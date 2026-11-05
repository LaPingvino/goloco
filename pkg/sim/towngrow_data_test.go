package sim

import (
	"github.com/LaPingvino/goloco/pkg/assets"
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
	if g1, err := assets.LoadG1("../../../locomotion/Data/g1.DAT"); err == nil { // sprites as the game loads them
		om.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
		om.G1File = g1
	}
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
				if a.EntityType != b.EntityType || a.ObjectSpriteType != b.ObjectSpriteType || a.SpriteWidth != b.SpriteWidth ||
					a.SpriteHeightNegative != b.SpriteHeightNegative || a.SpriteHeightPositive != b.SpriteHeightPositive ||
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

// autoLayoutTrain: the title map's saved trains were laid out by the game, so
// laying them out again keeps each one's order, orientation and sprites (a
// few may differ where the player shunted); and a train built wagon first,
// locomotive second gets the locomotive moved to the front.
func TestAutoLayoutTrain(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	snap := func(head uint16) []uint32 {
		tr, err := m.TrainOf(head)
		if err != nil {
			t.Fatal(err)
		}
		var out []uint32
		for _, car := range tr.Cars {
			for _, cc := range car {
				for _, i := range []int{cc.Front, cc.Back, cc.Body} {
					e := m.Entities[i]
					out = append(out, uint32(e.ID)<<16|uint32(e.ObjectSpriteType)<<8|uint32(e.Flags38&f38IsReversed))
				}
			}
		}
		return out
	}
	same, changed := 0, 0
	var loco, wagon *Train
	for _, tr := range trains {
		head := m.Entities[tr.Head].ID
		before := snap(head)
		m.autoLayoutTrain(head, om)
		after := snap(head)
		if len(before) == len(after) && func() bool {
			for i := range before {
				if before[i] != after[i] {
					return false
				}
			}
			return true
		}() {
			same++
		} else {
			changed++
			if changed <= 3 {
				t.Logf("train %d (%d cars) changed: %d → %d components", head, len(tr.Cars), len(before), len(after))
			}
		}
		if loco == nil && len(tr.Cars) >= 2 && m.Entities[tr.Head].Mode == 0 {
			first, last := om.Vehicles[m.Entities[tr.Cars[0][0].Front].ObjectID], om.Vehicles[m.Entities[tr.Cars[len(tr.Cars)-1][0].Front].ObjectID]
			if first != nil && last != nil && first.Power != 0 && last.Power == 0 {
				x := tr
				loco, wagon = &x, &x
			}
		}
	}
	t.Logf("%d trains laid out again: %d unchanged, %d changed", len(trains), same, changed)
	if changed*10 > len(trains) {
		t.Errorf("too many saved trains change: %d of %d", changed, len(trains))
	}
	if loco == nil {
		t.Skip("no rail train with a locomotive in front and a wagon behind")
	}
	head := m.Entities[loco.Head]
	locoObj, wagonObj := m.Entities[loco.Cars[0][0].Front].ObjectID, m.Entities[wagon.Cars[len(wagon.Cars)-1][0].Front].ObjectID
	h, _ := m.createBaseVehicle(head.Owner, head.Mode, head.VehicleType, head.TrackType)
	m.createCar(head.Owner, h, wagonObj, om)
	m.createCar(head.Owner, h, locoObj, om)
	m.autoLayoutTrain(h, om)
	nt, _ := m.TrainOf(h)
	if got := m.Entities[nt.Cars[0][0].Front].ObjectID; got != locoObj {
		t.Errorf("first car after the layout: object %d, want the locomotive %d", got, locoObj)
	}
}

// TestTrackTypeForSaved: VehicleObject::load's trackType (ObjectManager.
// TrackTypeFor) against every saved train — the head's track type is its first
// car's object's, as createVehicle passes it to createBaseVehicle.
func TestTrackTypeForSaved(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	checked, bad := 0, 0
	for _, tr := range trains {
		if len(tr.Cars) == 0 {
			continue
		}
		head := m.Entities[tr.Head]
		obj := om.Vehicles[m.Entities[tr.Cars[0][0].Body].ObjectID]
		if got := om.TrackTypeFor(obj); got != head.TrackType {
			bad++
			if bad <= 5 {
				t.Errorf("%s (mode %d, track %q): TrackTypeFor %d, the saved head has %d", obj.DisplayName, obj.Mode, obj.TrackName, got, head.TrackType)
			}
		}
		checked++
	}
	t.Logf("%d trains checked, %d differ", checked, bad)
}

// TestCreateVehicleCommand: recalculateTransportCounts against the saved
// counts; then buying a vehicle (createVehicle, no head): one more of its type,
// one train with one car, the inflation-adjusted cost; locked refused; and
// createNewVehicle's cleanup giving back every slot, order byte and routing.
func TestCreateVehicleCommand(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	for id := range m.Companies {
		saved := m.Companies[id].TransportTypeCount
		m.RecalculateTransportCounts(uint8(id))
		if m.Companies[id].TransportTypeCount != saved {
			t.Errorf("company %d: counts %v, saved %v", id, m.Companies[id].TransportTypeCount, saved)
		}
	}
	trains, _ := m.Trains()
	head := m.Entities[trains[0].Head]
	company, objID := head.Owner, m.Entities[trains[0].Cars[0][0].Body].ObjectID
	c := &Cmd{Company: company, Objects: om}
	obj := om.Vehicles[objID]
	before := m.Companies[company].TransportTypeCount[obj.Type]
	n := len(trains)
	var h uint16
	cost, ok := m.CreateVehicle(c, CmdApply, objID, true, &h)
	if !ok || h == entityIDNull {
		t.Fatalf("buying %s refused", obj.DisplayName)
	}
	if want := m.InflationAdjustedCost(obj.CostFactor, obj.CostIndex, 6); cost != want || cost <= 0 {
		t.Errorf("cost %d, want %d", cost, want)
	}
	if got := m.Companies[company].TransportTypeCount[obj.Type]; got != before+1 {
		t.Errorf("count of type %d: %d, want %d", obj.Type, got, before+1)
	}
	tr, err := m.TrainOf(h)
	if err != nil || len(tr.Cars) != 1 {
		t.Errorf("the new train: %+v %v", tr, err)
	}
	if after, _ := m.Trains(); len(after) != n+1 {
		t.Errorf("%d trains, want %d", len(after), n+1)
	}
	if nh := m.entity(h); nh.TrackType != om.TrackTypeFor(obj) || nh.TileX != -1 {
		t.Errorf("head track type %d tileX %d", nh.TrackType, nh.TileX)
	}
	// locked: refused (unless lockedOK)
	m.UnlockedVehicles[company][objID/32] &^= 1 << (objID % 32)
	if _, ok := m.CreateVehicle(c, CmdApply, objID, false, nil); ok {
		t.Error("a locked vehicle was bought")
	}
	// the cleanup
	free, orders, ents := m.EntityLists.numFreeEntities(), m.OrderTableLength, len(m.Entities)
	b, ok := m.createBaseVehicle(company, uint8(obj.Mode), uint8(obj.Type), 0)
	if !ok {
		t.Fatal("createBaseVehicle")
	}
	routing := m.entity(b).RoutingHandle
	m.deleteBaseVehicle(b)
	if m.EntityLists.numFreeEntities() != free || m.OrderTableLength != orders || len(m.Entities) != ents {
		t.Errorf("after the cleanup: free %d (was %d), orders %d (was %d), entities %d (was %d)",
			m.EntityLists.numFreeEntities(), free, m.OrderTableLength, orders, len(m.Entities), ents)
	}
	if m.vehicleRoutings(routing / 64)[0] != RoutingNull {
		t.Error("the routing slot was not freed")
	}
}
