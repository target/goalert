package smoke

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // import db driver
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/devtools/pgdump-lite"
	"github.com/target/goalert/migrate"
	"github.com/target/goalert/test/smoke/harness"
	"github.com/target/goalert/util/sqlutil"
)

type ignoreRule struct {
	MigrationName string
	TableName     string
	ColumnName    string
	ExtraRows     bool
	MissingRows   bool
}

var ignoreRules = []ignoreRule{
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

type pgDumpEntry struct {
	Name string
	Body string
}

var enumRx = regexp.MustCompile(`(?s)CREATE TYPE ([\w_.]+) AS ENUM \(\s*(.*)\s*\);`)

// enumOK handles checking for safe enum differences. This case is that migrate
// up can add, but migrate down will not remove new enum values.
//
// migrate down can't safely remove enum values, but it's safe for new ones
// to exist. So we simply check that all original items exist.
func enumOK(got, want string) bool {
	partsW := enumRx.FindStringSubmatch(want)
	if len(partsW) != 3 {
		return false
	}
	partsG := enumRx.FindStringSubmatch(got)
	if len(partsG) != 3 {
		return false
	}
	if partsW[1] != partsG[1] {
		return false
	}

	gotItems := strings.Split(partsG[2], ",\n")
	wantItems := strings.Split(partsW[2], ",\n")

	g := make(map[string]bool, len(gotItems))
	for _, v := range gotItems {
		g[strings.TrimSpace(v)] = true
	}

	for _, v := range wantItems {
		if !g[strings.TrimSpace(v)] {
			return false
		}
	}

	return true
}

func TestEnumOK(t *testing.T) {
	const got = `CREATE TYPE enum_alert_log_event AS ENUM (
'created',
'reopened',
'status_changed',
'assignment_changed',
'escalated',
'closed',
'notification_sent',
'response_received',
'acknowledged',
'policy_updated',
'duplicate_suppressed',
'escalation_request'
);`
	const want = `CREATE TYPE enum_alert_log_event AS ENUM (
'created',
'reopened',
'status_changed',
'assignment_changed',
'escalated',
'closed',
'notification_sent',
'response_received'
);`

	if !enumOK(got, want) {
		t.Errorf("got false; want true")
	}
}

func processIgnoreRules(ignoreRules []ignoreRule, name, body string) string {
	for _, r := range ignoreRules {
		if r.MigrationName != "" && r.MigrationName != name {
			continue
		}

		if !strings.HasPrefix(body, "COPY "+(pgx.Identifier{r.TableName}).Sanitize()) {
			continue
		}
		lines := strings.Split(body, "\n")
		pref, cols, suf := getCols(lines[0])
		index := -1
		for i, v := range cols {
			if v == (pgx.Identifier{r.ColumnName}).Sanitize() {
				index = i
				break
			}
		}
		if index == -1 {
			continue
		}
		newLen := len(cols) - 1
		copy(cols[index:], cols[index+1:])
		cols = cols[:newLen]
		lines[0] = pref + strings.Join(cols, ", ") + suf

		data := lines[1 : len(lines)-1]
		for i, l := range data {
			cols = strings.Split(l, "\t")
			copy(cols[index:], cols[index+1:])
			cols = cols[:newLen]
			data[i] = strings.Join(cols, "\t")
		}
		body = strings.Join(lines, "\n")
	}
	return body
}

func TestProcessIgnoreRules(t *testing.T) {
	t.Parallel()
	const input = `COPY "my_table" ("foo", "bar", "baz") FROM stdin;
1	2	3
a	b	c
\.`
	const expected = `COPY "my_table" ("foo", "baz") FROM stdin;
1	3
a	c
\.`
	rules := []ignoreRule{
		{MigrationName: "foo", TableName: "my_table", ColumnName: "bar"},
	}
	result := processIgnoreRules(rules, "foo", input)
	if result != expected {
		t.Errorf("got\n%s\n\nwant\n%s", result, expected)
	}
}

func getCols(line string) (prefix string, cols []string, suffix string) {
	cols = strings.SplitN(line, "(", 2)

	prefix = cols[0] + "("
	suffix = cols[1]
	cols = strings.SplitN(suffix, ")", 2)
	suffix = ")" + cols[1]
	cols = strings.Split(cols[0], ", ")

	return prefix, cols, suffix
}

var sqlCommentRx = regexp.MustCompile(`(?m)^--.*\n`)

func parsePGDump(data []byte, name string) []pgDumpEntry {
	// remove all comments to simplify parsing
	data = sqlCommentRx.ReplaceAll(data, nil)
	stmts := sqlutil.SplitQuery(string(data))

	entries := make([]pgDumpEntry, 0, 10000)
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		entries = append(entries, pgDumpEntry{
			Name: strings.SplitN(stmt, "\n", 2)[0],
			Body: processIgnoreRules(ignoreRules, name, stmt),
		})
	}

	return entries
}

func indent(str string) string {
	return "    " + strings.Replace(str, "\n", "\n    ", -1)
}

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

