package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// SchemaVersion indicates the current config struct version.
const SchemaVersion = 1

// Config contains GoAlert application settings.
type Config struct {
	data        []byte
	fallbackURL string

	General struct {
		PublicURL                    string `public:"true" info:"Publicly routable URL for UI links and API calls."`
		GoogleAnalyticsID            string `public:"true"`
		NotificationDisclaimer       string `public:"true" info:"Disclaimer text for receiving pre-recorded notifications (appears on profile page)."`
		MessageBundles               bool   `public:"true" info:"Enables bundling status updates and alert notifications. Also allows 'ack/close all' responses to bundled alerts."`
		ShortURL                     string `public:"true" info:"If set, messages will contain a shorter URL using this as a prefix (e.g. http://example.com). It should point to GoAlert and can be the same as the PublicURL."`
		DisableSMSLinks              bool   `public:"true" info:"If set, SMS messages will not contain a URL pointing to GoAlert."`
		DisableLabelCreation         bool   `public:"true" info:"Disables the ability to create new labels for services."`
		DisableCalendarSubscriptions bool   `public:"true" info:"If set, disables all active calendar subscriptions as well as the ability to create new calendar subscriptions."`
		DisableV1GraphQL             bool   `info:"Disables the deprecated /v1/graphql endpoint (replaced by /api/graphql)."`
	}

	Maintenance struct {
		AlertCleanupDays int `public:"true" info:"Closed alerts will be deleted after this many days (0 means disable cleanup)."`
		APIKeyExpireDays int `public:"true" info:"Unused calendar API keys will be disabled after this many days (0 means disable cleanup)."`
	}

	Auth struct {
		RefererURLs  []string `info:"Allowed referer URLs for auth and redirects."`
		DisableBasic bool     `public:"true" info:"Disallow username/password login."`
	}

	GitHub struct {
		Enable bool `public:"true" info:"Enable GitHub authentication."`

		NewUsers bool `info:"Allow new user creation via GitHub authentication."`

		ClientID     string
		ClientSecret string `password:"true"`

		AllowedUsers []string `info:"Allow any of the listed GitHub usernames to authenticate. Use '*' to allow any user."`
		AllowedOrgs  []string `info:"Allow any member of any listed GitHub org (or team, using the format 'org/team') to authenticate."`

		EnterpriseURL string `info:"GitHub URL (without /api) when used with GitHub Enterprise."`
	}

	OIDC struct {
		Enable bool `public:"true" info:"Enable OpenID Connect authentication."`

		NewUsers     bool   `info:"Allow new user creation via OIDC authentication."`
		OverrideName string `info:"Set the name/label on the login page to something other than OIDC."`

		IssuerURL    string
		ClientID     string
		ClientSecret string `password:"true"`

		Scopes                    string `info:"Requested scopes for authentication. If left blank, openid, profile, and email will be used."`
		UserInfoEmailPath         string `info:"JMESPath expression to find email address in UserInfo. If set, the email claim will be ignored in favor of this. (suggestion: email)."`
		UserInfoEmailVerifiedPath string `info:"JMESPath expression to find email verification state in UserInfo. If set, the email_verified claim will be ignored in favor of this. (suggestion: email_verified)."`
		UserInfoNamePath          string `info:"JMESPath expression to find full name in UserInfo. If set, the name claim will be ignored in favor of this. (suggestion: name || cn || join(' ', [firstname, lastname]))"`
	}

	Mailgun struct {
		Enable bool `public:"true"`

		APIKey      string `password:"true"`
		EmailDomain string `info:"The TO address for all incoming alerts."`
	}

	Slack struct {
		Enable bool `public:"true"`

		ClientID     string
		ClientSecret string `password:"true"`

		// The `xoxb-` prefix is documented by Slack.
		// https://api.slack.com/docs/token-types#bot
		AccessToken string `password:"true" info:"Slack app bot user OAuth access token (should start with xoxb-)."`
	}

	Twilio struct {
		Enable bool `public:"true" info:"Enables sending and processing of Voice and SMS messages through the Twilio notification provider."`

		AccountSID string
		AuthToken  string `password:"true" info:"The primary Auth Token for Twilio. Must be primary (not secondary) for request valiation."`
		FromNumber string `public:"true" info:"The Twilio number to use for outgoing notifications."`

		DisableTwoWaySMS      bool     `info:"Disables SMS reply codes for alert messages."`
		SMSCarrierLookup      bool     `info:"Perform carrier lookup of SMS contact methods (required for SMSFromNumberOverride). Extra charges may apply."`
		SMSFromNumberOverride []string `info:"List of 'carrier=number' pairs, SMS messages to numbers of the provided carrier string (exact match) will use the alternate From Number."`
	}

	SMTP struct {
		Enable bool `public:"true" info:"Enables email as a contact method."`

		From string `public:"true" info:"The email address messages should be sent from."`

		Address    string `info:"The server address to use for sending email. Port is optional."`
		DisableTLS bool   `info:"Disables TLS on the connection (STARTTLS will still be used if supported)."`
		SkipVerify bool   `info:"Disables certificate validation for TLS/STARTTLS (insecure)."`

		Username string `info:"Username for authentication."`
		Password string `password:"true" info:"Password for authentication."`
	}

	Webhook struct {
		Enable      bool     `public:"true" info:"Enables webhook as a contact method."`
		AllowedURLs []string `public:"true" info:"If set, allows webhooks for these domains only."`
	}

	Feedback struct {
		Enable      bool   `public:"true" info:"Enables Feedback link in nav bar."`
		OverrideURL string `public:"true" info:"Use a custom URL for Feedback link in nav bar."`
	}
}

