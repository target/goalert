package twilio

import "golang.org/x/text/language"

// Cebuano voice translations. Registered under the base language tag so the
// regional variant (ceb-PH) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Make("ceb"), map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Aron makumpirma ang pagtangtang niini nga numero, pislita ang %s.",
		"To go back to the previous menu, press %s.":               "Aron mobalik sa miaging menu, pislita ang %s.",
		"To disable voice notifications to this number, press %s.": "Aron i-disable ang mga voice notification niini nga numero, pislita ang %s.",
		"To repeat this message, press star.":                      "Aron usbon kini nga mensahe, pislita ang star.",
		"To acknowledge, press %s.":                                "Aron i-acknowledge, pislita ang %s.",
		"To escalate, press %s.":                                   "Aron i-escalate, pislita ang %s.",
		"To close, press %s.":                                      "Aron isira, pislita ang %s.",
		"To acknowledge all, press %s.":                            "Aron i-acknowledge ang tanan, pislita ang %s.",
		"To close all, press %s.":                                  "Aron isira ang tanan, pislita ang %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Kung nahuman ka na, mahimo ka lang nga mag-hang up.",
		"Sorry, I didn't understand that.":         "Pasayloa ko, wala ko makasabot niana.",
		"Goodbye.":                                 "Babay.",
		// call flow
		"Hello! This is %s":   "Maayong adlaw! Kini si %s",
		"Hello! This is %s. ": "Maayong adlaw! Kini si %s. ",
		"Please use the application dashboard to manage alerts.": "Palihug gamita ang dashboard sa aplikasyon aron madumala ang mga alerto.",
		"Unenrolled.":        "Natangtang na.",
		"One moment please.": "Palihug, kadiyot lang.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Adunay sayop nga nahitabo. Palihug gamita ang dashboard aron madumala ang mga alerto.",
		"The menu options have changed. To acknowledge, press %s.":          "Nausab ang mga opsyon sa menu. Aron i-acknowledge, pislita ang %s.",
		"The menu options have changed. To close, press %s.":                "Nausab ang mga opsyon sa menu. Aron isira, pislita ang %s.",
		// action confirmations
		"Acknowledged":                     "Na-acknowledge",
		"Acknowledged all alerts.":         "Na-acknowledge ang tanan nga alerto.",
		"Closed":                           "Sirado",
		"Closed all alerts.":               "Gisira ang tanan nga alerto.",
		"Escalation requested":             "Gihangyo ang escalation",
		"Escalation requested all alerts.": "Gihangyo ang escalation sa tanan nga alerto.",
		// error messages
		"Already %s":                                "%s na",
		"Alert is already closed.":                  "Sirado na ang alerto.",
		"Alert is already acknowledged.":            "Na-acknowledge na ang alerto.",
		"Error: %s":                                 "Sayop: %s",
		"System error. Please visit the dashboard.": "Sayop sa sistema. Palihug bisitaha ang dashboard.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s nga adunay mga notipikasyon sa alerto. Ang serbisyo nga '%s' adunay %d ka alerto nga wala pa ma-acknowledge.",
		"%s with an alert notification. %s.":                                      "%s nga adunay notipikasyon sa alerto. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s nga adunay update sa kahimtang sa alerto nga '%s'. %s",
		"%s with a test message.":                                                 "%s nga adunay test nga mensahe.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s nga adunay imong %d ka digit nga verification code. Ang code mao ang: %s. Usab, ang imong %d ka digit nga verification code mao ang: %s.",
		"No summary provided": "Walay summary nga gihatag",
	})
}
