package graphql

import (
	"fmt"
	"github.com/target/goalert/alert"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/user"
	"github.com/target/goalert/validation"
	"sort"
	"strings"
	"time"

	g "github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

type escalationPolicySnapshot struct {
	Repeat         int                            `json:"repeat"`
	CurrentLevel   int                            `json:"current_level"`
	LastEscalation time.Time                      `json:"last_escalation"`
	Steps          []escalationPolicySnapshotStep `json:"steps"`
}

type escalationPolicySnapshotStep struct {
	DelayMinutes int      `json:"delay_minutes"`
	UserIDs      []string `json:"user_ids"`
	ScheduleIDs  []string `json:"schedule_ids"`
}

type alertLogSearchResult struct {
	ID        int               `json:"id"`
	AlertID   int               `json:"alert_id"`
	Log       string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Subject   *alertlog.Subject `json:"subject"`
	Event     alertlog.Type     `json:"event"`
}

func (h *Handler) alertLogSubjectFields() g.Fields {
	return g.Fields{
		"id":         &g.Field{Type: g.String},
		"name":       &g.Field{Type: g.String},
		"type":       &g.Field{Type: g.String},
		"classifier": &g.Field{Type: g.String},
	}
}

var alertLogEventType = g.NewEnum(g.EnumConfig{
	Name: "AlertLogEventType",
	Values: g.EnumValueConfigMap{
		"created":              &g.EnumValueConfig{Value: alertlog.TypeCreated},
		"closed":               &g.EnumValueConfig{Value: alertlog.TypeClosed},
		"escalated":            &g.EnumValueConfig{Value: alertlog.TypeEscalated},
		"acknowledged":         &g.EnumValueConfig{Value: alertlog.TypeAcknowledged},
		"escalation_request":   &g.EnumValueConfig{Value: alertlog.TypeEscalationRequest},
		"notification_sent":    &g.EnumValueConfig{Value: alertlog.TypeNotificationSent},
		"no_notification_sent":    &g.EnumValueConfig{Value: alertlog.TypeNoNotificationSent},
		"policy_updated":       &g.EnumValueConfig{Value: alertlog.TypePolicyUpdated},
		"duplicate_suppressed": &g.EnumValueConfig{Value: alertlog.TypeDuplicateSupressed},
	},
})

func getEPSStep(src interface{}) (*escalationPolicySnapshotStep, error) {
	switch s := src.(type) {
	case *escalationPolicySnapshotStep:
		return s, nil
	case escalationPolicySnapshotStep:
		return &s, nil
	default:
		return nil, errors.Errorf("could not id of EPS Step (unknown source type %T)", s)
	}
}
func getAlertSummary(p interface{}) (*alert.Summary, error) {
	switch s := p.(type) {
	case alert.Summary:
		return &s, nil
	case *alert.Summary:
		return s, nil
	default:
		return nil, errors.Errorf("could not get Summary (unknown source type %T)", s)
	}
}

func (h *Handler) resolveUsersFromIDs(p g.ResolveParams) (interface{}, error) {
	src, err := getEPSStep(p.Source)
	if err != nil {
		return nil, err
	}

	users := make([]user.User, 0, len(src.UserIDs))
	var u *user.User
	for _, id := range src.UserIDs {
		u, err = h.c.UserStore.FindOne(p.Context, id)
		if err != nil {
			return newScrubber(p.Context).scrub(nil, err)
		}
		users = append(users, *u)
	}
	return users, nil
}

func (h *Handler) resolveSchedulesFromIDs(p g.ResolveParams) (interface{}, error) {
	src, err := getEPSStep(p.Source)
	if err != nil {
		return nil, err
	}

	schedules := make([]schedule.Schedule, 0, len(src.ScheduleIDs))
	var s *schedule.Schedule
	for _, id := range src.ScheduleIDs {
		s, err = h.c.ScheduleStore.FindOne(p.Context, id)
		if err != nil {
			return newScrubber(p.Context).scrub(nil, err)
		}
		schedules = append(schedules, *s)
	}
	return schedules, nil
}

func (h *Handler) alertSummaryFields() g.Fields {
	return g.Fields{
		"service_id":   &g.Field{Type: g.String},
		"service_name": &g.Field{Type: g.String},
		"service": &g.Field{
			Type: h.service,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				a, err := getAlertSummary(p.Source)
				if err != nil {
					return nil, err
				}

				userID := permission.UserID(p.Context)
				return newScrubber(p.Context).scrub(h.c.ServiceStore.FindOneForUser(p.Context, userID, a.ServiceID))
			},
		},

		"totals": &g.Field{
			Type: g.NewObject(g.ObjectConfig{
				Name: "AlertTotals",
				Fields: g.Fields{
					"unacknowledged": &g.Field{Type: g.Int},
					"acknowledged":   &g.Field{Type: g.Int},
					"closed":         &g.Field{Type: g.Int},
				},
			}),
		},
	}
}

