package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/target/goalert/devtools/pgdump-lite"
)

func main() {
	log.SetFlags(log.Lshortfile)
	file := flag.String("f", "", "Output file (default is stdout).")
	db := flag.String("d", os.Getenv("DBURL"), "DB URL") // use same env var as pg_dump
	dataOnly := flag.Bool("a", false, "dump only the data, not the schema")
	schemaOnly := flag.Bool("s", false, "dump only the schema, no data")
	parallel := flag.Bool("p", false, "dump data in parallel (note: separate tables will still be dumped in order, but not in the same transaction, so may be inconsistent between tables)")
	skip := flag.String("T", "", "skip tables")
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
	cfg, err := pgxpool.ParseConfig(*db)
	if err != nil {
		log.Fatalln("ERROR: invalid db url:", err)
	}
	cfg.ConnConfig.RuntimeParams["client_encoding"] = "UTF8"

	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalln("ERROR: connect:", err)
	}
	defer conn.Close()

	var s *pgdump.Schema
	if !*dataOnly {
		s, err = pgdump.DumpSchema(ctx, conn)
		if err != nil {
			log.Fatalln("ERROR: dump data:", err)
		}
		_, err = out.WriteString("--\n-- pgdump-lite database dump\n--\n\n")
		if err != nil {
			log.Fatalln("ERROR: write header:", err)
		}
		_, err = out.WriteString(s.String())
		if err != nil {
			log.Fatalln("ERROR: write schema:", err)
		}
	}

	if !*schemaOnly {
		if *parallel {
			err = pgdump.DumpDataWithSchemaParallel(ctx, conn, out, strings.Split(*skip, ","), s)
		} else {
			err = pgdump.DumpDataWithSchema(ctx, conn, out, strings.Split(*skip, ","), s)
		}
		if err != nil {
			log.Fatalln("ERROR: dump data:", err)
		}
	}
}
