package twilio

import "golang.org/x/text/language"

// Swedish voice translations. Registered under the base language tag so the
// regional variant (sv-SE) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Swedish, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Tryck %s för att bekräfta avregistreringen av det här numret.",
		"To go back to the previous menu, press %s.":               "Tryck %s för att gå tillbaka till föregående meny.",
		"To disable voice notifications to this number, press %s.": "Tryck %s för att stänga av röstaviseringar till det här numret.",
		"To repeat this message, press star.":                      "Tryck stjärna för att upprepa det här meddelandet.",
		"To acknowledge, press %s.":                                "Tryck %s för att kvittera.",
		"To escalate, press %s.":                                   "Tryck %s för att eskalera.",
		"To close, press %s.":                                      "Tryck %s för att stänga.",
		"To acknowledge all, press %s.":                            "Tryck %s för att kvittera alla.",
		"To close all, press %s.":                                  "Tryck %s för att stänga alla.",
		// general prompts
		"If you are done, you may simply hang up.": "Om du är klar kan du bara lägga på.",
		"Sorry, I didn't understand that.":         "Förlåt, jag uppfattade inte det.",
		"Goodbye.":                                 "Hej då.",
		// call flow
		"Hello! This is %s":   "Hej! Det här är %s",
		"Hello! This is %s. ": "Hej! Det här är %s. ",
		"Please use the application dashboard to manage alerts.": "Använd applikationens instrumentpanel för att hantera larm.",
		"Unenrolled.":        "Avregistrerad.",
		"One moment please.": "Ett ögonblick, tack.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Ett fel har uppstått. Använd instrumentpanelen för att hantera larm.",
		"The menu options have changed. To acknowledge, press %s.":          "Menyalternativen har ändrats. Tryck %s för att kvittera.",
		"The menu options have changed. To close, press %s.":                "Menyalternativen har ändrats. Tryck %s för att stänga.",
		// action confirmations
		"Acknowledged":                     "Kvitterat",
		"Acknowledged all alerts.":         "Alla larm har kvitterats.",
		"Closed":                           "Stängt",
		"Closed all alerts.":               "Alla larm har stängts.",
		"Escalation requested":             "Eskalering begärd",
		"Escalation requested all alerts.": "Eskalering begärd för alla larm.",
		// error messages
		"Already %s":                                "Redan %s",
		"Alert is already closed.":                  "Larmet är redan stängt.",
		"Alert is already acknowledged.":            "Larmet är redan kvitterat.",
		"Error: %s":                                 "Fel: %s",
		"System error. Please visit the dashboard.": "Systemfel. Besök instrumentpanelen.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s med larmaviseringar. Tjänsten '%s' har %d okvitterade larm.",
		"%s with an alert notification. %s.":                                      "%s med en larmavisering. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s med en statusuppdatering för larmet '%s'. %s",
		"%s with a test message.":                                                 "%s med ett testmeddelande.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s med din %d-siffriga verifieringskod. Koden är: %s. Igen, din %d-siffriga verifieringskod är: %s.",
		"No summary provided": "Ingen sammanfattning angiven",
	})
}
