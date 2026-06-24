package nfymsg

// SignalMessage is a dynamic message that is sent to a notification destination.
type SignalMessage struct {
	Base

	Params map[string]string
}

func (t SignalMessage) Param(name string) string {
	if t.Params == nil {
		return ""
	}
	return t.Params[name]
}
