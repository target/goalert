package twilio

import (
	"net/http"
)

const (
	msgParamID    = "msgID"
	msgParamSubID = "msgSubjectID"
	msgParamBody  = "msgBody"

	msgParamBundle = "msgBundle"
)

// Config contains the details needed to interact with Twilio for SMS
type Config struct {

	// APIURL can be used to override the Twilio API URL
	APIURL string

	// Client is an optional net/http client to use, if nil the global default is used.
	Client *http.Client
}
