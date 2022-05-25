package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func isAlphaNum(r byte) bool {
	switch {
	case r >= 'a' && r <= 'z':
	case r >= 'A' && r <= 'Z':
	case r >= '0' && r <= '9':
	case r == '_':
	default:
		return false
	}
	return true
}

func Parse(r io.Reader) ([]Task, error) {
	s := bufio.NewScanner(r)
	var tasks []Task
	var t Task
	var line int
	for s.Scan() {
		line++
		str := strings.TrimSpace(s.Text())
		if str == "" {
			t = Task{}
			continue
		}
		if str[0] == '@' {
			// parameter
			parts := strings.SplitN(str[1:], "=", 2)
			switch strings.TrimSpace(parts[0]) {
			case "oneshot":
				t.OneShot = true
			case "watch-file":
				if len(parts) != 2 {
					return nil, fmt.Errorf("line %d: missing file path for watch-file", line)
				}
				t.WatchFiles = append(t.WatchFiles, parts[1])
			default:
				return nil, fmt.Errorf("line %d: invalid option '%s'", line, parts[0])
			}
		}
		if !isAlphaNum(str[0]) {
			// comment/ignored
			continue
		}

		parts := strings.SplitN(str, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: invalid proc definition '%s' (missing ':')", line, parts[0])
		}
		t.Name = strings.TrimSpace(parts[0])
		t.Command = strings.TrimSpace(parts[1])
		tasks = append(tasks, t)
		t = Task{}
	}
	return tasks, s.Err()
}
