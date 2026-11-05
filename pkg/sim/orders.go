package sim

// Orders — a port of Vehicles/Orders.h / Orders.cpp (order types, kOrderFlags,
// the field getters) and OrderManager's OrderRingView, over the GameState
// orders table (Map.Orders), plus VehicleHead::advanceToNextRoutableOrder.

// OrderType.
const (
	OrderEnd           uint8 = 0
	OrderStopAt        uint8 = 1
	OrderRouteThrough  uint8 = 2
	OrderRouteWaypoint uint8 = 3
	OrderUnloadAll     uint8 = 4
	OrderWaitFor       uint8 = 5
)

// OrderFlags.
const (
	OrderIsRoutable uint8 = 1 << 0
	OrderHasNumber  uint8 = 1 << 1
	OrderHasCargo   uint8 = 1 << 2
	OrderHasStation uint8 = 1 << 3
)

// kOrderFlags (0x004FE088) and kOrderSizes, by OrderType.
var (
	orderFlags = [6]uint8{0, OrderIsRoutable | OrderHasNumber | OrderHasStation, OrderIsRoutable | OrderHasNumber | OrderHasStation,
		OrderIsRoutable | OrderHasNumber, OrderHasCargo, OrderHasCargo}
	orderSizes = [6]int{1, 2, 2, 6, 1, 1}
)

// Order is a view of the order at Offset in Map.Orders.
type Order struct {
	m      *Map
	Offset int
}

func (o Order) b(i int) uint8 {
	if o.Offset+i < len(o.m.Orders) {
		return o.m.Orders[o.Offset+i]
	}
	return 0
}

// Type is getType.
func (o Order) Type() uint8 { return o.b(0) & 7 }

// HasFlags is hasFlags (kOrderFlags).
func (o Order) HasFlags(f uint8) bool {
	t := o.Type()
	return int(t) < len(orderFlags) && orderFlags[t]&f != 0
}

// Station is OrderStation::getStation (StopAt, RouteThrough).
func (o Order) Station() int { return int(o.b(0)&0xC0)<<2 | int(o.b(1)) }

// Cargo is OrderCargo::getCargo (UnloadAll, WaitFor).
func (o Order) Cargo() uint8 { return o.b(0) >> 3 }

// Waypoint is OrderRouteWaypoint::getWaypoint (world units).
func (o Order) Waypoint() (x, y, z int32) {
	x = (int32(o.b(0)&0x80)<<1 | int32(o.b(1))) * 32
	y = (int32(o.b(2)&0x80)<<1 | int32(o.b(3))) * 32
	z = int32(o.b(2)&0x7F) * 8
	return
}

// Direction / TrackID are OrderRouteWaypoint::getDirection / getTrackId.
func (o Order) Direction() uint8 { return o.b(4) & 7 }
func (o Order) TrackID() uint8   { return o.b(4)>>3 | (o.b(5)&1)<<5 }

// OrderRing is OrderRingView(beginTableOffset, currentOrderOffset) iterated:
// from the current order, each by its size, an End wrapping to the table's
// start, until back at the current order; empty when that is an End. An
// invalid order type ends the walk (upstream throws "Invalid order type!").
func (m *Map) OrderRing(begin uint32, current uint16) []Order {
	start := int(begin) + int(current)
	cur := start
	o := Order{m, cur}
	if cur >= len(m.Orders) || o.Type() == OrderEnd {
		return nil
	}
	var out []Order
	for {
		out = append(out, o)
		t := o.Type()
		if int(t) >= len(orderSizes) {
			return out
		}
		cur += orderSizes[t]
		o = Order{m, cur}
		if cur >= len(m.Orders) || o.Type() == OrderEnd {
			cur = int(begin)
			o = Order{m, cur}
		}
		if cur == start || len(out) > len(m.Orders) {
			return out
		}
	}
}

// AdvanceToNextRoutableOrder is VehicleHead::advanceToNextRoutableOrder:
// currentOrder moves to the first routable order from it in the ring.
func (m *Map) AdvanceToNextRoutableOrder(orderTableOffset uint32, currentOrder *uint16, sizeOfOrderTable uint16) {
	if sizeOfOrderTable == 1 {
		return
	}
	for _, o := range m.OrderRing(orderTableOffset, *currentOrder) {
		if o.HasFlags(OrderIsRoutable) {
			*currentOrder = uint16(o.Offset - int(orderTableOffset))
			return
		}
	}
}

// OrderTable is getOrderTable(head): the vehicle's orders from its table's
// start (orderTableOffset) up to, not including, the End order. An invalid
// order type ends the walk.
func (m *Map) OrderTable(begin uint32) []Order {
	var out []Order
	for cur := int(begin); cur < len(m.Orders); {
		o := Order{m, cur}
		t := o.Type()
		if t == OrderEnd || int(t) >= len(orderSizes) {
			break
		}
		out = append(out, o)
		cur += orderSizes[t]
	}
	return out
}
