package sim

// OrderManager's insertOrder / deleteOrder (OrderManager.cpp 0x004704AB /
// 0x004705C0) and the order game commands: vehicleOrderInsert (0x0047036E),
// vehicleOrderDelete (0x0047057A), vehicleOrderSkip (0x0047071A).
// PARTIAL: the order window's invalidation and display frames (goloco keeps
// none; reoffsetVehicleOrderTables moves them upstream), the error texts.

const kMaxOrdersPerVehicle = 63 // kMaxRoutingsPerVehicle - 1

// reoffsetVehicleOrderTables is OrderManager's: every head's table at or past
// offset moves by delta.
func (m *Map) reoffsetVehicleOrderTables(offset uint32, delta int) {
	for i := range m.Entities {
		e := &m.Entities[i]
		if e.EntityType == 0 && e.OrderTableOffset >= offset {
			e.OrderTableOffset = uint32(int64(e.OrderTableOffset) + int64(delta))
		}
	}
}

// insertOrder is OrderManager::insertOrder: raw (its type's size) put at the
// head's order offset, the orders after shifted right, the current order
// kept on the same order, the other vehicles' tables moved.
func (m *Map) insertOrder(headID uint16, orderOffset uint16, raw []byte) {
	h := m.entity(headID)
	n := orderSizes[raw[0]&7]
	if h.SizeOfOrderTable > 1 && h.CurrentOrder >= orderOffset {
		h.CurrentOrder += uint16(n)
	}
	h.SizeOfOrderTable += uint16(n)
	at, end := int(h.OrderTableOffset)+int(orderOffset), int(m.OrderTableLength)
	if len(m.Orders) < end+n {
		m.Orders = append(m.Orders, make([]byte, end+n-len(m.Orders))...)
	}
	copy(m.Orders[at+n:end+n], m.Orders[at:end]) // shiftOrdersRight
	m.OrderTableLength += uint32(n)
	copy(m.Orders[at:at+n], raw[:n])
	m.reoffsetVehicleOrderTables(h.OrderTableOffset+1, n)
}

// deleteOrder is OrderManager::deleteOrder.
func (m *Map) deleteOrder(headID uint16, orderOffset uint16) {
	h := m.entity(headID)
	at := int(h.OrderTableOffset) + int(orderOffset)
	n := orderSizes[m.Orders[at]&7]
	h.SizeOfOrderTable -= uint16(n)
	if h.CurrentOrder > orderOffset {
		h.CurrentOrder -= uint16(n)
	}
	if h.CurrentOrder+1 >= h.SizeOfOrderTable {
		h.CurrentOrder = 0
	}
	end := int(m.OrderTableLength)
	tmp := append([]byte(nil), m.Orders[at:at+n]...)
	copy(m.Orders[at:], m.Orders[at+n:end]) // shiftOrdersLeft (std::rotate)
	copy(m.Orders[end-n:end], tmp)
	m.OrderTableLength -= uint32(n)
	m.reoffsetVehicleOrderTables(h.OrderTableOffset+uint32(orderOffset)+1, -n)
}

// VehicleOrderInsert is the vehicleOrderInsert command: raw is the order's
// bytes (its type's size). Refused: a station of another company, a route
// order for an aircraft (or route-through for a ship), no room in the game's
// or the vehicle's order table, an offset past the table. A stop at the
// station the order before stops at turns that order into route-through.
func (m *Map) VehicleOrderInsert(c *Cmd, flags uint8, headID, orderOffset uint16, raw []byte) (int32, bool) {
	h := m.entity(headID)
	if h == nil || h.EntityType != 0 || len(raw) == 0 || int(raw[0]&7) >= len(orderSizes) || len(raw) < orderSizes[raw[0]&7] {
		return 0, false
	}
	o := Order{&Map{Orders: raw}, 0}
	if o.HasFlags(OrderHasStation) {
		if st := o.Station(); st >= len(m.StationNames) || m.StationNames[st].Owner != h.Owner {
			return 0, false // stationOwnedByAnotherCompany
		}
	}
	if t := o.Type(); t == OrderRouteThrough || t == OrderRouteWaypoint {
		if h.Mode == 3 && t == OrderRouteThrough || h.Mode == 2 {
			return 0, false // not valid for ships / aircraft
		}
	}
	if int(m.OrderTableLength)+orderSizes[o.Type()] > maxOrders {
		return 0, false // no_space_for_more_vehicle_orders
	}
	if len(m.OrderRing(h.OrderTableOffset, 0)) >= kMaxOrdersPerVehicle {
		return 0, false // tooManyOrdersForThisVehicle
	}
	if orderOffset > h.SizeOfOrderTable {
		return 0, false
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	if o.Type() == OrderStopAt && h.Mode != 3 && h.Mode != 2 {
		ring := m.OrderRing(h.OrderTableOffset, 0)
		target := int(h.OrderTableOffset) + int(orderOffset)
		// upstream's lagging iterator: previous stays at the first order when
		// that is the target (so the order AT the insertion point is compared
		// then), else it ends one behind the target (or on the last order)
		prev := 0
		if len(ring) > 0 && ring[0].Offset != target {
			for i := 1; i < len(ring) && ring[i].Offset != target; i++ {
				prev++
			}
		}
		if len(ring) > 0 {
			if p := ring[prev]; p.Type() == OrderStopAt && p.Station() == o.Station() {
				m.Orders[p.Offset] = m.Orders[p.Offset]&^7 | OrderRouteThrough
				return 0, true
			}
		}
	}
	m.insertOrder(headID, orderOffset, raw)
	return 0, true
}

// VehicleOrderDelete is the vehicleOrderDelete command.
func (m *Map) VehicleOrderDelete(c *Cmd, flags uint8, headID, orderOffset uint16) (int32, bool) {
	h := m.entity(headID)
	if h == nil || h.EntityType != 0 {
		return 0, false
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	if orderOffset > h.SizeOfOrderTable {
		return 0, false
	}
	m.deleteOrder(headID, orderOffset)
	return 0, true
}

// VehicleOrderSkip is the vehicleOrderSkip command: the current order becomes
// the next one in the ring.
func (m *Map) VehicleOrderSkip(c *Cmd, flags uint8, headID uint16) (int32, bool) {
	h := m.entity(headID)
	if h == nil || h.EntityType != 0 {
		return 0, false
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	ring := m.OrderRing(h.OrderTableOffset, h.CurrentOrder)
	if len(ring) > 1 {
		h.CurrentOrder = uint16(ring[1].Offset - int(h.OrderTableOffset))
	} else if len(ring) == 1 {
		h.CurrentOrder = uint16(ring[0].Offset - int(h.OrderTableOffset))
	}
	return 0, true
}

// OrderStopAtRaw / OrderRouteThroughRaw are those orders' bytes for a station.
func OrderStopAtRaw(station int) []byte {
	return []byte{OrderStopAt | uint8(station>>2)&0xC0, uint8(station)}
}
