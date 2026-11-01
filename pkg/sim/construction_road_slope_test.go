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

// Each road slope piece must be buildable from a head on flat ground, in every
// rotation, and move the head up/down by its DZ (RoadCoordinates).
func TestRoadSlopePiecesPlace(t *testing.T) {
	for _, id := range []int{5, 6, 7, 8} {
		for rot := uint8(0); rot < 4; rot++ {
			s := flatState(20, 16)
			s.ConsHead = ConstructionHead{Active: true, X: 10, Y: 10, BaseZ: 16, Rotation: rot}
			if !s.CanPlaceRoadAtHead(id) {
				c := RoadCoordinates[id<<3|int(rot)]
				t.Errorf("road %d rot %d: cannot place (coords %+v, tiles %v)", id, rot, c, RoadPieceTiles[id])
				continue
			}
			before := s.ConsHead.BaseZ
			if !s.PlaceRoadAtHead(id, 0) {
				t.Errorf("road %d rot %d: place failed", id, rot)
				continue
			}
			dz := int(RoadCoordinates[id<<3|int(rot)].DZ) / 4
			if int(s.ConsHead.BaseZ) != int(before)+dz {
				t.Errorf("road %d rot %d: head z %d, want %d", id, rot, s.ConsHead.BaseZ, int(before)+dz)
			}
		}
	}
}
