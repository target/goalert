package graphql

import (
	"context"
	"encoding/json"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/shiftcalc"
	"github.com/target/goalert/service"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"io/ioutil"
	"net/http"
	"sort"

	g "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/pkg/errors"
)

type requestInfoKey int

// RequestInfoContextKey is used to store RequestInfo in a context.Context.
const RequestInfoContextKey = requestInfoKey(0)

// RequestInfo carries useful information about the request.
type RequestInfo struct {
	Mutations []string
	Queries   []string
}

type Handler struct {
	c Config

	alert                *g.Object
	alertSummary         *g.Object
	alertLog             *g.Object
	alertLogSubject      *g.Object
	user                 *g.Object
	service              *g.Object
	contactMethod        *g.Object
	notificationRule     *g.Object
	schedule             *g.Object
	rotation             *g.Object
	rotationParticipant  *g.Object
	escalationPolicyStep *g.Object
	escalationPolicy     *g.Object
	integrationKey       *g.Object
	scheduleRule         *g.Object
	scheduleAssignment   *g.Object
	scheduleShift        *g.Object
	onCallAssignment     *g.Object
	createAll            *g.Object
	heartbeat            *g.Object
	userOverride         *g.Object
	deleteAll            *g.Object
	serviceOnCallUser    *g.Object

	rotationShift *g.Object

	sourceType       *g.Union
	targetType       *g.Union
	assignmentSource *g.Object
	assignmentTarget *g.Object
	label            *g.Object

	legacyDB *legacyDB

	shiftCalc *shiftcalc.ShiftCalculator
	// resolver  resolver.ResolveWalker

	schema g.Schema
}

type fieldConfigAdder interface {
	AddFieldConfig(string, *g.Field)
}

func addFields(o fieldConfigAdder, f g.Fields) {
	for n, f := range f {
		if f.Resolve != nil {
			f.Resolve = wrapFieldResolver(f.Resolve)
		}
		o.AddFieldConfig(n, f)
	}
}

