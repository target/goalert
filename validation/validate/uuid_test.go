package validate

import (
	"strings"
	"testing"
)

func TestParseUUID(t *testing.T) {
	test := func(valid bool, id string) {
		var title string
		if valid {
			title = "Valid"
		} else {
			title = "Invalid"
		}
		t.Run(title, func(t *testing.T) {
			parsed, err := ParseUUID("UUID", id)
			if err == nil && !valid {
				t.Errorf("ParseUUID(%s) err = nil; want error", id)
			} else if err != nil && valid {
				t.Errorf("ParseUUID(%s) err = %v; want nil", id, err)
			} else if valid && !strings.EqualFold(parsed.String(), id) {
				t.Errorf("ParseUUID(%s) parsed = %s; want %s", id, parsed.String(), id)
			}
		})
	}

	invalid := []string{
		"", "12345", "b8b3ee1d-5ff8-4751-9$08-cb3e8b214790", "b8b3ee1d5ff847519208cb3e8b214790",
	}
	valid := []string{
		"b8b3ee1d-5ff8-4751-9208-cb3e8b214790", "00000000-0000-0000-0000-000000000000", "B8B3EE1D-5FF8-4751-9208-cb3e8b214790",
	}
	for _, n := range invalid {
		test(false, n)
	}
	for _, n := range valid {
		test(true, n)
	}
}

func TestParseManyUUID(t *testing.T) {
	maxLength := 3
	test := func(valid bool, ids []string) {
		var title string
		if valid {
			title = "Valid"
		} else {
			title = "Invalid"
		}
		t.Run(title, func(t *testing.T) {
			parsed, err := ParseManyUUID("UUID", ids, maxLength)
			if err == nil && !valid {
				t.Errorf("ParseManyUUID(%v) err = nil; want error", ids)
			} else if err != nil && valid {
				t.Errorf("ParseUUID(%v) err = %v; want nil", ids, err)
			} else if valid {
				for i, p := range parsed {
					if !strings.EqualFold(p.String(), ids[i]) {
						t.Errorf("ParseUUID(%v) parsed[%d] = %s; want %s", ids, i, p.String(), ids[i])
					}
				}
			}
		})
	}

	invalid := [][]string{
		{"b8b3ee1d-5ff8-4751-9208-cb3e8b214790", "", "00000000-0000-0000-0000-000000000000"},
		{"b8b3ee1d-5ff8-4751-9208-cb3e8b214790", "12345", "00000000-0000-0000-0000-000000000000"},
		{"b8b3ee1d-5ff8-4751-9208-cb3e8b214790", "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"},
	}
	valid := [][]string{
		{"B8B3EE1D-5FF8-4751-9208-cb3e8b214790"},
		{"b8b3ee1d-5ff8-4751-9208-cb3e8b214790", "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000001"},
		{"b8b3ee1d-5ff8-4751-9208-cb3e8b214790"},
		{},
	}
	for _, n := range invalid {
		test(false, n)
	}
	for _, n := range valid {
		test(true, n)
	}
}
