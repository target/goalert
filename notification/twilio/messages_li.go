package twilio

import "golang.org/x/text/language"

// Limburgish voice translations. Registered under the "li" language tag so the
// regional variant (li-NL) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Make("li"), map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Veur 't aafmelje van dit nummer te bevestige, druk %s.",
		"To go back to the previous menu, press %s.":               "Veur trök te gaon nao 't veurige menu, druk %s.",
		"To disable voice notifications to this number, press %s.": "Veur spraokmeldinge nao dit nummer oet te zètte, druk %s.",
		"To repeat this message, press star.":                      "Veur dit berich te herhaole, druk steer.",
		"To acknowledge, press %s.":                                "Veur te bevestige, druk %s.",
		"To escalate, press %s.":                                   "Veur te escalere, druk %s.",
		"To close, press %s.":                                      "Veur te sloete, druk %s.",
		"To acknowledge all, press %s.":                            "Veur alles te bevestige, druk %s.",
		"To close all, press %s.":                                  "Veur alles te sloete, druk %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Es geer klaor zeet, kènt geer gewoon ophange.",
		"Sorry, I didn't understand that.":         "Sorry, ich heb dat neet verstange.",
		"Goodbye.":                                 "Daag.",
		// call flow
		"Hello! This is %s":   "Hallo! Dit is %s",
		"Hello! This is %s. ": "Hallo! Dit is %s. ",
		"Please use the application dashboard to manage alerts.": "Gebruuk 't dashboard van de applicatie veur de meldinge te behere.",
		"Unenrolled.":        "Aafgemeld.",
		"One moment please.": "Ein memènt, aujublie.",
		"An error has occurred. Please use the dashboard to manage alerts.": "D'r is ein faut opgetraoje. Gebruuk 't dashboard veur de meldinge te behere.",
		"The menu options have changed. To acknowledge, press %s.":          "De menu-opties zeen verangerd. Veur te bevestige, druk %s.",
		"The menu options have changed. To close, press %s.":                "De menu-opties zeen verangerd. Veur te sloete, druk %s.",
		// action confirmations
		"Acknowledged":                     "Bevestig",
		"Acknowledged all alerts.":         "Alle meldinge bevestig.",
		"Closed":                           "Geslaote",
		"Closed all alerts.":               "Alle meldinge geslaote.",
		"Escalation requested":             "Escalatie aangevraog",
		"Escalation requested all alerts.": "Escalatie aangevraog veur alle meldinge.",
		// error messages
		"Already %s":                                "Al %s",
		"Alert is already closed.":                  "De melding is al geslaote.",
		"Alert is already acknowledged.":            "De melding is al bevestig.",
		"Error: %s":                                 "Faut: %s",
		"System error. Please visit the dashboard.": "Systeemfaut. Bezeuk 't dashboard.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s mèt meldinge. Service '%s' haet %d neet-bevestigde meldinge.",
		"%s with an alert notification. %s.":                                      "%s mèt ein melding. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s mèt ein statusupdate veur melding '%s'. %s",
		"%s with a test message.":                                                 "%s mèt ein testberich.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s mèt eure verificatiecode van %d cijfers. De code is: %s. Nog ein kier, eure verificatiecode van %d cijfers is: %s.",
		"No summary provided": "Gein samevatting opgegeve",
	})
}
