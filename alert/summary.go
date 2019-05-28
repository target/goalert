package alert

type Summary struct {
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
	Totals      struct {
		Unack  int `json:"unacknowledged"`
		Ack    int `json:"acknowledged"`
		Closed int `json:"closed"`
	} `json:"totals"`
}
