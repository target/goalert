package twilio

import "golang.org/x/text/language"

// Vietnamese voice translations. Registered under the base language tag so the
// regional variant (vi-VN) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Vietnamese, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "Để xác nhận hủy đăng ký số này, nhấn phím %s.",
		"To go back to the previous menu, press %s.":               "Để quay lại menu trước, nhấn phím %s.",
		"To disable voice notifications to this number, press %s.": "Để tắt thông báo bằng giọng nói tới số này, nhấn phím %s.",
		"To repeat this message, press star.":                      "Để nghe lại tin nhắn này, nhấn phím sao.",
		"To acknowledge, press %s.":                                "Để xác nhận, nhấn phím %s.",
		"To escalate, press %s.":                                   "Để chuyển cấp, nhấn phím %s.",
		"To close, press %s.":                                      "Để đóng, nhấn phím %s.",
		"To acknowledge all, press %s.":                            "Để xác nhận tất cả, nhấn phím %s.",
		"To close all, press %s.":                                  "Để đóng tất cả, nhấn phím %s.",
		// general prompts
		"If you are done, you may simply hang up.": "Nếu bạn đã xong, bạn có thể gác máy.",
		"Sorry, I didn't understand that.":         "Xin lỗi, tôi không hiểu.",
		"Goodbye.":                                 "Tạm biệt.",
		// call flow
		"Hello! This is %s":   "Xin chào! Đây là %s",
		"Hello! This is %s. ": "Xin chào! Đây là %s. ",
		"Please use the application dashboard to manage alerts.": "Vui lòng dùng bảng điều khiển ứng dụng để quản lý cảnh báo.",
		"Unenrolled.":        "Đã hủy đăng ký.",
		"One moment please.": "Xin chờ một lát.",
		"An error has occurred. Please use the dashboard to manage alerts.": "Đã xảy ra lỗi. Vui lòng dùng bảng điều khiển để quản lý cảnh báo.",
		"The menu options have changed. To acknowledge, press %s.":          "Các tùy chọn menu đã thay đổi. Để xác nhận, nhấn phím %s.",
		"The menu options have changed. To close, press %s.":                "Các tùy chọn menu đã thay đổi. Để đóng, nhấn phím %s.",
		// action confirmations
		"Acknowledged":                     "Đã xác nhận",
		"Acknowledged all alerts.":         "Đã xác nhận tất cả cảnh báo.",
		"Closed":                           "Đã đóng",
		"Closed all alerts.":               "Đã đóng tất cả cảnh báo.",
		"Escalation requested":             "Đã yêu cầu chuyển cấp",
		"Escalation requested all alerts.": "Đã yêu cầu chuyển cấp tất cả cảnh báo.",
		// error messages
		"Already %s":                                "Đã %s rồi",
		"Alert is already closed.":                  "Cảnh báo đã được đóng.",
		"Alert is already acknowledged.":            "Cảnh báo đã được xác nhận.",
		"Error: %s":                                 "Lỗi: %s",
		"System error. Please visit the dashboard.": "Lỗi hệ thống. Vui lòng truy cập bảng điều khiển.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s với thông báo cảnh báo. Dịch vụ '%s' có %d cảnh báo chưa được xác nhận.",
		"%s with an alert notification. %s.":                                      "%s với một thông báo cảnh báo. %s.",
		"%s with a status update for alert '%s'. %s":                              "%s với cập nhật trạng thái cho cảnh báo '%s'. %s",
		"%s with a test message.":                                                 "%s với một tin nhắn thử nghiệm.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s với mã xác minh %d chữ số của bạn. Mã là: %s. Nhắc lại, mã xác minh %d chữ số của bạn là: %s.",
		"No summary provided": "Không có tóm tắt",
	})
}
