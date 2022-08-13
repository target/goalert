package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"time"
)

//go:embed assets
var assets embed.FS

var tmpl = template.Must(template.New("").ParseFS(assets, "assets/*.html"))

// replace all nondigits
var rx = regexp.MustCompile(`[^0-9]+`)

func (m Message) Classname() string {
	if m.SMS == nil {
		return ""
	}

	if m.SMS.IsRejected() {
		return "rejected"
	}

	return ""
}

func (m Message) Error() string {
	if m.SMS == nil {
		return ""
	}

	if m.SMS.IsRejected() {
		return "Rejected"
	}

	return ""
}

func (m Message) Active() bool { return m.SMS != nil && m.SMS.IsActive() }
func (m Message) Status() string {
	if m.SMS == nil {
		return ""
	}

	switch {
	case m.SMS.IsAccepted():
		return "Accepted"
	case m.SMS.IsActive():
		return "Pending"
	case m.SMS.IsRejected():
		return "Rejected"
	}

	return "Unknown"
}

func timeHeader(eventTime time.Time) string {
	if sameDate(eventTime, time.Now()) {
		return eventTime.Format("3:04 PM")
	}

	// if yesterday
	if sameDate(eventTime, time.Now().Add(-24*time.Hour)) {
		return "Yesterday - " + eventTime.Format("3:04 PM")
	}

	return eventTime.Format("Monday - 3:04 PM")
}

func (s *State) renderUI(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/ui/action/sms":
		for _, msg := range s.SMS() {
			if msg.SMS == nil {
				continue
			}

			if msg.SMS.ID() != req.FormValue("id") {
				continue
			}
			switch req.FormValue("action") {
			case "accept":
				msg.SMS.Accept()
			case "reject":
				msg.SMS.Reject()
			}
			break
		}
	case "/ui/action/send":
		num := "+" + rx.ReplaceAllString(req.FormValue("from"), "")
		log.Printf("send sms: from=%s body=%s", num, req.FormValue("body"))
		s.LastSent = num
		s.sendSMS <- sendSMS{From: num, Body: req.FormValue("body")}

	case "/ui/action/call":
		log.Println("ERROR: starting voice calls not yet supported")
	case "/ui/sms":
		type msgWrap struct {
			Message
			TimeHeader   string
			FriendlyTime string
		}

		var data struct {
			Device   string
			Name     string
			Messages []msgWrap
		}
		data.Device = req.FormValue("dev")
		data.Name = formatNumber(data.Device)
		msgs := s.SMS()
		var prevTime time.Time
		for _, msg := range msgs {
			if msg.DeviceNumber != data.Device {
				continue
			}
			w := msgWrap{Message: msg, FriendlyTime: msg.Time.Format("Monday, Jan 2 3:04:05 PM MST 2006")}
			if msg.Time.Sub(prevTime) > 5*time.Minute {
				w.TimeHeader = timeHeader(msg.Time)
			}
			prevTime = msg.Time
			data.Messages = append(data.Messages, w)
		}

		err := tmpl.ExecuteTemplate(w, "sms.html", data)
		if err != nil {
			// print error to output
			fmt.Fprintf(w, `"><hr style="color:red"><code>%v</code>`, err)
		}
		return
	case "/ui":
		err := tmpl.ExecuteTemplate(w, "index.html", s)
		if err != nil {
			// print error to output
			fmt.Fprintf(w, `"><hr style="color:red"><code>%v</code>`, err)
		}
		return
	default:
		http.NotFound(w, req)
		return
	}

	// redirect back to referer
	http.Redirect(w, req, req.Referer(), http.StatusFound)
}