func (h *Handler) alertFields() g.Fields {
	return g.Fields{
		"_id": &g.Field{Type: g.Int},
		"id": &g.Field{
			Type:        g.String,
			Description: "Provides a unique string-version of id for use with Relay in the form of Alert(id).",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				a, err := getAlert(p.Source)
				if err != nil {
					return nil, err
				}

				return fmt.Sprintf("Alert(%d)", a.ID), nil
			},
		},
		"assignments": &g.Field{
			Type: g.NewList(h.user),
		},

		"status": &g.Field{Type: alertStatus, DeprecationReason: "Use the 'status_2' field instead."},
		"status_2": &g.Field{
			Type: alertStatus,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				a, err := getAlert(p.Source)
				if err != nil {
					return "", err
				}
				return a.Status, nil
			},
		},
		"description": &g.Field{
			Type:              g.String,
			DeprecationReason: "Use the 'summary' and 'details' fields instead.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				a, err := getAlert(p.Source)
				if err != nil {
					return nil, err
				}

				return a.Description(), nil
			},
		},
		"source":     &g.Field{Type: g.String},
		"service_id": &g.Field{Type: g.String},
		"service_name": &g.Field{
			Type: g.String,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				a, err := getAlert(p.Source)
				if err != nil {
					return nil, err
				}
				s, err := h.c.ServiceStore.FindOneForUser(p.Context, permission.UserID(p.Context), a.ServiceID)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}
				return s.Name, nil
			},
		},
		"service": &g.Field{
			Type: h.service,
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				a, err := getAlert(p.Source)
				if err != nil {
					return nil, err
				}

				userID := permission.UserID(p.Context)
				return newScrubber(p.Context).scrub(h.c.ServiceStore.FindOneForUser(p.Context, userID, a.ServiceID))
			},
		},

		"summary": &g.Field{Type: g.String},
		"details": &g.Field{Type: g.String},

		"escalation_level": &g.Field{Type: g.Int, Description: "The total number of escalations for this alert.",
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				a, err := getAlert(p.Source)
				if err != nil {
					return nil, err
				}
				l, err := h.c.AlertStore.State(p.Context, []int{a.ID})
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}
				if len(l) == 0 {
					return -1, nil
				}
				return l[0].StepNumber, nil
			},
		},

		"logs": &g.Field{
			Type: g.NewList(h.alertLog),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				a, err := getAlert(p.Source)
				if err != nil {
					return nil, err
				}
				var opts alertlog.LegacySearchOptions
				opts.AlertID = a.ID
				opts.Limit = 50
				entries, _, err := h.c.AlertLogStore.LegacySearch(p.Context, &opts)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}
				custom := make([]alertLogSearchResult, len(entries))
				for i, e := range entries {
					custom[i].AlertID = e.AlertID()
					custom[i].Log = e.String()
					custom[i].ID = e.ID()
					custom[i].Timestamp = e.Timestamp()
					custom[i].Subject = e.Subject()
				}
				return custom, nil
			},
			DeprecationReason: "Use the 'logs_2' field instead.",
		},
		"logs_2": &g.Field{
			Type: g.NewList(h.alertLog),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				a, err := getAlert(p.Source)
				if err != nil {
					return nil, err
				}

				var opts alertlog.LegacySearchOptions
				opts.AlertID = a.ID
				entries, _, err := h.c.AlertLogStore.LegacySearch(p.Context, &opts)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}
				custom := make([]alertLogSearchResult, len(entries))
				for i, e := range entries {
					custom[i].AlertID = e.AlertID()
					custom[i].Log = e.String()
					custom[i].ID = e.ID()
					custom[i].Timestamp = e.Timestamp()
					custom[i].Subject = e.Subject()
				}
				return custom, nil
			},
		},
		"created_at": &g.Field{Type: ISOTimestamp},
		"escalation_policy_snapshot": &g.Field{
			Type: g.NewObject(g.ObjectConfig{
				Name:        "EscalationPolicySnapshot",
				Description: "Snapshot of an escalation policy used by an Alert.",
				Fields: g.Fields{
					"repeat":          &g.Field{Type: g.Int},
					"current_level":   &g.Field{Type: g.Int},
					"last_escalation": &g.Field{Type: ISOTimestamp},
					"steps": &g.Field{Type: g.NewList(g.NewObject(g.ObjectConfig{
						Name: "EscalationPolicySnapshotStep",
						Fields: g.Fields{
							"delay_minutes": &g.Field{Type: g.Int},
							"user_ids":      &g.Field{Type: g.NewList(g.String)},
							"users":         &g.Field{Type: g.NewList(h.user), Resolve: h.resolveUsersFromIDs},
							"schedule_ids":  &g.Field{Type: g.NewList(g.String)},
							"schedules":     &g.Field{Type: g.NewList(h.schedule), Resolve: h.resolveSchedulesFromIDs},
						},
					}))},
				},
			}),
			Resolve: func(p g.ResolveParams) (interface{}, error) {
				a, err := getAlert(p.Source)
				if err != nil {
					return nil, err
				}

				polID, err := h.c.Resolver.AlertEPID(p.Context, a.ID)

				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}

				if polID == "" {
					return nil, fmt.Errorf("no escalation policy found")
				}

				pol, err := h.c.EscalationStore.FindOnePolicy(p.Context, polID)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}
				act, err := h.c.EscalationStore.ActiveStep(p.Context, a.ID, polID)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}
				steps, err := h.c.EscalationStore.FindAllSteps(p.Context, polID)
				if err != nil {
					return newScrubber(p.Context).scrub(nil, err)
				}
				sort.Slice(steps, func(i, j int) bool { return steps[i].StepNumber < steps[j].StepNumber })

				if act == nil {
					act = &escalation.ActiveStep{
						AlertID:        a.ID,
						LastEscalation: time.Now(),
						PolicyID:       polID,
					}
					if len(steps) > 0 {
						act.StepID = steps[0].ID
					}
				}

				eps := &escalationPolicySnapshot{
					Repeat:         pol.Repeat,
					LastEscalation: act.LastEscalation,
				}

				for _, step := range steps {
					if step.ID == act.StepID {
						eps.CurrentLevel = len(steps)*act.LoopCount + step.StepNumber
					}
					epss := escalationPolicySnapshotStep{
						DelayMinutes: step.DelayMinutes,
					}
					asn, err := h.c.EscalationStore.FindAllStepTargets(p.Context, step.ID)
					if err != nil {
						return newScrubber(p.Context).scrub(nil, err)
					}
					for _, a := range asn {
						switch a.TargetType() {
						case assignment.TargetTypeUser:
							epss.UserIDs = append(epss.UserIDs, a.TargetID())
						case assignment.TargetTypeSchedule:
							epss.ScheduleIDs = append(epss.ScheduleIDs, a.TargetID())
						}
					}
					eps.Steps = append(eps.Steps, epss)
				}

				return eps, nil
			},
		},
	}
}

