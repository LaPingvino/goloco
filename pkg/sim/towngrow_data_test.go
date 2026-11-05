package sim

import (
	"fmt"
	"github.com/LaPingvino/goloco/pkg/assets"
	"math"
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

// TestAddCarToVehicle: canVehiclesCouple holds between the cars of every saved
// train; a car added to a newly bought vehicle (two cars, still one vehicle,
// the cost); refused for another company, another kind of vehicle, and (for
// now, 5P.b) a vehicle on the track.
func TestAddCarToVehicle(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	pairs := 0
	for _, tr := range trains {
		for i := range tr.Cars {
			for j := range tr.Cars {
				if i == j {
					continue // a car isn't coupled to itself
				}
				a, b := m.Entities[tr.Cars[i][0].Body].ObjectID, m.Entities[tr.Cars[j][0].Front].ObjectID
				if !canVehiclesCouple(a, b, om) {
					t.Errorf("saved train %d: cars %s and %s can't couple", m.Entities[tr.Head].ID, om.Vehicles[a].DisplayName, om.Vehicles[b].DisplayName)
				}
				pairs++
			}
		}
	}
	t.Logf("%d saved car pairs couple", pairs)
	saved := trains[0]
	company := m.Entities[saved.Head].Owner
	objID := m.Entities[saved.Cars[len(saved.Cars)-1][0].Body].ObjectID // a wagon, often
	c := &Cmd{Company: company, Objects: om}
	var h uint16
	if _, ok := m.CreateVehicle(c, CmdApply, objID, true, &h); !ok {
		t.Fatal("buying refused")
	}
	counts := m.Companies[company].TransportTypeCount
	cost, ok := m.AddCarToVehicle(c, CmdApply, objID, h, true)
	obj := om.Vehicles[objID]
	if !ok || cost != m.InflationAdjustedCost(obj.CostFactor, obj.CostIndex, 6) {
		t.Fatalf("adding a car: ok %v cost %d", ok, cost)
	}
	if tr, _ := m.TrainOf(h); len(tr.Cars) != 2 {
		t.Errorf("%d cars, want 2", len(tr.Cars))
	}
	if m.Companies[company].TransportTypeCount != counts {
		t.Error("adding a car changed the vehicle counts")
	}
	other := &Cmd{Company: company ^ 1, Objects: om}
	if _, ok := m.AddCarToVehicle(other, CmdApply, objID, h, true); ok {
		t.Error("another company added a car")
	}
	for id, v := range om.Vehicles {
		if v != nil && v.Type != obj.Type {
			if _, ok := m.AddCarToVehicle(c, 0, uint16(id), h, true); ok {
				t.Errorf("%s (type %d) fits a type %d vehicle", v.DisplayName, v.Type, obj.Type)
			}
			break
		}
	}
	// a vehicle on the track: lifted, the car added, put back (a stopped rail one)
	for _, tr := range trains {
		hh := m.Entities[tr.Head]
		if hh.TileX == -1 || hh.Mode != 0 || !m.canBeModified(tr) {
			continue
		}
		car := m.Entities[tr.Cars[len(tr.Cars)-1][0].Body].ObjectID
		cc := &Cmd{Company: hh.Owner, Objects: om}
		m.UnlockedVehicles[hh.Owner][car/32] |= 1 << (car % 32)
		n := len(tr.Cars)
		if _, ok := m.AddCarToVehicle(cc, CmdApply, car, hh.ID, false); !ok {
			t.Fatalf("vehicle %d: adding a car on the track refused", hh.ID)
		}
		nt, _ := m.TrainOf(hh.ID)
		if len(nt.Cars) != n+1 || m.Entities[nt.Head].TileX == -1 || m.Entities[nt.Head].Status != StatusStopped {
			t.Errorf("vehicle %d: %d cars (was %d), tileX %d status %d", hh.ID, len(nt.Cars), n, m.Entities[nt.Head].TileX, m.Entities[nt.Head].Status)
		}
		break
	}
}

// TestVehiclePickup: saved trains on the track lifted off (VehiclePickup):
// every component off the map, stopped (commandStop, status 0, no station),
// empty; the routings beyond the vehicle released; another company refused.
func TestVehiclePickup(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	picked := 0
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.TileX == -1 || h.Mode == 2 || h.Mode == 3 { // land vehicles on the map
			continue
		}
		if !m.canBeModified(tr) { // moving, broken down, stuck …: refused (as checked here)
			if _, ok := m.VehiclePickup(&Cmd{Company: h.Owner, Objects: om}, CmdApply, h.ID); ok {
				t.Errorf("vehicle %d picked up though it can't be modified", h.ID)
			}
			continue
		}
		if _, ok := m.VehiclePickup(&Cmd{Company: h.Owner ^ 1, Objects: om}, CmdApply, h.ID); ok {
			t.Fatalf("vehicle %d picked up by another company", h.ID)
		}
		if _, ok := m.VehiclePickup(&Cmd{Company: h.Owner, Objects: om}, CmdApply, h.ID); !ok {
			t.Fatalf("vehicle %d: pickup refused", h.ID)
		}
		nt, err := m.TrainOf(h.ID)
		if err != nil {
			t.Fatal(err)
		}
		m.eachComponent(nt, func(i int) {
			if e := m.Entities[i]; e.TileX != -1 || e.Flags38&flags38Unk2 != 0 {
				t.Errorf("vehicle %d component %d: tileX %d flags38 %#x", h.ID, e.ID, e.TileX, e.Flags38)
			}
		})
		nh := m.Entities[nt.Head]
		if nh.VehicleFlags&VehicleFlagCommandStop == 0 || nh.Status != 0 || nh.StationID != 0xFFFF {
			t.Errorf("vehicle %d: flags %#x status %d station %d", h.ID, nh.VehicleFlags, nh.Status, nh.StationID)
		}
		for _, car := range nt.Cars {
			for _, cc := range car {
				if m.Entities[cc.Body].Cargo.Qty != 0 || m.Entities[cc.Front].Cargo.Qty != 0 {
					t.Errorf("vehicle %d: cargo left", h.ID)
				}
			}
		}
		for _, rh := range m.RoutingRing(RoutingHandle(m.Entities[nt.Veh2].RoutingHandle), false)[1:] {
			if r := m.Routing(rh); r != RoutingAllocatedButFree && r != RoutingNull {
				t.Errorf("vehicle %d: a routing beyond veh2 still held (%#x)", h.ID, r)
				break
			}
		}
		picked++
	}
	t.Logf("%d vehicles picked up", picked)
	if picked == 0 {
		t.Fatal("no vehicle to pick up")
	}
}