func (e pgDumpEntry) matchesBody(migrationName string, body string) bool {
	if e.Body == body {
		return true
	}
	if enumOK(body, e.Body) {
		return true
	}

	if !strings.HasPrefix(e.Body, "COPY ") {
		return false
	}

	// check for extra rows rule
	var extraRows, missingRows bool
	for _, r := range ignoreRules {
		if r.MigrationName != migrationName {
			continue
		}
		if !strings.HasPrefix(e.Name, r.TableName+";") {
			continue
		}
		extraRows = extraRows || r.ExtraRows
		missingRows = missingRows || r.MissingRows
	}
	if !extraRows && !missingRows {
		return false
	}

	e.Body = strings.TrimSuffix(e.Body, "\n\\.")
	body = strings.TrimSuffix(body, "\n\\.")
	if extraRows {
		rows := strings.Split(body, "\n")
		for i := range rows {
			if e.Body == strings.Join(rows[:len(rows)-i], "\n") {
				return true
			}
		}
	}

	if missingRows {
		rows := strings.Split(e.Body, "\n")
		for i := range rows {
			if body == strings.Join(rows[:len(rows)-i], "\n") {
				return true
			}
		}
	}

	return false
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

	_, err = db.Exec("create database " + sqlutil.QuoteID(dbName))
	if err != nil {
		t.Fatal("failed to create db:", err)
	}
	defer func() { _, _ = db.Exec("drop database " + sqlutil.QuoteID(dbName)) }()

	n, err := migrate.Up(context.Background(), harness.DBURL(dbName), start)
	if err != nil {
		t.Fatal("failed to apply initial migrations:", err)
	}

	initSQL := renderQuery(t, migrateInitData)

	err = harness.ExecSQLBatch(context.Background(), harness.DBURL(dbName), initSQL)
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
		start = "switchover-mk2" // default skip_to
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
		n, err := migrate.Up(context.Background(), harness.DBURL(dbName), start)
		if err != nil {
			t.Fatal("failed to apply skip migrations:", err)
		}
		if n == 0 {
			t.Fatal("SKIP_TO already applied")
		}
		t.Logf("Skipping to %s", start)
	}

	snapshot := func(t *testing.T, name string) []pgDumpEntry {
		// need a new connection as migrations may have changed the schema
		pgxConn, err := pgx.Connect(context.Background(), harness.DBURL(""))
		require.NoError(t, err, "failed to open db")
		defer pgxConn.Close(context.Background())

		schema, err := pgdump.DumpSchema(context.Background(), pgxConn)
		require.NoError(t, err, "failed to dump schema")

		var buf bytes.Buffer
		buf.WriteString(schema.String())

		err = pgdump.DumpData(context.Background(), pgxConn, &buf,
			// We ignore the (old) notifications table since it's trigger based
			// and always re-calculated which makes it near impossible to test
			// migrations.
			//
			// It also doesn't work properly anyhow, which is why it has been
			// replaced.
			[]string{"notifications"})
		require.NoError(t, err, "failed to dump data")

		return parsePGDump(buf.Bytes(), name)
	}
	mm := 0
	checkDiff := func(t *testing.T, typ, migrationName string, a, b []pgDumpEntry) bool {
		m1 := make(map[string]string)
		m2 := make(map[string]string)
		for _, e := range a {
			m1[e.Name] = e.Body
		}
		for _, e := range b {
			m2[e.Name] = e.Body
		}
		var mismatch bool
		for _, e := range a {
			body, ok := m2[e.Name]
			if !ok {
				mismatch = true
				t.Errorf("%s missing\n%s\n%s", typ, e.Name, indent(e.Body))
				continue
			}
			if !e.matchesBody(migrationName, body) {
				mismatch = true
				t.Errorf("%s mismatch\n%s\ngot\n%s\nwant\n%s", typ, e.Name, indent(body), indent(e.Body))
				continue
			}
		}
		for _, e := range b {
			_, ok := m1[e.Name]
			if !ok {
				mismatch = true
				t.Errorf("%s leftover\n%s\n%s", typ, e.Name, indent(e.Body))
			}
		}

		mm++
		return mismatch
	}
	names = names[1:]
	for i, migrationName := range names[1:] {
		lastMigrationName := names[i]
		var applied bool
		pass := t.Run(migrationName, func(t *testing.T) {
			ctx := context.Background()
			orig := snapshot(t, migrationName)
			n, err = migrate.Up(ctx, harness.DBURL(dbName), migrationName)
			if err != nil {
				t.Fatalf("failed to apply UP migration: %v", err)
			}
			if n == 0 {
				return
			}
			applied = true
			upSnap := snapshot(t, migrationName)
			_, err = migrate.Down(ctx, harness.DBURL(dbName), lastMigrationName)
			if err != nil {
				t.Fatalf("failed to apply DOWN migration: %v", err)
			}
			applied = false
			s := snapshot(t, migrationName)
			if checkDiff(t, "DOWN", migrationName, orig, s) {
				t.Fatalf("DOWN migration did not restore previous schema")
			}

			_, err = migrate.Up(ctx, harness.DBURL(dbName), migrationName)
			if err != nil {
				t.Fatalf("failed to apply UP migration (2nd time): %v", err)
			}
			applied = true
			s = snapshot(t, migrationName)
			if checkDiff(t, "UP", migrationName, upSnap, s) {
				t.Fatalf("UP migration did not restore previous schema")
			}
		})
		if !pass && !applied {
			n, err = migrate.Up(context.Background(), harness.DBURL(dbName), migrationName)
			if err != nil || n == 0 {
				t.Fatalf("failed to apply UP migration; abort")
			}
		}
	}
}
