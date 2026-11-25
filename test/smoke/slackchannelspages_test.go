package smoke

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/target/goalert/app"
	"github.com/target/goalert/test/smoke/harness"
)

// TestSlackChannelsPages tests that listing slack channels fails when they
// span more pages than the configured limit, and that raising the limit
// allows them to be listed.
func TestSlackChannelsPages(t *testing.T) {
	t.Parallel()

	h := harness.NewStoppedHarness(t, "", nil, "")
	defer h.Close()

	// GoAlert requests 200 channels per page, so anything
	// over 2000 guarantees more than 10 pages.
	const numChannels = 2001
	for i := range numChannels {
		h.Slack().Channel(fmt.Sprintf("chan-%d", i))
	}

	const query = `{slackChannels(input:{first: 1}){nodes{id}}}`
	h.IgnoreErrorsWith("exceeded limit of 10 pages")
	h.StartWithAppCfgHook(func(cfg *app.Config) {
		cfg.SlackMaxChannelPages = 10
	})

	resp := h.GraphQLQuery2(query)
	const wantErr = "more channels than GoAlert is configured to load"
	if len(resp.Errors) != 1 || !strings.Contains(resp.Errors[0].Message, wantErr) {
		t.Fatalf("with limit of 10 pages: got errors %v; want one containing %q", resp.Errors, wantErr)
	}

	h.RestartGoAlertWithAppCfgHook(func(cfg *app.Config) {
		// The mock server returns a random number of channels per page (up to
		// the requested limit), so allow for the worst case of one per page.
		cfg.SlackMaxChannelPages = numChannels
	})

	resp = h.GraphQLQuery2(query)
	for _, err := range resp.Errors {
		t.Error("graphql:", err.Message)
	}

	var data struct {
		SlackChannels struct {
			Nodes []struct{ ID string }
		}
	}
	err := json.Unmarshal(resp.Data, &data)
	if err != nil {
		t.Fatal("parse graphql response:", err)
	}
	if len(data.SlackChannels.Nodes) != 1 {
		t.Errorf("got %d channels; want 1", len(data.SlackChannels.Nodes))
	}
}
