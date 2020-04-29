package harness

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/alert"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/config"
	"github.com/target/goalert/devtools/mockslack"
	"github.com/target/goalert/devtools/mocktwilio"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/notificationrule"
	"github.com/target/goalert/util/sqlutil"
)

const dbTimeFormat = "2006-01-02 15:04:05.999999-07:00"

var (
	dbURLStr string
	dbURL    *url.URL
)

func init() {
	dbURLStr = os.Getenv("DB_URL")
	if dbURLStr == "" {
		dbURLStr = "postgres://goalert@127.0.0.1:5432?sslmode=disable"
	}
	var err error
	dbURL, err = url.Parse(dbURLStr)
	if err != nil {
		panic(err)
	}
}

func DBURL(name string) string {
	if name == "" {
		return dbURLStr
	}
	u := *dbURL
	u.Path = "/" + url.PathEscape(name)
	return u.String()
}

// Harness is a helper for smoketests. It deals with assertions, database management, and backend monitoring during tests.
type Harness struct {
	phoneG, phoneCCG, uuidG *DataGen
	t                       *testing.T
	closing                 bool

	tw  *twilioAssertionAPI
	twS *httptest.Server

	cfg config.Config

	slack     *slackServer
	slackS    *httptest.Server
	slackApp  mockslack.AppInfo
	slackUser mockslack.UserInfo

	ignoreErrors []string

	cmd *exec.Cmd

	backendURL  string
	dbURL       string
	dbName      string
	delayOffset time.Duration
	mx          sync.Mutex

	start          time.Time
	resumed        time.Time
	lastTimeChange time.Time
	pgResume       time.Time

	db       *pgxpool.Pool
	dbStdlib *sql.DB
	nr       notificationrule.Store
	a        alert.Store

	usr                user.Store
	userGeneratedIndex int
	addGraphUser       sync.Once

	sessToken string
	sessKey   keyring.Keyring
	authH     *auth.Handler
}

func (h *Harness) Config() config.Config {
	return h.cfg
}

// NewHarness will create a new database, perform `migrateSteps` migrations, inject `initSQL` and return a new Harness bound to
// the result. It starts a backend process pre-configured to a mock twilio server for monitoring notifications as well.
func NewHarness(t *testing.T, initSQL, migrationName string) *Harness {
	t.Helper()
	h := NewStoppedHarness(t, initSQL, migrationName)
	h.Start()
	return h
}

// NewHarnessDebugDB works like NewHarness, but fails the test immediately after
// migrations have been run. It is used to debug data & queries from a smoketest.
//
// Note that the now() function will be locked to the init timestamp for inspection.
func NewHarnessDebugDB(t *testing.T, initSQL, migrationName string) *Harness {
	t.Helper()
	h := NewStoppedHarness(t, initSQL, migrationName)
	h.Migrate("")

	t.Fatal("DEBUG DB ::", h.dbURL)
	return nil
}

const (
	twilioAuthToken  = "11111111111111111111111111111111"
	twilioAccountSID = "AC00000000000000000000000000000000"
	mailgunAPIKey    = "key-00000000000000000000000000000000"
)

// NewStoppedHarness will create a NewHarness, but will not call Start.
func NewStoppedHarness(t *testing.T, initSQL, migrationName string) *Harness {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping Harness tests for short mode")
	}

	t.Logf("Using DB URL: %s", dbURL)
	start := time.Now()
	name := strings.Replace("smoketest_"+time.Now().Format("2006_01_02_15_04_05")+uuid.NewV4().String(), "-", "", -1)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, DBURL(""))
	if err != nil {
		t.Fatal("connect to db:", err)
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, "create database "+sqlutil.QuoteID(name))
	if err != nil {
		t.Fatal("create db:", err)
	}
	conn.Close(ctx)

	t.Logf("created test database '%s': %s", name, dbURL)
	g := NewDataGen(t, "Phone", DataGenFunc(GenPhone))

	twCfg := mocktwilio.Config{
		AuthToken:    twilioAuthToken,
		AccountSID:   twilioAccountSID,
		MinQueueTime: 100 * time.Millisecond, // until we have a stateless backend for answering calls
	}

	h := &Harness{
		phoneG:         g,
		uuidG:          NewDataGen(t, "UUID", DataGenFunc(GenUUID)),
		phoneCCG:       NewDataGen(t, "PhoneCC", DataGenArgFunc(GenPhoneCC)),
		dbName:         name,
		dbURL:          DBURL(name),
		lastTimeChange: start,
		start:          start,

		t: t,
	}

	h.tw = newTwilioAssertionAPI(func() {
		h.FastForward(time.Minute)
		h.Trigger()
	}, mocktwilio.NewServer(twCfg), h.cfg.Twilio.FromNumber)

	h.twS = httptest.NewServer(h.tw)

	h.cmd = exec.Command(
		"goalert",
		"-l", "localhost:0",
		"-v",
		"--db-url", h.dbURL,
		"--json",
		"--twilio-base-url", h.twS.URL,
		"--db-max-open", "5", // 2 for API 1 for engine
	)
	h.cmd.Env = os.Environ()

	// freeze DB time until backend starts
	h.execQuery(`
		create schema testing_overrides;
		alter database ` + sqlutil.QuoteID(name) + ` set search_path = "$user", public,testing_overrides, pg_catalog;
		

		create or replace function testing_overrides.now()
		returns timestamp with time zone
		as $$
			begin
			return '` + start.Format(dbTimeFormat) + `';
			end;
		$$ language plpgsql;
	`)

	h.Migrate(migrationName)
	h.initSlack()
	h.execQuery(initSQL)

	return h
}

