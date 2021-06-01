package smoketest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/override"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/smoketest/harness"
)

func TestOverrideConflict(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values
		({{uuid "u1"}}, 'bob', 'bob@example.com');

	insert into schedules (id, name, time_zone) 
	values
		({{uuid "sid"}}, 'schedule', 'UTC');
	`

	h := harness.NewHarness(t, sql, "sched-module-v3")
	defer h.Close()

	db := h.App().DB()

	ctx := permission.SystemContext(context.Background(), "Smoketest")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tx1, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx1.Rollback()

	tx2, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx2.Rollback()

	start := time.Now().Add(-time.Hour)
	end := start.Add(8 * time.Hour)

	_, err = h.App().OverrideStore.CreateUserOverrideTx(ctx, tx1, &override.UserOverride{
		AddUserID: h.UUID("u1"),
		Target:    assignment.ScheduleTarget(h.UUID("sid")),
		Start:     start,
		End:       end,
	})
	require.NoError(t, err)

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		_, err := h.App().OverrideStore.CreateUserOverrideTx(ctx, tx2, &override.UserOverride{
			AddUserID: h.UUID("u1"),
			Target:    assignment.ScheduleTarget(h.UUID("sid")),
			Start:     start,
			End:       end,
		})
		if err != nil {
			errCh <- err
			return
		}

		errCh <- tx2.Commit()
	}()
	require.NoError(t, tx1.Commit())
	require.Error(t, <-errCh)
}
