package nonce

import (
	"context"
	"database/sql"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// Store allows generating and consuming nonce values.
type Store interface {
	New() [16]byte
	Consume(context.Context, [16]byte) (bool, error)
	Shutdown(context.Context) error
}

// DB implements the Store interface using postgres as it's backend.
type DB struct {
	db       *sql.DB
	shutdown chan context.Context

	consume *sql.Stmt
	cleanup *sql.Stmt
}

// NewDB prepares a new DB instance for the given sql.DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	d := &DB{
		db:       db,
		shutdown: make(chan context.Context),

		consume: p.P(`
			insert into auth_nonce (id)
			values ($1)
			on conflict do nothing
		`),
		cleanup: p.P(`
			delete from auth_nonce
			where created_at < now() - '1 week'::interval
		`),
	}
	if p.Err != nil {
		return nil, p.Err
	}
	go d.loop()

	return d, nil
}

func (db *DB) loop() {
	defer close(db.shutdown)
	t := time.NewTicker(time.Hour * 24)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			_, err := db.cleanup.ExecContext(context.Background())
			if err != nil {
				log.Log(context.Background(), errors.Wrap(err, "cleanup old nonce values"))
			}
		case <-db.shutdown:
			return
		}
	}
}

// Shutdown allows gracefully shutting down the nonce store.
func (db *DB) Shutdown(ctx context.Context) error {
	if db == nil {
		return nil
	}
	db.shutdown <- ctx

	// wait for it to complete
	<-db.shutdown
	return nil
}

// New will generate a new cryptographically random nonce value.
func (db *DB) New() (id [16]byte) {
	copy(id[:], uuid.NewV4().Bytes())
	return id
}

// Consume will record the use of a nonce value.
//
// An error is returned if it is not possible to validate the nonce value.
// Otherwise true/false is returned to indicate if the id is valid.
//
// The first call to Consume for a given ID will return true, subsequent calls
// for the same ID will return false.
func (db *DB) Consume(ctx context.Context, id [16]byte) (bool, error) {
	res, err := db.consume.ExecContext(ctx, uuid.UUID(id).String())
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}
