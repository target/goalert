package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/target/goalert/devtools/mocktwilio"
)

func main() {
	log.SetFlags(log.Lshortfile)

	var cfg mocktwilio.Config
	addr := flag.String("l", "localhost:3099", "Address to listen on.")
	flag.StringVar(&cfg.AccountSID, "sid", "AC00000000000000000000000000000000", "Mock Account SID")
	flag.StringVar(&cfg.PrimaryAuthToken, "token", "00000000000000000000000000000000", "Mock Auth Token")

	fromNumber := flag.String("from", "+17635555555", "From phone number.")
	voiceURL := flag.String("voice-url", "http://localhost:3030/api/v2/twilio/call", "URL to receive voice calls.")
	smsURL := flag.String("sms-url", "http://localhost:3030/api/v2/twilio/message", "URL to receive SMS messages.")
	msgSID := flag.String("msg-sid", "MG00000000000000000000000000000000", "Message SID to simulate the messaging service.")
	saveFile := flag.String("save", "messages.txt", "File to save state to.")
	flag.Parse()

	srv := mocktwilio.NewServer(cfg)
	defer srv.Close()

	err := srv.AddUpdateNumber(mocktwilio.Number{
		Number:          *fromNumber,
		VoiceWebhookURL: *voiceURL,
		SMSWebhookURL:   *smsURL,
	})
	if err != nil {
		log.Fatalln("register number:", err)
	}

	err = srv.AddUpdateMsgService(mocktwilio.MsgService{
		ID:            *msgSID,
		Numbers:       []string{*fromNumber},
		SMSWebhookURL: *smsURL,
	})
	if err != nil {
		log.Fatalln("register messaging service:", err)
	}

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalln("listen:", err)
	}
	defer l.Close()
	log.Println("Account SID:", cfg.AccountSID)
	log.Println("Auth Token: ", cfg.PrimaryAuthToken)
	log.Println("Message SID:", *msgSID)
	log.Println("From Number:", *fromNumber)
	log.Println("UI:          http://" + *addr)

	s := NewState(srv, *saveFile)
	s.FromNumber = *fromNumber
	s.MessageSID = *msgSID

	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	mux.Handle("/", srv)

	go s.loop()

	err = http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			log.Println(req.Method, req.URL.Path)
		}
		mux.ServeHTTP(w, req)
	}))
	if err != nil {
		log.Fatalln("http:", err)
	}
}
