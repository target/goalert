package graphql

import (
	"net/url"
	"time"

	g "github.com/graphql-go/graphql"
	"github.com/pkg/errors"
	"github.com/target/goalert/heartbeat"
)

func getHeartbeatMonitor(src interface{}) (*heartbeat.Monitor, error) {
	switch s := src.(type) {
	case *heartbeat.Monitor:
		return s, nil
	case heartbeat.Monitor:
		return &s, nil
	default:
		return nil, errors.Errorf("could not get heartbeat key (unknown source type %T)", s)
	}
}

var heartbeatMonitorState = g.NewEnum(g.EnumConfig{
	Name: "HeartbeatMonitorState",
	Values: g.EnumValueConfigMap{
		"inactive":  &g.EnumValueConfig{Value: heartbeat.StateInactive},
		"healthy":   &g.EnumValueConfig{Value: heartbeat.StateHealthy},
		"unhealthy": &g.EnumValueConfig{Value: heartbeat.StateUnhealthy},
	},
})

func (h *Handler) heartbeatMonitorFields() g.Fields {
	return g.Fields{
		"id":         &g.Field{Type: g.String},
		"name":       &g.Field{Type: g.String},
		"service_id": &g.Field{Type: g.String},
		"last_state": &g.Field{
			Type: heartbeatMonitorState,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				h, err := getHeartbeatMonitor(p.Source)
				if err != nil {
					return nil, err
				}
				return h.LastState(), nil
			},
		},
		"last_heartbeat_minutes": &g.Field{
			Type:        g.Int,
			Description: "Number of full minutes elapsed since last heartbeat.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				h, err := getHeartbeatMonitor(p.Source)
				if err != nil {
					return nil, err
				}
				last := h.LastHeartbeat()
				if last.IsZero() {
					return nil, nil
				}

				return time.Since(last) / time.Minute, nil
			},
		},
		"interval_minutes": &g.Field{Type: g.Int},

		"href": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				h, err := getHeartbeatMonitor(p.Source)
				if err != nil {
					return nil, err
				}
				return "/v1/api/heartbeat/" + url.PathEscape(h.ID), nil
			},
		},
	}
}

func (h *Handler) deleteHeartbeatMonitorField() *g.Field {
	return &g.Field{
		Type: g.NewObject(g.ObjectConfig{
			Name:   "DeleteHeartbeatMonitorOutput",
			Fields: g.Fields{"deleted_id": &g.Field{Type: g.String}},
		}),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "DeleteHeartbeatMonitorInput",
					Fields: g.InputObjectConfigFieldMap{
						"id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var r struct {
				ID string `json:"deleted_id"`
			}

			r.ID, _ = m["id"].(string)
			return newScrubber(p.Context).scrub(r, h.c.HeartbeatStore.DeleteTx(p.Context, nil, r.ID))
		},
	}
}
