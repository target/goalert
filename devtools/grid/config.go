package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

// Config represents advanced configuration of environmental variables.
type Config struct {
	Env      []string
	Variants []struct {
		Prefix string
		Env    []string
	}
}

type Instance struct {
	Done func()
	Env  []string
}

var (
	nameIdx    = 1
	namePrefix = "grid"
)

func Run(ctx context.Context, params []string, arg string, env []string) error {
	resolvedEnv := make([]string, len(env))
	for i, val := range env {
		val = strings.ReplaceAll(val, "$_BASENAME", path.Base(arg))
		val = strings.ReplaceAll(val, "$_ARG", arg)

		names := make(map[int]string)
		ports := make(map[int]int)
		for i := 0; i < 10; i++ {
			n := strconv.Itoa(i)
			if strings.Contains(val, "$_NAME"+n) {
				if names[i] == "" {
					names[i] = fmt.Sprintf("%s%d", namePrefix, nameIdx)
					nameIdx++
				}
				val = strings.ReplaceAll(val, "$_NAME"+n, names[i])
			}

			if strings.Contains(val, "$_PORT"+n) {
				if ports[i] == 0 {
					ports[i] = NextPort()
					defer DonePort(ports[i])
				}
				val = strings.ReplaceAll(val, "$_PORT"+n, strconv.Itoa(ports[i]))
			}
		}

		resolvedEnv[i] = val
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd := exec.CommandContext(ctx, params[0], params[1:]...)
	cmd.Env = append(os.Environ(), resolvedEnv...)

	data, err := cmd.CombinedOutput()
	fmt.Println(string(data))
	if err != nil {
		return err
	}

	return nil
}
