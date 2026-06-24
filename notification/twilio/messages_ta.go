package twilio

import "golang.org/x/text/language"

// Tamil voice translations. Registered under the base language tag so every
// regional variant (ta-IN) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Tamil, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "இந்த எண்ணை நீக்குவதை உறுதிப்படுத்த, %s ஐ அழுத்தவும்.",
		"To go back to the previous menu, press %s.":               "முந்தைய மெனுவிற்குத் திரும்ப, %s ஐ அழுத்தவும்.",
		"To disable voice notifications to this number, press %s.": "இந்த எண்ணுக்கான குரல் அறிவிப்புகளை முடக்க, %s ஐ அழுத்தவும்.",
		"To repeat this message, press star.":                      "இந்தச் செய்தியை மீண்டும் கேட்க, நட்சத்திரத்தை அழுத்தவும்.",
		"To acknowledge, press %s.":                                "ஒப்புக்கொள்ள, %s ஐ அழுத்தவும்.",
		"To escalate, press %s.":                                   "மேலேற்ற, %s ஐ அழுத்தவும்.",
		"To close, press %s.":                                      "மூட, %s ஐ அழுத்தவும்.",
		"To acknowledge all, press %s.":                            "அனைத்தையும் ஒப்புக்கொள்ள, %s ஐ அழுத்தவும்.",
		"To close all, press %s.":                                  "அனைத்தையும் மூட, %s ஐ அழுத்தவும்.",
		// general prompts
		"If you are done, you may simply hang up.": "நீங்கள் முடித்துவிட்டால், தொலைபேசியை வைத்துவிடலாம்.",
		"Sorry, I didn't understand that.":         "மன்னிக்கவும், எனக்குப் புரியவில்லை.",
		"Goodbye.":                                 "நன்றி, வணக்கம்.",
		// call flow
		"Hello! This is %s":   "வணக்கம்! இது %s",
		"Hello! This is %s. ": "வணக்கம்! இது %s. ",
		"Please use the application dashboard to manage alerts.": "எச்சரிக்கைகளை நிர்வகிக்க, பயன்பாட்டின் டாஷ்போர்டைப் பயன்படுத்தவும்.",
		"Unenrolled.":        "நீக்கப்பட்டது.",
		"One moment please.": "ஒரு நிமிடம், தயவுசெய்து காத்திருங்கள்.",
		"An error has occurred. Please use the dashboard to manage alerts.": "ஒரு பிழை ஏற்பட்டது. எச்சரிக்கைகளை நிர்வகிக்க டாஷ்போர்டைப் பயன்படுத்தவும்.",
		"The menu options have changed. To acknowledge, press %s.":          "மெனு விருப்பங்கள் மாறிவிட்டன. ஒப்புக்கொள்ள, %s ஐ அழுத்தவும்.",
		"The menu options have changed. To close, press %s.":                "மெனு விருப்பங்கள் மாறிவிட்டன. மூட, %s ஐ அழுத்தவும்.",
		// action confirmations
		"Acknowledged":                     "ஒப்புக்கொள்ளப்பட்டது",
		"Acknowledged all alerts.":         "அனைத்து எச்சரிக்கைகளும் ஒப்புக்கொள்ளப்பட்டன.",
		"Closed":                           "மூடப்பட்டது",
		"Closed all alerts.":               "அனைத்து எச்சரிக்கைகளும் மூடப்பட்டன.",
		"Escalation requested":             "மேலேற்றம் கோரப்பட்டது",
		"Escalation requested all alerts.": "அனைத்து எச்சரிக்கைகளுக்கும் மேலேற்றம் கோரப்பட்டது.",
		// error messages
		"Already %s":                                "ஏற்கனவே %s",
		"Alert is already closed.":                  "எச்சரிக்கை ஏற்கனவே மூடப்பட்டுள்ளது.",
		"Alert is already acknowledged.":            "எச்சரிக்கை ஏற்கனவே ஒப்புக்கொள்ளப்பட்டுள்ளது.",
		"Error: %s":                                 "பிழை: %s",
		"System error. Please visit the dashboard.": "கணினி பிழை. தயவுசெய்து டாஷ்போர்டைப் பார்வையிடவும்.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s எச்சரிக்கை அறிவிப்புகளுடன். '%s' சேவையில் %d ஒப்புக்கொள்ளப்படாத எச்சரிக்கைகள் உள்ளன.",
		"%s with an alert notification. %s.":                                      "%s ஒரு எச்சரிக்கை அறிவிப்புடன். %s.",
		"%s with a status update for alert '%s'. %s":                              "%s '%s' எச்சரிக்கைக்கான நிலை மேம்படுத்தலுடன். %s",
		"%s with a test message.":                                                 "%s ஒரு சோதனை செய்தியுடன்.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s உங்கள் %d இலக்க சரிபார்ப்புக் குறியீட்டுடன். குறியீடு: %s. மீண்டும், உங்கள் %d இலக்க சரிபார்ப்புக் குறியீடு: %s.",
		"No summary provided": "சுருக்கம் எதுவும் வழங்கப்படவில்லை",
	})
}
