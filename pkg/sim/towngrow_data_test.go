package sim

import (
	"fmt"
	"github.com/LaPingvino/goloco/pkg/assets"
	"math"
	"strings"
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
					// maxQty only where cargo is carried: createBogie leaves a bogie's secondaryCargo.maxQty
					// unset and OpenLoco's createEntity zeroes the slot, but the vanilla save keeps a
					// reused slot's old value (a Hopper Truck's front bogie: 60 with no accepted types)
					a.Cargo.AcceptedTypes != 0 && a.Cargo.MaxQty != b.Cargo.MaxQty || a.ColourPrimary != b.ColourPrimary || a.ColourSecondary != b.ColourSecondary {
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
	saved := firstRailTrain(t, m, trains, 2) // a train with wagons (road vehicles take no trailers by default)
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
// started (vehicleChangeRunningMode); then the game runs — the vehicles, the
// date (industries produce daily), the tile loop (town buildings produce) —:
// it leaves, travels, calls at both stations, and earns: cargo the town and
// the industries made is delivered and paid to its company.
func TestPlayableVehicleEndToEnd(t *testing.T) {
	m0, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m0}
	m := &s.Map
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
	for _, o := range m.OrderRing(oh.OrderTableOffset, 0) { // all of them, the waits for a full load too
		orders = append(orders, append([]byte(nil), m.Orders[o.Offset:o.Offset+orderSizes[o.Type()]]...))
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
	cash0 := m.Companies[c.Company].Cash
	waiting := func() (n int) {
		for id := range m.StationNames {
			for _, cg := range m.StationNames[id].Cargo {
				n += int(cg.Quantity)
			}
		}
		return
	}
	waiting0 := waiting()
	var earned int64
	prevRev := m.Entities[tr.Veh2].CurMonthRevenue
	for tick := 0; tick < 200000 && (len(called) < 2 || earned == 0); tick++ {
		m.UpdateVehicles(om)
		s.UpdateTiles(func(int, int) bool { return false }, om)
		tr, err := m.TrainOf(h)
		if err != nil {
			t.Fatal(err)
		}
		if rev := m.Entities[tr.Veh2].CurMonthRevenue; rev != prevRev {
			if rev > prevRev {
				if earned == 0 {
					t.Logf("tick %d: first payment %d", tick, rev-prevRev)
				}
				earned += int64(rev - prevRev)
			}
			prevRev = rev
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
	t.Logf("moved up to %.0f px; called at %v; earned %d (company cash %d → %d); cargo waiting at stations %d → %d", moved, called, earned, cash0, m.Companies[c.Company].Cash, waiting0, waiting())
	if moved < 100 || len(called) < 2 {
		t.Errorf("the bought train did not run its route: moved %.0f px, called at %v", moved, called)
	}
	if earned <= 0 {
		t.Error("the bought train earned nothing")
	}
	// The Finances tab's "last income": the paid run recorded on Vehicle1
	// (updateLastIncomeStats), today, its profits adding up to what was earned.
	li := m.Entities[tr.Veh1].LastIncome
	var sum int64
	lines := 0
	for i := 0; i < 4; i++ {
		if li.CargoTypes[i] != 0xFF {
			sum += int64(li.CargoProfits[i])
			lines++
			t.Logf("last income: cargo %d × %d over %d tiles in %d days = %d", li.CargoTypes[i], li.CargoQtys[i], li.CargoDistances[i], li.CargoAges[i], li.CargoProfits[i])
		}
	}
	if earned > 0 && (lines == 0 || li.Day <= 0 || li.Day > int32(m.CurrentDay) || sum != earned) {
		t.Errorf("last income: day %d (today %d), %d lines summing to %d, earned %d", li.Day, m.CurrentDay, lines, sum, earned)
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
	h := m.Entities[firstRailTrain(t, m, trains, 1).Head] // the trains tab and its track
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
			tr, _ := m.TrainOf(h)
			if want := int(m.Entities[tr.Cars[0][0].Front].Reliability / 256); r.Reliability != want || want == 0 {
				t.Errorf("the bought train's reliability %d%%, want %d%% (recalculateTrainMinReliability)", r.Reliability, want)
			}
		}
	}
	if !found {
		t.Error("the bought vehicle's row is missing or not created today")
	}
}

// TestRailPlacementAtCursor: for each rail vehicle on the title map, the
// cursor put on the screen point of each sub-position of the piece its
// Vehicle2 stands on (gameToScreen at rotation 0) gives that sub-position —
// or, when it faces the other way for the view, the same point counted from
// the piece's far end with the direction reversed; and the vehicle, picked
// up, can be placed back where the cursor points (VehiclePlace, query).
func TestRailPlacementAtCursor(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	toScreen := func(x, y, z int32) (int, int) { return int(y - x), int((y+x)>>1 - z) }
	trains, _ := m.Trains()
	checked, placeable, tried := 0, 0, 0
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.TileX == -1 || h.Mode != 0 || !m.canBeModified(tr) {
			continue
		}
		v2 := m.Entities[tr.Veh2]
		tad := v2.TrackAndDirection &^ (1 << 2)
		tile := m.TileAt(int(v2.TileX)/32, int(v2.TileY)/32)
		var te *scenario.TrackElement
		for i := range tile.Tracks {
			e := &tile.Tracks[i]
			if e.TrackID == uint8(tad>>3&0x3F) && e.Rotation == uint8(tad&3) && e.SequenceIndex == 0 {
				te = e
			}
		}
		if te == nil || v2.TrackAndDirection&(1<<2) != 0 {
			continue // the start tile holds sequence 0 only for an unreversed vehicle
		}
		moves := TrackSubPosition(tad)
		for k := range moves {
			cx, cy := toScreen(int32(v2.TileX)+int32(moves[k].X), int32(v2.TileY)+int32(moves[k].Y), int32(v2.TileBaseZ)*4+int32(moves[k].Z))
			p, ok := m.RailPlacementAtCursor(h.ID, int32(v2.TileX), int32(v2.TileY), te, 0, cx, cy, toScreen)
			if !ok {
				t.Fatalf("vehicle %d: no placement", h.ID)
			}
			switch p.TrackAndDirection {
			case tad:
				if int(p.TrackProgress) != k || p.X != int32(v2.TileX) || p.Y != int32(v2.TileY) {
					t.Errorf("vehicle %d sub %d: progress %d at %d,%d", h.ID, k, p.TrackProgress, p.X, p.Y)
				}
			case tad ^ 4:
				if want := min(len(moves)-k, len(moves)-1); int(p.TrackProgress) != want {
					t.Errorf("vehicle %d sub %d reversed: progress %d, want %d", h.ID, k, p.TrackProgress, want)
				}
			default:
				t.Errorf("vehicle %d: tad %#x from %#x", h.ID, p.TrackAndDirection, tad)
			}
			checked++
		}
		own := &Cmd{Company: h.Owner, Objects: om}
		m.VehiclePickup(own, CmdApply, h.ID)
		k := len(moves) / 2
		cx, cy := toScreen(int32(v2.TileX)+int32(moves[k].X), int32(v2.TileY)+int32(moves[k].Y), int32(v2.TileBaseZ)*4+int32(moves[k].Z))
		p, _ := m.RailPlacementAtCursor(h.ID, int32(v2.TileX), int32(v2.TileY), te, 0, cx, cy, toScreen)
		tried++
		if _, ok := m.VehiclePlace(own, CmdApply, p); ok { // applied: a query returns before any check (as upstream)
			placeable++
		} else {
			t.Logf("vehicle %d: not placeable at %+v", h.ID, p)
		}
	}
	t.Logf("%d sub-positions checked; %d of %d picked-up vehicles placeable where the cursor points", checked, placeable, tried)
	if checked == 0 || tried == 0 || placeable == 0 {
		t.Fatal("nothing checked")
	}
}

// TestRoadPlacementAtCursor: TestRailPlacementAtCursor for road vehicles —
// the cursor on each getRoadPlacementSubPositon point of the road piece a
// vehicle's Vehicle2 stands on gives that point (or the same point from the
// far end, reversed), and a picked-up vehicle can be placed back there.
func TestRoadPlacementAtCursor(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	toScreen := func(x, y, z int32) (int, int) { return int(y - x), int((y+x)>>1 - z) }
	trains, _ := m.Trains()
	checked, placeable, tried := 0, 0, 0
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.TileX == -1 || h.Mode != 1 || !m.canBeModified(tr) {
			continue
		}
		v2 := m.Entities[tr.Veh2]
		if v2.TrackAndDirection&(1<<2) != 0 {
			continue // the start tile holds sequence 0 only for an unreversed vehicle
		}
		tad := v2.TrackAndDirection & 0x7F // the road piece and direction, without the lane bits
		tile := m.TileAt(int(v2.TileX)/32, int(v2.TileY)/32)
		var re *scenario.RoadElement
		for i := range tile.Roads {
			e := &tile.Roads[i]
			if e.RoadID == uint8(tad>>3&0xF) && e.Rotation == uint8(tad&3) && e.SequenceIndex == 0 {
				re = e
			}
		}
		if re == nil {
			continue
		}
		moves := RoadPlacementSubPosition(tad)
		for k := range moves {
			cx, cy := toScreen(int32(v2.TileX)+int32(moves[k].X), int32(v2.TileY)+int32(moves[k].Y), int32(re.BaseZ)*4+int32(moves[k].Z))
			p, ok := m.RoadPlacementAtCursor(h.ID, int32(v2.TileX), int32(v2.TileY), re, 0, cx, cy, toScreen)
			if !ok {
				t.Fatalf("vehicle %d: no placement", h.ID)
			}
			switch p.TrackAndDirection {
			case tad:
				if int(p.TrackProgress) != k || p.X != int32(v2.TileX) || p.Y != int32(v2.TileY) {
					t.Errorf("vehicle %d sub %d: progress %d at %d,%d", h.ID, k, p.TrackProgress, p.X, p.Y)
				}
			case tad ^ 4:
				if want := min(len(moves)-k, len(moves)-1); int(p.TrackProgress) != want {
					t.Errorf("vehicle %d sub %d reversed: progress %d, want %d", h.ID, k, p.TrackProgress, want)
				}
			default:
				t.Errorf("vehicle %d: tad %#x from %#x", h.ID, p.TrackAndDirection, tad)
			}
			checked++
		}
		own := &Cmd{Company: h.Owner, Objects: om}
		m.VehiclePickup(own, CmdApply, h.ID)
		k := len(moves) / 2
		cx, cy := toScreen(int32(v2.TileX)+int32(moves[k].X), int32(v2.TileY)+int32(moves[k].Y), int32(re.BaseZ)*4+int32(moves[k].Z))
		p, _ := m.RoadPlacementAtCursor(h.ID, int32(v2.TileX), int32(v2.TileY), re, 0, cx, cy, toScreen)
		tried++
		if _, ok := m.VehiclePlace(own, CmdApply, p); ok { // applied: a query returns before any check (as upstream)
			placeable++
		} else {
			t.Logf("vehicle %d: not placeable at %+v", h.ID, p)
		}
	}
	t.Logf("%d sub-positions checked; %d of %d picked-up road vehicles placeable where the cursor points", checked, placeable, tried)
	if checked == 0 || tried == 0 || placeable == 0 {
		t.Fatal("nothing checked")
	}
}

// TestVehicleWaterPickupPlace: the title map has no docks and no boats, so
// the test makes both — a dock element for a company's station on a tile
// whose boat position (the dock object's, rotated, + 32, 32) is on water,
// and a ship bought for that company (CreateVehicle). VehiclePlaceWater puts
// it there as a ghost (the ghost flag) and VehiclePickupWater lifts it (the
// dock free again); then for real: at the boat position, on water, facing
// out of the dock, stopped, the station its, the dock occupied; placing it
// twice is refused; lifted again, every component is off the map; a dock
// whose boat position is on land refuses it (noWater).
func TestVehicleWaterPickupPlace(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	var shipObj uint16 = 0xFFFF
	for i, v := range om.Vehicles {
		if v != nil && v.Mode == objects.TransportModeWater {
			shipObj = uint16(i)
			break
		}
	}
	if shipObj == 0xFFFF || len(om.DockObjects) == 0 || om.DockObjects[0] == nil {
		t.Skip("no ship or dock objects")
	}
	dockObj := om.DockObjects[0]
	var company uint8 = 0xFF
	var stationID uint16
	for i, st := range m.StationNames {
		if st.Name != 0 && st.Owner < 16 {
			company, stationID = st.Owner, uint16(i)
			break
		}
	}
	if company == 0xFF {
		t.Skip("no company station")
	}
	addDock := func(wantWater bool) (int32, int32, int32, bool) {
		for ty := 2; ty < m.Height-4; ty++ {
			for tx := 2; tx < m.Width-4; tx++ {
				x, y := int32(tx)*32, int32(ty)*32
				ox, oy := rotOffset(dockObj.BoatPosition[0], dockObj.BoatPosition[1], 0)
				if _, water := m.GetHeight(x+int32(ox)+32, y+int32(oy)+32); (water != 0) != wantWater || len(m.TileAt(tx, ty).Stations) > 0 {
					continue
				}
				z := int32(m.TileAt(tx, ty).BaseZ) * 4 // the element's base height (whole small-z steps)
				m.TileAt(tx, ty).Stations = append(m.TileAt(tx, ty).Stations, scenario.StationElement{StationType: 3, StationID: stationID, Owner: company, BaseZ: uint8(z / 4)})
				return x, y, z, true
			}
		}
		return 0, 0, 0, false
	}
	dx, dy, dz, ok := addDock(true)
	if !ok {
		t.Skip("no tile with water at a dock's boat position")
	}
	own := &Cmd{Company: company, Objects: om}
	var head uint16
	if _, ok := m.CreateVehicle(own, CmdApply, shipObj, true, &head); !ok {
		t.Fatalf("buying ship %d for company %d refused", shipObj, company)
	}
	tr, _ := m.TrainOf(head)
	p := VehicleWaterPlacement{X: dx, Y: dy, Z: dz, Head: head}
	if _, ok := m.VehiclePlaceWater(own, CmdApply|CmdGhost, p); !ok {
		t.Fatal("ghost placement refused")
	}
	if m.Entities[tr.Head].Flags38&flags38IsGhost == 0 {
		t.Error("the ghost isn't marked")
	}
	m.VehiclePickupWater(own, CmdApply|CmdGhost, head)
	if m.dockAt(dx, dy, dz).Flag6 {
		t.Error("lifting the ghost left the dock occupied")
	}
	if _, ok := m.VehiclePlaceWater(own, CmdApply, p); !ok {
		t.Fatal("placement refused where its ghost stood")
	}
	el := m.dockAt(dx, dy, dz)
	hd, v2 := m.Entities[tr.Head], m.Entities[tr.Veh2]
	if !el.Flag6 || hd.TileX == -1 || hd.Status != StatusStopped || hd.StationID != stationID || hd.VehicleFlags&VehicleFlagCommandStop == 0 {
		t.Errorf("placed: dock %v, tile %d, status %d, station %d", el.Flag6, hd.TileX, hd.Status, hd.StationID)
	}
	ox, oy := rotOffset(dockObj.BoatPosition[0], dockObj.BoatPosition[1], 0)
	if int32(v2.PosX) != dx+int32(ox)+32 || int32(v2.PosY) != dy+int32(oy)+32 {
		t.Errorf("veh2 at %d,%d, want the boat position %d,%d", v2.PosX, v2.PosY, dx+int32(ox)+32, dy+int32(oy)+32)
	}
	if body := m.Entities[tr.Cars[0][0].Body]; body.SpriteYaw != 16 { // ((rotation 0 + 1) & 3) * 16
		t.Errorf("body yaw %d, want 16", body.SpriteYaw)
	}
	if _, ok := m.VehiclePlaceWater(own, CmdApply, p); ok {
		t.Error("placed twice")
	}
	m.VehiclePickupWater(own, CmdApply, head)
	tr, _ = m.TrainOf(head)
	m.eachComponent(tr, func(i int) {
		if m.Entities[i].TileX != -1 || m.Entities[i].PosX != -32768 {
			t.Errorf("component %d still on the map after the pickup", i)
		}
	})
	if lx, ly, lz, ok := addDock(false); ok {
		if _, ok := m.VehiclePlaceWater(own, CmdApply, VehicleWaterPlacement{X: lx, Y: ly, Z: lz, Head: head}); ok {
			t.Error("placed at a dock whose boat position is on land")
		}
	}
	t.Logf("ship object %d of company %d: placed at the dock at tile %d,%d and lifted", shipObj, company, dx/32, dy/32)
}

// TestWaterPathStep: on the title map's water, a boat stepping by
// waterPathStep (each step the centre of a neighbouring tile of the same
// water level) from a water tile towards another one 6 tiles off on the same
// water reaches it, never leaving that water; the first tiles found by a
// breadth-first walk over equal water levels.
func TestWaterPathStep(t *testing.T) {
	m, _ := loadDataMap(t, "title.dat")
	var sx, sy, gx, gy = -1, -1, -1, -1
	var level uint8
	for y := 2; y < m.Height-2 && sx < 0; y++ {
		for x := 2; x < m.Width-2 && sx < 0; x++ {
			l := m.TileAt(x, y).WaterLevel
			if l == 0 {
				continue
			}
			// a tile of the same water 6 steps away (breadth-first over 4-neighbours)
			type p struct{ x, y, d int }
			seen := map[[2]int]bool{{x, y}: true}
			q := []p{{x, y, 0}}
			for len(q) > 0 && gx < 0 {
				c := q[0]
				q = q[1:]
				if c.d == 6 {
					gx, gy = c.x, c.y
					break
				}
				for i := 0; i < 4; i++ {
					nx, ny := c.x+int(RotationOffset[i][0])/32, c.y+int(RotationOffset[i][1])/32
					if nt := m.TileAt(nx, ny); nt != nil && nt.WaterLevel == l && !seen[[2]int{nx, ny}] {
						seen[[2]int{nx, ny}] = true
						q = append(q, p{nx, ny, c.d + 1})
					}
				}
			}
			if gx >= 0 {
				sx, sy, level = x, y, l
			}
		}
	}
	if sx < 0 {
		t.Skip("no water body 6 tiles across on the title map")
	}
	px, py, pz := int32(sx)*32+16, int32(sy)*32+16, int32(level)*16
	var yaw uint8
	for step := 0; step < 20; step++ {
		if int(px)>>5 == gx && int(py)>>5 == gy {
			t.Logf("from tile %d,%d to %d,%d (water level %d) in %d steps", sx, sy, gx, gy, level, step)
			return
		}
		r := m.waterPathStep(px, py, pz, yaw, gx, gy, nil)
		ntx, nty := int(r.HeadX)>>5, int(r.HeadY)>>5
		if d := abs32(int32(ntx-int(px)>>5)) + abs32(int32(nty-int(py)>>5)); d != 1 {
			t.Fatalf("step %d: from tile %d,%d to %d,%d — not a neighbour", step, int(px)>>5, int(py)>>5, ntx, nty)
		}
		if m.TileAt(ntx, nty).WaterLevel != level {
			t.Fatalf("step %d: onto tile %d,%d of water level %d", step, ntx, nty, m.TileAt(ntx, nty).WaterLevel)
		}
		px, py = r.HeadX, r.HeadY
	}
	t.Fatalf("not at tile %d,%d after 20 steps (at %d,%d)", gx, gy, int(px)>>5, int(py)>>5)
}

// TestBoatSailsDockToDock: two docks on the title map's lake (station
// elements added for two of a company's stations, each listed in its
// station's tiles), a ship placed at the first (VehiclePlaceWater) with the
// order "stop at" the second, started: within the ticks it sails there over
// water, takes the second dock (flag_6), frees the first, and begins
// unloading — the vehicle tick now runs water vehicles.
func TestBoatSailsDockToDock(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	var shipObj uint16 = 0xFFFF
	for i, v := range om.Vehicles {
		if v != nil && v.Mode == objects.TransportModeWater {
			shipObj = uint16(i)
			break
		}
	}
	if shipObj == 0xFFFF || len(om.DockObjects) == 0 || om.DockObjects[0] == nil {
		t.Skip("no ship or dock objects")
	}
	dockObj := om.DockObjects[0]
	var company uint8 = 0xFF
	var stations []uint16
	for i, st := range m.StationNames {
		if st.Name != 0 && st.Owner < 16 && (company == 0xFF || st.Owner == company) && int(st.TileSize) < len(st.Tiles) {
			company = st.Owner
			stations = append(stations, uint16(i))
			if len(stations) == 2 {
				break
			}
		}
	}
	if len(stations) < 2 {
		t.Skip("no company with two stations")
	}
	ox, oy := rotOffset(dockObj.BoatPosition[0], dockObj.BoatPosition[1], 0)
	// dock tiles: shore tiles whose boat position lies on the water level 4 lake of TestWaterPathStep, some tiles apart
	var docks [][3]int32
	for y := 1; y < 20 && len(docks) < 2; y++ {
		for x := 66; x < 90 && len(docks) < 2; x++ {
			t := m.TileAt(x, y)
			bx, by := int32(x)*32+int32(ox)+32, int32(y)*32+int32(oy)+32
			bt := m.TileAt(int(bx>>5), int(by>>5))
			if t == nil || bt == nil || t.WaterLevel != 0 || bt.WaterLevel != 4 || len(t.Stations) > 0 {
				continue
			}
			if len(docks) == 1 && abs32(int32(x)*32-docks[0][0])+abs32(int32(y)*32-docks[0][1]) < 5*32 {
				continue
			}
			docks = append(docks, [3]int32{int32(x) * 32, int32(y) * 32, int32(bt.WaterLevel) * 16}) // a dock stands at the water's height
		}
	}
	if len(docks) < 2 {
		t.Skip("no two shore tiles on the lake")
	}
	for i, d := range docks {
		tl := m.TileAt(int(d[0]>>5), int(d[1]>>5))
		tl.Stations = append(tl.Stations, scenario.StationElement{StationType: 3, StationID: stations[i], Owner: company, BaseZ: uint8(d[2] / 4)})
		st := &m.StationNames[stations[i]]
		st.Tiles[st.TileSize] = [3]int16{int16(d[0]), int16(d[1]), int16(d[2])}
		st.TileSize++
	}
	own := &Cmd{Company: company, Objects: om}
	var head uint16
	if _, ok := m.CreateVehicle(own, CmdApply, shipObj, true, &head); !ok {
		t.Fatal("buying the ship refused")
	}
	if _, ok := m.VehiclePlaceWater(own, CmdApply, VehicleWaterPlacement{X: docks[0][0], Y: docks[0][1], Z: docks[0][2], Head: head}); !ok {
		t.Fatal("placing the ship refused")
	}
	if _, ok := m.VehicleOrderInsert(own, CmdApply, head, 0, OrderStopAtRaw(int(stations[1]))); !ok {
		t.Fatal("the order refused")
	}
	if _, ok := m.VehicleChangeRunningMode(own, CmdApply, head, RunningModeStart); !ok {
		t.Fatal("starting refused")
	}
	tr, _ := m.TrainOf(head)
	startX, startY := m.Entities[tr.Veh2].PosX, m.Entities[tr.Veh2].PosY
	for tick := 0; tick < 6000; tick++ {
		m.UpdateVehicles(om)
		h := m.Entities[tr.Head]
		if h.StationID == stations[1] && (h.Status == headStatusUnloading || h.Status == headStatusLoading) {
			v2 := m.Entities[tr.Veh2]
			if !m.dockAt(docks[1][0], docks[1][1], docks[1][2]).Flag6 || m.dockAt(docks[0][0], docks[0][1], docks[0][2]).Flag6 {
				t.Errorf("docks' flag_6: first %v, second %v", m.dockAt(docks[0][0], docks[0][1], docks[0][2]).Flag6, m.dockAt(docks[1][0], docks[1][1], docks[1][2]).Flag6)
			}
			t.Logf("tick %d: at the second dock (station %d), from %d,%d to %d,%d; status %d", tick, stations[1], startX, startY, v2.PosX, v2.PosY, h.Status)
			return
		}
		if v2 := m.Entities[tr.Veh2]; tick%500 == 0 {
			t.Logf("tick %d: at %d,%d yaw %d speed %.1f mph, status %d, station %d; head target %d,%d", tick, v2.PosX, v2.PosY, v2.SpriteYaw, float64(v2.CurrentSpeed)/65536, h.Status, h.StationID, h.PosX, h.PosY)
		}
	}
	v2 := m.Entities[tr.Veh2]
	t.Fatalf("not at the second dock after 6000 ticks: at %d,%d (from %d,%d), status %d, station %d", v2.PosX, v2.PosY, startX, startY, m.Entities[tr.Head].Status, m.Entities[tr.Head].StationID)
}

// TestDeliverCargoToNearbyStations: on the title map, cargo handed in at a
// station's own tile reaches the stations around it that take it from
// producers — each its rating-squared share, capped at rating / 256 of the
// quantity, one more over 66 — the waiting quantity growing by it; a cargo no
// station takes, nowhere.
func TestDeliverCargoToNearbyStations(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			for _, el := range m.TileAt(x, y).Stations {
				if int(el.StationID) >= len(m.StationNames) || el.Ghost || el.AiAllocated {
					continue
				}
				for cargo := range m.StationNames[el.StationID].Cargo {
					cs := m.StationNames[el.StationID].Cargo[cargo]
					if cs.Flags&stationCargoAcceptedFromProducer == 0 || cs.Rating == 0 {
						continue
					}
					px, py := int32(x)*32, int32(y)*32
					found := m.findStationsForCargoType(uint8(cargo), px, py, 1, 1)
					before := map[uint16]uint16{}
					total := 0
					for _, f := range found {
						before[f.id] = m.StationNames[f.id].Cargo[cargo].Quantity
						total += int(f.rating) * int(f.rating)
					}
					got := m.DeliverCargoToNearbyStations(uint8(cargo), 100, px, py, 1, 1, om)
					sum := 0
					for _, f := range found {
						r := int(f.rating)
						want := min(r*r*100/total, r*100/256)
						if r > 66 {
							want++
						}
						sum += want
						if q := m.StationNames[f.id].Cargo[cargo].Quantity; int(q)-int(before[f.id]) != want {
							t.Errorf("station %d (rating %d): %d → %d, want +%d", f.id, r, before[f.id], q, want)
						}
					}
					if int(got) != min(sum, 100) {
						t.Errorf("delivered %d, want %d", got, min(sum, 100))
					}
					t.Logf("cargo %d at tile %d,%d: %d stations took %d of 100", cargo, x, y, len(found), got)
					return
				}
			}
		}
	}
	t.Skip("no station taking cargo from producers on the title map")
}

