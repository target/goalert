package graphqlapp

import (
	context "context"
	"io"

	"github.com/google/uuid"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/alertmetrics"
	"github.com/target/goalert/dataloader"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/service"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"

	"github.com/pkg/errors"
)

type dataLoaderKey int

const (
	dataLoaderKeyUnknown = dataLoaderKey(iota)

	dataLoaderKeyAlert
	dataLoaderKeyAlertState
	dataLoaderKeyEP
	dataLoaderKeyRotation
	dataLoaderKeySchedule
	dataLoaderKeyService
	dataLoaderKeyUser
	dataLoaderKeyCM
	dataLoaderKeyHeartbeatMonitor
	dataLoaderKeyNotificationMessageStatus
	dataLoaderKeyNC
	dataLoaderAlertMetrics
	dataLoaderAlertFeedback
	dataLoaderAlertMetadata

	dataLoaderKeyLast // always keep as last
)

func (a *App) registerLoaders(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, dataLoaderKeyAlert, dataloader.NewStoreLoaderInt(ctx, a.AlertStore.FindMany))
	ctx = context.WithValue(ctx, dataLoaderKeyAlertState, dataloader.NewStoreLoaderInt(ctx, a.AlertStore.State))
	ctx = context.WithValue(ctx, dataLoaderKeyEP, dataloader.NewStoreLoader(ctx, a.PolicyStore.FindManyPolicies))
	ctx = context.WithValue(ctx, dataLoaderKeyRotation, dataloader.NewStoreLoader(ctx, a.RotationStore.FindMany))
	ctx = context.WithValue(ctx, dataLoaderKeySchedule, dataloader.NewStoreLoader(ctx, a.ScheduleStore.FindMany))
	ctx = context.WithValue(ctx, dataLoaderKeyService, dataloader.NewStoreLoader(ctx, a.ServiceStore.FindMany))
	ctx = context.WithValue(ctx, dataLoaderKeyUser, dataloader.NewStoreLoader(ctx, a.UserStore.FindMany))
	ctx = context.WithValue(ctx, dataLoaderKeyCM, dataloader.NewStoreLoaderWithDB(ctx, a.DBTX, a.CMStore.FindMany))
	ctx = context.WithValue(ctx, dataLoaderKeyNotificationMessageStatus, dataloader.NewStoreLoader(ctx, a.NotificationStore.FindManyMessageStatuses))
	ctx = context.WithValue(ctx, dataLoaderKeyHeartbeatMonitor, dataloader.NewStoreLoader(ctx, a.HeartbeatStore.FindMany))
	ctx = context.WithValue(ctx, dataLoaderKeyNC, dataloader.NewStoreLoader(ctx, a.NCStore.FindMany))
	ctx = context.WithValue(ctx, dataLoaderAlertMetrics, dataloader.NewStoreLoaderInt(ctx, a.AlertMetricsStore.FindMetrics))
	ctx = context.WithValue(ctx, dataLoaderAlertFeedback, dataloader.NewStoreLoaderInt(ctx, a.AlertStore.Feedback))
	ctx = context.WithValue(ctx, dataLoaderAlertMetadata, dataloader.NewStoreLoaderInt(ctx, func(ctx context.Context, i []int) ([]alert.MetadataAlertID, error) {
		return a.AlertStore.FindManyMetadata(ctx, a.DBTX, i)
	}))
	return ctx
}

func (a *App) closeLoaders(ctx context.Context) {
	for key := dataLoaderKeyUnknown; key < dataLoaderKeyLast; key++ {
		loader, ok := ctx.Value(key).(io.Closer)
		if !ok {
			continue
		}
		loader.Close()
	}
}

func (app *App) FindOneAlertMetadata(ctx context.Context, id int) (map[string]string, error) {
	loader, ok := ctx.Value(dataLoaderAlertMetadata).(*dataloader.Loader[int, alert.MetadataAlertID])
	if !ok {
		return app.AlertStore.Metadata(ctx, app.DBTX, id)
	}

	md, err := loader.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if md == nil {
		return map[string]string{}, nil
	}

	return md.Meta, nil
}

