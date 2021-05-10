package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	tool := flag.String("t", "", "Tool to fetch.")
	version := flag.String("v", "", "Version of the tool to fetch.")
	output := flag.String("o", "", "Output file/dir.")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	if *tool == "" {
		log.Fatal("-t flag is required")
	}
	if *version == "" {
		log.Fatal("-v flag is required")
	}
	if *output == "" {
		log.Fatal("-o flag is required")
	}

	switch *tool {
	case "prometheus":
		getProm(*version, *output)
	case "protoc":
		getProtoC(*version, *output)
	default:
		log.Fatalf("unknown tool \"%s\"", *tool)
	}
}

func extractTGZFile(r *tar.Reader, src, dest string, exec bool) {
	os.MkdirAll(filepath.Dir(dest), 0755)
	tmpFile := dest + ".tmp"
	outFile, err := os.Create(tmpFile)
	if err != nil {
		log.Fatalf(`ERROR: create output file "%s": %v`, dest, err)
	}
	defer outFile.Close()
	for {
		hdr, err := r.Next()
		if err != nil {
			log.Fatalf("ERROR: find '%s' in tgz: %v", src, err)
		}
		ok, err := filepath.Match(src, hdr.Name)
		if err != nil {
			log.Fatalf("ERROR: invalid pattern '%s': %v", src, err)
		}
		if !ok {
			continue
		}

		break
	}

	_, err = io.Copy(outFile, r)
	if err != nil {
		log.Fatalf(`ERROR: extract "%s": %v`, src, err)
	}

	outFile.Close()
	if exec {
		os.Chmod(tmpFile, 0755)
	}
	err = os.Rename(tmpFile, dest)
	if err != nil {
		log.Fatalf(`ERROR: rename "%s" file: %v`, tmpFile, err)
	}
}

func extractZipFile(z *zip.Reader, src, dest string, exec bool) {
	os.MkdirAll(filepath.Dir(dest), 0755)
	tmpFile := dest + ".tmp"
	outFile, err := os.Create(tmpFile)
	if err != nil {
		log.Fatalf(`ERROR: create output file "%s": %v`, dest, err)
	}
	defer outFile.Close()
	bin, err := z.Open(src)
	if err != nil {
		log.Fatalf(`ERROR: find "%s" in zip: %v`, src, err)
	}
	defer bin.Close()
	_, err = io.Copy(outFile, bin)
	if err != nil {
		log.Fatalf(`ERROR: extract "%s": %v`, src, err)
	}
	outFile.Close()
	if exec {
		os.Chmod(tmpFile, 0755)
	}
	err = os.Rename(tmpFile, dest)
	if err != nil {
		log.Fatalf(`ERROR: rename "%s" file: %v`, tmpFile, err)
	}
}

func getFile(url string) (*os.File, int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, 0, fmt.Errorf("non-200 response: %s", resp.Status)
	}

	fd, err := ioutil.TempFile("", "*.zip")
	if err != nil {
		return nil, 0, fmt.Errorf("create temp file: %w", err)
	}

	n, err := io.Copy(fd, resp.Body)
	if err != nil {
		fd.Close()
		return nil, 0, fmt.Errorf("download file '%s': %w", url, err)
	}
	_, err = fd.Seek(0, 0)
	if err != nil {
		fd.Close()
		return nil, 0, fmt.Errorf("seek: %w", err)
	}

	return fd, n, nil
}

func getProm(version, output string) {
	url := fmt.Sprintf("https://github.com/prometheus/prometheus/releases/download/v%s/prometheus-%s.%s-%s.tar.gz",
		version, version, runtime.GOOS, runtime.GOARCH,
	)
	fd, _, err := getFile(url)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	defer fd.Close()

	gzr, err := gzip.NewReader(fd)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	defer gzr.Close()

	r := tar.NewReader(gzr)

	extractTGZFile(r, "*/prometheus", output, true)
}

func getProtoC(version, output string) {

	var variant string
	binFile := "bin/protoc"
	switch runtime.GOOS + "-" + runtime.GOARCH {
	case "linux-amd64":
		variant = "linux-x86_64"
	case "linux-386":
		variant = "linux-x86_32"
	case "linux-arm64":
		variant = "linux-"
	case "darwin-amd64", "darwin-arm64":
		variant = "osx-x86_64"
	case "windows-amd64":
		variant = "win64"
		binFile = "bin/protoc.exe"
	case "windows-386":
		variant = "win32"
		binFile = "bin/protoc.exe"
	default:
		log.Fatalf("unsupported OS/Arch combination")
	}

	url := fmt.Sprintf("https://github.com/protocolbuffers/protobuf/releases/download/v%s/protoc-%s-%s.zip",
		version, version,
		variant,
	)

	fd, n, err := getFile(url)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	defer fd.Close()

	z, err := zip.NewReader(fd, n)
	if err != nil {
		log.Fatalf("ERROR: unzip protoc: %v", err)
	}

	outDir := filepath.Dir(output)
	for _, f := range z.File {
		if strings.HasSuffix(f.Name, "/") {
			continue
		}
		if !strings.HasPrefix(f.Name, "include/") {
			continue
		}

		extractZipFile(z, f.Name, filepath.Join(outDir, f.Name), false)
	}

	extractZipFile(z, binFile, output, true)
}
