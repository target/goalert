package main

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"log"
	"net"

	"github.com/oauth2-proxy/mockoidc"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9998", "Server listen address.")
	flag.Parse()

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	m, err := mockoidc.NewServer(rsaKey)
	if err != nil {
		log.Fatal(err)
	}
	m.ClientID = "test-client"
	m.ClientSecret = "test-secret"

	mockoidc.DefaultUser()

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	err = m.Start(l, nil)
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
