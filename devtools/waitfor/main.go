package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"sort"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type waitFunc func(context.Context, string) error

var (
	waitFuncs   = map[string]waitFunc{}
	timeout     time.Duration
	connTimeout time.Duration
	retryTime   time.Duration
)

func register(fn waitFunc, schema ...string) {
	for _, s := range schema {
		waitFuncs[s] = fn
	}
}

func waitFor(ctx context.Context, urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("url parse '%s': %w", urlStr, err)
	}

	t := time.NewTicker(retryTime)
	defer t.Stop()
	for {
		fn, ok := waitFuncs[u.Scheme]
		if !ok {
			return fmt.Errorf("unsupported schema %q", u.Scheme)
		}

		ct, cancel := context.WithTimeout(ctx, connTimeout)
		err = fn(ct, urlStr)
		cancel()
		if err == nil {
			return nil
		}

		log.Println("Waiting for", urlStr, err)
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %s", urlStr)
		case <-t.C:
		}
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	log.SetPrefix("waitfor: ")
	flag.DurationVar(&timeout, "timeout", time.Minute, "Timeout to wait for all checks to complete.")
	flag.DurationVar(&connTimeout, "connect-timeout", 5*time.Second, "Timeout to wait for a single check to complete.")
	flag.DurationVar(&retryTime, "retry-time", 3*time.Second, "Time to wait between retries.")
	flag.Usage = func() {
		fmt.Println("Usage: waitfor [flags] [schema://]host[:port]")

		var schemas []string
		for schema := range waitFuncs {
			schemas = append(schemas, schema)
		}
		sort.Strings(schemas)
		fmt.Println("\nSupported schemas:")
		for _, schema := range schemas {
			fmt.Println("  ", schema)
		}
		fmt.Println()

		fmt.Println("Flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, urlStr := range flag.Args() {
		err := waitFor(ctx, urlStr)
		if err != nil {
			log.Fatal(err)
		}
	}
}
