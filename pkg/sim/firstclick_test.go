package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The first click (onToolUpSingle + constructionLoop) builds the piece at the
// tile and leaves the head after it; a refused click turns the construction
// rotation by 90°.
func TestBuildTrackAtTile(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	c := &Cmd{Company: s.ControllingCompany, Objects: om}
	if len(om.TrackObjects) == 0 || om.TrackObjects[0] == nil {
		t.Skip("no track object")
	}
	for y := 8; y < s.Height-8; y++ {
		for x := 8; x < s.Width-8; x++ {
			tl := s.TileAt(x, y)
			if len(tl.Order) != 1 || tl.Slope != 0 || tl.WaterLevel != 0 {
				continue
			}
			s.ConsHead = ConstructionHead{Rotation: 0}
			if _, ok := s.BuildTrackAtTile(c, x, y, 0, 0); !ok {
				continue // money, ownership, clearance …: another tile
			}
			tracks := 0
			for _, r := range tl.Order {
				if r.Kind == scenario.KindTrack && !tl.Tracks[r.Index].Ghost {
					tracks++
				}
			}
			if tracks != 1 {
				t.Fatalf("real track elements on the clicked tile: %d", tracks)
			}
			h := s.ConsHead
			if !h.Active || (h.X == x && h.Y == y) {
				t.Fatalf("the head after the first piece: %+v (clicked %d,%d)", h, x, y)
			}
			// A click that can't build (off the map's drawable area) turns the rotation.
			s.ConsHead.Rotation = 1
			if _, ok := s.BuildTrackAtTile(c, 0, 0, 0, 0); ok {
				t.Fatal("built at the map's edge")
			}
			if s.ConsHead.Rotation != 2 {
				t.Errorf("after a refused click the rotation is %d, want 2", s.ConsHead.Rotation)
			}
			return
		}
	}
	t.Skip("no tile took a straight")
}

// The road tool as the track tool: the hover ghost (UpdateRoadGhost) stands
// as a ghost road on the tile, and the first click (BuildRoadAtTile) builds a
// real piece there and leaves the head after it; a refused click turns the
// rotation.
func TestBuildRoadAtTile(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	c := &Cmd{Company: s.ControllingCompany, Objects: om}
	obj := -1
	for i, o := range om.RoadObjects {
		if o != nil {
			obj = i
			break
		}
	}
	if obj < 0 {
		t.Skip("no road object")
	}
	for y := 8; y < s.Height-8; y++ {
		for x := 8; x < s.Width-8; x++ {
			tl := s.TileAt(x, y)
			if len(tl.Order) != 1 || tl.Slope != 0 || tl.WaterLevel != 0 {
				continue
			}
			if _, ok := s.UpdateRoadGhost(c, x, y, 0, 0, uint8(obj)); !ok {
				continue
			}
			ghosts := 0
			for _, r := range tl.Order {
				if r.Kind == scenario.KindRoad && tl.Roads[r.Index].Ghost {
					ghosts++
				}
			}
			if ghosts != 1 || s.RoadGhost == nil {
				t.Fatalf("the hover ghost on %d,%d: %d ghost roads, remembered %v", x, y, ghosts, s.RoadGhost != nil)
			}
			s.RemoveRoadGhost(c)
			s.ConsHead = ConstructionHead{Rotation: 0}
			if _, ok := s.BuildRoadAtTile(c, x, y, 0, uint8(obj)); !ok {
				continue // money, clearance …: another tile
			}
			real := 0
			for _, r := range tl.Order {
				if r.Kind == scenario.KindRoad && !tl.Roads[r.Index].Ghost {
					real++
				}
			}
			if real != 1 {
				t.Fatalf("real road elements on the clicked tile: %d", real)
			}
			if h := s.ConsHead; !h.Active || (h.X == x && h.Y == y) {
				t.Fatalf("the head after the first piece: %+v (clicked %d,%d)", h, x, y)
			}
			s.ConsHead.Rotation = 1
			if _, ok := s.BuildRoadAtTile(c, 0, 0, 0, uint8(obj)); ok {
				t.Fatal("built at the map's edge")
			}
			if s.ConsHead.Rotation != 2 {
				t.Errorf("after a refused click the rotation is %d, want 2", s.ConsHead.Rotation)
			}
			return
		}
	}
	t.Skip("no tile took a road")
}