func getAlert(src interface{}) (*alert.Alert, error) {
	switch a := src.(type) {
	case alert.Alert:
		return &a, nil
	case *alert.Alert:
		return a, nil
	default:
		return nil, fmt.Errorf("could not resolve id of alert (unknown source type %T)", a)
	}
}

func (h *Handler) updateAlertStatusByServiceField() *g.Field {
	return &g.Field{
		Type: g.NewList(h.alert),
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "UpdateAlertStatusByServiceInput",
					Fields: g.InputObjectConfigFieldMap{
						"service_id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String), Description: "The ID of the service"},
						"status":     &g.InputObjectFieldConfig{Type: g.NewNonNull(alertStatus)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			serviceID, _ := m["service_id"].(string)
			status, _ := m["status"].(alert.Status)

			return newScrubber(p.Context).scrub(nil, h.c.AlertStore.UpdateStatusByService(p.Context, serviceID, status))
		},
	}
}

func (h *Handler) escalateAlertField() *g.Field {
	return &g.Field{
		Type: h.alert,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "EscalateAlertInput",
					Fields: g.InputObjectConfigFieldMap{
						"id": &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int), Description: "The ID of the alert (_id field in GraphQL)"},
						"current_escalation_level": &g.InputObjectFieldConfig{
							Type:        g.NewNonNull(g.Int),
							Description: "The current escalation level of the alert.",
						},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}
			id, _ := m["id"].(int)
			lvl, _ := m["current_escalation_level"].(int)
			err := h.c.AlertStore.Escalate(p.Context, id, lvl)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}

			return newScrubber(p.Context).scrub(h.c.AlertStore.FindOne(p.Context, id))
		},
	}
}

