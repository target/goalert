package graphqlapp

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/service"
	"github.com/target/goalert/service/rule"
	"github.com/target/goalert/signal"
)

type Signal App

func (a *App) Signal() graphql2.SignalResolver { return (*Signal)(a) }

func (s *Signal) ID(ctx context.Context, raw *signal.Signal) (string, error) {
	return strconv.FormatInt(raw.ID, 10), nil
}

func (s *Signal) SignalID(ctx context.Context, raw *signal.Signal) (int, error) {
	return int(raw.ID), nil
}

func (s *Signal) OutgoingPayload(ctx context.Context, raw *signal.Signal) (string, error) {
	payloadBytes, err := json.Marshal(raw.OutgoingPayload)
	if err != nil {
		return "", err
	}
	return string(payloadBytes), nil
}

func (s *Signal) Service(ctx context.Context, raw *signal.Signal) (*service.Service, error) {
	return (*App)(s).FindOneService(ctx, raw.ServiceID)
}

func (s *Signal) ServiceRule(ctx context.Context, raw *signal.Signal) (*rule.Rule, error) {
	return s.ServiceRuleStore.FindOne(ctx, raw.ServiceRuleID)
}

func (q *Query) Signal(ctx context.Context, id int) (*signal.Signal, error) {
	return q.SignalStore.FindOne(ctx, id)
}
