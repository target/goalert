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
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
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
	NRMax                = 15    // select count(id) from user_notification_rules group by user_id order by count desc limit 1
	NRCMMax              = 11    // select count(id) from user_notification_rules group by user_id,contact_method_id order by count desc limit 1
	EPCount              = 371   // select count(id) from escalation_policies
	EPMaxStep            = 8     // select count(id) from escalation_policy_steps group by escalation_policy_id order by count desc limit 1
	EPMaxAssigned        = 19    // select count(id) from escalation_policy_actions group by escalation_policy_step_id order by count desc limit 1
	SvcCount             = 397   // select count(id) from services
	RotationMaxPart      = 50    // select count(id) from rotation_participants group by rotation_id order by count desc limit 1
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

func NewTable(ctx context.Context, conn *pgx.Conn, tx *sql.Tx, name string, cols []string, vals [][]interface{}) error {
	rows, err := conn.CopyFrom(pgx.Identifier{name}, cols, pgx.CopyFromRows(vals))
	if err != nil {
		return err
	}
	genTables++
	genRecords += rows
	return nil
}

func marshal(x string) ([]byte, error) {
	bID, err := uuid.FromString(x)
	if err != nil {
		return nil, err
	}
	id, err := bID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return id, nil
}

func fillDB() error {
	db, err := openDB()
	if err != nil {
		return errors.Wrap(err, "open DB")
	}
	defer db.Close()

	conn, err := stdlib.AcquireConn(db)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx = permission.SystemContext(ctx, "resetdb")
	start := time.Now()
	tx, err := db.BeginTx(ctx, nil)
	noErr(ctx, err)

	defer tx.Rollback()

	usrGen := newGen()
	var userIDs []string
	var users [][]interface{}

	for i := 0; i < UserCount; i++ {
		uid := gofakeit.UUID()
		userIDs = append(userIDs, uid)

		id, err := marshal(uid)
		if err != nil {
			return err
		}

		tuple := []interface{}{id, usrGen.Gen(gofakeit.Name), permission.RoleUser, usrGen.Gen(gofakeit.Email)}
		users = append(users, tuple)
	}
	err = NewTable(ctx, conn, tx, "users", []string{"id", "name", "role", "email"}, users)
	if err != nil {
		return err
	}

	p := 0
	phone := func() string {
		p++
		return fmt.Sprintf("+17633%06d", p)
	}
	var nRules [][]interface{}
	var cms [][]interface{}

	for _, userID := range userIDs {
		// For userID also
		uID, err := marshal(userID)
		if err != nil {
			return err
		}

		gen := newGen()
		typ := contactmethod.TypeSMS
		if gofakeit.Bool() {
			typ = contactmethod.TypeVoice
		}
		n := rand.Intn(CMMax)
		var cmIDs []string
		for i := 0; i < n; i++ {
			c := gofakeit.UUID()
			cmIDs = append(cmIDs, c)

			id, err := marshal(c)
			if err != nil {
				return err
			}
			tuple := []interface{}{id, gen.Gen(gofakeit.FirstName), phone(), uID, typ, true}
			cms = append(cms, tuple)
		}
		nr := 0
		nrTotal := rand.Intn(NRMax)
		for _, cmID := range cmIDs {
			nrGen := newIntGen()
			n := rand.Intn(NRCMMax) + nr
			for ; nr <= n && nr <= nrTotal; nr++ {
				n := gofakeit.UUID()
				id, err := marshal(n)
				if err != nil {
					return err
				}
				// For contact_method_id also
				cid, err := marshal(cmID)
				if err != nil {
					return err
				}
				nRules = append(nRules, []interface{}{id, nrGen.Gen(60), cid, uID})
			}
		}
	}
	err = NewTable(ctx, conn, tx, "user_contact_methods", []string{"id", "name", "value", "user_id", "type", "disabled"}, cms)
	if err != nil {
		return err
	}

	err = NewTable(ctx, conn, tx, "user_notification_rules", []string{"id", "delay_minutes", "contact_method_id", "user_id"}, nRules)
	if err != nil {
		return err
	}

	zones := []string{"America/Chicago", "Europe/Berlin", "UTC"}
	rotTypes := []rotation.Type{rotation.TypeDaily, rotation.TypeHourly, rotation.TypeWeekly}

	rotGen := newGen()
	var rotationIDs []string
	var rots [][]interface{}

	for i := 0; i < RotationCount; i++ {
		rid := gofakeit.UUID()
		rotationIDs = append(rotationIDs, rid)

		id, err := marshal(rid)
		if err != nil {
			return err
		}

		tuple := []interface{}{id,
			rotGen.Gen(idName("Rotation")),
			gofakeit.Sentence(rand.Intn(10) + 3),
			zones[rand.Intn(len(zones))],
			rand.Intn(14) + 1,
			gofakeit.DateRange(time.Now().AddDate(-3, 0, 0), time.Now()),
			rotTypes[rand.Intn(len(rotTypes))]}
		rots = append(rots, tuple)

	}
	err = NewTable(ctx, conn, tx, "rotations", []string{"id", "name", "description", "time_zone", "shift_length", "start_time", "type"}, rots)
	if err != nil {
		return err
	}

	var parts [][]interface{}
	for _, rotID := range rotationIDs {
		n := rand.Intn(RotationMaxPart)
		for i := 0; i < n; i++ {
			pid := gofakeit.UUID()
			id, err := marshal(pid)
			if err != nil {
				return err
			}

			// For rotation_id also
			rID, err := marshal(rotID)
			if err != nil {
				return err
			}
			// For user_id also
			uid := pickOneStr(userIDs)
			userID, err := marshal(uid)
			if err != nil {
				return err
			}

			tuple := []interface{}{id, rID, i, userID} //duplicates ok
			parts = append(parts, tuple)
		}
	}
	err = NewTable(ctx, conn, tx, "rotation_participants", []string{"id", "rotation_id", "position", "user_id"}, parts)
	if err != nil {
		return err
	}

	schedGen := newGen()

	var scheduleIDs []string
	var scheds [][]interface{}
	for i := 0; i < ScheduleCount; i++ {
		sid := gofakeit.UUID()
		scheduleIDs = append(scheduleIDs, sid)

		id, err := marshal(sid)
		if err != nil {
			return err
		}

		tuple := []interface{}{id,
			schedGen.Gen(idName("Schedule")),
			gofakeit.Sentence(rand.Intn(10) + 3),
			zones[rand.Intn(len(zones))]}
		scheds = append(scheds, tuple)
	}
	err = NewTable(ctx, conn, tx, "schedules", []string{"id", "name", "description", "time_zone"}, scheds)
	if err != nil {
		return err
	}

	var overrides [][]interface{}
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

			oid := gofakeit.UUID()
			id, err := marshal(oid)
			if err != nil {
				return err
			}
			// For tgt_schedule_id also
			sID, err := marshal(schedID)
			if err != nil {
				return err
			}
			var aID, rID []byte
			// For add_user_id also
			if add.Valid {
				aID, err = marshal(add.String)
				if err != nil {
					return err
				}
			}

			// For remove_user_id also
			if rem.Valid {
				rID, err = marshal(rem.String)
				if err != nil {
					return err
				}
			}
			tuple := []interface{}{id,
				sID,
				aID, rID,
				start, end}
			overrides = append(overrides, tuple)
		}
	}
	err = NewTable(ctx, conn, tx, "user_overrides",
		[]string{
			"id",
			"tgt_schedule_id",
			"add_user_id",
			"remove_user_id",
			"start_time",
			"end_time",
		}, overrides)
	if err != nil {
		return err
	}

	/*
		var rules [][]interface{}
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
				rid := gofakeit.UUID()
				id, err := marshal(rid)
				if err != nil {
					fmt.Println("Error: ", err)
				}
				// For schedule_id also
				sID, err := marshal(schedID)
				if err != nil {
					fmt.Println("Error: ", err)
				}
				var uID, rID []byte
				// For tgt_user_id also
				if usr.Valid {
					uID, err = marshal(usr.String)
					if err != nil {
						fmt.Println("Error: ", err)
					}
				}
				// For tgt_rotation_id also
				if rot.Valid {
					rID, err = marshal(rot.String)
					if err != nil {
						fmt.Println("Error: ", err)
					}
				}
				startDate, err := gofakeit.Date().MarshalBinary()
				if err != nil {
					fmt.Println("Error: ", err)
				}
				endDate, err := gofakeit.Date().MarshalBinary()
				if err != nil {
					fmt.Println("Error: ", err)
				}

				tuple := []interface{}{id,
					sID,
					gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(),
					startDate, endDate,
					uID, rID}
				rules = append(rules, tuple)
			}
		}
		_, err = NewTable(ctx, conn, tx, "schedule_rules",
			[]string{
				"id",
				"schedule_id",
				"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
				"start_time", "end_time",
				"tgt_user_id", "tgt_rotation_id",
			}, rules)
		if err != nil {
			return err
		} */

	var epIDs []string
	var eps [][]interface{}

	epGen := newGen()
	for i := 0; i < EPCount; i++ {
		eid := gofakeit.UUID()
		epIDs = append(epIDs, eid)

		id, err := marshal(eid)
		if err != nil {
			return err
		}

		tuple := []interface{}{id, epGen.Gen(idName("Policy")), gofakeit.Sentence(rand.Intn(10) + 3), rand.Intn(3)}
		eps = append(eps, tuple)
	}
	err = NewTable(ctx, conn, tx, "escalation_policies", []string{"id", "name", "description", "repeat"}, eps)
	if err != nil {
		return err
	}

	var epStepIDs []string
	var epSteps [][]interface{}

	for _, epID := range epIDs {
		n := rand.Intn(EPMaxStep)
		for i := 0; i < n; i++ {
			sid := gofakeit.UUID()
			epStepIDs = append(epStepIDs, sid)

			id, err := marshal(sid)
			if err != nil {
				return err
			}
			// For escalation_policy_id also
			eID, err := marshal(epID)
			if err != nil {
				return err
			}

			tuple := []interface{}{id,
				eID,
				i,
				rand.Intn(25) + 5}
			epSteps = append(epSteps, tuple)
		}
	}
	err = NewTable(ctx, conn, tx, "escalation_policy_steps", []string{"id", "escalation_policy_id", "step_number", "delay"}, epSteps)
	if err != nil {
		return err
	}

	var epActions [][]interface{}
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

			aid := gofakeit.UUID()
			id, err := marshal(aid)
			if err != nil {
				return err
			}

			// For escalation_policy_step_id also
			eID, err := marshal(epStepID)
			if err != nil {
				return err
			}
			var uID, sID, rID []byte
			// For user_id also
			if usr.Valid {
				uID, err = marshal(usr.String)
				if err != nil {
					return err
				}
			}
			// For schedule_id also
			if sched.Valid {
				sID, err = marshal(sched.String)
				if err != nil {
					return err
				}
			}
			// For rotation_id also
			if rot.Valid {
				rID, err = marshal(rot.String)
				if err != nil {
					return err
				}
			}

			tuple := []interface{}{id,
				eID,
				uID, sID, rID}
			epActions = append(epActions, tuple)
		}
	}
	err = NewTable(ctx, conn, tx, "escalation_policy_actions", []string{"id", "escalation_policy_step_id", "user_id", "schedule_id", "rotation_id"}, epActions)
	if err != nil {
		return err
	}

	var serviceIDs []string
	var svcs [][]interface{}

	svcGen := newGen()

	for i := 0; i < SvcCount; i++ {
		sid := gofakeit.UUID()
		serviceIDs = append(serviceIDs, sid)

		id, err := marshal(sid)
		if err != nil {
			return err
		}
		// For escalation_policy_id also
		eID, err := marshal(pickOneStr(epIDs))
		if err != nil {
			return err
		}

		tuple := []interface{}{id,
			svcGen.Gen(idName("Service")),
			gofakeit.Sentence(rand.Intn(10) + 3),
			eID}
		svcs = append(svcs, tuple)
	}
	err = NewTable(ctx, conn, tx, "services", []string{"id", "name", "description", "escalation_policy_id"}, svcs)
	if err != nil {
		return err
	}

	var iKeys [][]interface{}
	for _, serviceID := range serviceIDs {
		genIKey := newGen()
		n := rand.Intn(IntegrationKeyMax)
		for i := 0; i < n; i++ {
			typ := integrationkey.TypeGrafana
			if gofakeit.Bool() {
				typ = integrationkey.TypeGeneric
			}
			kid := gofakeit.UUID()
			id, err := marshal(kid)
			if err != nil {
				return err
			}
			// For service_id also
			sID, err := marshal(serviceID)
			if err != nil {
				return err
			}

			tuple := []interface{}{id,
				genIKey.Gen(idName("Key")),
				typ,
				sID}
			iKeys = append(iKeys, tuple)
		}
	}
	err = NewTable(ctx, conn, tx, "integration_keys", []string{"id", "name", "type", "service_id"}, iKeys)
	if err != nil {
		return err
	}

	var alerts [][]interface{}
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
		// For service_id also
		sID, err := marshal(a.ServiceID)
		if err != nil {
			return err
		}

		tuple := []interface{}{a.Summary,
			a.Details,
			a.Status,
			sID,
			a.Source,
			dedup}
		alerts = append(alerts, tuple)
	}
	err = NewTable(ctx, conn, tx, "alerts", []string{"summary", "details", "status", "service_id", "source", "dedup_key"}, alerts)
	if err != nil {
		return err
	}

	noErr(ctx, tx.Commit())
	fmt.Printf("Finished %d records across %d tables in %s\n", genRecords, genTables, time.Since(start).String())

	return nil
}

// openDB will open dbconfig.yml to detect the datasource, and attempt to open a DB connection.
func openDB() (*sql.DB, error) {
	return sql.Open("pgx", "user=goalert dbname=goalert sslmode=disable")
}

func recreateDB() error {
	db, err := sql.Open("pgx", "user=goalert dbname=postgres sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("drop database if exists goalert")
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
