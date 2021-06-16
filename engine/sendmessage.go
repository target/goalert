package engine

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/engine/message"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"go.opencensus.io/trace"
)

func (p *Engine) sendMessage(ctx context.Context, msg *message.Message) (*notification.SendResult, error) {
	ctx, sp := trace.StartSpan(ctx, "Engine.SendMessage")
	defer sp.End()
	sp.AddAttributes(
		trace.StringAttribute("message.id", msg.ID),
		trace.StringAttribute("message.type", msg.Type.String()),
		trace.StringAttribute("dest.type", msg.Dest.Type.String()),
		trace.StringAttribute("dest.id", msg.Dest.ID),
	)
	ctx = log.WithField(ctx, "CallbackID", msg.ID)

	if msg.Dest.Type.IsUserCM() {
		ctx = permission.UserSourceContext(ctx, msg.UserID, permission.RoleUser, &permission.SourceInfo{
			Type: permission.SourceTypeContactMethod,
			ID:   msg.Dest.ID,
		})
	} else {
		ctx = permission.SystemContext(ctx, "SendMessage")
		ctx = permission.SourceContext(ctx, &permission.SourceInfo{
			Type: permission.SourceTypeNotificationChannel,
			ID:   msg.Dest.ID,
		})
	}

	var notifMsg notification.Message
	switch msg.Type {
	case notification.MessageTypeAlertBundle:
		name, count, err := p.am.ServiceInfo(ctx, msg.ServiceID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup service info")
		}
		if count == 0 {
			// already acked/closed, don't send bundled notification
			return &notification.SendResult{
				ID: msg.ID,
				Status: notification.Status{
					Details: "alerts acked/closed before message sent",
					State:   notification.StateFailedPerm,
				},
			}, nil
		}
		notifMsg = notification.AlertBundle{
			Dest:        msg.Dest,
			CallbackID:  msg.ID,
			ServiceID:   msg.ServiceID,
			ServiceName: name,
			Count:       count,
		}
	case notification.MessageTypeAlert:
		a, err := p.am.FindOne(ctx, msg.AlertID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup alert")
		}
		stat, err := p.cfg.NotificationStore.OriginalMessageStatus(ctx, msg.AlertID, msg.Dest)
		if err != nil {
			return nil, fmt.Errorf("lookup original message: %w", err)
		}
		if stat != nil && stat.ID == msg.ID {
			// set to nil if it's the current message
			stat = nil
		}
		notifMsg = notification.Alert{
			Dest:       msg.Dest,
			AlertID:    msg.AlertID,
			Summary:    a.Summary,
			Details:    a.Details,
			CallbackID: msg.ID,

			OriginalStatus: stat,
		}
	case notification.MessageTypeAlertStatusBundle:
		e, err := p.cfg.AlertLogStore.FindOne(ctx, msg.AlertLogID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup alert log entry")
		}
		notifMsg = notification.AlertStatusBundle{
			Dest:       msg.Dest,
			CallbackID: msg.ID,
			LogEntry:   e.String(),
			AlertID:    e.AlertID(),
			Count:      len(msg.StatusAlertIDs),
		}
	case notification.MessageTypeAlertStatus:
		e, err := p.cfg.AlertLogStore.FindOne(ctx, msg.AlertLogID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup alert log entry")
		}
		notifMsg = notification.AlertStatus{
			Dest:       msg.Dest,
			AlertID:    e.AlertID(),
			CallbackID: msg.ID,
			LogEntry:   e.String(),
		}
	case notification.MessageTypeTest:
		notifMsg = notification.Test{
			Dest:       msg.Dest,
			CallbackID: msg.ID,
		}
	case notification.MessageTypeVerification:
		code, err := p.cfg.NotificationStore.Code(ctx, msg.VerifyID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup verification code")
		}
		notifMsg = notification.Verification{
			Dest:       msg.Dest,
			CallbackID: msg.ID,
			Code:       code,
		}
	case notification.MessageTypeScheduleOnCallStatus:
		users, err := p.cfg.OnCallStore.OnCallUsersBySchedule(ctx, msg.ScheduleID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup on call users by schedule")
		}
		sched, err := p.cfg.ScheduleStore.FindOne(ctx, msg.ScheduleID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup schedule by id")
		}

		var onCallUsers []notification.User
		for _, u := range users {
			onCallUsers = append(onCallUsers, notification.User{
				Name: u.Name,
				ID:   u.ID,
				URL:  p.cfg.ConfigSource.Config().CallbackURL("/users/" + u.ID),
			})
		}

		notifMsg = notification.ScheduleOnCallStatus{
			Dest:       msg.Dest,
			CallbackID: msg.ID,
			Schedule: notification.Schedule{
				Name: sched.Name,
				URL:  p.cfg.ConfigSource.Config().CallbackURL("/schedules/" + msg.ScheduleID),
				ID:   msg.ScheduleID,
			},
			Users: onCallUsers,
		}
	default:
		log.Log(ctx, errors.New("SEND NOT IMPLEMENTED FOR MESSAGE TYPE"))
		return &notification.SendResult{ID: msg.ID, Status: notification.Status{State: notification.StateFailedPerm}}, nil
	}

	meta := alertlog.NotificationMetaData{
		MessageID: msg.ID,
	}

	res, err := p.cfg.NotificationManager.SendMessage(ctx, notifMsg)
	if err != nil {
		return nil, err
	}

	switch msg.Type {
	case notification.MessageTypeAlert:
		p.cfg.AlertLogStore.MustLog(ctx, msg.AlertID, alertlog.TypeNotificationSent, meta)
	case notification.MessageTypeAlertBundle:
		err = p.cfg.AlertLogStore.LogServiceTx(ctx, nil, msg.ServiceID, alertlog.TypeNotificationSent, meta)
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "append alert log"))
		}
	}

	return res, nil
}
