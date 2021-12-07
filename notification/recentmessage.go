package notification

import (
	"context"
	"database/sql"
	"fmt"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"
)

type RecentMessage struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Type        MessageType
	Status      Status
	UserID      string
	UserName    string
	Dest        Dest
	ServiceID   string
	ServiceName string
	AlertID     int
	ProviderID  ProviderMessageID
}

var searchTemplate = template.Must(template.New("search").Parse(`
	SELECT
		m.id, m.created_at, m.last_status_at, m.message_type,
		m.last_status, m.status_details, m.provider_seq, m.src_value,
		m.user_id, u.name, m.contact_method_id, m.channel_id, cm.type, c.type, cm.value, c.value,
		m.service_id, s.name, m.alert_id, m.provider_msg_id
	FROM outgoing_messages m
	LEFT JOIN users u ON u.id = m.user_id
	LEFT JOIN services s ON s.id = m.service_id
	LEFT JOIN user_contact_methods cm ON cm.id = m.contact_method_id
	LEFT JOIN notification_channels c ON c.id = m.channel_id
	WHERE true
	{{if not .Before.IsZero}}
		AND m.created_at < :before
	{{end}}
	{{if not .After.IsZero}}
		AND m.created_at >= :after
	{{end}}
	ORDER BY m.created_at DESC
	LIMIT {{.Limit}}
`))

type RecentMessageSearchOptions struct {
	Before time.Time
	After  time.Time
	Limit  int
}

type renderData RecentMessageSearchOptions

func (opts renderData) Normalize() (*renderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	// set limit higher than normal since this is for admin use
	return &opts, validate.Range("Limit", opts.Limit, 0, 1000)
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("before", opts.Before),
		sql.Named("after", opts.After),
	}
}

func (db *DB) RecentMessages(ctx context.Context, opts *RecentMessageSearchOptions) ([]RecentMessage, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &RecentMessageSearchOptions{}
	}
	data, err := (*renderData)(opts).Normalize()
	if err != nil {
		return nil, err
	}
	query, args, err := search.RenderQuery(ctx, searchTemplate, data)
	if err != nil {
		return nil, errors.Wrap(err, "render query")
	}

	rows, err := db.db.QueryContext(ctx, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []RecentMessage
	for rows.Next() {
		var lastStatus sql.NullTime
		var srcValue, userID, userName, serviceID, serviceName, cmID, chID, providerID, cmVal, chVal sql.NullString
		var alertID sql.NullInt64
		var dstType ScannableDestType
		var msg RecentMessage
		err := rows.Scan(
			&msg.ID, &msg.CreatedAt, &lastStatus, &msg.Type,
			&msg.Status.State, &msg.Status.Details, &msg.Status.Sequence, &srcValue,
			&userID, &userName,
			&cmID, &chID, &dstType.CM, &dstType.NC, &cmVal, &chVal,
			&serviceID, &serviceName, &alertID, &providerID,
		)
		if err != nil {
			return nil, err
		}
		if lastStatus.Valid {
			msg.UpdatedAt = lastStatus.Time
		} else {
			msg.UpdatedAt = msg.CreatedAt
		}
		msg.Status.SrcValue = srcValue.String
		msg.UserID = userID.String
		msg.UserName = userName.String
		msg.ServiceID = serviceID.String
		msg.ServiceName = serviceName.String
		msg.AlertID = int(alertID.Int64)
		if providerID.Valid {
			msg.ProviderID, err = ParseProviderMessageID(providerID.String)
			if err != nil {
				log.Log(ctx, fmt.Errorf("invalid provider message id '%s': %w", providerID.String, err))
			}
		}
		msg.Dest.Type = dstType.DestType()
		switch {
		case msg.Dest.Type.IsUserCM():
			msg.Dest.ID = cmID.String
			msg.Dest.Value = cmVal.String
		default:
			msg.Dest.ID = chID.String
			msg.Dest.Value = chVal.String
		}

		msgs = append(msgs, msg)
	}

	return msgs, err
}
