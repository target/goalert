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

func (ig messageMatcher) match(msg devMessage) bool {
	return msg.To() == ig.number && containsAllIgnoreCase(msg.Body(), ig.keywords)
}

type anyMessage []messageMatcher

func (any anyMessage) match(msg devMessage) bool {
	for _, ig := range any {
		if ig.match(msg) {
			return true
		}
	}
	return false
}