func NewHandler(ctx context.Context, c Config) (*Handler, error) {

	obj := func(name, desc string, ifaces ...*g.Interface) *g.Object {
		return g.NewObject(g.ObjectConfig{
			Name:        name,
			Description: desc,
			Fields:      g.Fields{},
			Interfaces:  ifaces,
		})
	}
	c = cachedConfig(c)

	db, err := newLegacyDB(ctx, c.DB)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		c: c,
		shiftCalc: &shiftcalc.ShiftCalculator{
			RotStore:   c.RotationStore,
			RuleStore:  c.ScheduleRuleStore,
			SchedStore: c.ScheduleStore,
		},

		legacyDB: db,

		alert:                obj("Alert", "An alert."),
		alertLogSubject:      obj("AlertLogSubject", "The entity associated with the log event (e.g. the user who closed the alert)."),
		user:                 obj("User", "A user."),
		service:              obj("Service", "A registered service."),
		alertLog:             obj("AlertLog", "A log entry for Alert activity."),
		contactMethod:        obj("ContactMethod", "A method of notifying (contacting) a User."),
		notificationRule:     obj("NotificationRule", "A rule controlling how/when to use a ContactMethod to notifiy a User."),
		schedule:             obj("Schedule", "An on-call schedule."),
		rotation:             obj("Rotation", "An on-call rotation."),
		rotationParticipant:  obj("RotationParticipant", "A participant in an on-call rotation."),
		escalationPolicyStep: obj("EscalationPolicyStep", "A single step of an escalation policy."),
		escalationPolicy:     obj("EscalationPolicy", "An escalation policy."),
		integrationKey:       obj("IntegrationKey", "An Integration."),
		assignmentSource:     obj("AssignmentSource", "The source of an assignment."),
		assignmentTarget:     obj("AssignmentTarget", "The target of an assignment."),
		scheduleRule:         obj("ScheduleRule", "A schedule rule."),
		scheduleAssignment:   obj("ScheduleAssignment", "A schedule assignment"),
		scheduleShift:        obj("ScheduleShift", "A single shift of a schedule."),
		onCallAssignment:     obj("OnCallAssignment", "Contains an assignment for a user oncall."),
		createAll:            obj("CreateAll", "Creates multiple resources at once."),
		alertSummary:         obj("AlertSummary", "Contains alert totals for a service."),
		heartbeat:            obj("Heartbeat", "A heartbeat check for a service."),
		userOverride:         obj("UserOverride", "An override event to add, remove, or swap a user."),
		deleteAll:            obj("DeleteAll", "Deletes multiple resources at once."),
		label:                obj("Label", "Labels are key/value pairs that are attached to objects, such as services."),
		rotationShift:        obj("RotationShift2", "A single shift of a rotation."),
		serviceOnCallUser:    obj("ServiceOnCallUser", "An on-call user assigned to a service."),
	}

	h.sourceType = g.NewUnion(g.UnionConfig{
		Name:        "SourceType",
		Description: "Source object type of an assignment.",
		Types:       []*g.Object{h.alert, h.user, h.service, h.rotationParticipant, h.escalationPolicyStep},
		ResolveType: func(p g.ResolveTypeParams) *g.Object {
			src, ok := p.Value.(assignment.Source)
			if !ok {
				return nil
			}
			switch src.SourceType() {
			case assignment.SrcTypeAlert:
				return h.alert
			case assignment.SrcTypeEscalationPolicyStep:
				return h.escalationPolicyStep
			case assignment.SrcTypeRotationParticipant:
				return h.rotationParticipant
			case assignment.SrcTypeScheduleRule:
				return h.scheduleRule
			case assignment.SrcTypeService:
				return h.service
			case assignment.SrcTypeUser:
				return h.user
			}
			return nil
		},
	})

	h.targetType = g.NewUnion(g.UnionConfig{
		Name:        "TargetType",
		Description: "Target object type of an assignment.",
		Types:       []*g.Object{h.user, h.service, h.schedule, h.rotation, h.escalationPolicy},
		ResolveType: func(p g.ResolveTypeParams) *g.Object {
			switch p.Value.(type) {
			case user.User, *user.User:
				return h.user
			case service.Service, *service.Service:
				return h.service
			case schedule.Schedule, *schedule.Schedule:
				return h.schedule
			case rotation.Rotation, *rotation.Rotation:
				return h.rotation
			case escalation.Policy, *escalation.Policy:
				return h.escalationPolicy
			}

			return nil
		},
	})

	addFields(h.user, h.userFields())
	addFields(h.alertLogSubject, h.alertLogSubjectFields())
	addFields(h.alert, h.alertFields())
	addFields(h.alertLog, h.alertLogFields())
	addFields(h.contactMethod, h.CMFields())
	addFields(h.notificationRule, h.NRFields())
	addFields(h.service, h.serviceFields())
	addFields(h.schedule, h.scheduleFields())
	addFields(h.rotation, h.rotationFields())
	addFields(h.rotationParticipant, h.rotationParticipantFields())
	addFields(h.escalationPolicyStep, h.escalationPolicyStepFields())
	addFields(h.escalationPolicy, h.escalationPolicyFields())
	addFields(h.integrationKey, h.integrationKeyFields())
	addFields(h.assignmentSource, h.assignmentSourceFields())
	addFields(h.assignmentTarget, h.assignmentTargetFields())
	addFields(h.scheduleRule, h.scheduleRuleFields())
	addFields(h.scheduleAssignment, h.scheduleAssignmentFields())
	addFields(h.scheduleShift, h.scheduleShiftFields())
	addFields(h.onCallAssignment, h.onCallAssignmentFields())
	addFields(h.createAll, h.createAllFields())
	addFields(h.alertSummary, h.alertSummaryFields())
	addFields(h.heartbeat, h.heartbeatMonitorFields())
	addFields(h.userOverride, h.userOverrideFields())
	addFields(h.deleteAll, h.deleteAllFields())
	addFields(h.label, h.labelFields())
	addFields(h.rotationShift, h.rotationShiftFields())
	addFields(h.serviceOnCallUser, h.serviceOnCallUserFields())

	err = h.buildSchema()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Debug(ctx, errors.Wrap(err, "read GraphQL query"))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var b struct {
		Query     string
		Variables map[string]interface{}
	}

	err = json.Unmarshal(data, &b)
	if err != nil {
		log.Debug(ctx, errors.Wrap(err, "parse GraphQL query"))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var r *g.Result
	params := g.Params{
		Context:        ctx,
		Schema:         h.schema,
		RequestString:  b.Query,
		VariableValues: b.Variables,
	}

	if info, ok := ctx.Value(RequestInfoContextKey).(*RequestInfo); ok && info != nil {
		// If we have access to the RequestInfo pointer, we parse the query and try to gleam some
		// useful info to store.
		//
		// If the request info is missing (e.g. a future option to disable it) we just gracefully
		// ignore it and move on as before.
		source := source.NewSource(&source.Source{
			Body: []byte(b.Query),
			Name: "GraphQL Request",
		})
		a, err := parser.Parse(parser.ParseParams{Source: source})
		if err != nil {
			log.Debug(ctx, errors.Wrap(err, "parse GraphQL query"))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Safely walk the graphql AST by whatever means necessary...
		for _, n := range a.Definitions {
			switch def := n.(type) {
			case *ast.OperationDefinition:
				for _, sel := range def.GetSelectionSet().Selections {
					f, ok := sel.(*ast.Field)
					if !ok {
						continue
					}

					switch def.GetOperation() {
					case "query":
						info.Queries = append(info.Queries, f.Name.Value)
					case "mutation":
						info.Mutations = append(info.Mutations, f.Name.Value)
					}
				}
			}
		}

		// Sorted things are appreciated in logs.
		sort.Strings(info.Queries)
		sort.Strings(info.Mutations)
	}

	r = g.Do(params)
	err = json.NewEncoder(w).Encode(r)
	if errutil.HTTPError(ctx, w, errors.Wrap(err, "serialize GraphQL response")) {
		return
	}
}
