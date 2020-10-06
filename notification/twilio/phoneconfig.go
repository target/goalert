package twilio

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/target/goalert/config"
)

type PhoneNumberConfig struct {
	SID string

	Capabilities struct {
		SMS   bool
		Voice bool
	}

	SMSMethod string `json:"sms_method"`
	SMSURL    string `json:"sms_url"`

	VoiceMethod string `json:"voice_method"`
	VoiceURL    string `json:"voice_url"`
}

// PhoneNumberConfig will return the configuration of the provided number.
// If there is no matching number on the account, nil is returned.
func (c *Config) PhoneNumberConfig(ctx context.Context, number string) (*PhoneNumberConfig, error) {
	cfg := config.FromContext(ctx)
	urlStr := c.url("Accounts", cfg.Twilio.AccountSID, "IncomingPhoneNumbers.json") + "?PhoneNumber=" + url.QueryEscape(number)
	fmt.Println("GET", urlStr)
	resp, err := c.get(ctx, urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response: %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result struct {
		Numbers []PhoneNumberConfig `json:"incoming_phone_numbers"`
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("parse response: %w\n%s", err, string(data))
	}
	if len(result.Numbers) == 0 {
		return nil, nil
	}
	if len(result.Numbers) > 1 {
		return nil, fmt.Errorf("expected 1 number to match, got %d", len(result.Numbers))
	}

	return &result.Numbers[0], nil
}
