package schedule

import (
	"encoding"
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/util/timeutil"
	"github.com/target/goalert/validation"
)

type OnCallNotificationRule struct {
	ID RuleID

	// ChannelID is the notification channel ID for notifications.
	ChannelID uuid.UUID

	Time          *timeutil.Clock
	WeekdayFilter *timeutil.WeekdayFilter
}

type RuleID struct {
	ScheduleID uuid.UUID
	ID         int
	Valid      bool
}

var _ encoding.TextMarshaler = RuleID{}
var _ encoding.TextUnmarshaler = &RuleID{}
var _ graphql.Marshaler = RuleID{}
var _ graphql.Unmarshaler = &RuleID{}

func (r RuleID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(r.String()).MarshalGQL(w)
}
func (r *RuleID) UnmarshalGQL(v interface{}) error {
	s, err := graphql.UnmarshalString(v)
	if err != nil {
		return err
	}
	err = r.UnmarshalText([]byte(s))
	if err != nil {
		return validation.WrapError(err)
	}

	return nil
}
func (r RuleID) String() string {
	if !r.Valid {
		return ""
	}

	return fmt.Sprintf("%s:%d", r.ScheduleID.String(), r.ID)
}
func (r RuleID) MarshalText() ([]byte, error) {
	if !r.Valid {
		return nil, nil
	}

	return []byte(r.String()), nil
}

func (r *RuleID) UnmarshalText(data []byte) error {
	s := string(data)
	if s == "" {
		r.Valid = false
		r.ID = 0
		r.ScheduleID = uuid.UUID{}
		return nil
	}
	if len(s) < 38 {
		return fmt.Errorf("input too short")
	}

	var err error
	r.ScheduleID, err = uuid.FromString(s[:36])
	if err != nil {
		return err
	}
	i, err := strconv.ParseInt(string(data[37:]), 10, 64)
	if err != nil {
		return err
	}
	r.ID = int(i)
	r.Valid = true

	return nil
}
