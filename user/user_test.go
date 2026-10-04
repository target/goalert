package user

import (
	"github.com/target/goalert/permission"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchOptions_Only(t *testing.T) {
	validUUID := "A035FD3C-73C8-4F72-BECD-36B027AE1374"

	_, err := (*renderData)(&SearchOptions{Only: []string{validUUID}}).Normalize()
	require.NoError(t, err)

	_, err = (*renderData)(&SearchOptions{Only: []string{"not-a-uuid"}}).Normalize()
	require.Error(t, err)

	tooMany := make([]string, 51)
	for i := range tooMany {
		tooMany[i] = validUUID
	}
	_, err = (*renderData)(&SearchOptions{Only: tooMany}).Normalize()
	require.Error(t, err)

	args := (*renderData)(&SearchOptions{Only: []string{validUUID}}).QueryArgs()
	require.Condition(t, func() bool {
		for _, a := range args {
			if a.Name == "only" {
				return a.Value != nil
			}
		}
		return false
	}, "QueryArgs() should include a non-nil 'only' parameter")
}

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
