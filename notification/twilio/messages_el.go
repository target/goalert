package twilio

import "golang.org/x/text/language"

// Greek voice translations. Registered under the base language tag so every
// regional variant (el-GR) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Greek, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Για να επιβεβαιώσετε την κατάργηση εγγραφής αυτού του αριθμού, πατήστε %s.",
		"To go back to the previous menu, press %s.":               "Για να επιστρέψετε στο προηγούμενο μενού, πατήστε %s.",
		"To disable voice notifications to this number, press %s.": "Για να απενεργοποιήσετε τις φωνητικές ειδοποιήσεις σε αυτόν τον αριθμό, πατήστε %s.",
		"To repeat this message, press star.":                      "Για να επαναλάβετε αυτό το μήνυμα, πατήστε αστερίσκο.",
		"To acknowledge, press %s.":                                "Για αναγνώριση, πατήστε %s.",
		"To escalate, press %s.":                                   "Για κλιμάκωση, πατήστε %s.",
		"To close, press %s.":                                      "Για κλείσιμο, πατήστε %s.",
		"To acknowledge all, press %s.":                            "Για αναγνώριση όλων, πατήστε %s.",
		"To close all, press %s.":                                  "Για κλείσιμο όλων, πατήστε %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Αν τελειώσατε, μπορείτε απλώς να κλείσετε.",
		"Sorry, I didn't understand that.":         "Συγγνώμη, δεν το κατάλαβα.",
		"Goodbye.":                                 "Αντίο.",
		// call flow
		"Hello! This is %s":   "Γεια σας! Εδώ %s",
		"Hello! This is %s. ": "Γεια σας! Εδώ %s. ",
		"Please use the application dashboard to manage alerts.": "Χρησιμοποιήστε τον πίνακα ελέγχου της εφαρμογής για τη διαχείριση των ειδοποιήσεων.",
		"Unenrolled.":        "Η εγγραφή καταργήθηκε.",
		"One moment please.": "Μια στιγμή παρακαλώ.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Παρουσιάστηκε σφάλμα. Χρησιμοποιήστε τον πίνακα ελέγχου για τη διαχείριση των ειδοποιήσεων.",
		"The menu options have changed. To acknowledge, press %s.":          "Οι επιλογές του μενού άλλαξαν. Για αναγνώριση, πατήστε %s.",
		"The menu options have changed. To close, press %s.":                "Οι επιλογές του μενού άλλαξαν. Για κλείσιμο, πατήστε %s.",
		// action confirmations
		"Acknowledged":                     "Αναγνωρίστηκε",
		"Acknowledged all alerts.":         "Αναγνωρίστηκαν όλες οι ειδοποιήσεις.",
		"Closed":                           "Έκλεισε",
		"Closed all alerts.":               "Έκλεισαν όλες οι ειδοποιήσεις.",
		"Escalation requested":             "Ζητήθηκε κλιμάκωση",
		"Escalation requested all alerts.": "Ζητήθηκε κλιμάκωση για όλες τις ειδοποιήσεις.",
		// error messages
		"Already %s":                                "Ήδη %s",
		"Alert is already closed.":                  "Η ειδοποίηση έχει ήδη κλείσει.",
		"Alert is already acknowledged.":            "Η ειδοποίηση έχει ήδη αναγνωριστεί.",
		"Error: %s":                                 "Σφάλμα: %s",
		"System error. Please visit the dashboard.": "Σφάλμα συστήματος. Επισκεφθείτε τον πίνακα ελέγχου.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s με ειδοποιήσεις. Η υπηρεσία «%s» έχει %d μη αναγνωρισμένες ειδοποιήσεις.",
		"%s with an alert notification. %s.":                                      "%s με μια ειδοποίηση. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s με ενημέρωση κατάστασης για την ειδοποίηση «%s». %s",
		"%s with a test message.":                                                 "%s με ένα δοκιμαστικό μήνυμα.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s με τον %d-ψήφιο κωδικό επαλήθευσης. Ο κωδικός είναι: %s. Ξανά, ο %d-ψήφιος κωδικός επαλήθευσης είναι: %s.",
		"No summary provided": "Δεν δόθηκε περίληψη",
	})
}
