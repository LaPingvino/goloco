package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A small town building on an empty flat tile beside a town: the dry run
// changes nothing and costs what applying does; the element stands with all
// quarters occupied, the town gains a building and the capacity; removing it
// takes them back.
func TestCreateBuilding(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	obj := -1
	for i, b := range om.BuildingObjects {
		if b != nil && b.Flags&(objects.BuildingFlagLargeTile|objects.BuildingFlagMiscBuilding|objects.BuildingFlagIsHeadquarters) == 0 && b.ProducedQuantity[0] > 0 && len(b.GetBuildingParts(0)) > 0 {
			obj = i
			break
		}
	}
	if obj < 0 {
		t.Skip("no small town building object")
	}
	tw := m.Towns[0]
	x, y := -1, -1
	for r := 2; r < 12 && x < 0; r++ {
		for dy := -r; dy <= r && x < 0; dy++ {
			for dx := -r; dx <= r; dx++ {
				tx, ty := int(tw.X)/32+dx, int(tw.Y)/32+dy
				if tl := m.TileAt(tx, ty); tl != nil && len(tl.Order) == 1 && tl.Slope == 0 && tl.WaterLevel == 0 && !tl.IsIndustrial {
					x, y = tx, ty
					break
				}
			}
		}
	}
	if x < 0 {
		t.Skip("no empty tile near town 0")
	}
	tl := m.TileAt(x, y)
	town, _, _ := m.ClosestTownAndDensity(int32(x)*32, int32(y)*32)
	before := m.Towns[town]
	a := BuildingPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(tl.BaseZ) * 4, Type: uint8(obj), BuildImmediately: true}

	dry, ok := m.CreateBuilding(c, 0, a)
	if !ok || len(tl.Buildings) != 0 {
		t.Fatalf("dry run: ok %v, buildings %d", ok, len(tl.Buildings))
	}
	cost, ok := m.CreateBuilding(c, CmdApply, a)
	if !ok || cost != dry {
		t.Fatalf("apply: ok %v cost %d (dry %d)", ok, cost, dry)
	}
	if len(tl.Buildings) != 1 || !tl.Buildings[0].IsConstructed || tl.Buildings[0].ClearZ <= tl.Buildings[0].BaseZ {
		t.Fatalf("buildings %+v", tl.Buildings)
	}
	for _, r := range tl.Order {
		if r.Kind == scenario.KindBuilding && r.Flags&0xF != 0xF {
			t.Errorf("the building's ref has quarters %#x, want 0xF", r.Flags&0xF)
		}
	}
	after := m.Towns[town]
	capacity := uint32(om.BuildingObjects[obj].ProducedQuantity[0])
	if after.NumBuildings != before.NumBuildings+1 || after.PopulationCapacity != before.PopulationCapacity+capacity || after.Population != before.Population+capacity {
		t.Errorf("town: buildings %d → %d, capacity %d → %d, population %d → %d (+%d)", before.NumBuildings, after.NumBuildings, before.PopulationCapacity, after.PopulationCapacity, before.Population, after.Population, capacity)
	}
	// Built again on the same tile: the building in the way is cleared.
	if _, ok := m.CreateBuilding(c, CmdApply|CmdPreventBuildingClearing, a); ok {
		t.Error("preventBuildingClearing still cleared the building")
	}
	if _, ok := m.RemoveBuilding(c, a.X, a.Y, a.Z, CmdApply); !ok || len(tl.Buildings) != 0 {
		t.Fatalf("removal: ok %v, buildings %d", ok, len(tl.Buildings))
	}
	if m.Towns[town].NumBuildings != before.NumBuildings || m.Towns[town].PopulationCapacity != before.PopulationCapacity {
		t.Errorf("after removal: buildings %d (was %d), capacity %d (was %d)", m.Towns[town].NumBuildings, before.NumBuildings, m.Towns[town].PopulationCapacity, before.PopulationCapacity)
	}
}
