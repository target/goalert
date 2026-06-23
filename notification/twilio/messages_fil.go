package twilio

import "golang.org/x/text/language"

// Filipino (Tagalog) voice translations. Registered under the base language tag
// so the regional variant (fil-PH) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Filipino, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Upang kumpirmahin ang pag-alis ng numerong ito, pindutin ang %s.",
		"To go back to the previous menu, press %s.":               "Upang bumalik sa nakaraang menu, pindutin ang %s.",
		"To disable voice notifications to this number, press %s.": "Upang i-off ang mga voice notification sa numerong ito, pindutin ang %s.",
		"To repeat this message, press star.":                      "Upang ulitin ang mensaheng ito, pindutin ang star.",
		"To acknowledge, press %s.":                                "Upang kilalanin, pindutin ang %s.",
		"To escalate, press %s.":                                   "Upang i-escalate, pindutin ang %s.",
		"To close, press %s.":                                      "Upang isara, pindutin ang %s.",
		"To acknowledge all, press %s.":                            "Upang kilalanin lahat, pindutin ang %s.",
		"To close all, press %s.":                                  "Upang isara lahat, pindutin ang %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Kung tapos na kayo, maaari na kayong magbaba ng telepono.",
		"Sorry, I didn't understand that.":         "Paumanhin, hindi ko naintindihan iyon.",
		"Goodbye.":                                 "Paalam.",
		// call flow
		"Hello! This is %s":   "Kumusta! Ito ay %s",
		"Hello! This is %s. ": "Kumusta! Ito ay %s. ",
		"Please use the application dashboard to manage alerts.": "Mangyaring gamitin ang dashboard ng application upang pamahalaan ang mga alerto.",
		"Unenrolled.":        "Naalis na.",
		"One moment please.": "Sandali lamang po.",
		"An error has occurred. Please use the dashboard to manage alerts.": "May naganap na error. Mangyaring gamitin ang dashboard upang pamahalaan ang mga alerto.",
		"The menu options have changed. To acknowledge, press %s.":          "Nagbago ang mga opsyon sa menu. Upang kilalanin, pindutin ang %s.",
		"The menu options have changed. To close, press %s.":                "Nagbago ang mga opsyon sa menu. Upang isara, pindutin ang %s.",
		// action confirmations
		"Acknowledged":                     "Kinilala",
		"Acknowledged all alerts.":         "Kinilala lahat ng alerto.",
		"Closed":                           "Isinara",
		"Closed all alerts.":               "Isinara lahat ng alerto.",
		"Escalation requested":             "Hiniling ang escalation",
		"Escalation requested all alerts.": "Hiniling ang escalation para sa lahat ng alerto.",
		// error messages
		"Already %s":                                "%s na",
		"Alert is already closed.":                  "Naisara na ang alerto.",
		"Alert is already acknowledged.":            "Nakilala na ang alerto.",
		"Error: %s":                                 "Error: %s",
		"System error. Please visit the dashboard.": "System error. Mangyaring bisitahin ang dashboard.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s na may mga alert notification. Ang serbisyong '%s' ay may %d na hindi pa kinilalang alerto.",
		"%s with an alert notification. %s.":                                      "%s na may alert notification. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s na may status update para sa alertong '%s'. %s",
		"%s with a test message.":                                                 "%s na may test message.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s na may inyong %d-digit na verification code. Ang code ay: %s. Muli, ang inyong %d-digit na verification code ay: %s.",
		"No summary provided": "Walang ibinigay na buod",
	})
}
