package alert

import (
	"crypto/sha512"
	"encoding/hex"
	"strings"
	"time"

	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// maximum lengths
const (
	MaxSummaryLength = 1024     // 1KiB
	MaxDetailsLength = 6 * 1024 // 6KiB
)

// An Alert represents an ongoing situation.
type Alert struct {
	ID        int       `json:"_id"`
	Status    Status    `json:"status"`
	Summary   string    `json:"summary"`
	Details   string    `json:"details"`
	Source    Source    `json:"source"`
	ServiceID string    `json:"service_id"`
	CreatedAt time.Time `json:"created_at"`
	Dedup     *DedupID  `json:"dedup"`
	Meta      AlertMeta `json:"meta"`
}

// DedupKey will return the de-duplication key for the alert.
// The Dedup prop is used if non-nil, otherwise one is generated
// using the Description of the Alert.
func (a *Alert) DedupKey() *DedupID {
	if a.Dedup != nil {
		return a.Dedup
	}

	// fallback is auto:1:<lcase(Sum512(Description))>
	sum := sha512.Sum512([]byte(a.Description()))
	return &DedupID{
		Type:    DedupTypeAuto,
		Version: 1,
		Payload: hex.EncodeToString(sum[:]),
	}
}

func (a *Alert) scanFrom(scanFn func(...interface{}) error) error {
	return scanFn(&a.ID, &a.Summary, &a.Details, &a.ServiceID, &a.Source, &a.Status, &a.CreatedAt, &a.Dedup)
}

func (a Alert) Normalize() (*Alert, error) {
	if string(a.Source) == "" {
		a.Source = SourceManual
	}
	if string(a.Status) == "" {
		a.Status = StatusTriggered
	}
	a.Summary = strings.Replace(a.Summary, "\n", " ", -1)
	a.Summary = strings.Replace(a.Summary, "  ", " ", -1)

	var validateMeta error
	for k := range a.Meta.AlertMetaV1 {
		if k == "" {
			validateMeta = validation.NewFieldError("Meta", "key must be non empty string")
		}
	}

	err := validate.Many(
		validate.Text("Summary", a.Summary, 1, MaxSummaryLength),
		validate.Text("Details", a.Details, 0, MaxDetailsLength),
		validate.OneOf("Source", a.Source, SourceManual, SourceGrafana, SourceSite24x7, SourcePrometheusAlertmanager, SourceEmail, SourceGeneric),
		validate.OneOf("Status", a.Status, StatusTriggered, StatusActive, StatusClosed),
		validate.UUID("ServiceID", a.ServiceID),
		validateMeta,
	)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func (a Alert) Description() string {
	if a.Details == "" {
		return a.Summary
	}
	return a.Summary + "\n" + a.Details
}