// TestPlaceDownRoundTrip: stopped saved vehicles picked up and put back where
// their head was (as addCarToVehicle does around a new car). A saved stop is
// where motion left the train, a placement settles it onto exact sub-positions,
// so the saved game is no exact oracle: every component (but veh1, the road
// vehicles' look-ahead, which reserves afresh) within 2 px of where it was;
// a second pickup and placement lands exactly where the first did; refusals
// only as "in the way" (Unk1).
func TestPlaceDownRoundTrip(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	placed, inTheWay, repeat := 0, 0, 0
	positions := func(tr Train) map[uint16][3]int32 {
		out := map[uint16][3]int32{}
		m.eachComponent(tr, func(i int) {
			e := m.Entities[i]
			out[e.ID] = [3]int32{int32(e.PosX), int32(e.PosY), int32(e.PosZ)}
		})
		return out
	}
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.TileX == -1 || h.Mode == 2 || h.Mode == 3 || !m.canBeModified(tr) {
			continue
		}
		saved := positions(tr)
		// at veh2, the vehicle's front (a road vehicle's head runs far ahead,
		// leading its route; addCarToVehicle's placement reads the head, as upstream)
		v2 := m.Entities[tr.Veh2]
		x, y, z, tad, sub := int32(v2.TileX), int32(v2.TileY), v2.TileBaseZ, v2.TrackAndDirection, v2.SubPosition
		roundTrip := func() PlaceDownResult {
			if _, ok := m.VehiclePickup(&Cmd{Company: h.Owner, Objects: om}, CmdApply, h.ID); !ok {
				t.Fatalf("vehicle %d: pickup refused", h.ID)
			}
			return m.PlaceDownVehicle(h.ID, x, y, z, tad, sub, om)
		}
		switch r := roundTrip(); r {
		case PlaceDownUnk1:
			inTheWay++
			continue
		case PlaceDownOkay:
		default:
			t.Errorf("vehicle %d: placed back: result %d", h.ID, r)
			continue
		}
		placed++
		nt, _ := m.TrainOf(h.ID)
		first := positions(nt)
		for _, i := range trainComponents(nt) {
			e := m.Entities[i]
			if i == nt.Veh1 || (i == nt.Head && h.Mode == 1) {
				continue // the look-ahead; a road head (it led the route, placing brings it back)
			}
			a, b := saved[e.ID], first[e.ID]
			for k := range 3 {
				if d := a[k] - b[k]; d < -2 || d > 2 {
					t.Errorf("vehicle %d component %d (type %d): saved %v, placed %v", h.ID, e.ID, e.EntityType, a, b)
					break
				}
			}
		}
		if m.Entities[nt.Head].Status != StatusStopped || m.Entities[nt.Head].TileX == -1 {
			t.Errorf("vehicle %d: status %d tileX %d after placing", h.ID, m.Entities[nt.Head].Status, m.Entities[nt.Head].TileX)
		}
		nv2 := m.Entities[nt.Veh2]
		x, y, z, tad, sub = int32(nv2.TileX), int32(nv2.TileY), nv2.TileBaseZ, nv2.TrackAndDirection, nv2.SubPosition
		if r := roundTrip(); r != PlaceDownOkay {
			t.Errorf("vehicle %d: second placement %d", h.ID, r)
			continue
		}
		nt, _ = m.TrainOf(h.ID)
		second, same := positions(nt), true
		for id, p := range first {
			q := second[id]
			if q != p {
				same = false
			}
			for k := range 3 {
				if d := p[k] - q[k]; d < -2 || d > 2 {
					t.Errorf("vehicle %d component %d: first placement %v, second %v", h.ID, id, p, q)
					break
				}
			}
		}
		if same {
			repeat++
		} else {
			t.Logf("vehicle %d: the second placement is a sub-position step off the first (open: OpenLoco's own behaviour, LOOP)", h.ID)
		}
	}
	t.Logf("%d placed back (twice; %d the second time exactly as the first), %d in the way", placed, repeat, inTheWay)
	if placed < 10 {
		t.Fatalf("only %d placed", placed)
	}
}

