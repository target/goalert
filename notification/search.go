package notification

import (
	"context"
	"database/sql"
	"sort"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type MessageLog struct {
	ID           string
	CreatedAt    time.Time
	LastStatusAt time.Time
	MessageType  MessageType

	LastStatus    State
	StatusDetails string
	SrcValue      string

	AlertID       int
	ProviderMsgID *ProviderMessageID

	UserID   string
	UserName string

	ContactMethodID string

	ChannelID uuid.UUID

	ServiceID   string
	ServiceName string

	SentAt     *time.Time
	RetryCount int
}

// SearchOptions allow filtering and paginating the list of messages.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	CreatedAfter  time.Time `json:"ca,omitempty"`
	CreatedBefore time.Time `json:"cb,omitempty"`

	// Omit specifies a list of message IDs to exclude from the results
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	ID        string    `json:"i,omitempty"`
	CreatedAt time.Time `json:"n,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Funcs(search.Helpers()).Parse(`
	{{if .TimeSeriesOpts}}
	SELECT
		(trunc((extract('epoch' from om.created_at)-:timeSeriesOrigin)/:timeSeriesInterval))::bigint AS bucket,
		count(*)
		{{if .SegmentByService}}
		, s.name
		{{else if .SegmentByUser}}
		, u.name
		{{else if .SegmentByMessageType}}
		, om.message_type
		{{ else }}
		, ''
		{{end}}
	{{else}}
	SELECT
		om.id, om.created_at, om.last_status_at, om.message_type, om.last_status, om.status_details,
		om.src_value, om.alert_id, om.provider_msg_id,
		om.user_id, u.name, om.contact_method_id, om.channel_id, om.service_id, s.name,
		om.sent_at, om.retry_count
	{{end}}
	FROM outgoing_messages om
	LEFT JOIN users u ON om.user_id = u.id
	LEFT JOIN services s ON om.service_id = s.id
	LEFT JOIN user_contact_methods cm ON om.contact_method_id = cm.id
	LEFT JOIN notification_channels nc ON om.channel_id = nc.id
	WHERE true
	{{if .Omit}}
		AND NOT om.id = any(:omit)
	{{end}}
	{{if not .CreatedAfter.IsZero}}
		AND om.created_at >= :createdAfter
	{{end}}
	{{if not .CreatedBefore.IsZero}}
		AND om.created_at < :createdBefore
	{{end}}
	{{if .Search}}
		AND (
			{{orderedPrefixSearch "search" "u.name"}}
			OR
			{{orderedPrefixSearch "search" "s.name"}}
			OR
				cm.value ILIKE '%' || :search || '%'
			OR
				nc.name ILIKE '%' || :search || '%'
			OR
				lower(cm.type::text) = lower(:search)
			OR
				lower(nc.type::text) = lower(:search)
		)
	{{end}}
	{{if .After.ID}}
		AND om.created_at < :cursorCreatedAt
		OR (om.created_at = :cursorCreatedAt AND om.id > :afterID)
	{{end}}
		AND om.last_status != 'bundled'

	{{if .SegmentByService}}
	GROUP BY bucket, om.created_at, s.name

	{{else if .SegmentByUser}}
	GROUP BY bucket, om.created_at, u.name

	{{else if .SegmentByMessageType}}
	GROUP BY bucket, om.created_at, om.message_type

	{{else if .TimeSeriesOpts}}
	GROUP BY bucket
	
	{{else}}
	ORDER BY om.last_status = 'pending' desc, coalesce(om.sent_at, om.last_status_at) desc, om.created_at desc, om.id asc

	LIMIT {{.Limit}}
	{{end}}
`))

type renderData struct {
	SearchOptions
	*TimeSeriesOpts
}

func (opts renderData) SegmentByService() bool {
	return opts.TimeSeriesOpts != nil && opts.TimeSeriesOpts.SegmentBy == SegmentByService
}

func (opts renderData) SegmentByUser() bool {
	return opts.TimeSeriesOpts != nil && opts.TimeSeriesOpts.SegmentBy == SegmentByUser
}

func (opts renderData) SegmentByMessageType() bool {
	return opts.TimeSeriesOpts != nil && opts.TimeSeriesOpts.SegmentBy == SegmentByMessageType
}

func (opts renderData) Normalize() (*renderData, error) {
	if opts.Limit == 0 {
		opts.Limit = 50
	}

	err := validate.Many(
		validate.Search("Search", opts.Search),
		// should be 1 more than the expected limit
		validate.Range("Limit", opts.Limit, 0, 101),
		validate.ManyUUID("Omit", opts.Omit, 50),
	)
	if err != nil {
		return nil, err
	}

	if opts.TimeSeriesOpts != nil {
		opts.TimeSeriesInterval = opts.TimeSeriesInterval.Truncate(time.Second)
		if opts.CreatedBefore.IsZero() {
			return nil, validation.NewFieldError("CreatedBefore", "required for time series queries")
		}
		if opts.CreatedAfter.IsZero() {
			return nil, validation.NewFieldError("CreatedAfter", "required for time series queries")
		}

		diff := opts.CreatedBefore.Sub(opts.CreatedAfter)
		minInterval := diff/1000 + 1
		minInterval = minInterval.Truncate(time.Minute)
		err = validate.Duration("TimeSeriesInterval", opts.TimeSeriesInterval, minInterval, time.Hour*24*365)
	}
	if err != nil {
		return nil, err
	}

	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	var tsOrigin int64
	var tsInterval int64
	if opts.TimeSeriesOpts != nil {
		tsOrigin = opts.TimeSeriesOrigin.Unix()
		tsInterval = int64(opts.TimeSeriesInterval.Seconds())
	}

	return []sql.NamedArg{
		sql.Named("search", opts.Search),
		sql.Named("cursorCreatedAt", opts.After.CreatedAt),
		sql.Named("createdAfter", opts.CreatedAfter),
		sql.Named("afterID", opts.After.ID),
		sql.Named("createdBefore", opts.CreatedBefore),
		sql.Named("omit", sqlutil.UUIDArray(opts.Omit)),
		sql.Named("timeSeriesOrigin", tsOrigin),
		sql.Named("timeSeriesInterval", tsInterval),
	}
}

type TimeSeriesOpts struct {
	SearchOptions
	TimeSeriesOrigin   time.Time
	TimeSeriesInterval time.Duration
	SegmentBy          SegmentBy
}

type TimeSeriesBucket struct {
	Start        time.Time
	End          time.Time
	Count        int
	SegmentLabel string
}

type bucketID struct {
	Bucket int
	Label  string
}

type SegmentBy string

const SegmentByService SegmentBy = "service"
const SegmentByUser SegmentBy = "user"
const SegmentByMessageType SegmentBy = "messageType"

// TimeSeries returns a list of time series buckets for the given search options.
func (s *Store) TimeSeries(ctx context.Context, opts TimeSeriesOpts) ([]TimeSeriesBucket, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return nil, err
	}

	data := &renderData{
		SearchOptions:  opts.SearchOptions,
		TimeSeriesOpts: &opts,
	}
	data, err = data.Normalize()
	if err != nil {
		return nil, err
	}

	query, args, err := search.RenderQuery(ctx, searchTemplate, data)
	if err != nil {
		return nil, errors.Wrap(err, "render query")
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bucketCounts := make(map[bucketID]int)
	for rows.Next() {
		var bucket, count int
		var segmentLabel sql.NullString

		err := rows.Scan(&bucket, &count, &segmentLabel)
		if err != nil {
			return nil, err
		}

		bucketCounts[bucketID{
			Bucket: bucket,
			Label:  segmentLabel.String,
		}] = count
	}
	ts := makeTimeSeries(data.CreatedAfter, data.CreatedBefore, data.TimeSeriesOrigin, data.TimeSeriesInterval, bucketCounts)
	return ts, nil
}

func timeToIndex(origin time.Time, interval time.Duration, t time.Time) int {
	return int(t.Sub(origin) / interval)
}

// makeTimeSeries steps through the start and end time given, using respective the duration.
//
// This is to ensure a time series with a bucket for every time interval is returned,
// even if there are no counts for that interval for a particular label.
func makeTimeSeries(start, end, origin time.Time, duration time.Duration, buckets map[bucketID]int) []TimeSeriesBucket {
	var tsb []TimeSeriesBucket
	for t := start; t.Before(end); t = t.Add(duration) {
		for bID := range buckets {
			var b TimeSeriesBucket
			b.SegmentLabel = bID.Label
			b.Start = t
			b.End = t.Add(duration)

			// gets the count from original buckets
			// returning 0 if nothing present for current time range
			b.Count = buckets[bucketID{
				Bucket: timeToIndex(origin, duration, t),
				Label:  bID.Label,
			}]
			tsb = append(tsb, b)
		}
	}

	// keep names in alphabetical order for graph labels
	sort.SliceStable(tsb, func(i, j int) bool {
		return tsb[i].SegmentLabel < tsb[j].SegmentLabel
	})

	return tsb
}

func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]MessageLog, error) {
	if opts == nil {
		opts = &SearchOptions{}
	}

	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return nil, err
	}

	data := &renderData{SearchOptions: *opts}
	data, err = data.Normalize()
	if err != nil {
		return nil, err
	}

	query, args, err := search.RenderQuery(ctx, searchTemplate, data)
	if err != nil {
		return nil, errors.Wrap(err, "render query")
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MessageLog
	for rows.Next() {
		var l MessageLog
		var alertID sql.NullInt64
		var retryCount sql.NullInt32
		var chanID sqlutil.NullUUID
		var serviceID, svcName sql.NullString
		var srcValue sql.NullString
		var userID, userName sql.NullString
		var cmID sql.NullString
		var providerID sql.NullString
		var lastStatusAt, sentAt sql.NullTime
		err = rows.Scan(
			&l.ID,
			&l.CreatedAt,
			&lastStatusAt,
			&l.MessageType,
			&l.LastStatus,
			&l.StatusDetails,
			&srcValue,
			&alertID,
			&providerID,
			&userID,
			&userName,
			&cmID,
			&chanID,
			&serviceID,
			&svcName,
			&sentAt,
			&retryCount,
		)
		if err != nil {
			return nil, err
		}

		// set all the nullable fields
		if providerID.String != "" {
			pm, err := ParseProviderMessageID(providerID.String)
			if err != nil {
				return nil, err
			}
			l.ProviderMsgID = &pm
		}
		l.AlertID = int(alertID.Int64)
		l.ChannelID = chanID.UUID
		l.ServiceID = serviceID.String
		l.ServiceName = svcName.String
		l.SrcValue = srcValue.String
		l.UserID = userID.String
		l.UserName = userName.String
		l.ContactMethodID = cmID.String
		l.LastStatusAt = lastStatusAt.Time
		if sentAt.Valid {
			l.SentAt = &sentAt.Time
		}
		l.RetryCount = int(retryCount.Int32)

		result = append(result, l)
	}

	return result, nil
}
