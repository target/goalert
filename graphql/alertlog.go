package graphql

import (
	g "github.com/graphql-go/graphql"
)

func (h *Handler) alertLogFields() g.Fields {
	return g.Fields{
		"alert_id":  &g.Field{Type: g.String},
		"timestamp": &g.Field{Type: ISOTimestamp},
		"event":     &g.Field{Type: alertLogEventType},
		"message":   &g.Field{Type: g.String},
		"subject":   &g.Field{Type: h.alertLogSubject},
	}
}