// TestVehiclePlaceCommand: the vehiclePlace command on the stopped saved
// vehicles, picked up first: put back on veh2's piece (validate*Placement
// finds the piece, the owner, the extras); refused for another company, the
// piece turned a quarter, an empty spot, a vehicle already placed; a ghost
// placement flagged, then made real (convertGhost).
func TestVehiclePlaceCommand(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	placed := 0
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.TileX == -1 || h.Mode == 2 || h.Mode == 3 || !m.canBeModified(tr) {
			continue
		}
		v2 := m.Entities[tr.Veh2]
		own := &Cmd{Company: h.Owner, Objects: om}
		a := VehiclePlacement{X: int32(v2.TileX), Y: int32(v2.TileY), Z: int32(v2.TileBaseZ) * 4, TrackAndDirection: v2.TrackAndDirection, TrackProgress: v2.SubPosition, Head: h.ID}
		if _, ok := m.VehiclePlace(own, CmdApply, a); ok {
			t.Errorf("vehicle %d: placed while on the track", h.ID)
		}
		m.VehiclePickup(own, CmdApply, h.ID)
		if _, ok := m.VehiclePlace(&Cmd{Company: h.Owner ^ 1, Objects: om}, CmdApply, a); ok {
			t.Errorf("vehicle %d: placed by another company", h.ID)
		}
		turned := a
		turned.TrackAndDirection = a.TrackAndDirection&^3 | (a.TrackAndDirection+1)&3
		if _, ok := m.VehiclePlace(own, CmdApply, turned); ok {
			t.Errorf("vehicle %d: placed on the piece turned a quarter", h.ID)
		}
		empty := a
		empty.X, empty.Y = 32, 32
		if _, ok := m.VehiclePlace(own, CmdApply, empty); ok {
			t.Errorf("vehicle %d: placed on the map's edge", h.ID)
		}
		if _, ok := m.VehiclePlace(own, CmdApply|CmdGhost, a); !ok {
			t.Errorf("vehicle %d (mode %d, tad %#x): placing it back refused", h.ID, h.Mode, a.TrackAndDirection)
			continue
		}
		nt, _ := m.TrainOf(h.ID)
		nh := m.Entities[nt.Head]
		if nh.TileX == -1 || nh.Flags38&flags38IsGhost == 0 || nh.VehicleFlags&VehicleFlagCommandStop == 0 || nh.ManualPower != -40 {
			t.Errorf("vehicle %d after placing: tileX %d flags38 %#x flags %#x manual %d", h.ID, nh.TileX, nh.Flags38, nh.VehicleFlags, nh.ManualPower)
		}
		if _, ok := m.VehiclePlace(own, CmdApply, VehiclePlacement{Head: h.ID, ConvertGhost: true}); !ok || m.Entities[nt.Head].Flags38&flags38IsGhost != 0 {
			t.Errorf("vehicle %d: convertGhost", h.ID)
		}
		placed++
	}
	t.Logf("%d vehicles placed by the command", placed)
	if placed < 10 {
		t.Fatalf("only %d", placed)
	}
}

