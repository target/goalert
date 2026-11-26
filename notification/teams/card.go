package teams

import (
	"context"
	"fmt"
	"strings"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
)

// Adaptive Card color values (https://adaptivecards.io/explorer/TextBlock.html).
const (
	colorClosed  = "Good"
	colorUnacked = "Attention"
	colorAcked   = "Warning"
)

// CardVersion is the Adaptive Card schema version used for all outgoing
// cards. 1.2 is the safest widely-supported version in Microsoft Teams,
// including cards posted through Power Automate Workflows.
const CardVersion = "1.2"

const maxDetailsLen = 2000

// AdaptiveCard is a minimal Adaptive Card representation, limited to the
// conservative subset of the schema that Microsoft Teams supports when cards
// are posted via a Power Automate Workflow webhook.
type AdaptiveCard struct {
	Type    string       `json:"type"`
	Schema  string       `json:"$schema"`
	Version string       `json:"version"`
	Body    []any        `json:"body"`
	Actions []CardAction `json:"actions,omitempty"`
	MSTeams *MSTeams     `json:"msteams,omitempty"`
}

// MSTeams contains Teams-specific rendering properties.
type MSTeams struct {
	Width string `json:"width,omitempty"`
}

// TextBlock is an Adaptive Card TextBlock element.
type TextBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Wrap     bool   `json:"wrap,omitempty"`
	Weight   string `json:"weight,omitempty"`
	Size     string `json:"size,omitempty"`
	Color    string `json:"color,omitempty"`
	IsSubtle bool   `json:"isSubtle,omitempty"`
	Spacing  string `json:"spacing,omitempty"`
}

// FactSet is an Adaptive Card FactSet element.
type FactSet struct {
	Type  string `json:"type"`
	Facts []Fact `json:"facts"`
}

// Fact is a single title/value pair within a FactSet.
type Fact struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// CardAction is an Adaptive Card action; only Action.OpenUrl is used since
// submit-style actions require a bot to receive them.
type CardAction struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url,omitempty"`
}

func newCard(body []any, actions ...CardAction) AdaptiveCard {
	return AdaptiveCard{
		Type:    "AdaptiveCard",
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Version: CardVersion,
		Body:    body,
		Actions: actions,
		MSTeams: &MSTeams{Width: "Full"},
	}
}

func openURLAction(title, url string) CardAction {
	return CardAction{Type: "Action.OpenUrl", Title: title, URL: url}
}

func title(text string) TextBlock {
	return TextBlock{Type: "TextBlock", Text: text, Wrap: true, Weight: "Bolder", Size: "Medium"}
}

func body(text string) TextBlock {
	return TextBlock{Type: "TextBlock", Text: text, Wrap: true}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func stateInfo(state notification.AlertState) (color, defaultText string) {
	switch state {
	case notification.AlertStateAcknowledged:
		return colorAcked, "Acknowledged"
	case notification.AlertStateClosed:
		return colorClosed, "Closed"
	default:
		return colorUnacked, "Unacknowledged"
	}
}

// alertCard renders an alert notification or status update as an Adaptive
// Card, colored by the current alert state (matching Slack's color scheme).
func alertCard(ctx context.Context, alertID int, summary, details, serviceName, logEntry string, state notification.AlertState) AdaptiveCard {
	cfg := config.FromContext(ctx)

	color, defaultText := stateInfo(state)
	if logEntry == "" {
		logEntry = defaultText
	}

	elements := []any{
		title(fmt.Sprintf("Alert #%d: %s", alertID, summary)),
		TextBlock{Type: "TextBlock", Text: logEntry, Wrap: true, Color: color, Weight: "Bolder"},
	}

	var facts []Fact
	if serviceName != "" {
		facts = append(facts, Fact{Title: "Service", Value: serviceName})
	}
	if len(facts) > 0 {
		elements = append(elements, FactSet{Type: "FactSet", Facts: facts})
	}

	if details != "" {
		elements = append(elements, TextBlock{
			Type: "TextBlock", Text: truncate(details, maxDetailsLen),
			Wrap: true, IsSubtle: true, Spacing: "Medium",
		})
	}

	return newCard(elements,
		openURLAction("Open Alert", cfg.CallbackURL(fmt.Sprintf("/alerts/%d", alertID))),
	)
}

// alertBundleCard renders a bundled-alerts notification for a service.
func alertBundleCard(ctx context.Context, serviceID, serviceName string, count int) AdaptiveCard {
	cfg := config.FromContext(ctx)

	return newCard([]any{
		title(fmt.Sprintf("Service '%s' has %d unacknowledged alerts.", serviceName, count)),
	},
		openURLAction("Open Alerts", cfg.CallbackURL("/services/"+serviceID+"/alerts")),
	)
}

// onCallCard renders a schedule on-call notification.
func onCallCard(_ context.Context, m notification.ScheduleOnCallUsers) AdaptiveCard {
	var text string
	if len(m.Users) == 0 {
		text = fmt.Sprintf("No one is on-call for %s.", m.ScheduleName)
	} else {
		names := make([]string, len(m.Users))
		for i, u := range m.Users {
			names[i] = fmt.Sprintf("[%s](%s)", u.Name, u.URL)
		}
		text = fmt.Sprintf("On-call for %s: %s", m.ScheduleName, strings.Join(names, ", "))
	}

	return newCard([]any{
		title(fmt.Sprintf("On-Call Notification: %s", m.ScheduleName)),
		body(text),
	},
		openURLAction("Open Schedule", m.ScheduleURL),
	)
}

// testCard renders a test notification.
func testCard(ctx context.Context) AdaptiveCard {
	cfg := config.FromContext(ctx)
	return newCard([]any{
		title(fmt.Sprintf("%s Test Notification", cfg.ApplicationName())),
		body("This is a test message."),
	})
}

// signalCard renders a dynamic-action signal message.
func signalCard(_ context.Context, message string) AdaptiveCard {
	return newCard([]any{body(message)})
}
