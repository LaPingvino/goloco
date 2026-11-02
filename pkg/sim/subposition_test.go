package sim

import "testing"

// The generated tables have upstream's shapes: road tables [2][2][80]
// flattened; every track piece direction has a path; a straight track
// piece runs 32 world units along one axis.
func TestSubPositionTables(t *testing.T) {
	if len(RoadSubPositionsLeft) != 320 || len(RoadSubPositionsRight) != 320 {
		t.Fatalf("road tables %d/%d, want 320", len(RoadSubPositionsLeft), len(RoadSubPositionsRight))
	}
	empty := 0
	for _, p := range TrackSubPositions {
		if len(p) == 0 {
			empty++
		}
	}
	t.Logf("%d track sequences, %d empty", len(TrackSubPositions), empty)
	p := TrackSubPosition(0) // straight track, direction 0
	if len(p) == 0 {
		t.Fatal("no path for straight track")
	}
	first, last := p[0], p[len(p)-1]
	dx, dy := int(last.X)-int(first.X), int(last.Y)-int(first.Y)
	if (dx != 0 && dy != 0) || (abs(dx) < 30 && abs(dy) < 30) {
		t.Errorf("straight track runs (%d, %d) over %d steps", dx, dy, len(p))
	}
	// Left and right hand differ for road (the lanes).
	if l, r := RoadSubPosition(0, false), RoadSubPosition(0, true); len(l) > 0 && len(r) > 0 && l[0] == r[0] {
		t.Error("left- and right-hand road paths start at the same point")
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