// TestVehicleChangeRunningMode: every saved running vehicle passes
// canVehicleBeStarted (it was started once); stop / start, local / express and
// manual driving on one; a vehicle of wagons only refused.
func TestVehicleChangeRunningMode(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	running := 0
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.VehicleFlags&VehicleFlagCommandStop != 0 || h.Status == StatusCrashed || h.Status == StatusStuck {
			continue
		}
		if !m.canVehicleBeStarted(tr, om) {
			t.Errorf("saved running vehicle %d (%s): can't be started", h.ID, om.Vehicles[m.Entities[tr.Cars[0][0].Front].ObjectID].DisplayName)
		}
		running++
	}
	t.Logf("%d running vehicles startable", running)
	tr := trains[0]
	id := m.Entities[tr.Head].ID
	c := &Cmd{Company: m.Entities[tr.Head].Owner, Objects: om}
	flags := func() uint16 { return m.Entities[tr.Head].VehicleFlags }
	m.VehicleChangeRunningMode(c, CmdApply, id, RunningModeStop)
	if flags()&VehicleFlagCommandStop == 0 {
		t.Error("stop: not stopped")
	}
	if _, ok := m.VehicleChangeRunningMode(c, CmdApply, id, RunningModeStart); !ok || flags()&VehicleFlagCommandStop != 0 {
		t.Errorf("start: ok %v flags %#x", ok, flags())
	}
	x := m.Entities[tr.Veh1].Var48 & flags48ExpressMode
	m.VehicleChangeRunningMode(c, CmdApply, id, RunningModeToggleLocalExpress)
	if m.Entities[tr.Veh1].Var48&flags48ExpressMode == x {
		t.Error("local / express not toggled")
	}
	m.VehicleChangeRunningMode(c, CmdApply, id, RunningModeDriveManually)
	if flags()&VehicleFlagManualControl == 0 || flags()&VehicleFlagCommandStop != 0 || m.Entities[tr.Head].ManualPower != -40 {
		t.Errorf("manual on: flags %#x power %d", flags(), m.Entities[tr.Head].ManualPower)
	}
	m.VehicleChangeRunningMode(c, CmdApply, id, RunningModeDriveManually)
	if flags()&VehicleFlagManualControl != 0 || flags()&VehicleFlagCommandStop == 0 {
		t.Errorf("manual off: flags %#x", flags())
	}
	// wagons only: no power
	for oid, v := range om.Vehicles {
		if v == nil || v.Power != 0 || v.Mode != 0 || v.Type != 0 {
			continue
		}
		var h uint16
		if _, ok := m.CreateVehicle(c, CmdApply, uint16(oid), true, &h); !ok {
			continue
		}
		if _, ok := m.VehicleChangeRunningMode(c, CmdApply, h, RunningModeStart); ok && m.Entities[m.mustTrain(t, h).Head].VehicleFlags&VehicleFlagCommandStop == 0 {
			t.Errorf("a %s alone started", v.DisplayName)
		}
		break
	}
}

func (m *Map) mustTrain(t *testing.T, h uint16) Train {
	tr, err := m.TrainOf(h)
	if err != nil {
		t.Fatal(err)
	}
	return tr
}

