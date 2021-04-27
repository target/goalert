package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"

	"github.com/target/goalert/sysapi"
	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", "localhost:1234", "Server address.")
	flag.Parse()

	ctx := context.Background()
	// creds, err := credentials.NewClientTLSFromFile("server-cert.pem", "localhost")
	// if err != nil {
	// 	panic(err)
	// }

	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := sysapi.NewSysAPIClient(conn)
	resp, err := c.AuthSubjects(ctx, &sysapi.AuthSubjectsRequest{ProviderId: "basic"})
	if err != nil {
		panic(err)
	}
	for {
		sub, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}
		log.Println(sub.ProviderId, sub.SubjectId, sub.UserId)
	}

}
