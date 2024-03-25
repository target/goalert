package smoke

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // import db driver
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/migrate"
	"github.com/target/goalert/test/smoke/harness"
	"github.com/target/goalert/test/smoke/migratetest"
	"github.com/target/goalert/util/sqlutil"
)

// DefaultSkipToMigration is the default migration to skip to when running the migration tests.
//
// It can be overriden by setting the SKIP_TO environment variable.
const DefaultSkipToMigration = "switchover-mk2"

var rules = migratetest.RuleSet{
	// All migration timestamps will differ as they applied/re-applied
	{TableName: "gorp_migrations", ColumnName: "applied_at"},

	// id will be regenerated each time the table is created
	{MigrationName: "ev3-assign-schedule-rotations", TableName: "assignments", ColumnName: "src_schedule_rule_id"},

	// id will be regenerated each time the table is created
	{MigrationName: "ev3-assign-schedule-rotations", TableName: "schedule_rules", ColumnName: "id"},

	// id will be regenerated each time the table is created
	{MigrationName: "ev3-assign-escalation-policy-steps", TableName: "escalation_policy_actions", ColumnName: "id"},

	// client_id will be regenerated each time
	{MigrationName: "ev3-notification-policy", TableName: "notification_policy_rule_state", ColumnName: "client_id"},

	// rotation names were never used before, and was always "Default Rotation". Since they now need to be unique
	// the migration uses the schedule name as a prefix, and throws away the old name.
	{MigrationName: "ev3-assign-schedule-rotations", TableName: "rotations", ColumnName: "name"},

	// rules always get recreated
	{MigrationName: "ev3-assign-schedule-rotations", TableName: "schedule_rules", ColumnName: "created_at"},

	// actions get recreated when migrating down
	{MigrationName: "ev3-assign-alert", TableName: "escalation_policy_actions", ColumnName: "id"},

	// process timestamp will change
	{MigrationName: "twilio-sms-multiple-callbacks", TableName: "notification_policy_cycles", ColumnName: "last_tick"},
	{MigrationName: "ncycle-tick", TableName: "notification_policy_cycles", ColumnName: "last_tick"},

	// migrate down should not end cycles once started
	{MigrationName: "update-existing-escalations", TableName: "notification_policy_cycles", ExtraRows: true},

	// migrate down should not clear on-call data
	{MigrationName: "update-existing-escalations", TableName: "ep_step_on_call_users", ExtraRows: true},

	// Old tables, data is safe to drop
	{MigrationName: "drop-alert-escalation-policy-snapshots", TableName: "alert_escalation_policy_snapshots", MissingRows: true},
	{MigrationName: "drop-notification-logs", TableName: "notification_logs", MissingRows: true},
	{MigrationName: "drop-sent-notifications", TableName: "sent_notifications", MissingRows: true},
	{MigrationName: "drop-user-notification-cycles", TableName: "user_notification_cycles", MissingRows: true},

	// Timestamp will change
	{MigrationName: "drop-throttle", TableName: "throttle", ColumnName: "last_action_time"},

	// End times are truncated to the minute
	{MigrationName: "schedule-rule-endtime-fix", TableName: "schedule_rules", ColumnName: "end_time"},

	// System default limits once set are not unset
	{MigrationName: "set-default-system-limits", TableName: "config_limits", ExtraRows: true},

	// Every DB must have a unique ID.
	{MigrationName: "switchover-mk2", TableName: "switchover_state", ColumnName: "db_id"},
}

