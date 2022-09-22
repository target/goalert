package swo

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConnInfo(t *testing.T) {
	info := &ConnInfo{
		Version: "v0.31.0",
		Type:    ConnTypeMainMgr,
		ID:      uuid.Nil,
	}

	assert.Equal(t, "GoAlert v0.31.0 SWO:A:AAAAAAAAAAAAAAAAAAAAAA", info.String())

	_, err := ParseConnInfo("GoAlert 1.0.0 SWO:0:AAAAAAAAAAAAAAAAAAAAAA")
	assert.ErrorContains(t, err, "invalid connection type")

	info, err = ParseConnInfo("GoAlert 1.0.0 SWO:A:AAAAAAAAAAAAAAAAAAAAAA")
	assert.NoError(t, err)
	assert.Equal(t, ConnTypeMainMgr, info.Type)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "GoAlert 1.0.0 SWO:A:AAAAAAAAAAAAAAAAAAAAAA", info.String())
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", info.ID.String())
}
