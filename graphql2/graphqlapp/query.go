package graphqlapp

import (
	context "context"
	"fmt"
	"strings"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

type Query App
type DebugMessage App

func (a *App) Query() graphql2.QueryResolver { return (*Query)(a) }

func (a *App) formatNC(ctx context.Context, id string) (string, error) {
	n, err := a.FindOneNC(ctx, id)
	if err != nil {
		return "", err
	}
	var typeName string
	switch n.Type {
	case notificationchannel.TypeSlack:
		typeName = "Slack"
	default:
		typeName = string(n.Type)
	}

	return fmt.Sprintf("%s (%s)", n.Name, typeName), nil
}

func (a *Query) formatDest(ctx context.Context, dst notification.Dest) (string, error) {
	if !dst.Type.IsUserCM() {
		return (*App)(a).formatNC(ctx, dst.ID)
	}

	var str strings.Builder
	str.WriteString((*App)(a).FormatDestFunc(ctx, dst.Type, dst.Value))
	switch dst.Type {
	case notification.DestTypeSMS:
		str.WriteString(" (SMS)")
	case notification.DestTypeUserEmail:
		str.WriteString(" (Email)")
	case notification.DestTypeVoice:
		str.WriteString(" (Voice)")
	case notification.DestTypeUserWebhook:
		str.Reset()
		str.WriteString("Webhook")
	default:
		str.Reset()
		str.WriteString(dst.Type.String())
	}

	return str.String(), nil
}

func msgStatus(stat notification.Status) string {
	var str strings.Builder
	switch stat.State {
	case notification.StateBundled:
		str.WriteString("Bundled")
	case notification.StateUnknown:
		str.WriteString("Unknown")
	case notification.StateSending:
		str.WriteString("Sending")
	case notification.StatePending:
		str.WriteString("Pending")
	case notification.StateSent:
		str.WriteString("Sent")
	case notification.StateDelivered:
		str.WriteString("Delivered")
	case notification.StateFailedTemp:
		str.WriteString("Failed (temporary)")
	case notification.StateFailedPerm:
		str.WriteString("Failed (permanent)")
	}
	if stat.Details != "" {
		str.WriteString(": ")
		str.WriteString(stat.Details)
	}
	return str.String()
}

func (a *Query) DebugMessages(ctx context.Context, input *graphql2.DebugMessagesInput) ([]graphql2.DebugMessage, error) {
	var options notification.RecentMessageSearchOptions
	if input == nil {
		input = &graphql2.DebugMessagesInput{}
	}
	if input.CreatedAfter != nil {
		options.After = *input.CreatedAfter
	}
	if input.CreatedBefore != nil {
		options.Before = *input.CreatedBefore
	}
	if input.First != nil {
		options.Limit = *input.First
	}
	msgs, err := a.NotificationStore.RecentMessages(ctx, &options)
	if err != nil {
		return nil, err
	}

	var res []graphql2.DebugMessage
	for _, _m := range msgs {
		m := _m // clone since we're taking pointers to fields
		dest, err := a.formatDest(ctx, m.Dest)
		if err != nil {
			return nil, fmt.Errorf("format dest: %w", err)
		}

		msg := graphql2.DebugMessage{
			ID:          m.ID,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
			Type:        strings.TrimPrefix(m.Type.String(), "MessageType"),
			Status:      msgStatus(m.Status),
			Destination: dest,
		}
		if m.UserID != "" {
			msg.UserID = &m.UserID
		}
		if m.UserName != "" {
			msg.UserName = &m.UserName
		}
		if m.Status.SrcValue != "" && m.Dest.Type.IsUserCM() {
			src, err := a.formatDest(ctx, notification.Dest{Type: m.Dest.Type, Value: m.Status.SrcValue})
			if err != nil {
				return nil, fmt.Errorf("format src: %w", err)
			}
			msg.Source = &src
		}
		if m.ServiceID != "" {
			msg.ServiceID = &m.ServiceID
		}
		if m.ServiceName != "" {
			msg.ServiceName = &m.ServiceName
		}
		if m.ProviderID.ExternalID != "" {
			msg.ProviderID = &m.ProviderID.ExternalID
		}

		res = append(res, msg)
	}

	return res, nil
}

func (a *Query) AuthSubjectsForProvider(ctx context.Context, _first *int, _after *string, providerID string) (conn *graphql2.AuthSubjectConnection, err error) {
	var first int
	var after string
	if _after != nil {
		after = *_after
	}
	if _first != nil {
		first = *_first
	} else {
		first = 15
	}
	err = validate.Range("First", first, 1, 300)
	if err != nil {
		return nil, err
	}

	var c struct {
		ProviderID string
		LastID     string
	}

	if after != "" {
		err = search.ParseCursor(after, &c)
		if err != nil {
			return nil, errors.Wrap(err, "parse cursor")
		}
	} else {
		c.ProviderID = providerID
	}

	conn = new(graphql2.AuthSubjectConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	conn.Nodes, err = a.UserStore.FindSomeAuthSubjectsForProvider(ctx, first+1, c.LastID, c.ProviderID)
	if err != nil {
		return nil, err
	}
	if len(conn.Nodes) > first {
		conn.Nodes = conn.Nodes[:first]
		conn.PageInfo.HasNextPage = true
	}
	if len(conn.Nodes) > 0 {
		c.LastID = conn.Nodes[len(conn.Nodes)-1].SubjectID
	}

	cur, err := search.Cursor(c)
	if err != nil {
		return nil, err
	}
	conn.PageInfo.EndCursor = &cur
	return conn, nil
}
