package main

import (
	"fmt"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
)

func gen(types ...interface{}) {
	for _, t := range types {
		elem := reflect.TypeOf(t).Elem()
		run(elem.PkgPath(), elem.Name())
	}
}
func run(pkg string, iface string) {
	parts := strings.Split(pkg, "/")
	pName := parts[len(parts)-1]
	dir := path.Join("internal", "mocks", "mock_"+pName)
	os.MkdirAll(dir, 0755)

	file := path.Join(dir, "mock"+strings.ToLower(iface)+".go")
	fd, err := os.Create(file)
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}
	defer fd.Close()

	cmd := exec.Command("mockgen", pkg, iface)
	cmd.Stdout = fd
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		os.Exit(1)
	}
}

func main() {
	gen(
		(*escalation.Store)(nil),
		(*escalation.Manager)(nil),
		(*rule.Store)(nil),
		(*schedule.Store)(nil),
		(*rotation.Store)(nil),
		(*alert.Store)(nil),
	)
}
