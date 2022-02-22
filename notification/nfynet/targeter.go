package nfynet

// A Targeter can resolve a TargetID.
type Targeter interface {
	TargetID() TargetID
}
