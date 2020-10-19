package twilio

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/target/goalert/util/sqlutil"

	"github.com/target/goalert/util"
)

const (
	banErrorCount = 20
	banDuration   = 4 * time.Hour
)

type dbBan struct {
	db       *sql.DB
	insert   *sql.Stmt
	isBanned *sql.Stmt
	c        *Config
}

func newBanDB(ctx context.Context, db *sql.DB, c *Config, tableName string) (*dbBan, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	//  will register these sql statements by Prepared statements
	return &dbBan{
		db: db,
		c:  c,
		insert: p.P(fmt.Sprintf(`
			INSERT INTO %s (
				phone_number,
				outgoing,
				error_message
			)
			VALUES
				($1, $2, $3)
		`, sqlutil.QuoteID(tableName))),
		isBanned: p.P(fmt.Sprintf(`
			select count(1)
			from %s
			where outgoing = $1
				and phone_number = $2
				and occurred_at between
					(now() - '%f minutes'::interval) and now()
		`,
			sqlutil.QuoteID(tableName),
			banDuration.Minutes(),
		)),
	}, p.Err
}
func (db *dbBan) IsBanned(ctx context.Context, number string, outgoing bool) (bool, error) {
	row := db.isBanned.QueryRowContext(ctx, outgoing, number)
	var count int
	err := row.Scan(&count)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return count >= banErrorCount, nil
}
func (db *dbBan) RecordError(ctx context.Context, number string, outgoing bool, message string) error {
	_, err := db.insert.ExecContext(ctx, number, outgoing, message)
	return err
}
