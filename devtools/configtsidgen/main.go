package main

import (
	"flag"
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/target/goalert/config"
)

func main() {
	out := flag.String("out", "", "Output file.")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	w := os.Stdout
	if *out != "" {
		fd, err := os.OpenFile(*out, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("ERROR:", err)
		}
		defer fd.Close()
		w = fd
	}

	var fields []string
	typ := reflect.TypeOf(config.Config{})
	for i := 0; i < typ.NumField(); i++ {
		grpField := typ.Field(i)
		if grpField.Type.Kind() != reflect.Struct {
			continue
		}
		for j := 0; j < grpField.Type.NumField(); j++ {
			fields = append(fields, "'"+grpField.Name+"."+grpField.Type.Field(j).Name+"'")
		}
	}

	_, err := io.WriteString(w, "\ntype ConfigID = "+strings.Join(fields, "|")+"\n")
	if err != nil {
		log.Fatal(err)
	}
}
