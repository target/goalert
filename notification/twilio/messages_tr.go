package twilio

import "golang.org/x/text/language"

// Turkish voice translations. Registered under the base language tag so the
// regional variant (tr-TR) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Turkish, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Bu numaranın kaydını silmek için %s tuşuna basın.",
		"To go back to the previous menu, press %s.":               "Önceki menüye dönmek için %s tuşuna basın.",
		"To disable voice notifications to this number, press %s.": "Bu numaraya sesli bildirimleri kapatmak için %s tuşuna basın.",
		"To repeat this message, press star.":                      "Bu mesajı tekrarlamak için yıldız tuşuna basın.",
		"To acknowledge, press %s.":                                "Onaylamak için %s tuşuna basın.",
		"To escalate, press %s.":                                   "Yükseltmek için %s tuşuna basın.",
		"To close, press %s.":                                      "Kapatmak için %s tuşuna basın.",
		"To acknowledge all, press %s.":                            "Tümünü onaylamak için %s tuşuna basın.",
		"To close all, press %s.":                                  "Tümünü kapatmak için %s tuşuna basın.",
		// general prompts
		"If you are done, you may simply hang up.": "İşiniz bittiyse telefonu kapatabilirsiniz.",
		"Sorry, I didn't understand that.":         "Üzgünüm, anlayamadım.",
		"Goodbye.":                                 "Hoşça kalın.",
		// call flow
		"Hello! This is %s":   "Merhaba! Ben %s",
		"Hello! This is %s. ": "Merhaba! Ben %s. ",
		"Please use the application dashboard to manage alerts.": "Lütfen uyarıları yönetmek için uygulama panosunu kullanın.",
		"Unenrolled.":        "Kaydı silindi.",
		"One moment please.": "Lütfen bir saniye.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Bir hata oluştu. Lütfen uyarıları yönetmek için panoyu kullanın.",
		"The menu options have changed. To acknowledge, press %s.":          "Menü seçenekleri değişti. Onaylamak için %s tuşuna basın.",
		"The menu options have changed. To close, press %s.":                "Menü seçenekleri değişti. Kapatmak için %s tuşuna basın.",
		// action confirmations
		"Acknowledged":                     "Onaylandı",
		"Acknowledged all alerts.":         "Tüm uyarılar onaylandı.",
		"Closed":                           "Kapatıldı",
		"Closed all alerts.":               "Tüm uyarılar kapatıldı.",
		"Escalation requested":             "Yükseltme istendi",
		"Escalation requested all alerts.": "Tüm uyarılar için yükseltme istendi.",
		// error messages
		"Already %s":                                "Zaten %s",
		"Alert is already closed.":                  "Uyarı zaten kapatıldı.",
		"Alert is already acknowledged.":            "Uyarı zaten onaylandı.",
		"Error: %s":                                 "Hata: %s",
		"System error. Please visit the dashboard.": "Sistem hatası. Lütfen panoyu ziyaret edin.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s, uyarı bildirimleriyle. '%s' hizmetinde %d onaylanmamış uyarı var.",
		"%s with an alert notification. %s.":                                      "%s, bir uyarı bildirimiyle. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s, '%s' uyarısı için bir durum güncellemesiyle. %s",
		"%s with a test message.":                                                 "%s, bir test mesajıyla.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s, %d haneli doğrulama kodunuzla. Kod: %s. Tekrar, %d haneli doğrulama kodunuz: %s.",
		"No summary provided": "Özet sağlanmadı",
	})
}
