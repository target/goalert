package main

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"runtime"
)

func getSqlc(version, output string) error {
	url := fmt.Sprintf("https://github.com/kyleconroy/sqlc/releases/download/v%s/sqlc_%s_%s_%s.zip",
		version, version, runtime.GOOS, runtime.GOARCH,
	)
	fd, n, err := fetchFile(url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer fd.Close()

	name := "sqlc"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	outDir := filepath.Dir(output)
	z, err := zip.NewReader(fd, n)
	if err != nil {
		return fmt.Errorf("unzip: %w", err)
	}

	err = extractFromZip(z, name, filepath.Join(outDir, name), true)
	if err != nil {
		return fmt.Errorf("extract bin: %w", err)
	}

	return nil
}
