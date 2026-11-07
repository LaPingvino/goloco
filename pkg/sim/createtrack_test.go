package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// createTrack's start: an extra that isn't loaded is dropped, a height off
// the 16 grid is refused, and the cost is the track's price times the
// piece's cost factor / 256 (the curve's factor).
func TestCreateTrackCost(t *testing.T) {
	h := &objects.ObjectHeader{}
	copy(h.Name[:], "TRACKST ")
	track := &objects.TrackObject{Header: h, BuildCostFactor: 100, CostIndex: 0, ModNames: []string{"OVERHEAD"}}
	c := &Cmd{Company: 0, Objects: &objects.ObjectManager{TrackObjects: []*objects.TrackObject{track}}}
	m := &Map{}
	base := m.InflationAdjustedCost(100, 0, 10)
	mods, cost, ok := m.createTrackCost(c, trackPlacementArgs{Z: 32, Mods: 1})
	if !ok || mods != 0 {
		t.Errorf("mods %b ok %v", mods, ok)
	}
	if want := base * int32(TrackCostFactor[0]) / 256; cost != want {
		t.Errorf("straight cost %d, want %d", cost, want)
	}
	if _, _, ok := m.createTrackCost(c, trackPlacementArgs{Z: 40}); ok {
		t.Error("a height off the 16 grid must fail")
	}
	_, curve, _ := m.createTrackCost(c, trackPlacementArgs{Z: 32, TrackID: 6})
	if want := base * int32(TrackCostFactor[6]) / 256; curve != want {
		t.Errorf("curve cost %d, want %d", curve, want)
	}
	if _, _, ok := m.createTrackCost(c, trackPlacementArgs{TrackObjectID: 3}); ok {
		t.Error("an empty slot must fail")
	}
}

// createTrack's checks: a flat straight on the ground passes with its cost;
// raised without a bridge, or below 16, it fails; a long curve covers its
// tiles.
func TestCreateTrackCheck(t *testing.T) {
	h := &objects.ObjectHeader{}
	copy(h.Name[:], "TRACKST ")
	track := &objects.TrackObject{Header: h, BuildCostFactor: 100}
	c := &Cmd{Company: 0, Objects: &objects.ObjectManager{TrackObjects: []*objects.TrackObject{track}}}
	m := &Map{Width: 8, Height: 8, Tiles: make([][]Tile, 8)}
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, 8)
		for x := range m.Tiles[y] {
			m.Tiles[y][x].BaseZ = 4 // the ground at z 16
		}
	}
	a := trackPlacementArgs{X: 3 * 32, Y: 3 * 32, Z: 16, TrackID: 0, Bridge: 0xFF}
	r, ok := m.createTrackCheck(c, a, 0)
	_, base, _ := m.createTrackCost(c, a)
	if !ok || len(r.Pieces) != 1 || r.Pieces[0].BaseZ != 4 || r.Cost != base {
		t.Fatalf("flat straight: ok %v, %+v", ok, r)
	}
	a.Z = 32
	if _, ok := m.createTrackCheck(c, a, 0); ok {
		t.Error("raised without a bridge: must fail")
	}
	a.Z = 8
	if _, ok := m.createTrackCheck(c, a, 0); ok {
		t.Error("below 16: too low")
	}
	a = trackPlacementArgs{X: 4 * 32, Y: 4 * 32, Z: 16, TrackID: 1, Bridge: 0xFF} // the diagonal piece: several tiles
	if r, ok := m.createTrackCheck(c, a, 0); !ok || len(r.Pieces) != len(TrackPieces[1]) {
		t.Errorf("a multi-tile piece: ok %v, %d pieces", ok, len(r.Pieces))
	}
}

// createTrack whole: a straight is placed (owner, last flag, the ref's
// quarters) and costs what the check said; the same piece again is refused
// (clearTrack: already built); a ghost isn't marked as owned walls-removal.
func TestCreateTrack(t *testing.T) {
	h := &objects.ObjectHeader{}
	copy(h.Name[:], "TRACKST ")
	track := &objects.TrackObject{Header: h, BuildCostFactor: 100}
	c := &Cmd{Company: 2, Objects: &objects.ObjectManager{TrackObjects: []*objects.TrackObject{track}}}
	m := &Map{Width: 8, Height: 8, Tiles: make([][]Tile, 8)}
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, 8)
		for x := range m.Tiles[y] {
			m.Tiles[y][x].BaseZ = 4
		}
	}
	a := trackPlacementArgs{X: 3 * 32, Y: 3 * 32, Z: 16, TrackID: 0, Bridge: 0xFF}
	check, _ := m.createTrackCheck(c, a, 0)
	cost, ok := m.createTrack(c, a, CmdApply)
	tile := m.TileAt(3, 3)
	if !ok || cost != check.Cost || len(tile.Tracks) != 1 {
		t.Fatalf("placed: ok %v cost %d tracks %d", ok, cost, len(tile.Tracks))
	}
	e := tile.Tracks[0]
	if e.Owner != 2 || !e.IsLast || e.BaseZ != 4 || e.Ghost {
		t.Errorf("the element: %+v", e)
	}
	if _, ok := m.createTrack(c, a, CmdApply); ok || len(tile.Tracks) != 1 {
		t.Error("the same piece again was built")
	}
	g := trackPlacementArgs{X: 5 * 32, Y: 5 * 32, Z: 16, TrackID: 0, Bridge: 0xFF}
	if _, ok := m.createTrack(c, g, CmdApply|CmdGhost); !ok || !m.TileAt(5, 5).Tracks[0].Ghost {
		t.Error("a ghost piece")
	}
	if _, ok := m.createTrack(c, trackPlacementArgs{X: 6 * 32, Y: 6 * 32, Z: 16, Bridge: 0xFF}, 0); !ok || len(m.TileAt(6, 6).Tracks) != 0 {
		t.Error("without apply nothing is placed")
	}
}
