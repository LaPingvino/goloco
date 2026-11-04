package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// loadDataMap loads a scenario from the local game data with its objects in
// the scenario's slot order (as cmd/render8); skips without the data.
func loadDataMap(t *testing.T, file string) (*Map, *objects.ObjectManager) {
	t.Helper()
	sc, err := scenario.LoadScenarioData("../../../locomotion/Data/" + file)
	if err != nil {
		t.Skip("game data not present")
	}
	om := objects.NewObjectManager("../../../locomotion/ObjData")
	if err := om.LoadAllObjects(); err != nil {
		t.Skip("objects not present")
	}
	for _, po := range sc.PackedObjectsRaw {
		_, _ = om.LoadObjectFromHeaderAndData(po.HeaderBytes, po.Data)
	}
	ReorderObjectsForScenario(om, sc)
	var m Map
	m.LoadScenario(sc)
	return &m, om
}

// Towns grow their roads on the title map through the game tick's
// TownManager::update: after 20 000 ticks (every town visited ~20 times)
// there are more town road pieces, every new one a sane neutral town road.
func TestTownsGrowOnTitleMap(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	type key struct{ x, y, i int }
	before := map[key]bool{}
	count := func() int {
		n := 0
		for y := range m.Tiles {
			for x := range m.Tiles[y] {
				n += len(m.Tiles[y][x].Roads)
			}
		}
		return n
	}
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			for i := range m.Tiles[y][x].Roads {
				before[key{x, y, i}] = true
			}
		}
	}
	n0 := count()
	for range 20000 {
		m.ScenarioTicks++
		m.UpdateTowns(om)
	}
	n1 := count()
	t.Logf("%d towns: road pieces %d → %d", len(m.Towns), n0, n1)
	if n1 <= n0 {
		t.Fatalf("no town roads grew (%d → %d)", n0, n1)
	}
	logged := 0
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			tl := &m.Tiles[y][x]
			for i, e := range tl.Roads {
				if before[key{x, y, i}] {
					continue
				}
				if logged < 6 {
					t.Logf("new road at tile %d,%d: id %d rot %d z %d bridge %v", x, y, e.RoadID, e.Rotation, e.BaseZ, e.HasBridge)
					logged++
				}
				if e.Owner != CompanyNeutral || e.ClearZ <= e.BaseZ || e.RoadID > 9 || e.Ghost {
					t.Errorf("tile %d,%d: new road %+v", x, y, e)
				}
			}
			refs := 0
			for _, r := range tl.Order {
				if r.Kind == scenario.KindRoad {
					refs++
				}
			}
			if refs != len(tl.Roads) {
				t.Errorf("tile %d,%d: %d road refs for %d roads", x, y, refs, len(tl.Roads))
			}
		}
	}
}

// CargoFor (VehicleObject::load's cargo) against the title map's saved
// vehicles: each first body's primary cargo — capacity and accepted types —
// as the object gives it (refits keep both).
func TestVehicleCargoFromObjects(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	checked, bad := 0, 0
	for _, e := range m.Entities {
		if e.EntityType != scenario.VehicleTypeBodyStart || e.BodyIndex != 0 || int(e.ObjectID) >= len(om.Vehicles) || om.Vehicles[e.ObjectID] == nil {
			continue
		}
		c := om.CargoFor(om.Vehicles[e.ObjectID])
		checked++
		if c.NumSimultaneousCargoTypes == 0 {
			if e.Cargo.AcceptedTypes != 0 {
				bad++
				t.Errorf("%s: no cargo by the object, saved accepts %#x", om.Vehicles[e.ObjectID].DisplayName, e.Cargo.AcceptedTypes)
			}
			continue
		}
		if e.Cargo.AcceptedTypes != c.CompatibleCargoCategories[0] || e.Cargo.MaxQty != c.MaxCargo[0] {
			bad++
			if bad < 6 {
				t.Errorf("%s: saved accepts %#x max %d, object %#x max %d", om.Vehicles[e.ObjectID].DisplayName, e.Cargo.AcceptedTypes, e.Cargo.MaxQty, c.CompatibleCargoCategories[0], c.MaxCargo[0])
			}
		}
	}
	if checked == 0 {
		t.Fatal("no vehicle bodies checked")
	}
	t.Logf("%d first bodies checked, %d differ", checked, bad)
}
