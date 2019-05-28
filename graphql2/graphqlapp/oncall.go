package graphqlapp

import (
	context "context"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/user"
)

type OnCallShift App

func (a *App) OnCallShift() graphql2.OnCallShiftResolver { return (*OnCallShift)(a) }

func (oc *OnCallShift) User(ctx context.Context, raw *oncall.Shift) (*user.User, error) {
	return (*App)(oc).FindOneUser(ctx, raw.UserID)
}
