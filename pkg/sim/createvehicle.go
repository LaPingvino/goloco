package sim

import (
	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The CreateVehicle game command (GameCommands/Vehicles/CreateVehicle.cpp),
// item 5P.a — in parts. This part: a new vehicle's base, the head, Vehicle1,
// Vehicle2 and tail entities (createBaseVehicle, 0x004AE318) in fresh entity
// slots from the free list (entitylists.go), a routing handle, an empty order
// table. The cars (bogies, bodies), the cost and the command's checks follow.

const (
	maxOrders                = 256000 // Limits::kMaxOrders
	numVehicleComponentsBase = 4      // kNumVehicleComponentsInBase: head, veh1, veh2, tail
	vehicleStatusUnk0        = 0      // Status::unk_0
	orderEndSize             = 1      // kOrderSizes[OrderType::End]
)

// newVehicleEntity is createVehicleBaseEntity: a slot from the free list made
// a vehicle of this sub-type, zeroed as createEntity does (position null).
func (m *Map) newVehicleEntity(entityType uint8) (*scenario.VehicleEntity, bool) {
	id, ok := m.EntityLists.createEntityVehicle()
	if !ok {
		return nil, false
	}
	m.Entities = append(m.Entities, scenario.VehicleEntity{ID: id, EntityType: entityType,
		PosX: coords.WorldUnits(locationNull), NextCarID: entityIDNull, HeadID: entityIDNull, Name: stringIDEmptyPop})
	return &m.Entities[len(m.Entities)-1], true
}

const (
	stringIDEmptyPop = 0x08 // StringIds::empty_pop — what createEntity names a new entity
)

// createUniqueTypeNumber is CreateVehicle.cpp 0x004B64F9: the lowest ordinal
// number (from 1) not taken by the company's vehicles of this type.
func (m *Map) createUniqueTypeNumber(company, vehicleType uint8) int16 {
	taken := map[int16]bool{}
	for i := range m.Entities {
		e := &m.Entities[i]
		if e.EntityType == scenario.VehicleTypeHead && e.Owner == company && e.VehicleType == vehicleType && e.OrdinalNumber != 0 {
			taken[e.OrdinalNumber] = true
		}
	}
	n := int16(1)
	for taken[n] {
		n++
	}
	return n
}

// allocateOrders is OrderManager::allocateOrders: the head's order table, at
// the end of the orders in use, holding only the end order.
func (m *Map) allocateOrders(head *scenario.VehicleEntity) {
	if m.Orders == nil {
		m.Orders = make([]byte, maxOrders)
	}
	head.OrderTableOffset = m.OrderTableLength
	m.Orders[m.OrderTableLength] = OrderEnd
	m.OrderTableLength += orderEndSize
	head.CurrentOrder = 0
	head.SizeOfOrderTable = orderEndSize
}

// createBaseVehicle is CreateVehicle.cpp 0x004AE318 (with createHead
// 0x004AE34B, createVehicle1 0x004AE40E, createVehicle2 0x004AE4A0,
// createVehicleTail 0x004AE54E): the head's id; false when there aren't
// enough free entities, no room for orders, or no routing handle.
// PARTIAL: aiIsBelowVehicleLimit for an AI company (companies' transport type
// counts are not kept yet — a player is always below it, as upstream);
// head->updateTrainProperties runs when the cars are added (5P.a part 3); the
// sound and AI fields goloco does not keep.
func (m *Map) createBaseVehicle(company, mode, vehicleType, trackType uint8) (uint16, bool) {
	if m.EntityLists.numFreeEntities() <= numVehicleComponentsBase { // checkNumFreeEntities
		return 0, false
	}
	if m.OrderTableLength >= maxOrders { // no space for more vehicle orders
		return 0, false
	}
	routing, ok := m.AllocateRoutingHandle()
	if !ok { // too many vehicles
		return 0, false
	}
	common := func(e *scenario.VehicleEntity, head uint16) {
		e.Owner, e.HeadID, e.TrackType, e.Mode = company, head, trackType, mode
		e.TileX, e.TileY, e.TileBaseZ = -1, 0, 0
		e.RemainingDistance, e.SubPosition, e.TrackAndDirection = 0, 0, 0
		e.RoutingHandle = uint16(routing)
		e.Flags38, e.Var3C = 0, 0
	}
	head, _ := m.newVehicleEntity(scenario.VehicleTypeHead)
	m.EntityLists.moveEntityToList(head.ID, EntityListVehicleHead)
	headID := head.ID
	common(head, headID)
	head.VehicleFlags = VehicleFlagCommandStop
	head.VehicleType = vehicleType
	head.Name = uint16(vehicleType) + 4
	head.OrdinalNumber = 0
	head.OrdinalNumber = m.createUniqueTypeNumber(company, vehicleType)
	head.HeadVar52, head.Var5C = 0, 0
	head.Status, head.StationID, head.BreakdownFlags = vehicleStatusUnk0, 0xFFFF, 0
	head.TotalRefundCost, head.LastAverageSpeed = 0, 0
	m.allocateOrders(head)
	last := headID
	link := func(e *scenario.VehicleEntity) { // lastVeh->setNextCar
		for i := range m.Entities {
			if m.Entities[i].ID == last {
				m.Entities[i].NextCarID = e.ID
				break
			}
		}
		last = e.ID
	}
	v1, _ := m.newVehicleEntity(scenario.VehicleTypeVeh1)
	common(v1, headID)
	v1.TargetSpeed, v1.TimeAtSignal, v1.Var48 = 0, 0, 0
	link(v1)
	v2, _ := m.newVehicleEntity(scenario.VehicleTypeVeh2)
	common(v2, headID)
	v2.CurrentSpeed, v2.MotorState, v2.BrakeLightTimeout = 0, 0, 0
	v2.CurMonthRevenue, v2.Profit, v2.Veh2Reliability, v2.Var73 = 0, [4]int32{}, 0, 0
	link(v2)
	tail, _ := m.newVehicleEntity(scenario.VehicleTypeTail)
	common(tail, headID)
	link(tail)
	tail.NextCarID = entityIDNull
	return headID, true
}
