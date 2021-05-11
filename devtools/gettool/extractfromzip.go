package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// extractFromZip will extract `name` from the zip file and write it to `out`.
// If `exec` is set, the file will be marked executable.
func extractFromZip(z *zip.Reader, name, out string, exec bool) error {
	err := os.MkdirAll(filepath.Dir(out), 0755)
	if err != nil {
		return fmt.Errorf("create output dir '%s': %w", filepath.Dir(out), err)
	}

	tmpFile := out + ".tmp"
	mode := fs.FileMode(0666)
	if exec {
		mode = 0755
	}
	outFile, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create tmp file '%s': %w", tmpFile, err)
	}
	defer outFile.Close()

	bin, err := z.Open(name)
	if err != nil {
		return fmt.Errorf("find '%s' in zip: %w", name, err)
	}
	defer bin.Close()
	_, err = io.Copy(outFile, bin)
	if err != nil {
		return fmt.Errorf("extract '%s' to '%s': %w", name, tmpFile, err)
	}

	err = outFile.Close()
	if err != nil {
		return fmt.Errorf("close '%s': %w", tmpFile, err)
	}

	err = os.Rename(tmpFile, out)
	if err != nil {
		return fmt.Errorf(`ERROR: rename '%s' to '%s': %w`, tmpFile, out, err)
	}

	return nil
}
