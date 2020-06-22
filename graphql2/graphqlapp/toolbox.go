package graphqlapp

import (
	context "context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"github.com/ttacon/libphonenumber"
)

const twLookupURL = "https://lookups.twilio.com/v1/PhoneNumbers/"

type DebugPhoneNumberInfo App

func (a *App) DebugPhoneNumberInfo() graphql2.DebugPhoneNumberInfoResolver {
	return (*DebugPhoneNumberInfo)(a)
}

func (a *DebugPhoneNumberInfo) Carrier(ctx context.Context, info *graphql2.DebugPhoneNumberInfo) (*graphql2.DebugPhoneNumberCarrierInfo, error) {
	// must be admin to fetch carrier info
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	cfg := config.FromContext(ctx)
	url := twLookupURL + info.ID + "?Type=carrier"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(cfg.Twilio.AccountSID, cfg.Twilio.AuthToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response from Twilio: %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Carrier struct {
			CC   string `json:"mobile_country_code"`
			NC   string `json:"mobile_network_code"`
			Name string
			Type string
		}
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &graphql2.DebugPhoneNumberCarrierInfo{
		Name:              result.Carrier.Name,
		Type:              result.Carrier.Type,
		MobileCountryCode: result.Carrier.CC,
		MobileNetworkCode: result.Carrier.NC,
	}, nil
}

func (a *Mutation) DebugPhoneNumberInfo(ctx context.Context, input graphql2.DebugPhoneNumberInfoInput) (*graphql2.DebugPhoneNumberInfo, error) {
	err := validate.Phone("Number", input.Number)
	if err != nil {
		return nil, err
	}

	p, err := libphonenumber.Parse(input.Number, "")
	if err != nil {
		return nil, validation.NewFieldError("Number", fmt.Sprintf("must be a valid number: %s", err.Error()))
	}

	info := &graphql2.DebugPhoneNumberInfo{
		ID:          input.Number,
		CountryCode: fmt.Sprintf("+%d", p.GetCountryCode()),
		RegionCode:  libphonenumber.GetRegionCodeForNumber(p),
		Formatted:   libphonenumber.Format(p, libphonenumber.INTERNATIONAL),
	}

	return info, nil
}
