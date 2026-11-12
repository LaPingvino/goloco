package core

import "testing"

func TestStrLogicalCmp(t *testing.T) {
	less := [][2]string{
		{"Train 2", "Train 10"},
		{"train 2", "Train 3"}, // case-insensitive
		{"Bus", "Bus 1"},       // a prefix first
		{"A 1,000", "A 1,001"}, // ',' inside a number skipped
		{"Truck 9", "truck 99"},
	}
	for _, c := range less {
		if StrLogicalCmp(c[0], c[1]) >= 0 || StrLogicalCmp(c[1], c[0]) <= 0 {
			t.Errorf("%q should sort before %q", c[0], c[1])
		}
	}
	if StrLogicalCmp("Ship 7", "ship 7") != 0 {
		t.Error("equal but for case should compare equal")
	}
	if n, _ := parseNextNumber("99999999999 x"); n != 2147483647 {
		t.Errorf("overflow: %d", n)
	}
}
