package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"path"
	"runtime"
	"strings"
)

func getK6(version, output string) error {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	osName := runtime.GOOS
	if osName == "darwin" {
		osName = "macos"
		ext = "zip"
	}

	url := fmt.Sprintf("https://github.com/grafana/k6/releases/download/v%s/k6-v%s-%s-%s.%s",
		version, version, osName, runtime.GOARCH, ext,
	)
	fd, n, err := fetchFile(url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer fd.Close()

	if ext == "zip" {
		z, err := zip.NewReader(fd, n)
		if err != nil {
			return fmt.Errorf("open zip: %w", err)
		}
		name := strings.TrimSuffix(path.Base(url), ".zip") + "/k6"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		err = extractFromZip(z, name, output, true)
		if err != nil {
			return fmt.Errorf("extract: %w", err)
		}

		return nil
	}

	gzr, err := gzip.NewReader(fd)
	if err != nil {
		return fmt.Errorf("deflate: %w", err)
	}
	defer gzr.Close()

	r := tar.NewReader(gzr)

	err = extractFromTar(r, "*/k6", output, true)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}
