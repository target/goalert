package twilio

import "golang.org/x/text/language"

// Russian voice translations. Registered under the base language tag so every
// regional variant (ru-RU) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Russian, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Чтобы подтвердить отключение этого номера, нажмите %s.",
		"To go back to the previous menu, press %s.":               "Чтобы вернуться в предыдущее меню, нажмите %s.",
		"To disable voice notifications to this number, press %s.": "Чтобы отключить голосовые уведомления на этот номер, нажмите %s.",
		"To repeat this message, press star.":                      "Чтобы повторить это сообщение, нажмите звёздочку.",
		"To acknowledge, press %s.":                                "Чтобы принять, нажмите %s.",
		"To escalate, press %s.":                                   "Чтобы эскалировать, нажмите %s.",
		"To close, press %s.":                                      "Чтобы закрыть, нажмите %s.",
		"To acknowledge all, press %s.":                            "Чтобы принять все, нажмите %s.",
		"To close all, press %s.":                                  "Чтобы закрыть все, нажмите %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Если вы закончили, просто положите трубку.",
		"Sorry, I didn't understand that.":         "Извините, я не понял.",
		"Goodbye.":                                 "До свидания.",
		// call flow
		"Hello! This is %s":   "Здравствуйте! Это %s",
		"Hello! This is %s. ": "Здравствуйте! Это %s. ",
		"Please use the application dashboard to manage alerts.": "Пожалуйста, используйте панель управления приложения для управления оповещениями.",
		"Unenrolled.":        "Отключено.",
		"One moment please.": "Одну минуту, пожалуйста.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Произошла ошибка. Пожалуйста, используйте панель управления для управления оповещениями.",
		"The menu options have changed. To acknowledge, press %s.":          "Пункты меню изменились. Чтобы принять, нажмите %s.",
		"The menu options have changed. To close, press %s.":                "Пункты меню изменились. Чтобы закрыть, нажмите %s.",
		// action confirmations
		"Acknowledged":                     "Принято",
		"Acknowledged all alerts.":         "Все оповещения приняты.",
		"Closed":                           "Закрыто",
		"Closed all alerts.":               "Все оповещения закрыты.",
		"Escalation requested":             "Запрошена эскалация",
		"Escalation requested all alerts.": "Запрошена эскалация всех оповещений.",
		// error messages
		"Already %s":                                "Уже %s",
		"Alert is already closed.":                  "Оповещение уже закрыто.",
		"Alert is already acknowledged.":            "Оповещение уже принято.",
		"Error: %s":                                 "Ошибка: %s",
		"System error. Please visit the dashboard.": "Системная ошибка. Пожалуйста, перейдите в панель управления.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s с уведомлениями об оповещениях. У службы «%s» %d непринятых оповещений.",
		"%s with an alert notification. %s.":                                      "%s с уведомлением об оповещении. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s с обновлением статуса оповещения «%s». %s",
		"%s with a test message.":                                                 "%s с тестовым сообщением.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s с вашим %d-значным кодом подтверждения. Код: %s. Повторяю, ваш %d-значный код подтверждения: %s.",
		"No summary provided": "Сводка не предоставлена",
	})
}
