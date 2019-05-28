package permission

import (
	"context"
	"strings"
	"sync/atomic"

	"go.opencensus.io/trace"
)

// A Checker is used to give a pass-or-fail result for a given context.
type Checker func(context.Context) bool

func checkLimit(ctx context.Context) error {
	n, ok := ctx.Value(contextKeyCheckCount).(*uint64)
	if !ok {
		return newGeneric(false, "invalid auth context for check limit")
	}
	max, ok := ctx.Value(contextKeyCheckCountMax).(uint64)
	if !ok {
		return newGeneric(false, "invalid auth context for max check limit")
	}

	v := atomic.AddUint64(n, 1) // always add
	if max > 0 && v > max {
		return newGeneric(false, "exceeded auth check limit")
	}

	return nil
}

// LimitCheckAny will return a permission error if none of the checks pass, or
// if the auth check limit is reached. If no checks are provided, only
// the limit check, and a check that the context has SOME type authorization is
// performed. nil can be passed as an always-fail check option (useful to prevent
// the no-check behavior, if required).
func LimitCheckAny(ctx context.Context, checks ...Checker) error {
	if !All(ctx) {
		return newGeneric(true, "")
	}

	// ensure we don't get hammered with auth checks (or DB calls, for example)
	err := checkLimit(ctx)
	if err != nil {
		return err
	}

	if len(checks) == 0 {
		return nil
	}
	for _, c := range checks {
		if c != nil && c(ctx) {
			addAuthAttrs(ctx)
			return nil
		}
	}

	return newGeneric(false, "")
}

func addAuthAttrs(ctx context.Context) {
	sp := trace.FromContext(ctx)
	if sp == nil {
		return
	}
	var attrs []trace.Attribute
	if User(ctx) {
		attrs = append(attrs,
			sourceAttrs(ctx,
				trace.StringAttribute("auth.user.id", UserID(ctx)),
				trace.StringAttribute("auth.user.role", string(userRole(ctx))),
			)...)
	}
	if System(ctx) {
		attrs = append(attrs, trace.StringAttribute("auth.system.componentName", SystemComponentName(ctx)))
	}
	if Service(ctx) {
		attrs = append(attrs, sourceAttrs(ctx,
			trace.StringAttribute("auth.service.id", ServiceID(ctx)),
		)...)
	}
	if len(attrs) == 0 {
		return
	}
	sp.AddAttributes(attrs...)
}

// All is a Checker that checks against ALL providers, returning true
// if any are found.
func All(ctx context.Context) bool {
	if v, ok := ctx.Value(contextHasAuth).(int); ok && v == 1 {
		return true
	}

	return false
}

// Admin is a Checker that determines if a context has the Admin role.
func Admin(ctx context.Context) bool {
	r, ok := ctx.Value(contextKeyUserRole).(Role)
	if ok && r == RoleAdmin {
		return true
	}

	return false
}

// User is a Checker that determines if a context has the User or Admin role.
func User(ctx context.Context) bool {
	r, ok := ctx.Value(contextKeyUserRole).(Role)
	if ok && (r == RoleUser || r == RoleAdmin) {
		return true
	}

	return false
}

// Service is a Checker that determines if a context has a serviceID.
func Service(ctx context.Context) bool {
	return ServiceID(ctx) != ""
}

// System is a Checker that determines if a context has system privileges.
func System(ctx context.Context) bool {
	return SystemComponentName(ctx) != ""
}

// Team is a Checker that determines if a context has team privileges.
func Team(ctx context.Context) bool {
	return TeamID(ctx) != ""
}

// MatchService will return a Checker that ensures the context has the given ServiceID.
func MatchService(serviceID string) Checker {
	return func(ctx context.Context) bool {
		if serviceID == "" {
			return false
		}
		return ServiceID(ctx) == strings.ToLower(serviceID)
	}
}

// MatchTeam will return a Checker that ensures the context has the given TeamID.
func MatchTeam(teamID string) Checker {
	return func(ctx context.Context) bool {
		return TeamID(ctx) == strings.ToLower(teamID)
	}
}

// MatchUser will return a Checker that ensures the context has the given UserID.
func MatchUser(userID string) Checker {
	return func(ctx context.Context) bool {
		if userID == "" {
			return false
		}
		return UserID(ctx) == strings.ToLower(userID)
	}
}
