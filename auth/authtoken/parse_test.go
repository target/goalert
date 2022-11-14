package authtoken

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzParse(f *testing.F) {
	f.Add("01020304-0506-0708-090a-0b0c0d0e0f10") // v0
	f.Add("01020304-0506-0708-090a-0b0c0d0e0f1z") // v0 invalid

	f.Add("U9obklyVC0wduWIy75nbivABDxwc-rANyqNA4CZQzhkJHuNlUCfJDPpcG6W9bEIPddqPbh-sxMS1Km87jC9yLASp3i1UWtdDu2udCzM=")  // v1
	f.Add("U9obklyVC0wduWIy75nbivABDxwc-rANyqNA4CZQzhkJHuNlUCfJDPpcG6W9bEIPddqPbh-sxMS1Km87jC9yLASp3i1UWtdDu2udCzM==") // v1, invalid base64

	f.Add("VgICAQIDBAUGBwgJCgsMDQ4PEAAAAAAAAAU5c2ln") // v2

	f.Fuzz(func(t *testing.T, a string) {
		verifyFn := func(t Type, payload, signature []byte) (isValid bool, isOldKey bool) {
			return true, true
		}
		tok, _, err := Parse(a, verifyFn)
		if err != nil {
			return
		}

		s, err := tok.Encode(func(payload []byte) (signature []byte, err error) {
			return []byte("sig"), nil
		})
		require.NoError(t, err)

		tok2, _, err := Parse(s, verifyFn)
		require.NoError(t, err)
		assert.Equal(t, tok, tok2)
	})
}
