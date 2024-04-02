package migratetest

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// CopyScanner will scan COPY statements, returning TableSnapshots.
type CopyScanner struct {
	s   *bufio.Scanner
	t   TableSnapshot
	err error
}

// NewCopyScanner will return a new CopyScanner.
func NewCopyScanner(r io.Reader) *CopyScanner {
	return &CopyScanner{s: bufio.NewScanner(r)}
}

func (d *CopyScanner) nextTableName() string {
	for d.s.Scan() {
		if !strings.HasPrefix(d.s.Text(), "COPY ") {
			continue
		}
		parts := strings.Split(d.s.Text(), `"`)
		if len(parts) < 3 {
			d.err = fmt.Errorf("invalid COPY line: %s", d.s.Text())
			return ""
		}
		return parts[1]
	}

	d.err = d.s.Err()
	return ""
}

func (d *CopyScanner) csvData() string {
	var b strings.Builder
	for d.s.Scan() {
		if d.s.Text() == "\\." {
			return b.String()
		}
		b.WriteString(d.s.Text())
		b.WriteByte('\n')
	}

	d.err = d.s.Err()
	return ""
}

// Scan will scan the next COPY statement.
func (d *CopyScanner) Scan() bool {
	if d.err != nil {
		return false
	}

	var t TableSnapshot
	t.Name = d.nextTableName()
	if t.Name == "" {
		return false
	}

	data := d.csvData()
	if data == "" {
		return false
	}
	r := csv.NewReader(strings.NewReader(data))
	t.Columns, d.err = r.Read()
	if d.err != nil {
		return false
	}
	t.Rows, d.err = r.ReadAll()
	if d.err != nil {
		return false
	}
	t.Sort()
	d.t = t

	return true
}

func (d *CopyScanner) Err() error           { return d.err }
func (d *CopyScanner) Table() TableSnapshot { return d.t }
