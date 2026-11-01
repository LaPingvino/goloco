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
