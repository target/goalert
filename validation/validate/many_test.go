package validate

import (
	"github.com/target/goalert/validation"
	"strings"
	"testing"
)

func TestMany(t *testing.T) {
	err := validation.NewFieldError("Test", "test error")

	e := Many(err)
	if !validation.IsValidationError(e) {
		t.Errorf("got %v; want %v", err, e)
	}

	e = Many(err, nil)
	if e == nil {
		t.Error("got nil; want err")
	}

	e = Many(err, validation.NewFieldError("Other", "other error"))
	if e == nil {
		t.Error("got nil; want err")
	}
	if !strings.Contains(e.Error(), "Test") {
		t.Errorf("got '%s'; should contain 'Test'", e.Error())
	}
	if !strings.Contains(e.Error(), "Other") {
		t.Errorf("got '%s'; should contain 'Other'", e.Error())
	}

	if !validation.IsValidationError(e) {
		t.Error("IsValidationError = false; want true")
	}
}
