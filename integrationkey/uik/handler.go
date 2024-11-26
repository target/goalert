package uik

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/expr-lang/expr/vm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/engine/signalmgr"
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

	r *river.Client[pgx.Tx]
}

func NewHandler(db gadb.DBTX, intStore *integrationkey.Store, aStore *alert.Store, r *river.Client[pgx.Tx]) *Handler {
	return &Handler{intStore: intStore, db: db, alertStore: aStore, r: r}
}

func (h *Handler) handleAction(ctx context.Context, act gadb.UIKActionV1) (inserted bool, err error) {
	var didInsertSignals bool
	switch act.Dest.Type {
	case "builtin-webhook":
		req, err := http.NewRequest("POST", act.Dest.Arg("webhook_url"), strings.NewReader(act.Param("body")))
		if err != nil {
			return false, err
		}
		req.Header.Set("Content-Type", act.Param("content-type"))

		_, err = http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return false, err
		}

	case "builtin-alert":
		status := alert.StatusTriggered
		if act.Param("close") == "true" {
			status = alert.StatusClosed
		}

		_, _, err := h.alertStore.CreateOrUpdate(ctx, &alert.Alert{
			ServiceID: permission.ServiceID(ctx),
			Summary:   act.Param("summary"),
			Details:   act.Param("details"),
			Source:    alert.SourceUniversal,
			Status:    status,
		})
		if err != nil {
			return false, err
		}
	default:
		data, err := json.Marshal(act.Params)
		if err != nil {
			return false, err
		}

		err = gadb.New(h.db).IntKeyInsertSignalMessage(ctx, gadb.IntKeyInsertSignalMessageParams{
			DestID:    act.ChannelID,
			ServiceID: permission.ServiceNullUUID(ctx).UUID,
			Params:    data,
		})
		if err != nil {
			return false, err
		}
		didInsertSignals = true
	}

	return didInsertSignals, nil
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

	// TODO: cache
	compiled, err := NewCompiledConfig(*cfg)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	q := req.URL.Query()
	query := make(map[string]string)
	for key := range q {
		query[key] = q.Get(key)
	}
	querya := map[string][]string(q)
	env := map[string]any{
		"sprintf": fmt.Sprintf,
		"req": map[string]any{
			"body":   body,
			"query":  query,
			"querya": querya,
			"ua":     req.UserAgent(),
			"ip":     req.RemoteAddr,
		},
	}

	var vm vm.VM
	actions, err := compiled.Run(&vm, env)
	if errutil.HTTPError(ctx, w, validation.WrapError(err)) {
		return
	}

	var insertedAny bool
	for _, act := range actions {
		inserted, err := h.handleAction(ctx, act)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
		insertedAny = insertedAny || inserted
	}

	if insertedAny {
		// schedule job
		err := signalmgr.TriggerService(ctx, h.r, permission.ServiceNullUUID(ctx).UUID)
		if err != nil {
			log.Log(ctx, fmt.Errorf("schedule signal message: %w", err))
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
