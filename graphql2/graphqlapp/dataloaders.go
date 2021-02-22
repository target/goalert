package graphqlapp

import (
	context "context"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/dataloader"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/service"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"

	"github.com/pkg/errors"
)

type dataLoaderKey int

const (
	dataLoaderKeyAlert = dataLoaderKey(iota)
	dataLoaderKeyEP
	dataLoaderKeyRotation
	dataLoaderKeySchedule
	dataLoaderKeyService
	dataLoaderKeyUser
	dataLoaderKeyCM
	dataLoaderKeyHeartbeatMonitor
	dataLoaderKeyNotificationMessageStatus
)

func (a *App) registerLoaders(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, dataLoaderKeyAlert, dataloader.NewAlertLoader(ctx, a.AlertStore))
	ctx = context.WithValue(ctx, dataLoaderKeyEP, dataloader.NewPolicyLoader(ctx, a.PolicyStore))
	ctx = context.WithValue(ctx, dataLoaderKeyRotation, dataloader.NewRotationLoader(ctx, a.RotationStore))
	ctx = context.WithValue(ctx, dataLoaderKeySchedule, dataloader.NewScheduleLoader(ctx, a.ScheduleStore))
	ctx = context.WithValue(ctx, dataLoaderKeyService, dataloader.NewServiceLoader(ctx, a.ServiceStore))
	ctx = context.WithValue(ctx, dataLoaderKeyUser, dataloader.NewUserLoader(ctx, a.UserStore))
	ctx = context.WithValue(ctx, dataLoaderKeyCM, dataloader.NewCMLoader(ctx, a.CMStore))
	ctx = context.WithValue(ctx, dataLoaderKeyNotificationMessageStatus, dataloader.NewNotificationMessageStatusLoader(ctx, a.NotificationStore))
	ctx = context.WithValue(ctx, dataLoaderKeyHeartbeatMonitor, dataloader.NewHeartbeatMonitorLoader(ctx, a.HeartbeatStore))
	return ctx
}

func (app *App) FindOneNotificationMessageStatus(ctx context.Context, id string) (*notification.MessageStatus, error) {
	loader, ok := ctx.Value(dataLoaderKeyNotificationMessageStatus).(*dataloader.NotificationMessageStatusLoader)
	if !ok {
		ms, err := app.NotificationStore.FindManyMessageStatuses(ctx, id)
		if err != nil {
			return nil, err
		}
		return &ms[0], nil
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneRotation(ctx context.Context, id string) (*rotation.Rotation, error) {
	loader, ok := ctx.Value(dataLoaderKeyRotation).(*dataloader.RotationLoader)
	if !ok {
		return app.RotationStore.FindRotation(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneSchedule(ctx context.Context, id string) (*schedule.Schedule, error) {
	loader, ok := ctx.Value(dataLoaderKeySchedule).(*dataloader.ScheduleLoader)
	if !ok {
		return app.ScheduleStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneUser(ctx context.Context, id string) (*user.User, error) {
	loader, ok := ctx.Value(dataLoaderKeyUser).(*dataloader.UserLoader)
	if !ok {
		return app.UserStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

// FindOneCM will return a single contact method for the given id, using the contexts dataloader if enabled.
func (app *App) FindOneCM(ctx context.Context, id string) (*contactmethod.ContactMethod, error) {
	loader, ok := ctx.Value(dataLoaderKeyUser).(*dataloader.CMLoader)
	if !ok {
		return app.CMStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOnePolicy(ctx context.Context, id string) (*escalation.Policy, error) {
	loader, ok := ctx.Value(dataLoaderKeyEP).(*dataloader.PolicyLoader)
	if !ok {
		return app.PolicyStore.FindOnePolicy(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneService(ctx context.Context, id string) (*service.Service, error) {
	loader, ok := ctx.Value(dataLoaderKeyService).(*dataloader.ServiceLoader)
	if !ok {
		return app.ServiceStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}
func (app *App) FindOneAlertState(ctx context.Context, alertID int) (*alert.State, error) {
	loader, ok := ctx.Value(dataLoaderKeyAlert).(*dataloader.AlertLoader)
	if !ok {
		epState, err := app.AlertStore.State(ctx, []int{alertID})
		if err != nil {
			return nil, err
		}
		if len(epState) == 0 {
			return nil, errors.New("no current epState for alert")
		}
		return &epState[0], nil
	}

	return loader.FetchOneAlertState(ctx, alertID)
}
func (app *App) FindOneAlert(ctx context.Context, id int) (*alert.Alert, error) {
	loader, ok := ctx.Value(dataLoaderKeyAlert).(*dataloader.AlertLoader)
	if !ok {
		return app.AlertStore.FindOne(ctx, id)
	}

	return loader.FetchOneAlert(ctx, id)
}

func (app *App) FindOneHeartbeatMonitor(ctx context.Context, id string) (*heartbeat.Monitor, error) {
	loader, ok := ctx.Value(dataLoaderKeyHeartbeatMonitor).(*dataloader.HeartbeatMonitorLoader)
	if !ok {
		hb, err := app.HeartbeatStore.FindMany(ctx, id)
		if err != nil {
			return nil, err
		}
		if len(hb) == 0 {
			return nil, nil
		}
		return &hb[0], nil
	}

	return loader.FetchOne(ctx, id)
}
