package sim

import (
	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/objects"
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
// (aiIsBelowVehicleLimit is the command's check before it: CreateVehicle's
// isEmptyVehicleSlotAvailable, as upstream.)
// PARTIAL: head->updateTrainProperties runs when the cars are added (5P.a
// part 3); the sound and AI fields goloco does not keep.
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
	v1.LastIncome.Day = -1 // no income yet: Finances shows "Last income: N/A"
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

const (
	maxNumVehicleComponentsInCar = 3 * 4  // kNumVehicleComponentsInCarComponent × kMaxNumCarComponentsInCar
	spriteIndexNull              = 0xFF   // SpriteIndex::null
	spriteIndexIsReversed        = 0x80   // SpriteIndex::isReversed
	f38Unk0                      = 1 << 0 // Flags38::unk_0
	f38IsReversed                = 1 << 1 // Flags38::isReversed
	f38JacobsBogieAvailable      = 1 << 3 // Flags38::jacobsBogieAvailable
	vehObjFlagJacobsBogieFront   = 1 << 2 // VehicleObjectFlags::jacobsBogieFront
	vehObjFlagJacobsBogieRear    = 1 << 3 // VehicleObjectFlags::jacobsBogieRear
)

// entity is the entity with this id (nil if none); appending entities moves
// the slice, so pointers are looked up again after each new one.
func (m *Map) entity(id uint16) *scenario.VehicleEntity {
	for i := range m.Entities {
		if m.Entities[i].ID == id {
			return &m.Entities[i]
		}
	}
	return nil
}

// setNextCar is lastVeh->setNextCar.
func (m *Map) setNextCar(last, next uint16) {
	if e := m.entity(last); e != nil {
		e.NextCarID = next
	}
}

// bitScanForward is Numerics::bitScanForward: the lowest set bit, -1 for none.
func bitScanForward(v uint32) int {
	for i := range 32 {
		if v&(1<<i) != 0 {
			return i
		}
	}
	return -1
}

// createBogie is CreateVehicle.cpp's createBogie (0x004AE8F1 / 0x004AEA9E's
// common part): a bogie entity after last, its object, gPrng1's var_44, the
// day it was made, no cargo yet, the colours.
func (m *Map) createBogie(company uint8, head, objID uint16, bodyNumber uint8, last uint16, scheme [2]uint8) uint16 {
	b, _ := m.newVehicleEntity(scenario.VehicleTypeBogie)
	lv := m.entity(last)
	b.Owner, b.HeadID, b.BodyIndex = company, head, bodyNumber
	b.TrackType, b.Mode = lv.TrackType, lv.Mode
	b.TileX, b.TileY, b.TileBaseZ = -1, 0, 0
	b.SubPosition, b.TrackAndDirection, b.RoutingHandle = 0, 0, lv.RoutingHandle
	b.ObjectID = objID
	b.Var44 = uint16(m.Rng1.RandNext())
	b.CreationDay = m.CurrentDay
	b.Animation = 0
	b.Cargo = scenario.VehicleCargo{AcceptedTypes: 0, Type: 0xFF, Qty: 0}
	b.WheelSlipping, b.BreakdownFlags = 0, 0
	b.SpriteWidth, b.SpriteHeightNegative, b.SpriteHeightPositive = 1, 1, 1
	b.ColourPrimary, b.ColourSecondary = scheme[0], scheme[1]
	id := b.ID
	m.setNextCar(last, id)
	return id
}

// sub4BA873 is VehicleBogie.cpp's sub_4BA873: the days to the next breakdown
// from the reliability (gPrng1), 0xFFFF for a vehicle that never breaks down.
func (m *Map) sub4BA873(b *scenario.VehicleEntity) {
	b.TimeoutToBreakdown = 0xFFFF
	if b.Reliability == 0 {
		return
	}
	f := int32(b.Reliability / 256)
	f *= f
	f /= 16
	randVal := int32(m.Rng1.RandNextBound(65535)) * (f / 2) / 65536
	f -= f / 4
	f += randVal
	b.TimeoutToBreakdown = uint16(max(4, f))
}

// createFirstBogie is 0x4AE8F1: the car component's front bogie — its
// reliability (less in the first years after the design year; vanilla's bug
// of always two eighths kept), the breakdown timeout, the refund cost (7/8 of
// the price), the secondary cargo on the first component, the sprite.
func (m *Map) createFirstBogie(company uint8, head, objID uint16, obj *objects.VehicleObject, cargo objects.VehicleCargoInfo, bodyNumber uint8, last uint16, scheme [2]uint8, om *objects.ObjectManager) uint16 {
	id := m.createBogie(company, head, objID, bodyNumber, last, scheme)
	b := m.entity(id)
	b.Flags38 = 0
	rel := int32(obj.Reliability) * 256
	if int(obj.Designed)+2 > int(m.CurrentYear) {
		rel -= rel / 8
		rel -= rel / 8
	}
	if rel != 0 {
		rel += 255
	}
	b.Reliability = uint16(rel)
	m.sub4BA873(b)
	cost := m.InflationAdjustedCost(obj.CostFactor, obj.CostIndex, 6)
	b.RefundCost = uint32(cost - cost/8)
	if bodyNumber == 0 && cargo.NumSimultaneousCargoTypes > 1 {
		b.Cargo.MaxQty = cargo.MaxCargo[1]
		b.Cargo.AcceptedTypes = cargo.CompatibleCargoCategories[1]
		if ct := bitScanForward(b.Cargo.AcceptedTypes); ct != -1 {
			b.Cargo.Type = uint8(ct)
		}
	}
	b.ObjectSpriteType = obj.CarComponents[bodyNumber].FrontBogieSpriteIdx
	if st := b.ObjectSpriteType; st != spriteIndexNull && int(st) < len(obj.BogieSprites) {
		s := obj.BogieSprites[st]
		b.SpriteWidth, b.SpriteHeightNegative, b.SpriteHeightPositive = s.Width, s.HeightNegative, s.HeightPositive
	}
	return id
}

// createSecondBogie is 0x004AEA9E's bogie: the back bogie, reversed.
func (m *Map) createSecondBogie(company uint8, head, objID uint16, obj *objects.VehicleObject, bodyNumber uint8, last uint16, scheme [2]uint8) uint16 {
	id := m.createBogie(company, head, objID, bodyNumber, last, scheme)
	b := m.entity(id)
	b.Flags38 = f38IsReversed
	b.ObjectSpriteType = obj.CarComponents[bodyNumber].BackBogieSpriteIdx
	if st := b.ObjectSpriteType; st != spriteIndexNull && int(st) < len(obj.BogieSprites) {
		s := obj.BogieSprites[st]
		b.SpriteWidth, b.SpriteHeightNegative, b.SpriteHeightPositive = s.Width, s.HeightNegative, s.HeightPositive
	}
	return id
}

// createBody is 0x004AEA9E: the car component's body — body_start for the
// first, body_continued after; the primary cargo on the first; its sprite
// (reversed by the sprite index's high bit); the Jacobs bogie flags.
func (m *Map) createBody(company uint8, head, objID uint16, obj *objects.VehicleObject, cargo objects.VehicleCargoInfo, bodyNumber uint8, last uint16, scheme [2]uint8) uint16 {
	typ := scenario.VehicleTypeBodyContinued
	if bodyNumber == 0 {
		typ = scenario.VehicleTypeBodyStart
	}
	b, _ := m.newVehicleEntity(typ)
	lv := m.entity(last)
	b.Owner, b.HeadID, b.BodyIndex = company, head, bodyNumber
	b.TrackType, b.Mode = lv.TrackType, lv.Mode
	b.TileX, b.TileY, b.TileBaseZ = -1, 0, 0
	b.SubPosition, b.TrackAndDirection, b.RoutingHandle = 0, 0, lv.RoutingHandle
	b.Flags38 = f38Unk0
	b.ObjectID = objID
	b.Var44 = uint16(m.Rng1.RandNext())
	b.CreationDay = m.CurrentDay
	b.Animation, b.CargoFrame = 0, 0
	b.Cargo = scenario.VehicleCargo{AcceptedTypes: 0, Type: 0xFF, Qty: 0}
	b.WheelSlipping, b.BreakdownFlags = 0, 0
	if bodyNumber == 0 && cargo.NumSimultaneousCargoTypes != 0 {
		b.Cargo.MaxQty = cargo.MaxCargo[0]
		b.Cargo.AcceptedTypes = cargo.CompatibleCargoCategories[0]
		if ct := bitScanForward(b.Cargo.AcceptedTypes); ct != -1 {
			b.Cargo.Type = uint8(ct)
		}
	}
	b.SpriteWidth, b.SpriteHeightNegative, b.SpriteHeightPositive = 1, 1, 1
	st := obj.CarComponents[bodyNumber].BodySpriteIdx
	if st != spriteIndexNull && st&spriteIndexIsReversed != 0 {
		b.Flags38 |= f38IsReversed
		st &^= spriteIndexIsReversed
	}
	b.ObjectSpriteType = st
	if st != spriteIndexNull && int(st) < len(obj.BodySprites) {
		s := obj.BodySprites[st]
		b.SpriteWidth, b.SpriteHeightNegative, b.SpriteHeightPositive = s.Width, s.HeightNegative, s.HeightPositive
	}
	b.ColourPrimary, b.ColourSecondary = scheme[0], scheme[1]
	if bodyNumber == 0 && obj.Flags&vehObjFlagJacobsBogieFront != 0 {
		b.Flags38 |= f38JacobsBogieAvailable
	}
	if bodyNumber+1 == obj.NumCarComponents && obj.Flags&vehObjFlagJacobsBogieRear != 0 {
		b.Flags38 |= f38JacobsBogieAvailable
	}
	id := b.ID
	m.setNextCar(last, id)
	return id
}

// createCar is CreateVehicle.cpp 0x004AE86D: a car of vehicle object objID
// (each car component a front bogie, a back bogie and a body) added before
// the tail, in the company's colours for the object's colour type; then
// updateTrainProperties. False without free entities.
func (m *Map) createCar(company uint8, head, objID uint16, om *objects.ObjectManager) bool {
	if m.EntityLists.numFreeEntities() <= maxNumVehicleComponentsInCar {
		return false
	}
	if om == nil || int(objID) >= len(om.Vehicles) || om.Vehicles[objID] == nil {
		return false
	}
	tr, err := m.TrainOf(head)
	if err != nil {
		return false
	}
	last := m.Entities[tr.Veh2].ID
	if len(tr.Cars) > 0 {
		cc := tr.Cars[len(tr.Cars)-1]
		last = m.Entities[cc[len(cc)-1].Body].ID
	}
	tail := m.Entities[tr.Tail].ID
	obj := om.Vehicles[objID]
	cargo := om.CargoFor(obj)
	scheme := m.vehicleColourScheme(company, obj.ColourType)
	for bn := uint8(0); bn < obj.NumCarComponents; bn++ {
		last = m.createFirstBogie(company, head, objID, obj, cargo, bn, last, scheme, om)
		last = m.createSecondBogie(company, head, objID, obj, bn, last, scheme)
		last = m.createBody(company, head, objID, obj, cargo, bn, last, scheme)
	}
	m.setNextCar(last, tail)
	if tr, err := m.TrainOf(head); err == nil {
		m.UpdateTrainProperties(tr, om)
	}
	return true
}
