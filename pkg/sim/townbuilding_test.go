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

// A company's road beside town buildings is taken over by the town
// (neutralRoadTakeover): the owner becomes neutral, the object kept.
func TestTownRoadTakeover(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	// A neutral straight with buildings near, made a company's.
	tx, ty, ri := -1, -1, -1
	for y := 4; y < m.Height-4 && tx < 0; y++ {
		for x := 4; x < m.Width-4; x++ {
			tl := m.TileAt(x, y)
			if len(tl.Roads) != 1 || tl.Roads[0].RoadID != 0 || tl.Roads[0].Owner != CompanyNeutral || tl.Roads[0].SequenceIndex != 0 || !m.hasNearbyBuildings(x, y) {
				continue
			}
			if ro := om.GetRoadObjectByIndex(int(tl.Roads[0].RoadObjectID)); ro == nil || ro.Flags&roadFlagAnyRoadTypeCompatible == 0 {
				continue
			}
			tx, ty, ri = x, y, 0
			break
		}
	}
	if tx < 0 {
		t.Skip("no neutral town straight near buildings")
	}
	el := &m.TileAt(tx, ty).Roads[ri]
	el.Owner = 3
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	rad := uint16(el.Rotation & 3)
	info := m.getRoadInformation(c, int32(tx)*32, int32(ty)*32, int32(el.BaseZ)*4, rad)
	if info.Owner != 3 {
		t.Fatalf("road information owner %d", info.Owner)
	}
	obj := el.RoadObjectID
	if !m.townRoadUpdate(om, info, int32(tx)*32, int32(ty)*32, int32(el.BaseZ)*4, rad, false, obj, 1, TownGrowNeutralRoadTakeover) {
		t.Fatal("the takeover didn't stop the walk")
	}
	if el.Owner != CompanyNeutral || el.RoadObjectID != obj {
		t.Errorf("after the takeover: owner %d object %d (was %d)", el.Owner, el.RoadObjectID, obj)
	}
}

// generateNewBuildingArgs draws with the CLOSEST town's prng (town->prng
// after getClosestTownAndDensity) and only then refuses when that isn't the
// growing town: asked for town A at town B's centre, B's prng moves, A's
// doesn't, and nothing is placed.
func TestNewBuildingArgsUsesClosestTownPrng(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if len(m.Towns) < 2 {
		t.Skip("fewer than two towns")
	}
	bx, by := int32(m.Towns[1].X), int32(m.Towns[1].Y)
	if ti, _, _ := m.ClosestTownAndDensity(bx, by); ti != 1 {
		t.Skip("town 1's centre isn't closest to town 1")
	}
	m.CurrentYear = 1990 // candidates at every density
	a0, b0 := m.Towns[0].Prng, m.Towns[1].Prng
	_, ok := m.newBuildingArgs(om, 0, bx, by, 0, 255, 0, false, false)
	if ok {
		t.Error("a building for town 0 at town 1's centre")
	}
	if m.Towns[0].Prng != a0 {
		t.Error("the growing town's prng moved")
	}
	if m.Towns[1].Prng == b0 {
		t.Error("the closest town's prng didn't move (no draws?)")
	}
}
