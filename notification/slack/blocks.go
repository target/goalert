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
	summaryText := slack.NewTextBlockObject("mrkdwn", "Summary: "+summary, false, false)
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

func UserAuthMessageOption(clientID, uri string) slack.MsgOption {
	msg := slack.NewTextBlockObject("plain_text", "Please link your GoAlert account to continue", false, false)

	btnTxt := slack.NewTextBlockObject("plain_text", "Authenticate :link:", true, false)
	btn := slack.NewButtonBlockElement("auth", "", btnTxt)
	btn.URL = "https://slack.com/oauth/v2/authorize?user_scope=identity.basic&client_id=" + clientID + "&redirect_uri=" + uri // slack oauth endpoint
	accessory := slack.NewAccessory(btn)

	section := slack.NewSectionBlock(msg, nil, accessory)
	return slack.MsgOptionBlocks(section)
}

func CraftAlertMessage(a alert.Alert, url string) []slack.MsgOption {
	var msgOpt []slack.MsgOption
	var actions *slack.ActionBlock

	alertID := strconv.Itoa(a.ID)

	if a.Status == alert.StatusTriggered {
		actions = slack.NewActionBlock("", ackButton(alertID), escButton(alertID), closeButton(alertID), openLinkButton(url))
	} else if a.Status == alert.StatusActive {
		actions = slack.NewActionBlock("", escButton(alertID), closeButton(alertID), openLinkButton(url))
	} else {
		actions = slack.NewActionBlock("", openLinkButton(url))
	}

	msgOpt = append(msgOpt,
		// desktop notification text
		slack.MsgOptionText(a.Summary, false),

		// blockkit elements
		slack.MsgOptionBlocks(
			AlertIDAndStatusSection(a.ID, string(a.Status)),
			AlertSummarySection(a.Summary),
			actions,
		),
	)

	return msgOpt
}
