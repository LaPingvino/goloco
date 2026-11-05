package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The createVehicle game command (CreateVehicle.cpp 0x004AE5E4): buying a new
// vehicle (createNewVehicle 0x004AE74E). Adding a car to an existing vehicle
// (addCarToVehicle 0x004AE5FF) is the next step of item 5P.a.

const (
	kMaxAiVehicles               = 500 // CreateVehicle.cpp kMaxAiVehicles
	numVehicleComponentsInBaseCV = 4   // kNumVehicleComponentsInBase (head, veh1, veh2, tail)
)

// CreateVehicle is the createVehicle game command with no head given
// (headId null): the vehicle object's inflation-adjusted cost (divisor 6) and
// true, or false when refused — locked (unless lockedOK, Config::
// buildLockedVehicles), too few free entities, no vehicle slot, or the car
// could not be made. Applied (CmdApply), the new vehicle's head id is
// returned in *head (lastCreatedVehicleId; entityIDNull otherwise).
// PARTIAL: the expenditure type / payment (doCommand's note), the error
// texts, MessageManager::removeAllSubjectRefs in the cleanup (goloco keeps no
// messages naming vehicles), the vehicle list window's invalidation.
func (m *Map) CreateVehicle(c *Cmd, flags uint8, objID uint16, lockedOK bool, head *uint16) (int32, bool) {
	if head != nil {
		*head = entityIDNull
	}
	if !lockedOK && !m.isVehicleIndexUnlocked(c.Company, objID) {
		return 0, false // vehicle_is_locked
	}
	om := c.Objects
	if int(objID) >= len(om.Vehicles) || om.Vehicles[objID] == nil {
		return 0, false
	}
	obj := om.Vehicles[objID]
	// createNewVehicle
	if m.EntityLists.numFreeEntities() <= maxNumVehicleComponentsInCar+numVehicleComponentsInBaseCV {
		return 0, false // too_many_objects_in_game
	}
	if !m.isEmptyVehicleSlotAvailable(c) {
		return 0, false
	}
	if flags&CmdApply != 0 {
		h, ok := m.createBaseVehicle(c.Company, uint8(obj.Mode), uint8(obj.Type), om.TrackTypeFor(obj))
		if !ok {
			return 0, false
		}
		if head != nil {
			*head = h
		}
		if !m.createCar(c.Company, h, objID, om) {
			m.deleteBaseVehicle(h)
			if head != nil {
				*head = entityIDNull
			}
			return 0, false
		}
		// updateWholeVehicle 0x004AE6DE (no placement for a new vehicle)
		m.autoLayoutTrain(h, om)
		m.RecalculateTransportCounts(c.Company)
	}
	return m.InflationAdjustedCost(obj.CostFactor, obj.CostIndex, 6), true
}

// isVehicleIndexUnlocked is Company::isVehicleIndexUnlocked.
func (m *Map) isVehicleIndexUnlocked(company uint8, objID uint16) bool {
	if int(company) >= len(m.UnlockedVehicles) || objID >= 7*32 {
		return false
	}
	return m.UnlockedVehicles[company][objID/32]&(1<<(objID%32)) != 0
}

// isEmptyVehicleSlotAvailable is CreateVehicle.cpp 0x004B1E44: an AI company
// below kMaxAiVehicles in all (aiIsBelowVehicleLimit), then a free routing
// slot.
func (m *Map) isEmptyVehicleSlotAvailable(c *Cmd) bool {
	if !c.isPlayer(c.Company) {
		total := 0
		for id := range m.Companies {
			if c.isPlayer(uint8(id)) {
				continue
			}
			for _, n := range m.Companies[id].TransportTypeCount {
				total += int(n)
			}
		}
		if total > kMaxAiVehicles {
			return false // too_many_vehicles
		}
	}
	return m.IsEmptyRoutingSlotAvailable()
}

// RecalculateTransportCounts is Company::recalculateTransportCounts: the
// company's vehicles counted by vehicle type.
func (m *Map) RecalculateTransportCounts(company uint8) {
	if int(company) >= len(m.Companies) {
		return
	}
	c := &m.Companies[company]
	c.TransportTypeCount = [6]uint16{}
	for i := range m.Entities {
		e := &m.Entities[i]
		if e.EntityType == scenario.VehicleTypeHead && e.Owner == company && int(e.VehicleType) < len(c.TransportTypeCount) {
			c.TransportTypeCount[e.VehicleType]++
		}
	}
}

