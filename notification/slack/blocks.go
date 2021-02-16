package slack

import (
	"fmt"
	"strconv"

	"github.com/slack-go/slack"
)

func AlertIDAndStatusSection(id int, status string) *slack.HeaderBlock {
	txt := fmt.Sprintf("%d: %s", id, status)
	summaryText := slack.NewTextBlockObject("plain_text", txt, false, false)
	return slack.NewHeaderBlock(summaryText)
}

func AlertSummarySection(summary string) *slack.SectionBlock {
	summaryText := slack.NewTextBlockObject("mrkdwn", summary, false, false)
	return slack.NewSectionBlock(summaryText, nil, nil)
}

// ack button, close button, escalate button, open in goalert button
// +1 goal of see details button
func AlertActionsSection(alertID int, ack, escalate, close, openLink bool) *slack.ActionBlock {
	var ackButton, escButton, closeButton, openLinkButton *slack.ButtonBlockElement
	var buttons = make([]slack.BlockElement, 4)
	value := strconv.Itoa(alertID)

	if ack {
		txt := slack.NewTextBlockObject("plain_text", "Acknowledge :eyes:", true, false)
		ackButton = slack.NewButtonBlockElement("ack", value, txt)
		buttons = append(buttons, *ackButton)
	}

	if escalate {
		txt := slack.NewTextBlockObject("plain_text", "Escalate :arrow_up:", true, false)
		escButton = slack.NewButtonBlockElement("esc", value, txt)
		buttons = append(buttons, *escButton)
	}

	if close {
		txt := slack.NewTextBlockObject("plain_text", "Close :ballot_box_with_check:", true, false)
		closeButton = slack.NewButtonBlockElement("close", value, txt)
		buttons = append(buttons, *closeButton)
	}

	if openLink {
		txt := slack.NewTextBlockObject("plain_text", "Open in GoAlert :link:", true, false)
		openLinkButton = slack.NewButtonBlockElement("openLink", value, txt)
		buttons = append(buttons, *openLinkButton)
	}

	return slack.NewActionBlock("", ackButton, escButton, closeButton, openLinkButton)

}

func AlertLastStatusContext(lastStatus string) *slack.ContextBlock {
	lastStatusText := slack.NewTextBlockObject("plain_text", lastStatus, true, true)
	return slack.NewContextBlock("", []slack.MixedElement{lastStatusText}...)
}

// func example() {
// 	// Shared Assets for example
// 	chooseBtnText := slack.NewTextBlockObject("plain_text", "Choose", true, false)
// 	chooseBtnEle := slack.NewButtonBlockElement("", "click_me_123", chooseBtnText)
// 	divSection := slack.NewDividerBlock()

// 	// Option 1
// 	optionOneText := slack.NewTextBlockObject("mrkdwn", "*Today - 4:30-5pm*\nEveryone is available: @iris, @zelda", false, false)
// 	optionOneSection := slack.NewSectionBlock(optionOneText, nil, slack.NewAccessory(chooseBtnEle))

// 	// Build Message with blocks created above
// 	msg := slack.NewBlockMessage(
// 		divSection,
// 		divSection,
// 		optionOneSection,
// 	)

// 	b, err := json.MarshalIndent(msg, "", "    ")
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	fmt.Println(string(b))
// }
