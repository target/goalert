package apikey

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/permission"
)

// Note: This is a reference output, it should never be modified.
//
// Policy validation is done by comparing the hash of the policy data, and so the JSON representation of the policy must be consistent.
const v1ReferenceKey = `{"Version":1,"Query":"query","Role":"admin"}`

func TestGQLPolicy(t *testing.T) {
	data, err := json.Marshal(GQLPolicy{Version: 1, Query: "query", Role: permission.RoleAdmin})
	assert.NoError(t, err)
	assert.Equal(t, v1ReferenceKey, string(data))
}
