package calendarsubscription

import (
	"context"
	"testing"
	"time"

	"github.com/target/goalert/oncall"
	"github.com/target/goalert/permission"
)

func TestRenderICalFromShifts(t *testing.T) {
	check := func(ctx context.Context, schedID string, userID string) {
		t.Run("ical", func(t *testing.T) {
			var cs CalendarSubscription

			shifts := []oncall.Shift{{UserID: "cb75f78a-0f7c-42fa-99f8-6b30e92a9518", Start: time.Date(2020, 1, 1, 8, 0, 0, 0, time.UTC), End: time.Date(2020, 1, 15, 8, 0, 0, 0, time.UTC)}}

			cs.Config.ReminderMinutes = []int{5, 10}

			_, err := cs.renderICalFromShifts(shifts)
			if err != nil {
				t.Errorf("err = %v; want nil", err)
			}
		})
	}

	ctx := permission.UserContext(context.Background(), "cb75f78a-0f7c-42fa-99f8-6b30e92a9518", permission.RoleUser)
	check(ctx, "59aea4b0-75f0-4af3-9824-644abf8dd29a", "cb75f78a-0f7c-42fa-99f8-6b30e92a9518")
}
