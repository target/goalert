package cloudwatch

import (
	"fmt"
	"strings"
)

// SNS message types.
const (
	typeNotification             = "Notification"
	typeSubscriptionConfirmation = "SubscriptionConfirmation"
	typeUnsubscribeConfirmation  = "UnsubscribeConfirmation"
)

// signedFields lists the fields AWS includes in the string-to-sign for each
// message type, in the alphabetical order required by the signing scheme. The
// lists are literals rather than sorted at runtime; canonicalOrderTest asserts
// they are in fact sorted.
//
// https://docs.aws.amazon.com/sns/latest/dg/sns-verify-signature-of-message.html
var signedFields = map[string][]string{
	typeNotification:             {"Message", "MessageId", "Subject", "Timestamp", "TopicArn", "Type"},
	typeSubscriptionConfirmation: {"Message", "MessageId", "SubscribeURL", "Timestamp", "Token", "TopicArn", "Type"},
	typeUnsubscribeConfirmation:  {"Message", "MessageId", "SubscribeURL", "Timestamp", "Token", "TopicArn", "Type"},
}

// envelope is the SNS HTTP/S message envelope.
//
// https://docs.aws.amazon.com/sns/latest/dg/sns-message-and-json-formats.html
type envelope struct {
	Type      string `json:"Type"`
	MessageID string `json:"MessageId"`
	TopicARN  string `json:"TopicArn"`
	Message   string `json:"Message"`

	// Subject must stay a pointer. AWS omits the Subject block from the
	// string-to-sign entirely when the field is absent, and a plain string
	// cannot distinguish an absent key from `"Subject": ""`.
	Subject *string `json:"Subject"`

	// Timestamp is signed as the raw string; never parse and re-format it for
	// the canonical string.
	Timestamp string `json:"Timestamp"`

	SignatureVersion string `json:"SignatureVersion"`
	Signature        string `json:"Signature"`
	SigningCertURL   string `json:"SigningCertURL"`

	// Present on the two *Confirmation types. Token is a bearer credential that
	// can confirm or cancel a subscription: never log it above debug.
	SubscribeURL string `json:"SubscribeURL"`
	Token        string `json:"Token"`
}

// signedField returns the value of a field in the string-to-sign, and whether it
// is present at all. Absent fields are omitted from the canonical string rather
// than contributing an empty value.
func (e *envelope) signedField(name string) (value string, present bool) {
	switch name {
	case "Message":
		return e.Message, true
	case "MessageId":
		return e.MessageID, true
	case "Subject":
		if e.Subject == nil {
			return "", false
		}
		return *e.Subject, true
	case "SubscribeURL":
		return e.SubscribeURL, true
	case "Timestamp":
		return e.Timestamp, true
	case "Token":
		return e.Token, true
	case "TopicArn":
		return e.TopicARN, true
	case "Type":
		return e.Type, true
	}
	// Unreachable: signedFields is a package-level literal.
	panic("cloudwatch: unknown signed field " + name)
}

// canonicalString builds the AWS SNS string-to-sign. It is pure: no I/O, no
// clock, no package state.
func canonicalString(e *envelope) (string, error) {
	fields, ok := signedFields[e.Type]
	if !ok {
		return "", fmt.Errorf("cloudwatch: unknown message type %q", e.Type)
	}

	var b strings.Builder
	for _, f := range fields {
		v, present := e.signedField(f)
		if !present {
			continue
		}
		b.WriteString(f)
		b.WriteByte('\n')
		b.WriteString(v)
		b.WriteByte('\n')
	}

	return b.String(), nil
}
