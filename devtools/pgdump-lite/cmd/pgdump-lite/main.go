package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/devtools/pgdump-lite"
)

func main() {
	log.SetFlags(log.Lshortfile)
	file := flag.String("f", "", "Output file (default is stdout).")
	db := flag.String("d", os.Getenv("DBURL"), "DB URL") // use same env var as pg_dump
	dataOnly := flag.Bool("a", false, "dump only the data, not the schema")
	flag.Parse()

	out := os.Stdout
	if *file != "" {
		fd, err := os.Create(*file)
		if err != nil {
			log.Fatalln("ERROR: open output:", err)
		}
		out = fd
		defer fd.Close()
	}

	ctx := context.Background()
	cfg, err := pgx.ParseConfig(*db)
	if err != nil {
		log.Fatalln("ERROR: invalid db url:", err)
	}

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		log.Fatalln("ERROR: connect:", err)
	}
	defer conn.Close(ctx)

	err = pgdump.DumpData(ctx, conn, out)
	if err != nil {
		log.Fatalln("ERROR: dump data:", err)
	}

	if *dataOnly {
		return
	}

	// TODO: dump schema
}
