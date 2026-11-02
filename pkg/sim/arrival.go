package sim

import (
	"github.com/LaPingvino/goloco/pkg/olmath"
)

// Arriving at a station — VehicleHead.cpp's landTryBeginUnloading
// (0x004A8DC0-side), setStationVisitedTypes, checkIfAtOrderStation
// (0x004B9987), updateLastJourneyAverageSpeed (0x004BACAF) and
// beginUnloading (0x004B99E1).
//
// PARTIAL: updateLastJourneyAverageSpeed's speed records (CompanyManager
// records, the newSpeedRecord message, company emotion) are not ported; the
// unloading / loading that follow (updateUnloadCargo, updateLoadCargo,
// beginNewJourney, item 5.7 part 2) are not ported, so an unloading vehicle
// waits there.

// BreakdownFlags.
const (
	BreakdownAwaitingCargoTransfer uint8 = 1 << 0
	BreakdownJourneyStarted        uint8 = 1 << 3
)

// LandTryBeginUnloading is landTryBeginUnloading: once Vehicle2 has caught up
// with the head (the whole train is in), and not under manual control, the
// arrival is recorded and unloading begins.
func (m *Map) LandTryBeginUnloading(tr Train) {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	if head.RoutingHandle != v2.RoutingHandle || v2.SubPosition != head.SubPosition {
		return
	}
	if head.VehicleFlags&VehicleFlagManualControl != 0 {
		return
	}
	// setStationVisitedTypes
	if int(head.StationID) < len(m.StationNames) {
		m.StationNames[head.StationID].VisitedTypes |= 1 << (head.VehicleType & 7)
	}
	// checkIfAtOrderStation
	if orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder); len(orders) > 0 && orders[0].HasFlags(OrderHasStation) &&
		orders[0].Station() == int(head.StationID) {
		next := orders[1%len(orders)]
		head.CurrentOrder = uint16(next.Offset - int(head.OrderTableOffset))
	}
	m.updateLastJourneyAverageSpeed(tr)
	m.beginUnloading(tr)
}

// beginUnloading is beginUnloading.
func (m *Map) beginUnloading(tr Train) {
	head := &m.Entities[tr.Head]
	head.BreakdownFlags &^= BreakdownAwaitingCargoTransfer
	head.Status = StatusUnloading
	head.CargoTransferTimeout = 10
	head.Var58 = 0
	for _, car := range tr.Cars {
		for _, cc := range car {
			for _, i := range []int{cc.Front, cc.Back, cc.Body} {
				m.Entities[i].BreakdownFlags |= BreakdownAwaitingCargoTransfer
			}
		}
	}
}

// updateLastJourneyAverageSpeed: distance2D from the journey's start to
// Vehicle2, × 21 for land vehicles, over the ticks taken (both halved until
// the distance fits 32 bits), capped at Vehicle2's maxSpeed.
func (m *Map) updateLastJourneyAverageSpeed(tr Train) {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	if head.BreakdownFlags&BreakdownJourneyStarted == 0 {
		return
	}
	head.BreakdownFlags &^= BreakdownJourneyStarted
	dist := olmath.Distance2D(olmath.Vector2i{X: int32(head.JourneyStartX), Y: int32(head.JourneyStartY)}, olmath.Vector2i{X: int32(v2.PosX), Y: int32(v2.PosY)})
	ticks := m.ScenarioTicks - head.JourneyStartTicks
	adjDist := int64(dist) * 21
	adjTime := ticks
	for adjDist > 0xFFFFFFFF {
		adjDist >>= 1
		adjTime >>= 1
	}
	if adjTime == 0 {
		return
	}
	avg := int16(uint32(adjDist) / adjTime)
	head.LastAverageSpeed = min(v2.MaxSpeed, avg)
}