// TestBuildingsProduce: the title map's tile loop run for 20 passes (each
// tile is visited once per 256 ticks) — the towns' buildings produce
// (town var_19C's produced counters grow), some of it reaches stations
// (the delivered counters grow too), and the stations' waiting cargo rises.
func TestBuildingsProduce(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	var produced0, delivered0 [2]int
	for _, tw := range s.Towns {
		for i := 0; i < 2; i++ {
			produced0[i] += int(tw.Var19C[i][0])
			delivered0[i] += int(tw.Var19C[i][1])
		}
	}
	waiting := func() int {
		n := 0
		for _, st := range s.StationNames {
			for _, c := range st.Cargo {
				n += int(c.Quantity)
			}
		}
		return n
	}
	w0 := waiting()
	for i := 0; i < 256*20; i++ {
		s.UpdateTiles(nil, om)
	}
	var produced, delivered [2]int
	for _, tw := range s.Towns {
		for i := 0; i < 2; i++ {
			produced[i] += int(tw.Var19C[i][0])
			delivered[i] += int(tw.Var19C[i][1])
		}
	}
	for i := 0; i < 2; i++ {
		produced[i] -= produced0[i]
		delivered[i] -= delivered0[i]
	}
	t.Logf("produced %v, delivered %v; waiting at stations %d → %d", produced, delivered, w0, waiting())
	if produced[0] <= 0 {
		t.Fatal("the towns' buildings produced no passengers")
	}
	if delivered[0] <= 0 || waiting() <= w0 {
		t.Errorf("nothing reached a station: delivered %v, waiting %d → %d", delivered, w0, waiting())
	}
}

