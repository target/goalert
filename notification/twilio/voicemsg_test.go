package twilio

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

// TestVoiceMessageKeys guards the per-language translation files
// (messages_*.go) against silent drift: every map key they register must be a
// real key from voiceKeys. A typo'd key would never be matched at runtime and
// the message would silently fall back to English, so we fail the build for it.
// Missing keys are allowed (they fall back to English by design) but reported.
func TestVoiceMessageKeys(t *testing.T) {
	want := make(map[string]bool, len(voiceKeys))
	for _, k := range voiceKeys {
		want[k] = true
	}

	files, err := filepath.Glob("messages_*.go")
	assert.NoError(t, err)
	assert.NotEmpty(t, files, "expected at least one translation file")

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			keys := mapKeysFromFile(t, file)
			assert.NotEmpty(t, keys, "no translation keys found")

			seen := make(map[string]bool, len(keys))
			for _, k := range keys {
				assert.True(t, want[k], "unknown key (typo? not in voiceKeys): %q", k)
				assert.False(t, seen[k], "duplicate key: %q", k)
				seen[k] = true
			}

			if missing := len(want) - len(seen); missing > 0 {
				t.Logf("%s: %d/%d keys translated (%d fall back to English)", file, len(seen), len(want), missing)
			}
		})
	}
}

// TestVoicePrinterFallback documents how voicePrinter resolves the configured
// language to spoken text and the <Say> language attribute: empty config keeps
// the historical default, a translated regional variant resolves to its base
// catalog while keeping its own attribute, and an untranslated language falls
// back to English text spoken with an en-US voice.
func TestVoicePrinterFallback(t *testing.T) {
	cases := []struct {
		voiceLang string
		wantAttr  string
		wantText  string // translation of "Goodbye."
	}{
		{"", "", "Goodbye."},              // default: English, no attribute
		{"en-GB", "en-GB", "Goodbye."},    // English variant
		{"fr-CA", "fr-CA", "Au revoir."},  // regional variant -> fr catalog
		{"ar-EG", "ar-EG", "مع السلامة."}, // non-Latin, translated
		{"xx-YY", "en-US", "Goodbye."},    // unknown: English text + en-US
	}
	for _, tc := range cases {
		p, attr := voicePrinter(tc.voiceLang)
		assert.Equal(t, tc.wantAttr, attr, "attr for %q", tc.voiceLang)
		assert.Equal(t, tc.wantText, p.Sprintf("Goodbye."), "text for %q", tc.voiceLang)
	}
}

// TestVoiceMultiArgRendering proves x/text substitutes every argument of the
// multi-placeholder templates correctly across several languages (including
// non-Latin scripts), regardless of where the placeholders sit in the
// translated sentence.
func TestVoiceMultiArgRendering(t *testing.T) {
	for _, lang := range []string{"fr-FR", "ja-JP", "ar-EG", "de-DE", "ru-RU"} {
		p, _ := voicePrinter(lang)

		bundle, err := buildMessage(p, "P", notification.AlertBundle{ServiceName: "Widget", Count: 5})
		assert.NoError(t, err)
		assert.Contains(t, bundle, "Widget", lang)
		assert.Contains(t, bundle, "5", lang)

		verify, err := buildMessage(p, "P", notification.Verification{Code: "1234"})
		assert.NoError(t, err)
		// the 4-digit code is spelled out as "1. 2. 3. 4"
		assert.Contains(t, verify, spellCode("1234"), lang)
		// no leftover/missing-arg markers from a placeholder mismatch
		assert.NotContains(t, verify, "%!", lang)
		assert.NotContains(t, bundle, "%!", lang)
	}
}

func mapKeysFromFile(t *testing.T, file string) []string {
	t.Helper()
	src, err := os.ReadFile(file)
	assert.NoError(t, err)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, 0)
	assert.NoError(t, err)

	var keys []string
	ast.Inspect(f, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if _, ok := cl.Type.(*ast.MapType); !ok {
			return true
		}
		for _, e := range cl.Elts {
			kv, ok := e.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			bl, ok := kv.Key.(*ast.BasicLit)
			if !ok || bl.Kind != token.STRING {
				continue
			}
			if s, err := strconv.Unquote(bl.Value); err == nil {
				keys = append(keys, s)
			}
		}
		return true
	})
	return keys
}
