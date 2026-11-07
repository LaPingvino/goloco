package sim

// ExpenditureTrainRunningCosts is ExpenditureType::TrainRunningCosts.
const ExpenditureTrainRunningCosts = 1

// VehiclePassSignal is the vehiclePassSignal game command (0x004B0B50):
// refused for another company's vehicle or one that can't be modified;
// applied to a placed vehicle, the head's var_5C becomes 2 and Vehicle1's
// var_48 gets passSignal, so it passes the next signal at danger. The
// caller pays under TrainRunningCosts (cost 0) at Vehicle2's position.
func (m *Map) VehiclePassSignal(c *Cmd, flags uint8, headID uint16) (int32, bool) {
	tr, err := m.TrainOf(headID)
	if err != nil {
		return 0, false
	}
	v2 := m.Entities[tr.Veh2]
	c.Pos = &[3]int32{int32(v2.PosX), int32(v2.PosY), int32(v2.PosZ)} // setPosition(veh2->position)
	head := &m.Entities[tr.Head]
	if !checkCompanyCompatibility(c, head.Owner) || !m.canBeModified(tr) {
		return 0, false
	}
	if flags&CmdApply == 0 || head.TileX == -1 {
		return 0, true
	}
	head.Var5C = 2
	m.Entities[tr.Veh1].Var48 |= flags48PassSignal
	return 0, true
}
