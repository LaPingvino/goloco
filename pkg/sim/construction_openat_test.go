package sim

import "testing"

// setNextAndPreviousTrackTile (Construction/Common.cpp): a straight's two
// ends are the tiles beyond it, each heading away; a diagonal's start stays
// on its own tile (rotations from 12 take no step back).
func TestNextAndPreviousTile(t *testing.T) {
	next, prev, ok := NextAndPreviousTile(false, 0, 0, 0, 64, 64, 16)
	if !ok || next != (ConstructionEnd{32, 64, 16, 0}) || prev != (ConstructionEnd{96, 64, 16, 2}) {
		t.Errorf("straight, rotation 0: next %+v previous %+v ok %v", next, prev, ok)
	}
	next, prev, ok = NextAndPreviousTile(true, 0, 0, 1, 64, 64, 0)
	if !ok || next != (ConstructionEnd{64, 96, 0, 1}) || prev != (ConstructionEnd{64, 32, 0, 3}) {
		t.Errorf("road straight, rotation 1: next %+v previous %+v ok %v", next, prev, ok)
	}
	// the second tile of a large curve leads back to the piece's first tile
	p := TrackPieces[4][1]
	c := TrackCoordinates[4<<3]
	next, _, ok = NextAndPreviousTile(false, 4, 1, 0, 64+int32(p.X), 64+int32(p.Y), int32(p.Z))
	if !ok || next != (ConstructionEnd{64 + int32(c.DX), 64 + int32(c.DY), int32(c.DZ), c.RotEnd}) {
		t.Errorf("curve from its second tile: next %+v ok %v", next, ok)
	}
	if _, _, ok := NextAndPreviousTile(false, 0, 3, 0, 0, 0, 0); ok {
		t.Error("a sequence index the piece doesn't have was taken")
	}
}