// TestVehicleOrders: orders inserted into a saved train — every OTHER train's
// orders read the same after (their tables moved as upstream reoffsets them);
// a repeated stop at the same station becomes route-through; delete and skip;
// a station of another company refused.
func TestVehicleOrders(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	ringOf := func(id uint16) string {
		h := m.entity(id)
		out := ""
		for _, o := range m.OrderRing(h.OrderTableOffset, 0) {
			n := orderSizes[o.Type()]
			out += fmt.Sprintf("%x ", m.Orders[o.Offset:o.Offset+n])
		}
		return out
	}
	before := map[uint16]string{}
	for _, tr := range trains {
		before[m.Entities[tr.Head].ID] = ringOf(m.Entities[tr.Head].ID)
	}
	h := m.Entities[trains[0].Head]
	c := &Cmd{Company: h.Owner, Objects: om}
	own, other := -1, -1
	for i, st := range m.StationNames {
		if st.Name == 0 {
			continue
		}
		if st.Owner == h.Owner && own < 0 {
			own = i
		} else if st.Owner != h.Owner && st.Owner < 16 && other < 0 {
			other = i
		}
	}
	if own < 0 {
		t.Fatal("no station of the company")
	}
	n := len(m.OrderRing(h.OrderTableOffset, 0))
	t.Logf("vehicle %d: %d orders; own station %d, another company's %d", h.ID, n, own, other)
	if _, ok := m.VehicleOrderInsert(c, CmdApply, h.ID, 0, OrderStopAtRaw(own)); !ok {
		t.Fatal("insert refused")
	}
	ring := m.OrderRing(m.entity(h.ID).OrderTableOffset, 0)
	if len(ring) != n+1 || ring[0].Type() != OrderStopAt || ring[0].Station() != own {
		t.Fatalf("after insert: %d orders, first %d at %d", len(ring), ring[0].Type(), ring[0].Station())
	}
	for _, tr := range trains[1:] {
		id := m.Entities[tr.Head].ID
		if ringOf(id) != before[id] {
			t.Errorf("vehicle %d's orders changed: %q → %q", id, before[id], ringOf(id))
		}
	}
	// the same stop again right after it: that order becomes route-through
	m.VehicleOrderInsert(c, CmdApply, h.ID, 2, OrderStopAtRaw(own))
	ring = m.OrderRing(m.entity(h.ID).OrderTableOffset, 0)
	if len(ring) != n+1 || ring[0].Type() != OrderRouteThrough {
		t.Errorf("repeated stop: %d orders, first type %d (want %d, route-through)", len(ring), ring[0].Type(), OrderRouteThrough)
	}
	m.VehicleOrderDelete(c, CmdApply, h.ID, 0)
	if ringOf(h.ID) != before[h.ID] {
		t.Errorf("after delete: %q, want %q", ringOf(h.ID), before[h.ID])
	}
	for _, tr := range trains[1:] {
		if id := m.Entities[tr.Head].ID; ringOf(id) != before[id] {
			t.Errorf("vehicle %d's orders changed after the delete", id)
		}
	}
	if n >= 2 {
		cur := m.entity(h.ID).CurrentOrder
		m.VehicleOrderSkip(c, CmdApply, h.ID)
		if m.entity(h.ID).CurrentOrder == cur {
			t.Error("skip: the current order stayed")
		}
	}
	if other >= 0 {
		if _, ok := m.VehicleOrderInsert(c, CmdApply, h.ID, 0, OrderStopAtRaw(other)); ok {
			t.Error("a stop at another company's station inserted")
		}
	}
}

// TestVehicleOrderMoves: a saved vehicle with 3+ orders: up swaps an order with
// the one before, down puts it back, reverse turns the list round and back —
// the current order always still the same order, the other vehicles' orders
// untouched.
func TestVehicleOrderMoves(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	list := func(id uint16) []string {
		h := m.entity(id)
		var out []string
		for _, o := range m.OrderRing(h.OrderTableOffset, 0) {
			out = append(out, fmt.Sprintf("%x", m.Orders[o.Offset:o.Offset+orderSizes[o.Type()]]))
		}
		return out
	}
	current := func(id uint16) string {
		h := m.entity(id)
		o := Order{m, int(h.OrderTableOffset) + int(h.CurrentOrder)}
		return fmt.Sprintf("%x", m.Orders[o.Offset:o.Offset+orderSizes[o.Type()]])
	}
	var id uint16
	found := false
	for _, tr := range trains {
		if h := m.Entities[tr.Head]; len(list(h.ID)) >= 3 {
			id, found = h.ID, true
			break
		}
	}
	if !found {
		t.Fatal("no vehicle with 3 orders")
	}
	others := map[uint16]string{}
	for _, tr := range trains {
		if oid := m.Entities[tr.Head].ID; oid != id {
			others[oid] = fmt.Sprint(list(oid))
		}
	}
	c := &Cmd{Company: m.entity(id).Owner, Objects: om}
	orig, cur := list(id), current(id)
	t.Logf("vehicle %d: %v, current %s", id, orig, cur)
	second := m.OrderRing(m.entity(id).OrderTableOffset, 0)[1].Offset - int(m.entity(id).OrderTableOffset)
	m.VehicleOrderUp(c, CmdApply, id, uint16(second))
	if l := list(id); l[0] != orig[1] || l[1] != orig[0] || current(id) != cur {
		t.Errorf("up: %v, current %s (was %v, %s)", l, current(id), orig, cur)
	}
	m.VehicleOrderDown(c, CmdApply, id, 0)
	if fmt.Sprint(list(id)) != fmt.Sprint(orig) || current(id) != cur {
		t.Errorf("down back: %v, current %s", list(id), current(id))
	}
	m.VehicleOrderReverse(c, CmdApply, id)
	l := list(id)
	for i := range orig {
		if l[i] != orig[len(orig)-1-i] {
			t.Fatalf("reverse: %v of %v", l, orig)
		}
	}
	if current(id) != cur {
		t.Errorf("reverse: current %s, was %s", current(id), cur)
	}
	m.VehicleOrderReverse(c, CmdApply, id)
	if fmt.Sprint(list(id)) != fmt.Sprint(orig) || current(id) != cur {
		t.Errorf("reversed twice: %v current %s", list(id), current(id))
	}
	for oid, was := range others {
		if fmt.Sprint(list(oid)) != was {
			t.Errorf("vehicle %d's orders changed", oid)
		}
	}
}