// TestIndustriesProduce: 60 days of the title map's industries
// (Industry::updateDaily) — their monthly produced totals grow, and those
// with stations for their cargo deliver some of it.
func TestIndustriesProduce(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	before := make([][2]uint16, len(m.Industries))
	delivered0 := make([][2]uint16, len(m.Industries))
	withStations := 0
	for i, ind := range m.Industries {
		before[i], delivered0[i] = ind.ProducedMonthlyTotal, ind.DeliveredMonthlyTotal
		for _, s := range ind.ProducedStations[0] {
			if s != stationIDNull {
				withStations++
				break
			}
		}
	}
	for day := 0; day < 60; day++ {
		m.industriesUpdateDaily(om)
	}
	produced, delivered := 0, 0
	for i, ind := range m.Industries {
		for k := 0; k < 2; k++ {
			produced += int(ind.ProducedMonthlyTotal[k]) - int(before[i][k])
			delivered += int(ind.DeliveredMonthlyTotal[k]) - int(delivered0[i][k])
		}
	}
	t.Logf("%d industries (%d with a station for their first cargo): produced %d, delivered %d in 60 days", len(m.Industries), withStations, produced, delivered)
	if produced <= 0 {
		t.Fatal("the industries produced nothing")
	}
	if withStations > 0 && delivered <= 0 {
		t.Error("industries with stations delivered nothing")
	}
}

