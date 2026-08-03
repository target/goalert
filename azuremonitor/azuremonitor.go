// Package azuremonitor implements ingress for Azure Monitor alerts delivered by
// an action-group webhook receiver.
//
// # Trust model
//
// Azure signs nothing. Unlike the sibling cloudwatch integration -- where SNS
// signs every message with a verifiable certificate chain -- the integration key
// in the query string is the only credential, so the webhook URL is
// credential-grade: anyone holding it can create arbitrary alerts on the service
// the key identifies. There is no public key to verify against, so asymmetric
// verification is not possible; anything built would be a shared secret with
// extra steps.
//
// The handler makes no outbound requests at all, so there is no SSRF surface and
// nothing analogous to cloudwatch's host allowlist or certificate cache.
//
// # Service resolution
//
// The integration key selects the GoAlert service. Nothing here knows or can know
// which action group, subscription, or team it is serving -- the only inputs are
// the request body and the key. That is what lets one deployed handler serve any
// number of action groups with no routing table and no per-group code.
package azuremonitor

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
)

// maxBodyBytes bounds the request body independently of the global
// maxBodySizeMiddleware, which is disabled when MaxReqBodyBytes is 0.
const maxBodyBytes = 256 * 1024

// Config configures a Handler.
type Config struct {
	AlertStore          *alert.Store
	IntegrationKeyStore *integrationkey.Store
}

// Handler serves the Azure Monitor ingress endpoint.
type Handler struct {
	cfg Config
}

// NewHandler returns a Handler for the given config.
func NewHandler(cfg Config) *Handler { return &Handler{cfg: cfg} }

// ServeIncoming handles an Azure Monitor action-group webhook delivery.
func (h *Handler) ServeIncoming(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := permission.LimitCheckAny(ctx, permission.Service)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	// MaxBytesReader rather than io.LimitReader: errutil.HTTPError maps
	// *http.MaxBytesError to a clean 413, whereas silent truncation would surface
	// as a confusing 400.
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	data, err := io.ReadAll(r.Body)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	a, meta, info, err := buildAlert(data)
	switch {
	case errors.Is(err, errLegacySchema):
		// Deliberately not degraded to the best-effort path: a legacy-schema
		// payload would yield alerts with no useful content and no indication
		// why. 400 rather than 5xx so Azure does not retry a receiver that is
		// permanently misconfigured.
		log.Logf(ctx, "azuremonitor: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		log.Debugf(ctx, "azuremonitor: bad request body: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// signalType alone does not identify the payload shape -- Platform and
	// Prometheus metric alerts share it, as do Log Alerts V2 and Azure Backup --
	// so log the whole triple that actually determines parsing.
	ctx = log.WithFields(ctx, log.Fields{
		"AzureSignalType":       info.SignalType,
		"AzureMonitorService":   info.MonitorService,
		"AzureConditionType":    info.ConditionType,
		"AzureMonitorCondition": meta["monitor_condition"],
	})

	// Nothing recognised alertContext, so this alert carries essentials only. That
	// is a deliberate graceful degradation rather than a failure, but it is also
	// how a newly-routed Azure alert type announces itself -- otherwise the only
	// symptom is thin alerts that nobody notices.
	if !info.ContextRendered {
		log.Logf(ctx, "azuremonitor: unrecognised alertContext, alert built from essentials only")
	}

	a.ServiceID = permission.ServiceID(ctx)
	if len(meta) == 0 {
		meta = nil
	}

	var created *alert.Alert
	err = retry.DoTemporaryError(func(int) error {
		var err error
		created, _, err = h.cfg.AlertStore.CreateOrUpdateWithMeta(ctx, &a, meta)
		return err
	},
		retry.Log(ctx),
		retry.Limit(5),
		retry.FibBackoff(250*time.Millisecond),
	)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	// created is nil with a nil error when the status was closed and no open alert
	// held this dedup key -- a Resolved delivery for an alert we never opened, or
	// already closed. Normal, so info level and a 2xx.
	if created == nil {
		log.Logf(ctx, "azuremonitor: no open alert to close")
	}

	w.WriteHeader(http.StatusNoContent)
}
