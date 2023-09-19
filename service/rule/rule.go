package rule

type Rule struct {
	ID              string                   `json:"id"`
	Name            string                   `json:"name"`
	ServiceID       string                   `json:"service_id"`
	IntegrationKeys []string                 `json:"integration_keys"`
	FilterString    string                   `json:"filters"`
	SendAlert       bool                     `json:"send_alert"`
	Actions         []map[string]interface{} `json:"actions"`
}
