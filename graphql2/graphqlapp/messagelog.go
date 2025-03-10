package graphqlapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"
)

type MessageLog App

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

type (
	MessageLogConnectionStats App
	TimeSeriesBucket          App
)

func (a *App) MessageLogConnectionStats() graphql2.MessageLogConnectionStatsResolver {
	return (*MessageLogConnectionStats)(a)
}

func (a *App) TimeSeriesBucket() graphql2.TimeSeriesBucketResolver {
	return (*TimeSeriesBucket)(a)
}

func (a *TimeSeriesBucket) Count(ctx context.Context, obj *graphql2.TimeSeriesBucket) (int, error) {
	return int(obj.Value), nil
}

func (q *MessageLogConnectionStats) TimeSeries(ctx context.Context, opts *notification.SearchOptions, input graphql2.TimeSeriesOptions) ([]graphql2.TimeSeriesBucket, error) {
	if opts == nil {
		opts = &notification.SearchOptions{}
	}

	dur := input.BucketDuration.TimePart()
	dur += time.Duration(input.BucketDuration.Days()) * 24 * time.Hour
	dur += time.Duration(input.BucketDuration.MonthPart) * 30 * 24 * time.Hour
	dur += time.Duration(input.BucketDuration.YearPart) * 365 * 24 * time.Hour

	var origin time.Time
	if input.BucketOrigin != nil {
		origin = *input.BucketOrigin
	}

	buckets, err := q.NotificationStore.TimeSeries(ctx, notification.TimeSeriesOpts{
		SearchOptions:      *opts,
		TimeSeriesInterval: dur,
		TimeSeriesOrigin:   origin,
	})
	if err != nil {
		return nil, err
	}

	out := make([]graphql2.TimeSeriesBucket, len(buckets))
	for i, b := range buckets {
		out[i] = graphql2.TimeSeriesBucket{
			Start: b.Start,
			End:   b.End,
			Value: float64(b.Count),
		}
	}

	return out, nil
}

func msgTypeFriendlyName(msgType gadb.EnumOutgoingMessagesType) string {
	switch msgType {
	case gadb.EnumOutgoingMessagesTypeAlertNotification:
		return "Alert"
	case gadb.EnumOutgoingMessagesTypeAlertNotificationBundle:
		return "Alert Bundle"
	case gadb.EnumOutgoingMessagesTypeAlertStatusUpdate:
		return "Status Update"
	case gadb.EnumOutgoingMessagesTypeScheduleOnCallNotification:
		return "On-Call Notification"
	case gadb.EnumOutgoingMessagesTypeSignalMessage:
		return "Signal Message"
	case gadb.EnumOutgoingMessagesTypeAlertStatusUpdateBundle:
		return "Status Bundle" // deprecated
	case gadb.EnumOutgoingMessagesTypeTestNotification:
		return "Test Message"
	case gadb.EnumOutgoingMessagesTypeVerificationMessage:
		return "Verification Code"
	}

	return fmt.Sprintf("Unknown: %s", msgType)
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
		err := validate.Range("First", *opts.First, 0, 100)
		if err != nil {
			return nil, err
		}
		searchOpts.Limit = *opts.First
	}
	if opts.CreatedAfter != nil {
		searchOpts.CreatedAfter = *opts.CreatedAfter
	}
	if opts.CreatedBefore != nil {
		searchOpts.CreatedBefore = *opts.CreatedBefore
	}
	if searchOpts.Limit == 0 {
		searchOpts.Limit = 50
	}

	searchOpts.Limit++
	logs, err := q.NotificationStore.Search(ctx, &searchOpts)
	hasNextPage := len(logs) == searchOpts.Limit
	searchOpts.Limit-- // prevent confusion later
	if err != nil {
		return nil, err
	}

	conn = new(graphql2.MessageLogConnection)
	conn.PageInfo = &graphql2.PageInfo{
		HasNextPage: hasNextPage,
	}

	if hasNextPage {
		last := logs[len(logs)-1]
		searchOpts.After.CreatedAt = last.CreatedAt
		searchOpts.After.ID = last.ID

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}

	if len(logs) > searchOpts.Limit {
		// If we have next page, we've fetched MORE than one page, but we only want to return one page.
		logs = logs[:searchOpts.Limit]
	}

	for _, _log := range logs {
		log := _log
		var dest gadb.DestV1
		switch {
		case log.ContactMethodID != uuid.Nil:
			dest, err = q.CMStore.FindDestByID(ctx, q.DB, log.ContactMethodID)
			if err != nil {
				return nil, fmt.Errorf("lookup contact method %s: %w", log.ContactMethodID, err)
			}

		case log.ChannelID != uuid.Nil:
			dest, err = q.NCStore.FindDestByID(ctx, q.DB, log.ChannelID)
			if err != nil {
				return nil, fmt.Errorf("lookup notification channel %s: %w", log.ChannelID, err)
			}
		}

		dm := graphql2.DebugMessage{
			ID:         log.ID,
			CreatedAt:  log.CreatedAt,
			UpdatedAt:  log.LastStatusAt,
			Type:       msgTypeFriendlyName(log.MessageType),
			Status:     msgStatus(notification.Status{State: log.LastStatus, Details: log.StatusDetails}),
			AlertID:    &log.AlertID,
			RetryCount: log.RetryCount,
			SentAt:     log.SentAt,
		}
		if dest.Type != "" {
			info, err := q.DestReg.DisplayInfo(ctx, dest)
			if err != nil {
				return nil, fmt.Errorf("lookup dest %s: %w", dest, err)
			}
			dm.Destination = info.Text
		}
		if log.UserID != "" {
			dm.UserID = &log.UserID
		}
		if log.UserName != "" {
			dm.UserName = &log.UserName
		}
		if log.SrcValue != "" {
			dm.Source = &log.SrcValue
		}
		if log.ServiceID != "" {
			dm.ServiceID = &log.ServiceID
		}
		if log.ServiceName != "" {
			dm.ServiceName = &log.ServiceName
		}
		if log.AlertID != 0 {
			dm.AlertID = &log.AlertID
		}
		if log.ProviderMsgID != nil {
			dm.ProviderID = &log.ProviderMsgID.ExternalID
		}

		conn.Nodes = append(conn.Nodes, dm)
	}
	conn.Stats = &searchOpts

	return conn, nil
}

func (q *Query) DebugMessages(ctx context.Context, input *graphql2.DebugMessagesInput) ([]graphql2.DebugMessage, error) {
	if input.First != nil && *input.First > 100 {
		*input.First = 100
	}
	conn, err := q.MessageLogs(ctx, &graphql2.MessageLogSearchOptions{
		CreatedBefore: input.CreatedBefore,
		CreatedAfter:  input.CreatedAfter,
		First:         input.First,
	})
	if err != nil {
		return nil, err
	}

	return conn.Nodes, nil
}
