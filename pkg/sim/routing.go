package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Routing — a port of Vehicles/RoutingManager.cpp over GameState routings
// (Map.Routings, flat [vehicleRef*64 + index]). Each routing is the
// trackAndDirection (with AdditionalTaDFlags) of a piece a vehicle is on or
// has reserved ahead; a vehicle's routings form a 64-entry ring.

const (
	// RoutingAllocatedButFree is kAllocatedButFreeRouting.
	RoutingAllocatedButFree uint16 = 0xFFFE
	// RoutingNull is kRoutingNull (slot not allocated to any vehicle).
	RoutingNull uint16 = 0xFFFF
)

// RoutingHandle is Vehicles::RoutingHandle: vehicleRef·64 | index.
type RoutingHandle uint16

func NewRoutingHandle(vehicleRef uint16, index uint8) RoutingHandle {
	return RoutingHandle(vehicleRef*scenario.RoutingsPerVehicle | uint16(index)%scenario.RoutingsPerVehicle)
}
func (h RoutingHandle) VehicleRef() uint16 { return uint16(h) / scenario.RoutingsPerVehicle }
func (h RoutingHandle) Index() uint8       { return uint8(uint16(h) % scenario.RoutingsPerVehicle) }

// WithIndex is setIndex (the index wraps within the vehicle's 64).
func (h RoutingHandle) WithIndex(i uint8) RoutingHandle {
	return RoutingHandle(uint16(h)&^(scenario.RoutingsPerVehicle-1) | uint16(i)&(scenario.RoutingsPerVehicle-1))
}

func (m *Map) ensureRoutings() {
	if len(m.Routings) == 0 {
		m.ResetRoutingTable()
	}
}

// ResetRoutingTable is resetRoutingTable (0x004A8810).
func (m *Map) ResetRoutingTable() {
	m.Routings = make([]uint16, scenario.MaxRoutingVehicles*scenario.RoutingsPerVehicle)
	for i := range m.Routings {
		m.Routings[i] = RoutingNull
	}
}

func (m *Map) vehicleRoutings(ref uint16) []uint16 {
	m.ensureRoutings()
	b := int(ref) * scenario.RoutingsPerVehicle
	return m.Routings[b : b+scenario.RoutingsPerVehicle]
}

func (m *Map) findFreeRoutingVehicleRef() (uint16, bool) {
	m.ensureRoutings()
	for ref := 0; ref < scenario.MaxRoutingVehicles; ref++ {
		if m.Routings[ref*scenario.RoutingsPerVehicle] == RoutingNull {
			return uint16(ref), true
		}
	}
	return 0, false
}

// IsEmptyRoutingSlotAvailable is isEmptyRoutingSlotAvailable.
func (m *Map) IsEmptyRoutingSlotAvailable() bool {
	_, ok := m.findFreeRoutingVehicleRef()
	return ok
}

// AllocateRoutingHandle is getAndAllocateFreeRoutingHandle (0x004B1E00).
func (m *Map) AllocateRoutingHandle() (RoutingHandle, bool) {
	ref, ok := m.findFreeRoutingVehicleRef()
	if !ok {
		return 0, false
	}
	fill(m.vehicleRoutings(ref), RoutingAllocatedButFree)
	return NewRoutingHandle(ref, 0), true
}

// FreeRoutingHandle is freeRoutingHandle (0x004B1E77).
func (m *Map) FreeRoutingHandle(h RoutingHandle) {
	fill(m.vehicleRoutings(h.VehicleRef()), RoutingNull)
}

// ResetRoutings is resetRoutings: every routing of h's vehicle freed.
func (m *Map) ResetRoutings(h RoutingHandle) {
	fill(m.vehicleRoutings(h.VehicleRef()), RoutingAllocatedButFree)
}

// Routing is getRouting.
func (m *Map) Routing(h RoutingHandle) uint16 { return m.vehicleRoutings(h.VehicleRef())[h.Index()] }

// SetRouting is setRouting.
func (m *Map) SetRouting(h RoutingHandle, routing uint16) {
	m.vehicleRoutings(h.VehicleRef())[h.Index()] = routing
}

// FreeRouting is freeRouting.
func (m *Map) FreeRouting(h RoutingHandle) { m.SetRouting(h, RoutingAllocatedButFree) }

// RoutingRing is RingView(begin) iterated forwards (begin()…end()) or, with
// reverse, backwards (rbegin()…rend()): empty if begin's routing is free;
// otherwise begin, then each next index (wrapping within the vehicle's 64)
// until a free routing or back at begin.
func (m *Map) RoutingRing(begin RoutingHandle, reverse bool) []RoutingHandle {
	if m.Routing(begin) == RoutingAllocatedButFree {
		return nil
	}
	out := []RoutingHandle{begin}
	cur := begin
	for {
		if reverse {
			cur = cur.WithIndex(cur.Index() - 1)
		} else {
			cur = cur.WithIndex(cur.Index() + 1)
		}
		if cur == begin || m.Routing(cur) == RoutingAllocatedButFree {
			return out
		}
		out = append(out, cur)
	}
}

func fill(s []uint16, v uint16) {
	for i := range s {
		s[i] = v
	}
}
