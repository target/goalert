package twilio

import "golang.org/x/text/language"

// Dutch voice translations. Registered under the base language tag so every
// regional variant (nl-BE, nl-NL) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Dutch, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Druk op %s om de afmelding van dit nummer te bevestigen.",
		"To go back to the previous menu, press %s.":               "Druk op %s om terug te gaan naar het vorige menu.",
		"To disable voice notifications to this number, press %s.": "Druk op %s om spraakmeldingen naar dit nummer uit te schakelen.",
		"To repeat this message, press star.":                      "Druk op sterretje om dit bericht te herhalen.",
		"To acknowledge, press %s.":                                "Druk op %s om te bevestigen.",
		"To escalate, press %s.":                                   "Druk op %s om te escaleren.",
		"To close, press %s.":                                      "Druk op %s om te sluiten.",
		"To acknowledge all, press %s.":                            "Druk op %s om alles te bevestigen.",
		"To close all, press %s.":                                  "Druk op %s om alles te sluiten.",
		// general prompts
		"If you are done, you may simply hang up.": "Als u klaar bent, kunt u gewoon ophangen.",
		"Sorry, I didn't understand that.":         "Sorry, dat heb ik niet begrepen.",
		"Goodbye.":                                 "Tot ziens.",
		// call flow
		"Hello! This is %s":   "Hallo! Dit is %s",
		"Hello! This is %s. ": "Hallo! Dit is %s. ",
		"Please use the application dashboard to manage alerts.": "Gebruik het dashboard van de applicatie om meldingen te beheren.",
		"Unenrolled.":        "Afgemeld.",
		"One moment please.": "Een ogenblik geduld alstublieft.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Er is een fout opgetreden. Gebruik het dashboard om meldingen te beheren.",
		"The menu options have changed. To acknowledge, press %s.":          "De menuopties zijn gewijzigd. Druk op %s om te bevestigen.",
		"The menu options have changed. To close, press %s.":                "De menuopties zijn gewijzigd. Druk op %s om te sluiten.",
		// action confirmations
		"Acknowledged":                     "Bevestigd",
		"Acknowledged all alerts.":         "Alle meldingen zijn bevestigd.",
		"Closed":                           "Gesloten",
		"Closed all alerts.":               "Alle meldingen zijn gesloten.",
		"Escalation requested":             "Escalatie aangevraagd",
		"Escalation requested all alerts.": "Escalatie aangevraagd voor alle meldingen.",
		// error messages
		"Already %s":                                "Al %s",
		"Alert is already closed.":                  "De melding is al gesloten.",
		"Alert is already acknowledged.":            "De melding is al bevestigd.",
		"Error: %s":                                 "Fout: %s",
		"System error. Please visit the dashboard.": "Systeemfout. Raadpleeg het dashboard.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s met meldingen. Service '%s' heeft %d onbevestigde meldingen.",
		"%s with an alert notification. %s.":                                      "%s met een melding. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s met een statusupdate voor melding '%s'. %s",
		"%s with a test message.":                                                 "%s met een testbericht.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s met uw verificatiecode van %d cijfers. De code is: %s. Nogmaals, uw verificatiecode van %d cijfers is: %s.",
		"No summary provided": "Geen samenvatting opgegeven",
	})
}
