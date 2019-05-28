package graphql

import (
	g "github.com/graphql-go/graphql"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

func wrapFieldResolver(fn g.FieldResolveFn) g.FieldResolveFn {
	return func(p g.ResolveParams) (interface{}, error) {
		ctx, span := trace.StartSpan(p.Context, "GraphQL."+p.Info.ParentType.Name()+"."+p.Info.FieldName)
		defer span.End()
		p.Context = ctx
		val, err := fn(p)
		if err != nil {
			span.Annotate([]trace.Attribute{trace.BoolAttribute("error", true)}, err.Error())
		}
		return val, err
	}
}

func (h *Handler) buildSchema() error {
	queryFields := g.Fields{
		"currentUser":        h.currentUserField(),
		"user":               h.userField(),
		"users":              h.usersField(),
		"alert":              h.alertField(),
		"alerts":             h.alertsField(),
		"alerts2":            h.searchAlertsField(),
		"alertSummaries":     h.alertSummariesField(),
		"service":            h.serviceField(),
		"services":           h.servicesField(),
		"services2":          h.searchServicesField(),
		"schedules":          h.schedulesField(),
		"rotations":          h.rotationsField(),
		"schedule":           h.scheduleField(),
		"rotation":           h.rotationField(),
		"escalationPolicy":   h.escalationPolicyField(),
		"escalationPolicies": h.escalationPoliciesField(),
		"integrationKey":     h.integrationKeyField(),
		"integrationKeys":    h.integrationKeysField(),
		"alertLogs":          h.searchAlertLogsField(),
		"labelKeys":          h.labelKeysField(),
	}

	for _, f := range queryFields {
		f.Resolve = wrapFieldResolver(f.Resolve)
	}

	rootQuery := g.ObjectConfig{Name: "RootQuery", Fields: queryFields}
	mutFields := g.Fields{
		"updateUser":                         h.updateUserField(),
		"deleteSchedule":                     h.deleteScheduleField(),
		"createSchedule":                     h.createScheduleField(),
		"updateSchedule":                     h.updateSchedule(),
		"createAlert":                        h.createAlertField(),
		"updateAlertStatus":                  h.updateStatusAlertField(),
		"updateAlertStatusByService":         h.updateAlertStatusByServiceField(),
		"escalateAlert":                      h.escalateAlertField(),
		"updateNotificationRule":             h.updateNotificationRuleField(),
		"createNotificationRule":             h.createNotificationRuleField(),
		"deleteNotificationRule":             h.deleteNotificationRuleField(),
		"updateContactMethod":                h.updateContactMethodField(),
		"createContactMethod":                h.createContactMethodField(),
		"deleteContactMethod":                h.deleteContactMethodField(),
		"addRotationParticipant":             h.addRotationParticipantField(),
		"deleteRotationParticipant":          h.deleteRotationParticipantField(),
		"moveRotationParticipant":            h.moveRotationParticipantField(),
		"setActiveParticipant":               h.setActiveParticipantField(),
		"createOrUpdateEscalationPolicyStep": h.createOrUpdateEscalationPolicyStepField(),
		"addEscalationPolicyStepTarget":      h.addEscalationPolicyStepTargetField(),
		"deleteEscalationPolicyStepTarget":   h.deleteEscalationPolicyStepTargetField(),
		"deleteEscalationPolicy":             h.deleteEscalationPolicyField(),
		"deleteEscalationPolicyStep":         h.deleteEscalationPolicyStepField(),
		"moveEscalationPolicyStep":           h.moveEscalationPolicyStepField(),
		"createOrUpdateEscalationPolicy":     h.createOrUpdateEscalationPolicyField(),
		"createService":                      h.createServiceField(),
		"updateService":                      h.updateServiceField(),
		"deleteService":                      h.deleteServiceField(),
		"createIntegrationKey":               h.createIntegrationKeyField(),
		"deleteIntegrationKey":               h.deleteIntegrationKeyField(),
		"createOrUpdateRotation":             h.createOrUpdateRotationField(),
		"deleteScheduleRule":                 h.deleteScheduleRuleField(),
		"deleteScheduleAssignment":           h.deleteScheduleAssignmentField(),
		"updateScheduleRule":                 h.updateScheduleRuleField(),
		"createScheduleRule":                 h.createScheduleRuleField(),
		"addRotationParticipant2":            h.addRotationParticipant2Field(),
		"deleteRotationParticipant2":         h.deleteRotationParticipant2Field(),
		"moveRotationParticipant2":           h.moveRotationParticipant2Field(),
		"deleteRotation":                     h.deleteRotationField(),
		"createAll":                          h.createAllField(),
		"updateConfigLimit":                  h.updateConfigLimitField(),
		"sendContactMethodTest":              h.sendContactMethodTest(),
		"sendContactMethodVerification":      h.sendContactMethodVerification(),
		"verifyContactMethod":                h.verifyContactMethod(),
		"deleteHeartbeatMonitor":             h.deleteHeartbeatMonitorField(),
		"updateUserOverride":                 h.updateUserOverrideField(),
		"deleteAll":                          h.deleteAllField(),
		"setUserFavorite":                    h.setUserFavoriteField(),
		"unsetUserFavorite":                  h.unsetUserFavoriteField(),
		"setLabel":                           h.setLabelField(),
	}
	for _, f := range mutFields {
		f.Resolve = wrapFieldResolver(f.Resolve)
	}
	rootMutation := g.ObjectConfig{Name: "RootMutation", Fields: mutFields}

	schemaConfig := g.SchemaConfig{
		Query:    g.NewObject(rootQuery),
		Mutation: g.NewObject(rootMutation),
	}
	schema, err := g.NewSchema(schemaConfig)
	if err != nil {
		return errors.Wrap(err, "generate GraphQL schema")
	}

	h.schema = schema
	return nil
}
