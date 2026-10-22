package harness

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLogEntry_errText(t *testing.T) {
	type testCase struct {
		raw     string
		errText string
		level   string
	}
	check := func(desc string, c testCase) {
		t.Helper()
		entry, err := parseLogEntry([]byte(c.raw))
		assert.NoError(t, err, desc)
		assert.Equal(t, c.level, entry.Level, desc)
		assert.Equal(t, c.errText, entry.errText(), desc)
	}

	check("legacy error key", testCase{
		raw:     `{"level":"error","msg":"something failed","error":"bad thing"}`,
		errText: `"bad thing"`,
		level:   "error",
	})
	check("structured error object", testCase{
		raw:     `{"level":"error","msg":"something failed","error":{"code":"E1","detail":"bad"}}`,
		errText: `{"code":"E1","detail":"bad"}`,
		level:   "error",
	})
	// River's ErrorHandler (app/initriver.go) logs under `err`.
	check("river job error key", testCase{
		raw:     `{"level":"error","msg":"Job returned error.","module":"river","job.queue":"rotation-manager","job.id":1,"job.kind":"rotation-manager-update","err":"update rotation: boom"}`,
		errText: `"update rotation: boom"`,
		level:   "error",
	})
	// River's ErrorHandler (app/initriver.go) logs panics under `panic`.
	check("river job panic key", testCase{
		raw:     `{"level":"error","msg":"Job panicked.","module":"river","job.queue":"rotation-manager","panic":"runtime error: nil pointer","trace":"goroutine 1 [running]"}`,
		errText: `"runtime error: nil pointer"`,
		level:   "error",
	})
	check("error key preferred over err", testCase{
		raw:     `{"level":"error","error":"first","err":"second"}`,
		errText: `"first"`,
		level:   "error",
	})
	check("null error is treated as absent", testCase{
		raw:     `{"level":"error","error":null,"err":"real"}`,
		errText: `"real"`,
		level:   "error",
	})
	check("no error keys", testCase{
		raw:     `{"level":"info","msg":"Engine cycle start."}`,
		errText: "",
		level:   "info",
	})

	_, err := parseLogEntry([]byte(`not json`))
	assert.Error(t, err, "invalid JSON should return an error")
}
