package graphqlapp

import (
	"context"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
)

type (
	Destination      App
	DestinationInput App
)

func (a *App) Destination() graphql2.DestinationResolver           { return (*Destination)(a) }
func (a *App) DestinationInput() graphql2.DestinationInputResolver { return (*DestinationInput)(a) }

func (a *DestinationInput) Values(ctx context.Context, obj *gadb.DestV1, values []graphql2.FieldValueInput) error {
	obj.Args = make(map[string]string, len(values))
	for _, val := range values {
		obj.Args[val.FieldID] = val.Value
	}
	return nil
}

func (a *Destination) Values(ctx context.Context, obj *gadb.DestV1) ([]graphql2.FieldValuePair, error) {
	if obj.Args != nil {
		pairs := make([]graphql2.FieldValuePair, 0, len(obj.Args))
		for k, v := range obj.Args {
			pairs = append(pairs, graphql2.FieldValuePair{FieldID: k, Value: v})
		}
		return pairs, nil
	}

	return nil, nil
}

// DisplayInfo will return the display information for a destination by mapping to Query.DestinationDisplayInfo.
func (a *Destination) DisplayInfo(ctx context.Context, obj *gadb.DestV1) (graphql2.InlineDisplayInfo, error) {
	info, err := (*Query)(a)._DestinationDisplayInfo(ctx, gadb.DestV1{Type: obj.Type, Args: obj.Args}, true)
	if err != nil {
		isUnsafe, safeErr := errutil.ScrubError(err)
		if isUnsafe {
			log.Log(ctx, err)
		}
		return &graphql2.DestinationDisplayInfoError{Error: safeErr.Error()}, nil
	}

	return info, nil
}

func (a *Query) DestinationDisplayInfo(ctx context.Context, dest gadb.DestV1) (*nfydest.DisplayInfo, error) {
	return a._DestinationDisplayInfo(ctx, dest, false)
}

func (a *Query) _DestinationDisplayInfo(ctx context.Context, dest gadb.DestV1, skipValidation bool) (*nfydest.DisplayInfo, error) {
	app := (*App)(a)
	if !skipValidation {
		if err := app.ValidateDestination(ctx, "input", &dest); err != nil {
			return nil, err
		}
	}

	return app.DestReg.DisplayInfo(ctx, dest)
}
