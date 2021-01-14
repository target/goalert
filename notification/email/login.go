package email

import (
	"errors"
	"fmt"
	"net/smtp"
)

type loginAuth struct {
	host string
	user string
	pass string
}

// LoginAuth implements the LOGIN authentication mechanism.
//
// Adapted from smtp.PlainAuth in the standard library.
func LoginAuth(user, pass, host string) smtp.Auth {
	return &loginAuth{host: host, user: user, pass: pass}
}

func isLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// Must have TLS, or else localhost server.
	// Note: If TLS is not true, then we can't trust ANYTHING in ServerInfo.
	// In particular, it doesn't matter if the server advertises PLAIN auth.
	// That might just be the attacker saying
	// "it's ok, you can trust me with your password."
	if !server.TLS && !isLocalhost(server.Name) {
		return "", nil, errors.New("unencrypted connection")
	}
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	switch string(fromServer) {
	case "Username:":
		return []byte(a.user), nil
	case "Password:":
		return []byte(a.pass), nil
	}

	return nil, fmt.Errorf("unexpected server challenge: %s", string(fromServer))
}
