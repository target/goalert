package graphqlapp

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user/contactmethod"
)

func TestFilterVisibleContactMethods(t *testing.T) {
	cms := []contactmethod.ContactMethod{
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), UserID: "user-a", Private: false},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), UserID: "user-a", Private: true},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), UserID: "user-b", Private: true},
	}

	t.Run("owner sees their own private contact method", func(t *testing.T) {
		ctx := permission.UserContext(t.Context(), "user-a", permission.RoleUser)
		result := filterVisibleContactMethods(ctx, cms)
		require.Len(t, result, 2)
		for _, cm := range result {
			require.Equal(t, "user-a", cm.UserID)
		}
	})

	t.Run("non-owner never sees a private contact method, even their own roster's", func(t *testing.T) {
		ctx := permission.UserContext(t.Context(), "user-b", permission.RoleUser)
		result := filterVisibleContactMethods(ctx, cms)
		require.Len(t, result, 2)
		require.Equal(t, "user-a", result[0].UserID)
		require.False(t, result[0].Private)
		require.Equal(t, "user-b", result[1].UserID)
	})

	t.Run("empty input returns empty, not nil", func(t *testing.T) {
		ctx := permission.UserContext(t.Context(), "user-a", permission.RoleUser)
		result := filterVisibleContactMethods(ctx, nil)
		require.Empty(t, result)
	})
}
