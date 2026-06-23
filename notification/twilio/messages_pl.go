package twilio

import "golang.org/x/text/language"

// Polish voice translations. Registered under the base language tag so every
// regional variant (pl-PL) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Polish, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Aby potwierdzić wyrejestrowanie tego numeru, naciśnij %s.",
		"To go back to the previous menu, press %s.":               "Aby wrócić do poprzedniego menu, naciśnij %s.",
		"To disable voice notifications to this number, press %s.": "Aby wyłączyć powiadomienia głosowe na tym numerze, naciśnij %s.",
		"To repeat this message, press star.":                      "Aby powtórzyć tę wiadomość, naciśnij gwiazdkę.",
		"To acknowledge, press %s.":                                "Aby potwierdzić, naciśnij %s.",
		"To escalate, press %s.":                                   "Aby eskalować, naciśnij %s.",
		"To close, press %s.":                                      "Aby zamknąć, naciśnij %s.",
		"To acknowledge all, press %s.":                            "Aby potwierdzić wszystkie, naciśnij %s.",
		"To close all, press %s.":                                  "Aby zamknąć wszystkie, naciśnij %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Jeśli skończyłeś, możesz po prostu się rozłączyć.",
		"Sorry, I didn't understand that.":         "Przepraszam, nie zrozumiałem.",
		"Goodbye.":                                 "Do widzenia.",
		// call flow
		"Hello! This is %s":   "Dzień dobry! Tu %s",
		"Hello! This is %s. ": "Dzień dobry! Tu %s. ",
		"Please use the application dashboard to manage alerts.": "Skorzystaj z panelu aplikacji, aby zarządzać alertami.",
		"Unenrolled.":        "Wyrejestrowano.",
		"One moment please.": "Chwileczkę proszę.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Wystąpił błąd. Skorzystaj z panelu, aby zarządzać alertami.",
		"The menu options have changed. To acknowledge, press %s.":          "Opcje menu uległy zmianie. Aby potwierdzić, naciśnij %s.",
		"The menu options have changed. To close, press %s.":                "Opcje menu uległy zmianie. Aby zamknąć, naciśnij %s.",
		// action confirmations
		"Acknowledged":                     "Potwierdzono",
		"Acknowledged all alerts.":         "Potwierdzono wszystkie alerty.",
		"Closed":                           "Zamknięto",
		"Closed all alerts.":               "Zamknięto wszystkie alerty.",
		"Escalation requested":             "Zażądano eskalacji",
		"Escalation requested all alerts.": "Zażądano eskalacji wszystkich alertów.",
		// error messages
		"Already %s":                                "Już %s",
		"Alert is already closed.":                  "Alert jest już zamknięty.",
		"Alert is already acknowledged.":            "Alert jest już potwierdzony.",
		"Error: %s":                                 "Błąd: %s",
		"System error. Please visit the dashboard.": "Błąd systemu. Skorzystaj z panelu.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s z powiadomieniami o alertach. Usługa „%s” ma %d niepotwierdzonych alertów.",
		"%s with an alert notification. %s.":                                      "%s z powiadomieniem o alercie. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s z aktualizacją statusu alertu „%s”. %s",
		"%s with a test message.":                                                 "%s z wiadomością testową.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s z Twoim %d-cyfrowym kodem weryfikacyjnym. Kod to: %s. Jeszcze raz, Twój %d-cyfrowy kod weryfikacyjny to: %s.",
		"No summary provided": "Brak podsumowania",
	})
}
