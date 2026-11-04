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

// Company colour schemes against the title map's saved vehicles: each car
// body's colours are its owner's scheme for the object's colour type
// (vehicleColourScheme, as createCar picks them), and the main primary is the
// one CompanyColours already had.
func TestCompanyColourSchemes(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	for o, cs := range m.CompanyColourSchemes {
		if o < len(m.CompanyColours) && !m.CompanyEmpty(o) && cs.Main[0] != m.CompanyColours[o] {
			t.Errorf("company %d: main %v, colours %d", o, cs.Main, m.CompanyColours[o])
		}
	}
	checked, bad := 0, 0
	for _, e := range m.Entities {
		if e.EntityType != scenario.VehicleTypeBodyStart || int(e.ObjectID) >= len(om.Vehicles) || om.Vehicles[e.ObjectID] == nil {
			continue
		}
		want := m.vehicleColourScheme(e.Owner, om.Vehicles[e.ObjectID].ColourType)
		checked++
		if [2]uint8{e.ColourPrimary, e.ColourSecondary} != want {
			bad++
			if bad < 5 {
				t.Logf("%s (owner %d): colours %d/%d, scheme %v", om.Vehicles[e.ObjectID].DisplayName, e.Owner, e.ColourPrimary, e.ColourSecondary, want)
			}
		}
	}
	t.Logf("%d bodies checked, %d differ", checked, bad)
	if checked == 0 || bad*10 > checked { // repaints are allowed, not a pattern
		t.Errorf("%d of %d bodies' colours are not their company's scheme", bad, checked)
	}
}

// createCar against the title map's saved trains: for each saved train's
// first car, a new train of the same object for the same owner gets the same
// structure (car components, entity types), per component the same sprite
// types and sizes, reversed / jacobs flags, cargo and colours.
func TestCreateCarMatchesSaved(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, tr := range trains[:min(len(trains), 25)] {
		if len(tr.Cars) == 0 {
			continue
		}
		car := tr.Cars[0]
		first := m.Entities[car[0].Body]
		head := m.Entities[tr.Head]
		h, ok := m.createBaseVehicle(head.Owner, head.Mode, head.VehicleType, head.TrackType)
		if !ok || !m.createCar(head.Owner, h, first.ObjectID, om) {
			t.Fatalf("could not build %d", first.ObjectID)
		}
		nt, err := m.TrainOf(h)
		if err != nil || len(nt.Cars) != 1 || len(nt.Cars[0]) != len(car) {
			t.Fatalf("object %d: new train %+v err %v (saved car has %d components)", first.ObjectID, nt, err, len(car))
		}
		const flagsCompared = f38IsReversed | f38JacobsBogieAvailable | f38Unk0
		for i, sc := range car {
			nc := nt.Cars[0][i]
			for _, p := range [][2]int{{sc.Front, nc.Front}, {sc.Back, nc.Back}, {sc.Body, nc.Body}} {
				a, b := m.Entities[p[0]], m.Entities[p[1]]
				// The sprite sizes are not compared yet: VehicleObject::load measures
				// them from the images (getImagesMaxExtent) — 5P.a part 3d.
				if a.EntityType != b.EntityType || a.ObjectSpriteType != b.ObjectSpriteType ||
					a.Flags38&flagsCompared != b.Flags38&flagsCompared || a.Cargo.AcceptedTypes != b.Cargo.AcceptedTypes ||
					a.Cargo.MaxQty != b.Cargo.MaxQty || a.ColourPrimary != b.ColourPrimary || a.ColourSecondary != b.ColourSecondary {
					t.Errorf("%s component %d type %d:\n saved %+v\n new   %+v", om.Vehicles[first.ObjectID].DisplayName, i, a.EntityType,
						[]any{a.ObjectSpriteType, a.SpriteWidth, a.SpriteHeightNegative, a.SpriteHeightPositive, a.Flags38 & flagsCompared, a.Cargo.AcceptedTypes, a.Cargo.MaxQty, a.ColourPrimary, a.ColourSecondary},
						[]any{b.ObjectSpriteType, b.SpriteWidth, b.SpriteHeightNegative, b.SpriteHeightPositive, b.Flags38 & flagsCompared, b.Cargo.AcceptedTypes, b.Cargo.MaxQty, b.ColourPrimary, b.ColourSecondary})
				}
			}
		}
		checked++
	}
	t.Logf("%d trains' first cars rebuilt and compared", checked)
}
