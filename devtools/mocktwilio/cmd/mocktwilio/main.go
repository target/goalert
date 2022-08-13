package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/target/goalert/devtools/mocktwilio"
)

func main() {
	log.SetFlags(log.Lshortfile)

	var cfg mocktwilio.Config
	addr := flag.String("l", "localhost:3099", "Address to listen on.")
	flag.StringVar(&cfg.AccountSID, "sid", "AC00000000000000000000000000000000", "Mock Account SID")
	flag.StringVar(&cfg.AuthToken, "token", "00000000000000000000000000000000", "Mock Auth Token")
	flag.DurationVar(&cfg.MinQueueTime, "delay", time.Second, "Time to wait before sending a message.")
	fromNumber := flag.String("from", "+17635555555", "From phone number.")
	voiceURL := flag.String("voice-url", "http://localhost:3030/api/v2/twilio/call", "URL to receive voice calls.")
	smsURL := flag.String("sms-url", "http://localhost:3030/api/v2/twilio/message", "URL to receive SMS messages.")
	msgSID := flag.String("msg-sid", "MG00000000000000000000000000000000", "Message SID to simulate the messaging service.")
	saveFile := flag.String("save", "messages.txt", "File to save state to.")
	flag.Parse()

	srv := mocktwilio.NewServer(cfg)
	defer srv.Close()

	err := srv.RegisterSMSCallback(*fromNumber, *smsURL)
	if err != nil {
		log.Fatalln("register SMS callback:", err)
	}
	err = srv.RegisterVoiceCallback(*fromNumber, *voiceURL)
	if err != nil {
		log.Fatalln("register voice callback:", err)
	}
	err = srv.RegisterMessagingService(*msgSID, *smsURL, *fromNumber)
	if err != nil {
		log.Fatalln("register messaging service:", err)
	}

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalln("listen:", err)
	}
	defer l.Close()
	log.Println("Account SID:", cfg.AccountSID)
	log.Println("Auth Token: ", cfg.AuthToken)
	log.Println("Message SID:", *msgSID)
	log.Println("From Number:", *fromNumber)
	log.Println("UI:          http://" + *addr)

	s := NewState(srv, *saveFile)
	s.Config = cfg
	s.FromNumber = *fromNumber
	s.MessageSID = *msgSID

	mux := http.NewServeMux()
	mux.HandleFunc("/ui/", s.renderUI)
	mux.HandleFunc("/ui", s.renderUI)
	mux.Handle("/ui/assets/", http.StripPrefix("/ui/", http.FileServer(http.FS(assets))))
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
