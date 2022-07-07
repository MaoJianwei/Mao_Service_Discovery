package Email

import (
	"errors"
	"fmt"
	"net/smtp"
)


func AuthLOGIN(username, password string) smtp.Auth {
	return &authLOGIN{username, password}
}


type authLOGIN struct {
	username, password string
}

func (a *authLOGIN) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *authLOGIN) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New(fmt.Sprintf("Unknown message from Server: %s", fromServer))
		}
	}
	return nil, nil
}
