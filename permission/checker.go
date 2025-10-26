package permission

import (
	"context"
	"strings"
)

// A Checker is used to give a pass-or-fail result for a given context.
type Checker func(context.Context) bool

// LimitCheckAny will return a permission error if none of the checks pass, or
// if the auth check limit is reached. If no checks are provided, only
// the limit check, and a check that the context has SOME type authorization is
// performed. nil can be passed as an always-fail check option (useful to prevent
// the no-check behavior, if required).
func LimitCheckAny(ctx context.Context, checks ...Checker) error {
	if !All(ctx) {
		return newGeneric(true, "")
	}

	if len(checks) == 0 {
		return nil
	}
	for _, c := range checks {
		if c != nil && c(ctx) {
			return nil
		}
	}

	return newGeneric(false, "")
}

// All is a Checker that checks against ALL providers, returning true
// if any are found.
func All(ctx context.Context) bool {
	if v, ok := ctx.Value(contextHasAuth).(int); ok && v == 1 {
		return true
	}

	return false
}

// Admin is a Checker that determines if a context has the Admin or System role.
func Admin(ctx context.Context) bool {
	if System(ctx) {
		return true
	}
	r, ok := ctx.Value(contextKeyUserRole).(Role)
	if ok && r == RoleAdmin {
		return true
	}

	return false
}

// User is a Checker that determines if a context has the User, Admin or System role.
func User(ctx context.Context) bool {
	if System(ctx) {
		return true
	}
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
