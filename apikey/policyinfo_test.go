package apikey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePolicyInfo(t *testing.T) {
	expected := &policyInfo{
		// Never change this hash -- it is used to verify future changes wont break existing keys.
		Hash: []byte{0x2d, 0x6d, 0xef, 0x3b, 0xfa, 0xd8, 0xef, 0x9d, 0x7e, 0x6c, 0xf4, 0xed, 0xd1, 0x81, 0x16, 0xf4, 0x23, 0xaa, 0x4a, 0xaf, 0x70, 0x1b, 0xfd, 0x1b, 0x26, 0x1d, 0xb3, 0x1, 0x4f, 0xdc, 0xa1, 0x61},
		Policy: GQLPolicy{
			Version: 1,
			Query:   "query",
			Role:    "admin",
		},
	}

	info, err := parsePolicyInfo([]byte(`{"Version":1,"Query":"query","Role":"admin"}`))
	assert.NoError(t, err)
	assert.Equal(t, expected, info)

	// add spaces and re-order keys
	info, err = parsePolicyInfo([]byte(`{ "Role":"admin", "Query":"query","Version":1}`))
	assert.NoError(t, err)
	assert.Equal(t, expected, info)

	// add extra field
	info, err = parsePolicyInfo([]byte(`{"Version":1,"Query":"query","Role":"admin","Extra":true}`))
	assert.NoError(t, err)
	assert.Equal(t, expected, info)

	// changed query
	info, err = parsePolicyInfo([]byte(`{"Version":1,"Query":"query2","Role":"admin"}`))
	assert.NoError(t, err)
	assert.NotEqual(t, expected.Hash, info.Hash)
}
