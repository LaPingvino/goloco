package construct

import "testing"

// getInflationAdjustedCost and the piece cost: worked by hand from the
// formulas (a factor of 1024 for index 3 makes the base cost the factor).
func TestPieceCost(t *testing.T) {
	var f [32]uint32
	f[3] = 1024
	if got := InflationAdjustedCost(120, 3, 10, &f); got != 120 {
		t.Errorf("inflation-adjusted %d, want 120", got)
	}
	if got := TrackPieceCost(0, 120, 3, &f); got != 120 { // straight: 256/256
		t.Errorf("straight %d", got)
	}
	if got := TrackPieceCost(6, 120, 3, &f); got != 120*1005/256 { // left curve
		t.Errorf("curve %d", got)
	}
	if got := RoadPieceCost(9, 64, 3, &f); got != 64*603/256 { // turnaround
		t.Errorf("road turnaround %d", got)
	}
}

// A chosen mod adds its cost when it suits the piece; one that doesn't suit
// it (validMods) or isn't chosen adds nothing.
func TestModsCost(t *testing.T) {
	var f [32]uint32
	f[1] = 1024
	wires := &Mod{Pieces: 0xFFFF, BuildCostFactor: 40, CostIndex: 1}
	straightOnly := &Mod{Pieces: 1, BuildCostFactor: 40, CostIndex: 1} // diagonal flag 1 only
	if got := ModsCost(false, 0, 0b01, []*Mod{wires}, &f); got != 40 {
		t.Errorf("wires on a straight: %d", got)
	}
	if got := ModsCost(false, 6, 0b01, []*Mod{wires}, &f); got != 40*1005/256 {
		t.Errorf("wires on a curve: %d", got)
	}
	if got := ModsCost(false, 6, 0b01, []*Mod{straightOnly}, &f); got != 0 {
		t.Errorf("an unsuitable mod: %d", got)
	}
	if got := ModsCost(false, 0, 0b00, []*Mod{wires}, &f); got != 0 {
		t.Errorf("not chosen: %d", got)
	}
}
