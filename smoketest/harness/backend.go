package harness

import (
	"encoding/json"
	"io"
	"strings"
)

func (h *Harness) watchBackendLogs(r io.Reader, urlCh chan string) {
	defer close(urlCh)
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
	var err error
	var sent bool
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
		if !sent && entry.URL != "" {
			sent = true
			urlCh <- entry.URL
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
func (h *Harness) watchBackend(c io.Closer) {
	defer c.Close()
	err := h.cmd.Wait()
	if err != nil && !h.isClosing() {
		h.t.Errorf("backend failed: %v", err)
	}
}
