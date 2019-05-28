package user

import (
	"github.com/target/goalert/permission"
	"testing"
)

func TestUser_Normalize(t *testing.T) {
	test := func(valid bool, u User) {
		name := "valid"
		if !valid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			t.Logf("%+v", u)
			_, err := u.Normalize()
			if valid && err != nil {
				t.Errorf("got %v; want nil", err)
			} else if !valid && err == nil {
				t.Errorf("got nil; want")
			}
		})
	}

	valid := []User{
		{Name: "Joe", Role: permission.RoleAdmin, Email: "foo@bar.com"},
	}
	invalid := []User{
		{},
	}
	for _, u := range valid {
		test(true, u)
	}
	for _, u := range invalid {
		test(false, u)
	}
}
