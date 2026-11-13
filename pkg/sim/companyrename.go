package sim

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
)

// RenameCompany is GameCommands::changeCompanyName (RenameCompanyName.cpp
// 0x00434914) with its three 12-character calls taken as one (the name, at
// most 36 bytes): an empty or unchanged name does nothing; one another
// company has is refused (chosen_name_in_use); else a user string is
// allocated (not necessarily unique), and applied the company takes it and
// its old name's string is freed.
func (m *Map) RenameCompany(c *Cmd, flags uint8, company int, name string) (int32, bool) {
	if len(name) > 36 {
		name = name[:36]
	}
	if name == "" || company < 0 || company >= len(m.CompanyNames) || m.CompanyName(company) == name {
		return 0, true
	}
	for i := range m.CompanyNames {
		if i != company && !m.CompanyEmpty(i) && m.CompanyName(i) == name {
			c.ErrorText = i18n.ChosenNameInUse
			return 0, false
		}
	}
	id := m.userStringAllocate(name, false)
	if id == 0 {
		return 0, false
	}
	if flags&CmdApply == 0 {
		m.emptyUserString(id)
		return 0, true
	}
	old := m.CompanyNames[company]
	m.CompanyNames[company] = id
	m.emptyUserString(old)
	return 0, true
}

// RenameCompanyOwner is GameCommands::changeCompanyOwnerName
// (RenameCompanyOwner.cpp 0x00434914) with its three 12-character calls taken
// as one (at most 36 bytes): an empty or unchanged name does nothing; one
// another company's owner has is refused (chosen_name_in_use); else a user
// string is allocated, and applied the owner takes it and the old string is
// freed. PARTIAL: tryLoadInstalledFaceForName (a competitor face whose name
// matches, loaded while the game runs — waits on loading objects at run time).
func (m *Map) RenameCompanyOwner(c *Cmd, flags uint8, company int, name string, om *objects.ObjectManager) (int32, bool) {
	if len(name) > 36 {
		name = name[:36]
	}
	if name == "" || company < 0 || company >= len(m.CompanyFaces) || m.OwnerName(company, om) == name {
		return 0, true
	}
	for i := range m.CompanyFaces {
		if i != company && !m.CompanyEmpty(i) && m.OwnerName(i, om) == name {
			c.ErrorText = i18n.ChosenNameInUse
			return 0, false
		}
	}
	id := m.userStringAllocate(name, false)
	if id == 0 {
		return 0, false
	}
	if flags&CmdApply == 0 {
		m.emptyUserString(id)
		return 0, true
	}
	old := m.CompanyFaces[company].OwnerName
	m.CompanyFaces[company].OwnerName = id
	m.emptyUserString(old)
	return 0, true
}
