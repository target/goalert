package twilio

import "golang.org/x/text/language"

// Danish voice translations. Registered under the base language tag so the
// regional variant (da-DK) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Danish, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Tryk %s for at bekræfte afmelding af dette nummer.",
		"To go back to the previous menu, press %s.":               "Tryk %s for at gå tilbage til forrige menu.",
		"To disable voice notifications to this number, press %s.": "Tryk %s for at deaktivere talebeskeder til dette nummer.",
		"To repeat this message, press star.":                      "Tryk stjerne for at gentage denne besked.",
		"To acknowledge, press %s.":                                "Tryk %s for at kvittere.",
		"To escalate, press %s.":                                   "Tryk %s for at eskalere.",
		"To close, press %s.":                                      "Tryk %s for at lukke.",
		"To acknowledge all, press %s.":                            "Tryk %s for at kvittere for alle.",
		"To close all, press %s.":                                  "Tryk %s for at lukke alle.",
		// general prompts
		"If you are done, you may simply hang up.": "Hvis du er færdig, kan du bare lægge på.",
		"Sorry, I didn't understand that.":         "Beklager, det forstod jeg ikke.",
		"Goodbye.":                                 "Farvel.",
		// call flow
		"Hello! This is %s":   "Hej! Dette er %s",
		"Hello! This is %s. ": "Hej! Dette er %s. ",
		"Please use the application dashboard to manage alerts.": "Brug venligst applikationens dashboard til at håndtere alarmer.",
		"Unenrolled.":        "Afmeldt.",
		"One moment please.": "Et øjeblik.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Der opstod en fejl. Brug venligst dashboardet til at håndtere alarmer.",
		"The menu options have changed. To acknowledge, press %s.":          "Menuvalgene er ændret. Tryk %s for at kvittere.",
		"The menu options have changed. To close, press %s.":                "Menuvalgene er ændret. Tryk %s for at lukke.",
		// action confirmations
		"Acknowledged":                     "Kvitteret",
		"Acknowledged all alerts.":         "Alle alarmer er kvitteret.",
		"Closed":                           "Lukket",
		"Closed all alerts.":               "Alle alarmer er lukket.",
		"Escalation requested":             "Eskalering anmodet",
		"Escalation requested all alerts.": "Eskalering anmodet for alle alarmer.",
		// error messages
		"Already %s":                                "Allerede %s",
		"Alert is already closed.":                  "Alarmen er allerede lukket.",
		"Alert is already acknowledged.":            "Alarmen er allerede kvitteret.",
		"Error: %s":                                 "Fejl: %s",
		"System error. Please visit the dashboard.": "Systemfejl. Besøg venligst dashboardet.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s med alarmbeskeder. Tjenesten '%s' har %d ukvitterede alarmer.",
		"%s with an alert notification. %s.":                                      "%s med en alarmbesked. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s med en statusopdatering for alarm '%s'. %s",
		"%s with a test message.":                                                 "%s med en testbesked.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s med din %d-cifrede verifikationskode. Koden er: %s. Igen, din %d-cifrede verifikationskode er: %s.",
		"No summary provided": "Ingen oversigt angivet",
	})
}
