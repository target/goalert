package calendarsubscription

import (
	"context"
	"database/sql"
	"github.com/target/goalert/util"
)

type Store interface {

}

type DB struct {
	db *sql.DB
}

func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	prep := &util.Prepare{DB: db, Ctx: ctx}
	//p := prep.P

	s := &DB{db: db}
	return s, prep.Err
}
