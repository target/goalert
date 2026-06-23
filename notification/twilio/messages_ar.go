package twilio

import "golang.org/x/text/language"

// Arabic voice translations. Registered under the base language tag so every
// regional variant (ar-EG, ar-KW, ar-LB, ar-MA, ar-SA) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Arabic, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "لتأكيد إلغاء تسجيل هذا الرقم، اضغط %s.",
		"To go back to the previous menu, press %s.":               "للعودة إلى القائمة السابقة، اضغط %s.",
		"To disable voice notifications to this number, press %s.": "لتعطيل الإشعارات الصوتية إلى هذا الرقم، اضغط %s.",
		"To repeat this message, press star.":                      "لتكرار هذه الرسالة، اضغط على النجمة.",
		"To acknowledge, press %s.":                                "للإقرار، اضغط %s.",
		"To escalate, press %s.":                                   "للتصعيد، اضغط %s.",
		"To close, press %s.":                                      "للإغلاق، اضغط %s.",
		"To acknowledge all, press %s.":                            "للإقرار بالكل، اضغط %s.",
		"To close all, press %s.":                                  "لإغلاق الكل، اضغط %s.",
		// general prompts
		"If you are done, you may simply hang up.": "إذا انتهيت، يمكنك ببساطة إنهاء المكالمة.",
		"Sorry, I didn't understand that.":         "عذرًا، لم أفهم ذلك.",
		"Goodbye.":                                 "مع السلامة.",
		// call flow
		"Hello! This is %s":   "مرحبًا! هذه %s",
		"Hello! This is %s. ": "مرحبًا! هذه %s. ",
		"Please use the application dashboard to manage alerts.": "يرجى استخدام لوحة تحكم التطبيق لإدارة التنبيهات.",
		"Unenrolled.":        "تم إلغاء التسجيل.",
		"One moment please.": "لحظة من فضلك.",
		"An error has occurred. Please use the dashboard to manage alerts.": "حدث خطأ. يرجى استخدام لوحة التحكم لإدارة التنبيهات.",
		"The menu options have changed. To acknowledge, press %s.":          "تغيرت خيارات القائمة. للإقرار، اضغط %s.",
		"The menu options have changed. To close, press %s.":                "تغيرت خيارات القائمة. للإغلاق، اضغط %s.",
		// action confirmations
		"Acknowledged":                     "تم الإقرار",
		"Acknowledged all alerts.":         "تم الإقرار بجميع التنبيهات.",
		"Closed":                           "تم الإغلاق",
		"Closed all alerts.":               "تم إغلاق جميع التنبيهات.",
		"Escalation requested":             "تم طلب التصعيد",
		"Escalation requested all alerts.": "تم طلب التصعيد لجميع التنبيهات.",
		// error messages
		"Already %s":                                "بالفعل %s",
		"Alert is already closed.":                  "التنبيه مغلق بالفعل.",
		"Alert is already acknowledged.":            "تم الإقرار بالتنبيه بالفعل.",
		"Error: %s":                                 "خطأ: %s",
		"System error. Please visit the dashboard.": "خطأ في النظام. يرجى زيارة لوحة التحكم.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s مع إشعارات التنبيه. الخدمة '%s' لديها %d تنبيهات لم يتم الإقرار بها.",
		"%s with an alert notification. %s.":                                      "%s مع إشعار تنبيه. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s مع تحديث حالة للتنبيه '%s'. %s",
		"%s with a test message.":                                                 "%s مع رسالة اختبار.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s مع رمز التحقق المكون من %d أرقام. الرمز هو: %s. مرة أخرى، رمز التحقق المكون من %d أرقام هو: %s.",
		"No summary provided": "لم يتم تقديم ملخص",
	})
}