// TestPlayableVehicleEndToEnd (5P.e): the player's side, all through the game
// commands: a stopped saved train on the title map is lifted off its line; a
// new train with the same cars is bought (createVehicle, then a car at a time),
// put where it stood (vehiclePlace), given its stops (vehicleOrderInsert) and
// started (vehicleChangeRunningMode); then the game runs: it leaves, travels
// and calls at both stations.
// PARTIAL: "and earns" — nothing produces cargo yet (buildings' and
// industries' production into stations, DeliverCargoToStation's callers, are
// not ported: LOOP item CP), so the saved train's wait-for-full-load order is
// left out and the stations stay empty.
func TestPlayableVehicleEndToEnd(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	var old Train
	found := false
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.Mode != 0 || h.TileX == -1 || !m.canBeModified(tr) {
			continue
		}
		stops := 0
		for _, o := range m.OrderRing(h.OrderTableOffset, 0) {
			if o.Type() == OrderStopAt {
				stops++
			}
		}
		if stops >= 2 {
			old, found = tr, true
			break
		}
	}
	if !found {
		t.Fatal("no stopped rail train with two stops")
	}
	oh := m.Entities[old.Head]
	v2 := m.Entities[old.Veh2]
	c := &Cmd{Company: oh.Owner, Objects: om, Players: func(uint8) bool { return true }}
	var orders [][]byte
	for _, o := range m.OrderRing(oh.OrderTableOffset, 0) {
		if o.Type() == OrderStopAt { // its stops (no wait for a full load: nothing produces cargo yet)
			orders = append(orders, append([]byte(nil), m.Orders[o.Offset:o.Offset+orderSizes[o.Type()]]...))
		}
	}
	var cars []uint16
	for _, car := range old.Cars {
		cars = append(cars, m.Entities[car[0].Body].ObjectID)
	}
	place := VehiclePlacement{X: int32(v2.TileX), Y: int32(v2.TileY), Z: int32(v2.TileBaseZ) * 4, TrackAndDirection: v2.TrackAndDirection, TrackProgress: v2.SubPosition}
	if _, ok := m.VehiclePickup(c, CmdApply, oh.ID); !ok {
		t.Fatal("lifting the old train refused")
	}
	var h uint16
	if _, ok := m.CreateVehicle(c, CmdApply, cars[0], true, &h); !ok {
		t.Fatal("buying refused")
	}
	for _, car := range cars[1:] {
		if _, ok := m.AddCarToVehicle(c, CmdApply, car, h, true); !ok {
			t.Fatalf("adding car %d refused", car)
		}
	}
	place.Head = h
	if _, ok := m.VehiclePlace(c, CmdApply, place); !ok {
		t.Fatal("placing refused")
	}
	for _, raw := range orders {
		if _, ok := m.VehicleOrderInsert(c, CmdApply, h, m.entity(h).SizeOfOrderTable-1, raw); !ok {
			t.Fatalf("order %x refused", raw)
		}
	}
	if _, ok := m.VehicleChangeRunningMode(c, CmdApply, h, RunningModeStart); !ok {
		t.Fatal("starting refused")
	}
	tr, _ := m.TrainOf(h)
	t.Logf("vehicle %d: %d cars of %v, %d orders, on %d,%d", h, len(tr.Cars), cars, len(orders), place.X, place.Y)
	startX, startY := m.entity(h).PosX, m.entity(h).PosY
	moved := 0.0
	called := map[uint16]bool{}
	for tick := 0; tick < 60000 && len(called) < 2; tick++ {
		m.UpdateVehicles(om)
		tr, err := m.TrainOf(h)
		if err != nil {
			t.Fatal(err)
		}
		hd := m.Entities[tr.Head]
		if d := math.Hypot(float64(hd.PosX-startX), float64(hd.PosY-startY)); d > moved {
			moved = d
		}
		if (hd.Status == StatusLoading || hd.Status == StatusUnloading) && hd.StationID != 0xFFFF {
			if !called[hd.StationID] {
				t.Logf("tick %d: at station %d (status %d)", tick, hd.StationID, hd.Status)
			}
			called[hd.StationID] = true
		}
	}
	t.Logf("moved up to %.0f px; called at %v", moved, called)
	if moved < 100 || len(called) < 2 {
		t.Errorf("the bought train did not run its route: moved %.0f px, called at %v", moved, called)
	}
}

