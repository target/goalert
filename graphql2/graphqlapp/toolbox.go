package graphqlapp

import (
	context "context"
	"fmt"
	"net/url"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation/validate"
	"github.com/ttacon/libphonenumber"
)

const twLookupURL = "https://lookups.twilio.com/v1/PhoneNumbers/"

type safeErr struct{ error }

func (safeErr) ClientError() bool { return true }

func (a *Mutation) DebugSendSms(ctx context.Context, input graphql2.DebugSendSMSInput) (*graphql2.DebugSendSMSInfo, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	err = validate.Many(
		validate.Phone("To", input.To),
		validate.Phone("From", input.From),
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
		ID:          msg.SID,
		ProviderURL: "https://www.twilio.com/console/sms/logs/" + url.PathEscape(msg.SID),
	}, nil
}

func (a *Mutation) DebugCarrierInfo(ctx context.Context, input graphql2.DebugCarrierInfoInput) (*twilio.CarrierInfo, error) {
	return a.Twilio.FetchCarrierInfo(ctx, input.Number)
}

func (a *Query) PhoneNumberInfo(ctx context.Context, number string) (*graphql2.PhoneNumberInfo, error) {
	p, err := libphonenumber.Parse(number, "")
	if err != nil {
		return &graphql2.PhoneNumberInfo{
			ID:    number,
			Error: err.Error(),
		}, nil
	}

	return &graphql2.PhoneNumberInfo{
		ID:          number,
		CountryCode: fmt.Sprintf("+%d", p.GetCountryCode()),
		RegionCode:  libphonenumber.GetRegionCodeForNumber(p),
		Formatted:   libphonenumber.Format(p, libphonenumber.INTERNATIONAL),
		Valid:       libphonenumber.IsValidNumber(p),
	}, nil
}
