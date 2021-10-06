package twilio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpellNumber(t *testing.T) {
	// Test the spell number function
	assert.Equal(t, "1. 2. 3. 4. 5. 6", spellNumber(123456))
}
