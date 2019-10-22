// Code generated by devtools/configparams DO NOT EDIT.

package graphql2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/target/goalert/config"
	"github.com/target/goalert/validation"
)

// MapConfigValues will map a Config struct into a flat list of ConfigValue structs.
func MapConfigValues(cfg config.Config) []ConfigValue {
	return []ConfigValue{
		{ID: "General.PublicURL", Type: ConfigTypeString, Description: "Publicly routable URL for UI links and API calls.", Value: cfg.General.PublicURL},
		{ID: "General.GoogleAnalyticsID", Type: ConfigTypeString, Description: "", Value: cfg.General.GoogleAnalyticsID},
		{ID: "General.NotificationDisclaimer", Type: ConfigTypeString, Description: "Disclaimer text for receiving pre-recorded notifications (appears on profile page).", Value: cfg.General.NotificationDisclaimer},
		{ID: "General.DisableLabelCreation", Type: ConfigTypeBoolean, Description: "Disables the ability to create new labels for services.", Value: fmt.Sprintf("%t", cfg.General.DisableLabelCreation)},
		{ID: "General.MessageBundles", Type: ConfigTypeBoolean, Description: "Enables bundling status updates and alert notifications. Also allows 'ack all' responses to bundled alerts.", Value: fmt.Sprintf("%t", cfg.General.MessageBundles)},
		{ID: "General.ShortURL", Type: ConfigTypeString, Description: "If set, messages will contain a shorter URL using this prefix (e.g. http://example.com). It should point to GoAlert and can be the same as the PublicURL.", Value: cfg.General.ShortURL},
		{ID: "Maintenance.AlertCleanupDays", Type: ConfigTypeInteger, Description: "Closed alerts will be deleted after this many days (0 means disable cleanup).", Value: fmt.Sprintf("%d", cfg.Maintenance.AlertCleanupDays)},
		{ID: "Auth.RefererURLs", Type: ConfigTypeStringList, Description: "Allowed referer URLs for auth and redirects.", Value: strings.Join(cfg.Auth.RefererURLs, "\n")},
		{ID: "Auth.DisableBasic", Type: ConfigTypeBoolean, Description: "Disallow username/password login.", Value: fmt.Sprintf("%t", cfg.Auth.DisableBasic)},
		{ID: "GitHub.Enable", Type: ConfigTypeBoolean, Description: "Enable GitHub authentication.", Value: fmt.Sprintf("%t", cfg.GitHub.Enable)},
		{ID: "GitHub.NewUsers", Type: ConfigTypeBoolean, Description: "Allow new user creation via GitHub authentication.", Value: fmt.Sprintf("%t", cfg.GitHub.NewUsers)},
		{ID: "GitHub.ClientID", Type: ConfigTypeString, Description: "", Value: cfg.GitHub.ClientID},
		{ID: "GitHub.ClientSecret", Type: ConfigTypeString, Description: "", Value: cfg.GitHub.ClientSecret, Password: true},
		{ID: "GitHub.AllowedUsers", Type: ConfigTypeStringList, Description: "Allow any of the listed GitHub usernames to authenticate. Use '*' to allow any user.", Value: strings.Join(cfg.GitHub.AllowedUsers, "\n")},
		{ID: "GitHub.AllowedOrgs", Type: ConfigTypeStringList, Description: "Allow any member of any listed GitHub org (or team, using the format 'org/team') to authenticate.", Value: strings.Join(cfg.GitHub.AllowedOrgs, "\n")},
		{ID: "GitHub.EnterpriseURL", Type: ConfigTypeString, Description: "GitHub URL (without /api) when used with GitHub Enterprise.", Value: cfg.GitHub.EnterpriseURL},
		{ID: "OIDC.Enable", Type: ConfigTypeBoolean, Description: "Enable OpenID Connect authentication.", Value: fmt.Sprintf("%t", cfg.OIDC.Enable)},
		{ID: "OIDC.NewUsers", Type: ConfigTypeBoolean, Description: "Allow new user creation via OIDC authentication.", Value: fmt.Sprintf("%t", cfg.OIDC.NewUsers)},
		{ID: "OIDC.OverrideName", Type: ConfigTypeString, Description: "Set the name/label on the login page to something other than OIDC.", Value: cfg.OIDC.OverrideName},
		{ID: "OIDC.IssuerURL", Type: ConfigTypeString, Description: "", Value: cfg.OIDC.IssuerURL},
		{ID: "OIDC.ClientID", Type: ConfigTypeString, Description: "", Value: cfg.OIDC.ClientID},
		{ID: "OIDC.ClientSecret", Type: ConfigTypeString, Description: "", Value: cfg.OIDC.ClientSecret, Password: true},
		{ID: "Mailgun.Enable", Type: ConfigTypeBoolean, Description: "", Value: fmt.Sprintf("%t", cfg.Mailgun.Enable)},
		{ID: "Mailgun.APIKey", Type: ConfigTypeString, Description: "", Value: cfg.Mailgun.APIKey, Password: true},
		{ID: "Mailgun.EmailDomain", Type: ConfigTypeString, Description: "The TO address for all incoming alerts.", Value: cfg.Mailgun.EmailDomain},
		{ID: "Slack.Enable", Type: ConfigTypeBoolean, Description: "", Value: fmt.Sprintf("%t", cfg.Slack.Enable)},
		{ID: "Slack.ClientID", Type: ConfigTypeString, Description: "", Value: cfg.Slack.ClientID},
		{ID: "Slack.ClientSecret", Type: ConfigTypeString, Description: "", Value: cfg.Slack.ClientSecret, Password: true},
		{ID: "Slack.AccessToken", Type: ConfigTypeString, Description: "Slack app bot user OAuth access token (should start with xoxb-).", Value: cfg.Slack.AccessToken, Password: true},
		{ID: "Twilio.Enable", Type: ConfigTypeBoolean, Description: "Enables sending and processing of Voice and SMS messages through the Twilio notification provider.", Value: fmt.Sprintf("%t", cfg.Twilio.Enable)},
		{ID: "Twilio.AccountSID", Type: ConfigTypeString, Description: "", Value: cfg.Twilio.AccountSID},
		{ID: "Twilio.AuthToken", Type: ConfigTypeString, Description: "The primary Auth Token for Twilio. Must be primary (not secondary) for request valiation.", Value: cfg.Twilio.AuthToken, Password: true},
		{ID: "Twilio.FromNumber", Type: ConfigTypeString, Description: "The Twilio number to use for outgoing notifications.", Value: cfg.Twilio.FromNumber},
		{ID: "Feedback.Enable", Type: ConfigTypeBoolean, Description: "Enables Feedback link in nav bar.", Value: fmt.Sprintf("%t", cfg.Feedback.Enable)},
		{ID: "Feedback.OverrideURL", Type: ConfigTypeString, Description: "Use a custom URL for Feedback link in nav bar.", Value: cfg.Feedback.OverrideURL},
	}
}

