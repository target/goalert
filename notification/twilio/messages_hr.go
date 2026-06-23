package twilio

import "golang.org/x/text/language"

// Croatian voice translations. Registered under the base language tag so the
// regional variant (hr-HR) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Croatian, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Za potvrdu odjave ovog broja, pritisnite %s.",
		"To go back to the previous menu, press %s.":               "Za povratak na prethodni izbornik, pritisnite %s.",
		"To disable voice notifications to this number, press %s.": "Za isključivanje glasovnih obavijesti na ovom broju, pritisnite %s.",
		"To repeat this message, press star.":                      "Za ponavljanje ove poruke, pritisnite zvjezdicu.",
		"To acknowledge, press %s.":                                "Za potvrdu, pritisnite %s.",
		"To escalate, press %s.":                                   "Za eskalaciju, pritisnite %s.",
		"To close, press %s.":                                      "Za zatvaranje, pritisnite %s.",
		"To acknowledge all, press %s.":                            "Za potvrdu svih, pritisnite %s.",
		"To close all, press %s.":                                  "Za zatvaranje svih, pritisnite %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Ako ste gotovi, jednostavno spustite slušalicu.",
		"Sorry, I didn't understand that.":         "Oprostite, nisam razumio.",
		"Goodbye.":                                 "Doviđenja.",
		// call flow
		"Hello! This is %s":   "Dobar dan! Ovdje %s",
		"Hello! This is %s. ": "Dobar dan! Ovdje %s. ",
		"Please use the application dashboard to manage alerts.": "Molimo koristite nadzornu ploču aplikacije za upravljanje upozorenjima.",
		"Unenrolled.":        "Odjavljeno.",
		"One moment please.": "Trenutak, molim.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Došlo je do pogreške. Molimo koristite nadzornu ploču za upravljanje upozorenjima.",
		"The menu options have changed. To acknowledge, press %s.":          "Opcije izbornika su se promijenile. Za potvrdu, pritisnite %s.",
		"The menu options have changed. To close, press %s.":                "Opcije izbornika su se promijenile. Za zatvaranje, pritisnite %s.",
		// action confirmations
		"Acknowledged":                     "Potvrđeno",
		"Acknowledged all alerts.":         "Sva upozorenja su potvrđena.",
		"Closed":                           "Zatvoreno",
		"Closed all alerts.":               "Sva upozorenja su zatvorena.",
		"Escalation requested":             "Eskalacija zatražena",
		"Escalation requested all alerts.": "Eskalacija zatražena za sva upozorenja.",
		// error messages
		"Already %s":                                "Već %s",
		"Alert is already closed.":                  "Upozorenje je već zatvoreno.",
		"Alert is already acknowledged.":            "Upozorenje je već potvrđeno.",
		"Error: %s":                                 "Pogreška: %s",
		"System error. Please visit the dashboard.": "Sistemska pogreška. Molimo posjetite nadzornu ploču.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s s obavijestima o upozorenjima. Usluga „%s” ima %d nepotvrđenih upozorenja.",
		"%s with an alert notification. %s.":                                      "%s s obavijesti o upozorenju. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s s ažuriranjem statusa za upozorenje „%s”. %s",
		"%s with a test message.":                                                 "%s s testnom porukom.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s s vašim %d-znamenkastim verifikacijskim kodom. Kod je: %s. Ponovno, vaš %d-znamenkasti verifikacijski kod je: %s.",
		"No summary provided": "Nema sažetka.",
	})
}
