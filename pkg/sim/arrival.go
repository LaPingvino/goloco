package sim

import (
	"github.com/LaPingvino/goloco/pkg/olmath"
)

// Arriving at a station — VehicleHead.cpp's landTryBeginUnloading
// (0x004A8DC0-side), setStationVisitedTypes, checkIfAtOrderStation
// (0x004B9987), updateLastJourneyAverageSpeed (0x004BACAF) and
// beginUnloading (0x004B99E1).
//
// updateLastJourneyAverageSpeed posts newSpeedRecord (its text PARTIAL:
// {VELOCITY}). PARTIAL: the
// owner's happy emotion (goloco has no messages or activeEmotions yet); the
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
	mod, recordType := int64(21), 0 // rail, road
	switch head.Mode {
	case 2: // air
		mod, recordType = 36, 1
	case 3: // water
		mod, recordType = 31, 2
	}
	adjDist := int64(dist) * mod
	adjTime := ticks
	for adjDist > 0xFFFFFFFF {
		adjDist >>= 1
		adjTime >>= 1
	}
	if adjTime == 0 {
		return
	}
	avg := int16(uint32(adjDist) / adjTime)
	avg = min(v2.MaxSpeed, avg)
	head.LastAverageSpeed = avg
	rec := &m.CompanyRecords // a new speed record for the mode's kind
	if avg <= int16(rec.Speed[recordType]) {
		return
	}
	rec.Speed[recordType] = uint16(avg)
	rec.Date[recordType] = m.CurrentDay
	rec.Company[recordType] = head.Owner
	m.PostMessage(nil, MsgNewSpeedRecord, head.Owner, head.ID, uint16(head.Owner), uint16(recordType))
}
