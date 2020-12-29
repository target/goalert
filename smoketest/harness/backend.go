package harness

import (
	"encoding/json"
	"io"
	"strings"
)

func (h *Harness) watchBackendLogs(r io.Reader) {
	dec := json.NewDecoder(r)
	var entry struct {
		Error        string
		Message      string `json:"msg"`
		Source       string
		Level        string
		ProviderType json.Number
		URL          string
	}

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

	var err error
	for {
		err = dec.Decode(&entry)
		if err != nil {
			break
		}
		if ignore(entry.Error) {
			entry.Level = "ignore[" + entry.Level + "]"
		}
		if entry.Level == "error" || entry.Level == "fatal" {
			h.t.Errorf("Backend: %s(%s) %s", strings.ToUpper(entry.Level), entry.Source, entry.Error)
			continue
		} else {
			h.t.Logf("Backend: %s %s", strings.ToUpper(entry.Level), entry.Message)
		}
	}
	if h.isClosing() {
		return
	}
	data := make([]byte, 32768)
	n, _ := dec.Buffered().Read(data)
	nx, _ := r.Read(data[n:])
	if n+nx > 0 {
		h.t.Logf("Buffered: %s", string(data[:n+nx]))
	}

	h.t.Errorf("failed to read/parse JSON logs: %v", err)
}
