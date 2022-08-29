package main

import (
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/target/goalert/devtools/mocktwilio"
)

func (s *State) HandleMessages(w http.ResponseWriter, req *http.Request) {
	dev := path.Base(req.URL.Path)

act:
	switch req.FormValue("action") {
	case "":
		type msgWrap struct {
			Message
			TimeHeader   string
			FriendlyTime string
		}
		var data struct {
			Device   string
			Name     string
			Error    string
			Messages []msgWrap
			*State
		}
		data.Device = dev
		data.Error = req.FormValue("error")
		data.Name = formatNumber(dev)
		data.State = s
		msgs := s.SMS()

		var prevTime time.Time
		for _, msg := range msgs {
			if msg.DeviceNumber != data.Device {
				continue
			}
			w := msgWrap{Message: msg, FriendlyTime: msg.Time.Format("Monday, Jan 2 3:04:05 PM MST 2006")}
			if msg.Time.Sub(prevTime) > 20*time.Minute {
				w.TimeHeader = timeHeader(msg.Time)
			}
			prevTime = msg.Time
			data.Messages = append(data.Messages, w)
		}
		// reverse order
		for i := len(data.Messages)/2 - 1; i >= 0; i-- {
			opp := len(data.Messages) - 1 - i
			data.Messages[i], data.Messages[opp] = data.Messages[opp], data.Messages[i]
		}
		render(w, "sms.html", data)
		return
	case "Accept":
		for _, msg := range s.SMS() {
			if msg.SMS == nil {
				continue
			}
			if msg.SMS.ID() != req.FormValue("id") {
				continue
			}

			err := msg.SMS.SetStatus(req.Context(), mocktwilio.MessageDelivered)
			if hasError(w, req, err) {
				return
			}

			break act
		}
		hasError(w, req, fmt.Errorf("unknown message %s", req.FormValue("id")))
		return
	case "Reject":
		for _, msg := range s.SMS() {
			if msg.SMS == nil {
				continue
			}
			if msg.SMS.ID() != req.FormValue("id") {
				continue
			}

			err := msg.SMS.SetStatus(req.Context(), mocktwilio.MessageFailed)
			if hasError(w, req, err) {
				return
			}

			break act
		}
		hasError(w, req, fmt.Errorf("unknown message %s", req.FormValue("id")))
		return

	case "sendSMS":
		msg, err := s.srv.SendMessage(req.Context(), dev, s.FromNumber, req.FormValue("body"))
		if hasError(w, req, err) {
			return
		}

		s.sendSMS <- sendSMS{From: msg.From(), Body: msg.Text()}
	default:
		hasError(w, req, fmt.Errorf("unknown action: %s", req.FormValue("action")))
		return
	}

	http.Redirect(w, req, req.URL.Path, http.StatusFound)
}
