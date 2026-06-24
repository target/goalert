package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"runtime"
)

func getMailpit(version, output string) error {
	url := fmt.Sprintf("https://github.com/axllent/mailpit/releases/download/v%s/mailpit-%s-%s.tar.gz",
		version, runtime.GOOS, runtime.GOARCH,
	)
	fd, _, err := fetchFile(url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer fd.Close()

	gzr, err := gzip.NewReader(fd)
	if err != nil {
		return fmt.Errorf("deflate: %w", err)
	}
	defer gzr.Close()

	r := tar.NewReader(gzr)

	err = extractFromTar(r, "mailpit", output, true)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}
