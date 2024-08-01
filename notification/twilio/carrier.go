package twilio

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util/log"
)

// CarrierInfo holds information about the carrier network for a particular number.
type CarrierInfo struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	MobileNetworkCode string `json:"mobile_network_code"`
	MobileCountryCode string `json:"mobile_country_code"`
}

// DefaultLookupURL is the value that will be used for lookup calls if Config.BaseURL is empty.
const DefaultLookupURL = "https://lookups.twilio.com"

// ErrCarrierStale is returned if the available carrier data is over a year old.
var ErrCarrierStale = errors.New("carrier data is stale")

// ErrCarrierUnavailable is returned if the carrier data is missing and fetch is disabled.
var ErrCarrierUnavailable = errors.New("carrier data is unavailable")

func (c *Config) dbCarrierInfo(ctx context.Context, number string) (*CarrierInfo, error) {
	m, err := c.CMStore.MetadataByDest(ctx, c.DB, NewSMSDest(number))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCarrierUnavailable
	}
	if err != nil {
		return nil, err
	}
	info := &CarrierInfo{
		Name:              m.CarrierV1.Name,
		Type:              m.CarrierV1.Type,
		MobileCountryCode: m.CarrierV1.MobileCountryCode,
		MobileNetworkCode: m.CarrierV1.MobileNetworkCode,
	}
	if m.FetchedAt.Sub(m.CarrierV1.UpdatedAt) > 365*24*time.Hour {
		// over a year old
		return info, ErrCarrierStale
	}

	return info, nil
}

// CarrierInfo will return carrier information for the provided number. If fetch is true, it will fetch
// data from the Twilio API if it is not available from the DB.
func (c *Config) CarrierInfo(ctx context.Context, number string, fetch bool) (*CarrierInfo, error) {
	if c.CMStore == nil {
		return nil, nil
	}
	info, err := c.dbCarrierInfo(ctx, number)
	if !fetch || err == nil {
		return info, err
	}

	if !errors.Is(err, ErrCarrierStale) && !errors.Is(err, ErrCarrierUnavailable) {
		log.Log(ctx, err)
	}

	return c.FetchCarrierInfo(ctx, number)
}

// FetchCarrierInfo will lookup carrier information for the provided number using the Twilio API.
func (c Config) FetchCarrierInfo(ctx context.Context, number string) (*CarrierInfo, error) {
	if c.CMStore == nil {
		return nil, nil
	}
	// must be admin to fetch carrier info
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	cfg := config.FromContext(ctx)
	if c.BaseURL == "" {
		c.BaseURL = DefaultLookupURL
	}
	c.BaseURL = strings.TrimSuffix(c.BaseURL, "/")

	url := c.BaseURL + "/v1/PhoneNumbers/" + url.PathEscape(number) + "?Type=carrier"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(cfg.Twilio.AccountSID, cfg.Twilio.AuthToken)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response from Twilio: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Carrier CarrierInfo
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	// merge into existing metadata (if possible)
	var m contactmethod.Metadata
	m.CarrierV1.Name = result.Carrier.Name
	m.CarrierV1.Type = result.Carrier.Type
	m.CarrierV1.MobileCountryCode = result.Carrier.MobileCountryCode
	m.CarrierV1.MobileNetworkCode = result.Carrier.MobileNetworkCode

	err = c.CMStore.SetCarrierV1MetadataByDest(ctx, c.DB, NewSMSDest(number), &m)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Log(ctx, err)
	}
	err = c.CMStore.SetCarrierV1MetadataByDest(ctx, c.DB, NewVoiceDest(number), &m)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Log(ctx, err)
	}

	return &result.Carrier, nil
}
