package expflag

import "context"

type flagSetKeyT struct{}

var flagSetKey flagSetKeyT

// Context returns a new context with the given FlagSet.
func Context(ctx context.Context, fs FlagSet) context.Context {
	return context.WithValue(ctx, flagSetKey, fs)
}

// ContextHas returns true if the given context has the given flag.
func ContextHas(ctx context.Context, flag Flag) bool {
	fs, ok := ctx.Value(flagSetKey).(FlagSet)
	if !ok {
		return false
	}

	return fs.Has(flag)
}
