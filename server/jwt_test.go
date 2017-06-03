package main

import (
	"fmt"
	"testing"
	"time"

	jwt2 "github.com/dgrijalva/jwt-go"
)

func TestGenJwt(t *testing.T) {
	conf := newConfig()
	jwt := newJwt(conf.tokenExpires, conf.tokenCookieName, conf.JWTSigningKey, conf.tokenCookieMaxAge)
	token := jwt.tokenString(accessClaims{
		"admin",
		"123456",
		"clientCheck",

		jwt2.StandardClaims{
			ExpiresAt: 1800 + time.Now().Unix(),
			Issuer:    "prod",
		},
	})
	t.Log(token)
}

func TestParseToken(t *testing.T) {
	conf := newConfig()
	jwt := newJwt(conf.tokenExpires, conf.tokenCookieName, conf.JWTSigningKey, conf.tokenCookieMaxAge)
	tokenString := jwt.tokenString(accessClaims{
		"admin",
		"123456",
		"clientCheck",
		jwt2.StandardClaims{
			ExpiresAt: 1800 + time.Now().Unix(),
			Issuer:    "prod",
		},
	})
	if token, err := jwt.parse(tokenString); err != nil {
		fmt.Println(token)
		t.Fatal(err)
	} else {
		t.Log(token)
	}
}

func TestCustom(t *testing.T) {

}
