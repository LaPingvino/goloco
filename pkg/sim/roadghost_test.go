package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The player's road goes in through createRoad (LOOP 5.9g): its element ref
// carries the occupied quarters the clearance checks read, it costs money,
// and the head moves past the piece.
func TestBuildRoadAtHead(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	objID := -1
	for i, r := range om.RoadObjects {
		if r != nil {
			objID = i
			break
		}
	}
	if objID < 0 {
		t.Skip("no road object")
	}
	c := &Cmd{Company: s.ControllingCompany, Objects: om}
	for y := 8; y < s.Height-8; y++ {
		for x := 8; x < s.Width-8; x++ {
			tl := s.TileAt(x, y)
			if len(tl.Order) != 1 || tl.Slope != 0 || tl.WaterLevel != 0 {
				continue
			}
			s.ConsHead = ConstructionHead{Active: true, X: x, Y: y, BaseZ: tl.BaseZ}
			cost, ok := s.BuildRoadAtHead(c, 0, uint8(objID))
			if !ok {
				continue
			}
			if cost <= 0 {
				t.Errorf("a straight cost %d", cost)
			}
			for _, r := range tl.Order {
				if r.Kind == scenario.KindRoad {
					if r.Flags&0xF == 0 {
						t.Errorf("the road's ref has no occupied quarters (flags %#x)", r.Flags)
					}
					if tl.Roads[r.Index].Ghost {
						t.Error("the built road is a ghost")
					}
				}
			}
			if s.ConsHead.X == x && s.ConsHead.Y == y {
				t.Error("the head didn't move past the piece")
			}
			return
		}
	}
	t.Skip("no tile took a road")
}

// The road ghost at the head: placed as a ghost road element on an empty flat
// tile, kept while nothing changes, and taken up again by RemoveRoadGhost.
func TestRoadGhostAtHead(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	objID := -1
	for i, r := range om.RoadObjects {
		if r != nil {
			objID = i
			break
		}
	}
	if objID < 0 {
		t.Skip("no road object")
	}
	c := &Cmd{Company: s.ControllingCompany, Objects: om}
	for y := 8; y < s.Height-8; y++ {
		for x := 8; x < s.Width-8; x++ {
			tl := s.TileAt(x, y)
			if len(tl.Order) != 1 || tl.Slope != 0 || tl.WaterLevel != 0 {
				continue
			}
			s.ConsHead = ConstructionHead{Active: true, X: x, Y: y, BaseZ: tl.BaseZ, Rotation: 0}
			s.UpdateHeadRoadGhost(c, 0, uint8(objID))
			if s.RoadGhost == nil {
				continue // refused here (owner, clearance …): try another tile
			}
			ghosts := 0
			for _, r := range tl.Roads {
				if r.Ghost {
					ghosts++
				}
			}
			if ghosts != 1 {
				t.Fatalf("ghost road elements on the tile: %d", ghosts)
			}
			g := s.RoadGhost
			s.UpdateHeadRoadGhost(c, 0, uint8(objID))
			if s.RoadGhost != g {
				t.Error("an unchanged head placed the ghost again")
			}
			s.RemoveRoadGhost(c)
			if s.RoadGhost != nil || len(tl.Roads) != 0 || len(tl.Order) != 1 {
				t.Fatalf("after removal: ghost %v, roads %d, order %d", s.RoadGhost, len(tl.Roads), len(tl.Order))
			}
			for _, r := range tl.Order {
				if r.Kind == scenario.KindRoad {
					t.Fatal("a road ref is left")
				}
			}
			return
		}
	}
	t.Skip("no tile took a road ghost")
}
