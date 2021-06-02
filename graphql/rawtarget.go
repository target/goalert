package graphql

import "github.com/target/goalert/assignment"

type RawTarget struct {
	Type assignment.TargetType `json:"target_type"`
	ID   string                `json:"target_id"`
	Name string                `json:"target_name"`
}

func (rt RawTarget) TargetType() assignment.TargetType {
	return rt.Type
}
func (rt RawTarget) TargetID() string {
	return rt.ID
}
