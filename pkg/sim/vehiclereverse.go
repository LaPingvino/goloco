package sim

import "github.com/LaPingvino/goloco/pkg/i18n"

// VehicleReverse is the vehicleReverse game command (0x004ADAA8): refused
// for another company's vehicle, one that can't be modified, or one not on
// the map (the error text empty); applied, the train turns round where it
// stands (sub_4ADB47 with the head's var_52 at 1 meanwhile), the command's
// position the head's. The caller pays under TrainRunningCosts (cost 0).
func (m *Map) VehicleReverse(c *Cmd, flags uint8, headID uint16) (int32, bool) {
	tr, err := m.TrainOf(headID)
	if err != nil {
		return 0, false
	}
	head := &m.Entities[tr.Head]
	if !checkCompanyCompatibility(c, head.Owner) || !m.canBeModified(tr) {
		return 0, false
	}
	if head.TileX == -1 {
		c.ErrorText = i18n.Empty
		return 0, false
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	backup := head.HeadVar52
	head.HeadVar52 = 1
	m.Sub4ADB47(tr, false, c.Objects)
	m.Entities[tr.Head].HeadVar52 = backup
	h := m.Entities[tr.Head]
	c.Pos = &[3]int32{int32(h.PosX), int32(h.PosY), int32(h.PosZ)}
	return 0, true
}
