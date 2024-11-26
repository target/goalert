package graphqlapp

import (
	context "context"
	"database/sql"

	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/label"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation"
)

func (q *Query) LabelKeys(ctx context.Context, input *graphql2.LabelKeySearchOptions) (conn *graphql2.StringConnection, err error) {
	if input == nil {
		input = &graphql2.LabelKeySearchOptions{}
	}

	var opts label.KeySearchOptions
	if input.Search != nil {
		opts.Search = *input.Search
	}
	opts.Omit = input.Omit
	if input.After != nil && *input.After != "" {
		err = search.ParseCursor(*input.After, &opts)
		if err != nil {
			return conn, err
		}
	}
	if input.First != nil {
		opts.Limit = *input.First
	}
	if opts.Limit == 0 {
		opts.Limit = 15
	}

	opts.Limit++
	labelKeys, err := q.LabelStore.SearchKeys(ctx, &opts)
	if err != nil {
		return nil, err
	}
	conn = new(graphql2.StringConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(labelKeys) == opts.Limit {
		labelKeys = labelKeys[:len(labelKeys)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(labelKeys) > 0 {
		last := labelKeys[len(labelKeys)-1]
		opts.After = last

		cur, err := search.Cursor(opts)
		if err != nil {
			return nil, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = labelKeys
	return conn, err
}

func (q *Query) LabelValues(ctx context.Context, input *graphql2.LabelValueSearchOptions) (conn *graphql2.StringConnection, err error) {
	if input == nil {
		input = &graphql2.LabelValueSearchOptions{}
	}

	var opts label.ValueSearchOptions
	if input.Search != nil {
		opts.Search = *input.Search
	}
	opts.Omit = input.Omit
	opts.Key = input.Key
	if input.After != nil && *input.After != "" {
		err = search.ParseCursor(*input.After, &opts)
		if err != nil {
			return conn, err
		}
	}
	if input.First != nil {
		opts.Limit = *input.First
	}
	if opts.Limit == 0 {
		opts.Limit = 15
	}

	opts.Limit++
	values, err := q.LabelStore.SearchValues(ctx, &opts)
	if err != nil {
		return nil, err
	}
	conn = new(graphql2.StringConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(values) == opts.Limit {
		values = values[:len(values)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(values) > 0 {
		last := values[len(values)-1]
		opts.After = last

		cur, err := search.Cursor(opts)
		if err != nil {
			return nil, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = values
	return conn, err
}

func (q *Query) Labels(ctx context.Context, input *graphql2.LabelSearchOptions) (conn *graphql2.LabelConnection, err error) {
	if input == nil {
		input = &graphql2.LabelSearchOptions{}
	}
	keyConn, err := q.LabelKeys(ctx, &graphql2.LabelKeySearchOptions{
		Search: input.Search,
		After:  input.After,
		Omit:   input.Omit,
		First:  input.First,
	})
	if err != nil {
		return nil, err
	}

	conn = new(graphql2.LabelConnection)
	conn.PageInfo = keyConn.PageInfo
	conn.Nodes = make([]label.Label, len(keyConn.Nodes))
	for i, k := range keyConn.Nodes {
		conn.Nodes[i].Key = k
	}

	return conn, nil
}

func (m *Mutation) SetLabel(ctx context.Context, input graphql2.SetLabelInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		cfg := config.FromContext(ctx)
		if cfg.General.DisableLabelCreation {
			allLabels, err := m.LabelStore.UniqueKeysTx(ctx, gadb.Compat(tx))
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

		return m.LabelStore.SetTx(ctx, gadb.Compat(tx), &label.Label{
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
