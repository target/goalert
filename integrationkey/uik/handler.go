package uik

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/expr-lang/expr/vm"
	"github.com/google/uuid"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/validation"
)

type Handler struct {
	intStore   *integrationkey.Store
	alertStore *alert.Store
	db         TxAble
}

type TxAble interface {
	gadb.DBTX
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

func NewHandler(db TxAble, intStore *integrationkey.Store, aStore *alert.Store) *Handler {
	return &Handler{intStore: intStore, db: db, alertStore: aStore}
}

func (h *Handler) handleAction(ctx context.Context, act integrationkey.Action, _params any) error {
	params := _params.(map[string]any)

	switch act.Type {
	case "builtin-webhook":
		req, err := http.NewRequest("POST", act.StaticParams["webhook-url"], strings.NewReader(params["body"].(string)))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", params["content-type"].(string))

		_, err = http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return err
		}

	case "builtin-alert":
		status := alert.StatusTriggered
		if params["close"] == "true" {
			status = alert.StatusClosed
		}

		_, _, err := h.alertStore.CreateOrUpdate(ctx, &alert.Alert{
			ServiceID: permission.ServiceID(ctx),
			Summary:   params["summary"].(string),
			Details:   params["details"].(string),
			Source:    alert.SourceUniversal,
			Status:    status,
		})
		if err != nil {
			return err
		}
	default:
		data, err := json.Marshal(params)
		if err != nil {
			return err
		}

		err = gadb.New(h.db).IntKeyInsertSignalMessage(ctx, gadb.IntKeyInsertSignalMessageParams{
			DestID:    act.ChannelID,
			ServiceID: permission.ServiceNullUUID(ctx).UUID,
			Params:    data,
		})
		if err != nil {
			return err
		}
	}

	return nil
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

	var vm vm.VM
	var matched bool
	for _, rule := range cfg.Rules {
		p, err := CompileRule(rule.ConditionExpr, rule.Actions)
		if errutil.HTTPError(ctx, w, err) {
			return
		}

		res, err := vm.Run(p, env)
		if errutil.HTTPError(ctx, w, err) {
			return
		}

		actions, ok := res.([]any)
		if !ok {
			// didn't match
			continue
		}
		matched = true
		if len(actions) != len(rule.Actions) {
			// This should never happen, but better than a panic.
			errutil.HTTPError(ctx, w, fmt.Errorf("rule %s: expected %d actions, got %d", rule.ID, len(rule.Actions), len(actions)))
			return
		}

		for i, act := range actions {
			err = h.handleAction(ctx, rule.Actions[i], act)
			if errutil.HTTPError(ctx, w, err) {
				return
			}
		}

		if rule.ContinueAfterMatch {
			continue
		}

		break
	}

	if !matched {
		p, err := CompileRule("", cfg.DefaultActions)
		if errutil.HTTPError(ctx, w, err) {
			return
		}

		res, err := vm.Run(p, env)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
		for i, act := range res.([]any) {
			err = h.handleAction(ctx, cfg.DefaultActions[i], act)
			if errutil.HTTPError(ctx, w, err) {
				return
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
