// 	blocks := fmt.Sprintf(`
// 	[
// 		{
// 			"type": "section",
// 			"text": {
// 				"type": "mrkdwn",
// 				"text": "%s"
// 			}
// 		},
// 		{
// 			"type": "actions",
// 			"elements": [
// 				{
// 					"type": "button",
// 					"text": {
// 						"type": "plain_text",
// 						"text": "Acknowledge :eyes:",
// 						"emoji": true
// 					},
// 					"action_id": "ack",
// 					"value": "%[2]d",
// 					"style": "primary"
// 				},
// 				{
// 					"type": "button",
// 					"text": {
// 						"type": "plain_text",
// 						"text": "Close :ballot_box_with_check:",
// 						"emoji": true
// 					},
// 					"action_id": "close",
// 					"value": "%[2]d"
// 				},
// 				{
// 					"type": "button",
// 					"text": {
// 						"type": "plain_text",
// 						"text": "Escalate :arrow_up:",
// 						"emoji": true
// 					},
// 					"action_id": "esc",
// 					"value": "%[2]d",
// 					"style": "danger"
// 				},
// 				{
// 					"type": "button",
// 					"text": {
// 						"type": "plain_text",
// 						"text": "Open in GoAlert :link:",
// 						"emoji": true
// 					},
// 					"action_id": "open",
// 					"url": "%s"
// 				}
// 			]
// 		}
// 	]
// `, summaryText, alertID, url)

// 	return blocks, nil
// }
package slack

import "github.com/slack-go/slack"

func ActionsBlock() slack.Block {
	return slack.Block{}
}
