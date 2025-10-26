package smoke

import (
	"testing"

	"github.com/target/goalert/app"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/test/smoke/harness"
)

// TestDataEncKeyRotation tests that the data encryption key can be rotated.
func TestDataEncKeyRotation(t *testing.T) {
	h := harness.NewStoppedHarnessWithFlags(t, "", nil, "", nil)
	defer h.Close()

	// first startup will generate keys using the provided key
	h.StartWithAppCfgHook(func(c *app.Config) {
		c.EncryptionKeys = keyring.Keys{[]byte("test-orig-key")}
	})

	// second startup will use the original key for existing data, and the new key for new data
	h.RestartGoAlertWithAppCfgHook(func(c *app.Config) {
		c.EncryptionKeys = keyring.Keys{[]byte("test-new-key"), []byte("test-orig-key")}
	})
	// ensure we re-encrypt _all_ data with the new key
	h.GraphQLQuery2(`mutation{ reEncryptKeyringsAndConfig }`)

	// Lastly, we should be able to startup with only the new key
	h.RestartGoAlertWithAppCfgHook(func(c *app.Config) {
		c.EncryptionKeys = keyring.Keys{[]byte("test-new-key")}
	})
}
