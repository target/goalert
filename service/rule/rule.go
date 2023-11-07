package rule

type Rule struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	ServiceID       string   `json:"service_id"`
	IntegrationKeys []string `json:"integration_keys"`
	FilterString    string   `json:"filters"`
	SendAlert       bool     `json:"send_alert"`
	Actions         []Action `json:"actions"`
}

type Action struct {
	DestType string    `json:"dest_type"`
	Contents []Content `json:"contents"`
}

type Content struct {
	Prop  string `json:"prop"`
	Value string `json:"value"`
}
