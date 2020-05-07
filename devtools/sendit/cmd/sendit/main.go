package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/target/goalert/devtools/sendit"
)

func main() {
	token := flag.String("token", os.Getenv("SENDIT_TOKEN"), "Auth token to present to the server [SENDIT_TOKEN].")
	connTTL := flag.Duration("max-ttl", 15*time.Second, "Maximum time for a tunnel to exist before making a new request.")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	if flag.Arg(0) == "" {
		log.Fatal("source URL argument is required")
	}
	if flag.Arg(1) == "" {
		log.Fatal("destination URL argument is required")
	}

	u, err := url.Parse(flag.Arg(0))
	if err != nil {
		log.Fatal("invalid source URL:", err)
	}
	if u.Path == "" || u.Path == "/" {
		log.Fatal("source URL must contain a path prefix (e.g. /foobar)")
	}

	u, err = url.Parse(flag.Arg(1))
	if err != nil {
		log.Fatal("invalid destination URL:", err)
	}
	if u.Path != "" {
		log.Fatal("destination URL must not contain a path but found:", u.Path)
	}

	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		log.Println("ERROR:", sendit.ConnectAndServe(flag.Arg(0), flag.Arg(1), *token, *connTTL))
		<-t.C
	}
}
