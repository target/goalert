package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/migrate"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

func main() {
	log.SetFlags(log.Lshortfile)

	rand := flag.Bool("with-rand-data", false, "Repopulates the DB with random data.")
	skipMigrate := flag.Bool("no-migrate", false, "Disables UP migration.")
	dbURL := flag.String("db-url", "user=goalert dbname=goalert sslmode=disable", "DB URL to use.")
	flag.Parse()

	cfg, err := pgx.ParseConnectionString(*dbURL)
	if err != nil {
		log.Fatal("parse config:", err)
	}

	err = doMigrations(*dbURL, cfg, skipMigrate)
	if err != nil {
		log.Fatal("apply migrations:", err)
	}

	if !*rand {
		return
	}

	err = fillDB(cfg)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

}

var (
	genRecords int
	genTables  int
)

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

func fillDB(cfg pgx.ConnConfig) error {
	s := time.Now()
	defer func() {
		log.Println("Completed in", time.Since(s))
	}()
	data := datagenConfig{}.Generate()
	log.Println("Generated random data in", time.Since(s))

	conn, err := pgx.Connect(cfg)
	if err != nil {
		return errors.Wrap(err, "connect to db")
	}
	defer conn.Close()

	var t pgTime
	conn.ConnInfo.RegisterDataType(pgtype.DataType{
		Value: &t,
		Name:  "time",
		OID:   1183,
	})

	must := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	mustExec := func(query string, args ...interface{}) {
		_, err := conn.Exec(query, args...)
		if err != nil {
			log.Fatal(err)
		}
	}
	asUUID := func(id string) (res [16]byte) {
		copy(res[:], uuid.FromStringOrNil(id).Bytes())
		return res
	}
	asUUIDPtr := func(id string) *[16]byte {
		if id == "" {
			return nil
		}
		uid := asUUID(id)
		return &uid
	}
	copyFrom := func(table string, cols []string, n int, get func(int) []interface{}) {
		s := time.Now()
		rows := make([][]interface{}, n)
		for i := 0; i < n; i++ {
			rows[i] = get(i)
		}
		_, err := conn.CopyFrom(pgx.Identifier{table}, cols, pgx.CopyFromRows(rows))
		must(err)
		log.Printf("inserted %d rows into %s in %s", n, table, time.Since(s).String())
	}

	_, _, _, _ = must, mustExec, asUUID, copyFrom

	// for _, u := range data.Users {
	// 	mustExec(`insert into users (id, name, role, email) values ($1, $2, $3, $4)`, u.ID, u.Name, u.Role, u.Email)
	// }

	// conn.Prepare("users", `insert into users (id, name, role, email) values ($1, $2, $3, $4)`)
	// b := conn.BeginBatch()
	// for _, u := range data.Users {
	// 	b.Queue("users", []interface{}{u.ID, u.Name, u.Role, u.Email}, nil, nil)
	// }
	// must(b.Send(ctx, nil))
	// must(b.Close())

	copyFrom("users", []string{"id", "name", "role", "email"}, len(data.Users), func(n int) []interface{} {
		u := data.Users[n]
		return []interface{}{asUUID(u.ID), u.Name, u.Role, u.Email}
	})
	copyFrom("user_contact_methods", []string{"id", "user_id", "name", "type", "value", "disabled"}, len(data.ContactMethods), func(n int) []interface{} {
		cm := data.ContactMethods[n]
		return []interface{}{asUUID(cm.ID), asUUID(cm.UserID), cm.Name, cm.Type, cm.Value, cm.Disabled}
	})
	copyFrom("user_notification_rules", []string{"id", "user_id", "contact_method_id", "delay_minutes"}, len(data.NotificationRules), func(n int) []interface{} {
		nr := data.NotificationRules[n]
		return []interface{}{asUUID(nr.ID), asUUID(nr.UserID), asUUID(nr.ContactMethodID), nr.DelayMinutes}
	})
	copyFrom("rotations", []string{"id", "name", "description", "type", "shift_length", "start_time", "time_zone"}, len(data.Rotations), func(n int) []interface{} {
		r := data.Rotations[n]
		zone, _ := r.Start.Zone()
		return []interface{}{asUUID(r.ID), r.Name, r.Description, r.Type, r.ShiftLength, r.Start, zone}
	})
	copyFrom("rotation_participants", []string{"id", "rotation_id", "user_id", "position"}, len(data.RotationParts), func(n int) []interface{} {
		p := data.RotationParts[n]
		return []interface{}{asUUID(p.ID), asUUID(p.RotationID), asUUID(p.UserID), p.Pos}
	})
	copyFrom("schedules", []string{"id", "name", "description", "time_zone"}, len(data.Schedules), func(n int) []interface{} {
		s := data.Schedules[n]
		return []interface{}{asUUID(s.ID), s.Name, s.Description, s.TimeZone.String()}
	})

	copyFrom("schedule_rules",
		[]string{"id", "schedule_id", "sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "start_time", "end_time", "tgt_user_id", "tgt_rotation_id"},
		len(data.ScheduleRules), func(n int) []interface{} {
			r := data.ScheduleRules[n]

			id := asUUID(r.Target.TargetID())
			var usr, rot *[16]byte
			switch r.Target.TargetType() {
			case assignment.TargetTypeRotation:
				rot = &id
			case assignment.TargetTypeUser:
				usr = &id
			}

			return []interface{}{
				asUUID(r.ID),
				asUUID(r.ScheduleID),
				r.Day(0), r.Day(1), r.Day(2), r.Day(3), r.Day(4), r.Day(5), r.Day(6),
				pgTime(r.Start),
				pgTime(r.End),
				usr,
				rot,
			}
		})

	copyFrom("user_overrides", []string{"id", "tgt_schedule_id", "start_time", "end_time", "add_user_id", "remove_user_id"}, len(data.Overrides), func(n int) []interface{} {
		o := data.Overrides[n]
		var schedID *[16]byte
		if o.Target.TargetType() == assignment.TargetTypeSchedule {
			schedID = asUUIDPtr(o.Target.TargetID())
		}
		addUser := asUUIDPtr(o.AddUserID)
		remUser := asUUIDPtr(o.RemoveUserID)
		return []interface{}{asUUID(o.ID), schedID, o.Start, o.End, addUser, remUser}
	})

	copyFrom("escalation_policies", []string{"id", "name", "description", "repeat"}, len(data.EscalationPolicies), func(n int) []interface{} {
		ep := data.EscalationPolicies[n]
		return []interface{}{asUUID(ep.ID), ep.Name, ep.Description, ep.Repeat}
	})
	copyFrom("escalation_policy_steps", []string{"id", "escalation_policy_id", "step_number", "delay"}, len(data.EscalationSteps), func(n int) []interface{} {
		step := data.EscalationSteps[n]
		return []interface{}{asUUID(step.ID), asUUID(step.PolicyID), step.StepNumber, step.DelayMinutes}
	})
	copyFrom("escalation_policy_actions", []string{"id", "escalation_policy_step_id", "user_id", "rotation_id", "schedule_id", "channel_id"}, len(data.EscalationActions), func(n int) []interface{} {
		act := data.EscalationActions[n]
		var u, r, s, c *[16]byte
		id := asUUID(act.Tgt.TargetID())
		switch act.Tgt.TargetType() {
		case assignment.TargetTypeUser:
			u = &id
		case assignment.TargetTypeRotation:
			r = &id
		case assignment.TargetTypeSchedule:
			s = &id
		case assignment.TargetTypeNotificationChannel:
			c = &id
		}
		return []interface{}{asUUID(act.ID), asUUID(act.StepID), u, r, s, c}
	})
	copyFrom("services", []string{"id", "name", "description", "escalation_policy_id"}, len(data.Services), func(n int) []interface{} {
		s := data.Services[n]
		return []interface{}{asUUID(s.ID), s.Name, s.Description, asUUID(s.EscalationPolicyID)}
	})
	copyFrom("integration_keys", []string{"id", "service_id", "name", "type"}, len(data.IntKeys), func(n int) []interface{} {
		key := data.IntKeys[n]
		return []interface{}{asUUID(key.ID), asUUID(key.ServiceID), key.Name, key.Type}
	})
	copyFrom("heartbeat_monitors", []string{"id", "service_id", "name", "heartbeat_interval"}, len(data.Monitors), func(n int) []interface{} {
		hb := data.Monitors[n]
		return []interface{}{asUUID(hb.ID), asUUID(hb.ServiceID), hb.Name, hb.Timeout}
	})

	copyFrom("alerts", []string{"status", "summary", "details", "dedup_key", "service_id", "source"}, len(data.Alerts), func(n int) []interface{} {
		a := data.Alerts[n]
		return []interface{}{a.Status, a.Summary, a.Details, a.DedupKey(), asUUID(a.ServiceID), a.Source}
	})
	// rows := make([][]interface{}, len(data.Users))
	// for i, u := range data.Users {
	// 	rows[i] = []interface{}{asUUID(u.ID), u.Name, u.Role, u.Email}
	// }
	// _, err = conn.CopyFrom(pgx.Identifier{"users"}, []string{"id", "name", "role", "email"}, pgx.CopyFromRows(rows))
	// must(err)

	_ = data
	return nil

	// db, err := openDB()
	// if err != nil {
	// 	return errors.Wrap(err, "open DB")
	// }
	// defer db.Close()

	// conn, err := stdlib.AcquireConn(db)
	// if err != nil {
	// 	return err
	// }

	// ctx := context.Background()
	// ctx, cancel := context.WithCancel(ctx)
	// defer cancel()

	// ctx = permission.SystemContext(ctx, "resetdb")
	// start := time.Now()
	// tx, err := db.BeginTx(ctx, nil)
	// noErr(ctx, err)

	// defer tx.Rollback()

	// usrGen := newGen()
	// var userIDs []string
	// var users [][]interface{}

	// for i := 0; i < UserCount; i++ {
	// 	uid := gofakeit.UUID()
	// 	userIDs = append(userIDs, uid)

	// 	id, err := marshal(uid)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	tuple := []interface{}{id, usrGen.Gen(gofakeit.Name), permission.RoleUser, usrGen.Gen(gofakeit.Email)}
	// 	users = append(users, tuple)
	// }
	// err = NewTable(ctx, conn, tx, "users", []string{"id", "name", "role", "email"}, users)
	// if err != nil {
	// 	return err
	// }

	// p := 0
	// phone := func() string {
	// 	p++
	// 	return fmt.Sprintf("+17633%06d", p)
	// }
	// var nRules [][]interface{}
	// var cms [][]interface{}

	// for _, userID := range userIDs {
	// 	// For userID also
	// 	uID, err := marshal(userID)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	gen := newGen()
	// 	typ := contactmethod.TypeSMS
	// 	if gofakeit.Bool() {
	// 		typ = contactmethod.TypeVoice
	// 	}
	// 	n := rand.Intn(CMMax)
	// 	var cmIDs []string
	// 	for i := 0; i < n; i++ {
	// 		c := gofakeit.UUID()
	// 		cmIDs = append(cmIDs, c)

	// 		id, err := marshal(c)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		tuple := []interface{}{id, gen.Gen(gofakeit.FirstName), phone(), uID, typ, true}
	// 		cms = append(cms, tuple)
	// 	}
	// 	nr := 0
	// 	nrTotal := rand.Intn(NRMax)
	// 	for _, cmID := range cmIDs {
	// 		nrGen := newIntGen()
	// 		n := rand.Intn(NRCMMax) + nr
	// 		for ; nr <= n && nr <= nrTotal; nr++ {
	// 			n := gofakeit.UUID()
	// 			id, err := marshal(n)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			// For contact_method_id also
	// 			cid, err := marshal(cmID)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			nRules = append(nRules, []interface{}{id, nrGen.Gen(60), cid, uID})
	// 		}
	// 	}
	// }
	// err = NewTable(ctx, conn, tx, "user_contact_methods", []string{"id", "name", "value", "user_id", "type", "disabled"}, cms)
	// if err != nil {
	// 	return err
	// }

	// err = NewTable(ctx, conn, tx, "user_notification_rules", []string{"id", "delay_minutes", "contact_method_id", "user_id"}, nRules)
	// if err != nil {
	// 	return err
	// }

	// zones := []string{"America/Chicago", "Europe/Berlin", "UTC"}
	// rotTypes := []rotation.Type{rotation.TypeDaily, rotation.TypeHourly, rotation.TypeWeekly}

	// rotGen := newGen()
	// var rotationIDs []string
	// var rots [][]interface{}

	// for i := 0; i < RotationCount; i++ {
	// 	rid := gofakeit.UUID()
	// 	rotationIDs = append(rotationIDs, rid)

	// 	id, err := marshal(rid)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	tuple := []interface{}{id,
	// 		rotGen.Gen(idName("Rotation")),
	// 		gofakeit.Sentence(rand.Intn(10) + 3),
	// 		zones[rand.Intn(len(zones))],
	// 		rand.Intn(14) + 1,
	// 		gofakeit.DateRange(time.Now().AddDate(-3, 0, 0), time.Now()),
	// 		rotTypes[rand.Intn(len(rotTypes))]}
	// 	rots = append(rots, tuple)

	// }
	// err = NewTable(ctx, conn, tx, "rotations", []string{"id", "name", "description", "time_zone", "shift_length", "start_time", "type"}, rots)
	// if err != nil {
	// 	return err
	// }

	// var parts [][]interface{}
	// for _, rotID := range rotationIDs {
	// 	n := rand.Intn(RotationMaxPart)
	// 	for i := 0; i < n; i++ {
	// 		pid := gofakeit.UUID()
	// 		id, err := marshal(pid)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		// For rotation_id also
	// 		rID, err := marshal(rotID)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		// For user_id also
	// 		uid := pickOneStr(userIDs)
	// 		userID, err := marshal(uid)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		tuple := []interface{}{id, rID, i, userID} //duplicates ok
	// 		parts = append(parts, tuple)
	// 	}
	// }
	// err = NewTable(ctx, conn, tx, "rotation_participants", []string{"id", "rotation_id", "position", "user_id"}, parts)
	// if err != nil {
	// 	return err
	// }

	// schedGen := newGen()

	// var scheduleIDs []string
	// var scheds [][]interface{}
	// for i := 0; i < ScheduleCount; i++ {
	// 	sid := gofakeit.UUID()
	// 	scheduleIDs = append(scheduleIDs, sid)

	// 	id, err := marshal(sid)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	tuple := []interface{}{id,
	// 		schedGen.Gen(idName("Schedule")),
	// 		gofakeit.Sentence(rand.Intn(10) + 3),
	// 		zones[rand.Intn(len(zones))]}
	// 	scheds = append(scheds, tuple)
	// }
	// err = NewTable(ctx, conn, tx, "schedules", []string{"id", "name", "description", "time_zone"}, scheds)
	// if err != nil {
	// 	return err
	// }

	// var overrides [][]interface{}
	// for _, schedID := range scheduleIDs {
	// 	n := rand.Intn(ScheduleMaxOverrides)
	// 	u := make(map[string]bool, len(userIDs))
	// 	nextUser := func() string {
	// 		for {
	// 			id := pickOneStr(userIDs)
	// 			if u[id] {
	// 				continue
	// 			}
	// 			u[id] = true
	// 			return id
	// 		}
	// 	}
	// 	for i := 0; i < n; i++ {
	// 		var add, rem sql.NullString
	// 		if gofakeit.Bool() {
	// 			add.Valid = true
	// 			add.String = nextUser()
	// 		}
	// 		if !add.Valid || gofakeit.Bool() {
	// 			rem.Valid = true
	// 			rem.String = nextUser()
	// 		}
	// 		end := gofakeit.DateRange(time.Now(), time.Now().AddDate(0, 1, 0))
	// 		start := gofakeit.DateRange(time.Now().AddDate(0, -1, 0), end.Add(-time.Minute))

	// 		oid := gofakeit.UUID()
	// 		id, err := marshal(oid)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		// For tgt_schedule_id also
	// 		sID, err := marshal(schedID)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		var aID, rID []byte
	// 		// For add_user_id also
	// 		if add.Valid {
	// 			aID, err = marshal(add.String)
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}

	// 		// For remove_user_id also
	// 		if rem.Valid {
	// 			rID, err = marshal(rem.String)
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}
	// 		tuple := []interface{}{id,
	// 			sID,
	// 			aID, rID,
	// 			start, end}
	// 		overrides = append(overrides, tuple)
	// 	}
	// }
	// err = NewTable(ctx, conn, tx, "user_overrides",
	// 	[]string{
	// 		"id",
	// 		"tgt_schedule_id",
	// 		"add_user_id",
	// 		"remove_user_id",
	// 		"start_time",
	// 		"end_time",
	// 	}, overrides)
	// if err != nil {
	// 	return err
	// }

	// /*
	// 	var rules [][]interface{}
	// 	for _, schedID := range scheduleIDs {
	// 		n := rand.Intn(ScheduleMaxRules)
	// 		for i := 0; i < n; i++ {
	// 			var usr, rot sql.NullString
	// 			if gofakeit.Bool() {
	// 				usr.Valid = true
	// 				usr.String = pickOneStr(userIDs)
	// 			} else {
	// 				rot.Valid = true
	// 				rot.String = pickOneStr(rotationIDs)
	// 			}
	// 			rid := gofakeit.UUID()
	// 			id, err := marshal(rid)
	// 			if err != nil {
	// 				fmt.Println("Error: ", err)
	// 			}
	// 			// For schedule_id also
	// 			sID, err := marshal(schedID)
	// 			if err != nil {
	// 				fmt.Println("Error: ", err)
	// 			}
	// 			var uID, rID []byte
	// 			// For tgt_user_id also
	// 			if usr.Valid {
	// 				uID, err = marshal(usr.String)
	// 				if err != nil {
	// 					fmt.Println("Error: ", err)
	// 				}
	// 			}
	// 			// For tgt_rotation_id also
	// 			if rot.Valid {
	// 				rID, err = marshal(rot.String)
	// 				if err != nil {
	// 					fmt.Println("Error: ", err)
	// 				}
	// 			}
	// 			startDate, err := gofakeit.Date().MarshalBinary()
	// 			if err != nil {
	// 				fmt.Println("Error: ", err)
	// 			}
	// 			endDate, err := gofakeit.Date().MarshalBinary()
	// 			if err != nil {
	// 				fmt.Println("Error: ", err)
	// 			}

	// 			tuple := []interface{}{id,
	// 				sID,
	// 				gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(), gofakeit.Bool(),
	// 				startDate, endDate,
	// 				uID, rID}
	// 			rules = append(rules, tuple)
	// 		}
	// 	}
	// 	_, err = NewTable(ctx, conn, tx, "schedule_rules",
	// 		[]string{
	// 			"id",
	// 			"schedule_id",
	// 			"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
	// 			"start_time", "end_time",
	// 			"tgt_user_id", "tgt_rotation_id",
	// 		}, rules)
	// 	if err != nil {
	// 		return err
	// 	} */

	// var epIDs []string
	// var eps [][]interface{}

	// epGen := newGen()
	// for i := 0; i < EPCount; i++ {
	// 	eid := gofakeit.UUID()
	// 	epIDs = append(epIDs, eid)

	// 	id, err := marshal(eid)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	tuple := []interface{}{id, epGen.Gen(idName("Policy")), gofakeit.Sentence(rand.Intn(10) + 3), rand.Intn(3)}
	// 	eps = append(eps, tuple)
	// }
	// err = NewTable(ctx, conn, tx, "escalation_policies", []string{"id", "name", "description", "repeat"}, eps)
	// if err != nil {
	// 	return err
	// }

	// var epStepIDs []string
	// var epSteps [][]interface{}

	// for _, epID := range epIDs {
	// 	n := rand.Intn(EPMaxStep)
	// 	for i := 0; i < n; i++ {
	// 		sid := gofakeit.UUID()
	// 		epStepIDs = append(epStepIDs, sid)

	// 		id, err := marshal(sid)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		// For escalation_policy_id also
	// 		eID, err := marshal(epID)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		tuple := []interface{}{id,
	// 			eID,
	// 			i,
	// 			rand.Intn(25) + 5}
	// 		epSteps = append(epSteps, tuple)
	// 	}
	// }
	// err = NewTable(ctx, conn, tx, "escalation_policy_steps", []string{"id", "escalation_policy_id", "step_number", "delay"}, epSteps)
	// if err != nil {
	// 	return err
	// }

	// var epActions [][]interface{}
	// for _, epStepID := range epStepIDs {
	// 	epActGen := newGen()
	// 	n := rand.Intn(EPMaxAssigned)
	// 	for i := 0; i < n; i++ {
	// 		var usr, sched, rot sql.NullString
	// 		switch rand.Intn(3) {
	// 		case 0:
	// 			usr.Valid = true
	// 			usr.String = epActGen.PickOne(userIDs)
	// 		case 1:
	// 			sched.Valid = true
	// 			sched.String = epActGen.PickOne(scheduleIDs)
	// 		case 2:
	// 			rot.Valid = true
	// 			rot.String = epActGen.PickOne(rotationIDs)
	// 		}

	// 		aid := gofakeit.UUID()
	// 		id, err := marshal(aid)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		// For escalation_policy_step_id also
	// 		eID, err := marshal(epStepID)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		var uID, sID, rID []byte
	// 		// For user_id also
	// 		if usr.Valid {
	// 			uID, err = marshal(usr.String)
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}
	// 		// For schedule_id also
	// 		if sched.Valid {
	// 			sID, err = marshal(sched.String)
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}
	// 		// For rotation_id also
	// 		if rot.Valid {
	// 			rID, err = marshal(rot.String)
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}

	// 		tuple := []interface{}{id,
	// 			eID,
	// 			uID, sID, rID}
	// 		epActions = append(epActions, tuple)
	// 	}
	// }
	// err = NewTable(ctx, conn, tx, "escalation_policy_actions", []string{"id", "escalation_policy_step_id", "user_id", "schedule_id", "rotation_id"}, epActions)
	// if err != nil {
	// 	return err
	// }

	// var serviceIDs []string
	// var svcs [][]interface{}

	// svcGen := newGen()

	// for i := 0; i < SvcCount; i++ {
	// 	sid := gofakeit.UUID()
	// 	serviceIDs = append(serviceIDs, sid)

	// 	id, err := marshal(sid)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// For escalation_policy_id also
	// 	eID, err := marshal(pickOneStr(epIDs))
	// 	if err != nil {
	// 		return err
	// 	}

	// 	tuple := []interface{}{id,
	// 		svcGen.Gen(idName("Service")),
	// 		gofakeit.Sentence(rand.Intn(10) + 3),
	// 		eID}
	// 	svcs = append(svcs, tuple)
	// }
	// err = NewTable(ctx, conn, tx, "services", []string{"id", "name", "description", "escalation_policy_id"}, svcs)
	// if err != nil {
	// 	return err
	// }

	// var iKeys [][]interface{}
	// for _, serviceID := range serviceIDs {
	// 	genIKey := newGen()
	// 	n := rand.Intn(IntegrationKeyMax)
	// 	for i := 0; i < n; i++ {
	// 		typ := integrationkey.TypeGrafana
	// 		if gofakeit.Bool() {
	// 			typ = integrationkey.TypeGeneric
	// 		}
	// 		kid := gofakeit.UUID()
	// 		id, err := marshal(kid)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		// For service_id also
	// 		sID, err := marshal(serviceID)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		tuple := []interface{}{id,
	// 			genIKey.Gen(idName("Key")),
	// 			typ,
	// 			sID}
	// 		iKeys = append(iKeys, tuple)
	// 	}
	// }
	// err = NewTable(ctx, conn, tx, "integration_keys", []string{"id", "name", "type", "service_id"}, iKeys)
	// if err != nil {
	// 	return err
	// }

	// var alerts [][]interface{}
	// totalAlerts := AlertActiveCount + AlertClosedCount
	// for i := 0; i < totalAlerts; i++ {
	// 	a := alert.Alert{
	// 		Summary:   gofakeit.Sentence(rand.Intn(10) + 3),
	// 		Source:    alert.SourceGrafana,
	// 		ServiceID: pickOneStr(serviceIDs),
	// 		Status:    alert.StatusClosed,
	// 	}

	// 	if gofakeit.Bool() {
	// 		a.Details = gofakeit.Sentence(rand.Intn(30) + 1)
	// 	}
	// 	if i < AlertActiveCount {
	// 		a.Status = alert.StatusActive
	// 	}
	// 	if gofakeit.Bool() {
	// 		a.Source = alert.SourceManual
	// 	}
	// 	var dedup *alert.DedupID
	// 	if a.Status != alert.StatusClosed {
	// 		dedup = a.DedupKey()
	// 	}
	// 	// For service_id also
	// 	sID, err := marshal(a.ServiceID)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	tuple := []interface{}{a.Summary,
	// 		a.Details,
	// 		a.Status,
	// 		sID,
	// 		a.Source,
	// 		dedup}
	// 	alerts = append(alerts, tuple)
	// }
	// err = NewTable(ctx, conn, tx, "alerts", []string{"summary", "details", "status", "service_id", "source", "dedup_key"}, alerts)
	// if err != nil {
	// 	return err
	// }

	// noErr(ctx, tx.Commit())
	// fmt.Printf("Finished %d records across %d tables in %s\n", genRecords, genTables, time.Since(start).String())

	// return nil
}

// openDB will open dbconfig.yml to detect the datasource, and attempt to open a DB connection.
func openDB() (*sql.DB, error) {
	return sql.Open("pgx", "user=goalert dbname=goalert sslmode=disable")
}

func recreateDB(cfg pgx.ConnConfig) error {
	conn, err := pgx.Connect(cfg)
	if err != nil {
		return errors.Wrap(err, "connect to DB")
	}
	defer conn.Close()

	_, err = conn.Exec("drop database if exists goalert")
	if err != nil {
		return err
	}
	_, err = conn.Exec("create database goalert")
	return err
}

func resetDB(url string) error {
	var err error
	if flag.Arg(0) != "" {
		_, err = migrate.Up(context.Background(), url, flag.Arg(0))
	} else {
		_, err = migrate.ApplyAll(context.Background(), url)
	}
	return err
}

func doMigrations(url string, cfg pgx.ConnConfig, skipMigrate *bool) error {
	cfg.Database = "postgres"
	err := recreateDB(cfg)
	if err != nil {
		return errors.Wrap(err, "recreate DB")
	}

	if *skipMigrate {
		return nil
	}

	err = resetDB(url)
	if err != nil {
		return errors.Wrap(err, "perform migration after resettting")
	}
	return nil
}
