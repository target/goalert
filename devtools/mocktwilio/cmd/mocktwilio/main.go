package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/target/goalert/devtools/mocktwilio"
)

func main() {
	log.SetFlags(log.Lshortfile)
	addr := flag.String("addr", "localhost:8087", "Address to listen on.")
	prefix := flag.String("prefix", "", "API URL prefix.")
	minQueueTime := flag.Duration("min-time", 100*time.Millisecond, "Minimum amount of time to wait to deliver an SMS or call.")
	sid := flag.String("account-sid", "", "Account SID to use.")
	token := flag.String("auth-token", "", "Auth token to require for the mock account.")
	sms := flag.String("sms", "", "List of comma-separated <+phone>=<url> pairs for callback URLs for incoming SMS messages.")
	voice := flag.String("voice", "", "List of comma-separated <+phone>=<url> pairs for callback URLs for incoming voice calls.")
	flag.Parse()

	srv := mocktwilio.NewServer(mocktwilio.Config{
		AccountSID:   *sid,
		AuthToken:    *token,
		MinQueueTime: *minQueueTime,
	})

	log.Printf("AccountSID      = %s", *sid)
	log.Printf("AuthToken       = %s", *token)

	for _, str := range strings.Split(*sms, ",") {
		parts := strings.SplitN(str, "=", 2)
		err := srv.RegisterSMSCallback(parts[0], parts[1])
		if err != nil {
			log.Fatal("ERROR: register sms: ", err)
		}
	}

	for _, str := range strings.Split(*voice, ",") {
		parts := strings.SplitN(str, "=", 2)
		err := srv.RegisterVoiceCallback(parts[0], parts[1])
		if err != nil {
			log.Fatal("ERROR: register sms: ", err)
		}
	}

	go func() {
		for {
			select {
			case err := <-srv.Errors():
				log.Println("ERROR:", err)
			case sms := <-srv.SMS():
				sms.Accept()
				log.Printf("SMS: FROM=%s;TO=%s;BODY=%s", sms.From(), sms.To(), sms.Body())
			case call := <-srv.VoiceCalls():
				call.Accept()
				log.Printf("Voice: FROM=%s;TO=%s;BODY=%s", call.From(), call.To(), call.Body())
				call.Hangup()
			}
		}
	}()

	err := http.ListenAndServe(*addr, http.StripPrefix(*prefix, srv))
	if err != nil {
		log.Fatal("ERROR: start http: ", err)
	}
}
