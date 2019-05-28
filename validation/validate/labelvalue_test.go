package validate

import "testing"

func TestLabelValue(t *testing.T) {
	check := func(valid bool, values ...string) {
		for _, val := range values {
			t.Run(val, func(t *testing.T) {
				t.Log("'" + val + "'")
				err := LabelValue("", val)
				if valid && err != nil {
					t.Errorf("got %v; want nil", err)
				} else if !valid && err == nil {
					t.Errorf("got nil; want err")
				}
			})
		}
	}

	check(true,
		"foo", "foo bar", "FooBar", "foo-bar", "foo- 9bar", "", "foo'/bar", "@okay", "&n*#9\\; wowz@ $ \\/yee",
	)
	check(false,
		"    ", " foo", "foo ", "fo", "-", "unprintable"+string('\t'), "unprintable"+string('\n'), "unprintable"+string('\v'), "unprintable"+string('\f'), "unprintable"+string('\r'),
	)
}
