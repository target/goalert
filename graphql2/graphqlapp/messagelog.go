package graphqlapp

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/search"
)

type MessageLog App

func (a *App) formatNC(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", nil
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return "", err
	}

	n, err := a.FindOneNC(ctx, uid)
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

func (q *Query) formatDest(ctx context.Context, dst notification.Dest) (string, error) {
	if !dst.Type.IsUserCM() {
		return (*App)(q).formatNC(ctx, dst.ID)
	}

	var str strings.Builder
	str.WriteString((*App)(q).FormatDestFunc(ctx, dst.Type, dst.Value))
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

func (q *Query) MessageLogs(ctx context.Context, opts *graphql2.MessageLogSearchOptions) (conn *graphql2.MessageLogConnection, err error) {
	if opts == nil {
		opts = &graphql2.MessageLogSearchOptions{}
	}
	var searchOpts notification.SearchOptions
	if opts.Search != nil {
		searchOpts.Search = *opts.Search
	}
	searchOpts.Omit = opts.Omit
	if opts.After != nil && *opts.After != "" {
		err = search.ParseCursor(*opts.After, &searchOpts)
		if err != nil {
			return nil, err
		}
	}
	if opts.First != nil {
		searchOpts.Limit = *opts.First
	}
	if searchOpts.Limit == 0 {
		searchOpts.Limit = 50 // default limit if unset
	}

	searchOpts.Limit++
	logs, err := q.NotificationStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}
	conn = new(graphql2.MessageLogConnection)
	conn.PageInfo = &graphql2.PageInfo{}

	// more than current limit exists, set page info and cursor
	if len(logs) == searchOpts.Limit {
		logs = logs[:len(logs)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(logs) > 0 {
		last := logs[len(logs)-1]
		searchOpts.After.SrcValue = last.SrcValue

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}

	// map to struct here (for loop/append)

	var logsMapped []graphql2.DebugMessage
	for _, log := range logs {
		dst := notification.DestFromPair(&log.ContactMethod, &log.Channel)
		destStr, err := q.formatDest(ctx, dst)
		if err != nil {
			return nil, fmt.Errorf("format dest: %w", err)
		}

		var x = graphql2.DebugMessage{
			ID:          log.ID,
			CreatedAt:   log.CreatedAt,
			UpdatedAt:   log.LastStatusAt,
			Type:        strings.TrimPrefix(log.MessageType.String(), "MessageType"),
			Status:      msgStatus(notification.Status{State: log.LastStatus, Details: log.StatusDetails}),
			UserID:      &log.User.ID,
			UserName:    &log.User.Name,
			Destination: destStr,
			ServiceID:   &log.Service.ID,
			ServiceName: &log.Service.Name,
			AlertID:     &log.AlertID,
			ProviderID:  &log.ProviderMsgID.ExternalID,
		}
		if log.SrcValue != "" && &log.ContactMethod != nil {
			src, err := q.formatDest(ctx, notification.Dest{Type: dst.Type, Value: log.SrcValue})
			if err != nil {
				return nil, fmt.Errorf("format src: %w", err)
			}
			x.Source = &src
		}
	}

	conn.Nodes = logsMapped
	return conn, nil
}
