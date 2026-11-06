package scene

import (
	"os"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// BenchmarkRenderTitleView renders an 800×600 view of the title map, the work
// the game does per frame in scene mode (CPU render + palette conversion).
func BenchmarkRenderTitleView(b *testing.B) {
	r, m, v := titleView(b)
	r.ToRGBA(r.Render(m, v)) // warm the sprite cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ToRGBA(r.Render(m, v))
	}
}

// TestRenderSteadyStateAllocs: once the sprite cache is warm, a frame must not
// allocate (the render loop runs every frame in the browser).
func TestRenderSteadyStateAllocs(t *testing.T) {
	r, m, v := titleView(t)
	r.ToRGBA(r.Render(m, v))
	n := testing.AllocsPerRun(5, func() { r.ToRGBA(r.Render(m, v)) })
	t.Logf("allocations per frame: %.0f", n)
	if n > 16 {
		t.Errorf("a warm frame allocates %.0f times; want ~0", n)
	}
}

func titleView(b testing.TB) (*Renderer, *sim.Map, View) {
	return saveView(b, "Data/title.dat", true)
}

// saveView loads a save or scenario from the game data directory (path
// relative to it) the way the game does; title marks title.dat.
func saveView(b testing.TB, path string, title bool) (*Renderer, *sim.Map, View) {
	const data = "../../../locomotion"
	if _, err := os.Stat(data + "/Data/g1.DAT"); err != nil {
		b.Skip("game data not present")
	}
	g1, err := assets.LoadG1(data + "/Data/g1.DAT")
	if err != nil {
		b.Fatal(err)
	}
	om := objects.NewObjectManager(data + "/ObjData")
	om.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
	om.G1File = g1
	if err := om.LoadAllObjects(); err != nil {
		b.Fatal(err)
	}
	if _, err := os.Stat(data + "/" + path); err != nil {
		b.Skip(path + " not present")
	}
	sc, err := scenario.LoadScenarioData(data + "/" + path)
	if err != nil {
		b.Fatal(err)
	}
	sim.ReorderObjectsForScenario(om, sc) // the game's own list (cargo and vehicles were missing here)
	om.ResolveCliffEdges()
	var m sim.Map
	m.LoadScenario(sc)
	m.TitleMode = title // title.dat runs as the title screen
	r := New(g1, om)
	cx, cy := paint.GameToScreen(paint.Pos3{X: 169*32 + 16, Y: 150*32 + 16}, 0)
	return r, &m, View{X: cx - 400, Y: cy - 300, W: 800, H: 600}
}

// The tram track's road extra (RDEXCAT1, overhead wires) is drawn: a view of
// a tram street at tile (225, 246) changes when the extra's image is taken away.
func TestRoadExtrasDrawn(t *testing.T) {
	r, m, _ := titleView(t)
	cx, cy := paint.GameToScreen(paint.Pos3{X: 225*32 + 16, Y: 246*32 + 16}, 0)
	v := View{X: cx - 200, Y: cy - 150, W: 400, H: 300}
	with := append([]uint8(nil), r.Render(m, v).Bits...)
	ex := r.Objects.Extra("RDEXCAT1")
	if ex == nil {
		t.Skip("RDEXCAT1 not loaded")
	}
	img := ex.Image
	ex.Image = 0
	defer func() { ex.Image = img }()
	without := r.Render(m, v).Bits
	diff := 0
	for i := range with {
		if with[i] != without[i] {
			diff++
		}
	}
	t.Logf("%d pixels differ", diff)
	if diff == 0 {
		t.Error("the overhead wires change nothing in the view")
	}
}

// Station cargo is drawn on the platforms: a view of the first straight
// station tile with waiting cargo changes when every station's cargo is cleared.
func TestStationCargoDrawn(t *testing.T) {
	r, m, _ := titleView(t)
	tx, ty, found := -1, -1, 0
	for y := range m.Height {
		for x := range m.Width {
			t0 := m.TileAt(x, y)
			for _, st := range t0.Stations {
				if st.StationType != 0 || int(st.StationID) >= len(m.StationNames) {
					continue
				}
				d := m.StationCargoDensities(int(st.StationID))
				sum := 0
				for _, v := range d {
					sum += int(v)
				}
				if sum > 0 && len(t0.Tracks) > 0 && t0.Tracks[0].TrackID == 0 {
					found++
					if tx < 0 {
						tx, ty = x*32+16, y*32+16
						t.Logf("station %d at tile (%d, %d): density %v", st.StationID, x, y, d)
					}
				}
			}
		}
	}
	t.Logf("%d straight station tiles with waiting cargo", found)
	if tx < 0 {
		t.Skip("no station with waiting cargo")
	}
	cx, cy := paint.GameToScreen(paint.Pos3{X: tx, Y: ty}, 0)
	v := View{X: cx - 300, Y: cy - 200, W: 600, H: 400}
	with := append([]uint8(nil), r.Render(m, v).Bits...)
	saved := append([]scenario.StationName(nil), m.StationNames...)
	for i := range m.StationNames {
		for c := range m.StationNames[i].Cargo {
			m.StationNames[i].Cargo[c].DensityPerTile = 0
		}
	}
	defer func() { m.StationNames = saved }()
	without := r.Render(m, v).Bits
	diff := 0
	for i := range with {
		if with[i] != without[i] {
			diff++
		}
	}
	t.Logf("%d pixels differ", diff)
	if diff == 0 {
		t.Error("station cargo changes nothing in the view")
	}
}

