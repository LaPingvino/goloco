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
