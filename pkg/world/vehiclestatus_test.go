package world

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// getStatus: heading for the first stop-at order round the ring (here found
// after wrapping past End), the speed second; signal waits, stopping,
// one-order tables, unplaced and red statuses.
func TestVehicleStatus(t *testing.T) {
	orders := make([]byte, 64)
	// table at 10: RouteThrough station 5, StopAt station 0x107, End.
	copy(orders[10:], []byte{2, 5, 1 | 0x40, 7, 0})
	w := &World{}
	w.Orders = orders
	v := scenario.VehicleListEntry{Status: scenario.VehicleStatusTravelling, OrderTableOffset: 10, CurrentOrder: 2, SizeOfOrderTable: 5, CurrentSpeed: 40 << 16}
	for _, c := range []struct {
		edit func(*scenario.VehicleListEntry)
		want string
		red  bool
	}{
		{func(*scenario.VehicleListEntry) {}, "Heading for Station 263 at 40mph", false},
		{func(v *scenario.VehicleListEntry) { v.CurrentOrder = 0 }, "Heading for Station 263 at 40mph", false},
		{func(v *scenario.VehicleListEntry) { v.Status = scenario.VehicleStatusWaitingAtSignal }, "Heading for Station 263 / Waiting at signal", false},
		{func(v *scenario.VehicleListEntry) { v.VehicleFlags = 1 << 1 }, "Stopping at 40mph", false},
		{func(v *scenario.VehicleListEntry) { v.SizeOfOrderTable = 1 }, "Travelling at 40mph", false},
		{func(v *scenario.VehicleListEntry) { v.TileX = -1 }, "No position", false},
		{func(v *scenario.VehicleListEntry) { v.Status = scenario.VehicleStatusCrashed }, "Crashed!", true},
		{func(v *scenario.VehicleListEntry) { v.BrokenDown = true }, "Broken down", true},
		{func(v *scenario.VehicleListEntry) { v.Status = scenario.VehicleStatusStopped }, "Stopped", false},
	} {
		e := v
		c.edit(&e)
		got, red := w.VehicleStatus(e)
		if got != c.want || red != c.red {
			t.Errorf("got %q (red %v), want %q (red %v)", got, red, c.want, c.red)
		}
	}
}
