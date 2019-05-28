package dbsync

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"

	"github.com/abiosoft/ishell"
)

type ctxShell struct {
	*ishell.Shell
	mx sync.Mutex

	ctx    context.Context
	cancel func()
}
type ctxCmd struct {
	Name, Help string
	HasFlags   bool
	Func       func(context.Context, *ishell.Context) error
}

func newCtxShell() *ctxShell {
	sh := &ctxShell{Shell: ishell.New()}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		for {
			<-ch
			sh.mx.Lock()
			sh.cancel()
			sh.ctx, sh.cancel = context.WithCancel(context.Background())
			sh.mx.Unlock()
		}
	}()
	sh.ctx, sh.cancel = context.WithCancel(context.Background())
	sh.Interrupt(func(c *ishell.Context, count int, input string) {
		if count > 1 {
			sh.Stop()
			return
		}
		c.Println("Interrupt")
		c.Println("Press CTRL+C again to quit.")
	})
	return sh
}

func (sh *ctxShell) AddCmd(cmd ctxCmd) {
	sh.Shell.AddCmd(&ishell.Cmd{
		Name: cmd.Name,
		Help: cmd.Help,
		Func: func(c *ishell.Context) {
			sh.mx.Lock()
			ctx := sh.ctx
			sh.mx.Unlock()
			var err error
			if !cmd.HasFlags {
				fset := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
				err = fset.Parse(c.Args)
			}
			if err == nil {
				err = cmd.Func(ctx, c)
			}
			if err == flag.ErrHelp {
				err = nil
			}
			if err != nil {
				c.Println("ERROR:", err)
			}
		},
	})
}
