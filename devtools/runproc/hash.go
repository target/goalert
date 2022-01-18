package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func fileHash(file string) []byte {
	fd, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer fd.Close()
	h := sha256.New224()
	io.Copy(h, fd)
	return h.Sum(nil)
}
func groupHash(patterns []string) string {
	if len(patterns) == 0 {
		return ""
	}

	var files []string
	for _, p := range patterns {
		_files, err := filepath.Glob(p)
		if err != nil {
			panic(err)
		}
		files = append(files, _files...)
	}

	h := sha256.New()
	sort.Strings(files)
	for _, f := range files {
		io.WriteString(h, f+"\n")
		h.Write(fileHash(f))
		io.WriteString(h, "\n")
	}

	return hex.EncodeToString(h.Sum(nil))
}
