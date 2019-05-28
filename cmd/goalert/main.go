package main

import (
	"context"
	"github.com/target/goalert/app"
	"github.com/target/goalert/util/log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	err := app.RootCmd.Execute()
	if err != nil {
		log.Log(context.TODO(), err)
		os.Exit(1)
	}
}
