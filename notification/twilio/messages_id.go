package twilio

import "golang.org/x/text/language"

// Indonesian voice translations. Registered under the base language tag so the
// regional variant (id-ID) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Indonesian, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Untuk mengonfirmasi pembatalan pendaftaran nomor ini, tekan %s.",
		"To go back to the previous menu, press %s.":               "Untuk kembali ke menu sebelumnya, tekan %s.",
		"To disable voice notifications to this number, press %s.": "Untuk menonaktifkan notifikasi suara ke nomor ini, tekan %s.",
		"To repeat this message, press star.":                      "Untuk mengulangi pesan ini, tekan bintang.",
		"To acknowledge, press %s.":                                "Untuk mengakui, tekan %s.",
		"To escalate, press %s.":                                   "Untuk meningkatkan, tekan %s.",
		"To close, press %s.":                                      "Untuk menutup, tekan %s.",
		"To acknowledge all, press %s.":                            "Untuk mengakui semua, tekan %s.",
		"To close all, press %s.":                                  "Untuk menutup semua, tekan %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Jika Anda sudah selesai, Anda boleh langsung menutup telepon.",
		"Sorry, I didn't understand that.":         "Maaf, saya tidak mengerti.",
		"Goodbye.":                                 "Sampai jumpa.",
		// call flow
		"Hello! This is %s":   "Halo! Ini %s",
		"Hello! This is %s. ": "Halo! Ini %s. ",
		"Please use the application dashboard to manage alerts.": "Silakan gunakan dasbor aplikasi untuk mengelola peringatan.",
		"Unenrolled.":        "Pendaftaran dibatalkan.",
		"One moment please.": "Mohon tunggu sebentar.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Terjadi kesalahan. Silakan gunakan dasbor untuk mengelola peringatan.",
		"The menu options have changed. To acknowledge, press %s.":          "Pilihan menu telah berubah. Untuk mengakui, tekan %s.",
		"The menu options have changed. To close, press %s.":                "Pilihan menu telah berubah. Untuk menutup, tekan %s.",
		// action confirmations
		"Acknowledged":                     "Diakui",
		"Acknowledged all alerts.":         "Semua peringatan telah diakui.",
		"Closed":                           "Ditutup",
		"Closed all alerts.":               "Semua peringatan telah ditutup.",
		"Escalation requested":             "Peningkatan diminta",
		"Escalation requested all alerts.": "Peningkatan diminta untuk semua peringatan.",
		// error messages
		"Already %s":                                "Sudah %s",
		"Alert is already closed.":                  "Peringatan sudah ditutup.",
		"Alert is already acknowledged.":            "Peringatan sudah diakui.",
		"Error: %s":                                 "Kesalahan: %s",
		"System error. Please visit the dashboard.": "Kesalahan sistem. Silakan kunjungi dasbor.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s dengan notifikasi peringatan. Layanan '%s' memiliki %d peringatan yang belum diakui.",
		"%s with an alert notification. %s.":                                      "%s dengan notifikasi peringatan. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s dengan pembaruan status untuk peringatan '%s'. %s",
		"%s with a test message.":                                                 "%s dengan pesan uji coba.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s dengan kode verifikasi %d digit Anda. Kodenya adalah: %s. Sekali lagi, kode verifikasi %d digit Anda adalah: %s.",
		"No summary provided": "Tidak ada ringkasan yang diberikan",
	})
}
