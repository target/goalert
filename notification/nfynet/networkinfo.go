package nfynet

// NetworkInfo provides information for configuring and displaying information related to this network.
type NetworkInfo struct {
	NetworkID

	Name       string
	SubNetName string

	// TypeNamePlural indicates how targets should be referred to (e.g., "Channels" or "Users" for Slack).
	TypeNamePlural string

	TargetIDType TargetIDType

	// IconURL is an optional URL to an icon for this specific network.
	//
	// A good example may be the Slack workspace/team icon.
	IconURL string
}

// TargetIDType is used to indicate the type of target ID, this is used to determine how filters are applied.
//
// For example, if the TargetIDType is TargetIDTypePhone,
// a filter may be for an individual country code. For email filters could be applied per domain, etc...
type TargetIDType int

const (
	TargetIDTypeUnknown TargetIDType = iota
	TargetIDTypePhone
	TargetIDTypeEmail
	TargetIDTypeURL
	TargetIDTypeText
)
