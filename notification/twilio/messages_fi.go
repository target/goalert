package twilio

import "golang.org/x/text/language"

// Finnish voice translations. Registered under the base language tag so the
// regional variant (fi-FI) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Finnish, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Vahvista tämän numeron poisto painamalla %s.",
		"To go back to the previous menu, press %s.":               "Palaa edelliseen valikkoon painamalla %s.",
		"To disable voice notifications to this number, press %s.": "Poista puheilmoitukset tästä numerosta käytöstä painamalla %s.",
		"To repeat this message, press star.":                      "Toista tämä viesti painamalla tähteä.",
		"To acknowledge, press %s.":                                "Kuittaa painamalla %s.",
		"To escalate, press %s.":                                   "Eskaloi painamalla %s.",
		"To close, press %s.":                                      "Sulje painamalla %s.",
		"To acknowledge all, press %s.":                            "Kuittaa kaikki painamalla %s.",
		"To close all, press %s.":                                  "Sulje kaikki painamalla %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Jos olet valmis, voit vain sulkea puhelun.",
		"Sorry, I didn't understand that.":         "Anteeksi, en ymmärtänyt.",
		"Goodbye.":                                 "Näkemiin.",
		// call flow
		"Hello! This is %s":   "Hei! Täällä %s",
		"Hello! This is %s. ": "Hei! Täällä %s. ",
		"Please use the application dashboard to manage alerts.": "Hallinnoi hälytyksiä sovelluksen hallintapaneelista.",
		"Unenrolled.":        "Poistettu.",
		"One moment please.": "Hetki, kiitos.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Tapahtui virhe. Hallinnoi hälytyksiä hallintapaneelista.",
		"The menu options have changed. To acknowledge, press %s.":          "Valikon vaihtoehdot ovat muuttuneet. Kuittaa painamalla %s.",
		"The menu options have changed. To close, press %s.":                "Valikon vaihtoehdot ovat muuttuneet. Sulje painamalla %s.",
		// action confirmations
		"Acknowledged":                     "Kuitattu",
		"Acknowledged all alerts.":         "Kaikki hälytykset kuitattu.",
		"Closed":                           "Suljettu",
		"Closed all alerts.":               "Kaikki hälytykset suljettu.",
		"Escalation requested":             "Eskalointia pyydetty",
		"Escalation requested all alerts.": "Eskalointia pyydetty kaikille hälytyksille.",
		// error messages
		"Already %s":                                "Jo %s",
		"Alert is already closed.":                  "Hälytys on jo suljettu.",
		"Alert is already acknowledged.":            "Hälytys on jo kuitattu.",
		"Error: %s":                                 "Virhe: %s",
		"System error. Please visit the dashboard.": "Järjestelmävirhe. Käy hallintapaneelissa.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s hälytysilmoituksilla. Palvelulla ”%s” on %d kuittaamatonta hälytystä.",
		"%s with an alert notification. %s.":                                      "%s hälytysilmoituksella. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s tilapäivityksellä hälytykselle ”%s”. %s",
		"%s with a test message.":                                                 "%s testiviestillä.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s %d-numeroisella vahvistuskoodillasi. Koodi on: %s. Vielä kerran, %d-numeroinen vahvistuskoodisi on: %s.",
		"No summary provided": "Yhteenvetoa ei annettu",
	})
}
