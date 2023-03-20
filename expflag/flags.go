package expflag

import "sort"

type Flag string

const (
	Example     Flag = "example"
	SlackDM     Flag = "slack-dm"
	ChanWebhook Flag = "chan-webhook"
	SlackUsrGrp Flag = "slack-usr-grp"
)

var desc = map[Flag]string{
	Example:     "An example experimental flag to demonstrate usage.",
	SlackDM:     "Enables sending notifications to Slack DMs as a user contact method.",
	ChanWebhook: "Enables webhooks as a notification channel type",
	SlackUsrGrp: "Enables updating Slack user groups with schedule on-call users.",
}

// AllFlags returns a slice of all experimental flags sorted by name.
func AllFlags() []Flag {
	var result []Flag
	for k := range desc {
		result = append(result, k)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

// Description returns the description of the given flag.
func Description(f Flag) string {
	return desc[f]
}
