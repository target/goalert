package twilio

import "golang.org/x/text/language"

// Portuguese voice translations. Registered under the base language tag so every
// regional variant (pt-BR, pt-PT) resolves to it. The wording is kept neutral so
// it is intelligible in both Brazil and Portugal.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Portuguese, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Para confirmar o cancelamento do registro deste número, pressione %s.",
		"To go back to the previous menu, press %s.":               "Para voltar ao menu anterior, pressione %s.",
		"To disable voice notifications to this number, press %s.": "Para desativar as notificações de voz neste número, pressione %s.",
		"To repeat this message, press star.":                      "Para repetir esta mensagem, pressione asterisco.",
		"To acknowledge, press %s.":                                "Para confirmar, pressione %s.",
		"To escalate, press %s.":                                   "Para escalar, pressione %s.",
		"To close, press %s.":                                      "Para encerrar, pressione %s.",
		"To acknowledge all, press %s.":                            "Para confirmar tudo, pressione %s.",
		"To close all, press %s.":                                  "Para encerrar tudo, pressione %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Se você terminou, basta desligar.",
		"Sorry, I didn't understand that.":         "Desculpe, não entendi.",
		"Goodbye.":                                 "Até logo.",
		// call flow
		"Hello! This is %s":   "Olá! Aqui é %s",
		"Hello! This is %s. ": "Olá! Aqui é %s. ",
		"Please use the application dashboard to manage alerts.": "Por favor, use o painel do aplicativo para gerenciar os alertas.",
		"Unenrolled.":        "Registro cancelado.",
		"One moment please.": "Um momento, por favor.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Ocorreu um erro. Por favor, use o painel para gerenciar os alertas.",
		"The menu options have changed. To acknowledge, press %s.":          "As opções do menu mudaram. Para confirmar, pressione %s.",
		"The menu options have changed. To close, press %s.":                "As opções do menu mudaram. Para encerrar, pressione %s.",
		// action confirmations
		"Acknowledged":                     "Confirmado",
		"Acknowledged all alerts.":         "Todos os alertas foram confirmados.",
		"Closed":                           "Encerrado",
		"Closed all alerts.":               "Todos os alertas foram encerrados.",
		"Escalation requested":             "Escalonamento solicitado",
		"Escalation requested all alerts.": "Escalonamento solicitado para todos os alertas.",
		// error messages
		"Already %s":                                "Já %s",
		"Alert is already closed.":                  "O alerta já está encerrado.",
		"Alert is already acknowledged.":            "O alerta já está confirmado.",
		"Error: %s":                                 "Erro: %s",
		"System error. Please visit the dashboard.": "Erro de sistema. Por favor, acesse o painel.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s com notificações de alerta. O serviço \"%s\" tem %d alertas não confirmados.",
		"%s with an alert notification. %s.":                                      "%s com uma notificação de alerta. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s com uma atualização de status para o alerta \"%s\". %s",
		"%s with a test message.":                                                 "%s com uma mensagem de teste.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s com o seu código de verificação de %d dígitos. O código é: %s. Novamente, o seu código de verificação de %d dígitos é: %s.",
		"No summary provided": "Nenhum resumo fornecido",
	})
}
