package imapmanager

import (
	"fmt"

	"github.com/emersion/go-sasl"
)

// xoauth2Client implements the XOAUTH2 SASL mechanism for Gmail.
type xoauth2Client struct {
	username    string
	accessToken string
}

// NewXOAUTH2Client creates a new XOAUTH2 SASL client.
func NewXOAUTH2Client(username, accessToken string) sasl.Client {
	return &xoauth2Client{
		username:    username,
		accessToken: accessToken,
	}
}

func (c *xoauth2Client) Start() (string, []byte, error) {
	// XOAUTH2 sends the auth string immediately
	authString := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", c.username, c.accessToken)
	return "XOAUTH2", []byte(authString), nil
}

func (c *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	// XOAUTH2 is a one-step mechanism, so there should be no additional challenges
	// If we get a challenge, it's usually an error response
	// Return empty byte slice to complete the exchange
	return []byte{}, nil
}
