package sim

// VehicleRename is GameCommands/Vehicles/RenameVehicle.cpp (0x004B6572) as the
// window's three calls (i = 1, 2, 0) leave it: the name (at most 36 bytes, the
// three 12-byte parts) given whole. The same name as now: nothing. Empty: the
// type's numbered name (train_number …). Else a unique user string — refused
// when it can't be allocated (taken, or the table full); without apply it is
// freed again. The old user string is emptied.
func (m *Map) VehicleRename(c *Cmd, flags uint8, headID uint16, name string) (int32, bool) {
	tr, err := m.TrainOf(headID)
	if err != nil {
		return 0, false
	}
	head := &m.Entities[tr.Head]
	if len(name) > 36 {
		name = name[:36]
	}
	if m.VehicleName(head.Name, int(head.OrdinalNumber)) == name {
		return 0, true
	}
	var id uint16
	if name != "" {
		if id = m.userStringAllocate(name, true); id == 0 {
			return 0, false
		}
		if flags&CmdApply == 0 {
			m.emptyUserString(id)
			return 0, true
		}
	} else {
		if flags&CmdApply == 0 {
			return 0, true
		}
		id = defaultVehicleNames[min(int(head.VehicleType), 5)]
	}
	old := head.Name
	head.Name = id
	m.emptyUserString(old)
	return 0, true
}

// defaultVehicleNames is defaultVehicleStringIdMap: train_number … ship_number
// by vehicle type.
var defaultVehicleNames = [6]uint16{4, 5, 6, 7, 8, 9}

// emptyUserString is StringManager::emptyUserString: a user string's text
// cleared (any other id: nothing).
func (m *Map) emptyUserString(id uint16) {
	if id >= kUserStringsStart && int(id-kUserStringsStart) < len(m.UserStrings) {
		m.UserStrings[id-kUserStringsStart] = ""
	}
}
