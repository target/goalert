package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	tool := flag.String("t", "", "Tool to fetch.")
	version := flag.String("v", "", "Version of the tool to fetch.")
	output := flag.String("o", "", "Output file.")
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
	case "protoc":
		getProtoC(*version, *output)
	default:
		log.Fatalf("unknown tool \"%s\"", *tool)
	}
}

func getProtoC(version, output string) {

	var variant string
	binFile := "bin/protoc"
	switch runtime.GOOS + "-" + runtime.GOARCH {
	case "linux-amd64":
		variant = "linux-x86_64"
	case "linux-386":
		variant = "linux-x86_32"
	case "darwin-amd64":
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

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln("ERROR: fetch: ", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("ERROR: non-200 response: %s", resp.Status)
	}

	fd, err := ioutil.TempFile("", "protoc.zip")
	if err != nil {
		log.Fatalf("ERROR: create temp file: %v", err)
	}
	defer fd.Close()

	n, err := io.Copy(fd, resp.Body)
	if err != nil {
		log.Fatalf("ERROR: download protoc binary: %v", err)
	}

	z, err := zip.NewReader(fd, n)
	if err != nil {
		log.Fatalf("ERROR: unzip protoc: %v", err)
	}

	os.MkdirAll(filepath.Dir(output), 0755)

	outFile, err := os.Create(output + ".tmp")
	if err != nil {
		log.Fatalf("ERROR: create output: %v", err)
	}
	defer outFile.Close()
	bin, err := z.Open(binFile)
	if err != nil {
		log.Fatalf("ERROR: find protoc binary in zip: %v", err)
	}
	defer bin.Close()
	_, err = io.Copy(outFile, bin)
	if err != nil {
		log.Fatalf("ERROR: extract protoc binary: %v", err)
	}
	outFile.Close()
	os.Chmod(output+".tmp", 0755)
	err = os.Rename(output+".tmp", output)
	if err != nil {
		log.Fatalf("ERROR: rename output file: %v", err)
	}
}
