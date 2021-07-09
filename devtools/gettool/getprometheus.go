package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"runtime"
)

func getPrometheus(version, output string) error {
	url := fmt.Sprintf("https://github.com/prometheus/prometheus/releases/download/v%s/prometheus-%s.%s-%s.tar.gz",
		version, version, runtime.GOOS, runtime.GOARCH,
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

	err = extractFromTar(r, "*/prometheus", output, true)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}
