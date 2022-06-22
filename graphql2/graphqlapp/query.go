package graphqlapp

import (
	context "context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/service"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

type (
	Query        App
	DebugMessage App
)

func (a *App) Query() graphql2.QueryResolver { return (*Query)(a) }

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

func (a *Query) DebugMessages(ctx context.Context, input *graphql2.DebugMessagesInput) ([]graphql2.DebugMessage, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	var msgs []*struct {
		ID           string
		CreatedAt    time.Time
		LastStatusAt time.Time
		MessageType  notification.MessageType

		LastStatus    notification.State
		StatusDetails string
		SrcValue      string

		UserID string
		User   *user.User `gorm:"foreignkey:ID;references:UserID"`

		ContactMethodID string
		ContactMethod   *contactmethod.ContactMethod `gorm:"foreignKey:ID;references:ContactMethodID"`

		ChannelID string
		Channel   *notificationchannel.Channel `gorm:"foreignKey:ID;references:ChannelID"`

		ServiceID string
		Service   *service.Service `gorm:"foreignKey:ID;references:ServiceID"`

		AlertID       int
		ProviderMsgID *notification.ProviderMessageID
	}

	db := sqlutil.FromContext(ctx).Table("outgoing_messages")

	if input.CreatedAfter != nil {
		db = db.Where("created_at >= ?", *input.CreatedAfter)
	}
	if input.CreatedBefore != nil {
		db = db.Where("created_at < ?", *input.CreatedBefore)
	}
	if input.First != nil {
		err = validate.Range("first", *input.First, 0, 1000)
		if err != nil {
			return nil, err
		}
		db = db.Limit(*input.First)
	} else {
		db = db.Limit(search.DefaultMaxResults)
	}

	err = db.
		Preload("User", sqlutil.Columns("ID", "Name")).
		Preload("Service", sqlutil.Columns("ID", "Name")).
		Preload("Channel", sqlutil.Columns("ID", "Type", "Value")).
		Preload("ContactMethod", sqlutil.Columns("ID", "Type", "Value")).
		Order("created_at DESC").
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}

	var res []graphql2.DebugMessage
	for _, m := range msgs {
		dst := notification.DestFromPair(m.ContactMethod, m.Channel)
		destStr, err := a.formatDest(ctx, dst)
		if err != nil {
			return nil, fmt.Errorf("format dest: %w", err)
		}

		// notifications that end up bundled are omitted
		if m.MessageType == notification.MessageTypeAlertBundle {
			return nil, nil
		}

		msg := graphql2.DebugMessage{
			ID:          m.ID,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.LastStatusAt,
			Type:        strings.TrimPrefix(m.MessageType.String(), "MessageType"),
			Status:      msgStatus(notification.Status{State: m.LastStatus, Details: m.StatusDetails}),
			Destination: destStr,
		}
		if m.User != nil {
			msg.UserID = &m.User.ID
			msg.UserName = &m.User.Name
		}

		if m.SrcValue != "" && m.ContactMethod != nil {
			src, err := a.formatDest(ctx, notification.Dest{Type: dst.Type, Value: m.SrcValue})
			if err != nil {
				return nil, fmt.Errorf("format src: %w", err)
			}
			msg.Source = &src
		}
		if m.Service != nil {
			msg.ServiceID = &m.Service.ID
			msg.ServiceName = &m.Service.Name
		}
		if m.AlertID != 0 {
			msg.AlertID = &m.AlertID
		}
		if m.ProviderMsgID != nil {
			msg.ProviderID = &m.ProviderMsgID.ExternalID
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
