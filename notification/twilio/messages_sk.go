package twilio

import "golang.org/x/text/language"

// Slovak voice translations. Registered under the base language tag so the
// regional variant (sk-SK) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Slovak, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Pre potvrdenie odhlásenia tohto čísla stlačte %s.",
		"To go back to the previous menu, press %s.":               "Pre návrat do predchádzajúcej ponuky stlačte %s.",
		"To disable voice notifications to this number, press %s.": "Pre vypnutie hlasových upozornení na toto číslo stlačte %s.",
		"To repeat this message, press star.":                      "Pre zopakovanie tejto správy stlačte hviezdičku.",
		"To acknowledge, press %s.":                                "Pre potvrdenie stlačte %s.",
		"To escalate, press %s.":                                   "Pre eskaláciu stlačte %s.",
		"To close, press %s.":                                      "Pre uzavretie stlačte %s.",
		"To acknowledge all, press %s.":                            "Pre potvrdenie všetkých stlačte %s.",
		"To close all, press %s.":                                  "Pre uzavretie všetkých stlačte %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Ak ste skončili, môžete jednoducho zavesiť.",
		"Sorry, I didn't understand that.":         "Prepáčte, nerozumel som tomu.",
		"Goodbye.":                                 "Dovidenia.",
		// call flow
		"Hello! This is %s":   "Dobrý deň! Tu je %s",
		"Hello! This is %s. ": "Dobrý deň! Tu je %s. ",
		"Please use the application dashboard to manage alerts.": "Na správu upozornení použite prosím nástenku aplikácie.",
		"Unenrolled.":        "Odhlásené.",
		"One moment please.": "Moment, prosím.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Vyskytla sa chyba. Na správu upozornení použite prosím nástenku.",
		"The menu options have changed. To acknowledge, press %s.":          "Možnosti ponuky sa zmenili. Pre potvrdenie stlačte %s.",
		"The menu options have changed. To close, press %s.":                "Možnosti ponuky sa zmenili. Pre uzavretie stlačte %s.",
		// action confirmations
		"Acknowledged":                     "Potvrdené",
		"Acknowledged all alerts.":         "Všetky upozornenia boli potvrdené.",
		"Closed":                           "Uzavreté",
		"Closed all alerts.":               "Všetky upozornenia boli uzavreté.",
		"Escalation requested":             "Eskalácia vyžiadaná",
		"Escalation requested all alerts.": "Eskalácia vyžiadaná pre všetky upozornenia.",
		// error messages
		"Already %s":                                "Už %s",
		"Alert is already closed.":                  "Upozornenie je už uzavreté.",
		"Alert is already acknowledged.":            "Upozornenie je už potvrdené.",
		"Error: %s":                                 "Chyba: %s",
		"System error. Please visit the dashboard.": "Systémová chyba. Navštívte prosím nástenku.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s s upozorneniami. Služba „%s“ má %d nepotvrdených upozornení.",
		"%s with an alert notification. %s.":                                      "%s s upozornením. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s s aktualizáciou stavu upozornenia „%s“. %s",
		"%s with a test message.":                                                 "%s s testovacou správou.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s s vaším %d-miestnym overovacím kódom. Kód je: %s. Ešte raz, váš %d-miestny overovací kód je: %s.",
		"No summary provided": "Žiadny súhrn nebol poskytnutý",
	})
}