// TestIndustryStationsRefresh: one whole tile-loop round (256 passes) on the
// title map rebuilds every industry's producedCargoStatsStation from the
// stations around its buildings — and gives back the lists the game saved.
func TestIndustryStationsRefresh(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	saved := make([][2][4]uint16, len(m.Industries))
	for i, ind := range m.Industries {
		saved[i] = ind.ProducedStations
	}
	s := &State{Map: *m}
	s.tileUpdateX, s.tileUpdateY = 0, 0
	for i := range s.Industries { // a fresh round: no stations carried over, the lists wiped
		s.Industries[i].StationsInRange = [32]uint32{}
		s.Industries[i].ProducedStations = [2][4]uint16{}
	}
	for pass := 0; pass < 256; pass++ {
		s.UpdateTiles(func(int, int) bool { return false }, om)
	}
	same, with := 0, 0
	for i, ind := range s.Industries {
		if ind.ProducedStations[0][0] != stationIDNull {
			with++
		}
		if ind.ProducedStations == saved[i] {
			same++
		} else if same+with < 200 {
			t.Logf("industry %d: saved %v, rebuilt %v", ind.ID, saved[i], ind.ProducedStations)
		}
	}
	t.Logf("%d of %d industries' station lists as saved (%d with a station)", same, len(s.Industries), with)
	if with == 0 {
		t.Fatal("no industry found a station")
	}
	if same*10 < len(s.Industries)*9 {
		t.Errorf("only %d of %d lists match the save", same, len(s.Industries))
	}
}

