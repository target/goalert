package main

import (
	"context"
	"flag"
	"os"

	"github.com/target/goalert/migrate"
)

func main() {
	db := flag.String("db-url", os.Getenv("GOALERT_DB_URL"), "Database URL")
	flag.Parse()

	_, err := migrate.ApplyAll(context.Background(), *db)
	if err != nil {
		panic(err)
	}
}