func (h *Harness) Start() {
	h.t.Helper()
	r, err := h.cmd.StderrPipe()
	if err != nil {
		h.t.Fatalf("failed to get pipe for backend logs: %v", err)
	}

	var cfg config.Config
	cfg.General.DisableV1GraphQL = true
	cfg.Slack.Enable = true
	cfg.Slack.AccessToken = h.slackApp.AccessToken
	cfg.Slack.ClientID = h.slackApp.ClientID
	cfg.Slack.ClientSecret = h.slackApp.ClientSecret
	cfg.Twilio.Enable = true
	cfg.Twilio.AccountSID = twilioAccountSID
	cfg.Twilio.AuthToken = twilioAuthToken
	cfg.Twilio.FromNumber = h.phoneG.Get("twilio")

	cfg.Mailgun.Enable = true
	cfg.Mailgun.APIKey = mailgunAPIKey
	cfg.Mailgun.EmailDomain = "smoketest.example.com"
	h.cfg = cfg
	data, err := json.Marshal(cfg)
	if err != nil {
		h.t.Fatalf("failed to marshal config: %v", err)
	}

	cfgCmd := exec.Command("goalert", "migrate", "--db-url="+h.dbURL)
	cfgCmd.Stdin = bytes.NewReader(data)
	out, err := cfgCmd.CombinedOutput()
	if err != nil {
		h.t.Fatalf("failed to migrate backend: %v\n%s", err, string(out))
	}
	cfgCmd = exec.Command("goalert", "set-config", "--allow-empty-data-encryption-key", "--db-url="+h.dbURL)
	cfgCmd.Stdin = bytes.NewReader(data)
	out, err = cfgCmd.CombinedOutput()
	if err != nil {
		h.t.Fatalf("failed to config backend: %v\n%s", err, string(out))
	}

	dbCfg, err := pgx.ParseConfig(h.dbURL)
	if err != nil {
		h.t.Fatalf("failed to parse db url: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(h.dbURL)
	if err != nil {
		h.t.Fatalf("failed to parse db url: %v", err)
	}
	poolCfg.MaxConns = 2

	h.db, err = pgxpool.ConnectConfig(ctx, poolCfg)
	if err != nil {
		h.t.Fatalf("failed to connect to db: %v", err)
	}

	// resume the flow of time
	err = h.db.QueryRow(ctx, `select pg_catalog.now()`).Scan(&h.pgResume)
	if err != nil {
		h.t.Fatalf("failed to get postgres timestamp: %v", err)
	}
	h.resumed = time.Now()
	h.lastTimeChange = time.Now().Add(100 * time.Millisecond)
	h.modifyDBOffset(0)

	h.cmd.Env = append(h.cmd.Env, "GOALERT_SLACK_BASE_URL="+h.slackS.URL)
	err = h.cmd.Start()
	if err != nil {
		h.t.Fatalf("failed to start backend: %v", err)
	}
	urlCh := make(chan string)
	go h.watchBackend(r)
	go h.watchBackendLogs(r, urlCh)

	h.backendURL = <-urlCh

	err = h.tw.RegisterSMSCallback(h.phoneG.Get("twilio"), h.backendURL+"/v1/twilio/sms/messages")
	if err != nil {
		h.t.Fatalf("failed to init twilio (SMS callback): %v", err)
	}
	err = h.tw.RegisterVoiceCallback(h.phoneG.Get("twilio"), h.backendURL+"/v1/twilio/voice/call")
	if err != nil {
		h.t.Fatalf("failed to init twilio (voice callback): %v", err)
	}

	h.dbStdlib = stdlib.OpenDB(*dbCfg)
	h.nr, err = notificationrule.NewDB(ctx, h.dbStdlib)
	if err != nil {
		h.t.Fatalf("failed to init notification rule backend: %v", err)
	}
	h.usr, err = user.NewDB(ctx, h.dbStdlib)
	if err != nil {
		h.t.Fatalf("failed to init user backend: %v", err)
	}

	aLog, err := alertlog.NewDB(ctx, h.dbStdlib)
	if err != nil {
		h.t.Fatalf("failed to init alert log backend: %v", err)
	}

	h.a, err = alert.NewDB(ctx, h.dbStdlib, aLog)
	if err != nil {
		h.t.Fatalf("failed to init alert backend: %v", err)
	}

	h.sessKey, err = keyring.NewDB(ctx, h.dbStdlib, &keyring.Config{
		Name: "browser-sessions",
	})
	if err != nil {
		h.t.Fatalf("failed to init keyring: %v", err)
	}

	h.authH, err = auth.NewHandler(ctx, h.dbStdlib, auth.HandlerConfig{
		SessionKeyring: h.sessKey,
	})
	if err != nil {
		h.t.Fatalf("failed to init auth handler: %v", err)
	}
}

// URL returns the backend server's URL
func (h *Harness) URL() string {
	return h.backendURL
}

// Migrate will perform `steps` number of migrations.
func (h *Harness) Migrate(migrationName string) {
	h.t.Helper()
	h.t.Logf("Running migrations (target: %s)", migrationName)
	data, err := exec.Command("goalert", "migrate", "--db-url", h.dbURL, "--up", migrationName).CombinedOutput()
	if err != nil {
		h.t.Log(string(data))
		h.t.Fatalf("failed to run migration: %v", err)
	}
}

// IgnoreErrorsWith will cause the Harness to ignore backend errors containing the specified substring.
func (h *Harness) IgnoreErrorsWith(substr string) {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.ignoreErrors = append(h.ignoreErrors, substr)
}

func (h *Harness) modifyDBOffset(d time.Duration) {
	n := time.Now()
	d -= n.Sub(h.lastTimeChange)
	if n.After(h.lastTimeChange) {
		h.lastTimeChange = n
	}

	h.delayOffset += d

	h.setDBOffset(h.delayOffset)
}
func (h *Harness) setDBOffset(d time.Duration) {
	elapsed := time.Since(h.resumed)
	h.t.Logf("Updating DB time offset to: %s (+ %s elapsed = %s since test start)", h.delayOffset.String(), elapsed.String(), (h.delayOffset + elapsed).String())

	h.execQuery(fmt.Sprintf(`
		create or replace function testing_overrides.now()
		returns timestamp with time zone
		as $$
			begin
			return cast('%s' as timestamp with time zone) + (pg_catalog.now() - cast('%s' as timestamp with time zone))::interval;
			end;
		$$ language plpgsql;
	`,
		h.start.Add(d).Format(dbTimeFormat),
		h.pgResume.Format(dbTimeFormat),
	))
	h.trigger()
}

// Delay will forward time (with regard to the database). It currently
// performs Sleep, but should be treated as a DB modification.
func (h *Harness) Delay(d time.Duration) {
	h.t.Helper()
	h.t.Logf("Wait %s", d.String())
	time.Sleep(d)
	h.trigger()
}

func (h *Harness) FastForward(d time.Duration) {
	h.t.Helper()
	h.t.Logf("Fast-forward %s", d.String())
	h.delayOffset += d
	h.setDBOffset(h.delayOffset)
}

func (h *Harness) execQuery(sql string) {
	h.t.Helper()
	t := template.New("sql")
	t.Funcs(template.FuncMap{
		"uuid":    func(id string) string { return fmt.Sprintf("'%s'", h.uuidG.Get(id)) },
		"phone":   func(id string) string { return fmt.Sprintf("'%s'", h.phoneG.Get(id)) },
		"phoneCC": func(cc, id string) string { return fmt.Sprintf("'%s'", h.phoneCCG.GetWithArg(cc, id)) },

		"slackChannelID": func(name string) string { return fmt.Sprintf("'%s'", h.Slack().Channel(name).ID()) },
	})
	_, err := t.Parse(sql)
	if err != nil {
		h.t.Fatalf("failed to parse query template: %v", err)
	}

	b := new(bytes.Buffer)
	err = t.Execute(b, nil)
	if err != nil {
		h.t.Fatalf("failed to render query template: %v", err)
	}

	err = ExecSQLBatch(context.Background(), h.dbURL, b.String())
	if err != nil {
		h.t.Fatalf("failed to exec query: %v", err)
	}
}

// CreateAlert will create one or more unacknowledged alerts for a service.
func (h *Harness) CreateAlert(serviceID string, summary ...string) {
	h.t.Helper()

	permission.SudoContext(context.Background(), func(ctx context.Context) {
		h.t.Helper()
		tx, err := h.dbStdlib.BeginTx(ctx, nil)
		if err != nil {
			h.t.Fatalf("failed to start tx: %v", err)
		}
		defer tx.Rollback()
		for _, sum := range summary {
			a := &alert.Alert{
				ServiceID: serviceID,
				Summary:   sum,
			}

			h.t.Logf("insert alert: %v", a)
			_, isNew, err := h.a.CreateOrUpdateTx(ctx, tx, a)
			if err != nil {
				h.t.Fatalf("failed to insert alert: %v", err)
			}
			if !isNew {
				h.t.Fatal("could not create duplicate alert with summary: " + sum)
			}
		}
		err = tx.Commit()
		if err != nil {
			h.t.Fatalf("failed to commit tx: %v", err)
		}
	})
	h.trigger()
}

// CreateManyAlert will create multiple new unacknowledged alerts for a given service.
func (h *Harness) CreateManyAlert(serviceID, summary string) {
	h.t.Helper()
	a := &alert.Alert{
		ServiceID: serviceID,
		Summary:   summary,
	}
	h.t.Logf("insert alert: %v", a)
	permission.SudoContext(context.Background(), func(ctx context.Context) {
		h.t.Helper()
		_, err := h.a.Create(ctx, a)
		if err != nil {
			h.t.Fatalf("failed to insert alert: %v", err)
		}
	})
	h.trigger()
}

// AddNotificationRule will add a notification rule to the database.
func (h *Harness) AddNotificationRule(userID, cmID string, delayMinutes int) {
	h.t.Helper()
	nr := &notificationrule.NotificationRule{
		DelayMinutes:    delayMinutes,
		UserID:          userID,
		ContactMethodID: cmID,
	}
	h.t.Logf("insert notification rule: %v", nr)
	permission.SudoContext(context.Background(), func(ctx context.Context) {
		h.t.Helper()
		_, err := h.nr.Insert(ctx, nr)
		if err != nil {
			h.t.Fatalf("failed to insert notification rule: %v", err)
		}
	})
	h.trigger()
}

// Trigger will trigger, and wait for, an engine cycle.
func (h *Harness) Trigger() {
	go h.trigger()

	// wait for the next cycle to start and end before returning
	http.Get(h.backendURL + "/health/engine")
}

// Escalate will escalate an alert in the database, when 'level' matches.
func (h *Harness) Escalate(alertID, level int) {
	h.t.Helper()
	h.t.Logf("escalate alert #%d (from level %d)", alertID, level)
	permission.SudoContext(context.Background(), func(ctx context.Context) {
		err := h.a.Escalate(ctx, alertID, level)
		if err != nil {
			h.t.Fatalf("failed to escalate alert: %v", err)
		}
	})
	h.trigger()
}

// Phone will return the generated phone number for the id provided.
func (h *Harness) Phone(id string) string { return h.phoneG.Get(id) }

// PhoneCC will return the generated phone number for the id provided.
func (h *Harness) PhoneCC(cc, id string) string { return h.phoneCCG.GetWithArg(cc, id) }

// UUID will return the generated UUID for the id provided.
func (h *Harness) UUID(id string) string { return h.uuidG.Get(id) }

func (h *Harness) isClosing() bool {
	h.mx.Lock()
	defer h.mx.Unlock()
	return h.closing
}

func (h *Harness) dumpDB() {
	testName := reflect.ValueOf(h.t).Elem().FieldByName("name").String()
	file := filepath.Join("smoketest_db_dump", testName+".sql")
	os.MkdirAll(filepath.Dir(file), 0755)
	var t time.Time
	err := h.db.QueryRow(context.Background(), "select now()").Scan(&t)
	if err != nil {
		h.t.Fatalf("failed to get current timestamp: %v", err)
	}
	err = exec.Command(
		"pg_dump",
		"-O", "-x", "-a",
		"-f", file,
		h.dbURL,
	).Run()
	if err != nil {
		h.t.Errorf("failed to dump database '%s': %v", h.dbName, err)
	}
	fd, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		h.t.Fatalf("failed to open DB dump: %v", err)
	}
	defer fd.Close()
	_, err = fmt.Fprintf(fd, "\n-- Last Timestamp: %s\n", t.Format(time.RFC3339Nano))
	if err != nil {
		h.t.Fatalf("failed to open DB dump: %v", err)
	}
}

// Close terminates any background processes, and drops the testing database.
// It should be called at the end of all tests (usually with `defer h.Close()`).
func (h *Harness) Close() error {
	h.t.Helper()
	h.tw.WaitAndAssert(h.t)
	h.slack.WaitAndAssert()
	h.slackS.Close()
	h.twS.Close()
	h.mx.Lock()
	h.closing = true
	h.mx.Unlock()
	if h.cmd.Process != nil {
		h.cmd.Process.Kill()
		h.cmd.Process.Wait()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	h.sessKey.Shutdown(ctx)
	h.tw.Close()
	h.dumpDB()

	h.dbStdlib.Close()
	h.db.Close()

	conn, err := pgx.Connect(ctx, DBURL(""))
	if err != nil {
		h.t.Error("failed to connect to DB:", err)
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, "drop database "+sqlutil.QuoteID(h.dbName))
	if err != nil {
		h.t.Errorf("failed to drop database '%s': %v", h.dbName, err)
	}

	return nil
}

// CreateUser generates a random user.
func (h *Harness) CreateUser() (u *user.User) {
	h.t.Helper()
	var err error
	permission.SudoContext(context.Background(), func(ctx context.Context) {
		u, err = h.usr.Insert(ctx, &user.User{
			Name:  fmt.Sprintf("Generated%d", h.userGeneratedIndex),
			ID:    uuid.NewV4().String(),
			Role:  permission.RoleUser,
			Email: fmt.Sprintf("generated%d@example.com", h.userGeneratedIndex),
		})
	})
	if err != nil {
		h.t.Fatal(errors.Wrap(err, "generate random user"))
	}
	h.userGeneratedIndex++
	return u
}

// WaitAndAssertOnCallUsers will ensure the correct set of users as on-call for the given serviceID.
func (h *Harness) WaitAndAssertOnCallUsers(serviceID string, userIDs ...string) {
	h.t.Helper()
	doQL := func(query string, res interface{}) {
		g := h.GraphQLQuery2(query)
		for _, err := range g.Errors {
			h.t.Error("GraphQL Error:", err.Message)
		}
		if len(g.Errors) > 0 {
			h.t.Fatal("errors returned from GraphQL")
		}
		if res == nil {
			return
		}
		err := json.Unmarshal(g.Data, &res)
		if err != nil {
			h.t.Fatal("failed to parse response:", err)
		}
	}

	getUsers := func() []string {
		var result struct {
			Service struct {
				OnCallUsers []struct {
					UserID   string
					UserName string
				}
			}
		}

		doQL(fmt.Sprintf(`
			query{
				service(id: "%s"){
					onCallUsers{
						userID
						userName
					}
				}
			}
		`, serviceID), &result)

		var ids []string
		for _, oc := range result.Service.OnCallUsers {
			ids = append(ids, oc.UserID)
		}
		if len(ids) == 0 {
			return nil
		}
		sort.Strings(ids)
		uniq := ids[:1]
		last := ids[0]
		for _, id := range ids[1:] {
			if id == last {
				continue
			}
			uniq = append(uniq, id)
			last = id
		}
		return uniq
	}
	sort.Strings(userIDs)
	match := func(final bool) bool {
		ids := getUsers()
		if len(ids) != len(userIDs) {
			if final {
				h.t.Fatalf("got %d on-call users; want %d", len(ids), len(userIDs))
			}
			return false
		}
		for i, id := range userIDs {
			if ids[i] != id {
				if final {
					h.t.Fatalf("on-call[%d] = %s; want %s", i, ids[i], id)
				}
				return false
			}
		}
		return true
	}

	h.Trigger() // run engine cycle

	match(true) // assert result
}
