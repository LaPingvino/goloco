package world

import (
	"os"
	"reflect"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// The construction tool's selected tiles reach the renderer while the tool
// is up (with unk_03's colour), and the tiles they leave and take are
// redrawn (mapInvalidateMapSelectionFreeFormTiles); off, none are.
func TestSyncConstructSelection(t *testing.T) {
	wld := newTestWorld(20, 20, 10)
	wld.scene = &scene.Renderer{}
	cx, cy := wld.Map.GameToScreen(5*32+16, 6*32+16, 40)
	wld.dirtyView = scene.View{X: int(cx) - 200, Y: int(cy) - 150, W: 400, H: 300}
	wld.dirtyGrid.Reset(400, 300, gfx8.InvalidationBlockWidth, gfx8.InvalidationBlockHeight)
	dirty := func() (n int) {
		wld.dirtyGrid.TraverseDirtyCells(func(int, int, int, int) { n++ })
		wld.dirtyGrid.Reset(400, 300, gfx8.InvalidationBlockWidth, gfx8.InvalidationBlockHeight)
		return n
	}
	wld.ConsSelection = sim.ConstructionSelection{On: true, Unk03: true, Tiles: [][2]int32{{5 * 32, 6 * 32}}}
	wld.SyncConstructionArrow(true)
	sel := wld.scene.ConstructSelection
	if sel == nil || !sel(5, 6) || sel(6, 6) || !wld.scene.ConstructSelectionUnk03 {
		t.Fatalf("the selection handed over: %v, unk_03 %v", sel != nil, wld.scene.ConstructSelectionUnk03)
	}
	if dirty() == 0 {
		t.Error("the selected tile wasn't redrawn")
	}
	wld.SyncConstructionArrow(true)
	if n := dirty(); n != 0 {
		t.Errorf("unchanged: %d cells redrawn", n)
	}
	wld.SyncConstructionArrow(false)
	if wld.scene.ConstructSelection != nil || dirty() == 0 {
		t.Errorf("off: selection %v, the tile not redrawn", wld.scene.ConstructSelection != nil)
	}
}

// dataWorld is the title game in a World with a renderer (game data needed).
func dataWorld(t *testing.T) *World {
	const data = "../../../locomotion"
	if _, err := os.Stat(data + "/Data/g1.DAT"); err != nil {
		t.Skip("game data not present")
	}
	g1, err := assets.LoadG1(data + "/Data/g1.DAT")
	if err != nil {
		t.Fatal(err)
	}
	om := objects.NewObjectManager(data + "/ObjData")
	om.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
	om.G1File = g1
	if err := om.LoadAllObjects(); err != nil {
		t.Fatal(err)
	}
	sc, err := scenario.LoadScenarioData(data + "/Data/title.dat")
	if err != nil {
		t.Skip("title.dat: ", err)
	}
	sim.ReorderObjectsForScenario(om, sc)
	wld := &World{renderer: &render.Renderer{G1: g1, ObjMgr: om}}
	wld.Map.LoadScenario(sc)
	return wld
}

// drawTrack / drawRoad: the preview paints the piece (a straight, a large
// curve onto the diagonal, two roads) on the target, and the map's tiles it
// borrowed are as they were.
func TestPaintPiecePreview(t *testing.T) {
	wld := dataWorld(t)
	var before [][]sim.Tile
	for y := 250; y <= 262; y++ {
		row := make([]sim.Tile, 0, 13)
		for x := 250; x <= 262; x++ {
			row = append(row, *wld.Map.TileAt(x, y))
		}
		before = append(before, row)
	}
	track, road := -1, -1
	for i, o := range wld.renderer.ObjMgr.TrackObjects {
		if o != nil {
			track = i
			break
		}
	}
	for i, o := range wld.renderer.ObjMgr.RoadObjects {
		if o != nil {
			road = i
			break
		}
	}
	for _, c := range []struct {
		road       bool
		obj, piece int
		direction  uint8
	}{{false, track, 0, 0}, {false, track, 10, 14}, {true, road, 0, 1}, {true, road, 3, 2}} {
		if c.obj < 0 {
			continue
		}
		rt := gfx8.NewRenderTarget(132, 100)
		wld.PaintPiecePreview(rt, c.road, c.piece, c.obj, c.direction, 0)
		n := 0
		for _, b := range rt.Bits {
			if b != 0 {
				n++
			}
		}
		if n < 200 {
			t.Errorf("road %v piece %d direction %d: %d pixels painted", c.road, c.piece, c.direction, n)
		}
	}
	for j, row := range before {
		for i := range row {
			if got := *wld.Map.TileAt(250+i, 250+j); !reflect.DeepEqual(got, row[i]) {
				t.Fatalf("tile (%d, %d) changed by the preview", 250+i, 250+j)
			}
		}
	}
}
