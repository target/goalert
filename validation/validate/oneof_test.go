package validate

import (
	"testing"
)

type testType string

const (
	testType1 = testType("a")
	testType2 = testType("b")
	testType3 = testType("c")
)

func TestOneOf(t *testing.T) {
	m := testType1
	err := OneOf("foo", m, testType1, testType2, testType3)
	if err != nil {
		t.Errorf("err was %+v; want nil", err)
	}
}
