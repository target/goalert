package azuremonitor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/validation/validate"
)

// commonAlertSchemaID is the only schema this integration parses. Azure's legacy
// schema is a different, flatter shape that varies by alert type; a receiver
// sending it is rejected with an actionable message rather than degraded.
const commonAlertSchemaID = "azureMonitorCommonAlertSchema"

// conditionType values we render natively. Anything else routes to the
// best-effort fallback.
const (
	condSingleResourceMultipleMetric = "SingleResourceMultipleMetricCriteria"
	condDynamicThreshold             = "DynamicThresholdCriteria"
	condLogQuery                     = "LogQueryCriteria"
)

const (
	// maxSearchQueryLen bounds the KQL query. It is the one unbounded field that
	// precedes the search-results link in details, so without a cap a long query
	// would push the link past MaxDetailsLength and truncate it mid-URL.
	maxSearchQueryLen = 1024

	// maxExpressionLen bounds the PromQL expression, which is likewise unbounded.
	maxExpressionLen = 1024

	// maxDescriptionLen bounds essentials.description, which is rendered last.
	maxDescriptionLen = 2048

	// maxMetaValueLen keeps metadata inside alert.ValidateMetadata's total cap;
	// values come straight from the payload.
	maxMetaValueLen = 1024
)

// envelope is the Azure Monitor common alert schema.
//
// https://learn.microsoft.com/azure/azure-monitor/alerts/alerts-common-schema
type envelope struct {
	SchemaID string `json:"schemaId"`
	Data     struct {
		Essentials   essentials        `json:"essentials"`
		AlertContext alertContext      `json:"alertContext"`
		CustomProps  map[string]string `json:"customProperties"`
	} `json:"data"`
}

// essentials is present and identically shaped on every payload regardless of
// signal type, which is what makes the fallback path viable.
type essentials struct {
	AlertID            string   `json:"alertId"`
	AlertRule          string   `json:"alertRule"`
	AlertRuleID        string   `json:"alertRuleId"`
	Severity           string   `json:"severity"`
	SignalType         string   `json:"signalType"`
	MonitorCondition   string   `json:"monitorCondition"`
	MonitoringService  string   `json:"monitoringService"`
	AlertTargetIDs     []string `json:"alertTargetIDs"`
	ConfigurationItems []string `json:"configurationItems"`
	OriginAlertID      string   `json:"originAlertId"`
	FiredDateTime      string   `json:"firedDateTime"`
	ResolvedDateTime   string   `json:"resolvedDateTime"`
	Description        string   `json:"description"`

	// TargetResourceGroup and TargetResourceType are the routing/context fields
	// Microsoft documents the essentials block as existing to provide.
	TargetResourceGroup string `json:"targetResourceGroup"`
	TargetResourceType  string `json:"targetResourceType"`

	// InvestigationLink opens the alert in Azure Monitor. Not present on every
	// payload, but a direct actionable link when it is.
	InvestigationLink string `json:"investigationLink"`
}

