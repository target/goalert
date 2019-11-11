package main

import (
	"context"
	"flag"
	"log"

	"github.com/target/goalert/smoketest/harness"
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

	var err error
	if *inTx {
		err = harness.ExecSQLBatch(ctx, *db, *cmd)
	} else {
		err = harness.ExecSQL(ctx, *db, *cmd)
	}
	if err != nil {
		log.Fatal(err)
	}
}
