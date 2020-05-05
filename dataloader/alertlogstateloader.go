package dataloader

import (
	"context"
	"strconv"
	"time"

	alertlog "github.com/target/goalert/alert/log"
)

type AlertLogStateLoader struct {
	*loader
	store alertlog.Store
}

func NewAlertLogStateLoader(ctx context.Context, store alertlog.Store) *AlertLogStateLoader {
	p := &AlertLogStateLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return strconv.Itoa(v.(*alertlog.LogState).LogID) },
		FetchFunc: p.fetch,
	})
	return p
}

func (l *AlertLogStateLoader) FetchOne(ctx context.Context, logID int) (*alertlog.LogState, error) {
	ls, err := l.loader.FetchOne(ctx, strconv.Itoa(logID))
	if err != nil {
		return nil, err
	}
	if ls == nil {
		return nil, err
	}
	return ls.(*alertlog.LogState), nil
}

func (l *AlertLogStateLoader) fetch(ctx context.Context, logIDs []string) ([]interface{}, error) {
	intIDs := make([]int, len(logIDs))
	for i, id := range logIDs {
		intIDs[i], _ = strconv.Atoi(id)
	}

	many, err := l.store.FindManyLogStates(ctx, intIDs)
	if err != nil {
		return nil, err
	}

	res := make([]interface{}, len(many))
	for i := range many {
		res[i] = &many[i]
	}
	return res, nil
}
