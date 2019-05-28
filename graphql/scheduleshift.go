package graphql

import (
	"fmt"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/schedule/shiftcalc"

	g "github.com/graphql-go/graphql"
)

func (h *Handler) scheduleShiftFields() g.Fields {
	return g.Fields{
		"start_time": &g.Field{Type: ISOTimestamp},
		"end_time":   &g.Field{Type: ISOTimestamp},
		"truncated":  &g.Field{Type: g.Boolean},
		"user_id":    &g.Field{Type: g.String},
		"user": &g.Field{
			Type: h.user,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				var userID string
				switch s := p.Source.(type) {
				case *shiftcalc.Shift:
					userID = s.UserID
				case shiftcalc.Shift:
					userID = s.UserID
				case oncall.Shift:
					userID = s.UserID
				case *oncall.Shift:
					userID = s.UserID
				default:
					return nil, fmt.Errorf("could not id of user (unknown source type %T)", s)
				}

				return newScrubber(p.Context).scrub(h.c.UserStore.FindOne(p.Context, userID))
			},
		},
	}
}
