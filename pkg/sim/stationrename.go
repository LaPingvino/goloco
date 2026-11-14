package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// RenameStation is GameCommands::renameStation (RenameStation.cpp
// 0x00490756) with its three 12-character calls taken as one (the name, at
// most 36 bytes): unchanged, nothing; empty, applied, a new name is
// generated (generateNewStationName); else a unique user string is
// allocated (refused when none is free; the query frees it again) and,
// applied, the station takes it; the old name's string is freed either way
// it changes, and the station's label is recomputed (updateLabel: its frame
// sized for the new name). PARTIAL: the screen invalidation (goloco redraws).
func (m *Map) RenameStation(c *Cmd, flags uint8, id int, name string, om *objects.ObjectManager, g1 G1Raw) (int32, bool) {
	if len(name) > 36 {
		name = name[:36]
	}
	if id < 0 || id >= len(m.StationNames) {
		return 0, false
	}
	if m.StationName(id) == name {
		return 0, true
	}
	st := &m.StationNames[id]
	old := st.Name
	if name == "" {
		if flags&CmdApply == 0 {
			return 0, true
		}
		st.Name = m.generateNewStationName(st.Town, int32(st.X), int32(st.Y), int32(st.Z), 0, om)
	} else {
		nid := m.userStringAllocate(name, true)
		if nid == 0 {
			return 0, false
		}
		if flags&CmdApply == 0 {
			m.emptyUserString(nid)
			return 0, true
		}
		st.Name = nid
	}
	m.emptyUserString(old)
	m.UpdateStationLabel(id, g1)
	return 0, true
}
