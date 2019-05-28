package graphqlapp

import (
	context "context"
	"sort"
	"strings"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/search"
)

func (q *Query) SlackChannel(ctx context.Context, id string) (*slack.Channel, error) {
	return q.SlackStore.Channel(ctx, id)
}

func (q *Query) SlackChannels(ctx context.Context, input *graphql2.SlackChannelSearchOptions) (conn *graphql2.SlackChannelConnection, err error) {
	if input == nil {
		input = &graphql2.SlackChannelSearchOptions{}
	}

	var searchOpts struct {
		Search string   `json:"s,omitempty"`
		Omit   []string `json:"m,omitempty"`
		After  struct {
			Name string `json:"n,omitempty"`
		} `json:"a,omitempty"`
	}
	searchOpts.Omit = input.Omit
	if input.Search != nil {
		searchOpts.Search = *input.Search
	}
	if input.After != nil && *input.After != "" {
		err = search.ParseCursor(*input.After, &searchOpts)
		if err != nil {
			return nil, err
		}
	}

	limit := 15
	if input.First != nil {
		limit = *input.First
	}

	channels, err := q.SlackStore.ListChannels(ctx)
	if err != nil {
		return nil, err
	}
	// Sort by name, case-insensitive, then sensitive.
	sort.Slice(channels, func(i, j int) bool {
		iName, jName := strings.ToLower(channels[i].Name), strings.ToLower(channels[j].Name)

		if iName != jName {
			return iName < jName
		}
		return channels[i].Name < channels[j].Name
	})

	// No DB search, so we manually filter for the cursor and search strings.
	s := strings.ToLower(searchOpts.Search)
	n := strings.ToLower(searchOpts.After.Name)
	filtered := channels[:0]
	for _, ch := range channels {
		chName := strings.ToLower(ch.Name)
		if !strings.Contains(chName, s) {
			continue
		}
		if n != "" && chName <= n {
			continue
		}
		if contains(searchOpts.Omit, ch.ID) {
			continue
		}
		filtered = append(filtered, ch)
	}
	channels = filtered

	conn = new(graphql2.SlackChannelConnection)
	if len(channels) > limit {
		channels = channels[:limit]
		conn.PageInfo.HasNextPage = true
	}

	if len(channels) > 0 {
		searchOpts.After.Name = channels[len(channels)-1].Name
		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}

	conn.Nodes = channels
	return conn, err
}
