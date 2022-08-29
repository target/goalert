package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/target/goalert/devtools/mocktwilio"
	"github.com/ttacon/libphonenumber"
)

type State struct {
	srv *mocktwilio.Server

	FromNumber string
	MessageSID string

	LastSent string
	LastCall string

	messages chan []Message
	calls    chan []PhoneCall

	saveFile string

	sendSMS chan sendSMS
}

func (s *State) Config() mocktwilio.Config { return s.srv.Config() }

type sendSMS struct {
	From string
	Body string
}

type Message struct {
	DeviceNumber string
	Body         string
	Time         time.Time
	Sent         bool

	SMS mocktwilio.Message `json:"-"`
}

type Conversation struct {
	ID              string
	Name            string
	LastMessage     string
	LastMessageTime time.Time
	Unread          bool
}

type PhoneCall struct {
	ID           string
	Name         string
	DeviceNumber string
	Time         time.Time

	Call mocktwilio.Call `json:"-"`
}

func (c Conversation) Since() string { return since(c.LastMessageTime) }
func (c PhoneCall) Since() string    { return since(c.Time) }

func NewState(srv *mocktwilio.Server, saveFile string) *State {
	s := &State{
		srv:      srv,
		messages: make(chan []Message),
		calls:    make(chan []PhoneCall),

		LastSent: "+16125555555",
		LastCall: "+16125555555",

		sendSMS:  make(chan sendSMS),
		saveFile: saveFile,
	}

	return s
}

func formatNumber(value string) string {
	num, err := libphonenumber.Parse(value, "")
	if err != nil {
		return value
	}

	return libphonenumber.Format(num, libphonenumber.INTERNATIONAL)
}

func (s *State) Conversations() []Conversation {
	convo := make(map[string]*Conversation)

	for _, msg := range s.SMS() {
		c := convo[msg.DeviceNumber]
		if c == nil {
			c = &Conversation{
				ID:   msg.DeviceNumber,
				Name: formatNumber(msg.DeviceNumber),
			}
			convo[msg.DeviceNumber] = c
		}
		if msg.Sent {
			msg.Body = "Sent: " + msg.Body
		}
		c.LastMessage = msg.Body
		c.LastMessageTime = msg.Time
		c.Unread = c.Unread || (msg.SMS != nil && msg.SMS.IsActive())
	}

	var convos []Conversation
	for _, c := range convo {
		convos = append(convos, *c)
	}

	sort.Slice(convos, func(i, j int) bool {
		// sort by unread, then time
		if convos[i].Unread != convos[j].Unread {
			return convos[i].Unread
		}
		return convos[i].LastMessageTime.After(convos[j].LastMessageTime)
	})

	return convos
}

// SMS returns the current set of SMS messages.
func (s *State) SMS() []Message {
	msgs := <-s.messages
	// clone slice
	return append([]Message{}, msgs...)
}

func (s *State) Calls() []PhoneCall { return <-s.calls }

func loadMessages(name string) ([]Message, *json.Encoder) {
	var msgs []Message
	if name == "" {
		return msgs, nil
	}

	fd, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		log.Printf("ERROR: open %s: %s", name, err)
		return msgs, nil
	}

	dec := json.NewDecoder(fd)
	for {
		var m Message
		err = dec.Decode(&m)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Printf("ERROR: decode message: %s", err)
			fd.Close()
			return msgs, nil
		}
		msgs = append(msgs, m)
	}

	return msgs, json.NewEncoder(fd)
}

func (s *State) loop() {
	msgs, enc := loadMessages(s.saveFile)

	add := func(dev, body string, sms mocktwilio.Message) {
		m := Message{
			DeviceNumber: dev,
			Body:         body,
			SMS:          sms,
			Sent:         sms == nil,
			Time:         time.Now(),
		}
		msgs = append(msgs, m)
		if enc != nil {
			enc.Encode(m)
		}
	}

	var calls []PhoneCall
	for {
		select {
		case call := <-s.srv.Calls():
			calls = append([]PhoneCall{{
				ID:           call.ID(),
				Name:         formatNumber(call.To()),
				DeviceNumber: call.To(),
				Time:         time.Now(),
				Call:         call,
			}}, calls...)
		case sms := <-s.srv.Messages():
			add(sms.To(), sms.Text(), sms)
		case send := <-s.sendSMS:
			add(send.From, send.Body, nil)
		case s.messages <- msgs:
		case s.calls <- calls:
		}
	}
}
