package notification

import "github.com/target/goalert/notification/nfymsg"

type (
	Alert               = nfymsg.Alert
	AlertStatus         = nfymsg.AlertStatus
	AlertBundle         = nfymsg.AlertBundle
	Message             = nfymsg.Message
	MessageType         = nfymsg.MessageType
	Test                = nfymsg.Test
	Verification        = nfymsg.Verification
	SignalMessage       = nfymsg.SignalMessage
	ScheduleOnCallUsers = nfymsg.ScheduleOnCallUsers

	State = nfymsg.State
	User  = nfymsg.User

	Status      = nfymsg.Status
	SendResult  = nfymsg.SendResult
	SentMessage = nfymsg.SentMessage

	AlertState = nfymsg.AlertState
)

const (
	StateSending    = nfymsg.StateSending
	StateFailedPerm = nfymsg.StateFailedPerm
	StateFailedTemp = nfymsg.StateFailedTemp
	StateDelivered  = nfymsg.StateDelivered
	StateSent       = nfymsg.StateSent
	StateBundled    = nfymsg.StateBundled
	StateUnknown    = nfymsg.StateUnknown
	StatePending    = nfymsg.StatePending

	AlertStateUnknown        = nfymsg.AlertStateUnknown
	AlertStateUnacknowledged = nfymsg.AlertStateUnacknowledged
	AlertStateAcknowledged   = nfymsg.AlertStateAcknowledged
	AlertStateClosed         = nfymsg.AlertStateClosed
)