// TestIndustriesMonthly: the save's own months agree with
// Industry::updateMonthly (each industry's last history entry is its
// previous month's production ÷ 50, the percentage transported its
// delivered share) — so the fields are read right; then a month end moves
// the totals to the previous month's, appends to the history, and works
// the percentage out the same way, and removes the ones closed down with
// nothing left to make.
func TestIndustriesMonthly(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	pct := func(prod, del uint16) uint8 {
		tr := uint32(min(prod, del))
		if prod != 0 {
			tr = tr * 100 / uint32(prod)
		}
		return uint8(tr)
	}
	checked := 0
	for _, ind := range m.Industries {
		for k := 0; k < 2; k++ {
			if ind.HistorySize[k] == 0 {
				continue
			}
			checked++
			if got, want := ind.History[k][ind.HistorySize[k]-1], uint8(min(uint32(ind.ProducedPrevMonth[k]), 12750)/50); got != want {
				t.Errorf("industry %d cargo %d: saved history ends %d, previous month %d ÷ 50 = %d", ind.ID, k, got, ind.ProducedPrevMonth[k], want)
			}
			if got, want := ind.PercentTransported[k], pct(ind.ProducedPrevMonth[k], ind.DeliveredPrevMonth[k]); got != want {
				t.Errorf("industry %d cargo %d: saved %d%% transported, its totals give %d%%", ind.ID, k, got, want)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no industry has a history")
	}
	before := map[uint8]scenario.Industry{}
	removable := 0
	for _, ind := range m.Industries {
		before[ind.ID] = ind
		if ind.Flags&industryFlagClosingDown != 0 && ind.DailyProduction == [2]uint16{} {
			removable++
		}
	}
	n0 := len(m.Industries)
	m.industriesUpdateMonthly(om)
	if len(m.Industries) != n0-removable {
		t.Errorf("%d industries after the month, want %d (%d closed down with nothing to make: removed)", len(m.Industries), n0-removable, removable)
	}
	for _, ind := range m.Industries {
		b := before[ind.ID]
		if b.Flags&industryFlagIsGhost != 0 || b.Flags&industryFlagClosingDown != 0 && b.DailyProduction == [2]uint16{} {
			continue
		}
		for k := 0; k < 2; k++ {
			if ind.ProducedPrevMonth[k] != b.ProducedMonthlyTotal[k] || ind.ProducedMonthlyTotal[k] != 0 || ind.DeliveredPrevMonth[k] != b.DeliveredMonthlyTotal[k] {
				t.Errorf("industry %d cargo %d: the month's totals not rolled over", ind.ID, k)
			}
			if want := min(b.HistorySize[k]+1, 240); ind.HistorySize[k] != want {
				t.Errorf("industry %d cargo %d: history size %d, want %d", ind.ID, k, ind.HistorySize[k], want)
			}
			if ind.PercentTransported[k] != pct(ind.ProducedPrevMonth[k], ind.DeliveredPrevMonth[k]) {
				t.Errorf("industry %d cargo %d: percentage %d", ind.ID, k, ind.PercentTransported[k])
			}
		}
	}
	t.Logf("%d histories agree with the save's previous month; a month end rolled %d industries over, removed %d closed down", checked, len(m.Industries), removable)
}

// TestAirportMovementGraph (AW.b's data): every loaded airport's movement
// nodes and edges as upstream's AirportObject::load reads them — the edges
// join nodes that exist, their clear-masks name only edges that exist, and
// each airport has terminals and a takeoff end (a runway's or a
// heliport's), a runway its touchdown.
func TestAirportMovementGraph(t *testing.T) {
	_, om := loadDataMap(t, "title.dat")
	n := 0
	for _, a := range om.AirportObjects {
		if a == nil {
			continue
		}
		n++
		nodes, edges := len(a.MovementNodes), len(a.MovementEdges)
		if nodes == 0 || edges == 0 || edges > 32 {
			t.Errorf("%s: %d nodes, %d edges", a.Name, nodes, edges)
			continue
		}
		var flags uint16
		for _, nd := range a.MovementNodes {
			flags |= nd.Flags
		}
		for i, e := range a.MovementEdges {
			if int(e.CurNode) >= nodes || int(e.NextNode) >= nodes {
				t.Errorf("%s edge %d: %d → %d of %d nodes", a.Name, i, e.CurNode, e.NextNode, nodes)
			}
			if mask := uint32(1)<<edges - 1; edges < 32 && (e.MustBeClear|e.AtLeastOneClear)&^mask != 0 {
				t.Errorf("%s edge %d: masks %x / %x name edges past %d", a.Name, i, e.MustBeClear, e.AtLeastOneClear, edges)
			}
		}
		if flags&objects.AirportNodeTerminal == 0 {
			t.Errorf("%s: no terminal", a.Name)
		}
		if flags&objects.AirportNodeTakeoffBegin != 0 && flags&objects.AirportNodeTouchdown == 0 { // a runway (planes take off along it) has a touchdown; a heliport neither
			t.Errorf("%s: a runway but no touchdown", a.Name)
		}
		if flags&(objects.AirportNodeTakeoffEnd|objects.AirportNodeHeliTakeoffEnd) == 0 {
			t.Errorf("%s: no takeoff end", a.Name)
		}
		t.Logf("%s: %d nodes, %d edges, %d buildings, node flags %03x", a.Name, nodes, edges, len(a.Buildings), flags)
	}
	if n == 0 {
		t.Skip("no airport objects loaded on the title map")
	}
}

// TestAirportEdgeTargets (AW.b): every airport station on the title map —
// its airportStartPos holds an airport station element — and every edge of
// its airport's graph gives a target: taxiing nodes on the ground within
// the airport's few tiles of its start, the others at the flight heights.
func TestAirportEdgeTargets(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	airports := 0
	for id := range m.StationNames {
		el, pos, ok := m.airportStationElement(uint16(id))
		if !ok || el.StationType != 2 || m.StationNames[id].Name == 0 {
			continue
		}
		airports++
		a := om.AirportObjects[el.ObjectID]
		for e := range a.MovementEdges {
			flags, x, y, z, ok := m.AirportGetMovementEdgeTarget(om, uint16(id), uint8(e))
			if !ok {
				t.Fatalf("station %d edge %d: no target", id, e)
			}
			if flags&objects.AirportNodeTaxiing != 0 {
				if dx, dy := x-int32(pos[0]), y-int32(pos[1]); dx < -6*32 || dx > 8*32 || dy < -6*32 || dy > 8*32 || z < int32(pos[2]) || z > int32(pos[2])+64 {
					t.Errorf("station %d edge %d: taxiing target (%d, %d, %d) far from the airport at %v", id, e, x, y, z, pos)
				}
			} else if z != int32(pos[2])+255 && z != 960 {
				t.Errorf("station %d edge %d: airborne target height %d", id, e, z)
			}
		}
		t.Logf("station %d: %s, rotation %d, start %v, %d edges, occupied %032b", id, a.Name, el.Rotation, pos, len(a.MovementEdges), m.StationNames[id].AirportOccupiedEdges)
	}
	if airports == 0 {
		t.Skip("no airports on the title map")
	}
}

// TestAirplaneGetNewStatus (AW.b): every aircraft of the title map in the
// air or on an airport's edges has the status airplaneGetNewStatus gives
// for its edge — the one the game itself set on its last tick (the
// stopped / loading / unloading ones keep theirs) — and its approach
// parameters point at an airport or a station; the target speed is the one
// saved in Vehicle1.
func TestAirplaneGetNewStatus(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	checked, same := 0, 0
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.Mode != 2 || h.TileX == -1 {
			continue
		}
		switch h.Status {
		case headStatusStopped, headStatusLoading, headStatusUnloading:
			continue
		}
		checked++
		st, speed := m.airplaneGetNewStatus(tr, om)
		a := m.airplaneApproachParams(tr, om)
		if v1 := m.Entities[tr.Veh1].TargetSpeed; speed != v1 {
			t.Errorf("aircraft %d: target speed %d, Vehicle1's saved %d", h.ID, speed, v1)
		}
		if st == h.Status {
			same++
		} else {
			t.Errorf("aircraft %d (station %d, edge %d): status %d, saved %d", h.ID, h.StationID, h.AirportMovementEdge, st, h.Status)
		}
		t.Logf("aircraft %d: station %d edge %d status %d target speed %d; heading yaw %d to z %d, %d away", h.ID, h.StationID, h.AirportMovementEdge, st, speed, a.targetYaw, a.targetZ, a.manhattan)
	}
	if checked == 0 {
		t.Skip("no aircraft on the move on the title map")
	}
	t.Logf("%d of %d aircraft: the saved status", same, checked)
}

// TestAirportEdgeWalk (AW.b): the title map's taxiing aircraft, put at the
// entrance of station 82's airport (its own edge freed), walks the airport
// by airportGetNextMovementEdge / airplaneTakeEdge — entering by an entry
// edge, on to a terminal node; then, as it leaves (status taking off once
// the edges say so), on to the runway's end where it flies off (the edge
// null). Each edge taken is marked occupied and the one left freed.
func TestAirportEdgeWalk(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	var tr Train
	found := false
	for _, x := range trains {
		if h := m.Entities[x.Head]; h.Mode == 2 && h.StationID != stationIDNull && h.AirportMovementEdge != airportEdgeNull {
			tr, found = x, true
			break
		}
	}
	if !found {
		t.Skip("no aircraft at an airport")
	}
	head := &m.Entities[tr.Head]
	sid := head.StationID
	st := &m.StationNames[sid]
	st.AirportOccupiedEdges &^= 1 << head.AirportMovementEdge
	head.AirportMovementEdge = airportEdgeNull
	el, _, _ := m.airportStationElement(sid)
	a := om.AirportObjects[el.ObjectID]
	var path []uint8
	reachedTerminal := false
	for step := 0; step < 64; step++ {
		cur := head.AirportMovementEdge
		if cur != airportEdgeNull {
			if a.MovementNodes[a.MovementEdges[cur].NextNode].Flags&objects.AirportNodeTerminal != 0 && !reachedTerminal {
				reachedTerminal = true
				t.Logf("at a terminal after %v", path)
				head.Status = headStatusTaxiing2 // loaded and leaving
			}
			if s, _ := m.airplaneGetNewStatus(tr, om); reachedTerminal {
				head.Status = s
			}
		}
		next := m.airportGetNextMovementEdge(tr, cur, om)
		if next == airportEdgeNoValid {
			t.Fatalf("stuck after %v (occupied %032b)", path, st.AirportOccupiedEdges)
		}
		m.airplaneTakeEdge(tr, next, om)
		if next == airportEdgeNull {
			if !reachedTerminal {
				t.Fatalf("flew off before a terminal: %v", path)
			}
			if st.AirportOccupiedEdges&(1<<cur) != 0 {
				t.Errorf("edge %d still occupied after leaving", cur)
			}
			t.Logf("took off after %v", path)
			return
		}
		if st.AirportOccupiedEdges&(1<<next) == 0 || (cur != airportEdgeNull && cur != next && st.AirportOccupiedEdges&(1<<cur) != 0) {
			t.Errorf("edge %d → %d: occupancy %032b", cur, next, st.AirportOccupiedEdges)
		}
		path = append(path, next)
	}
	t.Fatalf("no take-off in 64 edges: %v", path)
}

// TestAircraftFlies (AW.b): the title map runs; its taxiing aircraft (at
// Large Airport 82) reaches the terminal and calls there, taxis out, takes
// off and climbs to the flying height, flies to its next airport and lands
// there — each status change logged.
func TestAircraftFlies(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	var id uint16
	found := false
	for _, x := range trains {
		if h := m.Entities[x.Head]; h.Mode == 2 && h.StationID != stationIDNull && h.AirportMovementEdge != airportEdgeNull {
			id, found = h.ID, true
			break
		}
	}
	if !found {
		t.Skip("no aircraft at an airport")
	}
	first := m.entity(id).StationID
	seen := map[uint8]bool{}
	lastStatus := uint8(0xFF)
	maxZ := int16(0)
	landedElsewhere := false
	for tick := 0; tick < 40000 && !landedElsewhere; tick++ {
		m.UpdateVehicles(om)
		h := m.entity(id)
		tr, _ := m.TrainOf(id)
		maxZ = max(maxZ, h.PosZ)
		if h.Status != lastStatus {
			t.Logf("tick %d: status %d, station %d, edge %d, at (%d, %d, %d), speed %d mph", tick, h.Status, h.StationID, h.AirportMovementEdge, h.PosX, h.PosY, h.PosZ, m.Entities[tr.Veh2].CurrentSpeed>>16)
			lastStatus = h.Status
		}
		seen[h.Status] = true
		if h.StationID != first && h.StationID != stationIDNull && (h.Status == headStatusUnloading || h.Status == headStatusLoading) {
			landedElsewhere = true
		}
	}
	t.Logf("highest %d; statuses %v", maxZ, seen)
	for _, s := range []uint8{headStatusUnloading, headStatusTakingOff, headStatusTravelling, headStatusLanding} {
		if !seen[s] {
			t.Errorf("never status %d", s)
		}
	}
	if maxZ < 200 {
		t.Errorf("never climbed: highest %d", maxZ)
	}
	if !landedElsewhere {
		t.Error("never landed at its next airport")
	}
}

// TestAircraftPickupPlace (AW.b part 4, 5P.g's air placement): the title
// map's aircraft, loading at its terminal, is lifted off its airport (its
// edge freed, off the map)
// and put back on a terminal node of airport 82 by the game commands — the
// first edge into that node taken and occupied, stopped, facing out —;
// started, it calls, taxis out and takes off.
func TestAircraftPickupPlace(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	var id uint16
	for _, x := range trains {
		if h := m.Entities[x.Head]; h.Mode == 2 && h.StationID != stationIDNull && h.AirportMovementEdge != airportEdgeNull {
			id = h.ID
			break
		}
	}
	if id == 0 {
		t.Skip("no aircraft at an airport")
	}
	for tick := 0; tick < 2000; tick++ { // only stopped or loading aircraft at rest can be lifted (canBeModified)
		tr0, _ := m.TrainOf(id)
		if m.entity(id).Status == headStatusLoading && m.Entities[tr0.Veh2].CurrentSpeed == 0 {
			break
		}
		m.UpdateVehicles(om)
	}
	h := m.entity(id)
	sid, edge := h.StationID, h.AirportMovementEdge
	c := &Cmd{Company: h.Owner, Objects: om, Players: func(uint8) bool { return true }}
	if _, ok := m.VehiclePickupAir(c, CmdApply, id); !ok {
		tr0, _ := m.TrainOf(id)
		t.Fatalf("pickup refused: status %d speed %d type %d var73 %d", h.Status, m.Entities[tr0.Veh2].CurrentSpeed, h.VehicleType, m.Entities[tr0.Veh2].Var73)
	}
	if h := m.entity(id); h.TileX != -1 || h.StationID != stationIDNull || m.StationNames[sid].AirportOccupiedEdges&(1<<edge) != 0 {
		t.Fatalf("after pickup: tileX %d station %d occupied %032b", h.TileX, h.StationID, m.StationNames[sid].AirportOccupiedEdges)
	}
	el, _, _ := m.airportStationElement(sid)
	ap := om.AirportObjects[el.ObjectID]
	node := -1
	for i, n := range ap.MovementNodes {
		if n.Flags&objects.AirportNodeTerminal != 0 {
			node = i
			break
		}
	}
	args := VehicleAirPlacement{StationID: sid, AirportNode: uint8(node), Head: id}
	if _, ok := m.VehiclePlaceAir(c, 0, args); !ok {
		t.Fatal("placing (query) refused")
	}
	if _, ok := m.VehiclePlaceAir(c, CmdApply, args); !ok {
		t.Fatal("placing refused")
	}
	h = m.entity(id)
	tr, _ := m.TrainOf(id)
	nx, ny, nz, _ := m.AirportMovementNodeLoc(om, sid, uint8(node))
	if h.Status != StatusStopped || h.StationID != sid || int32(h.PosX) != nx || int32(h.PosY) != ny || int32(h.PosZ) != nz ||
		m.StationNames[sid].AirportOccupiedEdges&(1<<h.AirportMovementEdge) == 0 || ap.MovementEdges[h.AirportMovementEdge].NextNode != uint8(node) {
		t.Fatalf("placed: status %d station %d at (%d,%d,%d) want (%d,%d,%d), edge %d", h.Status, h.StationID, h.PosX, h.PosY, h.PosZ, nx, ny, nz, h.AirportMovementEdge)
	}
	t.Logf("placed at terminal node %d by edge %d, body yaw %d", node, h.AirportMovementEdge, m.Entities[tr.Cars[0][0].Body].SpriteYaw)
	if _, ok := m.VehicleChangeRunningMode(c, CmdApply, id, RunningModeStart); !ok {
		t.Fatal("starting refused")
	}
	took := false
	for tick := 0; tick < 5000 && !took; tick++ {
		m.UpdateVehicles(om)
		took = m.entity(id).Status == headStatusTakingOff
	}
	if !took {
		t.Errorf("never took off: status %d, edge %d", m.entity(id).Status, m.entity(id).AirportMovementEdge)
	}
}

// TestRemoveIndustry (CR): the save's industry tiles (0x15, numTiles 0x14)
// each hold that industry's element at their height — the offsets are
// right —; removeIndustry takes one away (its buildings' elements on all
// their tiles, its fields back to plain ground, gone from the list), and
// the month's update removes an industry that has closed down with nothing
// left to make.
func TestRemoveIndustry(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	count := func(id uint8) (els, fields int) {
		for ty := 0; ty < m.Height; ty++ {
			for tx := 0; tx < m.Width; tx++ {
				tl := m.TileAt(tx, ty)
				for _, e := range tl.Industries {
					if e.IndustryID == id {
						els++
					}
				}
				if tl.IsIndustrial && tl.Variation == id {
					fields++
				}
			}
		}
		return
	}
	for _, ind := range m.Industries {
		for _, p := range ind.Tiles {
			tl := m.TileAt(int(p[0])>>5, int(p[1])>>5)
			found := false
			for _, e := range tl.Industries {
				if e.IndustryID == ind.ID && int32(e.BaseZ)*4 == int32(p[2]&0x7FFF) {
					found = true
				}
			}
			if !found {
				t.Fatalf("industry %d: no element of it at its tile %v", ind.ID, p)
			}
		}
	}
	victim := -1
	for i, ind := range m.Industries {
		if _, f := count(ind.ID); f > 0 && victim < 0 {
			victim = i
		}
	}
	if victim < 0 {
		victim = 0
	}
	id := m.Industries[victim].ID
	els, fields := count(id)
	n := len(m.Industries)
	cost, ok := m.RemoveIndustry(&Cmd{Company: CompanyNeutral, Objects: om}, CmdApply, id)
	if !ok {
		t.Fatal("refused")
	}
	e2, f2 := count(id)
	t.Logf("industry %d: %d elements, %d fields → %d, %d; cost %d", id, els, fields, e2, f2, cost)
	if e2 != 0 || f2 != 0 || len(m.Industries) != n-1 || m.industryByID(id) != nil {
		t.Errorf("after removal: %d elements, %d fields, %d industries (was %d)", e2, f2, len(m.Industries), n)
	}
	other := &m.Industries[0]
	other.Flags |= industryFlagClosingDown
	other.DailyProduction = [2]uint16{}
	oid := other.ID
	m.industriesUpdateMonthly(om)
	if m.industryByID(oid) != nil {
		t.Error("a closed-down industry with no production survived its month")
	}
}

// TestNewIndustryChoice (CR part 2): on the title map the towns' most
// common cargo is passengers; every industry's own place is too close to an
// industry and inside its type's cluster; each loaded type's cap is
// ((factor + 1) × total / 3) less a quarter (the random part is 0 — integer
// arithmetic, as upstream), drawing gPrng1 once; which types may start up
// this year is logged.
func TestNewIndustryChoice(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	mc := mostCommonBuildingCargoType(om)
	name := ""
	if c := om.CargoObjects[mc]; c != nil && c.Header != nil {
		name = strings.TrimSpace(string(c.Header.Name[:]))
	}
	if !strings.HasPrefix(strings.ToUpper(name), "PASS") {
		t.Errorf("most common building cargo %d %q, want passengers", mc, name)
	}
	for _, ind := range m.Industries {
		if !m.isTooCloseToNearbyIndustries(int32(ind.X), int32(ind.Y)) || m.isOutwithCluster(int32(ind.X), int32(ind.Y), ind.ObjectID) {
			t.Errorf("industry %d at its own place: not too close, or outwith its cluster", ind.ID)
		}
	}
	for i := 0; i < 16; i++ {
		obj := om.GetIndustryObjectByIndex(i)
		if obj == nil {
			continue
		}
		before := m.Rng1
		capN := m.capOfTypeOfIndustry(obj, m.NumberOfIndustries)
		want := (int32(m.NumberOfIndustries) + 1) * int32(obj.TotalOfTypeInScenario) / 3
		want -= want / 4
		check := before
		check.RandNext()
		if capN != want || m.Rng1 != check {
			t.Errorf("%s: cap %d want %d, or not one draw", obj.Name, capN, want)
		}
		have := 0
		for _, ind := range m.Industries {
			if int(ind.ObjectID) == i {
				have++
			}
		}
		t.Logf("%-22s designed %d–%d, total %2d → cap %2d, has %2d; may start this year (%d): %v", obj.Name, obj.DesignedYear, obj.ObsoleteYear, obj.TotalOfTypeInScenario, capN, have, m.CurrentYear, m.canIndustryObjBeCreated(obj, om))
	}
}

// TestFindRandomNewIndustryLocation (CR part 2b): for every loaded industry
// type, the search on the title map either finds a place that keeps all the
// type's rules (480 from any industry, its clusters, ground, water, towns,
// trees …) or none in its 250 tries; the same gPrng1 state finds the same
// place.
func TestFindRandomNewIndustryLocation(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	found := 0
	for i := 0; i < 16; i++ {
		obj := om.GetIndustryObjectByIndex(i)
		if obj == nil {
			continue
		}
		start := m.Rng1
		x, y, ok := m.findRandomNewIndustryLocation(uint8(i), om)
		after := m.Rng1
		m.Rng1 = start
		x2, y2, ok2 := m.findRandomNewIndustryLocation(uint8(i), om)
		if x != x2 || y != y2 || ok != ok2 || m.Rng1 != after {
			t.Errorf("%s: not deterministic", obj.Name)
		}
		if !ok {
			t.Logf("%-22s flags %08x: no place in 250 tries", obj.Name, obj.Flags)
			continue
		}
		found++
		f := obj.Flags
		tl := m.TileAt(int(x)>>5, int(y)>>5)
		bad := m.isTooCloseToNearbyIndustries(x, y) ||
			f&industryBuiltInClusters != 0 && m.isOutwithCluster(x, y, uint8(i)) ||
			f&industryBuiltOnFlatGround != 0 && m.mountainHeight(x, y) > industryFlatGroundMountainMax ||
			f&industryBuiltNearWater != 0 && m.countSurroundingWaterTiles(int(x)>>5, int(y)>>5) < industryTilesToBeNearWaterMin ||
			f&industryBuiltAwayFromWater != 0 && m.countSurroundingWaterTiles(int(x)>>5, int(y)>>5) > industryTilesToBeAwayWaterMax ||
			f&industryBuiltNearTrees != 0 && m.countSurroundingTrees(x, y) < industryNumTreesToBeNearTreesMin ||
			f&industryBuiltRequiresOpenSpace != 0 && m.countSurroundingTrees(x, y) > industryNumTreesToBeOpenSpaceMax ||
			f&industryBuiltOnHighGround != 0 && tl.BaseZ < industryHighGroundMin ||
			f&industryBuiltOnLowGround != 0 && tl.BaseZ > industryLowGroundMax
		if bad {
			t.Errorf("%s: place (%d, %d) breaks one of its rules", obj.Name, x, y)
		}
		t.Logf("%-22s flags %08x: tile (%d, %d), height %d, mountain %d, water %d, trees %d", obj.Name, f, x>>5, y>>5, tl.BaseZ, m.mountainHeight(x, y), m.countSurroundingWaterTiles(int(x)>>5, int(y)>>5), m.countSurroundingTrees(x, y))
	}
	if found == 0 {
		t.Error("no type found a place")
	}
}

// TestAllocateNewIndustry (CR part 2c): sub_454C91 refuses a place too close
// to an industry (an existing one's own place); at a place the search finds
// far enough off, allocateNewIndustry takes the lowest free slot — a new,
// empty industry kept in slot order —; with every slot taken it refuses.
func TestAllocateNewIndustry(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	first := m.Industries[0]
	if _, reason, ok := m.industryAllocateChecked(first.ObjectID, int32(first.X), int32(first.Y), [2]uint32{1, 2}, om); ok || reason != "too_close_to_another_industry" {
		t.Fatalf("at industry %d's own place: ok %v, %q", first.ID, ok, reason)
	}
	free := -1
	used := map[uint8]bool{}
	for _, ind := range m.Industries {
		used[ind.ID] = true
	}
	for id := 0; id < maxIndustries && free < 0; id++ {
		if !used[uint8(id)] {
			free = id
		}
	}
	var id uint8
	placed := false
	for tries := 0; tries < 20 && !placed; tries++ {
		for k := 0; k < 16 && !placed; k++ {
			if om.GetIndustryObjectByIndex(k) == nil {
				continue
			}
			x, y, ok := m.findRandomNewIndustryLocation(uint8(k), om)
			if !ok {
				continue
			}
			if got, _, ok := m.industryAllocateChecked(uint8(k), x, y, [2]uint32{7, 9}, om); ok {
				id, placed = got, true
				ind := m.industryByID(id)
				if int(id) != free || ind.ObjectID != uint8(k) || int32(ind.X) != x || ind.ProductionRate != 25 || ind.UnderConstruction != 0 ||
					ind.FoundingYear != m.CurrentYear || ind.HistorySize != [2]uint8{1, 1} || ind.ProducedStations[0][0] != stationIDNull || ind.Prng != [2]uint32{7, 9} {
					t.Errorf("new industry %+v (want slot %d)", *ind, free)
				}
				t.Logf("industry %d (%s) allocated at tile (%d, %d), town %d", id, om.GetIndustryObjectByIndex(k).Name, x>>5, y>>5, ind.Town)
			}
		}
	}
	if !placed {
		t.Fatal("no place far enough from the others in 20 rounds")
	}
	for i := 1; i < len(m.Industries); i++ {
		if m.Industries[i-1].ID >= m.Industries[i].ID {
			t.Fatal("industries not in slot order")
		}
	}
	for len(m.Industries) < maxIndustries {
		if _, ok := m.allocateNewIndustry(0, 0, 0, [2]uint32{}, 0); !ok {
			t.Fatal("refused while slots were free")
		}
	}
	if _, ok := m.allocateNewIndustry(0, 0, 0, [2]uint32{}, 0); ok {
		t.Error("a 129th industry allocated")
	}
}

// TestPlaceIndustryBuilding (CR part 2c): the industry objects' new fields
// read sensibly (buildings, colours, counts, production ranges); on the
// title map a new industry's first building is placed by
// placeIndustryBuilding — the query changes nothing, the apply puts its
// elements on every tile of its footprint at one level (the ground there
// flattened, the trees gone), the tile in the industry's list —; a land
// industry is refused on water.
func TestPlaceIndustryBuilding(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	for i := 0; i < 16; i++ {
		obj := om.GetIndustryObjectByIndex(i)
		if obj == nil {
			continue
		}
		if len(obj.Buildings) != int(obj.MaxNumBuildings) || obj.MinNumBuildings > obj.MaxNumBuildings || obj.MinNumBuildings == 0 ||
			obj.InitialProductionRate[0][0] > obj.InitialProductionRate[0][1] {
			t.Errorf("%s: buildings %v (%d–%d), production %v", obj.Name, obj.Buildings, obj.MinNumBuildings, obj.MaxNumBuildings, obj.InitialProductionRate)
		}
		for _, b := range obj.Buildings {
			if int(b) >= len(obj.Variations) {
				t.Errorf("%s: building type %d of %d", obj.Name, b, len(obj.Variations))
			}
		}
	}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	var id uint8
	var x, y int32
	var obj *objects.IndustryObject
	for k := 0; k < 16 && obj == nil; k++ {
		o := om.GetIndustryObjectByIndex(k)
		if o == nil || o.Flags&industryBuiltOnWater != 0 {
			continue
		}
		for try := 0; try < 10; try++ {
			px, py, ok := m.findRandomNewIndustryLocation(uint8(k), om)
			if !ok {
				continue
			}
			if got, _, ok := m.industryAllocateChecked(uint8(k), px, py, [2]uint32{3, 4}, om); ok {
				id, x, y, obj = got, px, py, o
				break
			}
		}
	}
	if obj == nil {
		t.Fatal("no industry allocated")
	}
	b := obj.Buildings[0]
	large := obj.BuildingSizeFlags&(1<<b) != 0
	tl := m.TileAt(int(x)>>5, int(y)>>5)
	nIndBefore := len(tl.Industries)
	if _, ok := m.placeIndustryBuilding(c, id, x, y, 0, b, 3, true, 0); !ok {
		t.Skipf("%s's first building can't stand at (%d, %d) (the query)", obj.Name, x>>5, y>>5)
	}
	if len(tl.Industries) != nIndBefore {
		t.Fatal("the query placed something")
	}
	m.industryByID(id).Tiles = nil // the query adds the tile, as upstream (createIndustry takes it back)
	cost, ok := m.placeIndustryBuilding(c, id, x, y, 0, b, 3, true, CmdApply)
	if !ok {
		t.Fatal("placing refused after the query allowed it")
	}
	ind := m.industryByID(id)
	if len(ind.Tiles) != 1 {
		t.Fatalf("industry tiles %v", ind.Tiles)
	}
	z := uint8((ind.Tiles[0][2] & 0x7FFF) / 4)
	for seq, off := range buildingTileOffsets(large) {
		ft := m.TileAt(int(x+off[0])>>5, int(y+off[1])>>5)
		found := false
		for _, e := range ft.Industries {
			if e.IndustryID == id && e.BaseZ == z && e.SequenceIndex == uint8(seq) && e.BuildingType == b && e.Colour == 3 && e.IsConstructed {
				found = true
			}
		}
		if !found || ft.BaseZ != z || ft.Slope&0x1F != 0 || len(ft.Trees) != 0 {
			t.Errorf("tile %d of the building: element %v, ground %d (want %d), slope %x, %d trees", seq, found, ft.BaseZ, z, ft.Slope, len(ft.Trees))
		}
	}
	t.Logf("%s: building type %d (2×2 %v) at tile (%d, %d), height %d, cost %d", obj.Name, b, large, x>>5, y>>5, z, cost)
	for ty := 0; ty < m.Height; ty++ { // a land industry on water
		for tx := 0; tx < m.Width; tx++ {
			if w := m.TileAt(tx, ty); w.WaterLevel != 0 {
				if _, ok := m.placeIndustryBuilding(c, id, int32(tx)*32, int32(ty)*32, 0, b, 3, true, 0); ok {
					t.Error("a land industry's building allowed on water")
				}
				return
			}
		}
	}
}

// TestCreateIndustry (CR part 2e): the createIndustry command at a searched
// place on the title map — the query leaves the industries as they were; the
// apply builds between the object's min and max buildings, each element on
// the map, the industry under construction, its production targets in the
// object's range, its centre on a tile corner; the same seed builds the same.
func TestCreateIndustry(t *testing.T) {
	m0, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	run := func(m *Map) (int32, uint8, bool, IndustryPlacement) {
		for k := 0; k < 16; k++ {
			if om.GetIndustryObjectByIndex(k) == nil || om.GetIndustryObjectByIndex(k).Flags&industryBuiltOnWater != 0 {
				continue
			}
			for try := 0; try < 10; try++ {
				x, y, ok := m.findRandomNewIndustryLocation(uint8(k), om)
				if !ok {
					continue
				}
				a := IndustryPlacement{Type: uint8(k), X: x, Y: y, Srand1: 0x1234}
				n := len(m.Industries)
				if _, _, ok := m.CreateIndustry(c, 0, a); !ok {
					continue
				}
				if len(m.Industries) != n {
					t.Fatal("the query left an industry behind")
				}
				cost, id, ok := m.CreateIndustry(c, CmdApply, a)
				if ok {
					return cost, id, true, a
				}
			}
		}
		return 0, 0, false, IndustryPlacement{}
	}
	snap := func() *Map { m, _ := loadDataMap(t, "title.dat"); return m }
	m := m0
	cost, id, ok, a := run(m)
	if !ok {
		t.Fatal("no industry could be created")
	}
	ind := m.industryByID(id)
	obj := om.GetIndustryObjectByIndex(int(a.Type))
	if n := len(ind.Tiles); n < int(obj.MinNumBuildings) || n > int(obj.MaxNumBuildings) {
		t.Errorf("%d buildings, want %d–%d", n, obj.MinNumBuildings, obj.MaxNumBuildings)
	}
	for _, tl := range ind.Tiles {
		found := false
		for _, e := range m.TileAt(int(tl[0])>>5, int(tl[1])>>5).Industries {
			if e.IndustryID == id && e.SequenceIndex == 0 {
				found = true
			}
		}
		if !found {
			t.Errorf("no element of the industry at its tile %v", tl)
		}
	}
	for i := 0; i < 2; i++ {
		lo, hi := obj.InitialProductionRate[i][0], obj.InitialProductionRate[i][1]
		if tg := ind.DailyProductionTarget[i]; tg < lo || tg > hi {
			t.Errorf("production target %d: %d outside %d–%d", i, tg, lo, hi)
		}
	}
	if ind.UnderConstruction != 0 || ind.X%32 != 0 || ind.Y%32 != 0 {
		t.Errorf("under construction %d, centre (%d, %d)", ind.UnderConstruction, ind.X, ind.Y)
	}
	t.Logf("%s: industry %d, %d buildings, centre tile (%d, %d), targets %v, cost %d", obj.Name, id, len(ind.Tiles), ind.X/32, ind.Y/32, ind.DailyProductionTarget, cost)
	m2 := snap()
	if _, id2, ok2, _ := run(m2); !ok2 || id2 != id || fmt.Sprint(m2.industryByID(id2).Tiles) != fmt.Sprint(ind.Tiles) {
		t.Error("the same seed built differently")
	}
}

// TestCreateWallAndClaim (CR part 2f): a wall on a flat, dry, empty tile of
// the title map — the query places nothing, the apply puts it at the ground's
// height on its edge; a second on the same edge is refused, another edge is
// fine —; claimSurfaceForIndustry makes a free tile an industry's field and
// refuses water.
func TestCreateWallAndClaim(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	wt := -1
	for i, w := range om.WallObjects {
		if w != nil && w.Flags&objects.WallFlagOnlyOnLevelLand == 0 {
			wt = i
			break
		}
	}
	if wt < 0 {
		t.Skip("no wall object")
	}
	tx, ty := -1, -1
	for y := 10; y < m.Height-10 && tx < 0; y++ {
		for x := 10; x < m.Width-10; x++ {
			tl := m.TileAt(x, y)
			if tl.Slope&0x1F == 0 && tl.WaterLevel == 0 && len(tl.Order) == 1 {
				tx, ty = x, y
				break
			}
		}
	}
	if tx < 0 {
		t.Fatal("no empty flat tile")
	}
	tl := m.TileAt(tx, ty)
	a := WallPlacement{X: int32(tx) * 32, Y: int32(ty) * 32, Rotation: 0, Type: uint8(wt)}
	if _, ok := m.CreateWall(c, 0, a); !ok || len(tl.Walls) != 0 {
		t.Fatalf("query: ok %v, %d walls", ok, len(tl.Walls))
	}
	if _, ok := m.CreateWall(c, CmdApply, a); !ok || len(tl.Walls) != 1 || tl.Walls[0].BaseZ != tl.BaseZ || tl.Walls[0].Rotation != 0 {
		t.Fatalf("apply: ok %v, walls %+v (ground %d)", ok, tl.Walls, tl.BaseZ)
	}
	if _, ok := m.CreateWall(c, 0, a); ok {
		t.Error("a second wall on the same edge allowed")
	}
	a.Rotation = 2
	if _, ok := m.CreateWall(c, 0, a); !ok {
		t.Error("a wall on the other edge refused")
	}
	if !m.claimSurfaceForIndustry(tx, ty, 5, 3, 2) || !tl.IsIndustrial || tl.Variation != 5 || tl.GrowthStage != 3 || tl.UpdateTimer != 2 {
		t.Errorf("claim: industrial %v id %d growth %d timer %d", tl.IsIndustrial, tl.Variation, tl.GrowthStage, tl.UpdateTimer)
	}
	if len(tl.Walls) != 0 {
		t.Errorf("the claimed field kept %d walls above it", len(tl.Walls))
	}
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.TileAt(x, y).WaterLevel != 0 {
				if m.claimSurfaceForIndustry(x, y, 5, 0, 0) {
					t.Error("water claimed")
				}
				return
			}
		}
	}
}

