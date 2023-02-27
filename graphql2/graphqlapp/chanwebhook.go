package graphqlapp

import (
	"context"
	_ "embed"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/search"
)

func (q *Query) ChanWebhooks(ctx context.Context, input *graphql2.ChanWebhookSearchOptions) (conn *graphql2.ChanWebhookConnection, err error) {
	if input == nil {
		input = &graphql2.ChanWebhookSearchOptions{}
	}

	var searchOpts webhook.SearchOptions

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

	if searchOpts.Limit == 0 {
		searchOpts.Limit = 15
	}

	if input.EscalationPolicyID != nil {
		searchOpts.EscalationPolicyID = *input.EscalationPolicyID
	}

	webhooks, err := q.ChanWebhookStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}

	conn = new(graphql2.ChanWebhookConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(webhooks) > searchOpts.Limit {
		webhooks = webhooks[:searchOpts.Limit]
		conn.PageInfo.HasNextPage = true
	}

	if len(webhooks) > 0 {
		searchOpts.After.Name = webhooks[len(webhooks)-1].Name
		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}

	conn.Nodes = webhooks
	return conn, err
}

func (q *Query) ChanWebhook(ctx context.Context, id string) (webhook *webhook.ChanWebhook, err error) {
	return q.ChanWebhookStore.FindOne(ctx, id)
}
