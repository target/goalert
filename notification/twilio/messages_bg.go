package twilio

import "golang.org/x/text/language"

// Bulgarian voice translations. Registered under the base language tag so the
// regional variant (bg-BG) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Bulgarian, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "За да потвърдите премахването на този номер, натиснете %s.",
		"To go back to the previous menu, press %s.":               "За да се върнете към предишното меню, натиснете %s.",
		"To disable voice notifications to this number, press %s.": "За да изключите гласовите известия за този номер, натиснете %s.",
		"To repeat this message, press star.":                      "За да повторите това съобщение, натиснете звездичка.",
		"To acknowledge, press %s.":                                "За да потвърдите, натиснете %s.",
		"To escalate, press %s.":                                   "За да ескалирате, натиснете %s.",
		"To close, press %s.":                                      "За да затворите, натиснете %s.",
		"To acknowledge all, press %s.":                            "За да потвърдите всички, натиснете %s.",
		"To close all, press %s.":                                  "За да затворите всички, натиснете %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Ако сте готови, можете просто да затворите.",
		"Sorry, I didn't understand that.":         "Съжалявам, не разбрах.",
		"Goodbye.":                                 "Довиждане.",
		// call flow
		"Hello! This is %s":   "Здравейте! Това е %s",
		"Hello! This is %s. ": "Здравейте! Това е %s. ",
		"Please use the application dashboard to manage alerts.": "Моля, използвайте таблото на приложението, за да управлявате сигналите.",
		"Unenrolled.":        "Премахнато.",
		"One moment please.": "Един момент, моля.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Възникна грешка. Моля, използвайте таблото, за да управлявате сигналите.",
		"The menu options have changed. To acknowledge, press %s.":          "Опциите в менюто се промениха. За да потвърдите, натиснете %s.",
		"The menu options have changed. To close, press %s.":                "Опциите в менюто се промениха. За да затворите, натиснете %s.",
		// action confirmations
		"Acknowledged":                     "Потвърдено",
		"Acknowledged all alerts.":         "Всички сигнали са потвърдени.",
		"Closed":                           "Затворено",
		"Closed all alerts.":               "Всички сигнали са затворени.",
		"Escalation requested":             "Заявена ескалация",
		"Escalation requested all alerts.": "Заявена ескалация за всички сигнали.",
		// error messages
		"Already %s":                                "Вече %s",
		"Alert is already closed.":                  "Сигналът вече е затворен.",
		"Alert is already acknowledged.":            "Сигналът вече е потвърден.",
		"Error: %s":                                 "Грешка: %s",
		"System error. Please visit the dashboard.": "Системна грешка. Моля, посетете таблото.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s с известия за сигнали. Услугата „%s“ има %d непотвърдени сигнала.",
		"%s with an alert notification. %s.":                                      "%s с известие за сигнал. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s с актуализация на състоянието за сигнал „%s“. %s",
		"%s with a test message.":                                                 "%s с тестово съобщение.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s с вашия %d-цифрен код за потвърждение. Кодът е: %s. Отново, вашият %d-цифрен код за потвърждение е: %s.",
		"No summary provided": "Няма предоставено резюме",
	})
}