// Road station cargo: the first road station tile (straight road) with waiting
// cargo that has platform images changes when the cargo is cleared.
func TestRoadStationCargoDrawn(t *testing.T) {
	r, m, _ := titleView(t)
	tx, ty := -1, -1
	for y := range m.Height {
		for x := range m.Width {
			t0 := m.TileAt(x, y)
			for _, st := range t0.Stations {
				if st.StationType != 1 || int(st.StationID) >= len(m.StationNames) || len(t0.Roads) == 0 || t0.Roads[0].RoadID != 0 {
					continue
				}
				for c, v := range m.StationCargoDensities(int(st.StationID)) {
					// cargo that has platform images (e.g. grapes have none)
					if v != 0 && tx < 0 && c < len(r.Objects.CargoObjects) && r.Objects.CargoObjects[c] != nil && r.Objects.CargoObjects[c].NumPlatformVariations > 0 {
						tx, ty = x*32+16, y*32+16
						t.Logf("road station %d at tile (%d, %d)", st.StationID, x, y)
					}
				}
			}
		}
	}
	if tx < 0 {
		t.Skip("no road station with waiting cargo")
	}
	cx, cy := paint.GameToScreen(paint.Pos3{X: tx, Y: ty}, 0)
	v := View{X: cx - 300, Y: cy - 200, W: 600, H: 400}
	with := append([]uint8(nil), r.Render(m, v).Bits...)
	saved := append([]scenario.StationName(nil), m.StationNames...)
	for i := range m.StationNames {
		for c := range m.StationNames[i].Cargo {
			m.StationNames[i].Cargo[c].DensityPerTile = 0
		}
	}
	defer func() { m.StationNames = saved }()
	without := r.Render(m, v).Bits
	diff := 0
	for i := range with {
		if with[i] != without[i] {
			diff++
		}
	}
	t.Logf("%d pixels differ", diff)
	if diff == 0 {
		t.Error("road station cargo changes nothing in the view")
	}
}

// Buildings under construction draw their built sections only: the title
// map's unconstructed buildings change the view when marked constructed.
func TestBuildingsUnderConstruction(t *testing.T) {
	r, m, _ := titleView(t)
	tx, ty, n := -1, -1, 0
	for y := range m.Height {
		for x := range m.Width {
			for _, b := range m.TileAt(x, y).Buildings {
				if !b.IsConstructed {
					n++
					if tx < 0 {
						tx, ty = x, y
					}
				}
			}
		}
	}
	t.Logf("%d building elements under construction", n)
	if tx < 0 {
		t.Skip("none under construction")
	}
	cx, cy := paint.GameToScreen(paint.Pos3{X: tx*32 + 16, Y: ty*32 + 16, Z: int(m.TileAt(tx, ty).BaseZ) * 4}, 0)
	v := View{X: cx - 200, Y: cy - 200, W: 400, H: 300}
	with := append([]uint8(nil), r.Render(m, v).Bits...)
	bs := m.TileAt(tx, ty).Buildings
	saved := append(bs[:0:0], bs...)
	for i := range bs {
		bs[i].IsConstructed = true
	}
	defer copy(bs, saved)
	without := r.Render(m, v).Bits
	diff := 0
	for i := range with {
		if with[i] != without[i] {
			diff++
		}
	}
	t.Logf("tile (%d, %d): %d pixels differ", tx, ty, diff)
	if diff == 0 {
		t.Error("an unconstructed building draws like a finished one")
	}
}

// Waves are drawn on a water tile with a running wave (flag 6) at zoom 0:
// the view over the title map's lake changes when its middle tile gets a
// wave, and again when the wave's frame moves on (PaintSurface's +60 + frame).
func TestWavesDrawn(t *testing.T) {
	r, m, _ := titleView(t)
	tx, ty := 86, 8
	tl := m.TileAt(tx, ty)
	if tl == nil || tl.WaterLevel == 0 {
		t.Skip("no water at the lake tile")
	}
	cx, cy := paint.GameToScreen(paint.Pos3{X: tx*32 + 16, Y: ty*32 + 16, Z: int(tl.WaterLevel) * 16}, 0)
	v := View{X: cx - 64, Y: cy - 48, W: 128, H: 96}
	none := append([]uint8(nil), r.Render(m, v).Bits...)
	frame := uint8(3)
	r.WaveFrame = func(int, int) uint8 { return frame }
	tl.WaveFlag = true
	defer func() { tl.WaveFlag = false; r.WaveFrame = nil }()
	with := append([]uint8(nil), r.Render(m, v).Bits...)
	frame = 9
	later := r.Render(m, v).Bits
	diff := func(a, b []uint8) int {
		n := 0
		for i := range a {
			if a[i] != b[i] {
				n++
			}
		}
		return n
	}
	t.Logf("a wave changes %d pixels; frame 3 → 9 changes %d", diff(none, with), diff(with, later))
	if diff(none, with) == 0 || diff(with, later) == 0 {
		t.Error("the wave isn't drawn, or doesn't animate")
	}
}
