package config

import (
	"context"

	"github.com/target/goalert/permission"
)

// ResourceCheck returns a permission.Checker for creating, modifying, or deleting
// shared resources. If General.RestrictResourceChanges is enabled, only the Admin
// role (or a system context) passes; otherwise the provided fallback is used.
func ResourceCheck(fallback permission.Checker) permission.Checker {
	return func(ctx context.Context) bool {
		cfg, ok := ctx.Value(contextKeyConfig).(Config)
		if ok && cfg.General.RestrictResourceChanges {
			return permission.Admin(ctx)
		}

		return fallback(ctx)
	}
}
