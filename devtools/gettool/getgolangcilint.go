package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"runtime"
)

func getGolangCiLint(version, output string) error {
    os := runtime.GOOS + "-" + runtime.GOARCH
	url := fmt.Sprintf("https://github.com/golangci/golangci-lint/releases/download/v%s/golangci-lint-%s-%s.tar.gz",
		version, version,
		os,
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

	err = extractFromTar(r, "*/golangci-lint", output, true)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}
