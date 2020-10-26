package main

import (
	"context"
	"os"
	_ "time/tzdata"

	_ "github.com/joho/godotenv/autoload"
	"github.com/target/goalert/app"
	"github.com/target/goalert/util/log"
)

func main() {
	err := app.RootCmd.Execute()
	if err != nil {
		log.Log(context.TODO(), err)
		os.Exit(1)
	}
}