// TestBuildableVehicles (5P.f's list): the trains tab on a saved train's track
// type lists only unlocked rail vehicles of that track, powered first, each
// group by design year; the saved locomotive is in it; cargoless filters; a
// target vehicle limits the list to what couples to it.
func TestBuildableVehicles(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	var tr Train
	for _, x := range trains {
		if m.Entities[x.Head].Mode == 0 {
			tr = x
			break
		}
	}
	h := m.Entities[tr.Head]
	loco := m.Entities[tr.Cars[0][0].Body].ObjectID
	m.UnlockedVehicles[h.Owner][loco/32] |= 1 << (loco % 32)
	o := DefaultBuildVehicleOptions(h.Owner)
	list := m.BuildableVehicles(om, 0, h.TrackType, o)
	t.Logf("company %d, track type %d: %d vehicles", h.Owner, h.TrackType, len(list))
	if len(list) == 0 {
		t.Fatal("empty")
	}
	seenLoco, unpoweredSeen := false, false
	for i, id := range list {
		v := om.Vehicles[id]
		if v.Type != 0 || v.Mode != objects.TransportModeRail || om.TrackTypeFor(v) != h.TrackType || !m.isVehicleIndexUnlocked(h.Owner, id) {
			t.Errorf("%s doesn't belong on the list", v.DisplayName)
		}
		if v.Power == 0 {
			unpoweredSeen = true
		} else if unpoweredSeen {
			t.Errorf("%s (powered) after an unpowered one", v.DisplayName)
		}
		if i > 0 {
			p := om.Vehicles[list[i-1]]
			if (p.Power > 0) == (v.Power > 0) && p.Designed > v.Designed {
				t.Errorf("%s (%d) after %s (%d)", v.DisplayName, v.Designed, p.DisplayName, p.Designed)
			}
		}
		seenLoco = seenLoco || id == loco
	}
	if !seenLoco {
		t.Errorf("the saved locomotive %s is not offered", om.Vehicles[loco].DisplayName)
	}
	o.CargoFilter = 0xFE
	for _, id := range m.BuildableVehicles(om, 0, h.TrackType, o) {
		if c := om.CargoFor(om.Vehicles[id]); c.CompatibleCargoCategories != [2]uint32{} {
			t.Errorf("cargoless filter: %s carries", om.Vehicles[id].DisplayName)
		}
	}
	o = DefaultBuildVehicleOptions(h.Owner)
	o.TargetHead = h.ID
	for _, id := range m.BuildableVehicles(om, 0, h.TrackType, o) {
		if !m.isVehicleTypeCompatible(tr, id, om) {
			t.Errorf("target: %s doesn't couple", om.Vehicles[id].DisplayName)
		}
	}
}

