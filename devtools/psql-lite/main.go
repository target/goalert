package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/target/goalert/util/sqlutil"
)

func main() {
	log.SetFlags(log.Lshortfile)
	db := flag.String("d", "", "Database URL.")
	cmd := flag.String("c", "", "Queries to execute.")
	inTx := flag.Bool("tx", false, "Run in transaction (faster).")
	t := flag.Duration("t", 0, "Specify a timeout for the query(s) to execute.")
	flag.Parse()

	ctx := context.Background()
	if *t > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), *t)
		defer cancel()
	}

	conn, err := pgx.Connect(ctx, *db)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	var tx pgx.Tx
	if *inTx {
		tx, err = conn.Begin(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer sqlutil.RollbackContext(ctx, "psql-lite: exec sql", tx)
	}

	for _, q := range sqlutil.SplitQuery(*cmd) {
		rows, err := conn.Query(ctx, q)
		if errors.Is(err, pgx.ErrNoRows) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var s string
			err = rows.Scan(&s)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(s)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
	}

	if *inTx {
		err = tx.Commit(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}
