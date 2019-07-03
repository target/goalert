package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// SchemaVersion indicates the current config struct version.
const SchemaVersion = 1

// Config contains GoAlert application settings.
type Config struct {
	data        []byte
	fallbackURL string

	General struct {
		PublicURL              string `info:"Publicly routable URL for UI links and API calls."`
		GoogleAnalyticsID      string `public:"true"`
		NotificationDisclaimer string `public:"true" info:"Disclaimer text for receiving pre-recorded notifications (appears on profile page)."`
		DisableLabelCreation   bool   `public:"true" info:"Disables the ability to create new labels for services."`
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

		AccessToken string `password:"true" info:"Slack app OAuth access token."`
	}

	Twilio struct {
		Enable bool `public:"true" info:"Enables sending and processing of Voice and SMS messages through the Twilio notification provider."`

		AccountSID string
		AuthToken  string `password:"true" info:"The primary Auth Token for Twilio. Must be primary (not secondary) for request valiation."`
		FromNumber string `public:"true" info:"The Twilio number to use for outgoing notifications."`
	}

	Feedback struct {
		Enable      bool   `public:"true" info:"Enables Feedback link in nav bar."`
		OverrideURL string `public:"true" info:"Use a custom URL for Feedback link in nav bar."`
	}
}

// CallbackURL will return a public-routable URL to the given path.
// It will use PublicURL() to fill in missing pieces.
//
// It will panic if provided an invalid URL.
func (cfg Config) CallbackURL(path string, mergeParams ...url.Values) string {
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
	return base.String()
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
		return strings.HasPrefix(ref, u.String())
	}

	for _, u := range cfg.Auth.RefererURLs {
		if strings.HasPrefix(ref, u) {
			return true
		}
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

var (
	mailgunKeyRx = regexp.MustCompile(`^key-[0-9a-f]{32}$`)

	slackClientIDRx        = regexp.MustCompile(`^[0-9]{10,12}\.[0-9]{10,12}$`)
	slackClientSecretRx    = regexp.MustCompile(`^[0-9a-f]{32}$`)
	slackUserAccessTokenRx = regexp.MustCompile(`^xoxp-[0-9]{10,12}-[0-9]{10,12}-[0-9]{10,12}-[0-9a-f]{32}$`)
	slackBotAccessTokenRx  = regexp.MustCompile(`^xoxb-[0-9]{10,12}-[0-9]{10,12}-[0-9A-Za-z]{24}$`)

	twilioAccountSIDRx = regexp.MustCompile(`^AC[0-9a-f]{32}$`)
	twilioAuthTokenRx  = regexp.MustCompile(`^[0-9a-f]{32}$`)

	githubClientIDRx     = regexp.MustCompile(`^[0-9a-f]{20}$`)
	githubClientSecretRx = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

func validateRx(field string, rx *regexp.Regexp, value, msg string) error {
	if value == "" {
		return nil
	}

	if !rx.MatchString(value) {
		return validation.NewFieldError(field, msg)
	}

	return nil
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

	err = validate.Many(
		err,
		validate.Text("General.NotificationDisclaimer", cfg.General.NotificationDisclaimer, 0, 500),

		validateRx("Mailgun.APIKey", mailgunKeyRx, cfg.Mailgun.APIKey, "should be of the format: 'key-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'"),

		validateRx("Slack.ClientID", slackClientIDRx, cfg.Slack.ClientID, "should be of the format: '############.############'"),
		validateRx("Slack.ClientSecret", slackClientSecretRx, cfg.Slack.ClientSecret, "should be of the format: 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'"),

		validateRx("Twilio.AccountSID", twilioAccountSIDRx, cfg.Twilio.AccountSID, "should be of the format: 'ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'"),
		validateRx("Twilio.AuthToken", twilioAuthTokenRx, cfg.Twilio.AuthToken, "should be of the format: 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'"),

		validateRx("GitHub.ClientID", githubClientIDRx, cfg.GitHub.ClientID, "should be of the format: 'xxxxxxxxxxxxxxxxxxxx'"),
		validateRx("GitHub.ClientSecret", githubClientSecretRx, cfg.GitHub.ClientSecret, "should be of the format: 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'"),
	)

	if strings.HasPrefix(cfg.Slack.AccessToken, "xoxp") {
		err = validate.Many(err,
			validateRx("Slack.AccessToken", slackUserAccessTokenRx, cfg.Slack.AccessToken, "should be of the format: 'xoxb-############-############-zzzzzzzzzzzzzzzzzzzzzzzz'"),
		)
	} else {
		err = validate.Many(err,
			validateRx("Slack.AccessToken", slackBotAccessTokenRx, cfg.Slack.AccessToken, "should be of the format: 'xoxb-############-############-zzzzzzzzzzzzzzzzzzzzzzzz'"),
		)
	}

	if cfg.OIDC.IssuerURL != "" {
		err = validate.Many(err, validate.AbsoluteURL("OIDC.IssuerURL", cfg.OIDC.IssuerURL))
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
	)

	if cfg.Feedback.OverrideURL != "" {
		err = validate.Many(
			err,
			validate.AbsoluteURL("Feedback.OverrideURL", cfg.Feedback.OverrideURL),
		)
	}

	for i, urlStr := range cfg.Auth.RefererURLs {
		field := fmt.Sprintf("Auth.RefererURLs[%d]", i)
		err = validate.Many(
			err,
			validate.AbsoluteURL(field, urlStr),
		)
	}

	return err
}
