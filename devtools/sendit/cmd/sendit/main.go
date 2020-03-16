package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/target/goalert/devtools/sendit"
)

func main() {
	token := flag.String("token", os.Getenv("SENDIT_TOKEN"), "Auth token to present to the server [SENDIT_TOKEN].")
	connTTL := flag.Duration("max-ttl", 15*time.Second, "Maximum time for a tunnel to exist before making a new request.")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		log.Println("ERROR:", sendit.ConnectAndServe(flag.Arg(0), flag.Arg(1), *token, *connTTL))
		<-t.C
	}
}
