package graphqlapp

import (
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/user/contactmethod"
)

// CompatDestToCMTypeVal converts a graphql2.DestinationInput to a contactmethod.Type and string value
// for the built-in destination types.
func CompatDestToCMTypeVal(d graphql2.DestinationInput) (contactmethod.Type, string) {
	switch d.Type {
	case destTwilioSMS:
		return contactmethod.TypeSMS, d.FieldValue(fieldPhoneNumber)
	case destTwilioVoice:
		return contactmethod.TypeVoice, d.FieldValue(fieldPhoneNumber)
	case destSMTP:
		return contactmethod.TypeEmail, d.FieldValue(fieldEmailAddress)
	case destWebhook:
		return contactmethod.TypeWebhook, d.FieldValue(fieldWebhookURL)
	case destSlackDM:
		return contactmethod.TypeSlackDM, d.FieldValue(fieldSlackUserID)
	}

	return "", ""
}