// alertContext varies by signal type. Only the fields we render are declared;
// conditionType is the discriminator, and its absence is expected (Service
// Health and activity-log payloads have no condition at all).
type alertContext struct {
	ConditionType string `json:"conditionType"`

	Condition struct {
		WindowSize      string      `json:"windowSize"`
		WindowStartTime string      `json:"windowStartTime"`
		WindowEndTime   string      `json:"windowEndTime"`
		AllOf           []criterion `json:"allOf"`
	} `json:"condition"`

	// Properties duplicates data.customProperties on some payloads, and carries
	// the Service Health incident detail on others.
	Properties map[string]json.RawMessage `json:"properties"`

	// Azure Managed Prometheus. This shape has no conditionType and no
	// condition.allOf; the PromQL expression and the rule's annotations carry the
	// diagnostic content instead.
	Expression      string            `json:"expression"`
	ExpressionValue string            `json:"expressionValue"`
	For             string            `json:"for"`
	Interval        string            `json:"interval"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	RuleGroup       string            `json:"ruleGroup"`
}

// criterion is one entry of condition.allOf. Metric and log criteria share this
// envelope, differing only in which fields are populated.
type criterion struct {
	MetricName      string      `json:"metricName"`
	MetricNamespace string      `json:"metricNamespace"`
	Operator        string      `json:"operator"`
	TimeAggregation string      `json:"timeAggregation"`
	Dimensions      []dimension `json:"dimensions"`

	// Threshold is a string while MetricValue is a number -- do not assume they
	// share a type.
	Threshold   string   `json:"threshold"`
	MetricValue *float64 `json:"metricValue"`

	// DynamicThresholdCriteria only. Threshold is a sensitivity artifact for
	// these, not a meaningful limit, so these are rendered instead.
	AlertSensitivity string          `json:"alertSensitivity"`
	FailingPeriods   *failingPeriods `json:"failingPeriods"`

	// WebtestLocationAvailabilityCriteria only.
	WebTestName string `json:"webTestName"`

	// LogQueryCriteria only.
	SearchQuery                   string `json:"searchQuery"`
	MetricMeasureColumn           string `json:"metricMeasureColumn"`
	LinkToFilteredSearchResultsUI string `json:"linkToFilteredSearchResultsUI"`
	LinkToSearchResultsUI         string `json:"linkToSearchResultsUI"`
}

type dimension struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type failingPeriods struct {
	NumberOfEvaluationPeriods *float64 `json:"numberOfEvaluationPeriods"`
	MinFailingPeriodsToAlert  *float64 `json:"minFailingPeriodsToAlert"`
}

// errLegacySchema is returned for a payload that is not the common alert schema.
// The message names the fix because an operator seeing it needs to change an
// action-group setting, not debug GoAlert.
var errLegacySchema = fmt.Errorf(
	"azuremonitor: unsupported schemaId; enable the common alert schema on this action group's webhook receiver")

// parseInfo reports which shape a payload was recognised as, so the caller can
// log it. ContextRendered is the important one: false means the alert carries
// essentials only because nothing recognised alertContext, which is a graceful
// outcome but also the signal that a new Azure alert type has started arriving.
type parseInfo struct {
	SignalType      string
	MonitorService  string
	ConditionType   string
	ContextRendered bool
}

// buildAlert maps an Azure Monitor webhook body onto an alert.
//
// The returned alert has ServiceID unset; the caller fills it in from the
// integration key. Source is always alert.SourceAzureMonitor and Dedup is always
// non-nil -- a nil dedup silently falls back to a content hash that changes
// between the Fired and Resolved deliveries, which would break the close path.
//
// An error is returned only for a payload that is not the common alert schema.
// Every other input, however unrecognised, produces a best-effort alert built
// from essentials rather than a failure.
func buildAlert(body []byte) (alert.Alert, map[string]string, parseInfo, error) {
	var env envelope
	err := json.Unmarshal(body, &env)

	// Tolerate a type mismatch on an individual field: encoding/json still fills
	// everything it could decode, so a numerically-typed threshold or dimension
	// value costs that one value rather than the whole alert. Azure documents
	// these as strings but is not consistent across shapes, and a hard failure
	// here means a 400 -- which Azure does not retry -- so the page is lost.
	var typeErr *json.UnmarshalTypeError
	if err != nil && !errors.As(err, &typeErr) {
		return alert.Alert{}, nil, parseInfo{}, err
	}
	if env.SchemaID != commonAlertSchemaID {
		return alert.Alert{}, nil, parseInfo{}, errLegacySchema
	}

	e := env.Data.Essentials
	ctx := env.Data.AlertContext

	status := alert.StatusTriggered
	// monitorCondition is the schema-stable field; alertContext.status can
	// disagree with it (the incident resolved while the alert fired).
	if strings.EqualFold(e.MonitorCondition, "Resolved") {
		status = alert.StatusClosed
	}

	// Rendered once and passed in, so the caller can also report whether anything
	// recognised alertContext without parsing twice.
	ctxLines := contextLines(ctx)
	info := parseInfo{
		SignalType:      e.SignalType,
		MonitorService:  e.MonitoringService,
		ConditionType:   ctx.ConditionType,
		ContextRendered: len(ctxLines) > 0,
	}

	summary := validate.SanitizeText(alertSummary(e), alert.MaxSummaryLength)
	details := validate.SanitizeText(alertDetails(e, ctx, ctxLines, env.Data.CustomProps), alert.MaxDetailsLength)

	return alert.Alert{
		Summary: summary,
		Details: details,
		Source:  alert.SourceAzureMonitor,
		Status:  status,
		// alertId is the alert *instance* ID and is stable across the Fired and
		// Resolved deliveries of one firing (Azure alerts are stateful), so this
		// closes the alert it opened. Never use originAlertId: it is per-rule for
		// metric alerts, so a single missed close would mute that rule forever.
		Dedup: alert.NewUserDedup(sha256Hex(e.AlertID)),
	}, buildMeta(e, ctx), info, nil
}

// alertSummary falls back through progressively weaker identifiers. Azure always
// sends alertRule in practice, but an empty summary does not error -- it creates
// a blank, unactionable alert -- so the fallback is a correctness requirement.
func alertSummary(e essentials) string {
	if strings.TrimSpace(e.AlertRule) != "" {
		return e.AlertRule
	}
	if items := nonEmpty(e.ConfigurationItems); len(items) > 0 {
		return "Azure Monitor alert on " + strings.Join(items, ", ")
	}
	if strings.TrimSpace(e.SignalType) != "" {
		return "Azure Monitor " + e.SignalType + " alert"
	}
	return "Azure Monitor alert"
}

// alertDetails renders short high-value fields first so that truncation at
// MaxDetailsLength eats the least useful content. The search-results link is
// kept whole by capping searchQuery upstream of it.
func alertDetails(e essentials, ctx alertContext, ctxLines []string, custom map[string]string) string {
	var lines []string
	add := func(label, value string) {
		if value != "" {
			lines = append(lines, label+": "+value)
		}
	}

	add("Severity", e.Severity)
	add("Monitor condition", e.MonitorCondition)
	add("Signal type", e.SignalType)
	add("Monitoring service", e.MonitoringService)
	add("Fired", e.FiredDateTime)
	add("Resolved", e.ResolvedDateTime)

	// configurationItems are short resource names; alertTargetIDs are full ARM
	// paths, so prefer the former and fall back only when absent.
	if items := nonEmpty(e.ConfigurationItems); len(items) > 0 {
		add("Resource", strings.Join(items, ", "))
	} else if targets := nonEmpty(e.AlertTargetIDs); len(targets) > 0 {
		add("Resource", strings.Join(targets, ", "))
	}
	add("Resource group", e.TargetResourceGroup)
	add("Resource type", e.TargetResourceType)
	add("Investigate", e.InvestigationLink)

	if len(ctxLines) > 0 {
		lines = append(lines, "")
		lines = append(lines, ctxLines...)
	}

	// customProperties is operator-settable per rule and is the natural place for
	// a runbook URL -- the closest Azure analogue to CloudWatch's
	// AlarmDescription. alertContext.properties duplicates it on some payloads.
	props := custom
	if len(props) == 0 {
		props = stringProps(ctx.Properties)
	}
	if len(props) > 0 {
		lines = append(lines, "")
		for _, k := range sortedKeys(props) {
			add(k, props[k])
		}
	}

	if d := strings.TrimSpace(e.Description); d != "" {
		lines = append(lines, "", truncRunes(d, maxDescriptionLen))
	}

	return strings.Join(lines, "\n")
}

// contextLines renders alertContext by conditionType. An unrecognised or absent
// conditionType yields no lines at all, which is the fallback: essentials alone
// still produces a usable alert.
func contextLines(ctx alertContext) []string {
	// Dispatch on the presence of condition.allOf rather than an allowlist of
	// conditionType values. Every metric and log criteria shape shares this one
	// envelope, so this covers MultipleResourceMultipleMetricCriteria (used by
	// multi-resource and resource-group-scoped metric rules) and
	// WebtestLocationAvailabilityCriteria without special-casing either, and
	// stays correct for whatever Azure adds next.
	//
	// conditionType still selects the only two behaviours that genuinely differ:
	// suppressing the meaningless dynamic threshold, and the log query lines.
	if len(ctx.Condition.AllOf) == 0 {
		// Prometheus rule groups have no condition either, but carry their own
		// distinct fields.
		if lines := prometheusLines(ctx); len(lines) > 0 {
			return lines
		}
		// Service Health and activity-log payloads have no condition at all;
		// they carry their detail in properties instead.
		return serviceHealthLines(ctx)
	}

	var lines []string
	for _, c := range ctx.Condition.AllOf {
		lines = append(lines, criterionLines(c, ctx.ConditionType, ctx.Condition.WindowSize)...)
	}
	return lines
}

// criterionLines renders one condition.allOf entry. allOf is an array -- a rule
// can carry several criteria and all of them are rendered.
func criterionLines(c criterion, condType, windowSize string) []string {
	var lines []string

	if condType == condLogQuery {
		if q := strings.TrimSpace(c.SearchQuery); q != "" {
			lines = append(lines, "Query: "+truncRunes(q, maxSearchQueryLen))
		}
		if c.MetricMeasureColumn != "" {
			lines = append(lines, "Measure column: "+c.MetricMeasureColumn)
		}
	}

	if head := c.headline(condType, windowSize); head != "" {
		lines = append(lines, head)
	}

	// The namespace disambiguates same-named metrics across resource types.
	if c.MetricNamespace != "" {
		lines = append(lines, "Namespace: "+c.MetricNamespace)
	}
	if c.WebTestName != "" {
		lines = append(lines, "Web test: "+c.WebTestName)
	}

	if condType == condDynamicThreshold {
		// threshold is a sensitivity artifact here, not a limit, so render the
		// sensitivity and failing-period counts instead of a misleading number.
		if c.AlertSensitivity != "" {
			lines = append(lines, "Sensitivity: "+c.AlertSensitivity)
		}
	}
	if fp := c.FailingPeriods; fp != nil {
		if s := fp.String(); s != "" {
			lines = append(lines, "Failing periods: "+s)
		}
	}

	if dims := dimensionPairs(c.Dimensions); dims != "" {
		lines = append(lines, "Dimensions: "+dims)
	}

	// Exactly one link, and the UI one: it is dimension-scoped to this firing and
	// opens in the portal. The *API variants are not clickable, and all four
	// together exceed MaxDetailsLength on their own.
	if link := c.LinkToFilteredSearchResultsUI; link != "" {
		lines = append(lines, "Results: "+link)
	} else if link := c.LinkToSearchResultsUI; link != "" {
		lines = append(lines, "Results: "+link)
	}

	return lines
}

// headline renders the metric comparison line shared by metric and log criteria.
func (c criterion) headline(condType, windowSize string) string {
	name := c.MetricName
	if name == "" && condType == condLogQuery {
		name = "Results"
	}
	if name == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(name)
	if c.Operator != "" {
		b.WriteString(" " + c.Operator)
	}
	// Suppress the threshold for dynamic criteria: it is a sensitivity artifact.
	if c.Threshold != "" && condType != condDynamicThreshold {
		b.WriteString(" " + c.Threshold)
	}

	var qual []string
	if c.TimeAggregation != "" {
		qual = append(qual, c.TimeAggregation)
	}
	if windowSize != "" {
		qual = append(qual, windowSize)
	}
	if len(qual) > 0 {
		b.WriteString(" (" + strings.Join(qual, ", ") + ")")
	}
	if c.MetricValue != nil {
		b.WriteString(" = " + formatNum(*c.MetricValue))
	}

	return b.String()
}

func (f failingPeriods) String() string {
	if f.MinFailingPeriodsToAlert == nil || f.NumberOfEvaluationPeriods == nil {
		return ""
	}
	return formatNum(*f.MinFailingPeriodsToAlert) + " of " + formatNum(*f.NumberOfEvaluationPeriods)
}

// prometheusLines renders the Azure Managed Prometheus shape. The rule's
// annotations are author-written and are the most useful thing on the alert, so
// they lead; the PromQL expression and the value that tripped it follow.
func prometheusLines(ctx alertContext) []string {
	if ctx.Expression == "" && ctx.RuleGroup == "" {
		return nil
	}

	var lines []string

	// summary and description are the Prometheus conventions and are rendered
	// bare, as prose. Any other annotation is labelled.
	for _, k := range []string{"summary", "description"} {
		if v := strings.TrimSpace(ctx.Annotations[k]); v != "" {
			lines = append(lines, v)
		}
	}
	for _, k := range sortedKeys(ctx.Annotations) {
		if k == "summary" || k == "description" {
			continue
		}
		if v := strings.TrimSpace(ctx.Annotations[k]); v != "" {
			lines = append(lines, k+": "+v)
		}
	}

	add := func(label, value string) {
		if value != "" {
			lines = append(lines, label+": "+value)
		}
	}
	add("Expression", truncRunes(ctx.Expression, maxExpressionLen))
	add("Value", ctx.ExpressionValue)
	add("For", ctx.For)
	add("Interval", ctx.Interval)

	// Labels are Prometheus's equivalent of metric dimensions.
	var pairs []string
	for _, k := range sortedKeys(ctx.Labels) {
		pairs = append(pairs, k+"="+ctx.Labels[k])
	}
	add("Labels", strings.Join(pairs, ", "))

	return lines
}

// serviceHealthLines renders the Service Health / activity-log shape, which has
// no conditionType and carries its detail under properties.
func serviceHealthLines(ctx alertContext) []string {
	props := stringProps(ctx.Properties)
	if len(props) == 0 {
		return nil
	}

	var lines []string
	add := func(label, key string) {
		if v := strings.TrimSpace(props[key]); v != "" {
			lines = append(lines, label+": "+v)
		}
	}
	add("Title", "title")
	add("Service", "service")
	add("Region", "region")
	add("Incident type", "incidentType")
	add("Tracking ID", "trackingId")
	add("Impact start", "impactStartTime")
	add("Stage", "stage")

	return lines
}

func buildMeta(e essentials, ctx alertContext) map[string]string {
	return cleanMeta(map[string]string{
		// Prometheus payloads have no essentials.alertRuleId; ruleGroup is the
		// equivalent handle back to the rule in Azure.
		"rule_group": ctx.RuleGroup,

		"alert_id":            e.AlertID,
		"alert_rule":          e.AlertRule,
		"alert_rule_id":       e.AlertRuleID,
		"severity":            e.Severity,
		"signal_type":         e.SignalType,
		"monitoring_service":  e.MonitoringService,
		"monitor_condition":   e.MonitorCondition,
		"alert_target_ids":    strings.Join(nonEmpty(e.AlertTargetIDs), ","),
		"configuration_items": strings.Join(nonEmpty(e.ConfigurationItems), ","),

		"target_resource_group": e.TargetResourceGroup,
		"target_resource_type":  e.TargetResourceType,
	})
}

// htmlTagRe strips markup from Service Health fields like `communication`, which
// are HTML and unreadable raw on a pager.
var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// stringProps flattens a properties map to the string-valued entries only.
//
// Several Azure properties are strings that *contain* JSON (impactedServices,
// targetResourceTypes). They are deliberately kept as opaque strings -- decoding
// them into a struct fails.
func stringProps(raw map[string]json.RawMessage) map[string]string {
	if len(raw) == 0 {
		return nil
	}

	out := make(map[string]string, len(raw))
	for k, v := range raw {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			// Non-string values (numbers, objects, null) are skipped rather than
			// rendered as raw JSON.
			continue
		}
		s = strings.TrimSpace(htmlTagRe.ReplaceAllString(s, ""))
		if s != "" {
			out[k] = s
		}
	}
	if len(out) == 0 {
		return nil
	}

	return out
}

func dimensionPairs(dims []dimension) string {
	var pairs []string
	for _, d := range dims {
		if d.Name == "" && d.Value == "" {
			continue
		}
		pairs = append(pairs, d.Name+"="+d.Value)
	}
	return strings.Join(pairs, ", ")
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Stable output so details are deterministic across deliveries.
	sort.Strings(keys)
	return keys
}

func nonEmpty(in []string) []string {
	var out []string
	for _, s := range in {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

// formatNum renders a float without a trailing .0, so counts read as integers.
func formatNum(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

func truncRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// cleanMeta drops empty values and bounds the rest. Values are not sanitized:
// they are JSON-marshalled on write, so control characters are escaped rather
// than injected, and trimming would corrupt an ARM resource ID.
func cleanMeta(m map[string]string) map[string]string {
	for k, v := range m {
		if v == "" {
			delete(m, k)
			continue
		}
		m[k] = truncRunes(v, maxMetaValueLen)
	}
	return m
}
