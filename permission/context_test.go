package permission

import (
	"context"
	"testing"
)

// ExampleSudoContext shows how to use SudoContext.
func ExampleSudoContext() {
	// the original context could be from anywhere (req.Context() in an http.Handler for example)
	ctx := context.Background()
	SudoContext(ctx, func(ctx context.Context) {
		// within this function scope, ctx now has System privileges
	})
	// once the function returns, the elevated context is cancelled, but the original ctx is still valid
}

func TestSudoContext(t *testing.T) {
	SudoContext(context.Background(), func(ctx context.Context) {
		if !System(ctx) {
			t.Error("System(ctx) == false; want true")
		}
		err := LimitCheckAny(ctx, System)
		if err != nil {
			t.Errorf("err = %v; want nil", err)
		}
		err = LimitCheckAny(ctx, System, Admin, User)
		if err != nil {
			t.Errorf("err = %v; want nil", err)
		}
	})
}

func TestWithoutAuth(t *testing.T) {
	check := func(ctx context.Context, name string) {
		t.Run(name, func(t *testing.T) {
			ctx = WithoutAuth(ctx)
			if User(ctx) {
				t.Error("User() = true; want false")
			}
			if Admin(ctx) {
				t.Error("Admin() = true; want false")
			}
			if System(ctx) {
				t.Error("System() = true; want false")
			}
			if Service(ctx) {
				t.Error("Service() = true; want false")
			}
			if ServiceID(ctx) != "" {
				t.Errorf("SeriviceID() = %s; want empty string", ServiceID(ctx))
			}
			if UserID(ctx) != "" {
				t.Errorf("UserID() = %s; want empty string", UserID(ctx))
			}
			if SystemComponentName(ctx) != "" {
				t.Errorf("SystemComponentName() = %s; want empty string", SystemComponentName(ctx))
			}
		})
	}
	ctx := context.Background()
	data := []struct {
		ctx  context.Context
		name string
	}{
		{name: "user_role_user", ctx: UserContext(ctx, "bob", RoleUser)},
		{name: "user_role_unknown", ctx: UserContext(ctx, "bob", RoleUnknown)},
		{name: "user_role_admin", ctx: UserContext(ctx, "bob", RoleAdmin)},
		{name: "system", ctx: SystemContext(ctx, "test")},
		{name: "service", ctx: ServiceContext(ctx, "test")},
	}

	for _, d := range data {
		check(d.ctx, d.name)
	}
}