// TwilioSMSFromNumber will determine the appropriate FROM number to use for SMS messages to the given number
func (cfg Config) TwilioSMSFromNumber(carrier string) string {
	if carrier == "" {
		return cfg.Twilio.FromNumber
	}

	for _, s := range cfg.Twilio.SMSFromNumberOverride {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] != carrier {
			continue
		}
		return parts[1]
	}

	return cfg.Twilio.FromNumber
}

func (cfg Config) rawCallbackURL(path string, mergeParams ...url.Values) *url.URL {
	base, err := url.Parse(cfg.PublicURL())
	if err != nil {
		panic(errors.Wrap(err, "parse PublicURL"))
	}

	next, err := url.Parse(path)
	if err != nil {
		panic(errors.Wrap(err, "parse path"))
	}

	base.Path = strings.TrimSuffix(base.Path, "/") + "/" + strings.TrimPrefix(next.Path, "/")

	params := base.Query()
	nx := next.Query()
	// set/override any params provided with path
	for name, val := range nx {
		params[name] = val
	}

	// set/override with any additionally provided params
	for _, merge := range mergeParams {
		for name, val := range merge {
			params[name] = val
		}
	}

	base.RawQuery = params.Encode()
	return base
}

// CallbackURL will return a public-routable URL to the given path.
// It will use PublicURL() to fill in missing pieces.
//
// It will panic if provided an invalid URL.
func (cfg Config) CallbackURL(path string, mergeParams ...url.Values) string {
	base := cfg.rawCallbackURL(path, mergeParams...)

	newPath := ShortPath(base.Path)
	if newPath != "" && cfg.General.ShortURL != "" {
		short, err := url.Parse(cfg.General.ShortURL)
		if err != nil {
			panic(errors.Wrap(err, "parse ShortURL"))
		}
		base.Path = newPath
		base.Host = short.Host
		base.Scheme = short.Scheme
	}

	return base.String()
}

//  MatchURL will compare two url strings and will return true if they match.
func MatchURL(baseURL, testURL string) (bool, error) {
	compareQueryStrings := func (queryStr1, queryStr2 url.Values) bool {
		for query, param := range queryStr1 {
			if len(param[0]) > 0 {
				val, ok := queryStr2[query]
				if !ok || param[0] != val[0] {
					return false
				}
			}
		}
		return true
	}

	ignoreImplicitPorts := func(urlString *url.URL) bool {
		if strings.EqualFold(urlString.Scheme, "http") && urlString.Port() == "80" {
			return true
		}
		if strings.EqualFold(urlString.Scheme, "https") && urlString.Port() == "443" {
			return true
		}
		return false
	}
	base, err := url.Parse(baseURL)

	if err != nil {
		return false, err
	}
	test, err := url.Parse(testURL)

	if err != nil {
		return false, err
	}

	// ignore auth info
	test.User = nil
	base.User = nil

	// use full host that includes port by default, i.e example.com:8080
	baseHost := base.Host
	testHost := test.Host

	// if ignoreImplicitPorts, use host ignoring port, i.e. example.com
	if ignoreImplicitPorts(base) {
		baseHost = base.Hostname()
	}
	if ignoreImplicitPorts(test) {
		testHost = test.Hostname()
	}

	// scheme check
	if !strings.EqualFold(base.Scheme, test.Scheme) {
		return false, nil
	}

	// host/port check
	if !strings.EqualFold(baseHost, testHost) {
		return false, nil
	}

	// path check
	if len(base.Path) > 1 && !strings.HasPrefix(test.Path, base.Path) {
		return false, nil
	}

	// query check
	if len(test.Query()) > 0 || len(base.Query()) > 0 {
		if !compareQueryStrings(test.Query(), base.Query()) {
			return false, nil
		}
		if !compareQueryStrings(base.Query(), test.Query()) {
			return false, nil
		}
	}

	return true, nil
}

