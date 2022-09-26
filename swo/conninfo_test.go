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
	info.ID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	assert.Equal(t, "GoAlert v0.31.0 SWO:A:EREREREREREREREREREREQ", info.String())

	_, err := ParseConnInfo("GoAlert 1.0.0 SWO:0:EREREREREREREREREREREQ")
	assert.ErrorContains(t, err, "invalid connection type")

	info, err = ParseConnInfo("GoAlert 1.0.0 SWO:A:EREREREREREREREREREREQ")
	assert.NoError(t, err)
	assert.Equal(t, ConnTypeMainMgr, info.Type)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "GoAlert 1.0.0 SWO:A:EREREREREREREREREREREQ", info.String())
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", info.ID.String())
}
