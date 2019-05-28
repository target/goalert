package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const stampFormat = "20060102150405"

func runCmd(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	fmt.Println("+ ", name, strings.Join(args, " "))
	data, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(data))
}
func setTime(s string, t time.Time) string {
	return s[:19] + t.Format(stampFormat) + s[33:]
}

func main() {
	check := flag.Bool("check", false, "Exit with error status on wrong order, but don't actually rename anything.")
	flag.Parse()
	runCmd("git", "fetch", "--no-tags", "origin", "+refs/heads/master:")
	masterMigrations := strings.Split(runCmd("git", "ls-tree", "-r", "--name-only", "origin/master", "--", "migrate/migrations"), "\n")
	newMigrations := strings.Split(runCmd("git", "diff", "--name-only", "origin/master", "--", "migrate/migrations"), "\n")

	sort.Strings(masterMigrations)
	sort.Strings(newMigrations)

	if len(newMigrations) == 0 || len(newMigrations) == 1 && newMigrations[0] == "" {
		return
	}

	if newMigrations[0] > masterMigrations[len(masterMigrations)-1] {
		// already in order
		return
	}

	if *check {
		log.Println(newMigrations[0], "<=", masterMigrations[len(masterMigrations)-1])
		log.Fatal("found new migrations before those in master")
	}

	t := time.Now().Add(time.Minute)
	for _, m := range newMigrations {
		runCmd("git", "mv", m, setTime(m, t))
		t = t.Add(time.Minute)
	}

}
