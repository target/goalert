package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func waitForHTTP(ctx context.Context, url string) {
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	for {
		_, err := http.Get(url)
		if err == nil {
			return
		}

		log.Println("Waiting for", url, err)
		select {
		case <-ctx.Done():
			log.Fatal("Timeout waiting for", url)
		case <-t.C:
		}
	}
}
func waitForPostgres(ctx context.Context, connStr string) {
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal("db open:", err)
	}
	defer db.Close()
	for {
		err = db.PingContext(ctx)
		if err == nil {
			return
		}

		log.Println("Waiting for", connStr, err)
		select {
		case <-ctx.Done():
			log.Fatal("Timeout waiting for", connStr)
		case <-t.C:
		}
	}
}

func main() {
	timeout := flag.Duration("timeout", time.Minute, "Timeout to wait for all checks to complete.")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	for _, u := range flag.Args() {
		if strings.HasPrefix(u, "postgres://") || strings.HasPrefix(u, "postgresql://") {
			waitForPostgres(ctx, u)
		} else {
			waitForHTTP(ctx, u)
		}
	}
}
