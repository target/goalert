package harness

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/util/sqlutil"
)

func sqlSplitBlock(blockIdx int, data []byte, atEOF bool) (advance int, token []byte, err error) {
	nextBlockIdx := bytes.Index(data[blockIdx+2:], []byte("$$"))
	if nextBlockIdx == -1 {
		if atEOF {
			// return rest as the final query
			return len(data), data, nil
		}

		return 0, nil, nil
	}

	next := blockIdx + 2 + nextBlockIdx + 2

	advance, token, err = sqlSplitQuery(data[next:], atEOF)
	if err != nil {
		return 0, nil, err
	}

	if advance == 0 {
		return 0, nil, nil
	}

	return next + advance, data[:next+len(token)], nil
}

func sqlSplitQuery(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	semiIdx := bytes.IndexRune(data, ';')
	blockIdx := bytes.Index(data, []byte("$$"))
	if blockIdx != -1 && (semiIdx == -1 || semiIdx > blockIdx) {
		// have block start and it comes before semi (or no semi)
		return sqlSplitBlock(blockIdx, data, atEOF)
	}

	// no block, or it comes later
	if semiIdx == -1 {
		if atEOF {
			// return rest as the final query
			return len(data), data, nil
		}

		return 0, nil, nil
	}

	// return up to semi
	return semiIdx + 1, data[:semiIdx], nil
}

func sqlSplit(query string) []string {
	s := bufio.NewScanner(strings.NewReader(query))
	s.Split(sqlSplitQuery)

	var queries []string
	for s.Scan() {
		if strings.TrimSpace(s.Text()) == "" {
			continue
		}
		queries = append(queries, s.Text())
	}

	return queries
}

// ExecSQLBatch will execute all queries in a transaction by sending them all at once.
func ExecSQLBatch(ctx context.Context, url string, query string) error {
	queries := sqlSplit(query)

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "set statement_timeout = 3000")
	if err != nil {
		return err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sqlutil.RollbackContext(ctx, "harness: exec sql", tx)

	b := &pgx.Batch{}
	for _, q := range queries {
		b.Queue(q)
	}

	err = tx.SendBatch(ctx, b).Close()
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// SQLRollback will rollback the transaction for cleanup, failing the test on error.
func SQLRollback(t *testing.T, errMsg string, tx *sql.Tx) {
	err := tx.Rollback()
	switch {
	case err == nil:
	case errors.Is(err, sql.ErrTxDone):
	case errors.Is(err, sql.ErrConnDone):
	default:
		t.Fatalf("ERROR: %s: tx rollback: %v", errMsg, err)
	}
}
