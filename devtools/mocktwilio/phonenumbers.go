package mocktwilio

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/target/goalert/notification/twilio"
)

func (s *Server) servePhoneNumbers(w http.ResponseWriter, req *http.Request) {
	m := make(map[string]twilio.PhoneNumberConfig)
	filter := req.FormValue("PhoneNumber")
	s.mx.RLock()
	for key, url := range s.callbacks {
		parts := strings.Split(key, ":")
		if filter != "" && parts[1] != filter {
			continue
		}
		cfg := m[parts[1]]
		cfg.SID = "MOCKTWILIO_" + parts[1]
		switch parts[0] {
		case "SMS":
			cfg.SMSMethod = "POST"
			cfg.SMSURL = url
			cfg.Capabilities.SMS = true
		case "VOICE":
			cfg.VoiceMethod = "POST"
			cfg.VoiceURL = url
			cfg.Capabilities.Voice = true
		}
		m[parts[1]] = cfg
	}
	s.mx.RUnlock()

	var resp struct {
		Numbers []twilio.PhoneNumberConfig `json:"incoming_phone_numbers"`
	}
	for _, cfg := range m {
		resp.Numbers = append(resp.Numbers, cfg)
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		panic(err)
	}
}
