package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// The head's own update — VehicleHead::updateLand (0x004A8B81's land branch)
// for its normal-movement path: landNormalMovementUpdate (advance the order,
// sub_4ACEE7 = sub_4ACEF1 / sub_47DA8D) and trainNormalMovementUpdate /
// roadNormalMovementUpdate.
//
// Approaching runs landTryBeginUnloading (arrival.go), unloading
// updateUnloadCargo (unload.go). PARTIAL (item 5.6/5.7): loading, crashed, stoppingUpdate (command stop, breakdowns), manual
// control's sub_4A8C81, the road signal timeouts (categoriseTimeElapsed,
// sub_4A8DB7, tryReverse) and the rail signal-wait reversing
// (timeAtSignal, landReverseFromSignal, sub_4AC1C2) are not ported: a head in
// those states does nothing this tick.

// Head statuses (VehicleHead.h Status).
const (
	StatusStopped         uint8 = 1
	StatusWaitingAtSignal uint8 = 3
	StatusApproaching     uint8 = 4
	StatusUnloading       uint8 = 5
	StatusLoading         uint8 = 6
	StatusCrashed         uint8 = 7
	StatusStuck           uint8 = 8
)

// updateHead runs updateLand's normal movement for a head that is not
// approaching, (un)loading, crashed or stuck; unkDistance1 is the tick's.
func (m *Map) updateHead(tr Train, unkDistance1 int32, om *objects.ObjectManager) {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	notBroken := v2.Var73&1 == 0 || v2.Var73&2 != 0 // !isBrokenDown || isStillPowered
	if notBroken && head.Status == StatusApproaching {
		// (road signal timeouts, command stop, manual: PARTIAL)
		m.LandTryBeginUnloading(tr)
		return
	}
	switch head.Status {
	case StatusUnloading:
		m.UpdateUnloadCargo(tr, om) // unload.go
		return
	case StatusLoading, StatusCrashed, StatusStuck:
		return // PARTIAL: see the file comment
	}
	head.Status = StatusTravelling
	brokenDown := v2.Var73&1 != 0 && v2.Var73&2 == 0 // isBrokenDown and not isStillPowered
	if brokenDown || head.VehicleFlags&VehicleFlagCommandStop != 0 {
		return // stoppingUpdate: PARTIAL
	}
	if head.VehicleFlags&VehicleFlagManualControl != 0 && head.ManualPower <= -20 {
		return // sub_4A8C81: PARTIAL
	}
	// landNormalMovementUpdate.
	m.AdvanceToNextRoutableOrder(head.OrderTableOffset, &head.CurrentOrder, head.SizeOfOrderTable)
	var r Sub4ACEE7Result
	if head.Mode == 1 {
		r = m.Sub47DA8D(tr, 0xD4CB00, uint32(unkDistance1), om)
	} else {
		r = m.Sub4ACEF1(tr, 0xD4CB00, uint32(unkDistance1), false)
	}
	if r.Al == 4 { // both: arrived at a station to stop at
		head.Status = StatusApproaching
		head.StationID = uint16(r.StationID)
		return
	}
	if head.Mode == 0 && r.Al == 3 {
		if v2.RoutingHandle != head.RoutingHandle || v2.SubPosition != head.SubPosition {
			return
		}
		head.Status = StatusWaitingAtSignal
		// trainsReverseAtSignals (default on): timeAtSignal / reversing — PARTIAL
	}
	// road al == 2 (tryReverse when Vehicle2 has caught up): PARTIAL
}
