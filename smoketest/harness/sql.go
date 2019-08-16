package harness

import (
	"bufio"
	"bytes"
	"context"
	"strings"

	"github.com/jackc/pgx"
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

// ExecSQL will execute all queries one-by-one.
func ExecSQL(ctx context.Context, url string, query string) error {
	queries := sqlSplit(query)
	cfg, err := pgx.ParseConnectionString(url)
	if err != nil {
		return err
	}
	conn, err := pgx.Connect(cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, q := range queries {
		_, err := conn.ExecEx(ctx, q, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// ExecSQLBatch will execute all queries in a transaction by sending them all at once.
func ExecSQLBatch(ctx context.Context, url string, query string) error {
	queries := sqlSplit(query)

	cfg, err := pgx.ParseConnectionString(url)
	if err != nil {
		return err
	}

	conn, err := pgx.Connect(cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b := tx.BeginBatch()
	defer b.Close()
	for _, q := range queries {
		b.Queue(q, nil, nil, nil)
	}

	err = b.Send(ctx, nil)
	if err != nil {
		return err
	}

	for range queries {
		_, err = b.ExecResults()
		if err != nil {
			return err
		}
	}
	err = b.Close()
	if err != nil {
		return err
	}

	return tx.CommitEx(ctx)
}
