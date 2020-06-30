package graphqlapp

import (
	context "context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/calendarsubscription"
	"github.com/target/goalert/config"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/label"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/override"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/service"
	"github.com/target/goalert/timezone"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/favorite"
	"github.com/target/goalert/user/notificationrule"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.opencensus.io/trace"
)

type App struct {
	DB             *sql.DB
	UserStore      user.Store
	CMStore        contactmethod.Store
	NRStore        notificationrule.Store
	NCStore        notificationchannel.Store
	AlertStore     alert.Store
	AlertLogStore  alertlog.Store
	ServiceStore   service.Store
	FavoriteStore  favorite.Store
	PolicyStore    escalation.Store
	ScheduleStore  schedule.Store
	CalSubStore    *calendarsubscription.Store
	RotationStore  rotation.Store
	OnCallStore    oncall.Store
	IntKeyStore    integrationkey.Store
	LabelStore     label.Store
	RuleStore      rule.Store
	OverrideStore  override.Store
	ConfigStore    *config.Store
	LimitStore     *limit.Store
	SlackStore     *slack.ChannelSender
	HeartbeatStore heartbeat.Store

	NotificationStore notification.Store
	Twilio            *twilio.Config

	TimeZoneStore *timezone.Store
}

func mustAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := permission.LimitCheckAny(req.Context())
		if errutil.HTTPError(req.Context(), w, err) {
			return
		}

		h.ServeHTTP(w, req)
	})
}

func (a *App) PlayHandler() http.Handler {
	var data struct {
		Version string
	}
	data.Version = playVersion
	return mustAuth(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := playTmpl.Execute(w, data)
		if err != nil {
			log.Log(req.Context(), err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}))
}

type fieldErr struct {
	FieldName string `json:"fieldName"`
	Message   string `json:"message"`
}

type apolloTracer struct {
	apollotracing.Tracer
}

func (a apolloTracer) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	if !permission.Admin(ctx) {
		return next(ctx)
	}

	return a.Tracer.InterceptField(ctx, next)
}
func (a apolloTracer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	if !permission.Admin(ctx) {
		return next(ctx)
	}

	return a.Tracer.InterceptResponse(ctx, next)
}

func (a *App) Handler() http.Handler {
	h := handler.NewDefaultServer(
		graphql2.NewExecutableSchema(graphql2.Config{Resolvers: a}),
	)
	h.Use(apolloTracer{Tracer: apollotracing.Tracer{}})

	h.AroundFields(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
		defer func() {
			err := recover()
			if err != nil {
				panic(err)
			}
		}()
		fieldCtx := graphql.GetFieldContext(ctx)

		ctx, sp := trace.StartSpan(ctx, "GQL."+fieldCtx.Object+"."+fieldCtx.Field.Name, trace.WithSpanKind(trace.SpanKindServer))
		defer sp.End()
		sp.AddAttributes(
			trace.StringAttribute("graphql.object", fieldCtx.Object),
			trace.StringAttribute("graphql.field.name", fieldCtx.Field.Name),
		)
		res, err = next(ctx)
		if err != nil {
			sp.Annotate([]trace.Attribute{
				trace.BoolAttribute("error", true),
			}, err.Error())
		} else if fieldCtx.Object == "Mutation" {
			ctx = log.WithFields(ctx, log.Fields{
				"MutationName": fieldCtx.Field.Name,
			})
			log.Logf(ctx, "Mutation.")
		}

		return res, err
	})

	h.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		if e, ok := err.(*strconv.NumError); ok {
			// gqlgen doesn't handle exponent notation numbers properly
			// but we want to return a validation error instead of a 500 at least.
			err = validation.NewGenericError("parse '" + e.Num + "': " + e.Err.Error())
		}
		err = errutil.MapDBError(err)
		isUnsafe, safeErr := errutil.ScrubError(err)
		if isUnsafe {
			log.Log(ctx, err)
		}
		gqlErr := graphql.DefaultErrorPresenter(ctx, safeErr)

		if m, ok := errors.Cause(safeErr).(validation.MultiFieldError); ok {
			errs := make([]fieldErr, len(m.FieldErrors()))
			for i, err := range m.FieldErrors() {
				errs[i].FieldName = err.Field()
				errs[i].Message = err.Reason()
			}
			gqlErr.Message = "Multiple fields failed validation."
			gqlErr.Extensions = map[string]interface{}{
				"isMultiFieldError": true,
				"fieldErrors":       errs,
			}
		} else if e, ok := errors.Cause(safeErr).(validation.FieldError); ok {
			type reasonable interface {
				Reason() string
			}
			msg := e.Error()
			if rs, ok := e.(reasonable); ok {
				msg = rs.Reason()
			}
			gqlErr.Message = msg
			gqlErr.Extensions = map[string]interface{}{
				"fieldName":    e.Field(),
				"isFieldError": true,
			}
		}

		return gqlErr
	})

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		// ensure some sort of auth before continuing
		err := permission.LimitCheckAny(ctx)
		if errutil.HTTPError(ctx, w, err) {
			return
		}

		ctx = a.registerLoaders(ctx)

		h.ServeHTTP(w, req.WithContext(ctx))
	})
}
