package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiltersToExprString(t *testing.T) {
	t.Run("all types", func(t *testing.T) {
		str := `Temp > 25 && age <= 42.2 && other != false && message == "alert"`
		expected := []Filter{
			{
				Field:     "Temp",
				Operator:  ">",
				Value:     "25",
				ValueType: TypeNumber,
			},
			{
				Field:     "age",
				Operator:  "<=",
				Value:     "42.2",
				ValueType: TypeNumber,
			},
			{
				Field:     "other",
				Operator:  "!=",
				Value:     "false",
				ValueType: TypeBool,
			},
			{
				Field:     "message",
				Operator:  "==",
				Value:     "alert",
				ValueType: TypeString,
			},
		}
		filters, err := FiltersFromExprString(str)
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
				ValueType: TypeString,
			},
			{
				Field:     "message",
				Operator:  "contains",
				Value:     `alert\ && another`,
				ValueType: TypeString,
			},
		}
		filters, err := FiltersFromExprString(str)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, filters)
	})
}

func TestFiltersFromExprString(t *testing.T) {
	t.Run("all types", func(t *testing.T) {
		filters := []Filter{
			{
				Field:     "Temp",
				Operator:  ">",
				Value:     "25",
				ValueType: TypeNumber,
			},
			{
				Field:     "age",
				Operator:  "<=",
				Value:     "42.2",
				ValueType: TypeNumber,
			},
			{
				Field:     "other",
				Operator:  "!=",
				Value:     "FALSE",
				ValueType: TypeBool,
			},
			{
				Field:     "message",
				Operator:  "==",
				Value:     "alert",
				ValueType: TypeString,
			},
		}
		expected := `Temp > 25 && age <= 42.2 && other != false && message == "alert"`
		str, err := FiltersToExprString(filters)
		assert.NoError(t, err)
		assert.Equal(t, expected, str)
	})
	t.Run("string escaping", func(t *testing.T) {
		filters := []Filter{
			{
				Field:     "weather",
				Operator:  "startsWith",
				Value:     `"cloud"`,
				ValueType: TypeString,
			},
			{
				Field:     "message",
				Operator:  "contains",
				Value:     `alert\ && another`,
				ValueType: TypeString,
			},
		}
		expected := `weather startsWith "\"cloud\"" && message contains "alert\\ \u0026\u0026 another"`
		filterString, err := FiltersToExprString(filters)
		assert.NoError(t, err)
		assert.Equal(t, expected, filterString)
	})
}
