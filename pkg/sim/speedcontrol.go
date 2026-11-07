package sim

// VehicleSpeedControl is the vehicleSpeedControl game command (0x004BAB63):
// applied, the head's manualPower becomes speed (−40 … 40: brake … power);
// a vehicle stopped by command that is stuck or crashed is sold (upstream's
// own reading of vehicleChangeRunningMode's sale). The caller pays under
// TrainRunningCosts (cost 0).
func (m *Map) VehicleSpeedControl(c *Cmd, flags uint8, headID uint16, speed int8) (int32, bool) {
	if flags&CmdApply == 0 {
		return 0, true
	}
	head := m.entity(headID)
	if head == nil {
		return 0, false
	}
	head.ManualPower = speed
	if head.VehicleFlags&VehicleFlagCommandStop != 0 && (head.Status == StatusStuck || head.Status == StatusCrashed) {
		nested(CmdApply, func(f uint8) (int32, bool) { return m.VehicleSell(c, f, headID) })
	}
	return 0, true
}
