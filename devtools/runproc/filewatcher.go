package main

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Watcher struct {
	p []string

	state string
}

func calcTree(patterns []string) string {
	var files []string
	for _, p := range patterns {
		f, err := filepath.Glob(p)
		if err != nil {
			continue
		}
		files = append(files, f...)
	}

	sort.Strings(files)

	var t strings.Builder
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		t.WriteString(f + "\n")
		t.WriteString(info.ModTime().String() + "\n")
		t.WriteString(strconv.FormatInt(info.Size(), 10) + "\n")
	}

	return t.String()
}

func (w *Watcher) Changed() bool {
	tree := calcTree(w.p)
	if tree == w.state {
		return false
	}

	w.state = tree
	return true
}

func Watch(patterns []string) *Watcher {
	sort.Strings(patterns)

	return &Watcher{p: patterns, state: calcTree(patterns)}
}