// ValidWebhookURL returns true if the URL is an allowed webhook source.
func (cfg Config) ValidWebhookURL(whURL string) bool {
	if len(cfg.Webhook.AllowedURLs) == 0 {
		return true
	}
	for _, u := range cfg.Webhook.AllowedURLs {
		matched, err := MatchURL(whURL, u)
		if err != nil {
			return false
		}
		return matched
	}
	return false
}

// ValidReferer returns true if the URL is an allowed referer source.
func (cfg Config) ValidReferer(reqURL, ref string) bool {
	pubURL := cfg.PublicURL()
	if pubURL != "" && strings.HasPrefix(ref, pubURL) {
		return true
	}

	if len(cfg.Auth.RefererURLs) == 0 {
		u, err := url.Parse(reqURL)
		if err != nil {
			return false
		}
		// just ensure ref is same host/scheme as req
		u.Path = ""
		u.RawQuery = ""
		matched, err := MatchURL(u.String(), ref)
		if err != nil {
			return false
		}
		return matched
	}

	for _, u := range cfg.Auth.RefererURLs {
		matched, err := MatchURL(u, ref)
		if err != nil {
			return false
		}
		return matched
	}

	return false
}

// PublicURL will return the General.PublicURL or a fallback address (i.e. the app listening port).
func (cfg Config) PublicURL() string {
	if cfg.General.PublicURL == "" {
		return strings.TrimSuffix(cfg.fallbackURL, "/")
	}

	return strings.TrimSuffix(cfg.General.PublicURL, "/")
}

func validateEnable(prefix string, isEnabled bool, vals ...string) error {
	if !isEnabled {
		return nil
	}

	var err error
	for i := 0; i < len(vals); i += 2 {
		if vals[i+1] != "" {
			continue
		}
		err = validate.Many(
			err,
			validation.NewFieldError(prefix+".Enable", fmt.Sprintf("requires %s.%s to be set ", prefix, vals[i])),
			validation.NewFieldError(fmt.Sprintf("%s.%s", prefix, vals[i]), "required to enable "+prefix),
		)
	}

	return err
}

