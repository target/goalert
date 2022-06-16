package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/target/goalert/devtools/pgmocktime"
)

func main() {
	log.SetFlags(log.Lshortfile)
	url := flag.String("d", os.Getenv("DBURL"), "DB URL") // use same env var as pg_dump
	speed := flag.Float64("s", 1.0, "Speed of time (1.0 = real time).")
	setTime := flag.String("t", "", "Set time to the provided value.")
	advTime := flag.Duration("a", 0, "Advance time by this offset.")
	reset := flag.Bool("reset", false, "Reset time and speed.")
	inject := flag.Bool("inject", false, "Inject instrumentation.")
	remove := flag.Bool("remove", false, "Remove instrumentation.")
	flag.Parse()

	ctx := context.Background()

	m, err := pgmocktime.New(ctx, *url)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	if *inject {
		err = m.Inject(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *reset {
		err = m.Reset(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}

	var setSpeed bool
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "s" {
			setSpeed = true
		}
	})

	if setSpeed {
		err = m.SetSpeed(ctx, *speed)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *setTime != "" {
		t, err := time.Parse(time.RFC3339Nano, *setTime)
		if err != nil {
			log.Fatal(err)
		}

		err = m.SetTime(ctx, t)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *advTime != 0 {
		err = m.AdvanceTime(ctx, *advTime)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *remove {
		err = m.Remove(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}
