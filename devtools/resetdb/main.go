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

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     cfg,
		MaxConnections: 20,
		AfterConnect: func(conn *pgx.Conn) error {
			var t pgTime
			conn.ConnInfo.RegisterDataType(pgtype.DataType{
				Value: &t,
				Name:  "time",
				OID:   1183,
			})
			return nil
		},
	})
	if err != nil {
		return errors.Wrap(err, "connect to db")
	}
	defer pool.Close()

	must := func(err error) {
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
	dt := newDepTree()
	copyFrom := func(table string, cols []string, n int, get func(int) []interface{}, deps ...string) {
		dt.Start(table)
		go func() {
			defer dt.Done(table)
			rows := make([][]interface{}, n)
			for i := 0; i < n; i++ {
				rows[i] = get(i)
			}
			s := time.Now()
			// wait for deps to finish inserting
			dt.WaitFor(deps...)
			_, err = pool.CopyFrom(pgx.Identifier{table}, cols, pgx.CopyFromRows(rows))
			must(err)
			log.Printf("inserted %d rows into %s in %s", n, table, time.Since(s).String())
		}()
	}

	copyFrom("users", []string{"id", "name", "role", "email"}, len(data.Users), func(n int) []interface{} {
		u := data.Users[n]
		return []interface{}{asUUID(u.ID), u.Name, u.Role, u.Email}
	})
	copyFrom("user_contact_methods", []string{"id", "user_id", "name", "type", "value", "disabled"}, len(data.ContactMethods), func(n int) []interface{} {
		cm := data.ContactMethods[n]
		return []interface{}{asUUID(cm.ID), asUUID(cm.UserID), cm.Name, cm.Type, cm.Value, cm.Disabled}
	}, "users")
	copyFrom("user_notification_rules", []string{"id", "user_id", "contact_method_id", "delay_minutes"}, len(data.NotificationRules), func(n int) []interface{} {
		nr := data.NotificationRules[n]
		return []interface{}{asUUID(nr.ID), asUUID(nr.UserID), asUUID(nr.ContactMethodID), nr.DelayMinutes}
	}, "user_contact_methods")
	copyFrom("rotations", []string{"id", "name", "description", "type", "shift_length", "start_time", "time_zone"}, len(data.Rotations), func(n int) []interface{} {
		r := data.Rotations[n]
		zone, _ := r.Start.Zone()
		return []interface{}{asUUID(r.ID), r.Name, r.Description, r.Type, r.ShiftLength, r.Start, zone}
	})
	copyFrom("rotation_participants", []string{"id", "rotation_id", "user_id", "position"}, len(data.RotationParts), func(n int) []interface{} {
		p := data.RotationParts[n]
		return []interface{}{asUUID(p.ID), asUUID(p.RotationID), asUUID(p.UserID), p.Pos}
	}, "rotations", "users")
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
		}, "schedules", "rotations", "users")

	copyFrom("user_overrides", []string{"id", "tgt_schedule_id", "start_time", "end_time", "add_user_id", "remove_user_id"}, len(data.Overrides), func(n int) []interface{} {
		o := data.Overrides[n]
		var schedID *[16]byte
		if o.Target.TargetType() == assignment.TargetTypeSchedule {
			schedID = asUUIDPtr(o.Target.TargetID())
		}
		addUser := asUUIDPtr(o.AddUserID)
		remUser := asUUIDPtr(o.RemoveUserID)
		return []interface{}{asUUID(o.ID), schedID, o.Start, o.End, addUser, remUser}
	}, "schedules", "users")

	copyFrom("escalation_policies", []string{"id", "name", "description", "repeat"}, len(data.EscalationPolicies), func(n int) []interface{} {
		ep := data.EscalationPolicies[n]
		return []interface{}{asUUID(ep.ID), ep.Name, ep.Description, ep.Repeat}
	})
	copyFrom("escalation_policy_steps", []string{"id", "escalation_policy_id", "step_number", "delay"}, len(data.EscalationSteps), func(n int) []interface{} {
		step := data.EscalationSteps[n]
		return []interface{}{asUUID(step.ID), asUUID(step.PolicyID), step.StepNumber, step.DelayMinutes}
	}, "escalation_policies")

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
	}, "escalation_policy_steps", "users", "schedules", "rotations")
	copyFrom("services", []string{"id", "name", "description", "escalation_policy_id"}, len(data.Services), func(n int) []interface{} {
		s := data.Services[n]
		return []interface{}{asUUID(s.ID), s.Name, s.Description, asUUID(s.EscalationPolicyID)}
	}, "escalation_policies")
	copyFrom("integration_keys", []string{"id", "service_id", "name", "type"}, len(data.IntKeys), func(n int) []interface{} {
		key := data.IntKeys[n]
		return []interface{}{asUUID(key.ID), asUUID(key.ServiceID), key.Name, key.Type}
	}, "services")
	copyFrom("heartbeat_monitors", []string{"id", "service_id", "name", "heartbeat_interval"}, len(data.Monitors), func(n int) []interface{} {
		hb := data.Monitors[n]
		return []interface{}{asUUID(hb.ID), asUUID(hb.ServiceID), hb.Name, hb.Timeout}
	}, "services")

	_, err = pool.Exec("alter table alerts disable trigger trg_enforce_alert_limit")
	must(err)
	copyFrom("alerts", []string{"status", "summary", "details", "dedup_key", "service_id", "source"}, len(data.Alerts), func(n int) []interface{} {
		a := data.Alerts[n]
		return []interface{}{a.Status, a.Summary, a.Details, a.DedupKey(), asUUID(a.ServiceID), a.Source}
	}, "services")

	dt.Wait()
	_, err = pool.Exec("alter table alerts enable trigger all")
	must(err)
	return nil
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