// Validate will check that the Config values are valid.
func (cfg Config) Validate() error {
	var err error
	if cfg.General.PublicURL != "" {
		err = validate.Many(
			err,
			validate.AbsoluteURL("General.PublicURL", cfg.General.PublicURL),
		)
	}

	validateKey := func(fname, val string) error { return validate.ASCII(fname, val, 0, 128) }
	validatePath := func(fname, val string) error {
		if val == "" {
			return nil
		}
		return validate.JMESPath(fname, val)
	}
	validateScopes := func(fname, val string) error {
		if val == "" {
			return nil
		}
		return validate.OAuthScope(fname, val, "openid")
	}

	err = validate.Many(
		err,
		validate.Text("General.NotificationDisclaimer", cfg.General.NotificationDisclaimer, 0, 500),
		validateKey("Mailgun.APIKey", cfg.Mailgun.APIKey),
		validateKey("Slack.ClientID", cfg.Slack.ClientID),
		validateKey("Slack.ClientSecret", cfg.Slack.ClientSecret),
		validateKey("Twilio.AccountSID", cfg.Twilio.AccountSID),
		validateKey("Twilio.AuthToken", cfg.Twilio.AuthToken),
		validateKey("GitHub.ClientID", cfg.GitHub.ClientID),
		validateKey("GitHub.ClientSecret", cfg.GitHub.ClientSecret),
		validateKey("Slack.AccessToken", cfg.Slack.AccessToken),
		validate.Range("Maintenance.AlertCleanupDays", cfg.Maintenance.AlertCleanupDays, 0, 9000),
		validate.Range("Maintenance.APIKeyExpireDays", cfg.Maintenance.APIKeyExpireDays, 0, 9000),
		validateScopes("OIDC.Scopes", cfg.OIDC.Scopes),
		validatePath("OIDC.UserInfoEmailPath", cfg.OIDC.UserInfoEmailPath),
		validatePath("OIDC.UserInfoEmailVerifiedPath", cfg.OIDC.UserInfoEmailVerifiedPath),
		validatePath("OIDC.UserInfoNamePath", cfg.OIDC.UserInfoNamePath),
	)

	if cfg.OIDC.IssuerURL != "" {
		err = validate.Many(err, validate.AbsoluteURL("OIDC.IssuerURL", cfg.OIDC.IssuerURL))
	}
	if cfg.OIDC.Scopes != "" {
		err = validate.Many(err, validateScopes("OIDC.Scopes", cfg.OIDC.Scopes))
	}
	if cfg.GitHub.EnterpriseURL != "" {
		err = validate.Many(err, validate.AbsoluteURL("GitHub.EnterpriseURL", cfg.GitHub.EnterpriseURL))
	}
	if cfg.Twilio.FromNumber != "" {
		err = validate.Many(err, validate.Phone("Twilio.FromNumber", cfg.Twilio.FromNumber))
	}
	if cfg.Mailgun.EmailDomain != "" {
		err = validate.Many(err, validate.Email("Mailgun.EmailDomain", "example@"+cfg.Mailgun.EmailDomain))
	}
	if cfg.SMTP.From != "" {
		err = validate.Many(err, validate.Email("SMTP.From", cfg.SMTP.From))
	}

	err = validate.Many(
		err,

		validateEnable("Mailgun", cfg.Mailgun.Enable,
			"APIKey", cfg.Mailgun.APIKey,
			"EmailDomain", cfg.Mailgun.EmailDomain,
		),

		validateEnable("Slack", cfg.Slack.Enable,
			"ClientID", cfg.Slack.ClientID,
			"ClientSecret", cfg.Slack.ClientSecret,
		),

		validateEnable("Twilio", cfg.Twilio.Enable,
			"AccountSID", cfg.Twilio.AccountSID,
			"AuthToken", cfg.Twilio.AuthToken,
			"FromNumber", cfg.Twilio.FromNumber,
		),

		validateEnable("GitHub", cfg.GitHub.Enable,
			"ClientID", cfg.GitHub.ClientID,
			"ClientSecret", cfg.GitHub.ClientSecret,
		),

		validateEnable("OIDC", cfg.OIDC.Enable,
			"IssuerURL", cfg.OIDC.IssuerURL,
			"ClientID", cfg.OIDC.ClientID,
			"ClientSecret", cfg.OIDC.ClientSecret,
		),
		validateEnable("SMTP", cfg.SMTP.Enable,
			"From", cfg.SMTP.From,
			"Address", cfg.SMTP.Address,
		),
	)

	if cfg.Feedback.OverrideURL != "" {
		err = validate.Many(
			err,
			validate.AbsoluteURL("Feedback.OverrideURL", cfg.Feedback.OverrideURL),
		)
	}

	if cfg.General.ShortURL != "" {
		err = validate.Many(
			err,
			validate.AbsoluteURL("General.ShortURL", cfg.General.ShortURL),
		)
	}

	for i, urlStr := range cfg.Auth.RefererURLs {
		field := fmt.Sprintf("Auth.RefererURLs[%d]", i)
		err = validate.Many(
			err,
			validate.AbsoluteURL(field, urlStr),
		)
	}

	for i, urlStr := range cfg.Webhook.AllowedURLs {
		field := fmt.Sprintf("Webhook.AllowedURLs[%d]", i)
		err = validate.Many(err, validate.AbsoluteURL(field, urlStr))
	}

	m := make(map[string]bool)
	for i, str := range cfg.Twilio.SMSFromNumberOverride {
		parts := strings.SplitN(str, "=", 2)
		fname := fmt.Sprintf("Twilio.SMSFromNumberOverride[%d]", i)
		if len(parts) != 2 {
			err = validate.Many(err, validation.NewFieldError(
				fname,
				"must be in the format 'carrier=number'",
			))
			continue
		}
		err = validate.Many(err,
			validate.ASCII(fname+".Carrier", parts[0], 1, 255),
			validate.Phone(fname+".Phone", parts[1]),
		)
		if m[parts[0]] {
			err = validate.Many(err, validation.NewFieldError(fname, fmt.Sprintf("carrier override '%s' already set", parts[0])))
		}
		m[parts[0]] = true
	}

	return err
}
