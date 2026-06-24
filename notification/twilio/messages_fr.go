package twilio

import "golang.org/x/text/language"

// French voice translations. Registered under the base language tag so every
// regional variant (fr-FR, fr-CA, fr-BE, fr-CH, fr-TN) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.French, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Pour confirmer la désinscription de ce numéro, tapez %s.",
		"To go back to the previous menu, press %s.":               "Pour revenir au menu précédent, tapez %s.",
		"To disable voice notifications to this number, press %s.": "Pour désactiver les notifications vocales sur ce numéro, tapez %s.",
		"To repeat this message, press star.":                      "Pour répéter ce message, tapez étoile.",
		"To acknowledge, press %s.":                                "Pour acquitter, tapez %s.",
		"To escalate, press %s.":                                   "Pour escalader, tapez %s.",
		"To close, press %s.":                                      "Pour clôturer, tapez %s.",
		"To acknowledge all, press %s.":                            "Pour tout acquitter, tapez %s.",
		"To close all, press %s.":                                  "Pour tout clôturer, tapez %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Si vous avez terminé, vous pouvez simplement raccrocher.",
		"Sorry, I didn't understand that.":         "Désolé, je n'ai pas compris.",
		"Goodbye.":                                 "Au revoir.",
		// call flow
		"Hello! This is %s":   "Bonjour ! Ici %s",
		"Hello! This is %s. ": "Bonjour ! Ici %s. ",
		"Please use the application dashboard to manage alerts.": "Veuillez utiliser le tableau de bord de l'application pour gérer les alertes.",
		"Unenrolled.":        "Désinscrit.",
		"One moment please.": "Un instant, s'il vous plaît.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Une erreur est survenue. Veuillez utiliser le tableau de bord pour gérer les alertes.",
		"The menu options have changed. To acknowledge, press %s.":          "Les options du menu ont changé. Pour acquitter, tapez %s.",
		"The menu options have changed. To close, press %s.":                "Les options du menu ont changé. Pour clôturer, tapez %s.",
		// action confirmations
		"Acknowledged":                     "Acquittée",
		"Acknowledged all alerts.":         "Toutes les alertes ont été acquittées.",
		"Closed":                           "Clôturée",
		"Closed all alerts.":               "Toutes les alertes ont été clôturées.",
		"Escalation requested":             "Escalade demandée",
		"Escalation requested all alerts.": "Escalade demandée pour toutes les alertes.",
		// error messages
		"Already %s":                                "Déjà %s",
		"Alert is already closed.":                  "L'alerte est déjà clôturée.",
		"Alert is already acknowledged.":            "L'alerte est déjà acquittée.",
		"Error: %s":                                 "Erreur : %s",
		"System error. Please visit the dashboard.": "Erreur système. Veuillez consulter le tableau de bord.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s avec des notifications d'alerte. Le service « %s » a %d alertes non acquittées.",
		"%s with an alert notification. %s.":                                      "%s avec une notification d'alerte. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s avec une mise à jour de statut pour l'alerte « %s ». %s",
		"%s with a test message.":                                                 "%s avec un message de test.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s avec votre code de vérification à %d chiffres. Le code est : %s. Encore une fois, votre code de vérification à %d chiffres est : %s.",
		"No summary provided": "Aucun résumé fourni",
	})
}
