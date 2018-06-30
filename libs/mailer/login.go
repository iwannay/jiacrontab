package mailer

import (
	"fmt"
	"net/smtp"
)

type loginAuth struct {
	username, password string
}

// SMTP AUTH LOGIN Auth Handler
func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}
func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}
func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unknwon fromServer: %s", string(fromServer))
		}
	}
	return nil, nil
}