// TestIndustryGroundsAndMonthlyStartUp (CR part 2g): a farm built by the
// createIndustry command claims fields around its buildings and fences
// them; with every Sawmill of the title map removed (under its cap now),
// the month's tryCreateNewIndustriesMonthly founds a new one, under
// construction.
func TestIndustryGroundsAndMonthlyStartUp(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	farm, saw := -1, -1
	for k := 0; k < 16; k++ {
		o := om.GetIndustryObjectByIndex(k)
		if o == nil {
			continue
		}
		if farm < 0 && o.FarmTileGrowthStageNoProduction != 0xFF && o.BuildingWall != 0xFF {
			farm = k
		}
		if strings.Contains(o.Name, "Sawmill") {
			saw = k
		}
	}
	if farm < 0 || saw < 0 {
		t.Skip("no farm with fences, or no sawmill")
	}
	var id uint8
	built := false
	for try := 0; try < 40 && !built; try++ {
		x, y, ok := m.findRandomNewIndustryLocation(uint8(farm), om)
		if !ok {
			continue
		}
		_, id, built = m.CreateIndustry(c, CmdApply, IndustryPlacement{Type: uint8(farm), X: x, Y: y, Srand1: uint32(try)})
	}
	if !built {
		t.Fatal("no farm built")
	}
	fields, walls := 0, 0
	for ty := 0; ty < m.Height; ty++ {
		for tx := 0; tx < m.Width; tx++ {
			tl := m.TileAt(tx, ty)
			if tl.IsIndustrial && tl.Variation == id {
				fields++
				walls += len(tl.Walls)
			}
		}
	}
	ind := m.industryByID(id)
	t.Logf("%s: industry %d, %d buildings, %d fields, %d walls on them", om.GetIndustryObjectByIndex(farm).Name, id, len(ind.Tiles), fields, walls)
	if fields < 9 || walls == 0 {
		t.Errorf("the farm's grounds: %d fields, %d walls", fields, walls)
	}
	var sawmills []uint8
	for _, in := range m.Industries {
		if int(in.ObjectID) == saw {
			sawmills = append(sawmills, in.ID)
		}
	}
	for _, sid := range sawmills {
		m.RemoveIndustry(c, CmdApply, sid)
	}
	m.tryCreateNewIndustriesMonthly(om)
	n := 0
	for _, in := range m.Industries {
		if int(in.ObjectID) == saw {
			n++
			if in.UnderConstruction != 0 {
				t.Errorf("the new sawmill isn't under construction (%d)", in.UnderConstruction)
			}
		}
	}
	t.Logf("%d sawmills removed; the month founded %d", len(sawmills), n)
	if n == 0 {
		t.Error("no new sawmill founded under the cap")
	}
}

