package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// RemoveRoad on the title map: a plain straight is found and taken up (every
// tile of the piece), a ghost removal doesn't touch a real road, and without
// apply nothing changes.
func TestRemoveRoad(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, _ := om.G1File.(G1Raw)
	for y := 4; y < m.Height-4; y++ {
		for x := 4; x < m.Width-4; x++ {
			tl := m.TileAt(x, y)
			for _, r := range tl.Order {
				if r.Kind != scenario.KindRoad {
					continue
				}
				rd := tl.Roads[r.Index]
				if rd.RoadID != 0 || rd.HasStationElement || rd.HasLevelCrossing || rd.Ghost || rd.AiAllocated || rd.HasBridge {
					continue
				}
				c := &Cmd{Company: CompanyNeutral, Objects: om}
				a := RoadRemoval{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(rd.BaseZ) * 4, Rotation: rd.Rotation, RoadID: 0, Index: rd.SequenceIndex, RoadObjectID: rd.RoadObjectID}
				n := len(tl.Roads)
				if _, ok := m.RemoveRoad(c, CmdApply|CmdGhost, a, g1); ok {
					t.Fatal("a ghost removal took up a real road")
				}
				if _, ok := m.RemoveRoad(c, 0, a, g1); !ok || len(tl.Roads) != n {
					t.Fatalf("the check: ok %v, roads %d → %d", ok, n, len(tl.Roads))
				}
				if _, ok := m.RemoveRoad(c, CmdApply, a, g1); !ok {
					t.Fatal("removing a straight was refused")
				}
				if len(tl.Roads) != n-1 {
					t.Fatalf("roads on the tile: %d → %d", n, len(tl.Roads))
				}
				for _, o := range tl.Order {
					if o.Kind == scenario.KindRoad && int(o.Index) >= len(tl.Roads) {
						t.Fatal("the tile's order still points at the removed road")
					}
				}
				return
			}
		}
	}
	t.Skip("no plain straight road on the title map")
}
