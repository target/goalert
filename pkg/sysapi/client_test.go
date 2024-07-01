package sysapi_test

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/target/goalert/pkg/sysapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// List auth subjects from an insecure server (no certs configured)
func ExampleSysAPIClient_AuthSubjects() {
	target := flag.String("target", "localhost:1234", "Server address.")
	flag.Parse()

	conn, err := grpc.NewClient(*target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := sysapi.NewSysAPIClient(conn)
	if err != nil {
		log.Fatal(err)
	}

	c, err := client.AuthSubjects(context.Background(), &sysapi.AuthSubjectsRequest{})
	if err != nil {
		log.Fatal(err)
	}
	for {
		sub, err := c.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(sub.UserId, sub.ProviderId, sub.SubjectId)
	}
}

// Delete a user from a secure server (certificates configured)
func ExampleSysAPIClient_DeleteUser() {
	target := flag.String("target", "localhost:1234", "Server address.")
	caFile := flag.String("ca", "../../goalert-client.ca.pem", "CA cert file.")
	certFile := flag.String("cert", "../../goalert-client.pem", "Cert file.")
	keyFile := flag.String("key", "../../goalert-client.key", "Key file.")
	flag.Parse()

	cfg, err := sysapi.NewTLS(*caFile, *certFile, *keyFile)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.NewClient(*target, grpc.WithTransportCredentials(credentials.NewTLS(cfg)))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := sysapi.NewSysAPIClient(conn)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.DeleteUser(context.Background(), &sysapi.DeleteUserRequest{UserId: "foobar"})
	if err != nil {
		log.Fatal(err)
	}
}