// deleteBaseVehicle is createNewVehicle's cleanup when the car can't be made:
// the routing handle freed (RoutingManager::freeRoutingHandle), the orders
// freed (OrderManager::freeOrders), and the head, veh1, veh2 and tail freed.
func (m *Map) deleteBaseVehicle(headID uint16) {
	h := m.entity(headID)
	if h == nil {
		return
	}
	m.freeRoutingHandle(RoutingHandle(h.RoutingHandle))
	m.freeOrders(headID)
	ids := []uint16{}
	for id := headID; id != entityIDNull && len(ids) < 4; {
		e := m.entity(id)
		if e == nil {
			break
		}
		ids = append(ids, id)
		id = e.NextCarID
	}
	for _, id := range ids {
		m.removeEntity(id)
	}
}

// freeRoutingHandle is RoutingManager::freeRoutingHandle: the vehicle's
// routings all null.
func (m *Map) freeRoutingHandle(h RoutingHandle) {
	r := m.vehicleRoutings(uint16(h) / scenario.RoutingsPerVehicle)
	for i := range r {
		r[i] = RoutingNull
	}
}

// freeOrders is OrderManager::freeOrders: every head's table at or after this
// one's moves back by its size (reoffsetVehicleOrderTables, the freed head's
// own included, as upstream), the table shifted out, the length shortened.
// (The order display frames of reoffsetVehicleOrderTables goloco doesn't keep.)
func (m *Map) freeOrders(headID uint16) {
	h := m.entity(headID)
	if h == nil {
		return
	}
	offset, size := h.OrderTableOffset, h.SizeOfOrderTable
	for i := range m.Entities {
		e := &m.Entities[i]
		if e.EntityType == scenario.VehicleTypeHead && uint32(e.OrderTableOffset) >= uint32(offset) {
			e.OrderTableOffset = uint32(int64(e.OrderTableOffset) - int64(size))
		}
	}
	end := int(m.OrderTableLength)
	o, s := int(offset), int(size)
	if o+s <= end && end <= len(m.Orders) {
		tmp := append([]byte(nil), m.Orders[o:o+s]...)
		copy(m.Orders[o:], m.Orders[o+s:end])
		copy(m.Orders[end-s:end], tmp) // std::rotate
	}
	m.OrderTableLength -= uint32(size)
}

// removeEntity frees an entity's slot (EntityManager::freeEntity) and drops it
// from goloco's entity slice.
func (m *Map) removeEntity(id uint16) {
	for i := range m.Entities {
		if m.Entities[i].ID == id {
			m.Entities = append(m.Entities[:i], m.Entities[i+1:]...)
			break
		}
	}
	m.EntityLists.freeEntity(id)
}

const (
	statusBrokenDown       uint8 = 7       // VehicleHead Status::brokenDown
	kMaxRoadVehicleLength        = 176     // Vehicle.h
	vehObjFlagCannotCouple       = 1 << 11 // VehicleObjectFlags::cannotCoupleToSelf
	kMaxBodySprites              = 4       // VehicleObject::kMaxBodySprites
)

// AddCarToVehicle is the createVehicle game command with a head
// (addCarToVehicle 0x004AE5FF): the car's inflation-adjusted cost (divisor 6)
// and true, or false when refused — locked (unless lockedOK), another
// company's vehicle (checkCompanyCompatibility), the vehicle can't be modified
// (crashed, stuck, broken down, moving), the car doesn't fit (type, track,
// coupling, a road vehicle's length), or too few free entities.
// PARTIAL: a vehicle that is placed (tileX != -1) is refused: lifting it off
// the track and putting it back (liftUpVehicle / placeDownVehicle, item
// 5P.b) are not ported yet; and as CreateVehicle's note.
func (m *Map) AddCarToVehicle(c *Cmd, flags uint8, objID, headID uint16, lockedOK bool) (int32, bool) {
	if !lockedOK && !m.isVehicleIndexUnlocked(c.Company, objID) {
		return 0, false // vehicle_is_locked
	}
	om := c.Objects
	if int(objID) >= len(om.Vehicles) || om.Vehicles[objID] == nil {
		return 0, false
	}
	h := m.entity(headID)
	if h == nil || h.EntityType != scenario.VehicleTypeHead {
		return 0, false
	}
	if h.Owner != c.Company { // checkCompanyCompatibility
		return 0, false
	}
	tr, err := m.TrainOf(headID)
	if err != nil || !m.canBeModified(tr) || !m.isVehicleTypeCompatible(tr, objID, om) {
		return 0, false
	}
	if m.EntityLists.numFreeEntities() <= maxNumVehicleComponentsInCar {
		return 0, false // too_many_objects_in_game
	}
	if flags&CmdApply != 0 {
		if h.TileX != -1 {
			return 0, false // PARTIAL: liftUpVehicle / placeDownVehicle (5P.b)
		}
		if !m.createCar(c.Company, headID, objID, om) {
			return 0, false // nothing lifted, nothing to put back
		}
		m.autoLayoutTrain(headID, om) // updateWholeVehicle
		m.RecalculateTransportCounts(c.Company)
	}
	obj := om.Vehicles[objID]
	return m.InflationAdjustedCost(obj.CostFactor, obj.CostIndex, 6), true
}

