package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToExprString(t *testing.T) {
	t.Run("all types", func(t *testing.T) {
		str := `Temp > 25 && age <= 42.2 && other != false && message == "alert"`
		expected := []Filter{
			{
				Field:     "Temp",
				Operator:  ">",
				Value:     "25",
				ValueType: NumberType,
			},
			{
				Field:     "age",
				Operator:  "<=",
				Value:     "42.2",
				ValueType: NumberType,
			},
			{
				Field:     "other",
				Operator:  "!=",
				Value:     "false",
				ValueType: BoolType,
			},
			{
				Field:     "message",
				Operator:  "==",
				Value:     "alert",
				ValueType: StringType,
			},
		}
		filters, err := FromExprString(str)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, filters)
	})
	t.Run("string escaping", func(t *testing.T) {
		str := `weather startsWith "cloud" && message contains "alert\\ \u0026\u0026 another"`
		expected := []Filter{
			{
				Field:     "weather",
				Operator:  "startsWith",
				Value:     "cloud",
				ValueType: StringType,
			},
			{
				Field:     "message",
				Operator:  "contains",
				Value:     `alert\ && another`,
				ValueType: StringType,
			},
		}
		filters, err := FromExprString(str)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, filters)
	})
}

func TestFromExprString(t *testing.T) {
	t.Run("all types", func(t *testing.T) {
		filters := []Filter{
			{
				Field:     "Temp",
				Operator:  ">",
				Value:     "25",
				ValueType: NumberType,
			},
			{
				Field:     "age",
				Operator:  "<=",
				Value:     "42.2",
				ValueType: NumberType,
			},
			{
				Field:     "other",
				Operator:  "!=",
				Value:     "FALSE",
				ValueType: BoolType,
			},
			{
				Field:     "message",
				Operator:  "==",
				Value:     "alert",
				ValueType: StringType,
			},
		}
		expected := `Temp > 25 && age <= 42.2 && other != false && message == "alert"`
		str, err := ToExprString(filters)
		assert.NoError(t, err)
		assert.Equal(t, expected, str)
	})
	t.Run("string escaping", func(t *testing.T) {
		filters := []Filter{
			{
				Field:     "weather",
				Operator:  "startsWith",
				Value:     `"cloud"`,
				ValueType: StringType,
			},
			{
				Field:     "message",
				Operator:  "contains",
				Value:     `alert\ && another`,
				ValueType: StringType,
			},
		}
		expected := `weather startsWith "\"cloud\"" && message contains "alert\\ \u0026\u0026 another"`
		filterString, err := ToExprString(filters)
		assert.NoError(t, err)
		assert.Equal(t, expected, filterString)
	})
}
