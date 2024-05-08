package integrationkey

import (
	"github.com/target/goalert/validation/validate"
)

type IntegrationKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      Type   `json:"type"`
	ServiceID string `json:"service_id"`

	ExternalSystemName string
}

func (i IntegrationKey) Normalize() (*IntegrationKey, error) {
	err := validate.Many(
		validate.IDName("Name", i.Name),
		validate.UUID("ServiceID", i.ServiceID),
		validate.OneOf("Type", i.Type, TypeGrafana, TypeSite24x7, TypePrometheusAlertmanager, TypeGeneric, TypeEmail, TypeUniversal),
		validate.ASCII("ExternalSystemName", i.ExternalSystemName, 0, 255),
	)
	if err != nil {
		return nil, err
	}

	return &i, nil
}
