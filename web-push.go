
package main

import (
	"fmt"
	"log"

	"github.com/SherClockHolmes/webpush-go"
)

func main() {
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("VAPID Public Key :", pub)  // base64url-encoded (no padding)
	fmt.Println("VAPID Private Key:", priv) // base64url-encoded (no padding)
}
