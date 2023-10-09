package graphqlapp

import (
	context "context"

	"github.com/target/goalert/expflag"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

type (
	Query        App
	DebugMessage App
)

func (a *App) Query() graphql2.QueryResolver { return (*Query)(a) }

func (q *Query) ListGQLFields(ctx context.Context, query *string) ([]string, error) {
	if query == nil || *query == "" {
		// List all fields if no query is provided.
		return graphql2.SchemaFields(), nil
	}

	return graphql2.QueryFields(*query)
}

func (a *Query) ExperimentalFlags(ctx context.Context) ([]string, error) {
	var flags []string
	for _, f := range expflag.AllFlags() {
		if !expflag.ContextHas(ctx, f) {
			continue
		}

		flags = append(flags, string(f))
	}

	return flags, nil
}

func (a *Query) LinkAccountInfo(ctx context.Context, token string) (*graphql2.LinkAccountInfo, error) {
	m, err := a.AuthLinkStore.FindLinkMetadata(ctx, token)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}

	info := &graphql2.LinkAccountInfo{
		UserDetails: m.UserDetails,
	}
	if m.AlertID > 0 {
		info.AlertID = &m.AlertID
	}
	var s graphql2.AlertStatus
	switch m.AlertAction {
	case notification.ResultAcknowledge.String():
		s = graphql2.AlertStatusStatusAcknowledged
		info.AlertNewStatus = &s
	case notification.ResultResolve.String():
		s = graphql2.AlertStatusStatusClosed
		info.AlertNewStatus = &s
	}

	return info, nil
}

func (a *Query) AuthSubjectsForProvider(ctx context.Context, _first *int, _after *string, providerID string) (conn *graphql2.AuthSubjectConnection, err error) {
	var first int
	var after string
	if _after != nil {
		after = *_after
	}
	if _first != nil {
		first = *_first
	} else {
		first = 15
	}
	err = validate.Range("First", first, 1, 300)
	if err != nil {
		return nil, err
	}

	var c struct {
		ProviderID string
		LastID     string
	}

	if after != "" {
		err = search.ParseCursor(after, &c)
		if err != nil {
			return nil, errors.Wrap(err, "parse cursor")
		}
	} else {
		c.ProviderID = providerID
	}

	conn = new(graphql2.AuthSubjectConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	conn.Nodes, err = a.UserStore.FindSomeAuthSubjectsForProvider(ctx, first+1, c.LastID, c.ProviderID)
	if err != nil {
		return nil, err
	}
	if len(conn.Nodes) > first {
		conn.Nodes = conn.Nodes[:first]
		conn.PageInfo.HasNextPage = true
	}
	if len(conn.Nodes) > 0 {
		c.LastID = conn.Nodes[len(conn.Nodes)-1].SubjectID
	}

	cur, err := search.Cursor(c)
	if err != nil {
		return nil, err
	}
	conn.PageInfo.EndCursor = &cur
	return conn, nil
}
