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
		ctx = permission.SourceContext(ctx, &permission.SourceInfo{
			Type: permission.SourceTypeNotificationChannel,
			ID:   msg.Dest.ID,
		})
	}

	var notifMsg notification.Message
	switch msg.Type {
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
	case message.TypeAlertStatusUpdate:
		e, err := p.cfg.AlertLogStore.FindOne(ctx, msg.AlertLogID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup alert log entry")
		}
		notifMsg = notification.AlertStatus{
			Dest:      msg.Dest,
			AlertID:   msg.AlertID,
			MessageID: msg.ID,
			Log:       e.String(),
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

	status, err := p.cfg.NotificationSender.Send(ctx, notifMsg)
	if err != nil {
		return nil, err
	}
	if msg.Type == message.TypeAlertNotification {
		p.cfg.AlertLogStore.MustLog(ctx, msg.AlertID, alertlog.TypeNotificationSent, nil)
	}

	return status, nil
}
