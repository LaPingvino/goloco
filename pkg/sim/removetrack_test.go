package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// removeTrack: a built piece comes up for its sell price (the query leaves
// it), another company's is refused, a ghost is only found with the ghost
// flag (and by its own company), and a multi-tile piece goes whole from any
// of its tiles.
func TestRemoveTrack(t *testing.T) {
	h := &objects.ObjectHeader{}
	copy(h.Name[:], "TRACKST ")
	track := &objects.TrackObject{Header: h, BuildCostFactor: 100, SellCostFactor: -80}
	c := &Cmd{Company: 2, Objects: &objects.ObjectManager{TrackObjects: []*objects.TrackObject{track}}}
	m := &Map{Width: 12, Height: 12, Tiles: make([][]Tile, 12)}
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, 12)
		for x := range m.Tiles[y] {
			m.Tiles[y][x].BaseZ = 4
		}
	}
	a := trackPlacementArgs{X: 3 * 32, Y: 3 * 32, Z: 16, TrackID: 0, Bridge: 0xFF}
	if _, ok := m.createTrack(c, a, CmdApply); !ok {
		t.Fatal("build")
	}
	r := TrackRemoval{X: a.X, Y: a.Y, Z: a.Z, TrackID: 0}
	want := m.InflationAdjustedCost(-80, 0, 10) * int32(TrackCostFactor[0]) / 256
	if cost, ok := m.RemoveTrack(c, 0, r, nil); !ok || cost != want || len(m.TileAt(3, 3).Tracks) != 1 {
		t.Fatalf("query: %d %v (want %d)", cost, ok, want)
	}
	if _, ok := m.RemoveTrack(&Cmd{Company: 3, Objects: c.Objects}, CmdApply, r, nil); ok {
		t.Fatal("another company removed it")
	}
	if _, ok := m.RemoveTrack(c, CmdApply|CmdGhost, r, nil); ok {
		t.Fatal("a real piece found as a ghost")
	}
	if _, ok := m.RemoveTrack(c, CmdApply, r, nil); !ok || len(m.TileAt(3, 3).Tracks) != 0 {
		t.Fatal("not removed")
	}
	g := trackPlacementArgs{X: 6 * 32, Y: 6 * 32, Z: 16, TrackID: 0, Bridge: 0xFF}
	m.createTrack(c, g, CmdApply|CmdGhost)
	gr := TrackRemoval{X: g.X, Y: g.Y, Z: g.Z, TrackID: 0}
	if _, ok := m.RemoveTrack(c, CmdApply, gr, nil); ok {
		t.Fatal("a ghost found without the ghost flag")
	}
	if _, ok := m.RemoveTrack(&Cmd{Company: 3, Objects: c.Objects}, CmdApply|CmdGhost, gr, nil); ok {
		t.Fatal("another company's ghost removed")
	}
	if _, ok := m.RemoveTrack(c, CmdApply|CmdGhost, gr, nil); !ok || len(m.TileAt(6, 6).Tracks) != 0 {
		t.Fatal("ghost not removed")
	}
	// A multi-tile piece (the first with more than one tile), removed by its last tile.
	id := -1
	for i, ps := range TrackPieces {
		if len(ps) > 1 && TrackPieces[i][0].Z == 0 {
			id = i
			break
		}
	}
	b := trackPlacementArgs{X: 6 * 32, Y: 6 * 32, Z: 16, TrackID: uint8(id), Bridge: 0xFF}
	if _, ok := m.createTrack(c, b, CmdApply); !ok {
		t.Fatalf("build piece %d", id)
	}
	last := TrackPieces[id][len(TrackPieces[id])-1]
	ox, oy := rotOffset(last.X, last.Y, 0)
	br := TrackRemoval{X: b.X + int32(ox), Y: b.Y + int32(oy), Z: b.Z + int32(last.Z), TrackID: uint8(id), Index: last.Index}
	if _, ok := m.RemoveTrack(c, CmdApply, br, nil); !ok {
		t.Fatalf("piece %d not removed by its last tile", id)
	}
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			if len(m.Tiles[y][x].Tracks) != 0 {
				t.Errorf("track left at %d,%d", x, y)
			}
		}
	}
}
