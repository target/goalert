package alert

import (
	"database/sql/driver"
	"fmt"
)

// Source is the entity that triggered an alert.
type Source string

// Source types
const (
	SourceEmail                  Source = "email"                  // email alert
	SourceGrafana                Source = "grafana"                // grafana alert
	SourceSite24x7               Source = "site24x7"               // site24x7 alert
	SourcePrometheusAlertmanager Source = "prometheusAlertmanager" // prometheus alertmanager alert
	SourceManual                 Source = "manual"                 // manually triggered
	SourceGeneric                Source = "generic"                // generic API
	SourceNotify                 Source = "notify"                 // notify API
)

func (s Source) Value() (driver.Value, error) {
	str := string(s)
	if str == "" {
		str = string(SourceManual)
	}
	return str, nil
}

func (s *Source) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*s = Source(t)
	case string:
		*s = Source(t)
	default:
		return fmt.Errorf("could not process unknown type for source %T", t)
	}
	return nil
}
