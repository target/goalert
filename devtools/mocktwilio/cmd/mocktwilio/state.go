package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/target/goalert/devtools/mocktwilio"
	"github.com/ttacon/libphonenumber"
)

type State struct {
	mocktwilio.Config
	FromNumber string
	MessageSID string

	LastSent string
	LastCall string

	messages chan []Message

	saveFile string

	sendSMS chan sendSMS

	srv *mocktwilio.Server
}

type sendSMS struct {
	From string
	Body string
}

type Message struct {
	DeviceNumber string
	Body         string
	Time         time.Time
	Sent         bool

	SMS *mocktwilio.SMS `json:"-"`
}

type Conversation struct {
	ID              string
	Name            string
	LastMessage     string
	LastMessageTime time.Time
	Unread          bool
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func (c Conversation) Since() string {
	switch {
	case time.Since(c.LastMessageTime) < time.Minute:
		return "Just Now"
	case time.Since(c.LastMessageTime) < time.Hour:
		return fmt.Sprintf("%d min", int(time.Since(c.LastMessageTime).Minutes()))
	case sameDate(c.LastMessageTime, time.Now()):
		return c.LastMessageTime.Format("3:04 PM")
	case c.LastMessageTime.After(time.Now().AddDate(0, 0, -7)):
		return c.LastMessageTime.Format("Mon")
	case c.LastMessageTime.Year() == time.Now().Year():
		return c.LastMessageTime.Format("Jan 2")
	default:
		return c.LastMessageTime.Format("1/2/06")
	}
}

func NewState(srv *mocktwilio.Server, saveFile string) *State {
	s := &State{
		srv:      srv,
		messages: make(chan []Message),

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
func (s *State) SMS() []Message { return <-s.messages }

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

	add := func(dev, body string, sms *mocktwilio.SMS) {
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

	for {
		select {
		case sms := <-s.srv.SMS():
			add(sms.To(), sms.Body(), sms)
		case send := <-s.sendSMS:
			add(send.From, send.Body, nil)
			go func() {
				err := s.srv.SendSMS(send.From, s.FromNumber, send.Body)
				if err != nil {
					log.Println("ERROR: send sms:", err)
				}
			}()
		case s.messages <- msgs:
		}
	}
}
