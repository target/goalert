package uik

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/gadb"
)

func TestNewAlert(t *testing.T) {
	const serviceID = "8b6c1d42-1a43-4c0e-9b7a-0d6f1e5b2c11"

	newAct := func(params map[string]string) gadb.UIKActionV1 {
		return gadb.UIKActionV1{Dest: gadb.DestV1{Type: "builtin-alert"}, Params: params}
	}

	t.Run("text that would fail validation is sanitized", func(t *testing.T) {
		a := newAlert(serviceID, newAct(map[string]string{
			// no-break space, zero-width space, trailing space
			"summary": "prod down\u00a0!\u200b ",
			// carriage returns, narrow no-break space, invalid UTF-8
			"details": "line one\r\nline two\u202f:\xe2\x82",
			"dedup":   "group-1",
		}))

		assert.Equal(t, "prod down !", a.Summary)
		assert.Equal(t, "line one\nline two :\ufffd\ufffd", a.Details)
		assert.Equal(t, alert.StatusTriggered, a.Status)
		assert.Equal(t, alert.SourceUniversal, a.Source)

		_, err := a.Normalize()
		require.NoError(t, err, "sanitized alert must pass validation")
	})

	t.Run("text is truncated to the maximum length", func(t *testing.T) {
		a := newAlert(serviceID, newAct(map[string]string{
			"summary": strings.Repeat("é", alert.MaxSummaryLength+10),
			"details": strings.Repeat("€", alert.MaxDetailsLength+10),
		}))

		assert.Len(t, []rune(a.Summary), alert.MaxSummaryLength)
		assert.Len(t, []rune(a.Details), alert.MaxDetailsLength)

		_, err := a.Normalize()
		require.NoError(t, err, "truncated alert must pass validation")
	})

	t.Run("close param closes the alert", func(t *testing.T) {
		a := newAlert(serviceID, newAct(map[string]string{"summary": "s", "close": "true"}))
		assert.Equal(t, alert.StatusClosed, a.Status)
	})
}
