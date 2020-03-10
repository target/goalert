package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/target/goalert/devtools/sendit"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:5050", "Local address to listen or connect to.")
	server := flag.Bool("server", false, "Run in server mode.")
	secret := flag.String("secret", os.Getenv("SENDIT_SECRET"), "Secret signing string (server mode) or auth token (client mode).")
	prefix := flag.String("http-prefix", os.Getenv("SENDIT_HTTP_PREFIX"), "HTTP prefix (server mode).")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	if !*server {
		if strings.Contains(flag.Arg(0), *addr) {
			log.Fatal("ERROR: addr must not be part of connect URL.")
		}
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			log.Println("ERROR:", sendit.ConnectAndServe(flag.Arg(0), *addr, *secret))
			<-t.C
		}
	}

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening:", l.Addr().String())
	log.Fatal(http.Serve(l, sendit.NewServer([]byte(*secret), *prefix)))
}
