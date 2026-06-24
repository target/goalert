package twilio

import "golang.org/x/text/language"

// Chinese (Simplified) voice translations. Registered under the base language
// tag so every regional variant (cmn-CN, cmn-TW) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Chinese, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "如需确认取消订阅此号码，请按 %s。",
		"To go back to the previous menu, press %s.":               "如需返回上一级菜单，请按 %s。",
		"To disable voice notifications to this number, press %s.": "如需关闭发送到此号码的语音通知，请按 %s。",
		"To repeat this message, press star.":                      "如需重复此消息，请按星号键。",
		"To acknowledge, press %s.":                                "如需确认，请按 %s。",
		"To escalate, press %s.":                                   "如需升级，请按 %s。",
		"To close, press %s.":                                      "如需关闭，请按 %s。",
		"To acknowledge all, press %s.":                            "如需全部确认，请按 %s。",
		"To close all, press %s.":                                  "如需全部关闭，请按 %s。",
		// general prompts
		"If you are done, you may simply hang up.": "如果您已完成操作，可以直接挂断电话。",
		"Sorry, I didn't understand that.":         "抱歉，我没有听明白。",
		"Goodbye.":                                 "再见。",
		// call flow
		"Hello! This is %s":   "您好！这里是 %s",
		"Hello! This is %s. ": "您好！这里是 %s。 ",
		"Please use the application dashboard to manage alerts.": "请使用应用程序的控制面板来管理警报。",
		"Unenrolled.":        "已取消订阅。",
		"One moment please.": "请稍候。",
		"An error has occurred. Please use the dashboard to manage alerts.": "发生了一个错误。请使用控制面板来管理警报。",
		"The menu options have changed. To acknowledge, press %s.":          "菜单选项已更改。如需确认，请按 %s。",
		"The menu options have changed. To close, press %s.":                "菜单选项已更改。如需关闭，请按 %s。",
		// action confirmations
		"Acknowledged":                     "已确认",
		"Acknowledged all alerts.":         "已确认所有警报。",
		"Closed":                           "已关闭",
		"Closed all alerts.":               "已关闭所有警报。",
		"Escalation requested":             "已请求升级",
		"Escalation requested all alerts.": "已为所有警报请求升级。",
		// error messages
		"Already %s":                                "已经%s",
		"Alert is already closed.":                  "该警报已经关闭。",
		"Alert is already acknowledged.":            "该警报已经确认。",
		"Error: %s":                                 "错误：%s",
		"System error. Please visit the dashboard.": "系统错误。请访问控制面板。",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s，发送警报通知。服务“%s”有 %d 条未确认的警报。",
		"%s with an alert notification. %s.":                                      "%s，发送一条警报通知。%s。",
		"%s with a status update for alert '%s'. %s":                              "%s，发送警报“%s”的状态更新。%s",
		"%s with a test message.":                                                 "%s，发送一条测试消息。",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s，发送您的 %d 位验证码。验证码是：%s。再说一遍，您的 %d 位验证码是：%s。",
		"No summary provided": "未提供摘要",
	})
}
