package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"sync"

	jwt "github.com/dgrijalva/jwt-go"
)

func newJwt(expire int64, name string, signKey []byte, maxAge int64) *mjwt {
	return &mjwt{
		expire:       expire,
		name:         name,
		signKey:      signKey,
		cookieMaxAge: maxAge,
	}
}

// json web token
type mjwt struct {
	expire       int64
	name         string
	ip           string
	signKey      []byte
	cookieMaxAge int64
	lock         sync.RWMutex
}

type accessClaims struct {
	User        string `json:"User"`
	Passwd      string `json:"passwd"`
	ClientCheck string `json:"ClientCheck"`

	jwt.StandardClaims
}

func (j *mjwt) tokenString(claims jwt.Claims) string {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(j.signKey)
	if err != nil {
		log.Println(err)
	}
	return ss

}

func (j *mjwt) parse(tokenString string) (jwt.MapClaims, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// MySigningKey is a []byte containing your secret, e.g. []byte("my_secret_key")
		return j.signKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New(fmt.Sprintf("parse token %s error", tokenString))
	}
}

func (j *mjwt) accessToken(rw http.ResponseWriter, r *http.Request, user, passwd string) {
	tokenString := j.tokenString(accessClaims{
		user,
		passwd,
		fmt.Sprintf("%s|%s", getHttpClientIp(r), r.Header.Get("User-Agent")),
		jwt.StandardClaims{
			ExpiresAt: j.expire + time.Now().Unix(),
			Issuer:    "prod",
		},
	})
	http.SetCookie(rw, &http.Cookie{
		Name:     j.name,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(j.cookieMaxAge),
		Value:    url.QueryEscape(tokenString),
	})
}

func (j *mjwt) accessTempToken(rw http.ResponseWriter, r *http.Request, user, passwd string) {
	tokenString := j.tokenString(accessClaims{
		user,
		passwd,
		fmt.Sprintf("%s|%s", getHttpClientIp(r), r.Header.Get("User-Agent")),
		jwt.StandardClaims{
			ExpiresAt: j.expire + time.Now().Unix(),
			Issuer:    "prod",
		},
	})
	http.SetCookie(rw, &http.Cookie{
		Name:     j.name,
		HttpOnly: true,
		Path:     "/",
		Value:    url.QueryEscape(tokenString),
	})
}

func (j *mjwt) cleanCookie(rw http.ResponseWriter) {
	http.SetCookie(rw, &http.Cookie{
		Name:     j.name,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().AddDate(-1, 0, 0),
	})
}

func (j *mjwt) auth(rw http.ResponseWriter, r *http.Request, data *map[string]interface{}) bool {
	var tokenString string
	token, err := r.Cookie(j.name)
	if err != nil {
		log.Println(err)
		return false
	}
	tokenString, err = url.QueryUnescape(token.Value)
	if err != nil {
		log.Println(err)
		return false
	}

	m, err := j.parse(tokenString)
	if err != nil {
		log.Println(err)
		return false
	}
	if fmt.Sprintf("%s|%s", getHttpClientIp(r), r.Header.Get("User-Agent")) != m["ClientCheck"] {
		log.Println("client sign changed.")
		return false
	}

	*data = m
	return true

}
