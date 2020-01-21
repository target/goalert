package calendarsubscription

import (
	"context"
	"testing"
	"time"

	"github.com/target/goalert/oncall"
	"github.com/target/goalert/permission"
)

func TestRenderICalFromShifts(t *testing.T) {
	check := func(ctx context.Context, userID string) {
		t.Run("ical", func(t *testing.T) {
			var cs CalendarSubscription
			cs.Config.ReminderMinutes = []int{5, 10}
			cs.UserID = userID
			shifts := []oncall.Shift{{UserID: cs.UserID, Start: time.Date(2020, 1, 1, 8, 0, 0, 0, time.UTC), End: time.Date(2020, 1, 15, 8, 0, 0, 0, time.UTC)}}

			_, err := cs.renderICalFromShifts(shifts)
			if err != nil {
				t.Errorf("err = %v; want nil", err)
			}
		})
	}

	ctx := permission.UserContext(context.Background(), "00000000-0000-0000-0000-000000000000", permission.RoleUser)
	check(ctx, permission.UserID(ctx))

}
