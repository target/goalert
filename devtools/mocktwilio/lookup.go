package mocktwilio

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"

	"github.com/target/goalert/notification/twilio"
	"github.com/ttacon/libphonenumber"
)

func (s *Server) serveLookup(w http.ResponseWriter, req *http.Request) {
	number := path.Base(req.URL.Path)
	inclCarrier := req.URL.Query().Get("Type") == "carrier"

	var info struct {
		CallerName *struct{}           `json:"caller_name"`
		Carrier    *twilio.CarrierInfo `json:"carrier"`
		CC         string              `json:"country_code"`
		Fmt        string              `json:"national_format"`
		Number     string              `json:"phone_number"`
		AddOns     *struct{}           `json:"add_ons"`
		URL        string              `json:"url"`
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
		s.carrierInfoMx.Lock()
		c, ok := s.carrierInfo[info.Number]
		s.carrierInfoMx.Unlock()
		if ok {
			info.Carrier = &c
		}
	}

	data, err := json.Marshal(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
