package graphqlapp

import (
	context "context"

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

const requestLoadersKey = dataLoaderKey(1)

type loaders struct {
	Alert                     *dataloader.Loader[int, alert.Alert]
	AlertState                *dataloader.Loader[int, alert.State]
	EP                        *dataloader.Loader[string, escalation.Policy]
	Rotation                  *dataloader.Loader[string, rotation.Rotation]
	Schedule                  *dataloader.Loader[string, schedule.Schedule]
	Service                   *dataloader.Loader[string, service.Service]
	User                      *dataloader.Loader[string, user.User]
	CM                        *dataloader.Loader[string, contactmethod.ContactMethod]
	Heartbeat                 *dataloader.Loader[string, heartbeat.Monitor]
	NotificationMessageStatus *dataloader.Loader[string, notification.SendResult]
	NC                        *dataloader.Loader[string, notificationchannel.Channel]
	AlertMetrics              *dataloader.Loader[int, alertmetrics.Metric]
	AlertFeedback             *dataloader.Loader[int, alert.Feedback]
	AlertMetadata             *dataloader.Loader[int, alert.MetadataAlertID]
}

func (a *App) registerLoaders(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, requestLoadersKey, &loaders{
		Alert:                     dataloader.NewStoreLoader(ctx, a.AlertStore.FindMany, func(a alert.Alert) int { return a.ID }),
		AlertState:                dataloader.NewStoreLoader(ctx, a.AlertStore.State, func(s alert.State) int { return s.ID }),
		EP:                        dataloader.NewStoreLoader(ctx, a.PolicyStore.FindManyPolicies, func(p escalation.Policy) string { return p.ID }),
		Rotation:                  dataloader.NewStoreLoader(ctx, a.RotationStore.FindMany, func(r rotation.Rotation) string { return r.ID }),
		Schedule:                  dataloader.NewStoreLoader(ctx, a.ScheduleStore.FindMany, func(s schedule.Schedule) string { return s.ID }),
		Service:                   dataloader.NewStoreLoader(ctx, a.ServiceStore.FindMany, func(s service.Service) string { return s.ID }),
		User:                      dataloader.NewStoreLoader(ctx, a.UserStore.FindMany, func(u user.User) string { return u.ID }),
		CM:                        dataloader.NewStoreLoaderWithDB(ctx, a.DB, a.CMStore.FindMany, func(cm contactmethod.ContactMethod) string { return cm.ID.String() }),
		Heartbeat:                 dataloader.NewStoreLoader(ctx, a.HeartbeatStore.FindMany, func(hb heartbeat.Monitor) string { return hb.ID }),
		NotificationMessageStatus: dataloader.NewStoreLoader(ctx, a.NotificationStore.FindManyMessageStatuses, func(n notification.SendResult) string { return n.ID }),
		NC:                        dataloader.NewStoreLoader(ctx, a.NCStore.FindMany, func(nc notificationchannel.Channel) string { return nc.ID.String() }),
		AlertMetrics:              dataloader.NewStoreLoader(ctx, a.AlertMetricsStore.FindMetrics, func(m alertmetrics.Metric) int { return m.ID }),
		AlertFeedback:             dataloader.NewStoreLoader(ctx, a.AlertStore.Feedback, func(f alert.Feedback) int { return f.ID }),
		AlertMetadata: dataloader.NewStoreLoader(ctx, func(ctx context.Context, i []int) ([]alert.MetadataAlertID, error) {
			return a.AlertStore.FindManyMetadata(ctx, a.DB, i)
		}, func(md alert.MetadataAlertID) int { return int(md.ID) }),
	})
	return ctx
}

func loadersFrom(ctx context.Context) loaders {
	loader, ok := ctx.Value(requestLoadersKey).(*loaders)
	if !ok {
		return loaders{}
	}
	if loader == nil {
		return loaders{}
	}

	return *loader
}

