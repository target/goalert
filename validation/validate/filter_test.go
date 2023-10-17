package validate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterField(t *testing.T) {
	valid := []string{"field", "field_1", "this-is-a-field!@#$%^*()", "field\u1234"}
	invalid := []string{"quote\"", "ampersand&", "\ttab\t", "new\nline"}

	t.Run("valid fields", func(t *testing.T) {
		for _, field := range valid {
			assert.NoError(t, FilterField("Field", field), fmt.Sprintf("'%s' was invalid field but should be valid", field))
		}
	})
	t.Run("invalid fields", func(t *testing.T) {
		for _, field := range invalid {
			assert.Error(t, FilterField("Field", field), fmt.Sprintf("'%s' was valid field but should be invalid", field))
		}
	})
}

func TestFilterStringOperator(t *testing.T) {
	valid := []string{"==", "!=", "contains", "startsWith", "endsWith"}
	invalid := []string{"matches", "CONTAINS", "<", ""}

	t.Run("valid string operator", func(t *testing.T) {
		for _, operator := range valid {
			assert.NoError(t, FilterStringOperator("Operator", operator), fmt.Sprintf("'%s' was invalid string operator but should be valid", operator))
		}
	})
	t.Run("invalid string operator", func(t *testing.T) {
		for _, operator := range invalid {
			assert.Error(t, FilterStringOperator("Operator", operator), fmt.Sprintf("'%s' was valid string operator but should be invalid", operator))
		}
	})
}

func TestFilterNumberOperator(t *testing.T) {
	valid := []string{"==", "!=", "<", "<=", ">", ">="}
	invalid := []string{"contains", "!<", ""}

	t.Run("valid number operator", func(t *testing.T) {
		for _, operator := range valid {
			assert.NoError(t, FilterNumberOperator("Operator", operator), fmt.Sprintf("'%s' was invalid number operator but should be valid", operator))
		}
	})
	t.Run("invalid number operator", func(t *testing.T) {
		for _, operator := range invalid {
			assert.Error(t, FilterNumberOperator("Operator", operator), fmt.Sprintf("'%s' was valid number operator but should be invalid", operator))
		}
	})
}

func TestFilterNumberValue(t *testing.T) {
	valid := []string{"0", "-30", "2.4", "-.3", "2000000000"}
	invalid := []string{"2.1.", "3a", "string", "1/2", ""}

	t.Run("valid number value", func(t *testing.T) {
		for _, value := range valid {
			assert.NoError(t, FilterNumberValue("Value", value), fmt.Sprintf("'%s' was invalid number value but should be valid", value))
		}
	})
	t.Run("invalid number value", func(t *testing.T) {
		for _, value := range invalid {
			assert.Error(t, FilterNumberValue("Value", value), fmt.Sprintf("'%s' was valid number value but should be invalid", value))
		}
	})
}

func TestFilterBoolOperator(t *testing.T) {
	valid := []string{"==", "!="}
	invalid := []string{"contains", "<=", ""}

	t.Run("valid boolean operator", func(t *testing.T) {
		for _, operator := range valid {
			assert.NoError(t, FilterBoolOperator("Operator", operator), fmt.Sprintf("'%s' was invalid boolean operator but should be valid", operator))
		}
	})
	t.Run("invalid boolean operator", func(t *testing.T) {
		for _, operator := range invalid {
			assert.Error(t, FilterBoolOperator("Operator", operator), fmt.Sprintf("'%s' was valid boolean operator but should be invalid", operator))
		}
	})
}

func TestFilterBoolValue(t *testing.T) {
	valid := []string{"true", "false"}
	invalid := []string{"True", "3", "string", ""}

	t.Run("valid boolean value", func(t *testing.T) {
		for _, value := range valid {
			assert.NoError(t, FilterBoolValue("Value", value), fmt.Sprintf("'%s' was invalid boolean value but should be valid", value))
		}
	})
	t.Run("invalid boolean value", func(t *testing.T) {
		for _, value := range invalid {
			assert.Error(t, FilterBoolValue("Value", value), fmt.Sprintf("'%s' was valid boolean value but should be invalid", value))
		}
	})
}
