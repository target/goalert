package harness

import "strings"

func containsAllIgnoreCase(s string, substrs []string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrs {
		if !strings.Contains(s, strings.ToLower(sub)) {
			return false
		}
	}

	return true
}

type messageMatcher struct {
	number   string
	keywords []string
}
type devMessage interface {
	To() string
	Body() string
}

func (m messageMatcher) match(msg devMessage) bool {
	return strings.TrimPrefix(msg.To(), "rcs:") == m.number && containsAllIgnoreCase(msg.Body(), m.keywords)
}

type anyMessage []messageMatcher

func (any anyMessage) match(msg devMessage) bool {
	for _, m := range any {
		if m.match(msg) {
			return true
		}
	}
	return false
}
