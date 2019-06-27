package dbsync

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/pkg/errors"
	"github.com/target/goalert/switchover"
)

func (s *Sync) status(ctx context.Context) (string, error) {
	rows, err := s.oldDB.QueryContext(ctx, `
		select count(*), coalesce(application_name, ''), coalesce(usename, '')
		from pg_stat_activity
		where datname=current_database()
		group by application_name, usename
		order by application_name, usename
	`)
	if err != nil {
		return "", errors.Wrap(err, "check DB connections")
	}
	defer rows.Close()
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Text: "Application"},
			{Text: "Username"},
			{Text: "Connections"},
		},
	}

	var issues []string
	for rows.Next() {
		var num int
		var name string
		var user string
		err = rows.Scan(&num, &name, &user)
		if err != nil {
			return "", errors.Wrap(err, "scan query results")
		}
		if strings.Contains(name, "GoAlert") && !strings.Contains(name, "S/O") {
			issues = append(issues, "Non-switchover GoAlert connection: "+name)
		}
		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{Text: name},
			{Text: user},
			{Text: strconv.Itoa(num)},
		})
	}
	rows.Close()
	buf := new(strings.Builder)
	buf.WriteString(table.String() + "\n\n")

	table = simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Text: "Node ID"},
			{Text: "Status"},
			{Text: "Offset"},
			{Text: "Config"},
			{Text: "Last Seen"},
			{Text: "ActiveRequests"},
		},
	}
	nodes := s.NodeStatus()

	for _, stat := range nodes {
		cfg := "Valid"
		if s.oldDBID != stat.DBID || s.newDBID != stat.DBNextID {
			cfg = "Invalid"
			issues = append(issues, fmt.Sprintf("Node[%s] has invalid config (try `reset` or checking db urls)", stat.NodeID))
		}
		since := time.Since(stat.At).Truncate(time.Millisecond)
		if since > 3*time.Second {
			issues = append(issues, fmt.Sprintf("Node[%s] has not been seen for >3 seconds (try `reset`)", stat.NodeID))
		}

		if stat.State == switchover.StateAbort {
			issues = append(issues, fmt.Sprintf("Node[%s] has aborted (try `reset`)", stat.NodeID))
		} else if stat.State != switchover.StateReady {
			issues = append(issues, fmt.Sprintf("Node[%s] is not ready", stat.NodeID))
		}

		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{Text: stat.NodeID},
			{Text: string(stat.State)},
			{Text: stat.Offset.String()},
			{Text: cfg},
			{Text: since.String()},
			{Text: strconv.Itoa(stat.ActiveRequests)},
		})
	}
	buf.WriteString(table.String() + "\n\n")

	if len(nodes) == 0 {
		issues = append(issues, "No nodes detected.")
	}

	fmt.Fprintf(buf, "Node Count: %d\n", len(nodes))
	fmt.Fprintln(buf, "Local offset:", s.Offset())
	var stat string
	err = s.oldDB.QueryRowContext(ctx, `select current_state from switchover_state`).Scan(&stat)
	if err != nil {
		return "", errors.Wrap(err, "lookup switchover state")
	}

	fmt.Fprintln(buf, "Switchover state:", stat)
	if stat == "idle" {
		return buf.String(), nil
	} else if stat == "use_next_db" {
		issues = append(issues, "Switchover has already been completed")
	}

	var changeMax int
	err = s.oldDB.QueryRowContext(ctx, `select coalesce(max(id),0) from change_log`).Scan(&changeMax)
	if err != nil {
		return "", errors.Wrap(err, "lookup change id")
	}
	fmt.Fprintln(buf, "Max change_log ID:", changeMax)

	err = s.newDB.QueryRowContext(ctx, `select coalesce(max(id),0) from change_log`).Scan(&changeMax)
	if err != nil {
		return "", errors.Wrap(err, "lookup change id (new)")
	}
	fmt.Fprintln(buf, "Max change_log ID (next DB):", changeMax)

	if len(issues) > 0 {
		fmt.Fprintln(buf, "\nPotential Problems Found:")
		for _, s := range issues {
			fmt.Fprintln(buf, "- "+s)
		}
	} else {
		fmt.Fprintln(buf, "\nNo Problems Found")
	}

	return buf.String(), nil
}
