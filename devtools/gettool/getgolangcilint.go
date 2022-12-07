package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"runtime"
)

func getGolangCiLint(version, output string) error {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	base := fmt.Sprintf("golangci-lint-%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://github.com/golangci/golangci-lint/releases/download/v%s/%s.%s",
		version, base, ext,
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
		err = extractFromZip(z, fmt.Sprintf("%s/golangci-lint.exe", base), output, true)
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

	err = extractFromTar(r, "*/golangci-lint", output, true)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	return nil
}
