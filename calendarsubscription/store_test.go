package calendarsubscription

import (
	"context"
	"testing"
	"time"

	"github.com/target/goalert/oncall"
	"github.com/target/goalert/permission"
)

// shifts in, iCal out

func TestRenderICalFromShifts(t *testing.T) {
	check := func(ctx context.Context, schedID string, userID string, start time.Time, end time.Time) {
		t.Run("ical", func(t *testing.T) {
			// todo: Sample input arguments for now

			cs := CalendarSubscription{}

			t1, _ := time.Parse(time.RFC3339, "2020-01-01T22:08:41+00:00")
			t2, _ := time.Parse(time.RFC3339, "2020-01-07T22:08:41+00:00")

			reminderMinutes := make([]int, 2)
			reminderMinutes[0] = 5
			reminderMinutes[1] = 10

			s1 := oncall.Shift{UserID: "cb75f78a-0f7c-42fa-99f8-6b30e92a9518", Start: t1, End: t2}
			shifts := []oncall.Shift{}
			shifts = append(shifts, s1)

			_, err := cs.RenderICalFromShifts(shifts, reminderMinutes, start, end)
			if err != nil {
				t.Errorf("err = %v; want nil", err)
			}
		})
	}

	t1, _ := time.Parse(time.RFC3339, "2020-01-01T22:08:41+00:00")
	t2, _ := time.Parse(time.RFC3339, "2020-01-07T22:08:41+00:00")
	ctx := permission.UserContext(context.Background(), "cb75f78a-0f7c-42fa-99f8-6b30e92a9518", permission.RoleUser)
	check(ctx, "59aea4b0-75f0-4af3-9824-644abf8dd29a", "cb75f78a-0f7c-42fa-99f8-6b30e92a9518", t1, t2)
}
