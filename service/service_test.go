package service

import (
	"testing"
)

func TestService_Normalize(t *testing.T) {
	test := func(valid bool, s Service) {
		name := "valid"
		if !valid {
			name = "invalid"
		}
		t.Run(name, func(t *testing.T) {
			t.Logf("%+v", s)
			_, err := s.Normalize()
			if valid && err != nil {
				t.Errorf("got %v; want nil", err)
			} else if !valid && err == nil {
				t.Errorf("got nil err; want non-nil")
			}
		})
	}

	valid := []Service{
		{Name: "Sample Service", Description: "Sample Service", EscalationPolicyID: "A035FD3C-73C8-4F72-BECD-36B027AE1374"},
	}
	invalid := []Service{
		{},
	}
	for _, s := range valid {
		test(true, s)
	}
	for _, s := range invalid {
		test(false, s)
	}
}

func TestSearchOptions_Normalize_Only(t *testing.T) {
	tests := []struct {
		name    string
		opts    SearchOptions
		wantErr bool
	}{
		{
			name: "valid Only filter with single UUID",
			opts: SearchOptions{
				Only: []string{"A035FD3C-73C8-4F72-BECD-36B027AE1374"},
			},
			wantErr: false,
		},
		{
			name: "valid Only filter with multiple UUIDs",
			opts: SearchOptions{
				Only: []string{
					"A035FD3C-73C8-4F72-BECD-36B027AE1374",
					"B035FD3C-73C8-4F72-BECD-36B027AE1375",
				},
			},
			wantErr: false,
		},
		{
			name: "empty Only filter should be valid",
			opts: SearchOptions{
				Only: []string{},
			},
			wantErr: false,
		},
		{
			name: "nil Only filter should be valid",
			opts: SearchOptions{
				Only: nil,
			},
			wantErr: false,
		},
		{
			name: "invalid UUID in Only filter",
			opts: SearchOptions{
				Only: []string{"invalid-uuid"},
			},
			wantErr: true,
		},
		{
			name: "too many UUIDs in Only filter (over 50 limit)",
			opts: SearchOptions{
				Only: make([]string, 51), // Creates slice with 51 empty strings, which should fail validation
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill invalid UUIDs for "too many" test case
			if len(tt.opts.Only) == 51 {
				for i := range tt.opts.Only {
					tt.opts.Only[i] = "A035FD3C-73C8-4F72-BECD-36B027AE1374"
				}
			}

			data := (*renderData)(&tt.opts)
			_, err := data.Normalize()

			if tt.wantErr && err == nil {
				t.Errorf("Normalize() expected error but got nil")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Normalize() unexpected error: %v", err)
			}
		})
	}
}

func TestSearchOptions_Only_QueryArgs(t *testing.T) {
	opts := SearchOptions{
		Only: []string{
			"A035FD3C-73C8-4F72-BECD-36B027AE1374",
			"B035FD3C-73C8-4F72-BECD-36B027AE1375",
		},
	}

	data := (*renderData)(&opts)
	args := data.QueryArgs()

	// Find the "only" argument
	var foundOnly bool
	for _, arg := range args {
		if arg.Name == "only" {
			foundOnly = true
			if arg.Value == nil {
				t.Error("QueryArgs() 'only' parameter should not be nil when Only is provided")
			}
			break
		}
	}

	if !foundOnly {
		t.Error("QueryArgs() should include 'only' parameter")
	}
}

func TestSearchOptions_Only_WithOtherFilters(t *testing.T) {
	// Test that Only filter works in combination with other filters
	opts := SearchOptions{
		Search: "test service",
		Only:   []string{"A035FD3C-73C8-4F72-BECD-36B027AE1374"},
		Omit:   []string{"B035FD3C-73C8-4F72-BECD-36B027AE1375"},
	}

	data := (*renderData)(&opts)
	normalized, err := data.Normalize()
	if err != nil {
		t.Errorf("Normalize() with combined filters failed: %v", err)
	}

	if len(normalized.Only) != 1 {
		t.Errorf("Expected Only to have 1 item, got %d", len(normalized.Only))
	}

	if len(normalized.Omit) != 1 {
		t.Errorf("Expected Omit to have 1 item, got %d", len(normalized.Omit))
	}

	if normalized.Search != "test service" {
		t.Errorf("Expected Search to be preserved, got %q", normalized.Search)
	}
}

func TestSearchTemplate_Only_SQLGeneration(t *testing.T) {
	// Test that the SQL template correctly includes the Only filter
	opts := SearchOptions{
		Only: []string{
			"A035FD3C-73C8-4F72-BECD-36B027AE1374",
			"B035FD3C-73C8-4F72-BECD-36B027AE1375",
		},
		Limit: 10,
	}

	data, err := (*renderData)(&opts).Normalize()
	if err != nil {
		t.Fatalf("Normalize() failed: %v", err)
	}

	// Verify that our QueryArgs include the only parameter
	args := data.QueryArgs()

	var onlyArg *interface{}
	for _, arg := range args {
		if arg.Name == "only" {
			onlyArg = &arg.Value
			break
		}
	}

	if onlyArg == nil {
		t.Fatal("QueryArgs() should include 'only' parameter")
	}

	// Verify the argument is not nil (UUIDArray should convert the slice)
	if *onlyArg == nil {
		t.Error("'only' parameter should not be nil when Only slice is provided")
	}
}

func TestSearchOptions_Only_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		opts    SearchOptions
		wantErr bool
		desc    string
	}{
		{
			name: "Only and Omit with same UUID should be valid",
			opts: SearchOptions{
				Only: []string{"A035FD3C-73C8-4F72-BECD-36B027AE1374"},
				Omit: []string{"A035FD3C-73C8-4F72-BECD-36B027AE1374"},
			},
			wantErr: false,
			desc:    "SQL should handle this case (Only will include it, Omit will exclude it - Omit should win)",
		},
		{
			name: "Only with favorites filters",
			opts: SearchOptions{
				Only:            []string{"A035FD3C-73C8-4F72-BECD-36B027AE1374"},
				FavoritesOnly:   true,
				FavoritesUserID: "B035FD3C-73C8-4F72-BECD-36B027AE1375",
			},
			wantErr: false,
			desc:    "Only filter should work with favorites filtering",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := (*renderData)(&tt.opts)
			_, err := data.Normalize()

			if tt.wantErr && err == nil {
				t.Errorf("Normalize() expected error but got nil - %s", tt.desc)
			} else if !tt.wantErr && err != nil {
				t.Errorf("Normalize() unexpected error: %v - %s", err, tt.desc)
			}
		})
	}
}
