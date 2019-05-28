package remotemonitor

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var numRx = regexp.MustCompile(`^\+\d{1,15}$`)

func validPhone(n string) string {
	if !numRx.MatchString(n) {
		return ""
	}

	return n
}
func (m *Monitor) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/v1/twilio/sms/status" || req.URL.Path == "/api/v2/twilio/message/status" {
		// ignore status notifications
		return
	}
	from := validPhone(req.FormValue("From"))
	to := validPhone(req.FormValue("To"))
	if from == "" || to == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if from == m.cfg.Twilio.FromNumber || to != m.cfg.Twilio.FromNumber {
		// status or something we don't care about
		return
	}

	body := req.FormValue("Body")
	go m.processSMS(from, body)
}
func (m *Monitor) sendSMS(to, body string) {
	msg, err := m.tw.SendSMS(m.context(), to, body, nil)
	if err != nil {
		log.Printf("ERROR: sending '%s' to %s: %s\n", strconv.Quote(body), to, err.Error())
		return
	}
	log.Printf("SENT SMS: %s -> %s %s; SID=%s; Status=%s\n", msg.From, msg.To, strconv.Quote(body), msg.SID, msg.Status)
}

var actionRx = regexp.MustCompile(`'(\d+c)'`)

func (m *Monitor) processSMS(from, body string) {
	log.Println("INCOMING SMS:", from, strconv.Quote(body))
	var i Instance
	var found bool
	for _, search := range m.cfg.Instances {
		if search.Phone == from {
			i = search
			found = true
			break
		}
	}
	if !found {
		log.Println("ERROR: unknown SMS source:", from, strconv.Quote(body))
		return
	}

	if strings.Contains(strings.ToLower(body), "closed") {
		for _, err := range i.heartbeat() {
			m.reportErr(i, err, "post to heartbeat endpoint")
		}
		m.finishCh <- i.Location
		return
	}

	if p := actionRx.FindStringSubmatch(body); len(p) == 2 {
		m.sendSMS(from, p[1])
		return
	}

	log.Println("ERROR: unrecognized SMS message:", from, body)
}
