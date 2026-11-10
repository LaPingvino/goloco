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

// Growing with road branching adds junctions: over many rounds the title
// towns gain road pieces (straights or very small curves off their roads).
func TestGrowTownAddsJunctions(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	roads := func() (n int) {
		for y := range m.Tiles {
			for x := range m.Tiles[y] {
				n += len(m.Tiles[y][x].Roads)
			}
		}
		return
	}
	r0 := roads()
	for round := 0; round < 60; round++ {
		for i := range m.Towns {
			m.GrowTown(om, i, TownGrowAllowRoadBranching)
		}
	}
	r1 := roads()
	t.Logf("road pieces %d → %d", r0, r1)
	if r1 <= r0 {
		t.Error("no junction was added")
	}
}

// The edge table as extracted: a straight at rotation 0 reaches edges 0 and 2.
func TestRoadTadConnectionEdge(t *testing.T) {
	if RoadTadConnectionEdge[0] != 0b0101 || RoadTadConnectionEdge[1] != 0b1010 || RoadTadConnectionEdge[79] != 0 {
		t.Errorf("table %v…", RoadTadConnectionEdge[:4])
	}
}
