package integrationkeyrule // unsure about this naming, seems confusing

type Rule struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	IntegrationKey string          `json:"integration_key"`
	Filter         string          `json:"filter"`
	Summary        string          `json:"summary"`
	Details        string          `json:"details"`
	Dedup          string          `json:"dedup"`
	Action         AlertActionType `json:"action"`
}

type AlertActionType int

const (
	CreateAlert AlertActionType = iota
	CloseAlert
)
