package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var errMiss = fmt.Errorf("file not found in cache")

func hash256(in string) string {
	sum := sha256.Sum256([]byte(in))
	return hex.EncodeToString(sum[:])
}

func getCacheFile(url string) (*os.File, int64, error) {
	if cacheDir == "" {
		return nil, 0, errMiss
	}

	file := filepath.Join(cacheDir, hash256(url)+".data")
	fd, err := os.Open(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, errMiss
	}
	info, err := fd.Stat()
	if err != nil {
		return nil, 0, fmt.Errorf("stat: %w", err)
	}
	return fd, info.Size(), nil
}

func mkCacheFile(url string) (*os.File, error) {
	file := filepath.Join(cacheDir, hash256(url)+".data.tmp")
	err := os.MkdirAll(cacheDir, 0o755)
	if err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	fd, err := os.Create(file)
	if err != nil {
		return nil, fmt.Errorf("create cache file: %w", err)
	}
	return fd, nil
}

// fetchFile will download and open a file with the contents of `url`.
func fetchFile(url string) (*os.File, int64, error) {
	file, size, err := getCacheFile(url)
	if err == nil {
		return file, size, nil
	}
	if err != errMiss {
		return nil, 0, fmt.Errorf("getCacheFile: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, 0, fmt.Errorf("non-200 response: %s", resp.Status)
	}

	var fd *os.File
	var cacheFailed bool
	if cacheDir != "" {
		fd, err = mkCacheFile(url)
		if err != nil {
			log.Println("failed to update cache:", err)
		}
	}

	if fd == nil {
		fd, err = os.CreateTemp("", "*.zip")
		if err != nil {
			return nil, 0, fmt.Errorf("create temp file: %w", err)
		}
		cacheFailed = true
	}

	n, err := io.Copy(fd, resp.Body)
	if err != nil {
		fd.Close()
		return nil, 0, fmt.Errorf("download file '%s': %w", url, err)
	}

	if !cacheFailed {
		err = os.Rename(fd.Name(), filepath.Join(cacheDir, hash256(url)+".data"))
		if err != nil {
			fd.Close()
			return nil, 0, fmt.Errorf("rename cache file: %w", err)
		}
	}

	_, err = fd.Seek(0, 0)
	if err != nil {
		fd.Close()
		return nil, 0, fmt.Errorf("seek: %w", err)
	}

	return fd, n, nil
}
