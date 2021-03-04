package slack

import (
	"fmt"
	"strconv"

	"github.com/slack-go/slack"
	"github.com/target/goalert/alert"
)

func AlertIDAndStatusSection(id int, status string) *slack.HeaderBlock {
	var s string
	if status == "triggered" {
		s = "Unacknowledged"
	} else if status == "active" {
		s = "Acknowledged"
	} else {
		s = "Closed"
	}
	txt := fmt.Sprintf("%d: %s", id, s)
	summaryText := slack.NewTextBlockObject("plain_text", txt, false, false)
	return slack.NewHeaderBlock(summaryText)
}

func AlertSummarySection(summary string) *slack.SectionBlock {
	summaryText := slack.NewTextBlockObject("mrkdwn", summary, false, false)
	return slack.NewSectionBlock(summaryText, nil, nil)
}

func ackButton(alertID string) slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Acknowledge :eyes:", true, false)
	return *slack.NewButtonBlockElement("ack", alertID, txt)
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
	s := slack.NewButtonBlockElement("openLink", "", txt)
	s.URL = url
	return s
}

// AlertActionsOnUpdate handles returning the block actions for an alert message
// within Slack. The alert a parameter represents the state of the alert after
// the action has been processed.
func AlertActionsOnUpdate(a int, status alert.Status, url string) *slack.ActionBlock {
	alertID := strconv.Itoa(a)

	if status == alert.StatusTriggered {
		return slack.NewActionBlock("", ackButton(alertID), escButton(alertID), closeButton(alertID), openLinkButton(url))
	} else if status == alert.StatusActive {
		return slack.NewActionBlock("", escButton(alertID), closeButton(alertID), openLinkButton(url))
	} else {
		return slack.NewActionBlock("", openLinkButton(url))
	}
}

func AlertLastStatusContext(lastStatus string) *slack.ContextBlock {
	lastStatusText := slack.NewTextBlockObject("plain_text", lastStatus, true, true)
	return slack.NewContextBlock("", []slack.MixedElement{lastStatusText}...)
}

func UserAuthMessageBlock() *slack.SectionBlock {
	msg := slack.NewTextBlockObject("plain_text", "Please authenticate with GoAlert before continuing", false, false)

	txt := slack.NewTextBlockObject("plain_text", "Authenticate :link:", true, false)
	btn := slack.NewButtonBlockElement("auth", "", txt)

	btn.URL = "google.com" // slack oauth endpoint

	accessory := slack.NewAccessory(btn)

	return slack.NewSectionBlock(msg, nil, accessory)
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
