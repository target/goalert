package harness

import (
	"encoding/json"
	"io"
	"strings"
)

// logEntry is the subset of a backend JSON log line that the harness inspects.
type logEntry struct {
	// Error, Err, and Panic are json.RawMessage because some log entries emit
	// them as objects (e.g. structured errors) rather than strings.
	//
	// Error is the key used by the legacy logger; Err and Panic are the keys
	// used by River's error handler (see riverErrs in app/initriver.go).
	Error json.RawMessage `json:"error"`
	Err   json.RawMessage `json:"err"`
	Panic json.RawMessage `json:"panic"`

	Message      string `json:"msg"`
	Source       string
	Level        string
	SQL          string
	ProviderType string
	URL          string
}

// parseLogEntry decodes a single JSON log line from the backend.
func parseLogEntry(raw []byte) (logEntry, error) {
	var entry logEntry
	err := json.Unmarshal(raw, &entry)
	return entry, err
}

// errText returns the raw JSON text of the entry's error, checking the
// `error`, `err`, and `panic` keys (in that order). It returns an empty string
// if none are present.
func (e logEntry) errText() string {
	for _, raw := range []json.RawMessage{e.Error, e.Err, e.Panic} {
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}
		return string(raw)
	}

	return ""
}

func (h *Harness) watchBackendLogs(r io.Reader) {
	defer close(h.logsDone)
	dec := json.NewDecoder(r)

	ignore := func(msg string) bool {
		h.mx.Lock()
		defer h.mx.Unlock()
		for _, s := range h.ignoreErrors {
			if strings.Contains(msg, s) {
				return true
			}
		}
		return false
	}

	h.IgnoreErrorsWith("rotation advanced late")

	for {
		var raw json.RawMessage
		err := dec.Decode(&raw)
		if err != nil {
			// Decoder is unrecoverable; must keep draining the pipe so backend
			// writers don't block forever. Without this, a parse failure here
			// can deadlock app startup/shutdown.
			if !h.isClosing() {
				h.t.Errorf("failed to read JSON logs: %v", err)
			}
			_, _ = io.Copy(io.Discard, r)
			return
		}

		entry, err := parseLogEntry(raw)
		if err != nil {
			h.t.Logf("Backend: failed to parse log entry: %v\n%s", err, string(raw))
			continue
		}

		errText := entry.errText()
		if ignore(errText) {
			entry.Level = "ignore[" + entry.Level + "]"
		}
		if entry.Level == "error" || entry.Level == "fatal" {
			if entry.SQL != "" {
				// ignore printed SQL errors
				continue
			}
			h.t.Errorf("Backend: %s(%s) %s: %s\n%s", strings.ToUpper(entry.Level), entry.Source, errText, entry.Message, string(raw))
			continue
		} else {
			h.t.Logf("Backend: %s %s", strings.ToUpper(entry.Level), entry.Message)
		}
	}
}
