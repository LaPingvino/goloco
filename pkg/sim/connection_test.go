package sim

import "testing"

// A piece's sub-position path ends where its connection end begins: the last
// MoveInfo of every track piece lies on the edge of the tile getUnkTrack
// points to (within 2 units of its 32×32 square). Straight track
// (trackAndDirection 0..3) moves exactly one tile with the rotation kept.
func TestTrackConnectionEnd(t *testing.T) {
	for d := uint16(0); d < 4; d++ {
		e, ok := TrackConnectionEnd(0, 0, 0, d)
		if !ok || e.Z != 0 || abs(int(e.X))+abs(int(e.Y)) != 32 || e.NextRotation != uint8(d) {
			t.Errorf("straight dir %d: %+v", d, e)
		}
	}
	far := 0
	for tad := range TrackCoordinates {
		p := TrackSubPosition(uint16(tad))
		e, _ := TrackConnectionEnd(0, 0, 0, uint16(tad))
		last := p[len(p)-1]
		if outside(int(last.X), int(e.X)) > 2 || outside(int(last.Y), int(e.Y)) > 2 {
			far++
		}
	}
	if far != 0 {
		t.Errorf("%d pieces whose path ends more than a tile from their connection end", far)
	}
	if _, ok := RoadConnectionEnd(0, 0, 0, 80); ok {
		t.Error("road trackAndDirection 80 out of range accepted")
	}
}

// outside is how far v lies outside the tile [lo, lo+31].
func outside(v, lo int) int {
	switch {
	case v < lo:
		return lo - v
	case v > lo+31:
		return v - lo - 31
	}
	return 0
}