// TestFarmProduction (CR part 3): the save holds each industry's farm counts
// halfway through its tile loop's round. For every industry whose loop has
// already passed all its fields this round, a fresh count of them
// (isFarmTileProducing over the whole map) gives back exactly the saved
// fields and idle fields; then a round run from the saved place ends in
// calculateFarmProduction, whose rate follows from those counts. (The saved
// rates themselves are the previous round's: the save is in November, snow
// and trees on the fields have changed since.)
func TestFarmProduction(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	checked := 0
	for i := range m.Industries {
		ind := &m.Industries[i]
		obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
		if obj == nil || ind.UnderConstruction != 0xFF {
			continue
		}
		loopAt := int(ind.TileLoop[1])/32*384 + int(ind.TileLoop[0])/32
		last := -1
		for ty := 0; ty < m.Height; ty++ {
			for tx := 0; tx < m.Width; tx++ {
				if tl := m.TileAt(tx, ty); tl.IsIndustrial && tl.Variation == ind.ID {
					last = ty*384 + tx
				}
			}
		}
		if last < 0 || last >= loopAt {
			continue
		}
		fresh := *ind
		fresh.NumFarmTiles, fresh.NumIdleFarmTiles = 0, 0
		for ty := 0; ty < m.Height; ty++ {
			for tx := 0; tx < m.Width; tx++ {
				m.isFarmTileProducing(&fresh, obj, [2]int16{int16(tx * 32), int16(ty * 32)})
			}
		}
		checked++
		snowy := int16(0) // the original game counted snowy fields idle even where the object ignores snow (OpenLoco honours it)
		if obj.Flags&industryFarmIgnoresSnow != 0 {
			for ty := 0; ty < m.Height; ty++ {
				for tx := 0; tx < m.Width; tx++ {
					if tl := m.TileAt(tx, ty); tl.IsIndustrial && tl.Variation == ind.ID && tl.Slope>>5 != 0 && len(tl.Trees) == 0 &&
						(tl.GrowthStage == 0 || tl.GrowthStage != obj.FarmTileGrowthStageNoProduction) {
						snowy++
					}
				}
			}
		}
		if fresh.NumFarmTiles != ind.NumFarmTiles || fresh.NumIdleFarmTiles+snowy != ind.NumIdleFarmTiles {
			t.Errorf("%s %d: counted %d fields / %d idle, saved %d / %d", obj.Name, ind.ID, fresh.NumFarmTiles, fresh.NumIdleFarmTiles, ind.NumFarmTiles, ind.NumIdleFarmTiles)
		}
		n, idle := ind.NumFarmTiles, ind.NumIdleFarmTiles
		for step := 0; step < 384*384/100+2; step++ { // the rest of the round: no more fields, then the rate
			m.industryUpdate(ind, om)
			if ind.NumFarmTiles == 0 {
				break
			}
		}
		adj := max(0, idle-n/16)
		size := min((n-adj)/25, 255)
		want := uint8(255)
		if size < int16(obj.FarmNumFields) {
			want = uint8(int32(size) * 256 / int32(obj.FarmNumFields))
		}
		if ind.ProductionRate != want {
			t.Errorf("%s %d: rate %d from %d / %d, want %d", obj.Name, ind.ID, ind.ProductionRate, n, idle, want)
		}
	}
	t.Logf("%d industries past all their fields: counts as saved, rate from them", checked)
	if checked == 0 {
		t.Fatal("none to check")
	}
}

