package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/target/goalert/smoketest/harness"
)

func main() {
	log.SetFlags(log.Lshortfile)
	db := flag.String("d", "", "Database URL.")
	cmd := flag.String("c", "", "Queries to execute.")
	inTx := flag.Bool("tx", false, "Run in transaction (faster).")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

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
