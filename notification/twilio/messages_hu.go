package twilio

import "golang.org/x/text/language"

// Hungarian voice translations. Registered under the base language tag so the
// regional variant (hu-HU) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Hungarian, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "A szám leiratkozásának megerősítéséhez nyomja meg a %s gombot.",
		"To go back to the previous menu, press %s.":               "Az előző menübe való visszatéréshez nyomja meg a %s gombot.",
		"To disable voice notifications to this number, press %s.": "A hangértesítések kikapcsolásához ezen a számon nyomja meg a %s gombot.",
		"To repeat this message, press star.":                      "Az üzenet megismétléséhez nyomja meg a csillag gombot.",
		"To acknowledge, press %s.":                                "A nyugtázáshoz nyomja meg a %s gombot.",
		"To escalate, press %s.":                                   "Az eszkalációhoz nyomja meg a %s gombot.",
		"To close, press %s.":                                      "A lezáráshoz nyomja meg a %s gombot.",
		"To acknowledge all, press %s.":                            "Az összes nyugtázásához nyomja meg a %s gombot.",
		"To close all, press %s.":                                  "Az összes lezárásához nyomja meg a %s gombot.",
		// general prompts
		"If you are done, you may simply hang up.": "Ha végzett, egyszerűen tegye le a telefont.",
		"Sorry, I didn't understand that.":         "Sajnálom, ezt nem értettem.",
		"Goodbye.":                                 "Viszonthallásra.",
		// call flow
		"Hello! This is %s":   "Üdvözöljük! Itt a %s",
		"Hello! This is %s. ": "Üdvözöljük! Itt a %s. ",
		"Please use the application dashboard to manage alerts.": "Kérjük, az alkalmazás vezérlőpultját használja a riasztások kezeléséhez.",
		"Unenrolled.":        "Leiratkozva.",
		"One moment please.": "Egy pillanat, kérem.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Hiba történt. Kérjük, a vezérlőpultot használja a riasztások kezeléséhez.",
		"The menu options have changed. To acknowledge, press %s.":          "A menüpontok megváltoztak. A nyugtázáshoz nyomja meg a %s gombot.",
		"The menu options have changed. To close, press %s.":                "A menüpontok megváltoztak. A lezáráshoz nyomja meg a %s gombot.",
		// action confirmations
		"Acknowledged":                     "Nyugtázva",
		"Acknowledged all alerts.":         "Minden riasztás nyugtázva.",
		"Closed":                           "Lezárva",
		"Closed all alerts.":               "Minden riasztás lezárva.",
		"Escalation requested":             "Eszkaláció kérve",
		"Escalation requested all alerts.": "Eszkaláció kérve minden riasztáshoz.",
		// error messages
		"Already %s":                                "Már %s",
		"Alert is already closed.":                  "A riasztás már le van zárva.",
		"Alert is already acknowledged.":            "A riasztás már nyugtázva van.",
		"Error: %s":                                 "Hiba: %s",
		"System error. Please visit the dashboard.": "Rendszerhiba. Kérjük, látogasson el a vezérlőpultra.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s riasztási értesítésekkel. A(z) „%s” szolgáltatásnak %d nyugtázatlan riasztása van.",
		"%s with an alert notification. %s.":                                      "%s egy riasztási értesítéssel. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s állapotfrissítéssel a(z) „%s” riasztáshoz. %s",
		"%s with a test message.":                                                 "%s egy tesztüzenettel.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s a(z) %d számjegyű ellenőrző kódjával. A kód: %s. Ismétlem, a(z) %d számjegyű ellenőrző kódja: %s.",
		"No summary provided": "Nincs megadva összefoglaló",
	})
}
