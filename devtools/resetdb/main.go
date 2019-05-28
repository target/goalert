package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/migrate"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"

	"github.com/brianvoe/gofakeit"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func main() {
	rand := flag.Bool("with-rand-data", false, "Repopulates the DB with random data.")
	skipMigrate := flag.Bool("no-migrate", false, "Disables UP migration.")
	flag.Parse()
	err := doMigrations(skipMigrate)
	if err != nil {
		log.Log(context.TODO(), err)
		os.Exit(1)
	}

	if *rand {
		err := fillDB()
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
	}
}
func noErr(ctx context.Context, err error) {
	if err == nil {
		return
	}

	log.Log(ctx, errors.WithStack(err))
	os.Exit(1)
}

// Constant values for data generation
const (
	UserCount            = 1619  // select count(id) from users
	CMMax                = 7     // select count(id) from user_contact_methods group by user_id order by count desc limit 1
	NRMax                = 20    // select count(id) from user_notification_rules group by user_id order by count desc limit 1
	NRCMMax              = 11    // select count(id) from user_notification_rules group by user_id,contact_method_id order by count desc limit 1
	EPCount              = 371   // select count(id) from escalation_policies
	EPMaxStep            = 8     // select count(id) from escalation_policy_steps group by escalation_policy_id order by count desc limit 1
	EPMaxAssigned        = 19    // select count(id) from escalation_policy_actions group by escalation_policy_step_id order by count desc limit 1
	SvcCount             = 397   // select count(id) from services
	RotationMaxPart      = 64    // select count(id) from rotation_participants group by rotation_id order by count desc limit 1
	ScheduleCount        = 404   // select count(id) from schedules
	AlertClosedCount     = 76909 // select count(id) from alerts where status = 'closed'
	AlertActiveCount     = 2762  // select count(id) from alerts where status = 'triggered' or status = 'active'
	RotationCount        = 529   // select count(id) from rotations
	IntegrationKeyMax    = 11    // select count(id) from integration_keys group by service_id order by count desc limit 1
	ScheduleMaxRules     = 10    // select count(id) from schedule_rules group by schedule_id order by count desc limit 1
	ScheduleMaxOverrides = 10
)

var (
	genRecords int
	genTables  int
)

type intGen struct {
	m  map[int]bool
	mx sync.Mutex
}

func newIntGen() *intGen {
	return &intGen{
		m: make(map[int]bool),
	}
}
func (g *intGen) Gen(n int) int {
	g.mx.Lock()
	defer g.mx.Unlock()
	for {
		value := rand.Intn(n)
		if g.m[value] {
			continue
		}
		g.m[value] = true
		return value
	}
}

type gen struct {
	m  map[string]bool
	mx sync.Mutex
}

func newGen() *gen {
	return &gen{
		m: make(map[string]bool),
	}
}
func (g *gen) PickOne(s []string) string {
	return g.Gen(func() string { return pickOneStr(s) })
}
func (g *gen) Gen(fn func() string) string {
	g.mx.Lock()
	defer g.mx.Unlock()
	for {
		value := fn()
		if g.m[value] {
			continue
		}
		g.m[value] = true
		return value
	}
}

func idName(suffix string) func() string {
	return func() string {
		var res string
		for {
			res = fmt.Sprintf("%s %s %s %s", gofakeit.JobDescriptor(), gofakeit.BuzzWord(), gofakeit.JobLevel(), suffix)
			err := validate.IDName("", res)
			if err == nil {
				return res
			}
		}
	}
}

func pickOneStr(s []string) string {
	return s[rand.Intn(len(s))]
}

type table struct {
	ctx  context.Context
	stmt *sql.Stmt
	n    int
	name string
	s    time.Time
}

