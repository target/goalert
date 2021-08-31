package slack

import (
	"strconv"

	"github.com/slack-go/slack"
	"github.com/target/goalert/alert"
)

func needsAuthMsgOpt() slack.MsgOption {
	msg := slack.NewTextBlockObject("plain_text", "Unauthorized. Please link your GoAlert account to continue.", false, false)
	return slack.MsgOptionBlocks(slack.NewSectionBlock(msg, nil, nil))
}

func makeAlertMessageHeader(a alert.Alert) *slack.HeaderBlock {
	var s string
	switch a.Status {
	case alert.StatusTriggered:
		s = "Unacknowledged"
	case alert.StatusActive:
		s = "Acknowledged"
	case alert.StatusClosed:
		s = "Closed"
	default:
		panic("alert type not supported")
	}
	txt := strconv.Itoa(a.ID) + ": " + s
	summaryText := slack.NewTextBlockObject("plain_text", txt, false, false)
	return slack.NewHeaderBlock(summaryText)
}

func makeActionButton(actionType, callbackID, url string) slack.ButtonBlockElement {
	var text string
	switch actionType {
	case "ack":
		text = "Acknowledge :eyes:"
	case "esc":
		text = "Escalate :arrow_up:"
	case "close":
		text = "Close :ballot_box_with_check:"
	case "openLink":
		text = "Open in GoAlert :link:"
	}

	txt := slack.NewTextBlockObject("plain_text", text, true, false)
	el := *slack.NewButtonBlockElement(actionType, callbackID, txt)
	if url != "" {
		el.URL = url
	}
	return el
}

func makeAlertMessageOptions(a alert.Alert, callbackID, url, responseURL string) []slack.MsgOption {
	var msgOpt []slack.MsgOption
	var actions *slack.ActionBlock

	switch a.Status {
	case alert.StatusTriggered:
		actions = slack.NewActionBlock(
			"",
			makeActionButton("ack", callbackID, ""),
			makeActionButton("esc", callbackID, ""),
			makeActionButton("close", callbackID, ""),
			makeActionButton("openLink", callbackID, url),
		)
	case alert.StatusActive:
		actions = slack.NewActionBlock(
			"",
			makeActionButton("esc", callbackID, ""),
			makeActionButton("close", callbackID, ""),
			makeActionButton("openLink", callbackID, url),
		)
	case alert.StatusClosed:
		actions = slack.NewActionBlock(
			"",
			makeActionButton("openLink", callbackID, url),
		)
	}

	if responseURL != "" {
		msgOpt = append(msgOpt, slack.MsgOptionReplaceOriginal(responseURL))
	}

	msgOpt = append(msgOpt,
		// desktop notification text
		slack.MsgOptionText(a.Summary, false),

		// blockkit elements
		slack.MsgOptionBlocks(
			makeAlertMessageHeader(a),
			slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", a.Summary, false, false), nil, nil),
			actions,
		),
	)

	return msgOpt
}
