package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// fetchFile will download and open a file with the contents of `url`.
func fetchFile(url string) (*os.File, int64, error) {
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
