package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/target/goalert/devtools/sendit"
)

func main() {
	secret := flag.String("secret", "", "Signing secret to use for token generation.")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	tok, err := sendit.GenerateToken([]byte(*secret), sendit.TokenAudienceAuth, "")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(tok)
}
