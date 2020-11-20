package pgdump

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func sortColumns(columns []string) {
	// alphabetical, but with id first
	sort.Slice(columns, func(i, j int) bool {
		ci, cj := columns[i], columns[j]
		if ci == cj {
			return false
		}
		if ci == "id" {
			return true
		}
		if cj == "id" {
			return false
		}

		return ci < cj
	})
}

func queryStrings(ctx context.Context, tx pgx.Tx, sql string, args ...interface{}) ([]string, error) {
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var value string
		err = rows.Scan(&value)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}

	return result, nil
}

type scannable string

func (s *scannable) DecodeText(ci *pgtype.ConnInfo, src []byte) error {
	if src == nil {
		*s = "\\N"
	} else {
		*s = scannable(strings.ReplaceAll(string(src), "\\", "\\\\"))
	}

	return nil
}

func DumpData(ctx context.Context, conn *pgx.Conn, out io.Writer) error {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tables, err := queryStrings(ctx, tx, "select table_name from information_schema.tables where table_schema = 'public'")
	if err != nil {
		return fmt.Errorf("read tables: %w", err)
	}
	sort.Strings(tables)

	for _, table := range tables {
		columns, err := queryStrings(ctx, tx, "select column_name from information_schema.columns where table_schema = 'public' and table_name = $1 order by ordinal_position", table)
		if err != nil {
			return fmt.Errorf("read columns for '%s': %w", table, err)
		}

		primaryCols, err := queryStrings(ctx, tx, `
			select col.column_name
			from information_schema.table_constraints tbl
			join information_schema.constraint_column_usage col on
				col.table_schema = 'public' and
				col.constraint_name = tbl.constraint_name
			where
				tbl.table_schema = 'public' and
				tbl.table_name = $1 and
				constraint_type = 'PRIMARY KEY'
		`, table)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("read primary key for '%s': %w", table, err)
		}
		sortColumns(primaryCols)

		colNames := strings.Join(columns, ", ")
		orderBy := strings.Join(primaryCols, ",")
		if orderBy == "" {
			orderBy = colNames
		}

		fmt.Fprintf(out, "COPY %s (%s) FROM stdin;\n", table, colNames)
		rows, err := tx.Query(ctx,
			fmt.Sprintf("select %s from %s order by %s",
				colNames,
				table,
				orderBy,
			),
			pgx.QuerySimpleProtocol(true),
		)
		if err != nil {
			return fmt.Errorf("read data on '%s': %w", table, err)
		}
		defer rows.Close()
		vals := make([]interface{}, len(columns))

		for i := range vals {
			vals[i] = new(scannable)
		}
		for rows.Next() {
			err = rows.Scan(vals...)
			if err != nil {
				return fmt.Errorf("read data on '%s': %w", table, err)
			}
			for i, v := range vals {
				if i > 0 {
					io.WriteString(out, "\t")
				}
				io.WriteString(out, string(*v.(*scannable)))
			}
			io.WriteString(out, "\n")
		}
		rows.Close()

		fmt.Fprintf(out, "\\.\n\n")
	}

	return tx.Commit(ctx)
}
