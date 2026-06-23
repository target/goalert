package twilio

import "golang.org/x/text/language"

// Japanese voice translations. Registered under the base language tag so every
// regional variant (ja-JP) resolves to it.
//
// Placeholders (%s, %d) must appear in the same order as the English source
// string. Keys must match those in voiceKeys exactly.
func init() {
	registerVoiceMessages(language.Japanese, map[string]string{
		// menu options
		"To confirm unenrollment of this number, press %s.":        "この番号の登録解除を確定するには、%s を押してください。",
		"To go back to the previous menu, press %s.":               "前のメニューに戻るには、%s を押してください。",
		"To disable voice notifications to this number, press %s.": "この番号への音声通知を無効にするには、%s を押してください。",
		"To repeat this message, press star.":                      "このメッセージをもう一度聞くには、アスタリスクを押してください。",
		"To acknowledge, press %s.":                                "受信確認するには、%s を押してください。",
		"To escalate, press %s.":                                   "エスカレーションするには、%s を押してください。",
		"To close, press %s.":                                      "クローズするには、%s を押してください。",
		"To acknowledge all, press %s.":                            "すべて受信確認するには、%s を押してください。",
		"To close all, press %s.":                                  "すべてクローズするには、%s を押してください。",
		// general prompts
		"If you are done, you may simply hang up.": "用件が済みましたら、そのまま電話をお切りください。",
		"Sorry, I didn't understand that.":         "申し訳ありません、聞き取れませんでした。",
		"Goodbye.":                                 "さようなら。",
		// call flow
		"Hello! This is %s":   "こんにちは。%s です",
		"Hello! This is %s. ": "こんにちは。%s です。",
		"Please use the application dashboard to manage alerts.": "アラートの管理には、アプリケーションのダッシュボードをご利用ください。",
		"Unenrolled.":        "登録を解除しました。",
		"One moment please.": "少々お待ちください。",
		"An error has occurred. Please use the dashboard to manage alerts.": "エラーが発生しました。アラートの管理には、ダッシュボードをご利用ください。",
		"The menu options have changed. To acknowledge, press %s.":          "メニューの選択肢が変更されました。受信確認するには、%s を押してください。",
		"The menu options have changed. To close, press %s.":                "メニューの選択肢が変更されました。クローズするには、%s を押してください。",
		// action confirmations
		"Acknowledged":                     "受信確認しました",
		"Acknowledged all alerts.":         "すべてのアラートを受信確認しました。",
		"Closed":                           "クローズしました",
		"Closed all alerts.":               "すべてのアラートをクローズしました。",
		"Escalation requested":             "エスカレーションを要求しました",
		"Escalation requested all alerts.": "すべてのアラートのエスカレーションを要求しました。",
		// error messages
		"Already %s":                                "すでに %s",
		"Alert is already closed.":                  "アラートはすでにクローズされています。",
		"Alert is already acknowledged.":            "アラートはすでに受信確認されています。",
		"Error: %s":                                 "エラー: %s",
		"System error. Please visit the dashboard.": "システムエラーです。ダッシュボードをご確認ください。",
		// buildMessage templates
		"%s with alert notifications. Service '%s' has %d unacknowledged alerts.": "%s よりアラート通知です。サービス「%s」に未確認のアラートが %d 件あります。",
		"%s with an alert notification. %s.":                                      "%s よりアラート通知です。%s。",
		"%s with a status update for alert '%s'. %s":                              "%s よりアラート「%s」の状況更新です。%s",
		"%s with a test message.":                                                 "%s よりテストメッセージです。",
		"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.": "%s より %d 桁の確認コードをお知らせします。コードは %s です。もう一度、%d 桁の確認コードは %s です。",
		"No summary provided": "概要はありません",
	})
}