func (app *App) FindOneNotificationMessageStatus(ctx context.Context, id string) (*notification.SendResult, error) {
	loader, ok := ctx.Value(dataLoaderKeyNotificationMessageStatus).(*dataloader.Loader[string, notification.SendResult])
	if !ok {
		ms, err := app.NotificationStore.FindManyMessageStatuses(ctx, []string{id})
		if err != nil {
			return nil, err
		}
		return &ms[0], nil
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneAlertFeedback(ctx context.Context, id int) (*alert.Feedback, error) {
	loader, ok := ctx.Value(dataLoaderAlertFeedback).(*dataloader.Loader[int, alert.Feedback])
	if !ok {
		feedback, err := app.AlertStore.Feedback(ctx, []int{id})
		if err != nil {
			return nil, err
		}
		if len(feedback) == 0 {
			return nil, nil
		}
		return &feedback[0], nil
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneRotation(ctx context.Context, id string) (*rotation.Rotation, error) {
	loader, ok := ctx.Value(dataLoaderKeyRotation).(*dataloader.Loader[string, rotation.Rotation])
	if !ok {
		return app.RotationStore.FindRotation(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneSchedule(ctx context.Context, id string) (*schedule.Schedule, error) {
	loader, ok := ctx.Value(dataLoaderKeySchedule).(*dataloader.Loader[string, schedule.Schedule])
	if !ok {
		return app.ScheduleStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneUser(ctx context.Context, id string) (*user.User, error) {
	loader, ok := ctx.Value(dataLoaderKeyUser).(*dataloader.Loader[string, user.User])
	if !ok {
		return app.UserStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneAlertMetric(ctx context.Context, id int) (*alertmetrics.Metric, error) {
	loader, ok := ctx.Value(dataLoaderAlertMetrics).(*dataloader.Loader[int, alertmetrics.Metric])
	if !ok {
		m, err := app.AlertMetricsStore.FindMetrics(ctx, []int{id})
		if err != nil {
			return nil, err
		}
		if len(m) == 0 {
			return nil, nil
		}
		return &m[0], nil
	}

	return loader.FetchOne(ctx, id)
}

// FindOneCM will return a single contact method for the given id, using the contexts dataloader if enabled.
func (app *App) FindOneCM(ctx context.Context, id uuid.UUID) (*contactmethod.ContactMethod, error) {
	loader, ok := ctx.Value(dataLoaderKeyCM).(*dataloader.Loader[uuid.UUID, contactmethod.ContactMethod])
	if !ok {
		return app.CMStore.FindOne(ctx, app.DBTX, id)
	}

	return loader.FetchOne(ctx, id)
}

// FindOneNC will return a single notification channel for the given id, using the contexts dataloader if enabled.
func (app *App) FindOneNC(ctx context.Context, id uuid.UUID) (*notificationchannel.Channel, error) {
	loader, ok := ctx.Value(dataLoaderKeyNC).(*dataloader.Loader[uuid.UUID, notificationchannel.Channel])
	if !ok {
		return app.NCStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOnePolicy(ctx context.Context, id string) (*escalation.Policy, error) {
	loader, ok := ctx.Value(dataLoaderKeyEP).(*dataloader.Loader[string, escalation.Policy])
	if !ok {
		return app.PolicyStore.FindOnePolicyTx(ctx, nil, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneService(ctx context.Context, id string) (*service.Service, error) {
	loader, ok := ctx.Value(dataLoaderKeyService).(*dataloader.Loader[string, service.Service])
	if !ok {
		return app.ServiceStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneAlertState(ctx context.Context, alertID int) (*alert.State, error) {
	loader, ok := ctx.Value(dataLoaderKeyAlertState).(*dataloader.Loader[int, alert.State])
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

	return loader.FetchOne(ctx, alertID)
}

func (app *App) FindOneAlert(ctx context.Context, id int) (*alert.Alert, error) {
	loader, ok := ctx.Value(dataLoaderKeyAlert).(*dataloader.Loader[int, alert.Alert])
	if !ok {
		return app.AlertStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneHeartbeatMonitor(ctx context.Context, id string) (*heartbeat.Monitor, error) {
	loader, ok := ctx.Value(dataLoaderKeyHeartbeatMonitor).(*dataloader.Loader[string, heartbeat.Monitor])
	if !ok {
		hb, err := app.HeartbeatStore.FindMany(ctx, []string{id})
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
