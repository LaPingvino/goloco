package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// RoadElement::tick on the title game: an AI company's town-type road in use
// (unk7_40) stays; the same road unused is removed by its owner; a player's
// stays either way. updateYearly moves this year's flag to last year's.
func TestRoadTick(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	var tl *Tile
	var k, tx, ty int
	found := false
	for y := range s.Tiles {
		for x := range s.Tiles[y] {
			t := &s.Tiles[y][x]
			for i, r := range t.Order {
				if r.Kind != scenario.KindRoad {
					continue
				}
				e := &t.Roads[r.Index]
				if e.Owner == CompanyNeutral || s.IsPlayerCompany(int(e.Owner)) || s.RoadAnyTypeCompatible&(1<<e.RoadObjectID) == 0 ||
					e.SequenceIndex != 0 || e.HasLevelCrossing || e.Unk7_10 || e.Ghost || e.AiAllocated || e.HasStationElement || len(t.Roads) != 1 {
					continue
				}
				tl, k, tx, ty, found = t, i, x, y, true
				break
			}
			if found {
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Skip("no AI town-type road on the title map")
	}
	e := &tl.Roads[tl.Order[k].Index]
	owner := e.Owner
	e.Mods = roadUnk7_40
	if !s.updateRoadElement(tl, k, int32(tx)*32, int32(ty)*32, om) || len(tl.Roads) != 1 {
		t.Fatal("a road in use was removed")
	}
	s.tilesUpdateYearly()
	if e.Mods != roadUnk7_80 {
		t.Fatalf("after updateYearly: mods %#x, want unk7_80", e.Mods)
	}
	s.PlayerCompanies[0] = owner // a player's: kept though unused
	e.Mods = 0
	if !s.updateRoadElement(tl, k, int32(tx)*32, int32(ty)*32, om) || len(tl.Roads) != 1 {
		t.Fatal("a player's road was removed")
	}
	s.PlayerCompanies[0] = CompanyNeutral
	if s.updateRoadElement(tl, k, int32(tx)*32, int32(ty)*32, om) || len(tl.Roads) != 0 {
		t.Fatalf("an unused AI road stayed: %d roads", len(tl.Roads))
	}
	t.Logf("company %d's road at %d,%d removed", owner, tx, ty)
}
