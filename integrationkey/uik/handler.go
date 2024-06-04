package uik

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/google/uuid"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
)

type Handler struct {
	intStore   *integrationkey.Store
	alertStore *alert.Store
	db         gadb.DBTX
}

func NewHandler(db gadb.DBTX, intStore *integrationkey.Store, aStore *alert.Store) *Handler {
	return &Handler{intStore: intStore, db: db, alertStore: aStore}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	if !expflag.ContextHas(ctx, expflag.UnivKeys) {
		errutil.HTTPError(ctx, w, validation.NewGenericError("universal keys are disabled"))
		return
	}

	err := permission.LimitCheckAny(req.Context(), permission.Service)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	src := permission.Source(ctx)
	if src.Type != permission.SourceTypeUIK {
		// we don't want to allow regular API keys to be used here
		errutil.HTTPError(ctx, w, permission.Unauthorized())
		return
	}

	keyID, err := uuid.Parse(src.ID)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	data, err := io.ReadAll(req.Body)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	var body any
	err = json.Unmarshal(data, &body)
	if errutil.HTTPError(ctx, w, validation.WrapError(err)) {
		return
	}

	cfg, err := h.intStore.Config(ctx, h.db, keyID)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	env := map[string]any{
		"sprintf": fmt.Sprintf,
		"req": map[string]any{
			"body": body,
		},
	}

	// We need to track if any rule matched, so we can apply default actions if none did.
	var anyMatched bool
	var results []integrationkey.ActionResult
	for _, rule := range cfg.Rules {
		result, err := expr.Eval("string("+rule.ConditionExpr+")", env)
		if errutil.HTTPError(ctx, w, validation.WrapError(err)) {
			return
		}
		r, ok := result.(string)
		if !ok {
			errutil.HTTPError(ctx, w, validation.NewGenericError("condition expression must return a boolean"))
			return
		}
		anyMatched = anyMatched || r == "true"
		if r != "true" {
			continue
		}

		for _, action := range rule.Actions {
			res := integrationkey.ActionResult{
				DestType: action.Type,
				Values:   action.StaticParams,
				Params:   make(map[string]string, len(action.DynamicParams)),
			}

			for name, exprStr := range action.DynamicParams {
				val, err := expr.Eval("string("+exprStr+")", env)
				if errutil.HTTPError(ctx, w, validation.WrapError(err)) {
					return
				}
				if _, ok := val.(string); !ok {
					errutil.HTTPError(ctx, w, validation.NewGenericError("dynamic param expressions must return a string"))
					return
				}
				res.Params[name] = val.(string)
			}
			results = append(results, res)
		}
	}

	if !anyMatched {
		// Default actions need to be applied if no rules matched (or if there are no rules at all).
		for _, action := range cfg.DefaultActions {
			res := integrationkey.ActionResult{
				DestType: action.Type,
				Values:   action.StaticParams,
				Params:   make(map[string]string, len(action.DynamicParams)),
			}

			for name, exprStr := range action.DynamicParams {
				val, err := expr.Eval("string("+exprStr+")", env)
				if errutil.HTTPError(ctx, w, validation.WrapError(err)) {
					return
				}
				if _, ok := val.(string); !ok {
					errutil.HTTPError(ctx, w, validation.NewGenericError("dynamic param expressions must return a string"))
					return
				}
				res.Params[name] = val.(string)
			}
			results = append(results, res)
		}
	}

	log.Logf(ctx, "uik: action result: %#v", results)

	for _, res := range results {
		switch res.DestType {
		case "builtin-webhook":
			req, err := http.NewRequest("POST", res.Values["webhook-url"], strings.NewReader(res.Params["body"]))
			if errutil.HTTPError(ctx, w, err) {
				return
			}
			req.Header.Set("Content-Type", res.Params["content-type"])

			_, err = http.DefaultClient.Do(req.WithContext(ctx))
			if errutil.HTTPError(ctx, w, err) {
				return
			}

		case "builtin-alert":
			status := alert.StatusTriggered
			if res.Params["close"] == "true" {
				status = alert.StatusClosed
			}

			_, _, err = h.alertStore.CreateOrUpdate(ctx, &alert.Alert{
				ServiceID: permission.ServiceID(ctx),
				Summary:   res.Params["summary"],
				Details:   res.Params["details"],
				Source:    alert.SourceUniversal,
				Status:    status,
			})
			if errutil.HTTPError(ctx, w, err) {
				return
			}
		default:
			errutil.HTTPError(ctx, w, validation.NewFieldError("action", "unknown action type"))
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
