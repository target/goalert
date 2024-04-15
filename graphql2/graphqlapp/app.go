package graphqlapp

import (
	context "context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/errcode"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/apollotracing"
	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/alert/alertmetrics"
	"github.com/target/goalert/apikey"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/auth/authlink"
	"github.com/target/goalert/auth/basic"
	"github.com/target/goalert/calsub"
	"github.com/target/goalert/config"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/label"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/notice"
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
	"github.com/target/goalert/swo"
	"github.com/target/goalert/timezone"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/favorite"
	"github.com/target/goalert/user/notificationrule"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type App struct {
	DB                *sql.DB
	AuthBasicStore    *basic.Store
	UserStore         *user.Store
	CMStore           *contactmethod.Store
	NRStore           *notificationrule.Store
	NCStore           *notificationchannel.Store
	AlertStore        *alert.Store
	AlertMetricsStore *alertmetrics.Store
	AlertLogStore     *alertlog.Store
	ServiceStore      *service.Store
	FavoriteStore     *favorite.Store
	PolicyStore       *escalation.Store
	ScheduleStore     *schedule.Store
	CalSubStore       *calsub.Store
	RotationStore     *rotation.Store
	OnCallStore       *oncall.Store
	IntKeyStore       *integrationkey.Store
	LabelStore        *label.Store
	RuleStore         *rule.Store
	OverrideStore     *override.Store
	ConfigStore       *config.Store
	LimitStore        *limit.Store
	SlackStore        *slack.ChannelSender
	HeartbeatStore    *heartbeat.Store
	NoticeStore       *notice.Store
	APIKeyStore       *apikey.Store

	AuthLinkStore *authlink.Store

	NotificationManager *notification.Manager

	AuthHandler *auth.Handler

	NotificationStore *notification.Store
	Twilio            *twilio.Config

	TimeZoneStore *timezone.Store

	SWO *swo.Manager

	FormatDestFunc func(context.Context, notification.DestType, string) string
}

type fieldErr struct {
	FieldName string `json:"fieldName"`
	Message   string `json:"message"`
}

type apolloTracer struct {
	apollotracing.Tracer
	shouldTrace func(context.Context) bool
}

func (a apolloTracer) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	if !a.shouldTrace(ctx) {
		return next(ctx)
	}

	return a.Tracer.InterceptField(ctx, next)
}

func (a apolloTracer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	if !a.shouldTrace(ctx) {
		return next(ctx)
	}

	return a.Tracer.InterceptResponse(ctx, next)
}

func isGQLValidation(gqlErr *gqlerror.Error) bool {
	if gqlErr == nil {
		return false
	}

	var numErr *strconv.NumError
	if errors.As(gqlErr, &numErr) {
		return true
	}

	if strings.HasPrefix(gqlErr.Message, "json request body") {
		var body string
		gqlErr.Message, body, _ = strings.Cut(gqlErr.Message, " body:") // remove body
		if !strings.HasPrefix(strings.TrimSpace(body), "{") {
			// Make the error more readable for common JSON errors.
			gqlErr.Message = "json request body could not be decoded: body must be an object, missing '{'"
		}

		return true
	}

	if gqlErr.Extensions == nil {
		return false
	}

	_, ok := gqlErr.Extensions["code"].(graphql2.ErrorCode)
	if ok {
		return true
	}

	code, ok := gqlErr.Extensions["code"].(string)
	if !ok {
		return false
	}

	switch code {
	case errcode.ValidationFailed, errcode.ParseFailed:
		// These are gqlgen validation errors.
		return true
	}

	return false
}

func (a *App) Handler() http.Handler {
	h := handler.NewDefaultServer(
		graphql2.NewExecutableSchema(graphql2.Config{
			Resolvers: a,
			Directives: graphql2.DirectiveRoot{
				Experimental: Experimental,
			},
		}),
	)

	type hasTraceKey int
	h.Use(apolloTracer{Tracer: apollotracing.Tracer{}, shouldTrace: func(ctx context.Context) bool {
		enabled, ok := ctx.Value(hasTraceKey(1)).(bool)
		return ok && enabled
	}})

	h.Use(apikey.Middleware{})

	h.AroundFields(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
		defer func() {
			err := recover()
			if err != nil {
				panic(err)
			}
		}()
		fieldCtx := graphql.GetFieldContext(ctx)

		start := time.Now()
		res, err = next(ctx)
		errVal := "0"
		if err != nil {
			errVal = "1"
		}
		if fieldCtx.IsMethod {
			metricResolverHist.
				WithLabelValues(fmt.Sprintf("%s.%s", fieldCtx.Object, fieldCtx.Field.Name), errVal).
				Observe(time.Since(start).Seconds())
		}
		if err == nil && fieldCtx.Object == "Mutation" {
			ctx = log.WithFields(ctx, log.Fields{
				"MutationName": fieldCtx.Field.Name,
			})
			log.Logf(ctx, "Mutation.")
		}

		return res, err
	})

	h.Use(&errSkipHandler{})

	h.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		if errors.Is(err, errAlreadySet) {
			// This error just indicates that a field error has already been set on the context
			// it should not be returned to the client.
			return &gqlerror.Error{
				Extensions: map[string]interface{}{
					"skip": true,
				},
			}
		}
		err = errutil.MapDBError(err)
		var gqlErr *gqlerror.Error

		isUnsafe, safeErr := errutil.ScrubError(err)
		if !errors.As(err, &gqlErr) {
			gqlErr = &gqlerror.Error{
				Message: safeErr.Error(),
			}
		}

		if isUnsafe && !isGQLValidation(gqlErr) {
			// context.Canceled is caused by normal things like closing a browser tab.
			if !errors.Is(err, context.Canceled) {
				log.Log(ctx, err)
			}
			gqlErr.Message = safeErr.Error()
		}

		var multiFieldErr validation.MultiFieldError
		var singleFieldErr validation.FieldError
		if errors.As(err, &multiFieldErr) {
			errs := make([]fieldErr, len(multiFieldErr.FieldErrors()))
			for i, err := range multiFieldErr.FieldErrors() {
				errs[i].FieldName = err.Field()
				errs[i].Message = err.Reason()
			}
			gqlErr.Message = "Multiple fields failed validation."
			gqlErr.Extensions = map[string]interface{}{
				"isMultiFieldError": true,
				"fieldErrors":       errs,
			}
		} else if errors.As(err, &singleFieldErr) {
			type reasonable interface {
				Reason() string
			}
			msg := singleFieldErr.Error()
			if rs, ok := singleFieldErr.(reasonable); ok {
				msg = rs.Reason()
			}
			gqlErr.Message = msg
			gqlErr.Extensions = map[string]interface{}{
				"fieldName":    singleFieldErr.Field(),
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
		defer a.closeLoaders(ctx)

		if req.URL.Query().Get("trace") == "1" && permission.Admin(ctx) {
			ctx = context.WithValue(ctx, hasTraceKey(1), true)
		}

		h.ServeHTTP(w, req.WithContext(ctx))
	})
}
