package smoketest

import (
	"encoding/json"
	"github.com/target/goalert/smoketest/harness"
	"sort"
	"testing"
)

// TestSlackChannels tests that slack channels are returned for configured users.
func TestSlackChannels(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, "", "")
	defer h.Close()

	ch := []harness.SlackChannel{
		h.Slack().Channel("foo"),
		h.Slack().Channel("bar"),
		h.Slack().Channel("baz"),
	}
	sort.Slice(ch, func(i, j int) bool { return ch[i].ID() < ch[j].ID() })

	resp := h.GraphQLQuery2(`{slackChannels{nodes{id,name}}}`)
	for _, err := range resp.Errors {
		t.Error("graphql:", err.Message)
	}

	var data struct {
		SlackChannels struct {
			Nodes []struct {
				ID, Name string
			}
		}
	}
	err := json.Unmarshal(resp.Data, &data)
	if err != nil {
		t.Fatal("parse graphql response:", err)
	}
	channels := data.SlackChannels.Nodes
	sort.Slice(channels, func(i, j int) bool { return channels[i].ID < channels[j].ID })

	if len(channels) > len(ch) {
		for _, n := range channels[len(ch)-1:] {
			t.Errorf("got extra channel: ID=%s, Name=%s", n.ID, n.Name)
		}
		channels = channels[:len(ch)]
	}
	if len(channels) < len(ch) {
		for _, c := range ch {
			t.Errorf("missing channel: ID=%s, Name=%s", c.ID(), c.Name())
		}
		ch = ch[:len(channels)]
	}

	for i, n := range channels {
		c := ch[i]
		if c.ID() != n.ID {
			t.Errorf("channel[%d].ID: got %s; want %s", i, n.ID, c.ID())
		}
		if c.Name() != n.Name {
			t.Errorf("channel[%d].Name: got %s; want %s", i, n.Name, c.Name())
		}
	}
}
