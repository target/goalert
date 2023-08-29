package integrationkey

import (
	"database/sql/driver"
	"fmt"
)

// Type is the entity that needs an integration.
type Type string

// Types
const (
	TypeGrafana                Type = "grafana"
	TypeSite24x7               Type = "site24x7"
	TypePrometheusAlertmanager Type = "prometheusAlertmanager"
	TypeGeneric                Type = "generic"
	TypeNotify                 Type = "notify"
	TypeEmail                  Type = "email"
)

func (s Type) Value() (driver.Value, error) {
	str := string(s)
	return str, nil
}

func (s *Type) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*s = Type(t)
	case string:
		*s = Type(t)
	default:
		return fmt.Errorf("could not process unknown type for source %T", t)
	}
	return nil
}
