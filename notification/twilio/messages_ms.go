package twilio

import "golang.org/x/text/language"

// Malay voice translations. Registered under the base language tag so the
// regional variant (ms-MY) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Malay, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Untuk mengesahkan pembatalan pendaftaran nombor ini, tekan %s.",
		"To go back to the previous menu, press %s.":               "Untuk kembali ke menu sebelumnya, tekan %s.",
		"To disable voice notifications to this number, press %s.": "Untuk mematikan pemberitahuan suara ke nombor ini, tekan %s.",
		"To repeat this message, press star.":                      "Untuk mengulang mesej ini, tekan bintang.",
		"To acknowledge, press %s.":                                "Untuk mengakui, tekan %s.",
		"To escalate, press %s.":                                   "Untuk meningkatkan, tekan %s.",
		"To close, press %s.":                                      "Untuk menutup, tekan %s.",
		"To acknowledge all, press %s.":                            "Untuk mengakui semua, tekan %s.",
		"To close all, press %s.":                                  "Untuk menutup semua, tekan %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Jika anda sudah selesai, anda boleh letak telefon sahaja.",
		"Sorry, I didn't understand that.":         "Maaf, saya tidak faham.",
		"Goodbye.":                                 "Selamat tinggal.",
		// call flow
		"Hello! This is %s":   "Helo! Ini %s",
		"Hello! This is %s. ": "Helo! Ini %s. ",
		"Please use the application dashboard to manage alerts.": "Sila gunakan papan pemuka aplikasi untuk menguruskan amaran.",
		"Unenrolled.":        "Pendaftaran dibatalkan.",
		"One moment please.": "Sila tunggu sebentar.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Ralat telah berlaku. Sila gunakan papan pemuka untuk menguruskan amaran.",
		"The menu options have changed. To acknowledge, press %s.":          "Pilihan menu telah berubah. Untuk mengakui, tekan %s.",
		"The menu options have changed. To close, press %s.":                "Pilihan menu telah berubah. Untuk menutup, tekan %s.",
		// action confirmations
		"Acknowledged":                     "Diakui",
		"Acknowledged all alerts.":         "Semua amaran telah diakui.",
		"Closed":                           "Ditutup",
		"Closed all alerts.":               "Semua amaran telah ditutup.",
		"Escalation requested":             "Peningkatan diminta",
		"Escalation requested all alerts.": "Peningkatan diminta untuk semua amaran.",
		// error messages
		"Already %s":                                "Sudah %s",
		"Alert is already closed.":                  "Amaran sudah ditutup.",
		"Alert is already acknowledged.":            "Amaran sudah diakui.",
		"Error: %s":                                 "Ralat: %s",
		"System error. Please visit the dashboard.": "Ralat sistem. Sila lawati papan pemuka.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s dengan pemberitahuan amaran. Perkhidmatan '%s' mempunyai %d amaran yang belum diakui.",
		"%s with an alert notification. %s.":                                      "%s dengan satu pemberitahuan amaran. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s dengan kemas kini status untuk amaran '%s'. %s",
		"%s with a test message.":                                                 "%s dengan mesej ujian.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s dengan kod pengesahan %d digit anda. Kodnya ialah: %s. Sekali lagi, kod pengesahan %d digit anda ialah: %s.",
		"No summary provided": "Tiada ringkasan diberikan",
	})
}
