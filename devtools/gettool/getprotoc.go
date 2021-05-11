package main

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

func getProtoC(version, output string) error {
	var variant string
	binFile := "bin/protoc"
	switch runtime.GOOS + "-" + runtime.GOARCH {
	case "linux-amd64":
		variant = "linux-x86_64"
	case "linux-386":
		variant = "linux-x86_32"
	case "linux-arm64":
		variant = "linux-aarch_64"
	case "darwin-amd64", "darwin-arm64":
		// TODO: use arm64 variant if/when supported by protoc.
		// M1 will work with `x86_64` for now.
		variant = "osx-x86_64"
	case "windows-amd64":
		variant = "win64"
		binFile = "bin/protoc.exe"
	case "windows-386":
		variant = "win32"
		binFile = "bin/protoc.exe"
	default:
		return fmt.Errorf("unsupported OS-Arch '%s'", runtime.GOOS+"-"+runtime.GOARCH)
	}

	url := fmt.Sprintf("https://github.com/protocolbuffers/protobuf/releases/download/v%s/protoc-%s-%s.zip",
		version, version,
		variant,
	)

	fd, n, err := fetchFile(url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer fd.Close()

	z, err := zip.NewReader(fd, n)
	if err != nil {
		return fmt.Errorf("unzip: %w", err)
	}

	outDir := filepath.Dir(output)
	for _, f := range z.File {
		if strings.HasSuffix(f.Name, "/") {
			continue
		}
		if !strings.HasPrefix(f.Name, "include/") {
			continue
		}

		err = extractFromZip(z, f.Name, filepath.Join(outDir, f.Name), false)
		if err != nil {
			return fmt.Errorf("extract lib: %w", err)
		}
	}

	err = extractFromZip(z, binFile, output, true)
	if err != nil {
		return fmt.Errorf("extract bin: %w", err)
	}

	return nil
}
