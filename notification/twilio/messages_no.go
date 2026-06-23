package twilio

import "golang.org/x/text/language"

// Norwegian (Bokmål) voice translations. Registered under the base language tag
// so the regional variant (no-NO) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Norwegian, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "For å bekrefte avregistrering av dette nummeret, trykk %s.",
		"To go back to the previous menu, press %s.":               "For å gå tilbake til forrige meny, trykk %s.",
		"To disable voice notifications to this number, press %s.": "For å slå av talevarsler til dette nummeret, trykk %s.",
		"To repeat this message, press star.":                      "For å gjenta denne meldingen, trykk stjerne.",
		"To acknowledge, press %s.":                                "For å bekrefte, trykk %s.",
		"To escalate, press %s.":                                   "For å eskalere, trykk %s.",
		"To close, press %s.":                                      "For å lukke, trykk %s.",
		"To acknowledge all, press %s.":                            "For å bekrefte alle, trykk %s.",
		"To close all, press %s.":                                  "For å lukke alle, trykk %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Hvis du er ferdig, kan du bare legge på.",
		"Sorry, I didn't understand that.":         "Beklager, jeg forsto ikke det.",
		"Goodbye.":                                 "Ha det.",
		// call flow
		"Hello! This is %s":   "Hei! Dette er %s",
		"Hello! This is %s. ": "Hei! Dette er %s. ",
		"Please use the application dashboard to manage alerts.": "Vennligst bruk applikasjonens dashbord for å håndtere varsler.",
		"Unenrolled.":        "Avregistrert.",
		"One moment please.": "Et øyeblikk, takk.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Det har oppstått en feil. Vennligst bruk dashbordet for å håndtere varsler.",
		"The menu options have changed. To acknowledge, press %s.":          "Menyvalgene har endret seg. For å bekrefte, trykk %s.",
		"The menu options have changed. To close, press %s.":                "Menyvalgene har endret seg. For å lukke, trykk %s.",
		// action confirmations
		"Acknowledged":                     "Bekreftet",
		"Acknowledged all alerts.":         "Alle varsler er bekreftet.",
		"Closed":                           "Lukket",
		"Closed all alerts.":               "Alle varsler er lukket.",
		"Escalation requested":             "Eskalering forespurt",
		"Escalation requested all alerts.": "Eskalering forespurt for alle varsler.",
		// error messages
		"Already %s":                                "Allerede %s",
		"Alert is already closed.":                  "Varselet er allerede lukket.",
		"Alert is already acknowledged.":            "Varselet er allerede bekreftet.",
		"Error: %s":                                 "Feil: %s",
		"System error. Please visit the dashboard.": "Systemfeil. Vennligst gå til dashbordet.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s med varslinger. Tjenesten «%s» har %d ubekreftede varsler.",
		"%s with an alert notification. %s.":                                      "%s med en varsling. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s med en statusoppdatering for varsel «%s». %s",
		"%s with a test message.":                                                 "%s med en testmelding.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s med din %d-sifrede verifiseringskode. Koden er: %s. En gang til, din %d-sifrede verifiseringskode er: %s.",
		"No summary provided": "Ingen oppsummering oppgitt",
	})
}
