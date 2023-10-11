package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/target/goalert/util/log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var queries = [][]string{
	{"UserCount", `select count(id) from users`},
	{"CMMax", `select count(id) from user_contact_methods group by user_id order by count desc limit 1`},
	{"NRMax", `select count(id) from user_notification_rules group by user_id order by count desc limit 1`},
	{"NRCMMax", `select count(id) from user_notification_rules group by user_id,contact_method_id order by count desc limit 1`},
	{"EPCount", `select count(id) from escalation_policies`},
	{"EPMaxStep", `select count(id) from escalation_policy_steps group by escalation_policy_id order by count desc limit 1`},
	{"EPMaxAssigned", `select count(id) from escalation_policy_actions group by escalation_policy_step_id order by count desc limit 1`},
	{"SvcCount", `select count(id) from services`},
	{"RotationMaxPart", `select count(id) from rotation_participants group by rotation_id order by count desc limit 1`},
	{"ScheduleCount", `select count(id) from schedules`},
	{"AlertClosedCount", `select count(id) from alerts where status = 'closed'`},
	{"AlertActiveCount", `select count(id) from alerts where status = 'triggered' or status = 'active'`},
	{"RotationCount", `select count(id) from rotations`},
	{"IntegrationKeyMax", `select count(id) from integration_keys group by service_id order by count desc limit 1`},
	{"ScheduleMaxRules", `select count(id) from schedule_rules group by schedule_id order by count desc limit 1`},
}

func noErr(err error) {
	if err == nil {
		return
	}
	log.Log(context.TODO(), err)
	os.Exit(1)
}

func main() {
	mult := flag.Float64("m", 1.5, "Multiplier for prod values.")
	url := flag.String("db", os.Getenv("DB_URL"), "DB connection URL.")
	flag.Parse()
	db, err := sql.Open("pgx", *url)
	noErr(err)

	for _, q := range queries {
		var n int
		row := db.QueryRow(q[1])
		noErr(row.Scan(&n))
		n = int(float64(n)**mult) + 1
		fmt.Printf("\t%s = %d // %s\n", q[0], n, q[1])
	}
	db.Close()
}