// canBeModified is VehicleHead::canBeModified (VehicleHead.cpp 0x004B6D1F).
func (m *Map) canBeModified(tr Train) bool {
	h := &m.Entities[tr.Head]
	switch h.Status {
	case StatusCrashed, StatusStuck, statusBrokenDown:
		return false
	}
	v2 := &m.Entities[tr.Veh2]
	if h.VehicleType == uint8(objects.VehicleTypeAircraft) || h.VehicleType == uint8(objects.VehicleTypeShip) {
		if v2.Var73&1 != 0 { // isBrokenDown
			return false
		}
		if h.TileX == -1 {
			return true
		}
		if h.Status != StatusLoading && h.Status != StatusStopped {
			return false // vehicle_must_be_stopped
		}
		return v2.CurrentSpeed == 0
	}
	if h.TileX == -1 || v2.CurrentSpeed == 0 {
		return true
	}
	return m.Entities[tr.Veh1].Var3C <= 0x3689
}

// canVehiclesCouple is VehicleHead.cpp's canVehiclesCouple.
func canVehiclesCouple(newID, srcID uint16, om *objects.ObjectManager) bool {
	n, s := om.Vehicles[newID], om.Vehicles[srcID]
	if n == nil || s == nil {
		return false
	}
	if n.Flags&vehObjFlagCannotCouple != 0 && s.Flags&vehObjFlagCannotCouple != 0 {
		return false
	}
	if newID == srcID {
		return true
	}
	nc, sc := om.CompatibleVehiclesFor(n), om.CompatibleVehiclesFor(s)
	for i := 0; i < int(n.NumCompatVehicles) && i < len(nc); i++ {
		if nc[i] == srcID {
			return true
		}
	}
	for i := 0; i < int(s.NumCompatVehicles) && i < len(sc); i++ {
		if sc[i] == newID {
			return true
		}
	}
	return n.NumCompatVehicles == 0 && s.NumCompatVehicles == 0
}

// isVehicleTypeCompatible is VehicleHead::isVehicleTypeCompatible (0x004B8FA2).
func (m *Map) isVehicleTypeCompatible(tr Train, objID uint16, om *objects.ObjectManager) bool {
	h := &m.Entities[tr.Head]
	obj := om.Vehicles[objID]
	if obj.Mode == objects.TransportModeAir || obj.Mode == objects.TransportModeWater {
		if len(tr.Cars) != 0 {
			return false
		}
	} else if om.TrackTypeFor(obj) != h.TrackType {
		return false
	}
	if uint8(obj.Mode) != h.Mode || uint8(obj.Type) != h.VehicleType {
		return false
	}
	for _, car := range tr.Cars {
		if !canVehiclesCouple(objID, m.Entities[car[0].Front].ObjectID, om) {
			return false
		}
	}
	if obj.Mode != objects.TransportModeRoad || h.TrackType != 0xFF {
		return true
	}
	total := 0 // getVehicleTotalLength
	for _, car := range tr.Cars {
		total += vehicleLength(om.Vehicles[m.Entities[car[0].Body].ObjectID])
	}
	return total+vehicleLength(obj) <= kMaxRoadVehicleLength // vehicle_too_long
}

// vehicleLength is VehicleObject::getLength.
func vehicleLength(v *objects.VehicleObject) int {
	if v == nil {
		return 0
	}
	n := 0
	for i := 0; i < int(v.NumCarComponents) && i < len(v.CarComponents); i++ {
		b := v.CarComponents[i].BodySpriteIdx
		if b == 0xFF {
			continue
		}
		n += int(v.BodySprites[b&(kMaxBodySprites-1)].HalfLength) * 2
	}
	return n
}
