package engine

import (
	"context"

	"github.com/pkg/errors"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/engine/message"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"go.opencensus.io/trace"
)

func (p *Engine) sendMessage(ctx context.Context, msg *message.Message) (*notification.MessageStatus, error) {
	ctx, sp := trace.StartSpan(ctx, "Engine.SendMessage")
	defer sp.End()
	sp.AddAttributes(
		trace.StringAttribute("message.id", msg.ID),
		trace.StringAttribute("message.type", string(msg.Type)),
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
	case message.TypeAlertNotificationBundle:
		name, count, err := p.am.ServiceInfo(ctx, msg.ServiceID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup service info")
		}
		if count == 0 {
			// already acked/closed, don't send bundled notification
			return &notification.MessageStatus{
				Ctx:     ctx,
				ID:      msg.ID,
				Details: "alerts acked/closed before message sent",
				State:   notification.MessageStateFailedPerm,
			}, nil
		}
		notifMsg = notification.AlertBundle{
			Dest:        msg.Dest,
			CallbackID:  msg.ID,
			ServiceID:   msg.ServiceID,
			ServiceName: name,
			Count:       count,
		}
	case message.TypeAlertNotification:
		a, err := p.am.FindOne(ctx, msg.AlertID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup alert")
		}
		notifMsg = notification.Alert{
			Dest:       msg.Dest,
			AlertID:    msg.AlertID,
			Summary:    a.Summary,
			Details:    a.Details,
			CallbackID: msg.ID,
		}
	case message.TypeAlertStatusUpdateBundle:
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
	case message.TypeAlertStatusUpdate:
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
	case message.TypeTestNotification:
		notifMsg = notification.Test{
			Dest:       msg.Dest,
			CallbackID: msg.ID,
		}
	case message.TypeVerificationMessage:
		code, err := p.cfg.NotificationStore.Code(ctx, msg.VerifyID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup verification code")
		}
		notifMsg = notification.Verification{
			Dest:       msg.Dest,
			CallbackID: msg.ID,
			Code:       code,
		}
	default:
		log.Log(ctx, errors.New("SEND NOT IMPLEMENTED FOR MESSAGE TYPE"))
		return &notification.MessageStatus{State: notification.MessageStateFailedPerm}, nil
	}

	meta := alertlog.NotificationMetaData{
		MessageID: msg.ID,
	}

	status, err := p.cfg.NotificationSender.Send(ctx, notifMsg)
	if err != nil {
		return nil, err
	}

	switch msg.Type {
	case message.TypeAlertNotification:
		p.cfg.AlertLogStore.MustLog(ctx, msg.AlertID, alertlog.TypeNotificationSent, meta)
	case message.TypeAlertNotificationBundle:
		err = p.cfg.AlertLogStore.LogServiceTx(ctx, nil, msg.ServiceID, alertlog.TypeNotificationSent, meta)
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "append alert log"))
		}
	}

	return status, nil
}
