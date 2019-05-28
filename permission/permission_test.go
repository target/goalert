package permission

import (
	"context"
	"fmt"
)

func ExampleUserContext() {
	// start with any context
	ctx := context.Background()

	// pass it through UserContext to assign a user ID and Role
	ctx = UserContext(ctx, "user-id-here", RoleAdmin)

	// later on it can be checked anywhere; this example will satisfy the Admin role requirement
	err := LimitCheckAny(ctx, Admin)

	fmt.Println(err)
	// output: <nil>
}
