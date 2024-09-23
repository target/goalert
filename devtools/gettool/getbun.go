package main

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"runtime"
)

func getBun(version, output string) error {
	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "x64"
	case "arm64":
		arch = "aarch64"
	default:
		return fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
	}

	url := fmt.Sprintf("https://github.com/oven-sh/bun/releases/download/bun-v%s/bun-%s-%s.zip",
		version, runtime.GOOS, arch,
	)
	fd, n, err := fetchFile(url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer fd.Close()

	name := fmt.Sprintf("bun-%s-%s/bun", runtime.GOOS, arch)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	outDir := filepath.Dir(output)
	z, err := zip.NewReader(fd, n)
	if err != nil {
		return fmt.Errorf("unzip: %w", err)
	}

	err = extractFromZip(z, name, filepath.Join(outDir, "bun"), true)
	if err != nil {
		return fmt.Errorf("extract bin: %w", err)
	}

	return nil
}
