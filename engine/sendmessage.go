package engine

import (
	"context"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

func (p *Engine) sendMessage(ctx context.Context, msgID string, destType notification.DestType, destID string, disabledOK bool, fn func(notification.Dest) notification.Message, afterFn func(context.Context)) (*notification.MessageStatus, error) {
	ctx, sp := trace.StartSpan(ctx, "Engine.SendMessage")
	defer sp.End()
	sp.AddAttributes(
		trace.StringAttribute("message.id", msgID),
		trace.StringAttribute("dest.type", destType.String()),
		trace.StringAttribute("dest.id", destID),
	)

	var dest notification.Dest
	if destType.IsUserCM() {
		cm, err := p.cfg.ContactMethodStore.FindOne(ctx, destID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup contact method")
		}
		trace.FromContext(ctx).AddAttributes(trace.StringAttribute("message.contactMethod.value", cm.Value))

		if !disabledOK && cm.Disabled {
			return nil, errDisabledCM
		}
		dest.Type = cm.Type.DestType()
		dest.Value = cm.Value
		ctx = permission.UserSourceContext(ctx, cm.UserID, permission.RoleUser, &permission.SourceInfo{
			Type: permission.SourceTypeContactMethod,
			ID:   cm.ID,
		})
	} else {
		ch, err := p.cfg.NCStore.FindOne(ctx, destID)
		if err != nil {
			return nil, errors.Wrap(err, "lookup notification channel")
		}
		dest.Type = ch.Type.DestType()
		dest.Value = ch.Value
		ctx = permission.SourceContext(ctx, &permission.SourceInfo{
			Type: permission.SourceTypeNotificationChannel,
			ID:   ch.ID,
		})
	}

	msg := fn(dest)

	ctx = log.WithField(ctx, "CallbackID", msgID)

	status, err := p.cfg.NotificationSender.Send(ctx, msg)
	if err != nil {
		return nil, err
	}
	if afterFn != nil {
		afterFn(ctx)
	}
	return status, nil
}

func (p *Engine) sendStatusUpdate(ctx context.Context, msgID string, alertLogID int, destType notification.DestType, destID string) (*notification.MessageStatus, error) {
	e, err := p.cfg.AlertlogStore.FindOne(ctx, alertLogID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup alert log entry")
	}

	return p.sendMessage(ctx, msgID, destType, destID, false, func(dest notification.Dest) notification.Message {
		return notification.AlertStatus{
			Dest:      dest,
			AlertID:   e.AlertID(),
			MessageID: msgID,
			Log:       e.String(),
		}
	}, nil)
}
func (p *Engine) sendNotification(ctx context.Context, msgID string, alertID int, destType notification.DestType, destID string) (*notification.MessageStatus, error) {
	a, err := p.am.FindOne(ctx, alertID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup alert")
	}
	return p.sendMessage(ctx, msgID, destType, destID, false, func(dest notification.Dest) notification.Message {
		return notification.Alert{
			Dest:       dest,
			AlertID:    a.ID,
			Summary:    a.Summary,
			Details:    a.Details,
			CallbackID: msgID,
		}
	}, func(ctx context.Context) {
		p.cfg.AlertlogStore.MustLog(ctx, alertID, alertlog.TypeNotificationSent, nil)
	})
}
func (p *Engine) sendTestNotification(ctx context.Context, msgID string, destType notification.DestType, destID string) (*notification.MessageStatus, error) {
	return p.sendMessage(ctx, msgID, destType, destID, false, func(dest notification.Dest) notification.Message {
		return notification.Test{
			Dest:       dest,
			CallbackID: msgID,
		}
	}, nil)
}
func (p *Engine) sendVerificationMessage(ctx context.Context, msgID string, destType notification.DestType, destID, verifyID string) (*notification.MessageStatus, error) {
	code, err := p.cfg.NotificationStore.Code(ctx, verifyID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup verification code")
	}
	return p.sendMessage(ctx, msgID, destType, destID, true, func(dest notification.Dest) notification.Message {
		return notification.Verification{
			Dest:       dest,
			CallbackID: msgID,
			Code:       code,
		}
	}, nil)
}
