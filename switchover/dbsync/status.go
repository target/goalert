package dbsync

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
	"github.com/pkg/errors"
)

func (s *Sync) status(ctx context.Context) (string, error) {
	rows, err := s.oldDB.QueryContext(ctx, `
		select count(*), application_name, usename
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
	for rows.Next() {
		var num int
		var name string
		var user string
		err = rows.Scan(&num, &name, &user)
		if err != nil {
			return "", errors.Wrap(err, "scan query results")
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
		if !stat.MatchDBNext(s.newURL) {
			cfg = "Invalid"
		}
		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{Text: stat.NodeID},
			{Text: string(stat.State)},
			{Text: stat.Offset.String()},
			{Text: cfg},
			{Text: time.Since(stat.At).Truncate(time.Millisecond).String()},
			{Text: strconv.Itoa(stat.ActiveRequests)},
		})
	}
	buf.WriteString(table.String() + "\n\n")

	fmt.Fprintf(buf, "Node Count: %d\n", len(nodes))
	fmt.Fprintln(buf, "Local offset:", s.Offset())
	var stat string
	err = s.oldDB.QueryRowContext(ctx, `select current_state from switchover_state`).Scan(&stat)
	if err != nil {
		return "", errors.Wrap(err, "lookup switchover state")
	}

	fmt.Fprintln(buf, "Switchover state:", stat)

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

	return buf.String(), nil
}
