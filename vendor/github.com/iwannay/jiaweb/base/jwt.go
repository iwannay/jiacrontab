package base

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func NewJwt(domain string, expire int64, tokenName string, signKey []byte, maxAge int64) *MJwt {
	return &MJwt{
		Domain:       domain,
		Expire:       expire,
		Name:         tokenName,
		SignKey:      signKey,
		CookieMaxAge: maxAge,
	}
}

// json web token
type MJwt struct {
	Domain       string
	Expire       int64
	Name         string
	SignKey      []byte
	CookieMaxAge int64
	lock         sync.RWMutex
}

// type accessClaims struct {
// 	user        string `json:"User"`
// 	passwd      string `json:"passwd"`
// 	clientCheck string `json:"ClientCheck"`

// 	jwt.StandardClaims
// }

func (j *MJwt) tokenString(claims jwt.MapClaims) string {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(j.SignKey)
	if err != nil {
		log.Println(err)
	}
	return ss

}

func (j *MJwt) parse(tokenString string) (jwt.MapClaims, error) {
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
		return j.SignKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New(fmt.Sprintf("parse token %s error", tokenString))

}

func (j *MJwt) GenerateToken(rw http.ResponseWriter, mapClaims jwt.MapClaims) {
	tokenString := j.tokenString(mapClaims)

	http.SetCookie(rw, &http.Cookie{
		Domain:   j.Domain,
		Name:     j.Name,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(j.CookieMaxAge),
		Value:    url.QueryEscape(tokenString),
	})
}

func (j *MJwt) GenerateSeesionToken(rw http.ResponseWriter, mapClaims jwt.MapClaims) {
	tokenString := j.tokenString(mapClaims)
	http.SetCookie(rw, &http.Cookie{
		Name:     j.Name,
		Domain:   j.Domain,
		HttpOnly: true,
		Path:     "/",
		Value:    url.QueryEscape(tokenString),
	})
}

func (j *MJwt) CleanCookie(rw http.ResponseWriter) {
	http.SetCookie(rw, &http.Cookie{
		Domain:   j.Domain,
		Name:     j.Name,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().AddDate(-1, 0, 0),
	})
}

func (j *MJwt) VerifyToken(r *http.Request, v *map[string]interface{}) bool {
	var tokenString string

	token, err := r.Cookie(j.Name)
	if err != nil {
		return false
	}
	tokenString, err = url.QueryUnescape(token.Value)
	if err != nil {
		return false
	}

	m, err := j.parse(tokenString)
	if err != nil {
		return false
	}

	*v = m
	return true

}
