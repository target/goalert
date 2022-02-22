package notificationchannel

import "github.com/target/goalert/notification/nfynet"

// legacy slack network -- no workspace/team ID
var (
	netSlackChannel = nfynet.NetworkID{ID: "slack", SubTypeID: "channel"}

	_ = nfynet.Target(Channel{})
)

// SetTargetID sets the Type and Value appropriately for the given TargetID.
func (c *Channel) SetTargetID(id nfynet.TargetID) {
	switch id.NetworkID {
	case netSlackChannel:
		c.Type = TypeSlack
		c.Value = id.ID
	default:
		c.Type = TypeV2
		c.Value = id.String()
	}
}

// TargetID returns the TargetID for the ContactMethod.
func (c Channel) TargetID() nfynet.TargetID {
	switch c.Type {
	case TypeSlack:
		return nfynet.TargetID{
			ID:        c.Value,
			NetworkID: netSlackChannel,
		}
	case TypeV2:
		t, _ := nfynet.ParseTargetID(c.Value)
		if t != nil {
			return *t
		}
	}

	// invalid/unknown
	return nfynet.TargetID{}
}