var alertStatus = g.NewEnum(g.EnumConfig{
	Name: "AlertStatus",
	Values: g.EnumValueConfigMap{
		"unacknowledged": &g.EnumValueConfig{Value: alert.StatusTriggered},
		"acknowledged":   &g.EnumValueConfig{Value: alert.StatusActive},
		"closed":         &g.EnumValueConfig{Value: alert.StatusClosed},
	},
})

func getAlertStatus(m map[string]interface{}) alert.Status {
	s, ok := m["status"].(alert.Status)
	if ok {
		return s
	}

	s2, ok := m["status_2"].(alert.Status)
	if ok {
		return s2
	}

	return ""
}

func (h *Handler) createAlertField() *g.Field {
	return &g.Field{
		Type: h.alert,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "CreateAlertInput",
					Fields: g.InputObjectConfigFieldMap{
						"status":      &g.InputObjectFieldConfig{Type: alertStatus},
						"status_2":    &g.InputObjectFieldConfig{Type: alertStatus},
						"description": &g.InputObjectFieldConfig{Type: g.String},
						"summary":     &g.InputObjectFieldConfig{Type: g.String},
						"details":     &g.InputObjectFieldConfig{Type: g.String},
						"service_id":  &g.InputObjectFieldConfig{Type: g.NewNonNull(g.String)},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			var a alert.Alert
			a.Status = getAlertStatus(m)

			if string(a.Status) == "" {
				a.Status = alert.StatusTriggered
			}
			a.Summary, _ = m["summary"].(string)
			a.Details, _ = m["details"].(string)
			if a.Summary == "" {
				desc, _ := m["description"].(string)
				parts := strings.SplitN(desc, "\n", 2)
				a.Summary = parts[0]
				if len(parts) == 2 {
					a.Details = parts[1]
				} else {
					a.Details = ""
				}
			}

			a.Source = alert.SourceManual
			a.ServiceID, _ = m["service_id"].(string)

			return newScrubber(p.Context).scrub(h.c.AlertStore.CreateOrUpdate(p.Context, &a))
		},
	}
}

func (h *Handler) updateStatusAlertField() *g.Field {
	return &g.Field{
		Type: h.alert,
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Description: "because bugs.",
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "UpdateAlertStatusInput",
					Fields: g.InputObjectConfigFieldMap{
						"id":       &g.InputObjectFieldConfig{Type: g.NewNonNull(g.Int)},
						"status":   &g.InputObjectFieldConfig{Type: alertStatus},
						"status_2": &g.InputObjectFieldConfig{Type: alertStatus},
					},
				}),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			id, _ := m["id"].(int)
			status := getAlertStatus(m)

			if string(status) == "" {
				status = alert.StatusTriggered
			}

			err := h.c.AlertStore.UpdateStatus(p.Context, id, status)
			if alert.IsAlreadyAcknowledged(err) || alert.IsAlreadyClosed(err) {
				err = nil
			}
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}
			return newScrubber(p.Context).scrub(h.c.AlertStore.FindOne(p.Context, id))
		},
	}
}