// MapPublicConfigValues will map a Config struct into a flat list of ConfigValue structs.
func MapPublicConfigValues(cfg config.Config) []ConfigValue {
	return []ConfigValue{
		{ID: "General.GoogleAnalyticsID", Type: ConfigTypeString, Description: "", Value: cfg.General.GoogleAnalyticsID},
		{ID: "General.NotificationDisclaimer", Type: ConfigTypeString, Description: "Disclaimer text for receiving pre-recorded notifications (appears on profile page).", Value: cfg.General.NotificationDisclaimer},
		{ID: "General.DisableLabelCreation", Type: ConfigTypeBoolean, Description: "Disables the ability to create new labels for services.", Value: fmt.Sprintf("%t", cfg.General.DisableLabelCreation)},
		{ID: "General.MessageBundles", Type: ConfigTypeBoolean, Description: "Enables bundling status updates and alert notifications. Also allows 'ack all' responses to bundled alerts.", Value: fmt.Sprintf("%t", cfg.General.MessageBundles)},
		{ID: "General.ShortURL", Type: ConfigTypeString, Description: "If set, messages will contain a shorter URL using this prefix (e.g. http://example.com). It should point to GoAlert and can be the same as the PublicURL.", Value: cfg.General.ShortURL},
		{ID: "Maintenance.AlertCleanupDays", Type: ConfigTypeInteger, Description: "Closed alerts will be deleted after this many days (0 means disable cleanup).", Value: fmt.Sprintf("%d", cfg.Maintenance.AlertCleanupDays)},
		{ID: "Auth.DisableBasic", Type: ConfigTypeBoolean, Description: "Disallow username/password login.", Value: fmt.Sprintf("%t", cfg.Auth.DisableBasic)},
		{ID: "GitHub.Enable", Type: ConfigTypeBoolean, Description: "Enable GitHub authentication.", Value: fmt.Sprintf("%t", cfg.GitHub.Enable)},
		{ID: "OIDC.Enable", Type: ConfigTypeBoolean, Description: "Enable OpenID Connect authentication.", Value: fmt.Sprintf("%t", cfg.OIDC.Enable)},
		{ID: "Mailgun.Enable", Type: ConfigTypeBoolean, Description: "", Value: fmt.Sprintf("%t", cfg.Mailgun.Enable)},
		{ID: "Slack.Enable", Type: ConfigTypeBoolean, Description: "", Value: fmt.Sprintf("%t", cfg.Slack.Enable)},
		{ID: "Twilio.Enable", Type: ConfigTypeBoolean, Description: "Enables sending and processing of Voice and SMS messages through the Twilio notification provider.", Value: fmt.Sprintf("%t", cfg.Twilio.Enable)},
		{ID: "Twilio.FromNumber", Type: ConfigTypeString, Description: "The Twilio number to use for outgoing notifications.", Value: cfg.Twilio.FromNumber},
		{ID: "Feedback.Enable", Type: ConfigTypeBoolean, Description: "Enables Feedback link in nav bar.", Value: fmt.Sprintf("%t", cfg.Feedback.Enable)},
		{ID: "Feedback.OverrideURL", Type: ConfigTypeString, Description: "Use a custom URL for Feedback link in nav bar.", Value: cfg.Feedback.OverrideURL},
	}
}

