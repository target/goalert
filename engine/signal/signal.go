package signal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/target/goalert/notification"
	"github.com/target/goalert/service/rule"
)

// Signal represents the data for an outgoing signal.
type OutgoingSignal struct {
	ID   string
	Type notification.MessageType
	Dest notification.Dest

	SignalID  int
	UserID    string
	ServiceID string
	CreatedAt time.Time
	SentAt    time.Time

	Message string
	Content json.RawMessage
}

type Destination interface {

	// Content processes action context returning the signal message along with optional destination value and additional json context if necessary
	Content(map[string]string) (json.RawMessage, string, string, error)
}

// MapContent returns a list of service rule contents as a map of props to values
func MapContent(content []rule.Content) map[string]string {
	var cm = make(map[string]string)
	for _, c := range content {
		cm[c.Prop] = c.Value
	}
	return cm
}

// ProcessContent handles preparing signals and their corresponding service rule actions to more easily usable data to create an outoging signal
func ProcessContent(action rule.Action) (string, string, json.RawMessage, error) {
	var dest Destination
	var message string = ""
	var destValue string

	// utilize a map of contents to more easily access contents
	contentMap := MapContent(action.Contents)

	switch destType := action.DestType; destType {
	case "SLACK":
		dest = SlackChannel{}
	case "WEBHOOK":
		dest = UserWebhook{}
	default:
		return message, destValue, json.RawMessage{}, fmt.Errorf("could not process unknown destination %s", destType)
	}

	// what is the point of this??? we already aquire the message/body in the switch statement
	// content should be the additional context needed that varies between the notification channels
	content, destValue, message, err := dest.Content(contentMap)
	if err != nil {
		err = fmt.Errorf("signalsmanager destination content error: %w", err)
	}

	return message, destValue, content, err
}

func mapToStruct(data map[string]string, obj interface{}) error {
	v := reflect.ValueOf(obj)

	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("obj must be a pointer to a struct")
	}

	elem := v.Elem()
	elemType := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		fieldName := strings.ToLower(elemType.Field(i).Name)
		jsonTag := elemType.Field(i).Tag.Get("json")

		if jsonTag != "" {
			jsonTag = strings.TrimSuffix(jsonTag, ",omitempty")
			fieldName = jsonTag
		}

		if val, ok := data[fieldName]; ok {
			if field.CanSet() {
				if field.Type().Kind() == reflect.String {
					field.SetString(val)
				}
			}
		}
	}

	return nil
}
