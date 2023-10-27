package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/migrate"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/util/timeutil"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

var adminID string

func main() {
	log.SetFlags(log.Lshortfile)
	flag.StringVar(&adminID, "admin-id", "", "Generate an admin user with the given ID.")
	seedVal := flag.Int64("seed", 1, "Change the random seed used to generate data.")
	mult := flag.Float64("mult", 1, "Multiply base type counts (e.g., alerts, users, services).")
	genData := flag.Bool("with-rand-data", false, "Repopulates the DB with random data.")
	skipMigrate := flag.Bool("no-migrate", false, "Disables UP migration.")
	skipDrop := flag.Bool("skip-drop", false, "Skip database drop/create step.")
	adminURL := flag.String("admin-db-url", "postgres://goalert@localhost/postgres", "Admin DB URL to use (used to recreate DB).")
	dbURL := flag.String("db-url", "postgres://goalert@localhost", "DB URL to use.")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := pgx.ParseConfig(*dbURL)
	if err != nil {
		log.Fatal("parse db-url:", err)
	}
	dbName := cfg.Database
	if dbName == "" {
		dbName = cfg.User
	}

	if !*skipDrop {
		err = recreateDB(ctx, *adminURL, dbName)
		if err != nil {
			log.Fatal("recreate DB:", err)
		}
	}

	if *skipMigrate {
		return
	}

	s := time.Now()
	n, err := resetDB(ctx, *dbURL)
	if err != nil {
		log.Fatalln("apply migrations:", err)
	}
	log.Printf("applied %d migrations in %s", n, time.Since(s).String())

	if !*genData {
		return
	}
	dataCfg := &datagenConfig{AdminID: adminID, Seed: *seedVal}
	dataCfg.SetDefaults()
	dataCfg.Multiply(*mult)
	err = fillDB(ctx, dataCfg, *dbURL)
	if err != nil {
		log.Fatal("insert random data:", err)
	}
}

