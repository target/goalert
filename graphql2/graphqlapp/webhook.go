package graphqlapp

import (
	"context"
	_ "embed"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/search"
)

func (q *Query) Webhooks(ctx context.Context, input *graphql2.WebhookSearchOptions) (conn *graphql2.WebhookConnection, err error) {
	if input == nil {
		input = &graphql2.WebhookSearchOptions{}
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

	webhooks, err := q.WebhookStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}

	conn = new(graphql2.WebhookConnection)
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

func (q *Query) Webhook(ctx context.Context, id string) (webhook *webhook.Webhook, err error) {
	return q.WebhookStore.FindOne(ctx, id)
}