func (a *App) closeLoaders(ctx context.Context) {
	loader := loadersFrom(ctx)

	if loader.Alert != nil {
		loader.Alert.Close()
	}
	if loader.AlertState != nil {
		loader.AlertState.Close()
	}
	if loader.EP != nil {
		loader.EP.Close()
	}
	if loader.Rotation != nil {
		loader.Rotation.Close()
	}
	if loader.Schedule != nil {
		loader.Schedule.Close()
	}
	if loader.Service != nil {
		loader.Service.Close()
	}
	if loader.User != nil {
		loader.User.Close()
	}
	if loader.CM != nil {
		loader.CM.Close()
	}
	if loader.Heartbeat != nil {
		loader.Heartbeat.Close()
	}
	if loader.NotificationMessageStatus != nil {
		loader.NotificationMessageStatus.Close()
	}
	if loader.NC != nil {
		loader.NC.Close()
	}
	if loader.AlertMetrics != nil {
		loader.AlertMetrics.Close()
	}
	if loader.AlertFeedback != nil {
		loader.AlertFeedback.Close()
	}
	if loader.AlertMetadata != nil {
		loader.AlertMetadata.Close()
	}
}

func (app *App) FindOneAlertMetadata(ctx context.Context, id int) (map[string]string, error) {
	loader := loadersFrom(ctx).AlertMetadata
	if loader == nil {
		return app.AlertStore.Metadata(ctx, app.DB, id)
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
	loader := loadersFrom(ctx).NotificationMessageStatus
	if loader == nil {
		ms, err := app.NotificationStore.FindManyMessageStatuses(ctx, []string{id})
		if err != nil {
			return nil, err
		}
		return &ms[0], nil
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneAlertFeedback(ctx context.Context, id int) (*alert.Feedback, error) {
	loader := loadersFrom(ctx).AlertFeedback
	if loader == nil {
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
	loader := loadersFrom(ctx).Rotation
	if loader == nil {
		return app.RotationStore.FindRotation(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneSchedule(ctx context.Context, id string) (*schedule.Schedule, error) {
	loader := loadersFrom(ctx).Schedule
	if loader == nil {
		return app.ScheduleStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneUser(ctx context.Context, id string) (*user.User, error) {
	loader := loadersFrom(ctx).User
	if loader == nil {
		return app.UserStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneAlertMetric(ctx context.Context, id int) (*alertmetrics.Metric, error) {
	loader := loadersFrom(ctx).AlertMetrics
	if loader == nil {
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
	loader := loadersFrom(ctx).CM
	if loader == nil {
		return app.CMStore.FindOne(ctx, app.DB, id)
	}

	return loader.FetchOne(ctx, id.String())
}

// FindOneNC will return a single notification channel for the given id, using the contexts dataloader if enabled.
func (app *App) FindOneNC(ctx context.Context, id uuid.UUID) (*notificationchannel.Channel, error) {
	loader := loadersFrom(ctx).NC
	if loader == nil {
		return app.NCStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id.String())
}

func (app *App) FindOnePolicy(ctx context.Context, id string) (*escalation.Policy, error) {
	loader := loadersFrom(ctx).EP
	if loader == nil {
		return app.PolicyStore.FindOnePolicyTx(ctx, nil, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneService(ctx context.Context, id string) (*service.Service, error) {
	loader := loadersFrom(ctx).Service
	if loader == nil {
		return app.ServiceStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneAlertState(ctx context.Context, alertID int) (*alert.State, error) {
	loader := loadersFrom(ctx).AlertState
	if loader == nil {
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
	loader := loadersFrom(ctx).Alert
	if loader == nil {
		return app.AlertStore.FindOne(ctx, id)
	}

	return loader.FetchOne(ctx, id)
}

func (app *App) FindOneHeartbeatMonitor(ctx context.Context, id string) (*heartbeat.Monitor, error) {
	loader := loadersFrom(ctx).Heartbeat
	if loader == nil {
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
