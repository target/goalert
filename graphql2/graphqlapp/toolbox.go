package graphqlapp

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nyaruka/phonenumbers"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/validation"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation/validate"
)

type safeErr struct{ error }

func (safeErr) ClientError() bool { return true }

func (q *Query) DebugMessageStatus(ctx context.Context, input graphql2.DebugMessageStatusInput) (*graphql2.DebugMessageStatusInfo, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	id, err := notification.ParseProviderMessageID(input.ProviderMessageID)
	if err != nil {
		return nil, validation.NewFieldError("ProviderMessageID", err.Error())
	}

	status, destType, err := q.NotificationManager.MessageStatus(ctx, id)
	if err != nil {
		return nil, err
	}

	return &graphql2.DebugMessageStatusInfo{
		State: notificationStateFromSendResult(*status, q.FormatDestFunc(ctx, destType, status.SrcValue)),
	}, nil
}

func (a *Mutation) DebugSendSms(ctx context.Context, input graphql2.DebugSendSMSInput) (*graphql2.DebugSendSMSInfo, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	err = validate.Many(
		validate.Phone("To", input.To),
		validate.TwilioFromValue("From", input.From),
		validate.Text("Body", input.Body, 1, 1000),
	)
	if err != nil {
		return nil, err
	}

	msg, err := a.Twilio.SendSMS(ctx, input.To, input.Body, &twilio.SMSOptions{
		FromNumber: input.From,
	})
	if err != nil {
		return nil, safeErr{error: err}
	}

	return &graphql2.DebugSendSMSInfo{
		ID: notification.ProviderMessageID{
			ExternalID:   msg.SID,
			ProviderName: "Twilio-SMS",
		}.String(),
		ProviderURL: "https://www.twilio.com/console/sms/logs/" + url.PathEscape(msg.SID),
		FromNumber:  msg.From,
	}, nil
}

func (a *Mutation) DebugCarrierInfo(ctx context.Context, input graphql2.DebugCarrierInfoInput) (*twilio.CarrierInfo, error) {
	return a.Twilio.FetchCarrierInfo(ctx, input.Number)
}

func (a *Query) PhoneNumberInfo(ctx context.Context, number string) (*graphql2.PhoneNumberInfo, error) {
	p, err := phonenumbers.Parse(number, "")
	if err != nil {
		return &graphql2.PhoneNumberInfo{
			ID:    number,
			Error: err.Error(),
		}, nil
	}

	return &graphql2.PhoneNumberInfo{
		ID:          number,
		CountryCode: fmt.Sprintf("+%d", p.GetCountryCode()),
		RegionCode:  phonenumbers.GetRegionCodeForNumber(p),
		Formatted:   phonenumbers.Format(p, phonenumbers.INTERNATIONAL),
		Valid:       phonenumbers.IsValidNumber(p),
	}, nil
}
