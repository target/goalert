package nfynet

// A Target can resolve a TargetID.
type Target interface {
	TargetID() TargetID
}
