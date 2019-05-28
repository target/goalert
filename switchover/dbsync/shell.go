package dbsync

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/target/goalert/migrate"
	"github.com/target/goalert/switchover"
	"net/url"
	"strings"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"

	"github.com/abiosoft/ishell"
	_ "github.com/jackc/pgx/stdlib" // load PGX driver
	"github.com/pkg/errors"
)

// RunShell will start the switchover shell.
func RunShell(oldURL, newURL string) error {
	ctx := context.Background()
	u, err := url.Parse(oldURL)
	if err != nil {
		return errors.Wrap(err, "parse old URL")
	}
	q := u.Query()
	q.Set("application_name", "GoAlert Switch-Over Shell")
	u.RawQuery = q.Encode()
	oldURL = u.String()

	u, err = url.Parse(newURL)
	if err != nil {
		return errors.Wrap(err, "parse new URL")
	}
	q = u.Query()
	q.Set("application_name", "GoAlert Switch-Over Shell")
	u.RawQuery = q.Encode()
	newURL = u.String()

	db, err := sql.Open("pgx", oldURL)
	if err != nil {
		return errors.Wrap(err, "open DB")
	}

	var numMigrations int
	err = db.QueryRowContext(ctx, `select count(*) from gorp_migrations`).Scan(&numMigrations)
	if err != nil {
		return errors.Wrap(err, "validate migration number")
	}
	if numMigrations != len(migrate.Names()) {
		return errors.Errorf("got %d migrations but expected %d", numMigrations, len(migrate.Names()))
	}

	fmt.Println("Applying migrations to next-db...")
	dbNew, err := sql.Open("pgx", newURL)
	if err != nil {
		return errors.Wrap(err, "open next-DB")
	}
	_, err = migrate.ApplyAll(ctx, dbNew)
	if err != nil {
		return errors.Wrap(err, "migrate next-DB")
	}
	dbNew.Close()

	dbNew, err = sql.Open("pgx", newURL)
	if err != nil {
		return errors.Wrap(err, "open next-DB")
	}
	sendNotif, err := db.PrepareContext(ctx, `select pg_notify($1, $2)`)
	if err != nil {
		return errors.Wrap(err, "prepare notify statement")
	}

	s, err := NewSync(ctx, db, dbNew, newURL)
	if err != nil {
		return errors.Wrap(err, "init sync manager")
	}

	sh := newCtxShell()
	sh.AddCmd(ctxCmd{
		Name:     "sync",
		Help:     "Execute DB sync.",
		HasFlags: true,
		Func: func(ctx context.Context, sh *ishell.Context) error {
			fset := flag.NewFlagSet("sync", flag.ContinueOnError)
			cont := fset.Bool("continuous", false, "Perform continuous sync (up to once per second).")
			err := fset.Parse(sh.Args)
			if err != nil {
				return err
			}

			if *cont {
				sh.Print("\033[H\033[2J")
			}
			start := time.Now()
			err = s.Sync(ctx, false, false)
			if err != nil {
				return err
			}
			sh.Printf("Completed sync in %s\n", time.Since(start).Truncate(time.Millisecond).String())

			if !*cont {
				return nil
			}

			t := time.NewTicker(time.Second)
			for {
				select {
				case <-t.C:
					if *cont {
						sh.Print("\033[H\033[2J")
					}
					start := time.Now()
					err = s.Sync(ctx, false, false)
					if err != nil {
						return err
					}
					sh.Printf("Completed sync in %s\n", time.Since(start).Truncate(time.Millisecond).String())
				case <-ctx.Done():
					return nil
				}
			}

		},
	})
	sh.AddCmd(ctxCmd{
		Name: "enable",
		Help: "Enable change_log",
		Func: func(ctx context.Context, sh *ishell.Context) error {
			res, err := db.ExecContext(ctx, `update switchover_state set current_state = 'in_progress' where current_state = 'idle'`)
			if err != nil {
				return err
			}
			r, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if r != 1 {
				return errors.New("not idle")
			}

			status, err := s.status(ctx)
			if err != nil {
				return err
			}
			sh.Println(status)
			sh.Println("change_log enabled.")

			return nil
		},
	})
	sh.AddCmd(ctxCmd{
		Name: "disable",
		Help: "Enable change_log",
		Func: func(ctx context.Context, sh *ishell.Context) error {
			res, err := db.ExecContext(ctx, `update switchover_state set current_state = 'idle' where current_state = 'in_progress'`)
			if err != nil {
				return err
			}
			r, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if r != 1 {
				return errors.New("not in_progress")
			}

			status, err := s.status(ctx)
			if err != nil {
				return err
			}
			sh.Println(status)
			sh.Println("change_log disabled")
			return nil
		},
	})

	sh.AddCmd(ctxCmd{
		Name: "reset",
		Help: "Reset node status",
		Func: func(ctx context.Context, sh *ishell.Context) error {
			s.mx.Lock()
			for key := range s.nodeStatus {
				delete(s.nodeStatus, key)
			}
			s.mx.Unlock()
			_, err := sendNotif.ExecContext(ctx, switchover.ControlChannel, "reset")
			if err != nil {
				return err

			}

			status, err := s.status(ctx)
			if err != nil {
				return err
			}
			sh.Println(status)
			sh.Println("Reset signal sent.")

			return nil
		},
	})
	sh.AddCmd(ctxCmd{
		Name:     "status",
		Help:     "Print current status.",
		HasFlags: true,
		Func: func(ctx context.Context, sh *ishell.Context) error {
			fset := flag.NewFlagSet("status", flag.ContinueOnError)
			watch := fset.Bool("w", false, "Watch mode.")
			dur := fset.Duration("n", 2*time.Second, "Time between updates")
			err := fset.Parse(sh.Args)
			if err != nil {
				return err
			}

			status, err := s.status(ctx)
			if err != nil {
				return err
			}
			if *watch {
				sh.Print("\033[H\033[2J")
			}
			sh.Println(status)
			if !*watch {
				return nil
			}

			t := time.NewTicker(*dur)
			for {
				select {
				case <-t.C:
					status, err := s.status(ctx)
					if err != nil {
						return err
					}
					sh.Print("\033[H\033[2J")
					sh.Println(status)
				case <-ctx.Done():
					return nil
				}
			}
		},
	})
	sh.AddCmd(ctxCmd{
		Name:     "execute",
		HasFlags: true,
		Help:     "Execute the switchover procedure.",
		Func: func(ctx context.Context, sh *ishell.Context) error {
			cfg := switchover.DefaultConfig()
			fset := flag.NewFlagSet("execute", flag.ContinueOnError)
			fset.BoolVar(&cfg.NoPauseAPI, "allow-api", cfg.NoPauseAPI, "Allow API requests during pause phase (DB calls will still pause during final sync).")
			fset.DurationVar(&cfg.ConsensusTimeout, "consensus-timeout", cfg.ConsensusTimeout, "Timeout to reach consensus.")
			fset.DurationVar(&cfg.PauseDelay, "pause-delay", cfg.PauseDelay, "Delay from start until global pause begins.")
			fset.DurationVar(&cfg.PauseTimeout, "pause-timeout", cfg.PauseTimeout, "Timeout to achieve global pause.")
			fset.DurationVar(&cfg.MaxPause, "max-pause", cfg.MaxPause, "Maximum duration for any pause/delay/impact during switchover.")
			extraSync := fset.Bool("extra-sync", false, "Do a second sync after pausing, immediately before the final sync (useful with -allow-api).")
			noSwitch := fset.Bool("no-switch", false, "Run the entire procedure, but omit the final use_next_db update.")
			err := fset.Parse(sh.Args)
			if err != nil {
				if err == flag.ErrHelp {
					return nil
				}
				return err
			}

			status, err := s.status(ctx)
			if err != nil {
				return errors.Wrap(err, "get status")
			}

			details := new(strings.Builder)

			pauseAPI := "yes"
			if cfg.NoPauseAPI {
				pauseAPI = "no"
			}
			fmt.Fprintln(details, status)
			fmt.Fprintln(details, "Switch-Over Details")
			fmt.Fprintln(details, "  Pause API Requests:", pauseAPI)
			fmt.Fprintln(details, "  Consensus Timeout :", cfg.ConsensusTimeout)
			fmt.Fprintln(details, "  Pause Starts After:", cfg.PauseDelay)
			fmt.Fprintln(details, "  Pause Timeout     :", cfg.PauseTimeout)
			fmt.Fprintln(details, "  Absolute Max Pause:", cfg.MaxPause)
			fmt.Fprintln(details, "  Avail. Sync Time  :", cfg.MaxPause-2*time.Second-cfg.PauseTimeout, "-", cfg.MaxPause-2*time.Second)
			fmt.Fprintln(details, "  Max Alloted Time  :", cfg.PauseDelay+cfg.MaxPause)
			fmt.Fprintln(details)
			fmt.Fprintln(details, "Ready?")

			if sh.MultiChoice([]string{"Cancel", "Go!"}, details.String()) != 1 {
				sh.Println()
				return nil
			}

			start := time.Now()
			err = s.Sync(ctx, false, false)
			if err != nil {
				return err
			}
			sh.Printf("Completed sync in %s\n", time.Since(start).Truncate(time.Second/10).String())

			nodes := s.NodeStatus()
			n := len(nodes)
			if n == 0 {
				return errors.New("no nodes are available")
			}

			if !s.Ready() {
				return errors.New("all nodes are not ready")
			}

			for _, stat := range nodes {
				if !stat.MatchDBNext(newURL) {
					return errors.New("one or more nodes (or this shell) have mismatched config, check db-url-next")
				}
				if stat.At.Before(time.Now().Add(-5 * time.Second)) {
					return errors.New("one or more nodes have timed out (try reset)")
				}
			}

			p := mpb.New()
			var done bool
			abort := func() {
				if !done {
					sh.Println("ABORT")
					sendNotif.ExecContext(context.Background(), switchover.ControlChannel, "abort")
				}
			}
			defer abort()

			cfg.BeginAt = time.Now()

			sh.Println()
			sh.Println("Switch-Over Start ::", cfg.BeginAt.Format(time.StampMilli))
			sh.Println()

			swDeadline := cfg.AbsoluteDeadline().Add(-2 * time.Second)
			ctx, cancel := context.WithDeadline(ctx, swDeadline)
			defer cancel()

			_, err = sendNotif.ExecContext(ctx, switchover.ControlChannel, cfg.Serialize(s.Offset()))
			if err != nil {
				return errors.Wrap(err, "send control message")
			}

			cBar := p.AddBar(int64(n),
				mpb.PrependDecorators(decor.Name("Consensus", decor.WCSyncSpaceR)),
				mpb.BarClearOnComplete(),
				mpb.AppendDecorators(
					decor.OnComplete(decor.CountersNoUnit("(%d of %d nodes)", decor.WCSyncSpaceR), "Done"),
				),
			)

			cCtx, cCancel := context.WithDeadline(ctx, cfg.ConsensusDeadline())
			defer cCancel()
			err = s.NodeStateWait(cCtx, n, cBar, switchover.StateArmed, switchover.StateArmWait)
			if err != nil {
				p.Abort(cBar, false)
				p.Wait()
				return errors.Wrap(err, "wait for consensus")
			}
			p.Wait()

			t := time.NewTicker(time.Second)
			tE := time.NewTimer(time.Until(cfg.PauseAt()))
		waitLoop:
			for {
				dur := time.Until(cfg.PauseAt()).Truncate(time.Second)
				if dur >= time.Second {
					sh.Printf("Stop-The-World Pause begins in %ds...\n", dur/time.Second)
				} else {
					break
				}
				select {
				case <-t.C:
				case <-tE.C:
					break waitLoop
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			p = mpb.New()
			pBar := p.AddBar(int64(n),
				mpb.PrependDecorators(decor.Name("STW Pause", decor.WCSyncSpaceR)),
				mpb.BarClearOnComplete(),
				mpb.AppendDecorators(
					decor.OnComplete(decor.CountersNoUnit("(%d of %d nodes)", decor.WCSyncSpaceR), "Done"),
				),
			)
			pCtx, pCancel := context.WithDeadline(ctx, cfg.PauseDeadline())
			defer pCancel()
			err = s.NodeStateWait(pCtx, n, pBar, switchover.StatePaused, switchover.StatePauseWait)
			if err != nil {
				p.Abort(pBar, false)
				p.Wait()
				return errors.Wrap(err, "wait for pause")
			}
			p.Wait()

			if *extraSync {
				start = time.Now()
				err = s.Sync(ctx, false, false)
				if err != nil {
					return err
				}
				sh.Printf("Completed extra sync in %s\n", time.Since(start).Truncate(time.Second/10).String())
			}

			sh.Println("Begin final synchronization")
			err = s.Sync(ctx, true, !*noSwitch)
			if err != nil {
				return err
			}

			if !*noSwitch {
				sh.Println("Next DB is now permanently active, switchover complete.")
			}

			_, err = sendNotif.ExecContext(ctx, switchover.ControlChannel, "done")
			done = true
			return errors.Wrap(err, "send done signal")
		},
	})

	fmt.Println("GoAlert Switch-Over Shell")
	fmt.Println(sh.HelpText())
	sh.Run()
	return nil
}
