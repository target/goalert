package twilio

import "golang.org/x/text/language"

// Ukrainian voice translations. Registered under the base language tag so every
// regional variant (uk-UA) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Ukrainian, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Щоб підтвердити скасування реєстрації цього номера, натисніть %s.",
		"To go back to the previous menu, press %s.":               "Щоб повернутися до попереднього меню, натисніть %s.",
		"To disable voice notifications to this number, press %s.": "Щоб вимкнути голосові сповіщення на цей номер, натисніть %s.",
		"To repeat this message, press star.":                      "Щоб повторити це повідомлення, натисніть зірочку.",
		"To acknowledge, press %s.":                                "Щоб підтвердити, натисніть %s.",
		"To escalate, press %s.":                                   "Щоб ескалувати, натисніть %s.",
		"To close, press %s.":                                      "Щоб закрити, натисніть %s.",
		"To acknowledge all, press %s.":                            "Щоб підтвердити всі, натисніть %s.",
		"To close all, press %s.":                                  "Щоб закрити всі, натисніть %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Якщо ви завершили, можете просто покласти слухавку.",
		"Sorry, I didn't understand that.":         "Вибачте, я не зрозумів.",
		"Goodbye.":                                 "До побачення.",
		// call flow
		"Hello! This is %s":   "Вітаю! Це %s",
		"Hello! This is %s. ": "Вітаю! Це %s. ",
		"Please use the application dashboard to manage alerts.": "Будь ласка, використовуйте панель керування додатком для керування сповіщеннями.",
		"Unenrolled.":        "Реєстрацію скасовано.",
		"One moment please.": "Зачекайте, будь ласка.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Сталася помилка. Будь ласка, використовуйте панель керування для керування сповіщеннями.",
		"The menu options have changed. To acknowledge, press %s.":          "Параметри меню змінилися. Щоб підтвердити, натисніть %s.",
		"The menu options have changed. To close, press %s.":                "Параметри меню змінилися. Щоб закрити, натисніть %s.",
		// action confirmations
		"Acknowledged":                     "Підтверджено",
		"Acknowledged all alerts.":         "Усі сповіщення підтверджено.",
		"Closed":                           "Закрито",
		"Closed all alerts.":               "Усі сповіщення закрито.",
		"Escalation requested":             "Запитано ескалацію",
		"Escalation requested all alerts.": "Запитано ескалацію для всіх сповіщень.",
		// error messages
		"Already %s":                                "Вже %s",
		"Alert is already closed.":                  "Сповіщення вже закрито.",
		"Alert is already acknowledged.":            "Сповіщення вже підтверджено.",
		"Error: %s":                                 "Помилка: %s",
		"System error. Please visit the dashboard.": "Системна помилка. Будь ласка, відвідайте панель керування.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s зі сповіщеннями. Служба «%s» має %d непідтверджених сповіщень.",
		"%s with an alert notification. %s.":                                      "%s зі сповіщенням. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s з оновленням статусу для сповіщення «%s». %s",
		"%s with a test message.":                                                 "%s з тестовим повідомленням.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s з вашим %d-значним кодом підтвердження. Код: %s. Ще раз, ваш %d-значний код підтвердження: %s.",
		"No summary provided": "Опис не надано",
	})
}