func fillDB(ctx context.Context, dataCfg *datagenConfig, url string) error {
	s := time.Now()
	defer func() {
		log.Println("Completed in", time.Since(s))
	}()
	data := dataCfg.Generate()
	log.Println("Generated random data in", time.Since(s))

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return err
	}
	cfg.MaxConns = 20

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "connect to db")
	}
	defer pool.Close()

	must := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	asTime := func(c timeutil.Clock) (t pgtype.Time) {
		t.Status = pgtype.Present
		t.Microseconds = time.Duration(c).Microseconds()
		return t
	}
	asUUID := func(id string) [16]byte { return uuid.MustParse(id) }
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
			_, err = pool.CopyFrom(ctx, pgx.Identifier{table}, cols, pgx.CopyFromRows(rows))
			must(errors.Wrap(err, table))
			log.Printf("inserted %d rows into %s in %s", n, table, time.Since(s).String())
		}()
	}

	copyFrom("users", []string{"id", "name", "role", "email"}, len(data.Users), func(n int) []interface{} {
		u := data.Users[n]
		return []interface{}{asUUID(u.ID), u.Name, u.Role, u.Email}
	})
	copyFrom("user_contact_methods", []string{"id", "user_id", "name", "type", "value", "disabled", "pending"}, len(data.ContactMethods), func(n int) []interface{} {
		cm := data.ContactMethods[n]
		return []interface{}{asUUID(cm.ID), asUUID(cm.UserID), cm.Name, cm.Type, cm.Value, cm.Disabled, cm.Pending}
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
				asTime(r.Start),
				asTime(r.End),
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
	copyFrom("labels", []string{"tgt_service_id", "key", "value"}, len(data.Labels), func(n int) []interface{} {
		lbl := data.Labels[n]
		var svc *[16]byte
		id := asUUID(lbl.Target.TargetID())
		switch lbl.Target.TargetType() {
		case assignment.TargetTypeService:
			svc = &id
		}
		return []interface{}{svc, lbl.Key, lbl.Value}
	}, "services")
	copyFrom("heartbeat_monitors", []string{"id", "service_id", "name", "heartbeat_interval"}, len(data.Monitors), func(n int) []interface{} {
		hb := data.Monitors[n]
		return []interface{}{asUUID(hb.ID), asUUID(hb.ServiceID), hb.Name, hb.Timeout}
	}, "services")
	copyFrom("user_favorites", []string{"user_id", "tgt_service_id", "tgt_schedule_id", "tgt_rotation_id"}, len(data.Favorites), func(n int) []interface{} {
		fav := data.Favorites[n]
		var svc, sched, rot *[16]byte
		id := asUUID(fav.Tgt.TargetID())
		switch fav.Tgt.TargetType() {
		case assignment.TargetTypeService:
			svc = &id
		case assignment.TargetTypeSchedule:
			sched = &id
		case assignment.TargetTypeRotation:
			rot = &id
		}
		return []interface{}{asUUID(fav.UserID), svc, sched, rot}
	}, "users", "services", "schedules", "rotations", "escalation_policies")

	_, err = pool.Exec(ctx, "alter table alerts disable trigger trg_enforce_alert_limit")
	must(err)
	copyFrom("alerts", []string{"id", "created_at", "status", "summary", "details", "dedup_key", "service_id", "source"}, len(data.Alerts), func(n int) []interface{} {
		a := data.Alerts[n]
		var dedup *alert.DedupID
		if a.Status != alert.StatusClosed {
			dedup = a.DedupKey()
		}
		return []interface{}{a.ID, a.CreatedAt, a.Status, a.Summary, a.Details, dedup, asUUID(a.ServiceID), a.Source}
	}, "services")

	copyFrom("alert_logs", []string{"alert_id", "timestamp", "event", "message", "sub_type", "sub_user_id", "sub_classifier", "meta"}, len(data.AlertLogs), func(n int) []interface{} {
		a := data.AlertLogs[n]
		var subType interface{}
		if a.UserID != "" {
			subType = "user"
		}
		return []interface{}{a.AlertID, a.Timestamp, a.Event, a.Message, subType, asUUIDPtr(a.UserID), a.Class, a.Meta}
	}, "alerts", "outgoing_messages", "users")

	copyFrom("alert_feedback", []string{"alert_id", "noise_reason"}, len(data.AlertFeedback), func(n int) []interface{} {
		f := data.AlertFeedback[n]
		return []interface{}{f.ID, f.NoiseReason}
	}, "alerts")

	copyFrom("outgoing_messages", []string{"id", "created_at", "alert_id", "service_id", "escalation_policy_id", "contact_method_id", "user_id", "message_type", "last_status", "sent_at"}, len(data.AlertMessages), func(n int) []interface{} {
		msg := data.AlertMessages[n]
		return []interface{}{asUUID(msg.ID), msg.CreatedAt, msg.AlertID, asUUID(msg.ServiceID), asUUID(msg.EPID), asUUID(msg.CMID), asUUID(msg.UserID), "alert_notification", msg.Status, msg.SentAt}
	}, "alerts", "services", "users", "user_contact_methods")

	dt.Wait()
	_, err = pool.Exec(ctx, "alter table alerts enable trigger all")
	must(err)

	// fix sequences
	_, err = pool.Exec(ctx, "SELECT pg_catalog.setval('public.alerts_id_seq', (select max(id)+1 from public.alerts), true)")
	must(err)

	_, err = pool.Exec(ctx, "vacuum analyze")
	must(err)

	return nil
}

func recreateDB(ctx context.Context, url, dbName string) error {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return errors.Wrap(err, "connect to DB")
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "drop database if exists "+sqlutil.QuoteID(dbName))
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, "create database "+sqlutil.QuoteID(dbName))
	return err
}

func resetDB(ctx context.Context, url string) (n int, err error) {
	if flag.Arg(0) != "" {
		n, err = migrate.Up(ctx, url, flag.Arg(0))
	} else {
		n, err = migrate.ApplyAll(ctx, url)
	}
	return n, err
}
