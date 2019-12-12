package config

// Hints contains information helpful for configuring GoAlert and various integrations.
type Hints struct {
	GitHub struct {
		AuthCallbackURL string
	}
	OIDC struct {
		RedirectURL string
	}
	Mailgun struct {
		ForwardURL string
	}
	Twilio struct {
		MessageWebhookURL string
		VoiceWebhookURL   string
	}
}

// Hints returns available hints for the current configuration.
func (cfg Config) Hints() Hints {
	var h Hints

	h.GitHub.AuthCallbackURL = cfg.CallbackURL("/api/v2/identity/providers/github/callback")
	h.OIDC.RedirectURL = cfg.CallbackURL("/api/v2/identity/providers/oidc/callback")
	h.Mailgun.ForwardURL = cfg.CallbackURL("/api/v2/mailgun/incoming")
	h.Twilio.MessageWebhookURL = cfg.CallbackURL("/api/v2/twilio/message")
	h.Twilio.VoiceWebhookURL = cfg.CallbackURL("/api/v2/twilio/call")

	return h
}
