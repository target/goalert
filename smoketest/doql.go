package smoketest

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

func contains(expectedErrors []string, err string) bool {
	for _, expectedErr := range expectedErrors {
		if strings.Contains(err, expectedErr) {
			return true
		}
	}
	return false
}

func DoGQL(t *testing.T, h *harness.Harness, query string, res interface{}, expectedErrs ...string) {
	t.Helper()
	g := h.GraphQLQuery2(query)

	if len(g.Errors) > 0 {
		for _, err := range g.Errors {
			if !contains(expectedErrs, err.Message) {
				t.Fatalf("unexpected graphql error: %s", err.Message)
			}
		}
	}

	t.Log("Response:", string(g.Data))
	if res == nil {
		return
	}

	err := json.Unmarshal(g.Data, &res)
	if err != nil {
		t.Fatal("failed to parse response:", err)
	}
}
