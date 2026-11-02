package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// The head's own update — VehicleHead::updateLand (0x004A8B81's land branch)
// for its normal-movement path: landNormalMovementUpdate (advance the order,
// sub_4ACEE7 = sub_4ACEF1 / sub_47DA8D) and trainNormalMovementUpdate /
// roadNormalMovementUpdate.
//
// Approaching runs landTryBeginUnloading (arrival.go), unloading
// updateUnloadCargo (unload.go), loading landLoadingUpdate (load.go).
// PARTIAL (item 5.6): crashed (landCrashedUpdate) and the manual
// control's crash (destroyTrain) is not ported.

// Head statuses (VehicleHead.h Status).
const (
	StatusStopped         uint8 = 1
	StatusWaitingAtSignal uint8 = 3
	StatusApproaching     uint8 = 4
	StatusUnloading       uint8 = 5
	StatusLoading         uint8 = 6
	StatusCrashed         uint8 = 8
	StatusStuck           uint8 = 9
)

// updateHead runs updateLand's normal movement for a head that is not
// approaching, (un)loading, crashed or stuck; unkDistance1 is the tick's.
func (m *Map) updateHead(tr Train, unkDistance1 int32, om *objects.ObjectManager) {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	notBroken := v2.Var73&1 == 0 || v2.Var73&2 != 0 // !isBrokenDown || isStillPowered
	if notBroken && head.Status == StatusApproaching {
		if head.Mode == 1 {
			switch m.CategoriseTimeElapsed(tr, om) {
			case TimeoutFirst:
				m.Sub4A8DB7(tr, om)
				return
			case TimeoutTurnaround:
				m.TryReverse(tr, om)
				return
			}
		}
		if head.VehicleFlags&VehicleFlagCommandStop != 0 {
			m.StoppingUpdate(tr)
			return
		}
		if head.VehicleFlags&VehicleFlagManualControl != 0 && head.ManualPower <= -20 {
			m.sub4A8C81(tr, unkDistance1, om)
			return
		}
		m.LandTryBeginUnloading(tr)
		return
	}
	switch head.Status {
	case StatusUnloading:
		m.UpdateUnloadCargo(tr, om) // unload.go
		return
	case StatusLoading:
		m.LandLoadingUpdate(tr, om) // load.go
		return
	case StatusCrashed, StatusStuck:
		return // PARTIAL: see the file comment
	}
	head.Status = StatusTravelling
	brokenDown := v2.Var73&1 != 0 && v2.Var73&2 == 0 // isBrokenDown and not isStillPowered
	if brokenDown || head.VehicleFlags&VehicleFlagCommandStop != 0 {
		m.StoppingUpdate(tr)
		return
	}
	if head.VehicleFlags&VehicleFlagManualControl != 0 && head.ManualPower <= -20 {
		m.sub4A8C81(tr, unkDistance1, om)
		return
	}
	m.landNormalMovementUpdate(tr, unkDistance1, om)
}

// sub4A8C81 is sub_4A8C81 (manual control braking): still moving → normal
// movement; stopped over a station → arrive (beginUnloading) and stop.
func (m *Map) sub4A8C81(tr Train, unkDistance1 int32, om *objects.ObjectManager) {
	head := &m.Entities[tr.Head]
	if m.Entities[tr.Veh2].CurrentSpeed > 1<<16 {
		m.landNormalMovementUpdate(tr, unkDistance1, om)
		return
	}
	if st := m.manualFindTrainStationAtLocation(tr); st >= 0 {
		head.StationID = uint16(st)
		if int(head.StationID) < len(m.StationNames) {
			m.StationNames[head.StationID].VisitedTypes |= 1 << (head.VehicleType & 7)
		}
		if orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder); len(orders) > 0 && orders[0].HasFlags(OrderHasStation) && orders[0].Station() == st {
			head.CurrentOrder = uint16(orders[1%len(orders)].Offset - int(head.OrderTableOffset))
		}
		m.updateLastJourneyAverageSpeed(tr)
		m.beginUnloading(tr)
	}
	m.StoppingUpdate(tr)
}

// landNormalMovementUpdate: advance the order, reserve (sub_4ACEE7), then the
// train / road normal movement update.
func (m *Map) landNormalMovementUpdate(tr Train, unkDistance1 int32, om *objects.ObjectManager) {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	m.AdvanceToNextRoutableOrder(head.OrderTableOffset, &head.CurrentOrder, head.SizeOfOrderTable)
	var r Sub4ACEE7Result
	if head.Mode == 1 {
		r = m.Sub47DA8D(tr, 0xD4CB00, uint32(unkDistance1), om)
		// roadNormalMovementUpdate: the timeouts first.
		switch m.CategoriseTimeElapsed(tr, om) {
		case TimeoutFirst:
			m.Sub4A8DB7(tr, om)
			return
		case TimeoutTurnaround:
			m.TryReverse(tr, om)
			return
		}
	} else {
		r = m.Sub4ACEF1(tr, 0xD4CB00, uint32(unkDistance1), false)
	}
	if r.Al == 4 { // both: arrived at a station to stop at
		head.Status = StatusApproaching
		head.StationID = uint16(r.StationID)
		return
	}
	if head.Mode == 1 && r.Al == 2 { // no route: turn round once Vehicle2 has caught up
		if head.RoutingHandle == v2.RoutingHandle && v2.SubPosition == head.SubPosition {
			m.TryReverse(tr, om)
		}
		return
	}
	if head.Mode == 0 {
		m.trainSignalUpdate(tr, r, om) // reversal.go
	}
}
