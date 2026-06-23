package twilio

import "golang.org/x/text/language"

// Italian voice translations. Registered under the base language tag so every
// regional variant (it-IT) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Italian, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Per confermare la cancellazione di questo numero, premi %s.",
		"To go back to the previous menu, press %s.":               "Per tornare al menu precedente, premi %s.",
		"To disable voice notifications to this number, press %s.": "Per disattivare le notifiche vocali su questo numero, premi %s.",
		"To repeat this message, press star.":                      "Per ripetere questo messaggio, premi asterisco.",
		"To acknowledge, press %s.":                                "Per confermare la presa in carico, premi %s.",
		"To escalate, press %s.":                                   "Per inoltrare, premi %s.",
		"To close, press %s.":                                      "Per chiudere, premi %s.",
		"To acknowledge all, press %s.":                            "Per prendere in carico tutte, premi %s.",
		"To close all, press %s.":                                  "Per chiudere tutte, premi %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Se hai finito, puoi semplicemente riagganciare.",
		"Sorry, I didn't understand that.":         "Scusa, non ho capito.",
		"Goodbye.":                                 "Arrivederci.",
		// call flow
		"Hello! This is %s":   "Ciao! Questo è %s",
		"Hello! This is %s. ": "Ciao! Questo è %s. ",
		"Please use the application dashboard to manage alerts.": "Usa la dashboard dell'applicazione per gestire gli avvisi.",
		"Unenrolled.":        "Cancellato.",
		"One moment please.": "Un momento, per favore.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Si è verificato un errore. Usa la dashboard per gestire gli avvisi.",
		"The menu options have changed. To acknowledge, press %s.":          "Le opzioni del menu sono cambiate. Per confermare la presa in carico, premi %s.",
		"The menu options have changed. To close, press %s.":                "Le opzioni del menu sono cambiate. Per chiudere, premi %s.",
		// action confirmations
		"Acknowledged":                     "Presa in carico",
		"Acknowledged all alerts.":         "Tutti gli avvisi sono stati presi in carico.",
		"Closed":                           "Chiusa",
		"Closed all alerts.":               "Tutti gli avvisi sono stati chiusi.",
		"Escalation requested":             "Inoltro richiesto",
		"Escalation requested all alerts.": "Inoltro richiesto per tutti gli avvisi.",
		// error messages
		"Already %s":                                "Già %s",
		"Alert is already closed.":                  "L'avviso è già chiuso.",
		"Alert is already acknowledged.":            "L'avviso è già stato preso in carico.",
		"Error: %s":                                 "Errore: %s",
		"System error. Please visit the dashboard.": "Errore di sistema. Consulta la dashboard.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s con notifiche di avviso. Il servizio «%s» ha %d avvisi non presi in carico.",
		"%s with an alert notification. %s.":                                      "%s con una notifica di avviso. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s con un aggiornamento di stato per l'avviso «%s». %s",
		"%s with a test message.":                                                 "%s con un messaggio di prova.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s con il tuo codice di verifica di %d cifre. Il codice è: %s. Di nuovo, il tuo codice di verifica di %d cifre è: %s.",
		"No summary provided": "Nessun riepilogo fornito",
	})
}
