package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// The vehicleChangeRunningMode game command (VehicleChangeRunningMode.cpp
// 0x004B694B): start, stop, local / express, manual driving.
// A player's stuck or crashed vehicle is released, stopped and sold
// (VehicleSell) as upstream. PARTIAL: the player's
// happy emotion when a profitable vehicle starts (company emotions, as
// elsewhere); the window invalidations and error texts.

// Running modes (VehicleChangeRunningModeArgs::Mode).
const (
	RunningModeStop = iota
	RunningModeStart
	RunningModeToggleLocalExpress
	RunningModeDriveManually
)

const (
	vehObjFlagMustHavePair = 1 << 12 // VehicleObjectFlags::mustHavePair
)

// VehicleChangeRunningMode is the vehicleChangeRunningMode game command: 0
// and true, or false when refused (starting or driving a vehicle that can't
// be started).
func (m *Map) VehicleChangeRunningMode(c *Cmd, flags uint8, headID uint16, mode int) (int32, bool) {
	tr, err := m.TrainOf(headID)
	if err != nil {
		return 0, false
	}
	head := &m.Entities[tr.Head]
	if mode == RunningModeToggleLocalExpress { // changeLocalExpressMode 0x004B6AEE
		if flags&CmdApply != 0 {
			m.Entities[tr.Veh1].Var48 ^= flags48ExpressMode
		}
		return 0, true
	}
	if (head.Status == StatusStuck || head.Status == StatusCrashed) && m.IsPlayerCompany(int(head.Owner)) {
		if flags&CmdApply != 0 { // 0x004B69C7
			m.Sub4AD778(tr, c.Objects)
			head = &m.Entities[tr.Head]
			head.Status = StatusStopped
			m.Entities[tr.Veh2].CurrentSpeed = 0
			m.VehicleSell(c, CmdApply, headID) // sargs.car = args.head
		}
		return 0, true
	}
	if mode == RunningModeDriveManually { // toggleManualDriving 0x004B6AAF
		if !m.canVehicleBeStarted(tr, c.Objects) {
			return 0, false
		}
		if flags&CmdApply == 0 {
			return 0, true
		}
		head.VehicleFlags ^= VehicleFlagManualControl
		head.ManualPower = -40
		if head.VehicleFlags&VehicleFlagManualControl != 0 {
			head.VehicleFlags &^= VehicleFlagCommandStop
		} else {
			head.VehicleFlags |= VehicleFlagCommandStop
		}
		if head.Status == StatusApproaching {
			head.StationID = 0xFFFF
			head.Status = StatusTravelling
		}
		return 0, true
	}
	// startStopVehicle 0x004B6A08
	start := mode == RunningModeStart
	if start && !m.canVehicleBeStarted(tr, c.Objects) {
		return 0, false
	}
	stopped := head.VehicleFlags&VehicleFlagCommandStop != 0
	if !start && stopped || start && !stopped || flags&CmdApply == 0 {
		return 0, true
	}
	head.VehicleFlags ^= VehicleFlagCommandStop
	if head.VehicleFlags&VehicleFlagCommandStop != 0 {
		head.VehicleFlags &^= VehicleFlagManualControl
	}
	return 0, true
}

// canVehicleBeStarted is VehicleChangeRunningMode.cpp 0x004B6B0C: not
// bankrupt, has cars, a driving cab in front (unless driven by hand), some
// power, and its must-have-a-pair vehicles in pairs.
func (m *Map) canVehicleBeStarted(tr Train, om *objects.ObjectManager) bool {
	head := &m.Entities[tr.Head]
	if int(head.Owner) < len(m.CompanyFaces) && m.CompanyFaces[head.Owner].ChallengeFlags&companyFlagBankrupt != 0 {
		return false // company_is_bankrupt
	}
	if len(tr.Cars) == 0 {
		return false // train_has_no_vehicles
	}
	manual := head.VehicleFlags&VehicleFlagManualControl != 0
	front := vehicleObject(om, m.Entities[tr.Cars[0][0].Front].ObjectID)
	if !manual && front != nil && front.Flags&vehicleFlagTopAndTail == 0 &&
		(front.Power == 0 || front.Flags&vehicleFlagCenter != 0) {
		return false // train_has_no_driving_cab
	}
	if m.Entities[tr.Veh2].TotalPower == 0 {
		return false // train_needs_a_locomotive_or_power_car
	}
	if !manual {
		pairID, pairs := uint16(0xFFFF), 0
		for _, car := range tr.Cars {
			id := m.Entities[car[0].Front].ObjectID
			if pairs == 0 {
				if o := vehicleObject(om, id); o != nil && o.Flags&vehObjFlagMustHavePair != 0 {
					pairID = id
					pairs++
				}
				continue
			}
			if id == pairID {
				pairs++
			}
		}
		if pairs&1 != 0 {
			return false // requires_an_extra_stringid
		}
	}
	return true
}
