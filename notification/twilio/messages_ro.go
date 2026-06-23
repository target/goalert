package twilio

import "golang.org/x/text/language"

// Romanian voice translations. Registered under the base language tag so the
// regional variant (ro-RO) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Romanian, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Pentru a confirma dezabonarea acestui număr, apăsați %s.",
		"To go back to the previous menu, press %s.":               "Pentru a reveni la meniul anterior, apăsați %s.",
		"To disable voice notifications to this number, press %s.": "Pentru a dezactiva notificările vocale pentru acest număr, apăsați %s.",
		"To repeat this message, press star.":                      "Pentru a repeta acest mesaj, apăsați asterisc.",
		"To acknowledge, press %s.":                                "Pentru a confirma, apăsați %s.",
		"To escalate, press %s.":                                   "Pentru a escalada, apăsați %s.",
		"To close, press %s.":                                      "Pentru a închide, apăsați %s.",
		"To acknowledge all, press %s.":                            "Pentru a confirma toate, apăsați %s.",
		"To close all, press %s.":                                  "Pentru a închide toate, apăsați %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Dacă ați terminat, puteți închide.",
		"Sorry, I didn't understand that.":         "Îmi pare rău, nu am înțeles.",
		"Goodbye.":                                 "La revedere.",
		// call flow
		"Hello! This is %s":   "Bună ziua! Aici %s",
		"Hello! This is %s. ": "Bună ziua! Aici %s. ",
		"Please use the application dashboard to manage alerts.": "Vă rugăm să folosiți panoul aplicației pentru a gestiona alertele.",
		"Unenrolled.":        "Dezabonat.",
		"One moment please.": "Un moment, vă rog.",
		"An error has occurred. Please use the dashboard to manage alerts.": "A apărut o eroare. Vă rugăm să folosiți panoul pentru a gestiona alertele.",
		"The menu options have changed. To acknowledge, press %s.":          "Opțiunile meniului s-au schimbat. Pentru a confirma, apăsați %s.",
		"The menu options have changed. To close, press %s.":                "Opțiunile meniului s-au schimbat. Pentru a închide, apăsați %s.",
		// action confirmations
		"Acknowledged":                     "Confirmată",
		"Acknowledged all alerts.":         "Toate alertele au fost confirmate.",
		"Closed":                           "Închisă",
		"Closed all alerts.":               "Toate alertele au fost închise.",
		"Escalation requested":             "Escaladare solicitată",
		"Escalation requested all alerts.": "Escaladare solicitată pentru toate alertele.",
		// error messages
		"Already %s":                                "Deja %s",
		"Alert is already closed.":                  "Alerta este deja închisă.",
		"Alert is already acknowledged.":            "Alerta este deja confirmată.",
		"Error: %s":                                 "Eroare: %s",
		"System error. Please visit the dashboard.": "Eroare de sistem. Vă rugăm să accesați panoul.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s cu notificări de alertă. Serviciul „%s” are %d alerte neconfirmate.",
		"%s with an alert notification. %s.":                                      "%s cu o notificare de alertă. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s cu o actualizare de stare pentru alerta „%s”. %s",
		"%s with a test message.":                                                 "%s cu un mesaj de test.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s cu codul dvs. de verificare din %d cifre. Codul este: %s. Din nou, codul dvs. de verificare din %d cifre este: %s.",
		"No summary provided": "Niciun rezumat furnizat",
	})
}