// TestBuildVehicleTabs: setDisabledTransportTabs — a company's own vehicles'
// types are never disabled; setTrackTypeTabs — the trains tab offers the
// track its train uses, each track tab lists vehicles, road tabs carry bit 7;
// the track tab chosen at opening is the toolbar's last choice when offered.
func TestBuildVehicleTabs(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if d := m.BuildVehicleDisabledTabs(h.Owner); d&(1<<h.VehicleType) != 0 {
			t.Errorf("company %d runs a type-%d vehicle but that tab is disabled (%#x)", h.Owner, h.VehicleType, d)
		}
	}
	h := m.Entities[trains[0].Head]
	o := DefaultBuildVehicleOptions(h.Owner)
	for vt := uint8(0); vt < 6; vt++ {
		tabs := m.BuildVehicleTrackTabs(om, vt, o)
		t.Logf("type %d: tabs %v, chosen %d (last rail %d, road %d)", vt, tabs, m.BuildVehicleTrackTab(vt, tabs), m.LastRailroadOption, m.LastRoadOption)
		for _, tt := range tabs {
			if len(m.BuildableVehicles(om, vt, tt, o)) == 0 {
				t.Errorf("type %d: tab %#x lists nothing", vt, tt)
			}
			if (vt == 1 || vt == 2) && tt&(1<<7) == 0 {
				t.Errorf("type %d (road): tab %#x without the road bit", vt, tt)
			}
		}
		if vt == h.VehicleType {
			found := false
			for _, tt := range tabs {
				found = found || tt == h.TrackType
			}
			if !found {
				t.Errorf("the trains tab doesn't offer track type %d", h.TrackType)
			}
		}
	}
}

// TestBuyVehiclePays: createVehicle through DoCommandFor — the cost comes off
// the company's cash and is booked under VehiclePurchases; a player company
// without the cash is refused (ensureCompanyFunding) and nothing is made.
func TestBuyVehiclePays(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	head := m.Entities[trains[0].Head]
	company, objID := head.Owner, m.Entities[trains[0].Cars[0][0].Body].ObjectID
	c := &Cmd{Company: company, Objects: om}
	co := &m.Companies[company]
	co.Cash = 10_000_000
	cash, booked := co.Cash, co.Expenditures[0][ExpenditureVehiclePurchases]
	buy := func() (int32, bool) {
		var h uint16
		return m.DoCommandFor(c, CmdApply, ExpenditureVehiclePurchases, func(f uint8) (int32, bool) {
			return m.CreateVehicle(c, f, objID, true, &h)
		})
	}
	cost, ok := buy()
	if !ok || cost <= 0 {
		t.Fatalf("buying refused (%d)", cost)
	}
	if co.Cash != cash-int64(cost) || co.Expenditures[0][ExpenditureVehiclePurchases] != booked-cost {
		t.Errorf("cash %d (was %d), booked %d (was %d), cost %d", co.Cash, cash, co.Expenditures[0][ExpenditureVehiclePurchases], booked, cost)
	}
	if !m.IsPlayerCompany(int(company)) {
		m.PlayerCompanies[0] = company // as the player's
	}
	co.Cash = int64(cost) - 1
	n, _ := m.Trains()
	if _, ok := buy(); ok {
		t.Error("bought without the cash")
	}
	if after, _ := m.Trains(); len(after) != len(n) || co.Cash != int64(cost)-1 {
		t.Errorf("refused purchase changed things: %d trains (was %d), cash %d", len(after), len(n), co.Cash)
	}
}

// TestLiveVehicleList: read from the entities right after loading, the list
// equals the save's own snapshot field for field (dayCreated included); a
// bought vehicle then shows up as one more row, created today.
func TestLiveVehicleList(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	live := m.LiveVehicleList()
	if len(live) != len(m.VehicleListData) {
		t.Fatalf("%d live rows, the save has %d", len(live), len(m.VehicleListData))
	}
	byHead := map[uint16]int{}
	for i, v := range m.VehicleListData {
		byHead[v.HeadID] = i
	}
	for _, v := range live {
		if s := m.VehicleListData[byHead[v.HeadID]]; v != s {
			t.Errorf("head %d: live %+v\nsave %+v", v.HeadID, v, s)
		}
	}
	trains, _ := m.Trains()
	company, objID := m.Entities[trains[0].Head].Owner, m.Entities[trains[0].Cars[0][0].Body].ObjectID
	var h uint16
	if _, ok := m.CreateVehicle(&Cmd{Company: company, Objects: om}, CmdApply, objID, true, &h); !ok {
		t.Fatal("buying refused")
	}
	rows := m.VehicleListRows()
	if len(rows) != len(live)+1 {
		t.Fatalf("%d rows after buying, want %d", len(rows), len(live)+1)
	}
	found := false
	for _, r := range rows {
		if r.HeadID == h {
			found = r.DayCreated == m.CurrentDay && r.AgeYears == 0
		}
	}
	if !found {
		t.Error("the bought vehicle's row is missing or not created today")
	}
}
