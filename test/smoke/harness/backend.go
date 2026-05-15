package harness

import (
	"encoding/json"
	"io"
	"strings"
)

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

		// Error is json.RawMessage because some log entries emit it as an object
		// (e.g. structured errors) rather than a string.
		var entry struct {
			Error        json.RawMessage
			Message      string `json:"msg"`
			Source       string
			Level        string
			SQL          string
			ProviderType string
			URL          string
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			h.t.Logf("Backend: failed to parse log entry: %v\n%s", err, string(raw))
			continue
		}

		if ignore(string(entry.Error)) {
			entry.Level = "ignore[" + entry.Level + "]"
		}
		if entry.Level == "error" || entry.Level == "fatal" {
			if entry.SQL != "" {
				// ignore printed SQL errors
				continue
			}
			h.t.Errorf("Backend: %s(%s) %s: %s\n%s", strings.ToUpper(entry.Level), entry.Source, string(entry.Error), entry.Message, string(raw))
			continue
		} else {
			h.t.Logf("Backend: %s %s", strings.ToUpper(entry.Level), entry.Message)
		}
	}
}
