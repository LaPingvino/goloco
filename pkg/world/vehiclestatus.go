package world

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/scenario"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
)

// Head Status values (VehicleHead.h Status).
const (
	headStatusUnk0            = scenario.VehicleStatusUnplaced
	headStatusStopped         = scenario.VehicleStatusStopped
	headStatusTravelling      = scenario.VehicleStatusTravelling
	headStatusWaitingAtSignal = scenario.VehicleStatusWaitingAtSignal
	headStatusApproaching     = scenario.VehicleStatusApproaching
	headStatusUnloading       = scenario.VehicleStatusUnloading
	headStatusLoading         = scenario.VehicleStatusLoading
	headStatusBrokenDown      = scenario.VehicleStatusBrokenDown
	headStatusCrashed         = scenario.VehicleStatusCrashed
	headStatusStuck           = scenario.VehicleStatusStuck
	headStatusLanding         = scenario.VehicleStatusLanding
	headStatusTaxiing1        = scenario.VehicleStatusTaxiing1
	headStatusTaxiing2        = scenario.VehicleStatusTaxiing2
	headStatusTakingOff       = scenario.VehicleStatusTakingOff
)

// orderSizes is OrderManager kOrderSizes (End, StopAt, RouteThrough,
// RouteWaypoint, UnloadAll, WaitFor).
var orderSizes = [6]int{1, 2, 2, 6, 1, 1}

// VehicleStatus is VehicleHead::getStatus formatted as VehicleList shows it
// after the name (vehicle_list_status_2pos / _3pos): status1, then status2
// after a space. red reports a {COLOUR RED} status (crashed, stuck, broken
// down), which colours the whole status.
func (w *World) VehicleStatus(v scenario.VehicleListEntry) (text string, red bool) {
	s1, s2, red := w.vehicleStatusParts(v)
	if s2 != "" {
		return s1 + " " + s2, red
	}
	return s1, red
}

func (w *World) vehicleStatusParts(v scenario.VehicleListEntry) (string, string, bool) {
	if v.TileX == -1 {
		return "No position", "", false
	}
	if v.BrokenDown {
		return "Broken down", "", true
	}
	at := func(format string, second bool) (string, string, bool) {
		if v.StationID == 0xFFFF {
			return w.statusTravelling(v)
		}
		s2 := ""
		if second {
			s2 = w.secondStatus(v)
		}
		return strings.Replace(format, "{STRINGID}", w.stationName(int(v.StationID)), 1), s2, false
	}
	switch v.Status {
	case headStatusUnk0:
		return "No position", "", false
	case headStatusStopped:
		return "Stopped", "", false
	case headStatusTravelling, headStatusWaitingAtSignal:
		return w.statusTravelling(v)
	case headStatusApproaching:
		return at("Approaching {STRINGID}", true)
	case headStatusUnloading:
		return at("Unloading at {STRINGID}", false)
	case headStatusLoading:
		return at("Loading at {STRINGID}", false)
	case headStatusBrokenDown:
		return "Broken down", "", true
	case headStatusCrashed:
		return "Crashed!", "", true
	case headStatusStuck:
		return "Stuck!", "", true
	case headStatusLanding:
		return at("Landing at {STRINGID}", true)
	case headStatusTaxiing1, headStatusTaxiing2:
		return at("Taxiing at {STRINGID}", true)
	case headStatusTakingOff:
		return at("Taking off from {STRINGID}", true)
	}
	return "", "", false
}

// statusTravelling is getStatusTravelling: stopping (commanded, or manual
// power ≤ −20), travelling with a one-entry order table, else heading for
// the first stop-at order from the current one round the ring.
func (w *World) statusTravelling(v scenario.VehicleListEntry) (string, string, bool) {
	const commandStop, manualControl = 1 << 1, 1 << 6
	s1 := "Travelling"
	switch {
	case v.VehicleFlags&commandStop != 0 || (v.VehicleFlags&manualControl != 0 && v.ManualPower <= -20):
		s1 = "Stopping"
	case v.SizeOfOrderTable == 1:
	default:
		if st, ok := w.firstStopAt(v); ok {
			s1 = "Heading for " + w.stationName(st)
		}
	}
	return s1, w.secondStatus(v), false
}

// firstStopAt walks OrderRingView(orderTableOffset, currentOrder) once: from
// the current order, each by its size, an End going back to the table's
// start, until back at the current one.
func (w *World) firstStopAt(v scenario.VehicleListEntry) (int, bool) {
	orders := w.Orders
	begin := int(v.OrderTableOffset)
	cur := begin + int(v.CurrentOrder)
	if cur >= len(orders) || orders[cur]&7 == 0 { // an empty table doesn't loop
		return 0, false
	}
	start := cur
	for steps := 0; steps < 4096; steps++ {
		if cur+1 >= len(orders) {
			return 0, false
		}
		t := orders[cur] & 7
		if t >= uint8(len(orderSizes)) {
			return 0, false // "Invalid order type!"
		}
		if t == 1 { // OrderStopAt: station = (type & 0xC0) << 2 | data
			return int(orders[cur]&0xC0)<<2 | int(orders[cur+1]), true
		}
		cur += orderSizes[t]
		if cur >= len(orders) || orders[cur]&7 == 0 {
			cur = begin
		}
		if cur == start {
			break
		}
	}
	return 0, false
}

// secondStatus is getSecondStatus: "/ Waiting at signal", or "at" the speed
// (toSpeed16(veh2 currentSpeed), {VELOCITY}).
func (w *World) secondStatus(v scenario.VehicleListEntry) string {
	if v.Status == headStatusWaitingAtSignal {
		return "/ Waiting at signal"
	}
	return "at " + L.Velocity(int(v.CurrentSpeed/65536)) // toSpeed16: raw / 65536
}

// VehicleCargos is what vehicleIsTransportingCargo reads: the cargo ids of
// the UnloadAll and WaitFor orders (getCargo = type >> 3) in the vehicle's
// order ring from the table's start.
func (w *World) VehicleCargos(v scenario.VehicleListEntry) []int {
	orders := w.Orders
	begin := int(v.OrderTableOffset)
	if begin >= len(orders) || orders[begin]&7 == 0 {
		return nil
	}
	var out []int
	cur := begin
	for steps := 0; steps < 4096; steps++ {
		t := orders[cur] & 7
		if t >= uint8(len(orderSizes)) {
			return out
		}
		if t == 4 || t == 5 { // UnloadAll, WaitFor
			out = append(out, int(orders[cur]>>3))
		}
		cur += orderSizes[t]
		if cur >= len(orders) || orders[cur]&7 == 0 {
			break // back to the start: the ring is done
		}
	}
	return out
}
