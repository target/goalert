package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx"
)

func main() {
	log.SetFlags(log.Lshortfile)
	db := flag.String("d", "", "Database URL.")
	cmd := flag.String("c", "", "Queries to execute.")
	inTx := flag.Bool("tx", false, "Run in transaction (faster).")
	flag.Parse()

	queries := strings.Split(*cmd, ";")
	q := queries[:0]
	for _, str := range queries {
		str = strings.TrimSpace(str)
		if str == "" {
			continue
		}
		q = append(q, str)
	}
	queries = q
	if len(queries) == 0 {
		return
	}

	cfg, err := pgx.ParseConnectionString(*db)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := pgx.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if !*inTx {
		// just run them one by one
		for _, q := range queries {
			tag, err := conn.ExecEx(ctx, q, nil)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%s %d", strings.SplitN(q, " ", 2)[0], tag.RowsAffected())
		}
		return
	}

	b := conn.BeginBatch()
	defer b.Close()

	for _, q := range queries {
		b.Queue(q, nil, nil, nil)
	}

	err = b.Send(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, q := range queries {
		tag, err := b.ExecResults()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s %d", strings.SplitN(q, " ", 2)[0], tag.RowsAffected())
	}

}
