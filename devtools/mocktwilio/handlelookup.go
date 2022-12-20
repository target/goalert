package mocktwilio

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ttacon/libphonenumber"
)

type CarrierInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`

	// MobileCC is the mobile country code.
	MobileCC string `json:"mobile_country_code"`

	// MobileNC is the mobile network code.
	MobileNC string `json:"mobile_network_code"`
}

// handleLookup is a handler for the Twilio Lookup API at /v1/PhoneNumbers/<number>.
func (s *Server) handleLookup(w http.ResponseWriter, req *http.Request) {
	number := strings.TrimPrefix(req.URL.Path, "/v1/PhoneNumbers/")
	inclCarrier := req.URL.Query().Get("Type") == "carrier"

	var info struct {
		CallerName *struct{}    `json:"caller_name"`
		Carrier    *CarrierInfo `json:"carrier"`
		CC         string       `json:"country_code"`
		Fmt        string       `json:"national_format"`
		Number     string       `json:"phone_number"`
		URL        string       `json:"url"`
	}
	n, err := libphonenumber.Parse(number, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	info.CC = strconv.Itoa(int(n.GetCountryCode()))
	info.Fmt = libphonenumber.Format(n, libphonenumber.NATIONAL)
	info.Number = libphonenumber.Format(n, libphonenumber.E164)
	req.URL.Host = req.Host
	info.URL = req.URL.String()

	if inclCarrier {
		db := <-s.numInfoCh
		info.Carrier = db[info.Number]
		s.numInfoCh <- db
	}

	data, err := json.Marshal(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
