package sim

import "testing"

func flatState(n int, baseZ uint8) *State {
	s := &State{Map: Map{Width: n, Height: n, Tiles: make([][]Tile, n)}}
	for y := range s.Tiles {
		s.Tiles[y] = make([]Tile, n)
		for x := range s.Tiles[y] {
			s.Tiles[y][x].BaseZ = baseZ
		}
	}
	return s
}

// testHead lays pieces one after another without the game commands
// (LayTrack / LayRoad): the lines the tests run on.
type testHead struct {
	s *State
	e ConnectionEnd // where the next piece starts
}

func headAt(s *State, tx, ty int, baseZ, rot uint8) *testHead {
	return &testHead{s, ConnectionEnd{X: int32(tx) * 32, Y: int32(ty) * 32, Z: int32(baseZ) * 4, NextRotation: rot}}
}

func (h *testHead) track(id int) bool {
	e, ok := h.s.LayTrack(h.e.X, h.e.Y, h.e.Z, h.e.NextRotation, id, 0)
	if ok {
		h.e = e
	}
	return ok
}

func (h *testHead) road(id int) bool {
	e, ok := h.s.LayRoad(h.e.X, h.e.Y, h.e.Z, h.e.NextRotation, id, 0)
	if ok {
		h.e = e
	}
	return ok
}

// Each road slope piece laid on flat ground, in every rotation, ends its DZ
// (RoadCoordinates) higher or lower.
func TestRoadSlopePiecesPlace(t *testing.T) {
	for _, id := range []int{5, 6, 7, 8} {
		for rot := uint8(0); rot < 4; rot++ {
			h := headAt(flatState(20, 16), 10, 10, 16, rot)
			if !h.road(id) {
				t.Errorf("road %d rot %d: can't be laid (coords %+v, tiles %v)", id, rot, RoadCoordinates[id<<3|int(rot)], RoadPieceTiles[id])
				continue
			}
			if want := 64 + int32(RoadCoordinates[id<<3|int(rot)].DZ); h.e.Z != want {
				t.Errorf("road %d rot %d: ends at z %d, want %d", id, rot, h.e.Z, want)
			}
		}
	}
}
