package twilio

import "golang.org/x/text/language"

// Hindi voice translations. Registered under the base language tag so the
// regional variant (hi-IN) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Hindi, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "इस नंबर को हटाने की पुष्टि के लिए, %s दबाएँ।",
		"To go back to the previous menu, press %s.":               "पिछले मेन्यू पर वापस जाने के लिए, %s दबाएँ।",
		"To disable voice notifications to this number, press %s.": "इस नंबर पर वॉइस सूचनाएँ बंद करने के लिए, %s दबाएँ।",
		"To repeat this message, press star.":                      "इस संदेश को दोहराने के लिए, तारा दबाएँ।",
		"To acknowledge, press %s.":                                "स्वीकार करने के लिए, %s दबाएँ।",
		"To escalate, press %s.":                                   "आगे बढ़ाने के लिए, %s दबाएँ।",
		"To close, press %s.":                                      "बंद करने के लिए, %s दबाएँ।",
		"To acknowledge all, press %s.":                            "सभी को स्वीकार करने के लिए, %s दबाएँ।",
		"To close all, press %s.":                                  "सभी को बंद करने के लिए, %s दबाएँ।",
		// general prompts
		"If you are done, you may simply hang up.": "यदि आपका काम हो गया है, तो आप फ़ोन रख सकते हैं।",
		"Sorry, I didn't understand that.":         "क्षमा करें, मैं समझ नहीं पाया।",
		"Goodbye.":                                 "नमस्ते।",
		// call flow
		"Hello! This is %s":   "नमस्ते! यह %s है",
		"Hello! This is %s. ": "नमस्ते! यह %s है। ",
		"Please use the application dashboard to manage alerts.": "कृपया अलर्ट प्रबंधित करने के लिए एप्लिकेशन डैशबोर्ड का उपयोग करें।",
		"Unenrolled.":        "हटा दिया गया।",
		"One moment please.": "कृपया एक क्षण प्रतीक्षा करें।",
		"An error has occurred. Please use the dashboard to manage alerts.": "एक त्रुटि हुई है। कृपया अलर्ट प्रबंधित करने के लिए डैशबोर्ड का उपयोग करें।",
		"The menu options have changed. To acknowledge, press %s.":          "मेन्यू विकल्प बदल गए हैं। स्वीकार करने के लिए, %s दबाएँ।",
		"The menu options have changed. To close, press %s.":                "मेन्यू विकल्प बदल गए हैं। बंद करने के लिए, %s दबाएँ।",
		// action confirmations
		"Acknowledged":                     "स्वीकार किया गया",
		"Acknowledged all alerts.":         "सभी अलर्ट स्वीकार किए गए।",
		"Closed":                           "बंद किया गया",
		"Closed all alerts.":               "सभी अलर्ट बंद किए गए।",
		"Escalation requested":             "आगे बढ़ाने का अनुरोध किया गया",
		"Escalation requested all alerts.": "सभी अलर्ट के लिए आगे बढ़ाने का अनुरोध किया गया।",
		// error messages
		"Already %s":                                "पहले से ही %s",
		"Alert is already closed.":                  "अलर्ट पहले से ही बंद है।",
		"Alert is already acknowledged.":            "अलर्ट पहले से ही स्वीकार किया जा चुका है।",
		"Error: %s":                                 "त्रुटि: %s",
		"System error. Please visit the dashboard.": "सिस्टम त्रुटि। कृपया डैशबोर्ड देखें।",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s, अलर्ट सूचनाओं के साथ। सेवा '%s' में %d अस्वीकृत अलर्ट हैं।",
		"%s with an alert notification. %s.":                                      "%s, एक अलर्ट सूचना के साथ। %s।",
		"%s with a status update for alert '%s'. %s":                              "%s, अलर्ट '%s' के लिए स्थिति अपडेट के साथ। %s",
		"%s with a test message.":                                                 "%s, एक परीक्षण संदेश के साथ।",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s, आपके %d अंकों के सत्यापन कोड के साथ। कोड है: %s। फिर से, आपका %d अंकों का सत्यापन कोड है: %s।",
		"No summary provided": "कोई सारांश नहीं दिया गया",
	})
}
