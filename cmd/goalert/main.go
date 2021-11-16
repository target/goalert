package main

import (
	"os"
	_ "time/tzdata"

	_ "github.com/joho/godotenv/autoload"
	"github.com/target/goalert/app"
	"github.com/target/goalert/util/log"
)

func main() {
	l := log.NewLogger()
	ctx := l.BackgroundContext()
	err := app.RootCmd.ExecuteContext(ctx)
	if err != nil {
		log.Log(ctx, err)
		os.Exit(1)
	}
}
