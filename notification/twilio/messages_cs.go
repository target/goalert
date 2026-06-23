package twilio

import "golang.org/x/text/language"

// Czech voice translations. Registered under the base language tag so the
// regional variant (cs-CZ) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Czech, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Pro potvrzení zrušení registrace tohoto čísla stiskněte %s.",
		"To go back to the previous menu, press %s.":               "Pro návrat do předchozí nabídky stiskněte %s.",
		"To disable voice notifications to this number, press %s.": "Pro vypnutí hlasových upozornění na toto číslo stiskněte %s.",
		"To repeat this message, press star.":                      "Pro zopakování této zprávy stiskněte hvězdičku.",
		"To acknowledge, press %s.":                                "Pro potvrzení stiskněte %s.",
		"To escalate, press %s.":                                   "Pro eskalaci stiskněte %s.",
		"To close, press %s.":                                      "Pro uzavření stiskněte %s.",
		"To acknowledge all, press %s.":                            "Pro potvrzení všech stiskněte %s.",
		"To close all, press %s.":                                  "Pro uzavření všech stiskněte %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Pokud jste hotovi, můžete jednoduše zavěsit.",
		"Sorry, I didn't understand that.":         "Omlouvám se, nerozuměl jsem.",
		"Goodbye.":                                 "Na shledanou.",
		// call flow
		"Hello! This is %s":   "Dobrý den! Tady je %s",
		"Hello! This is %s. ": "Dobrý den! Tady je %s. ",
		"Please use the application dashboard to manage alerts.": "Ke správě upozornění použijte prosím nástěnku aplikace.",
		"Unenrolled.":        "Registrace zrušena.",
		"One moment please.": "Okamžik prosím.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Došlo k chybě. Ke správě upozornění použijte prosím nástěnku.",
		"The menu options have changed. To acknowledge, press %s.":          "Možnosti nabídky se změnily. Pro potvrzení stiskněte %s.",
		"The menu options have changed. To close, press %s.":                "Možnosti nabídky se změnily. Pro uzavření stiskněte %s.",
		// action confirmations
		"Acknowledged":                     "Potvrzeno",
		"Acknowledged all alerts.":         "Všechna upozornění byla potvrzena.",
		"Closed":                           "Uzavřeno",
		"Closed all alerts.":               "Všechna upozornění byla uzavřena.",
		"Escalation requested":             "Eskalace vyžádána",
		"Escalation requested all alerts.": "Eskalace vyžádána pro všechna upozornění.",
		// error messages
		"Already %s":                                "Již %s",
		"Alert is already closed.":                  "Upozornění je již uzavřeno.",
		"Alert is already acknowledged.":            "Upozornění je již potvrzeno.",
		"Error: %s":                                 "Chyba: %s",
		"System error. Please visit the dashboard.": "Systémová chyba. Navštivte prosím nástěnku.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s s upozorněními. Služba „%s“ má %d nepotvrzených upozornění.",
		"%s with an alert notification. %s.":                                      "%s s upozorněním. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s s aktualizací stavu upozornění „%s“. %s",
		"%s with a test message.":                                                 "%s s testovací zprávou.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s s vaším %d místným ověřovacím kódem. Kód je: %s. Znovu, váš %d místný ověřovací kód je: %s.",
		"No summary provided": "Není uveden žádný souhrn",
	})
}
