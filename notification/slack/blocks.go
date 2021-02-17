package slack

import (
	"fmt"
	"strconv"

	"github.com/slack-go/slack"
	"github.com/target/goalert/alert"
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

func ackButton(alertID string) *slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Acknowledge :eyes:", true, false)
	return slack.NewButtonBlockElement("ack", alertID, txt)
}

func escButton(alertID string) *slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Escalate :arrow_up:", true, false)
	return slack.NewButtonBlockElement("esc", alertID, txt)
}

func closeButton(alertID string) *slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Close :ballot_box_with_check:", true, false)
	return slack.NewButtonBlockElement("close", alertID, txt)
}

func openLinkButton(url string) *slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Open in GoAlert :link:", true, false)
	return slack.NewButtonBlockElement("openLink", url, txt)
}

func AlertActionsOnCreate(id int, url string) *slack.ActionBlock {
	value := strconv.Itoa(id)

	blocks := []slack.BlockElement{
		ackButton(value),
		escButton(value),
		closeButton(value),
		openLinkButton(url),
	}

	return slack.NewActionBlock("", blocks...)
}

// AlertActionsOnUpdate handles returning the block actions for an alert message
// within Slack. The alert a parameter represents the state of the alert after
// the action has been processed.
func AlertActionsOnUpdate(a alert.Alert, url string) *slack.ActionBlock {
	var buttons = make([]slack.BlockElement, 4)
	alertID := strconv.Itoa(a.ID)

	if a.Status == alert.StatusTriggered {
		buttons = append(buttons, ackButton(alertID))
	}
	if a.Status == alert.StatusTriggered || a.Status == alert.StatusActive {
		buttons = append(buttons, escButton(alertID))
	}
	if a.Status == alert.StatusTriggered || a.Status == alert.StatusActive {
		buttons = append(buttons, closeButton(alertID))
	}
	if url != "" {
		buttons = append(buttons, openLinkButton(url))
	}

	return slack.NewActionBlock("", buttons...)
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