// ApplyConfigValues will apply a list of ConfigValues to a Config struct.
func ApplyConfigValues(cfg config.Config, vals []ConfigValueInput) (config.Config, error) {
	parseStringList := func(v string) []string {
		if v == "" {
			return nil
		}
		return strings.Split(v, "\n")
	}
	parseInt := func(id, v string) (int, error) {
		if v == "" {
			return 0, nil
		}
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, validation.NewFieldError("\""+id+"\".Value", "integer value invalid: "+err.Error())
		}
		return int(val), nil
	}
	parseBool := func(id, v string) (bool, error) {
		switch v {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return false, validation.NewFieldError("\""+id+"\".Value", "boolean value invalid: expected 'true' or 'false'")
		}
	}
	for _, v := range vals {
		switch v.ID {
		case "General.PublicURL":
			cfg.General.PublicURL = v.Value
		case "General.GoogleAnalyticsID":
			cfg.General.GoogleAnalyticsID = v.Value
		case "General.NotificationDisclaimer":
			cfg.General.NotificationDisclaimer = v.Value
		case "General.DisableLabelCreation":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.General.DisableLabelCreation = val
		case "General.MessageBundles":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.General.MessageBundles = val
		case "General.ShortURL":
			cfg.General.ShortURL = v.Value
		case "Maintenance.AlertCleanupDays":
			val, err := parseInt(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.Maintenance.AlertCleanupDays = val
		case "Auth.RefererURLs":
			cfg.Auth.RefererURLs = parseStringList(v.Value)
		case "Auth.DisableBasic":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.Auth.DisableBasic = val
		case "GitHub.Enable":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.GitHub.Enable = val
		case "GitHub.NewUsers":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.GitHub.NewUsers = val
		case "GitHub.ClientID":
			cfg.GitHub.ClientID = v.Value
		case "GitHub.ClientSecret":
			cfg.GitHub.ClientSecret = v.Value
		case "GitHub.AllowedUsers":
			cfg.GitHub.AllowedUsers = parseStringList(v.Value)
		case "GitHub.AllowedOrgs":
			cfg.GitHub.AllowedOrgs = parseStringList(v.Value)
		case "GitHub.EnterpriseURL":
			cfg.GitHub.EnterpriseURL = v.Value
		case "OIDC.Enable":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.OIDC.Enable = val
		case "OIDC.NewUsers":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.OIDC.NewUsers = val
		case "OIDC.OverrideName":
			cfg.OIDC.OverrideName = v.Value
		case "OIDC.IssuerURL":
			cfg.OIDC.IssuerURL = v.Value
		case "OIDC.ClientID":
			cfg.OIDC.ClientID = v.Value
		case "OIDC.ClientSecret":
			cfg.OIDC.ClientSecret = v.Value
		case "Mailgun.Enable":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.Mailgun.Enable = val
		case "Mailgun.APIKey":
			cfg.Mailgun.APIKey = v.Value
		case "Mailgun.EmailDomain":
			cfg.Mailgun.EmailDomain = v.Value
		case "Slack.Enable":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.Slack.Enable = val
		case "Slack.ClientID":
			cfg.Slack.ClientID = v.Value
		case "Slack.ClientSecret":
			cfg.Slack.ClientSecret = v.Value
		case "Slack.AccessToken":
			cfg.Slack.AccessToken = v.Value
		case "Twilio.Enable":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.Twilio.Enable = val
		case "Twilio.AccountSID":
			cfg.Twilio.AccountSID = v.Value
		case "Twilio.AuthToken":
			cfg.Twilio.AuthToken = v.Value
		case "Twilio.FromNumber":
			cfg.Twilio.FromNumber = v.Value
		case "Feedback.Enable":
			val, err := parseBool(v.ID, v.Value)
			if err != nil {
				return cfg, err
			}
			cfg.Feedback.Enable = val
		case "Feedback.OverrideURL":
			cfg.Feedback.OverrideURL = v.Value
		default:
			return cfg, validation.NewFieldError("ID", fmt.Sprintf("unknown config ID '%s'", v.ID))
		}
	}
	return cfg, nil
}
