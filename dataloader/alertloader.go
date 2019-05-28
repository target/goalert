package dataloader

import (
	"context"
	"github.com/target/goalert/alert"
	"strconv"
	"time"
)

type AlertLoader struct {
	alertLoader *loader
	stateLoader *loader

	store alert.Store
}

func NewAlertLoader(ctx context.Context, store alert.Store) *AlertLoader {
	p := &AlertLoader{
		store: store,
	}
	p.alertLoader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return strconv.Itoa(v.(*alert.Alert).ID) },
		FetchFunc: p.fetchAlerts,
	})
	p.stateLoader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return strconv.Itoa(v.(*alert.State).AlertID) },
		FetchFunc: p.fetchAlertsState,
	})
	return p
}

func (l *AlertLoader) FetchOneAlert(ctx context.Context, id int) (*alert.Alert, error) {
	v, err := l.alertLoader.FetchOne(ctx, strconv.Itoa(id))
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, err
	}
	return v.(*alert.Alert), nil
}
func (l *AlertLoader) FetchOneAlertState(ctx context.Context, alertID int) (*alert.State, error) {
	v, err := l.stateLoader.FetchOne(ctx, strconv.Itoa(alertID))
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, err
	}
	return v.(*alert.State), nil
}

func (l *AlertLoader) fetchAlerts(ctx context.Context, ids []string) ([]interface{}, error) {
	intIDs := make([]int, len(ids))
	for i, id := range ids {
		intIDs[i], _ = strconv.Atoi(id)
	}
	many, err := l.store.FindMany(ctx, intIDs)
	if err != nil {
		return nil, err
	}

	res := make([]interface{}, len(many))
	for i := range many {
		res[i] = &many[i]
	}
	return res, nil
}

func (l *AlertLoader) fetchAlertsState(ctx context.Context, ids []string) ([]interface{}, error) {
	intIDs := make([]int, len(ids))
	for i, id := range ids {
		intIDs[i], _ = strconv.Atoi(id)
	}
	many, err := l.store.State(ctx, intIDs)
	if err != nil {
		return nil, err
	}

	res := make([]interface{}, len(many))
	for i := range many {
		res[i] = &many[i]
	}
	return res, nil
}
