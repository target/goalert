package twilio

import "golang.org/x/text/language"

// Korean voice translations. Registered under the base language tag so every
// regional variant (ko-KR) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Korean, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "이 번호의 등록 해지를 확인하려면 %s 번을 누르세요.",
		"To go back to the previous menu, press %s.":               "이전 메뉴로 돌아가려면 %s 번을 누르세요.",
		"To disable voice notifications to this number, press %s.": "이 번호의 음성 알림을 끄려면 %s 번을 누르세요.",
		"To repeat this message, press star.":                      "이 메시지를 다시 들으려면 별표를 누르세요.",
		"To acknowledge, press %s.":                                "확인하려면 %s 번을 누르세요.",
		"To escalate, press %s.":                                   "에스컬레이션하려면 %s 번을 누르세요.",
		"To close, press %s.":                                      "종료하려면 %s 번을 누르세요.",
		"To acknowledge all, press %s.":                            "모두 확인하려면 %s 번을 누르세요.",
		"To close all, press %s.":                                  "모두 종료하려면 %s 번을 누르세요.",
		// general prompts
		"If you are done, you may simply hang up.": "끝나셨으면 그냥 전화를 끊으셔도 됩니다.",
		"Sorry, I didn't understand that.":         "죄송합니다. 이해하지 못했습니다.",
		"Goodbye.":                                 "안녕히 계세요.",
		// call flow
		"Hello! This is %s":   "안녕하세요! %s입니다",
		"Hello! This is %s. ": "안녕하세요! %s입니다. ",
		"Please use the application dashboard to manage alerts.": "알림을 관리하려면 애플리케이션 대시보드를 이용하세요.",
		"Unenrolled.":        "등록이 해지되었습니다.",
		"One moment please.": "잠시만 기다려 주세요.",
		"An error has occurred. Please use the dashboard to manage alerts.": "오류가 발생했습니다. 알림을 관리하려면 대시보드를 이용하세요.",
		"The menu options have changed. To acknowledge, press %s.":          "메뉴 옵션이 변경되었습니다. 확인하려면 %s 번을 누르세요.",
		"The menu options have changed. To close, press %s.":                "메뉴 옵션이 변경되었습니다. 종료하려면 %s 번을 누르세요.",
		// action confirmations
		"Acknowledged":                     "확인됨",
		"Acknowledged all alerts.":         "모든 알림이 확인되었습니다.",
		"Closed":                           "종료됨",
		"Closed all alerts.":               "모든 알림이 종료되었습니다.",
		"Escalation requested":             "에스컬레이션이 요청됨",
		"Escalation requested all alerts.": "모든 알림에 대해 에스컬레이션이 요청되었습니다.",
		// error messages
		"Already %s":                                "이미 %s",
		"Alert is already closed.":                  "알림이 이미 종료되었습니다.",
		"Alert is already acknowledged.":            "알림이 이미 확인되었습니다.",
		"Error: %s":                                 "오류: %s",
		"System error. Please visit the dashboard.": "시스템 오류입니다. 대시보드를 확인해 주세요.",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "알림 통지를 전하는 %s입니다. '%s' 서비스에 확인되지 않은 알림이 %d건 있습니다.",
		"%s with an alert notification. %s.":                                      "알림 통지를 전하는 %s입니다. %s.",
		"%s with a status update for alert '%s'. %s":                              "'%s' 알림의 상태 업데이트를 전하는 %s입니다. %s",
		"%s with a test message.":                                                 "테스트 메시지를 전하는 %s입니다.",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s입니다. %d자리 인증 코드를 전해 드립니다. 코드는 %s입니다. 다시 한 번, %d자리 인증 코드는 %s입니다.",
		"No summary provided": "요약이 제공되지 않았습니다",
	})
}
