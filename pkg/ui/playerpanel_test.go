package ui

import "testing"

// formatShortWithOneDecimal: value / 10 "." value % 10, C++ truncation —
// a negative value keeps the minus on both parts, as upstream prints it.
func TestInt16OneDecimal(t *testing.T) {
	for v, want := range map[int16]string{473: "47.3", 1000: "100.0", 5: "0.5", -15: "-1.-5"} {
		if got := int16OneDecimal(v); got != want {
			t.Errorf("%d: %q, want %q", v, got, want)
		}
	}
}
