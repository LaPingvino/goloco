package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Growing the title map's towns with constructBuildings puts new buildings
// beside their roads: the tiles gain building elements of town building
// objects and the towns count them; a road beside them is untouched.
func TestGrowTownConstructsBuildings(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	countTiles := func() (n int) {
		for y := range m.Tiles {
			for x := range m.Tiles[y] {
				for _, r := range m.Tiles[y][x].Order {
					if r.Kind == scenario.KindBuilding && m.Tiles[y][x].Buildings[r.Index].SequenceIndex == 0 {
						n++
					}
				}
			}
		}
		return
	}
	numBuildings := func() (n int) {
		for _, tw := range m.Towns {
			n += int(tw.NumBuildings)
		}
		return
	}
	roads := func() (n int) {
		for y := range m.Tiles {
			for x := range m.Tiles[y] {
				n += len(m.Tiles[y][x].Roads)
			}
		}
		return
	}
	b0, n0, r0 := countTiles(), numBuildings(), roads()
	for round := 0; round < 20; round++ {
		for i := range m.Towns {
			m.GrowTown(om, i, TownGrowConstructBuildings|TownGrowBuildImmediately)
		}
	}
	b1, n1, r1 := countTiles(), numBuildings(), roads()
	t.Logf("buildings on the map %d → %d, counted by the towns %d → %d, roads %d → %d", b0, b1, n0, n1, r0, r1)
	if b1 <= b0 {
		t.Fatal("no building was constructed")
	}
	if n1-n0 != b1-b0 {
		t.Errorf("the towns counted %d new buildings, the map has %d", n1-n0, b1-b0)
	}
	if r1 != r0 {
		t.Errorf("roads changed %d → %d without road flags", r0, r1)
	}
}