const migrateInitData = `

insert into users (id, bio, email, role, name)
values
	({{uuid "u1"}}, {{text 20}}, {{text 8}}, 'admin', {{text 10}}),
	({{uuid "u2"}}, {{text 20}}, {{text 8}}, 'admin', {{text 10}});

insert into schedules (id, name, description, time_zone)
values
	({{uuid "sched1"}}, {{text 10}}, {{text 20}}, 'America/Chicago'),
	({{uuid "sched2"}}, {{text 10}}, {{text 20}}, 'America/Chicago');

insert into rotations (id, schedule_id, name, description, type)
values
	({{uuid "rot1"}}, {{uuid "sched1"}}, {{text 10}}, {{text 20}}, 'daily'),
	({{uuid "rot2"}}, {{uuid "sched2"}}, {{text 10}}, {{text 20}}, 'daily');


insert into rotation_participants (id, rotation_id, position, user_id)
values
	({{uuid "rp1"}}, {{uuid "rot1"}}, 0, {{uuid "u1"}}),
	({{uuid "rp2"}}, {{uuid "rot1"}}, 1, {{uuid "u2"}});

insert into escalation_policies (id, name, description, repeat)
values
	({{uuid "e1"}}, {{text 10}}, {{text 20}}, 1),
	({{uuid "e2"}}, {{text 10}}, {{text 20}}, 0);

insert into escalation_policy_steps (id, delay, step_number, escalation_policy_id)
values
	({{uuid "es1"}}, 1, 0, {{uuid "e1"}}),
	({{uuid "es2"}}, 1, 1, {{uuid "e1"}}),
	({{uuid "es3"}}, 1, 2, {{uuid "e1"}}),
	({{uuid "es4"}}, 1, 0, {{uuid "e2"}});

insert into escalation_policy_actions (id, escalation_policy_step_id, user_id, schedule_id)
values
	({{uuid "epa1"}}, {{uuid "es1"}}, {{uuid "u1"}}, NULL),
	({{uuid "epa2"}}, {{uuid "es1"}}, NULL, {{uuid "sched1"}});

insert into services (id, name, description, escalation_policy_id)
values
	({{uuid "s1"}}, {{text 10}}, {{text 20}}, {{uuid "e1"}}),
	({{uuid "s2"}}, {{text 10}}, {{text 20}}, {{uuid "e1"}}),
	({{uuid "s3"}}, {{text 10}}, {{text 20}}, {{uuid "e2"}});

insert into alerts (description, service_id)
values
	({{text 30}}, {{uuid "s1"}}),
	({{text 30}}, {{uuid "s1"}}),
	({{text 30}}, {{uuid "s2"}}),
	({{text 30}}, {{uuid "s2"}}),
	({{text 30}}, {{uuid "s1"}}),
	({{text 30}}, {{uuid "s3"}});

insert into user_contact_methods (id, user_id, name, type, value)
values
	({{uuid "c1"}}, {{uuid "u1"}}, {{text 8}}, 'SMS', {{phone "1"}}),
	({{uuid "c2"}}, {{uuid "u1"}}, {{text 8}}, 'VOICE', {{phone "1"}}),
	({{uuid "c3"}}, {{uuid "u2"}}, {{text 8}}, 'SMS', {{phone "2"}});

insert into user_notification_rules (id, user_id, contact_method_id, delay_minutes)
values
	({{uuid "n1"}}, {{uuid "u1"}}, {{uuid "c1"}}, 0),
	({{uuid "n2"}}, {{uuid "u1"}}, {{uuid "c2"}}, 1);

insert into user_notification_cycles (id, user_id, alert_id, escalation_level, started_at)
values
	({{uuid "ncy1"}}, {{uuid "u1"}}, 1, 0, now());

insert into sent_notifications (id, cycle_id, alert_id, contact_method_id, notification_rule_id, sent_at)
values
	({{uuid "cb1"}}, {{uuid "ncy1"}}, 1, {{uuid "c1"}}, {{uuid "n1"}}, now()),
	({{uuid "cb2"}}, {{uuid "ncy1"}}, 1, {{uuid "c2"}}, {{uuid "n2"}}, now());
`

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func renderQuery(t *testing.T, sql string) string {
	tmpl := template.New("sql")
	uuidG := harness.NewDataGen(t, "UUID", harness.DataGenFunc(harness.GenUUID))
	phoneCCG := harness.NewDataGen(t, "Phone", harness.DataGenArgFunc(harness.GenPhoneCC))
	strs := make(map[string]bool)
	tmpl.Funcs(template.FuncMap{
		"uuid":    func(id string) string { return fmt.Sprintf("'%s'", uuidG.Get(id)) },
		"phone":   func(id string) string { return fmt.Sprintf("'%s'", phoneCCG.Get(id)) },
		"phoneCC": func(cc, id string) string { return fmt.Sprintf("'%s'", phoneCCG.GetWithArg(cc, id)) },
		"text": func(n int) string {
			val := randStringRunes(n)
			for strs[val] {
				val = randStringRunes(n)
			}
			strs[val] = true
			return fmt.Sprintf("'%s'", val)
		},
	})
	_, err := tmpl.Parse(sql)
	if err != nil {
		t.Fatalf("failed to parse query template: %v", err)
	}
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, nil)
	if err != nil {
		t.Fatalf("failed to render query template: %v", err)
	}
	return b.String()
}

func TestMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping migrations tests for short mode")
	}
	t.Parallel()
	start := "atomic-escalation-policies"
	t.Logf("Starting migration testing at %s", start)

	db, err := sql.Open("pgx", harness.DBURL(""))
	if err != nil {
		t.Fatal("failed to open db:", err)
	}
	defer db.Close()
	dbName := strings.Replace("migrations_smoketest_"+time.Now().Format("2006_01_02_03_04_05")+uuid.New().String(), "-", "", -1)

	testURL := harness.DBURL(dbName)

	_, err = db.Exec("create database " + sqlutil.QuoteID(dbName))
	if err != nil {
		t.Fatal("failed to create db:", err)
	}
	defer func() { _, _ = db.Exec("drop database " + sqlutil.QuoteID(dbName)) }()

	n, err := migrate.Up(context.Background(), testURL, start)
	if err != nil {
		t.Fatal("failed to apply initial migrations:", err)
	}

	initSQL := renderQuery(t, migrateInitData)

	err = harness.ExecSQLBatch(context.Background(), testURL, initSQL)
	if err != nil {
		t.Fatalf("failed to init db %v", err)
	}

	names := migrate.Names()
	env, _ := os.LookupEnv("SKIP_TO")
	var skipTo bool
	if env != "" {
		start = env
		skipTo = true
	} else {
		start = DefaultSkipToMigration
		skipTo = true
	}
	var idx int
	for idx = range names {
		if names[idx+1] == start {
			break
		}
	}

	names = names[idx:]
	if skipTo {
		n, err := migrate.Up(context.Background(), testURL, start)
		if err != nil {
			t.Fatal("failed to apply skip migrations:", err)
		}
		if n == 0 {
			t.Fatal("SKIP_TO already applied")
		}
		t.Logf("Skipping to %s", start)
	}

	snapshot := func(t *testing.T, name string) *migratetest.Snapshot {
		t.Helper()

		snap, err := migratetest.NewSnapshotURL(context.Background(), testURL)
		require.NoError(t, err, "failed to create snapshot")
		return snap
	}

	names = names[1:]
	for i, migrationName := range names[1:] {
		lastMigrationName := names[i]
		var beforeUpSnap *migratetest.Snapshot
		pass := t.Run(migrationName, func(t *testing.T) {
			ctx := context.Background()

			if beforeUpSnap == nil {
				beforeUpSnap = snapshot(t, migrationName)
			}

			n, err = migrate.Up(ctx, testURL, migrationName)
			require.NoError(t, err, "failed to apply UP migration")
			if n == 0 {
				// no more migrations are left, so end the test
				return
			}

			afterUpSnap1 := snapshot(t, migrationName)

			_, err = migrate.Down(ctx, testURL, lastMigrationName)
			require.NoError(t, err, "failed to apply DOWN migration")

			afterDownSnap := snapshot(t, migrationName)
			pass := t.Run("Down", func(t *testing.T) {
				rules.RequireEqualDown(t, beforeUpSnap, afterDownSnap)
			})
			if !pass {
				return
			}

			_, err = migrate.Up(ctx, testURL, migrationName)
			require.NoError(t, err, "failed to apply UP migration (2nd time)")

			afterUpSnap2 := snapshot(t, migrationName)
			pass = t.Run("Up", func(t *testing.T) {
				rules.RequireEqualUp(t, afterUpSnap1, afterUpSnap2)
			})
			if !pass {
				return
			}

			beforeUpSnap = afterUpSnap2 // save for next iteration
		})
		if !pass {
			return
		}
	}
}
