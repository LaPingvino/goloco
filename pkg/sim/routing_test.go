package sim

import "testing"

// RoutingManager: handles pack vehicleRef·64 | index and wrap the index;
// allocation takes the first null slot and fills it with
// kAllocatedButFreeRouting; free returns it to null.
func TestRoutingManager(t *testing.T) {
	var m Map
	h := NewRoutingHandle(3, 63)
	if h.VehicleRef() != 3 || h.Index() != 63 || h.WithIndex(64).Index() != 0 || h.WithIndex(64).VehicleRef() != 3 {
		t.Fatalf("handle %d: ref %d index %d", h, h.VehicleRef(), h.Index())
	}
	a, ok := m.AllocateRoutingHandle()
	b, _ := m.AllocateRoutingHandle()
	if !ok || a.VehicleRef() != 0 || b.VehicleRef() != 1 || a.Index() != 0 {
		t.Fatalf("allocated %d, %d", a, b)
	}
	if m.Routing(a.WithIndex(17)) != RoutingAllocatedButFree {
		t.Error("allocated slot not filled with kAllocatedButFreeRouting")
	}
	m.SetRouting(a.WithIndex(1), 0x123)
	if m.Routing(a.WithIndex(1)) != 0x123 {
		t.Error("setRouting/getRouting")
	}
	m.FreeRoutingHandle(a)
	if c, _ := m.AllocateRoutingHandle(); c.VehicleRef() != 0 || m.Routing(c.WithIndex(1)) != RoutingAllocatedButFree {
		t.Errorf("freed slot not reused cleanly: %d", c)
	}
}

// RingView: empty from a free routing; otherwise up to the first free one,
// wrapping past index 63; backwards likewise; a full ring stops at begin.
func TestRoutingRing(t *testing.T) {
	var m Map
	h, _ := m.AllocateRoutingHandle()
	if r := m.RoutingRing(h.WithIndex(5), false); len(r) != 0 {
		t.Errorf("from a free routing: %v", r)
	}
	for _, i := range []uint8{62, 63, 0, 1} {
		m.SetRouting(h.WithIndex(i), uint16(i))
	}
	r := m.RoutingRing(h.WithIndex(62), false)
	if len(r) != 4 || r[3].Index() != 1 {
		t.Errorf("forward from 62: %v", r)
	}
	r = m.RoutingRing(h.WithIndex(1), true)
	if len(r) != 4 || r[3].Index() != 62 {
		t.Errorf("backward from 1: %v", r)
	}
	for i := range 64 {
		m.SetRouting(h.WithIndex(uint8(i)), 7)
	}
	if r := m.RoutingRing(h.WithIndex(10), false); len(r) != 64 {
		t.Errorf("full ring: %d", len(r))
	}
}
