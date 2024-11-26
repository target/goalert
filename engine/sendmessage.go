package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/engine/message"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
)

func (p *Engine) sendMessage(ctx context.Context, msg *message.Message) (*notification.SendResult, error) {
	ctx = log.WithField(ctx, "CallbackID", msg.ID)

	if msg.DestID.IsUserCM() {
		ctx = permission.UserSourceContext(ctx, msg.UserID, permission.RoleUser, &permission.SourceInfo{
			Type: permission.SourceTypeContactMethod,
			ID:   msg.DestID.String(),
		})
	} else {
		ctx = permission.SystemContext(ctx, "SendMessage")
		ctx = permission.SourceContext(ctx, &permission.SourceInfo{
			Type: permission.SourceTypeNotificationChannel,
			ID:   msg.DestID.String(),
		})
	}

	var notifMsg notification.Message
	var isFirstAlertMessage bool
	switch msg.Type {
	case notification.MessageTypeAlertBundle:
		name, count, err := p.a.ServiceInfo(ctx, msg.ServiceID)
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
			Base:        msg.Base(),
			ServiceID:   msg.ServiceID,
			ServiceName: name,
			Count:       count,
		}
	case notification.MessageTypeAlert:
		name, _, err := p.a.ServiceInfo(ctx, msg.ServiceID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup service info")
		}
		a, err := p.a.FindOne(ctx, msg.AlertID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup alert")
		}
		stat, err := p.cfg.NotificationStore.OriginalMessageStatus(ctx, msg.AlertID, msg.DestID)
		if err != nil {
			return nil, fmt.Errorf("lookup original message: %w", err)
		}
		if stat != nil && stat.ID == msg.ID {
			// set to nil if it's the current message
			stat = nil
		}
		meta, err := p.a.Metadata(ctx, gadb.Compat(p.b.db), msg.AlertID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup alert metadata")
		}
		notifMsg = notification.Alert{
			Base:        msg.Base(),
			AlertID:     msg.AlertID,
			Summary:     a.Summary,
			Details:     a.Details,
			ServiceID:   a.ServiceID,
			ServiceName: name,
			Meta:        meta,

			OriginalStatus: stat,
		}
		isFirstAlertMessage = stat == nil
	case notification.MessageTypeAlertStatus:
		e, err := p.cfg.AlertLogStore.FindOne(ctx, msg.AlertLogID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup alert log entry")
		}
		a, err := p.cfg.AlertStore.FindOne(ctx, msg.AlertID)
		if err != nil {
			return nil, fmt.Errorf("lookup original alert: %w", err)
		}
		stat, err := p.cfg.NotificationStore.OriginalMessageStatus(ctx, msg.AlertID, msg.DestID)
		if err != nil {
			return nil, fmt.Errorf("lookup original message: %w", err)
		}
		if stat == nil {
			return nil, fmt.Errorf("could not find original notification for alert %d to %s", msg.AlertID, msg.Dest.String())
		}

		var status notification.AlertState
		switch e.Type() {
		case alertlog.TypeAcknowledged:
			status = notification.AlertStateAcknowledged
		case alertlog.TypeEscalated:
			status = notification.AlertStateUnacknowledged
		case alertlog.TypeClosed:
			status = notification.AlertStateClosed
		}

		notifMsg = notification.AlertStatus{
			Base:           msg.Base(),
			AlertID:        e.AlertID(),
			ServiceID:      a.ServiceID,
			LogEntry:       e.String(ctx),
			Summary:        a.Summary,
			Details:        a.Details,
			NewAlertState:  status,
			OriginalStatus: *stat,
		}
	case notification.MessageTypeTest:
		notifMsg = notification.Test{
			Base: msg.Base(),
		}
	case notification.MessageTypeVerification:
		code, err := p.cfg.NotificationStore.Code(ctx, msg.VerifyID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup verification code")
		}
		notifMsg = notification.Verification{
			Base: msg.Base(),
			Code: fmt.Sprintf("%06d", code),
		}
	case notification.MessageTypeScheduleOnCallUsers:
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

		notifMsg = notification.ScheduleOnCallUsers{
			Base:         msg.Base(),
			ScheduleName: sched.Name,
			ScheduleURL:  p.cfg.ConfigSource.Config().CallbackURL("/schedules/" + msg.ScheduleID),
			ScheduleID:   msg.ScheduleID,
			Users:        onCallUsers,
		}
	case notification.MessageTypeSignalMessage:
		id, err := uuid.Parse(msg.ID)
		if err != nil {
			return nil, errors.Wrap(err, "parse signal message id")
		}
		rawParams, err := gadb.NewCompat(p.b.db).EngineGetSignalParams(ctx, uuid.NullUUID{Valid: true, UUID: id})
		if err != nil {
			return nil, errors.Wrap(err, "get signal message params")
		}

		var params map[string]string
		err = json.Unmarshal(rawParams, &params)
		if err != nil {
			return nil, errors.Wrap(err, "parse signal message params")
		}

		notifMsg = notification.SignalMessage{
			Base:   msg.Base(),
			Params: params,
		}
	default:
		log.Log(ctx, errors.New("SEND NOT IMPLEMENTED FOR MESSAGE TYPE "+string(msg.Type)))
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

	if isFirstAlertMessage && res.State.IsOK() {
		_, err = p.b.trackStatus.ExecContext(ctx, msg.DestID.NCID, msg.DestID.CMID, msg.AlertID)
		if err != nil {
			// non-fatal, but log because it means status updates will not work for that alert/dest.
			log.Log(ctx, fmt.Errorf("track status updates for alert #%d for %s: %w", msg.AlertID, msg.Dest.String(), err))
		}
	}

	return res, nil
}