func NewTable(ctx context.Context, tx *sql.Tx, name string, cols []string) *table {
	stmt, err := tx.Prepare(pq.CopyIn(name, cols...))
	noErr(ctx, err)
	return &table{ctx: ctx, stmt: stmt, name: name, s: time.Now()}
}
func (t *table) Close() {
	noErr(t.ctx, t.stmt.Close())
	fmt.Printf("%s: %d records in %s\n", t.name, t.n, time.Since(t.s).String())
	genRecords += t.n
	genTables++
}
func (t *table) Insert(args ...interface{}) {
	t.n++
	_, err := t.stmt.Exec(args...)
	noErr(t.ctx, err)
}

func fillDB() error {
	db, err := openDB()
	if err != nil {
		return errors.Wrap(err, "open DB")
	}
	defer db.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx = permission.SystemContext(ctx, "resetdb")
	start := time.Now()
	tx, err := db.BeginTx(ctx, nil)
	noErr(ctx, err)

	defer tx.Rollback()

	users := NewTable(ctx, tx, "users", []string{"id", "name", "role", "email"})
	usrGen := newGen()
	var userIDs []string
	for i := 0; i < UserCount; i++ {
		id := gofakeit.UUID()
		userIDs = append(userIDs, id)
		users.Insert(id, usrGen.Gen(gofakeit.Name), permission.RoleUser, usrGen.Gen(gofakeit.Email))
	}
	users.Close()

	p := 0
	phone := func() string {
		p++
		return fmt.Sprintf("+17633%06d", p)
	}
	var nRules [][]interface{}
	cm := NewTable(ctx, tx, "user_contact_methods", []string{"id", "name", "value", "user_id", "type", "disabled"})
	for _, userID := range userIDs {
		gen := newGen()
		typ := contactmethod.TypeSMS
		if gofakeit.Bool() {
			typ = contactmethod.TypeVoice
		}
		n := rand.Intn(CMMax)
		var cmIDs []string
		for i := 0; i < n; i++ {
			id := gofakeit.UUID()
			cm.Insert(id, gen.Gen(gofakeit.FirstName), phone(), userID, typ, true)
			cmIDs = append(cmIDs, id)
		}
		nr := 0
		nrTotal := rand.Intn(NRMax)
		for _, cmID := range cmIDs {
			nrGen := newIntGen()
			n := rand.Intn(NRCMMax) + nr
			for ; nr <= n && nr <= nrTotal; nr++ {
				nRules = append(nRules, []interface{}{gofakeit.UUID(), nrGen.Gen(60), cmID, userID})
			}
		}
	}
	cm.Close()

	nr := NewTable(ctx, tx, "user_notification_rules", []string{"id", "delay_minutes", "contact_method_id", "user_id"})
	for _, rules := range nRules {
		nr.Insert(rules...)
	}
	nr.Close()

	zones := []string{"America/Chicago", "Europe/Berlin", "UTC"}
	rotTypes := []rotation.Type{rotation.TypeDaily, rotation.TypeHourly, rotation.TypeWeekly}

	rotGen := newGen()
	var rotationIDs []string
	rot := NewTable(ctx, tx, "rotations", []string{"id", "name", "description", "time_zone", "shift_length", "start_time", "type"})
	for i := 0; i < RotationCount; i++ {
		id := gofakeit.UUID()
		rot.Insert(
			id,
			rotGen.Gen(idName("Rotation")),
			gofakeit.Sentence(rand.Intn(10)+3),
			zones[rand.Intn(len(zones))],
			rand.Intn(14)+1,
			gofakeit.DateRange(time.Now().AddDate(-3, 0, 0), time.Now()),
			rotTypes[rand.Intn(len(rotTypes))],
		)
		rotationIDs = append(rotationIDs, id)
	}
	rot.Close()

	rPart := NewTable(ctx, tx, "rotation_participants", []string{"id", "rotation_id", "position", "user_id"})
	for _, rotID := range rotationIDs {
		n := rand.Intn(RotationMaxPart)
		for i := 0; i < n; i++ {
			rPart.Insert(gofakeit.UUID(), rotID, i, pickOneStr(userIDs)) //duplicates ok
		}
	}
	rPart.Close()

	schedGen := newGen()
	sc := NewTable(ctx, tx, "schedules", []string{"id", "name", "description", "time_zone"})
	var scheduleIDs []string
	for i := 0; i < ScheduleCount; i++ {
		id := gofakeit.UUID()
		sc.Insert(id,
			schedGen.Gen(idName("Schedule")),
			gofakeit.Sentence(rand.Intn(10)+3),
			zones[rand.Intn(len(zones))],
		)
		scheduleIDs = append(scheduleIDs, id)
	}
	sc.Close()

	uo := NewTable(ctx, tx, "user_overrides",
		[]string{
			"id",
			"tgt_schedule_id",
			"add_user_id",
			"remove_user_id",
			"start_time",
			"end_time",
		},
	)
	for _, schedID := range scheduleIDs {
		n := rand.Intn(ScheduleMaxOverrides)
		u := make(map[string]bool, len(userIDs))
		nextUser := func() string {
			for {
				id := pickOneStr(userIDs)
				if u[id] {
					continue
				}
				u[id] = true
				return id
			}
		}
		for i := 0; i < n; i++ {
			var add, rem sql.NullString
			if gofakeit.Bool() {
				add.Valid = true
				add.String = nextUser()
			}
			if !add.Valid || gofakeit.Bool() {
				rem.Valid = true
				rem.String = nextUser()
			}
			end := gofakeit.DateRange(time.Now(), time.Now().AddDate(0, 1, 0))
			start := gofakeit.DateRange(time.Now().AddDate(0, -1, 0), end.Add(-time.Minute))
			uo.Insert(
				gofakeit.UUID(),
				schedID,
				add, rem,
				start, end,
			)
		}
	}
	uo.Close()

	sr := NewTable(ctx, tx, "schedule_rules",
		[]string{
			"id",
			"schedule_id",
			"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
			"start_time", "end_time",
			"tgt_user_id", "tgt_rotation_id",
		})
	for _, schedID := range scheduleIDs {
		n := rand.Intn(ScheduleMaxRules)
		for i := 0; i < n; i++ {
			var usr, rot sql.NullString
			if gofakeit.Bool() {
				usr.Valid = true
				usr.String = pickOneStr(userIDs)
			} else {
				rot.Valid = true
				rot.String = pickOneStr(rotationIDs)
			}
			sr.Insert(
				gofakeit.UUID(),
				schedID,
				gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(),
				gofakeit.Date(), gofakeit.Date(),
				usr, rot,
			)
		}
	}
	sr.Close()

	var epIDs []string
	ep := NewTable(ctx, tx, "escalation_policies", []string{"id", "name", "description", "repeat"})
	epGen := newGen()
	for i := 0; i < EPCount; i++ {
		id := gofakeit.UUID()
		ep.Insert(id, epGen.Gen(idName("Policy")), gofakeit.Sentence(rand.Intn(10)+3), rand.Intn(3))
		epIDs = append(epIDs, id)
	}
	ep.Close()

	var epStepIDs []string
	eps := NewTable(ctx, tx, "escalation_policy_steps", []string{"id", "escalation_policy_id", "step_number", "delay"})
	for _, epID := range epIDs {
		n := rand.Intn(EPMaxStep)
		for i := 0; i < n; i++ {
			id := gofakeit.UUID()
			eps.Insert(
				id,
				epID,
				i,
				rand.Intn(25)+5,
			)
			epStepIDs = append(epStepIDs, id)
		}
	}
	eps.Close()

	epAct := NewTable(ctx, tx, "escalation_policy_actions", []string{"id", "escalation_policy_step_id", "user_id", "schedule_id", "rotation_id"})
	for _, epStepID := range epStepIDs {
		epActGen := newGen()
		n := rand.Intn(EPMaxAssigned)
		for i := 0; i < n; i++ {
			var usr, sched, rot sql.NullString
			switch rand.Intn(3) {
			case 0:
				usr.Valid = true
				usr.String = epActGen.PickOne(userIDs)
			case 1:
				sched.Valid = true
				sched.String = epActGen.PickOne(scheduleIDs)
			case 2:
				rot.Valid = true
				rot.String = epActGen.PickOne(rotationIDs)
			}
			epAct.Insert(
				gofakeit.UUID(),
				epStepID,
				usr, sched, rot,
			)
		}
	}
	epAct.Close()

	var serviceIDs []string
	svcGen := newGen()
	svc := NewTable(ctx, tx, "services", []string{"id", "name", "description", "escalation_policy_id"})
	for i := 0; i < SvcCount; i++ {
		id := gofakeit.UUID()
		svc.Insert(
			id,
			svcGen.Gen(idName("Service")),
			gofakeit.Sentence(rand.Intn(10)+3),
			pickOneStr(epIDs),
		)
		serviceIDs = append(serviceIDs, id)
	}
	svc.Close()

	iKey := NewTable(ctx, tx, "integration_keys", []string{"id", "name", "type", "service_id"})
	for _, serviceID := range serviceIDs {
		genIKey := newGen()
		n := rand.Intn(IntegrationKeyMax)
		for i := 0; i < n; i++ {
			typ := integrationkey.TypeGrafana
			if gofakeit.Bool() {
				typ = integrationkey.TypeGeneric
			}
			iKey.Insert(
				gofakeit.UUID(),
				genIKey.Gen(idName("Key")),
				typ,
				serviceID,
			)
		}
	}
	iKey.Close()

	aTbl := NewTable(ctx, tx, "alerts", []string{"summary", "details", "status", "service_id", "source", "dedup_key"})
	totalAlerts := AlertActiveCount + AlertClosedCount
	for i := 0; i < totalAlerts; i++ {
		a := alert.Alert{
			Summary:   gofakeit.Sentence(rand.Intn(10) + 3),
			Source:    alert.SourceGrafana,
			ServiceID: pickOneStr(serviceIDs),
			Status:    alert.StatusClosed,
		}

		if gofakeit.Bool() {
			a.Details = gofakeit.Sentence(rand.Intn(30) + 1)
		}
		if i < AlertActiveCount {
			a.Status = alert.StatusActive
		}
		if gofakeit.Bool() {
			a.Source = alert.SourceManual
		}
		var dedup *alert.DedupID
		if a.Status != alert.StatusClosed {
			dedup = a.DedupKey()
		}
		aTbl.Insert(
			a.Summary,
			a.Details,
			a.Status,
			a.ServiceID,
			a.Source,
			dedup,
		)
	}
	aTbl.Close()

	noErr(ctx, tx.Commit())
	fmt.Printf("Finished %d records across %d tables in %s\n", genRecords, genTables, time.Since(start).String())

	return nil
}

// openDB will open dbconfig.yml to detect the datasource, and attempt to open a DB connection.
func openDB() (*sql.DB, error) {
	return sql.Open("postgres", "user=goalert dbname=goalert sslmode=disable")
}

func recreateDB() error {
	db, err := sql.Open("postgres", "user=goalert dbname=postgres sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("drop database goalert")
	if err != nil {
		return err
	}
	_, err = db.Exec("create database goalert")
	return err
}

func resetDB(db *sql.DB) error {
	var err error
	if flag.Arg(0) != "" {
		_, err = migrate.Up(context.Background(), db, flag.Arg(0))
	} else {
		_, err = migrate.ApplyAll(context.Background(), db)
	}
	return err
}

func doMigrations(skipMigrate *bool) error {
	err := recreateDB()
	if err != nil {
		return errors.Wrap(err, "recreate DB")
	}

	db, err := openDB()
	if err != nil {
		return errors.Wrap(err, "open DB")
	}
	defer db.Close()

	if *skipMigrate {
		return nil
	}

	err = resetDB(db)
	if err != nil {
		return errors.Wrap(err, "perform migration after resettting")
	}
	return nil
}
