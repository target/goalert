package keyring

import (
	"crypto"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAFSplitMerge(t *testing.T) {
	const input = "Hello, World!"

	split := AFSplit([]byte(input), 10, crypto.SHA256)
	t.Log("split:", hex.EncodeToString(split))
	merged := AFMerge(split, 10, crypto.SHA256)
	t.Log("merge:", string(merged))
	assert.Equal(t, input, string(merged))
}
