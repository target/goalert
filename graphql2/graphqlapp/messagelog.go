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

func (q *Query) MessageLog(ctx context.Context, id string) (*notification.MessageLog, error) {
	return (*App)(q).FindOneMessageLog(ctx, id)
}

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

	conn.Nodes = logs
	return conn, nil
}

// func (a *Query) DebugMessages(ctx context.Context, input *graphql2.DebugMessagesInput) ([]graphql2.DebugMessage, error) {
// 	err := permission.LimitCheckAny(ctx, permission.Admin)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var msgs []*struct {
// 		ID           string
// 		CreatedAt    time.Time
// 		LastStatusAt time.Time
// 		MessageType  notification.MessageType

// 		LastStatus    notification.State
// 		StatusDetails string
// 		SrcValue      string

// 		UserID string
// 		User   *user.User `gorm:"foreignkey:ID;references:UserID"`

// 		ContactMethodID string
// 		ContactMethod   *contactmethod.ContactMethod `gorm:"foreignKey:ID;references:ContactMethodID"`

// 		ChannelID string
// 		Channel   *notificationchannel.Channel `gorm:"foreignKey:ID;references:ChannelID"`

// 		ServiceID string
// 		Service   *service.Service `gorm:"foreignKey:ID;references:ServiceID"`

// 		AlertID       int
// 		ProviderMsgID *notification.ProviderMessageID
// 	}

// 	db := sqlutil.FromContext(ctx).Table("outgoing_messages")

// 	if input.CreatedAfter != nil {
// 		db = db.Where("created_at >= ?", *input.CreatedAfter)
// 	}
// 	if input.CreatedBefore != nil {
// 		db = db.Where("created_at < ?", *input.CreatedBefore)
// 	}
// 	if input.First != nil {
// 		err = validate.Range("first", *input.First, 0, 1000)
// 		if err != nil {
// 			return nil, err
// 		}
// 		db = db.Limit(*input.First)
// 	} else {
// 		db = db.Limit(search.DefaultMaxResults)
// 	}

// 	err = db.
// 		Preload("User", sqlutil.Columns("ID", "Name")).
// 		Preload("Service", sqlutil.Columns("ID", "Name")).
// 		Preload("Channel", sqlutil.Columns("ID", "Type", "Value")).
// 		Preload("ContactMethod", sqlutil.Columns("ID", "Type", "Value")).
// 		Order("created_at DESC").
// 		Find(&msgs).Error
// 	if err != nil {
// 		return nil, err
// 	}

// 	var res []graphql2.DebugMessage
// 	for _, m := range msgs {
// 		dst := notification.DestFromPair(m.ContactMethod, m.Channel)
// 		destStr, err := a.formatDest(ctx, dst)
// 		if err != nil {
// 			return nil, fmt.Errorf("format dest: %w", err)
// 		}

// 		msg := graphql2.DebugMessage{
// 			ID:          m.ID,
// 			CreatedAt:   m.CreatedAt,
// 			UpdatedAt:   m.LastStatusAt,
// 			Type:        strings.TrimPrefix(m.MessageType.String(), "MessageType"),
// 			Status:      msgStatus(notification.Status{State: m.LastStatus, Details: m.StatusDetails}),
// 			Destination: destStr,
// 		}
// 		if m.User != nil {
// 			msg.UserID = &m.User.ID
// 			msg.UserName = &m.User.Name
// 		}

// 		if m.SrcValue != "" && m.ContactMethod != nil {
// 			src, err := a.formatDest(ctx, notification.Dest{Type: dst.Type, Value: m.SrcValue})
// 			if err != nil {
// 				return nil, fmt.Errorf("format src: %w", err)
// 			}
// 			msg.Source = &src
// 		}
// 		if m.Service != nil {
// 			msg.ServiceID = &m.Service.ID
// 			msg.ServiceName = &m.Service.Name
// 		}
// 		if m.AlertID != 0 {
// 			msg.AlertID = &m.AlertID
// 		}
// 		if m.ProviderMsgID != nil {
// 			msg.ProviderID = &m.ProviderMsgID.ExternalID
// 		}

// 		res = append(res, msg)
// 	}

// 	return res, nil
// }
