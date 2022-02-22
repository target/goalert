package contactmethod

import "github.com/target/goalert/notification/nfynet"

var (
	netVoice = nfynet.NetworkID{ID: "phone", SubTypeID: "voice"}
	netSMS   = nfynet.NetworkID{ID: "phone", SubTypeID: "sms"}
	netEmail = nfynet.NetworkID{ID: "email"}
	netWeb   = nfynet.NetworkID{ID: "webhook"}

	_ = nfynet.Targeter(ContactMethod{})
)

// SetTargetID sets the Type and Value appropriately for the given TargetID.
func (cm *ContactMethod) SetTargetID(id nfynet.TargetID) {
	switch id.NetworkID {
	case netVoice:
		cm.Type = TypeVoice
		cm.Value = id.ID
	case netSMS:
		cm.Type = TypeSMS
		cm.Value = id.ID
	case netEmail:
		cm.Type = TypeEmail
		cm.Value = id.ID
	case netWeb:
		cm.Type = TypeWebhook
		cm.Value = id.ID
	default:
		cm.Type = "V2"
		cm.Value = id.String()
	}
}

// TargetID returns the TargetID for the ContactMethod.
func (cm ContactMethod) TargetID() nfynet.TargetID {
	switch cm.Type {
	case TypeVoice:
		return nfynet.TargetID{
			ID:        cm.Value,
			NetworkID: netVoice,
		}
	case TypeSMS:
		return nfynet.TargetID{
			ID:        cm.Value,
			NetworkID: netSMS,
		}
	case TypeEmail:
		return nfynet.TargetID{
			ID:        cm.Value,
			NetworkID: netEmail,
		}
	case TypeWebhook:
		return nfynet.TargetID{
			ID:        cm.Value,
			NetworkID: netWeb,
		}
	case TypeV2:
		t, _ := nfynet.ParseTargetID(cm.Value)
		if t != nil {
			return *t
		}
	}

	// invalid/unknown
	return nfynet.TargetID{}
}
