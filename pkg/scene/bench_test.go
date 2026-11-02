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
	sc, err := scenario.LoadScenarioData(data + "/Data/title.dat")
	if err != nil {
		b.Fatal(err)
	}
	om.ReorderLandObjects(sc.LandObjectOrder)
	om.ReorderTreeObjects(sc.TreeObjectOrder)
	om.ReorderBuildingObjects(sc.BuildingObjectOrder)
	om.ReorderWallObjects(sc.WallObjectOrder)
	om.ReorderRoadObjects(sc.RoadObjectOrder)
	om.ReorderTrackObjects(sc.TrackObjectOrder)
	om.ReorderTrainSignalObjects(sc.TrainSignalObjectOrder)
	om.ReorderTrainStationObjects(sc.TrainStationObjectOrder)
	om.ReorderIndustryObjects(sc.IndustryObjectOrder)
	om.ReorderRoadStationObjects(sc.RoadStationObjectOrder)
	om.ReorderAirportObjects(sc.AirportObjectOrder)
	om.ReorderBridgeObjects(sc.BridgeObjectOrder)
	om.ReorderDockObjects(sc.DockObjectOrder)
	om.ResolveCliffEdges()
	var m sim.Map
	m.LoadScenario(sc)
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
				if st.StationType != 0 || int(st.StationID) >= len(m.StationCargoDensity) {
					continue
				}
				d := m.StationCargoDensity[st.StationID]
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
	saved := m.StationCargoDensity
	m.StationCargoDensity = make([][32]uint8, len(saved))
	defer func() { m.StationCargoDensity = saved }()
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
