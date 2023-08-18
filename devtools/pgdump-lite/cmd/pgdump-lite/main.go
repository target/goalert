package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/devtools/pgdump-lite"
)

func main() {
	log.SetFlags(log.Lshortfile)
	file := flag.String("f", "", "Output file (default is stdout).")
	db := flag.String("d", os.Getenv("DBURL"), "DB URL") // use same env var as pg_dump
	dataOnly := flag.Bool("a", false, "dump only the data, not the schema")
	schemaOnly := flag.Bool("s", false, "dump only the schema, no data")
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
	cfg, err := pgx.ParseConfig(*db)
	if err != nil {
		log.Fatalln("ERROR: invalid db url:", err)
	}
	cfg.RuntimeParams["client_encoding"] = "UTF8"

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		log.Fatalln("ERROR: connect:", err)
	}
	defer conn.Close(ctx)

	if !*dataOnly {
		s, err := pgdump.DumpSchema(ctx, conn)
		if err != nil {
			log.Fatalln("ERROR: dump data:", err)
		}
		_, err = out.WriteString(s.String())
		if err != nil {
			log.Fatalln("ERROR: write schema:", err)
		}
	}

	if !*schemaOnly {
		err = pgdump.DumpData(ctx, conn, out, strings.Split(*skip, ","))
		if err != nil {
			log.Fatalln("ERROR: dump data:", err)
		}
	}
}
