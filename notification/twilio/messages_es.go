package twilio

import "golang.org/x/text/language"

// Spanish voice translations. Registered under the base language tag so every
// regional variant (es-AR, es-CL, es-CO, es-ES, es-MX, es-PE, es-US, es-VE)
// resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Spanish, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Para confirmar la baja de este número, presione %s.",
		"To go back to the previous menu, press %s.":               "Para volver al menú anterior, presione %s.",
		"To disable voice notifications to this number, press %s.": "Para desactivar las notificaciones de voz a este número, presione %s.",
		"To repeat this message, press star.":                      "Para repetir este mensaje, presione asterisco.",
		"To acknowledge, press %s.":                                "Para reconocer, presione %s.",
		"To escalate, press %s.":                                   "Para escalar, presione %s.",
		"To close, press %s.":                                      "Para cerrar, presione %s.",
		"To acknowledge all, press %s.":                            "Para reconocer todas, presione %s.",
		"To close all, press %s.":                                  "Para cerrar todas, presione %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Si ha terminado, puede simplemente colgar.",
		"Sorry, I didn't understand that.":         "Lo siento, no entendí eso.",
		"Goodbye.":                                 "Adiós.",
		// call flow
		"Hello! This is %s":   "¡Hola! Le habla %s",
		"Hello! This is %s. ": "¡Hola! Le habla %s. ",
		"Please use the application dashboard to manage alerts.": "Por favor, utilice el panel de la aplicación para gestionar las alertas.",
		"Unenrolled.":        "Dado de baja.",
		"One moment please.": "Un momento, por favor.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Ha ocurrido un error. Por favor, utilice el panel para gestionar las alertas.",
		"The menu options have changed. To acknowledge, press %s.":          "Las opciones del menú han cambiado. Para reconocer, presione %s.",
		"The menu options have changed. To close, press %s.":                "Las opciones del menú han cambiado. Para cerrar, presione %s.",
		// action confirmations
		"Acknowledged":                     "Reconocida",
		"Acknowledged all alerts.":         "Se reconocieron todas las alertas.",
		"Closed":                           "Cerrada",
		"Closed all alerts.":               "Se cerraron todas las alertas.",
		"Escalation requested":             "Escalada solicitada",
		"Escalation requested all alerts.": "Se solicitó la escalada de todas las alertas.",
		// error messages
		"Already %s":                                "Ya %s",
		"Alert is already closed.":                  "La alerta ya está cerrada.",
		"Alert is already acknowledged.":            "La alerta ya está reconocida.",
		"Error: %s":                                 "Error: %s",
		"System error. Please visit the dashboard.": "Error del sistema. Por favor, consulte el panel.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s con notificaciones de alerta. El servicio «%s» tiene %d alertas sin reconocer.",
		"%s with an alert notification. %s.":                                      "%s con una notificación de alerta. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s con una actualización de estado para la alerta «%s». %s",
		"%s with a test message.":                                                 "%s con un mensaje de prueba.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s con su código de verificación de %d dígitos. El código es: %s. De nuevo, su código de verificación de %d dígitos es: %s.",
		"No summary provided": "No se proporcionó ningún resumen",
	})
}
