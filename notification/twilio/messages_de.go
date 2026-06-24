package twilio

import "golang.org/x/text/language"

// German voice translations. Registered under the base language tag so every
// regional variant (de-AT, de-DE) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.German, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Um die Abmeldung dieser Nummer zu bestätigen, drücken Sie %s.",
		"To go back to the previous menu, press %s.":               "Um zum vorherigen Menü zurückzukehren, drücken Sie %s.",
		"To disable voice notifications to this number, press %s.": "Um Sprachbenachrichtigungen an diese Nummer zu deaktivieren, drücken Sie %s.",
		"To repeat this message, press star.":                      "Um diese Nachricht zu wiederholen, drücken Sie die Sterntaste.",
		"To acknowledge, press %s.":                                "Um zu bestätigen, drücken Sie %s.",
		"To escalate, press %s.":                                   "Um zu eskalieren, drücken Sie %s.",
		"To close, press %s.":                                      "Um zu schließen, drücken Sie %s.",
		"To acknowledge all, press %s.":                            "Um alle zu bestätigen, drücken Sie %s.",
		"To close all, press %s.":                                  "Um alle zu schließen, drücken Sie %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Wenn Sie fertig sind, können Sie einfach auflegen.",
		"Sorry, I didn't understand that.":         "Entschuldigung, das habe ich nicht verstanden.",
		"Goodbye.":                                 "Auf Wiederhören.",
		// call flow
		"Hello! This is %s":   "Hallo! Hier ist %s",
		"Hello! This is %s. ": "Hallo! Hier ist %s. ",
		"Please use the application dashboard to manage alerts.": "Bitte verwenden Sie das Dashboard der Anwendung, um Alarme zu verwalten.",
		"Unenrolled.":        "Abgemeldet.",
		"One moment please.": "Einen Moment bitte.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Ein Fehler ist aufgetreten. Bitte verwenden Sie das Dashboard, um Alarme zu verwalten.",
		"The menu options have changed. To acknowledge, press %s.":          "Die Menüoptionen haben sich geändert. Um zu bestätigen, drücken Sie %s.",
		"The menu options have changed. To close, press %s.":                "Die Menüoptionen haben sich geändert. Um zu schließen, drücken Sie %s.",
		// action confirmations
		"Acknowledged":                     "Bestätigt",
		"Acknowledged all alerts.":         "Alle Alarme bestätigt.",
		"Closed":                           "Geschlossen",
		"Closed all alerts.":               "Alle Alarme geschlossen.",
		"Escalation requested":             "Eskalation angefordert",
		"Escalation requested all alerts.": "Eskalation für alle Alarme angefordert.",
		// error messages
		"Already %s":                                "Bereits %s",
		"Alert is already closed.":                  "Der Alarm ist bereits geschlossen.",
		"Alert is already acknowledged.":            "Der Alarm ist bereits bestätigt.",
		"Error: %s":                                 "Fehler: %s",
		"System error. Please visit the dashboard.": "Systemfehler. Bitte rufen Sie das Dashboard auf.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s mit Alarmbenachrichtigungen. Der Dienst „%s“ hat %d unbestätigte Alarme.",
		"%s with an alert notification. %s.":                                      "%s mit einer Alarmbenachrichtigung. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s mit einer Statusaktualisierung für den Alarm „%s“. %s",
		"%s with a test message.":                                                 "%s mit einer Testnachricht.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s mit Ihrem %d-stelligen Bestätigungscode. Der Code lautet: %s. Noch einmal, Ihr %d-stelliger Bestätigungscode lautet: %s.",
		"No summary provided": "Keine Zusammenfassung angegeben",
	})
}