func (h *Handler) alertField() *g.Field {
	return &g.Field{
		Type: h.alert,
		Args: g.FieldConfigArgument{
			"id": &g.ArgumentConfig{
				Type: g.NewNonNull(g.Int),
			},
		},
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			id, ok := p.Args["id"].(int)
			if !ok {
				return nil, validation.NewFieldError("id", "required")
			}

			return newScrubber(p.Context).scrub(h.c.AlertStore.FindOne(p.Context, id))
		},
	}
}

func (h *Handler) alertsField() *g.Field {
	return &g.Field{
		Name:              "Alerts",
		Type:              g.NewList(h.alert),
		DeprecationReason: "Use the 'alerts2' field instead.",
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			a, _, err := h.c.AlertStore.LegacySearch(p.Context, nil)
			return newScrubber(p.Context).scrub(a, err)
		},
	}
}

var alertSortByEnum = g.NewEnum(g.EnumConfig{
	Name: "AlertSortBy",
	Values: g.EnumValueConfigMap{
		"status":     &g.EnumValueConfig{Value: alert.SortByStatus},
		"id":         &g.EnumValueConfig{Value: alert.SortByID},
		"created_at": &g.EnumValueConfig{Value: alert.SortByCreatedTime},
		"summary":    &g.EnumValueConfig{Value: alert.SortBySummary},
		"service":    &g.EnumValueConfig{Value: alert.SortByServiceName},
	},
})

func (h *Handler) searchAlertsField() *g.Field {
	return &g.Field{
		Name: "Alerts2",
		Args: g.FieldConfigArgument{
			"options": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "AlertSearchOptions",
					Fields: g.InputObjectConfigFieldMap{
						"search": &g.InputObjectFieldConfig{
							Type:        g.String,
							Description: "Searches for case-insensitive summary or service name substring match, or exact id match.",
						},
						"service_id":     &g.InputObjectFieldConfig{Type: g.String},
						"sort_by":        &g.InputObjectFieldConfig{Type: alertSortByEnum},
						"sort_desc":      &g.InputObjectFieldConfig{Type: g.Boolean},
						"omit_triggered": &g.InputObjectFieldConfig{Type: g.Boolean},
						"omit_active":    &g.InputObjectFieldConfig{Type: g.Boolean},
						"omit_closed":    &g.InputObjectFieldConfig{Type: g.Boolean},
						"limit": &g.InputObjectFieldConfig{
							Type:        g.Int,
							Description: "Defaulted to 50 if not supplied.",
						},
						"offset":                 &g.InputObjectFieldConfig{Type: g.Int},
						"favorite_services_only": &g.InputObjectFieldConfig{Type: g.Boolean},
					},
				}),
			},
		},
		Type: g.NewObject(g.ObjectConfig{
			Name: "AlertSearchResult",
			Fields: g.Fields{
				"items":       &g.Field{Type: g.NewList(h.alert)},
				"total_count": &g.Field{Type: g.Int},
			},
		}),
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			var result struct {
				Items []alert.Alert `json:"items"`
				Total int           `json:"total_count"`
			}

			var opts alert.LegacySearchOptions
			if m, ok := p.Args["options"].(map[string]interface{}); ok {
				opts.Search, _ = m["search"].(string)
				opts.ServiceID, _ = m["service_id"].(string)
				opts.SortBy, _ = m["sort_by"].(alert.SortBy)
				opts.SortDesc, _ = m["sort_desc"].(bool)
				opts.OmitTriggered, _ = m["omit_triggered"].(bool)
				opts.OmitActive, _ = m["omit_active"].(bool)
				opts.OmitClosed, _ = m["omit_closed"].(bool)
				opts.Limit, _ = m["limit"].(int)
				opts.Offset, _ = m["offset"].(int)
				if v, ok := m["favorite_services_only"].(bool); ok && v {
					opts.FavoriteServicesOnlyUserID = permission.UserID(p.Context)
				}
			}

			a, ttl, err := h.c.AlertStore.LegacySearch(p.Context, &opts)
			result.Items = a
			result.Total = ttl

			return newScrubber(p.Context).scrub(result, err)
		},
	}
}

