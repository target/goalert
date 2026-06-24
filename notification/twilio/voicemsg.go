package twilio

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

// voiceCat is the message catalog used to translate the text spoken during
// Twilio voice calls. English is the fallback language: any key without a
// translation for the selected language is spoken in English.
var voiceCat = catalog.NewBuilder(catalog.Fallback(language.English))

// voiceKeys lists every translatable English source string spoken during a
// voice call. These strings double as the catalog keys: a translation file maps
// each of them to its localized form. Keep this list in sync with the
// Say/Sayf/Sprintf call sites in voice.go and twiml.go — it is both the
// canonical key list and the reference handed to translators.
var voiceKeys = []string{
	// twiml.go — menu options
	"To confirm unenrollment of this number, press %s.",
	"To go back to the previous menu, press %s.",
	"To disable voice notifications to this number, press %s.",
	"To repeat this message, press star.",
	"To acknowledge, press %s.",
	"To escalate, press %s.",
	"To close, press %s.",
	"To acknowledge all, press %s.",
	"To close all, press %s.",
	// twiml.go — general prompts
	"If you are done, you may simply hang up.",
	"Sorry, I didn't understand that.",
	"Goodbye.",
	// voice.go — call flow
	"Hello! This is %s",
	"Hello! This is %s. ",
	"Please use the application dashboard to manage alerts.",
	"Unenrolled.",
	"One moment please.",
	"An error has occurred. Please use the dashboard to manage alerts.",
	"The menu options have changed. To acknowledge, press %s.",
	"The menu options have changed. To close, press %s.",
	// voice.go — action confirmations
	"Acknowledged",
	"Acknowledged all alerts.",
	"Closed",
	"Closed all alerts.",
	"Escalation requested",
	"Escalation requested all alerts.",
	// voice.go — error messages
	"Already %s",
	"Alert is already closed.",
	"Alert is already acknowledged.",
	"Error: %s",
	"System error. Please visit the dashboard.",
	// voice.go — buildMessage templates
	"%s with alert notifications. Service '%s' has %d unacknowledged alerts.",
	"%s with an alert notification. %s.",
	"%s with a status update for alert '%s'. %s",
	"%s with a test message.",
	"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.",
	"No summary provided",
}

func init() {
	// Register English as the identity catalog so it is part of the matcher's
	// supported-language set and every key resolves to its English source form.
	for _, k := range voiceKeys {
		_ = voiceCat.SetString(language.English, k, k)
	}
}

// registerVoiceMessages registers a set of translated voice messages for the
// given language. It is called from the per-language message files
// (messages_*.go). Empty translations are skipped so the message falls back to
// English.
func registerVoiceMessages(tag language.Tag, msgs map[string]string) {
	for k, v := range msgs {
		if v == "" {
			continue
		}
		_ = voiceCat.SetString(tag, k, v)
	}
}

// voicePrinter returns a message.Printer for the configured voice language
// along with the language code to set on the <Say> element.
//
// When voiceLang is empty the historical default is preserved: English text
// with no language attribute (Twilio uses its default voice). When voiceLang
// has no matching translation, it falls back to English text spoken with an
// English ("en-US") voice so the spoken language matches the text.
func voicePrinter(voiceLang string) (*message.Printer, string) {
	if voiceLang == "" {
		return message.NewPrinter(language.English, message.Catalog(voiceCat)), ""
	}

	desired := language.Make(voiceLang)
	if _, _, conf := voiceCat.Matcher().Match(desired); conf == language.No {
		return message.NewPrinter(language.English, message.Catalog(voiceCat)), "en-US"
	}

	return message.NewPrinter(desired, message.Catalog(voiceCat)), voiceLang
}