// TestIndustryConstruction (CR part 4): a new industry founded under
// construction builds up in the tile loop (IndustryElement::update) — each
// building a section at a time —, until every building is constructed, its
// clear height its parts', and the industry is done (under_construction
// 0xFF); the objects' unk38 lists read where the reader used to skip them.
func TestIndustryConstruction(t *testing.T) {
	m0, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m0}
	m := &s.Map
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	unk := 0
	for k := 0; k < 16; k++ {
		if o := om.GetIndustryObjectByIndex(k); o != nil {
			unk += len(o.Unk38)
		}
	}
	var id uint8
	ok := false
	for k := 0; k < 16 && !ok; k++ {
		o := om.GetIndustryObjectByIndex(k)
		if o == nil || o.Flags&industryBuiltOnWater != 0 {
			continue
		}
		for try := 0; try < 10 && !ok; try++ {
			x, y, found := m.findRandomNewIndustryLocation(uint8(k), om)
			if found {
				_, id, ok = m.CreateIndustry(c, CmdApply, IndustryPlacement{Type: uint8(k), X: x, Y: y, Srand1: uint32(try + 1)})
			}
		}
	}
	if !ok {
		t.Fatal("no industry founded")
	}
	ind := m.industryByID(id)
	if ind.UnderConstruction != 0 {
		t.Fatalf("founded at %d, want 0 (under construction)", ind.UnderConstruction)
	}
	passes := 0
	for ; passes < 256*80 && m.industryByID(id).UnderConstruction != 0xFF; passes++ {
		s.UpdateTiles(func(int, int) bool { return false }, om)
	}
	ind = m.industryByID(id)
	if ind.UnderConstruction != 0xFF {
		t.Fatalf("still under construction (%d) after %d passes", ind.UnderConstruction, passes)
	}
	obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
	for _, tl := range ind.Tiles {
		for _, e := range m.TileAt(int(tl[0])>>5, int(tl[1])>>5).Industries {
			if e.IndustryID != id {
				continue
			}
			if !e.IsConstructed || e.ClearZ <= e.BaseZ {
				t.Errorf("building at %v: constructed %v, %d–%d", tl, e.IsConstructed, e.BaseZ, e.ClearZ)
			}
		}
	}
	t.Logf("%s (%d buildings) built in %d tile-loop passes; %d unk38 entries over the industry objects", obj.Name, len(ind.Tiles), passes, unk)
	if unk == 0 {
		t.Error("no unk38 entries read")
	}
}

// firstRailTrain is the first rail train (vehicle type 0) of at least minCars
// cars in Trains' order — VehicleList's, newest first, where a truck may lead.
func firstRailTrain(t *testing.T, m *Map, trains []Train, minCars int) Train {
	t.Helper()
	for _, tr := range trains {
		if m.Entities[tr.Head].VehicleType == 0 && len(tr.Cars) >= minCars {
			return tr
		}
	}
	t.Skipf("no rail train of %d cars", minCars)
	return Train{}
}

// The saved Vehicle1s' lastIncome (0x53) reads as upstream's IncomeStats:
// days not after today, cargo lines of loaded cargo, and some vehicle with one.
func TestSavedLastIncome(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	withIncome := 0
	for _, tr := range trains {
		li := m.Entities[tr.Veh1].LastIncome
		if li.Day == -1 {
			continue
		}
		if li.Day > int32(m.CurrentDay) || li.Day < 0 {
			t.Errorf("vehicle %d: last income on day %d, today %d", m.Entities[tr.Head].ID, li.Day, m.CurrentDay)
		}
		for i := 0; i < 4; i++ {
			if c := li.CargoTypes[i]; c != 0xFF {
				if int(c) >= len(om.CargoObjects) || om.CargoObjects[c] == nil {
					t.Errorf("vehicle %d: last income of cargo %d, not loaded", m.Entities[tr.Head].ID, c)
				}
				withIncome++
			}
		}
	}
	t.Logf("%d income lines over %d vehicles", withIncome, len(trains))
	if withIncome == 0 {
		t.Error("no saved vehicle has a last income")
	}
}

// VehicleRefit: a bought refittable vehicle refitted to another refittable
// cargo — its first body carries that cargo, empty, the capacity rescaled by
// unit size, accepting that cargo's category; another company can't.
func TestVehicleRefit(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	saved := firstRailTrain(t, m, func() []Train { tr, _ := m.Trains(); return tr }(), 1)
	company := m.Entities[saved.Head].Owner
	c := &Cmd{Company: company, Objects: om}
	for id, v := range om.Vehicles {
		if v == nil || v.Flags&vehicleFlagRefittable == 0 {
			continue
		}
		var h uint16
		if _, ok := m.CreateVehicle(c, CmdApply, uint16(id), true, &h); !ok {
			continue
		}
		tr, _ := m.TrainOf(h)
		body := &m.Entities[tr.Cars[0][0].Body]
		cargo := om.CargoFor(v)
		primary := bitScanForward32(uint32(cargo.CompatibleCargoCategories[0]))
		for nc, co := range om.CargoObjects {
			if co == nil || co.Flags&(1<<1) == 0 || uint8(nc) == body.Cargo.Type { // CargoObjectFlags::refit
				continue
			}
			if _, ok := m.VehicleRefit(&Cmd{Company: company ^ 1, Objects: om}, CmdApply, h, uint8(nc)); ok {
				t.Error("another company refitted the vehicle")
			}
			if _, ok := m.VehicleRefit(c, CmdApply, h, uint8(nc)); !ok {
				t.Fatalf("%s: refit to %s refused", v.DisplayName, co.Name)
			}
			body = &m.Entities[tr.Cars[0][0].Body]
			want := uint32(om.CargoObjects[primary].UnitSize) * uint32(cargo.MaxCargo[0]) / uint32(co.UnitSize)
			if body.Cargo.Type != uint8(nc) || body.Cargo.Qty != 0 || uint32(body.Cargo.MaxQty) != min(want, 0xFF) || body.Cargo.AcceptedTypes&(1<<nc) == 0 {
				t.Errorf("%s refitted to %s: %+v (want max %d)", v.DisplayName, co.Name, body.Cargo, want)
			}
			t.Logf("%s refitted to %s: max %d, accepts %#x", v.DisplayName, co.Name, body.Cargo.MaxQty, body.Cargo.AcceptedTypes)
			return
		}
	}
	t.Skip("no refittable vehicle and cargo")
}
