package normalization

import (
	"testing"
)

func TestParseEstimatedOwners(t *testing.T) {
	tests := []struct {
		input       string
		expectedMin int64
		expectedMax int64
		expectedMid int64
		expectedLbl string
		isValid     bool
	}{
		{"0 - 0", 0, 0, 0, "0 - 0", true},
		{"0 - 20000", 0, 20000, 10000, "0 - 20,000", true},
		{"20,000 - 50,000", 20000, 50000, 35000, "20,000 - 50,000", true},
		{"1000000 - 2000000", 1000000, 2000000, 1500000, "1,000,000 - 2,000,000", true},
		{"50000 - 20000", 20000, 50000, 35000, "20,000 - 50,000", true}, // reversed
		{"invalid", 0, 0, 0, "invalid", false},
		{"", 0, 0, 0, "Unknown", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			res := ParseEstimatedOwners(tc.input)
			if res.IsValid != tc.isValid {
				t.Fatalf("expected isValid=%v, got %v", tc.isValid, res.IsValid)
			}
			if tc.isValid {
				if res.MinOwners != tc.expectedMin || res.MaxOwners != tc.expectedMax || res.MidOwners != tc.expectedMid {
					t.Errorf("expected bounds %d-%d (mid %d), got %d-%d (mid %d)",
						tc.expectedMin, tc.expectedMax, tc.expectedMid, res.MinOwners, res.MaxOwners, res.MidOwners)
				}
				if res.NormalizedRange != tc.expectedLbl {
					t.Errorf("expected label %q, got %q", tc.expectedLbl, res.NormalizedRange)
				}
			}
		})
	}
}
