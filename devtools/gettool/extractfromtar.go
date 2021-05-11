package main

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// extractFromTar will extract a file matching `glob` to `out`.
// If `exec` is set, the file will be marked executable.
func extractFromTar(r *tar.Reader, glob, out string, exec bool) error {
	err := os.MkdirAll(filepath.Dir(out), 0755)
	if err != nil {
		return fmt.Errorf("make output dir '%s': %w", filepath.Dir(out), err)
	}

	tmpFile := out + ".tmp"
	mode := fs.FileMode(0666)
	if exec {
		mode = 0755
	}
	outFile, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create temp file '%s': %w", tmpFile, err)
	}
	defer outFile.Close()

	var name string
	for {
		hdr, err := r.Next()
		if err != nil {
			return fmt.Errorf("find '%s' in tgz: %w", glob, err)
		}
		ok, err := filepath.Match(glob, hdr.Name)
		if err != nil {
			return fmt.Errorf("invalid pattern '%s': %w", glob, err)
		}
		if !ok {
			continue
		}

		name = hdr.Name
		break
	}

	_, err = io.Copy(outFile, r)
	if err != nil {
		return fmt.Errorf("extract '%s' to '%s': %w", name, tmpFile, err)
	}

	err = outFile.Close()
	if err != nil {
		return fmt.Errorf("close '%s': %w", tmpFile, err)
	}

	err = os.Rename(tmpFile, out)
	if err != nil {
		return fmt.Errorf("rename '%s' to '%s': %w", tmpFile, out, err)
	}

	return nil
}
