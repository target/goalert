package keyring

import (
	"crypto"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAFSplitMerge(t *testing.T) {
	const input = "Hello, World!"

	split, err := AFSplit([]byte(input), 1000, crypto.SHA256)
	assert.NoError(t, err)
	t.Log("split:", hex.EncodeToString(split))
	merged, err := AFMerge(split, 1000, crypto.SHA256)
	assert.NoError(t, err)
	t.Log("merge:", string(merged))
	assert.Equal(t, input, string(merged))
}
