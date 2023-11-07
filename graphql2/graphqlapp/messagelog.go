package graphqlapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"
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
	case notificationchannel.TypeSlackChan:
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

type MessageLogConnectionStats App

func (a *App) MessageLogConnectionStats() graphql2.MessageLogConnectionStatsResolver {
	return (*MessageLogConnectionStats)(a)
}

func (q *MessageLogConnectionStats) TimeSeries(
	ctx context.Context,
	opts *notification.SearchOptions,
	input graphql2.TimeSeriesOptions,
	segmentBy *graphql2.MessageLogSegmentBy,
) ([]graphql2.TimeSeriesBucket, error) {
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

	var s notification.SegmentBy
	if segmentBy != nil {
		s = notification.SegmentBy(*segmentBy)
	}
	buckets, err := q.NotificationStore.TimeSeries(ctx, notification.TimeSeriesOpts{
		SearchOptions:      *opts,
		TimeSeriesInterval: dur,
		TimeSeriesOrigin:   origin,
		SegmentBy:          s,
	})
	if err != nil {
		return nil, err
	}

	out := make([]graphql2.TimeSeriesBucket, len(buckets))
	for i, b := range buckets {
		out[i] = graphql2.TimeSeriesBucket(b)
	}

	return out, nil
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
		var dest notification.Dest
		switch {
		case log.ContactMethodID != "":
			cm, err := (*App)(q).FindOneCM(ctx, log.ContactMethodID)
			if err != nil {
				return nil, fmt.Errorf("lookup contact method %s: %w", log.ContactMethodID, err)
			}
			dest = notification.DestFromPair(cm, nil)

		case log.ChannelID != uuid.Nil:
			nc, err := (*App)(q).FindOneNC(ctx, log.ChannelID)
			if err != nil {
				return nil, fmt.Errorf("lookup notification channel %s: %w", log.ChannelID, err)
			}
			dest = notification.DestFromPair(nil, nc)
		}

		dm := graphql2.DebugMessage{
			ID:         log.ID,
			CreatedAt:  log.CreatedAt,
			UpdatedAt:  log.LastStatusAt,
			Type:       strings.TrimPrefix(log.MessageType.String(), "MessageType"),
			Status:     msgStatus(notification.Status{State: log.LastStatus, Details: log.StatusDetails}),
			AlertID:    &log.AlertID,
			RetryCount: log.RetryCount,
			SentAt:     log.SentAt,
		}
		if dest.ID != "" {
			dm.Destination, err = q.formatDest(ctx, dest)
			if err != nil {
				return nil, fmt.Errorf("format dest: %w", err)
			}
		}
		if log.UserID != "" {
			dm.UserID = &log.UserID
		}
		if log.UserName != "" {
			dm.UserName = &log.UserName
		}
		if log.SrcValue != "" {
			src, err := q.formatDest(ctx, notification.Dest{Type: dest.Type, Value: log.SrcValue})
			if err != nil {
				return nil, fmt.Errorf("format src: %w", err)
			}
			dm.Source = &src
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
