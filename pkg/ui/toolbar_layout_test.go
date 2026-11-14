package ui

import "testing"

// ToolbarTopCommon.cpp justifyToolbar / centreToolbar (2da9f6f, 7ba7099) on an
// 800-wide screen: justified, the left groups from 0 (the hidden cheats slot
// taking no room) and the right ones ending at 799 (Towns 770–799); centred,
// all 13 visible buttons side by side, 11 between groups, in a toolbar of
// 13 × 30 + 3 × 11 = 423 at x 188.
func TestToolbarLayout(t *testing.T) {
	tb := NewToolbar(800)
	tb.Layout(800, false)
	x := func(name string) int {
		for _, b := range tb.Buttons {
			if b.Tooltip == name {
				return b.X
			}
		}
		t.Fatalf("no button %q", name)
		return 0
	}
	want := map[string]int{"Load/Save": 0, "Audio": 30, "Zoom": 71, "View": 131, "Towns": 770, "Stations": 740, "Vehicles": 710,
		"Build Vehicles": 669, "Terraform": 549}
	for k, v := range want {
		if got := x(k); got != v {
			t.Errorf("justified %s at %d, want %d", k, got, v)
		}
	}
	if tb.X != 0 || tb.Width != 800 {
		t.Errorf("justified toolbar %d+%d", tb.X, tb.Width)
	}
	tb.Layout(800, true)
	if tb.X != 188 || tb.Width != 423 {
		t.Fatalf("centred toolbar %d+%d, want 188+423", tb.X, tb.Width)
	}
	want = map[string]int{"Load/Save": 188, "Zoom": 188 + 60 + 11, "Terraform": 188 + 150 + 22, "Vehicles": 188 + 300 + 33, "Towns": 188 + 360 + 33}
	for k, v := range want {
		if got := x(k); got != v {
			t.Errorf("centred %s at %d, want %d", k, got, v)
		}
	}
}
