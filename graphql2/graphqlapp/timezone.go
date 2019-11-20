package graphqlapp

import (
	context "context"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/search"
	"github.com/target/goalert/timezone"
)

func (q *Query) TimeZones(ctx context.Context, input *graphql2.TimeZoneSearchOptions) (conn *graphql2.TimeZoneConnection, err error) {
	if input == nil {
		input = &graphql2.TimeZoneSearchOptions{}
	}

	var searchOpts timezone.SearchOptions
	if input.Search != nil {
		searchOpts.Search = *input.Search
	}
	searchOpts.Omit = input.Omit

	if input.After != nil && *input.After != "" {
		err = search.ParseCursor(*input.After, &searchOpts)
		if err != nil {
			return nil, err
		}
	}
	if input.First != nil {
		searchOpts.Limit = *input.First
	}
	if searchOpts.Limit == 0 {
		searchOpts.Limit = 15
	}

	searchOpts.Limit++
	names, err := q.TimeZoneStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}

	conn = new(graphql2.TimeZoneConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(names) == searchOpts.Limit {
		names = names[:len(names)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(names) > 0 {
		last := names[len(names)-1]
		searchOpts.After.Name = last

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = make([]graphql2.TimeZone, len(names))
	for i, n := range names {
		conn.Nodes[i].ID = n
	}
	return conn, err
}
