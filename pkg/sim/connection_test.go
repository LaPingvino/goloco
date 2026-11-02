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

// Three pieces built in a row (straight, a small left curve, straight) are
// found by getTrackConnections from each piece's connection end, forwards;
// and backwards from the first piece's start, the reversed (bit 2) entry of
// nothing — the line has no piece behind it. Another owner or track object
// finds nothing.
func TestTrackConnections(t *testing.T) {
	s := flatState(30, 64)
	s.ConsHead = ConstructionHead{Active: true, X: 10, Y: 10, BaseZ: 64, Rotation: 0}
	type placed struct {
		tad     uint16
		x, y, z int32
	}
	var line []placed
	for _, id := range []int{0, 4, 0} { // straight, leftCurveSmall, straight
		h := s.ConsHead
		if !s.PlaceTrackAtHead(id, 0) {
			t.Fatalf("track %d can't be built", id)
		}
		line = append(line, placed{uint16(id)<<3 | uint16(h.Rotation&3), int32(h.X) * 32, int32(h.Y) * 32, int32(h.BaseZ) * 4})
	}
	for i := 0; i+1 < len(line); i++ {
		p := line[i]
		e, _ := TrackConnectionEnd(p.x, p.y, p.z, p.tad)
		tc := s.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, 0, 0, 0, 0)
		if len(tc.Connections) != 1 || tc.Connections[0]&TaDBasicMask != line[i+1].tad {
			t.Errorf("after piece %d (tad %d): connections %v, want [%d]", i, p.tad, tc.Connections, line[i+1].tad)
		}
		if tc.StationID != -1 || tc.HasLevelCrossing {
			t.Errorf("after piece %d: station %d crossing %v", i, tc.StationID, tc.HasLevelCrossing)
		}
		if o := s.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, 1, 0, 0, 0); len(o.Connections) != 0 {
			t.Errorf("another company's train finds %v", o.Connections)
		}
	}
	// Reversed: the last straight, entered from its far end, is found as
	// tad|4 (the reversed direction), arriving with the opposite rotation.
	last := line[2]
	e, _ := TrackConnectionEnd(last.x, last.y, last.z, last.tad)
	rev := last.tad | 4
	back := TrackCoordinates[rev].RotBegin
	ex, ey := e.X, e.Y
	// The reversed piece begins on the tile the forward piece ended on: step back one tile.
	ex -= int32(TrackCoordinates[last.tad].DX)
	ey -= int32(TrackCoordinates[last.tad].DY)
	tc := s.TrackConnections(ex, ey, e.Z, back, 0, 0, 0, 0)
	found := false
	for _, c := range tc.Connections {
		if c&TaDBasicMask == rev {
			found = true
		}
	}
	if !found {
		t.Errorf("reversed straight: connections %v, want %d", tc.Connections, rev)
	}
}
