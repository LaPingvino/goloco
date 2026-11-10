package sim

// RenameTown is GameCommands::renameTown (RenameTown.cpp 0x00455029) with its
// three 12-character calls taken as one (the name, at most 36 bytes): an
// empty or unchanged name does nothing; else a unique user string is
// allocated (refused when none is free; the query frees it again) and,
// applied, the town takes it, its old string is freed, and the town's label
// and every station's (StationManager::updateLabels) are recomputed.
// PARTIAL: the screen invalidation (goloco redraws).
func (m *Map) RenameTown(c *Cmd, flags uint8, town int, name string, g1 G1Raw) (int32, bool) {
	if len(name) > 36 {
		name = name[:36]
	}
	if name == "" || town < 0 || town >= len(m.Towns) {
		return 0, true
	}
	t := &m.Towns[town]
	if m.TownName(t.Name, town) == name {
		return 0, true
	}
	id := m.userStringAllocate(name, true)
	if id == 0 {
		return 0, false
	}
	if flags&CmdApply == 0 {
		m.emptyUserString(id)
		return 0, true
	}
	old := t.Name
	t.Name = id
	m.emptyUserString(old)
	m.UpdateTownLabel(town, g1)
	for i := range m.StationNames {
		if m.StationNames[i].Name != 0 { // stations(): the live ones
			m.UpdateStationLabel(i, g1)
		}
	}
	return 0, true
}
