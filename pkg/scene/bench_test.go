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
	om.ResolveCliffEdges()
	var m sim.Map
	m.LoadScenario(sc)
	r := New(g1, om)
	cx, cy := paint.GameToScreen(paint.Pos3{X: 169*32 + 16, Y: 150*32 + 16}, 0)
	v := View{X: cx - 400, Y: cy - 300, W: 800, H: 600}
	r.ToRGBA(r.Render(&m, v)) // warm the sprite cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ToRGBA(r.Render(&m, v))
	}
}