func (h *Handler) searchAlertLogsField() *g.Field {
	return &g.Field{
		Name: "AlertLogs",
		Args: g.FieldConfigArgument{
			"input": &g.ArgumentConfig{
				Type: g.NewInputObject(g.InputObjectConfig{
					Name: "AlertLogSearchOptions",
					Fields: g.InputObjectConfigFieldMap{
						"service_id":         &g.InputObjectFieldConfig{Type: g.String},
						"alert_id":           &g.InputObjectFieldConfig{Type: g.Int},
						"user_id":            &g.InputObjectFieldConfig{Type: g.String},
						"integration_key_id": &g.InputObjectFieldConfig{Type: g.String},
						"start_time":         &g.InputObjectFieldConfig{Type: g.String},
						"end_time":           &g.InputObjectFieldConfig{Type: g.String},
						"event_type":         &g.InputObjectFieldConfig{Type: alertLogEventType},
						"sort_by":            &g.InputObjectFieldConfig{Type: alertLogsSortByEnum},
						"sort_desc":          &g.InputObjectFieldConfig{Type: g.Boolean},
						"limit": &g.InputObjectFieldConfig{
							Type:        g.Int,
							Description: "Defaulted to 25 if not supplied, Maximum is 50.",
						},
						"offset": &g.InputObjectFieldConfig{Type: g.Int},
					},
				}),
			},
		},

		Type: g.NewObject(g.ObjectConfig{
			Name: "AlertLogSearchResult",
			Fields: g.Fields{
				"items":       &g.Field{Type: g.NewList(h.alertLog)},
				"total_count": &g.Field{Type: g.Int},
			},
		}),

		Resolve: func(p g.ResolveParams) (interface{}, error) {
			var list struct {
				Items []alertLogSearchResult `json:"items"`
				Total int                    `json:"total_count"`
			}

			var opts alertlog.LegacySearchOptions
			m, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid input type")
			}

			opts.AlertID, _ = m["alert_id"].(int)
			opts.ServiceID, _ = m["service_id"].(string)
			opts.IntegrationKeyID, _ = m["integration_key_id"].(string)
			opts.UserID, _ = m["user_id"].(string)
			opts.Event, _ = m["event_type"].(alertlog.Type)
			opts.Start, _ = m["start_time"].(time.Time)
			opts.End, _ = m["end_time"].(time.Time)
			opts.SortBy, _ = m["sort_by"].(alertlog.SortBy)
			opts.SortDesc, _ = m["sort_desc"].(bool)
			opts.Limit, _ = m["limit"].(int)
			opts.Offset, _ = m["offset"].(int)

			entries, total, err := h.c.AlertLogStore.LegacySearch(p.Context, &opts)
			if err != nil {
				return newScrubber(p.Context).scrub(nil, err)
			}
			custom := make([]alertLogSearchResult, len(entries))
			for i, e := range entries {
				custom[i].AlertID = e.AlertID()
				custom[i].Log = e.String()
				custom[i].ID = e.ID()
				custom[i].Timestamp = e.Timestamp()
				custom[i].Subject = e.Subject()
				custom[i].Event = e.Type()
			}

			list.Items = custom
			list.Total = total
			return newScrubber(p.Context).scrub(list, err)
		},
	}
}

func (h *Handler) alertSummariesField() *g.Field {
	return &g.Field{
		Name: "AlertSummaries",
		Type: g.NewList(h.alertSummary),
		Resolve: func(p g.ResolveParams) (interface{}, error) {
			return newScrubber(p.Context).scrub(h.c.AlertStore.FindAllSummary(p.Context))
		},
	}
}

var alertLogsSortByEnum = g.NewEnum(g.EnumConfig{
	Name: "AlertLogsSortBy",
	Values: g.EnumValueConfigMap{
		"timestamp":  &g.EnumValueConfig{Value: alertlog.SortByTimestamp},
		"alert_id":   &g.EnumValueConfig{Value: alertlog.SortByAlertID},
		"event_type": &g.EnumValueConfig{Value: alertlog.SortByEventType},
		"user_name":  &g.EnumValueConfig{Value: alertlog.SortByUserName},
	},
})
