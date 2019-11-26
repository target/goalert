package graphqlapp

import (
	context "context"
	"database/sql"
	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/label"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation"
)

func (q *Query) Labels(ctx context.Context, input *graphql2.LabelSearchOptions) (conn *graphql2.LabelConnection, err error) {
	if input == nil {
		input = &graphql2.LabelSearchOptions{}
	}

	var searchOpts label.SearchOptions
	if input.Search != nil {
		searchOpts.Search = *input.Search
	}
	searchOpts.Omit = input.Omit
	if input.UniqueKeys != nil {
		searchOpts.UniqueKeys = *input.UniqueKeys
	}
	if input.After != nil && *input.After != "" {
		err = search.ParseCursor(*input.After, &searchOpts)
		if err != nil {
			return conn, err
		}
	}
	if input.First != nil {
		searchOpts.Limit = *input.First
	}
	if searchOpts.Limit == 0 {
		searchOpts.Limit = 15
	}

	searchOpts.Limit++
	labels, err := q.LabelStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}
	conn = new(graphql2.LabelConnection)
	if len(labels) == searchOpts.Limit {
		labels = labels[:len(labels)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(labels) > 0 {
		last := labels[len(labels)-1]
		searchOpts.After.Key = last.Key
		searchOpts.After.TargetType = last.Target.TargetType()
		searchOpts.After.TargetID = last.Target.TargetID()

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return nil, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = labels
	return conn, err
}
func (m *Mutation) SetLabel(ctx context.Context, input graphql2.SetLabelInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		cfg := config.FromContext(ctx)
		if cfg.General.DisableLabelCreation {
			allLabels, err := m.LabelStore.UniqueKeysTx(ctx, tx)
			if err != nil {
				return err
			}
			var keyExists bool
			for _, l := range allLabels {
				if input.Key == l {
					keyExists = true
					break
				}
			}
			if !keyExists {
				return validation.NewFieldError("Key", "Creating new labels is currently disabled.")
			}
		}

		return m.LabelStore.SetTx(ctx, tx, &label.Label{
			Key:    input.Key,
			Value:  input.Value,
			Target: input.Target,
		})
	})
	if err != nil {
		return false, err
	}

	return true, nil
}
