package universalapi

import (
	"github.com/target/goalert/alert"
)

// Temporary implementation
type AlertMapper struct {
	Summary string
	Details string
	Dedup   string
	Status  alert.Status
}

// BuildOutgoingAlert creates an outgoing alert applying templating for the
// summary and detail fields when applicable.
func BuildOutgoingAlert(payload map[string]interface{}, rule Rule) (outgoingAlert AlertMapper, err error) {
	outgoingAlert.Summary, err = applyTemplateOrDefault("summary", payload, rule.Template)
	if err != nil {
		return outgoingAlert, err
	}
	outgoingAlert.Details, err = applyTemplateOrDefault("details", payload, rule.Template)
	if err != nil {
		return outgoingAlert, err
	}

	outgoingAlert.Status = alert.StatusTriggered
	if rule.Action == "close" {
		outgoingAlert.Status = alert.StatusClosed
	}
	return
}

// TODO: applyTemplateOrDefault applies the given template to a field if the
// field value is not found in the payload
func applyTemplateOrDefault(fieldName string, payload map[string]interface{}, template string) (res string, err error) {
	return
}
