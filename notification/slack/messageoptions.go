package slack

import (
	"github.com/slack-go/slack"
	"github.com/target/goalert/alert"
)

func alertIDAndStatusSection(alertID string, status alert.Status) *slack.HeaderBlock {
	var s string
	switch status {
	case alert.StatusTriggered:
		s = "Unacknowledged"
	case alert.StatusActive:
		s = "Acknowledged"
	case alert.StatusClosed:
		s = "Closed"
	default:
		panic("alert type not supported")
	}
	txt := alertID + ": " + s
	summaryText := slack.NewTextBlockObject("plain_text", txt, false, false)
	return slack.NewHeaderBlock(summaryText)
}

func alertSummarySection(summary string) *slack.SectionBlock {
	summaryText := slack.NewTextBlockObject("mrkdwn", "Summary: "+summary, false, false)
	return slack.NewSectionBlock(summaryText, nil, nil)
}

func ackButton(callbackID string) slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Acknowledge :eyes:", true, false)
	return *slack.NewButtonBlockElement("ack", callbackID, txt)
}

func escButton(callbackID string) *slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Escalate :arrow_up:", true, false)
	return slack.NewButtonBlockElement("esc", callbackID, txt)
}

func closeButton(callbackID string) *slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Close :ballot_box_with_check:", true, false)
	return slack.NewButtonBlockElement("close", callbackID, txt)
}

func openLinkButton(url string) *slack.ButtonBlockElement {
	txt := slack.NewTextBlockObject("plain_text", "Open in GoAlert :link:", true, false)
	s := slack.NewButtonBlockElement("openLink", "", txt)
	s.URL = url
	return s
}

// func alertLastStatusContext(lastStatus string) *slack.ContextBlock {
// 	lastStatusText := slack.NewTextBlockObject("plain_text", lastStatus, true, true)
// 	return slack.NewContextBlock("", []slack.MixedElement{lastStatusText}...)
// }

func needsAuthMsgOpt() slack.MsgOption {
	msg := slack.NewTextBlockObject("plain_text", "Unauthorized. Please link your GoAlert account to continue", false, false)
	return slack.MsgOptionBlocks(slack.NewSectionBlock(msg, nil, nil))
}

func CraftAlertMessage(a alert.Alert, callbackID, url, responseURL string) []slack.MsgOption {
	var msgOpt []slack.MsgOption
	var actions *slack.ActionBlock

	if a.Status == alert.StatusTriggered {
		actions = slack.NewActionBlock("", ackButton(callbackID), escButton(callbackID), closeButton(callbackID), openLinkButton(url))
	} else if a.Status == alert.StatusActive {
		actions = slack.NewActionBlock("", escButton(callbackID), closeButton(callbackID), openLinkButton(url))
	} else {
		actions = slack.NewActionBlock("", openLinkButton(url))
	}

	if responseURL != "" {
		msgOpt = append(msgOpt, slack.MsgOptionReplaceOriginal(responseURL))
	}

	msgOpt = append(msgOpt,
		// desktop notification text
		slack.MsgOptionText(a.Summary, false),

		// blockkit elements
		slack.MsgOptionBlocks(
			alertIDAndStatusSection(callbackID, a.Status),
			alertSummarySection(a.Summary),
			actions,
		),
	)

	return msgOpt
}
