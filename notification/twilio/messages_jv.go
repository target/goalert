package twilio

import "golang.org/x/text/language"

// Javanese voice translations. Registered under the base language tag so the
// regional variant (jv-ID) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Make("jv"), map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Kanggo ngonfirmasi mbatalake registrasi nomer iki, pencet %s.",
		"To go back to the previous menu, press %s.":               "Kanggo bali menyang menu sadurunge, pencet %s.",
		"To disable voice notifications to this number, press %s.": "Kanggo mateni notifikasi swara menyang nomer iki, pencet %s.",
		"To repeat this message, press star.":                      "Kanggo mbaleni pesen iki, pencet bintang.",
		"To acknowledge, press %s.":                                "Kanggo ngakoni, pencet %s.",
		"To escalate, press %s.":                                   "Kanggo eskalasi, pencet %s.",
		"To close, press %s.":                                      "Kanggo nutup, pencet %s.",
		"To acknowledge all, press %s.":                            "Kanggo ngakoni kabeh, pencet %s.",
		"To close all, press %s.":                                  "Kanggo nutup kabeh, pencet %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Yen wis rampung, panjenengan bisa langsung nutup telpon.",
		"Sorry, I didn't understand that.":         "Nyuwun sewu, kula mboten mangertos.",
		"Goodbye.":                                 "Sugeng kondur.",
		// call flow
		"Hello! This is %s":   "Sugeng! Iki %s",
		"Hello! This is %s. ": "Sugeng! Iki %s. ",
		"Please use the application dashboard to manage alerts.": "Mangga ginakaken dashboard aplikasi kangge ngatur peringatan.",
		"Unenrolled.":        "Wis dibatalake.",
		"One moment please.": "Mangga ngentosi sekedhap.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Ana kesalahan. Mangga ginakaken dashboard kangge ngatur peringatan.",
		"The menu options have changed. To acknowledge, press %s.":          "Pilihan menu wis owah. Kanggo ngakoni, pencet %s.",
		"The menu options have changed. To close, press %s.":                "Pilihan menu wis owah. Kanggo nutup, pencet %s.",
		// action confirmations
		"Acknowledged":                     "Wis diakoni",
		"Acknowledged all alerts.":         "Kabeh peringatan wis diakoni.",
		"Closed":                           "Wis ditutup",
		"Closed all alerts.":               "Kabeh peringatan wis ditutup.",
		"Escalation requested":             "Eskalasi wis dijaluk",
		"Escalation requested all alerts.": "Eskalasi wis dijaluk kanggo kabeh peringatan.",
		// error messages
		"Already %s":                                "Wis %s",
		"Alert is already closed.":                  "Peringatan wis ditutup.",
		"Alert is already acknowledged.":            "Peringatan wis diakoni.",
		"Error: %s":                                 "Kesalahan: %s",
		"System error. Please visit the dashboard.": "Kesalahan sistem. Mangga bukak dashboard.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s kanthi notifikasi peringatan. Layanan '%s' nduwe %d peringatan sing durung diakoni.",
		"%s with an alert notification. %s.":                                      "%s kanthi notifikasi peringatan. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s kanthi nganyari status kanggo peringatan '%s'. %s",
		"%s with a test message.":                                                 "%s kanthi pesen tes.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s kanthi kode verifikasi %d angka. Kode kasebut: %s. Maleh, kode verifikasi %d angka panjenengan yaiku: %s.",
		"No summary provided": "Ora ana ringkesan",
	})
}
