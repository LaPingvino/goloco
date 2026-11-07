package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// ExpenditureMiscellaneous is ExpenditureType::Miscellaneous.
const ExpenditureMiscellaneous = 16

// ColourSchemeChange is ChangeCompanyColourSchemeArgs: a colour (value) for
// the main scheme (ColourType 0) or a vehicle type's (1–10), primary or
// secondary — or, with SetMode, that vehicle type's own scheme on (value 1)
// or off.
type ColourSchemeChange struct {
	Company    uint8
	ColourType uint8
	Value      uint8
	IsPrimary  bool
	SetMode    bool
}

// ChangeCompanyColourScheme is GameCommands::changeCompanyColour
// (ChangeCompanyColour.cpp 0x0043483D). Checked, only the main primary colour
// can be refused: one a competing company has. Applied: the scheme changed,
// then updateVehicleColours, CompanyManager::updateColours. It costs nothing.
// PARTIAL: updateHeadquartersColour (no headquarters are built yet).
func (m *Map) ChangeCompanyColourScheme(c *Cmd, flags uint8, a ColourSchemeChange, om *objects.ObjectManager) (int32, bool) {
	c.Pos = &[3]int32{-32768, 0, 0} // setPosition(0x8000): no money effect
	if int(a.Company) >= len(m.CompanyColourSchemes) {
		return 0, false
	}
	cs := &m.CompanyColourSchemes[a.Company]
	if flags&CmdApply == 0 {
		if !checkCompanyCompatibility(c, a.Company) {
			return 0, false
		}
		if a.SetMode || a.ColourType > 0 || !a.IsPrimary {
			return 0, true
		}
		if m.competingColourMask(int(a.Company))&(1<<a.Value) != 0 {
			c.ErrorText = 0 // StringIds::empty
			return 0, false
		}
		return 0, true
	}
	if a.SetMode {
		if a.Value != 0 {
			cs.CustomSet |= 1 << a.ColourType
		} else {
			cs.CustomSet &^= 1 << a.ColourType
		}
	} else {
		colours := &cs.Main
		if a.ColourType > 0 && int(a.ColourType) <= len(cs.Vehicle) {
			colours = &cs.Vehicle[a.ColourType-1]
		}
		if a.IsPrimary {
			colours[0] = a.Value
		} else {
			colours[1] = a.Value
		}
	}
	m.updateVehicleColours(int(a.Company), om)
	m.updateCompanyColours()
	return 0, true
}

// updateVehicleColours is Company::updateVehicleColours: every car of the
// company's vehicles takes its scheme — the main one, or its vehicle
// object's colour type's own when that is set.
func (m *Map) updateVehicleColours(company int, om *objects.ObjectManager) {
	if om == nil {
		return
	}
	trains, _ := m.Trains()
	for _, tr := range trains {
		if int(m.Entities[tr.Head].Owner) != company {
			continue
		}
		for _, car := range tr.Cars {
			if len(car) == 0 {
				continue
			}
			body := &m.Entities[car[0].Body]
			if int(body.ObjectID) >= len(om.Vehicles) || om.Vehicles[body.ObjectID] == nil {
				continue
			}
			scheme := m.vehicleColourScheme(uint8(company), om.Vehicles[body.ObjectID].ColourType)
			for _, comp := range car { // applyToComponents: front, back, body
				for _, id := range []int{comp.Front, comp.Back, comp.Body} {
					m.Entities[id].ColourPrimary, m.Entities[id].ColourSecondary = scheme[0], scheme[1]
				}
			}
		}
	}
}

// updateCompanyColours is CompanyManager::updateColours: each company's
// colour is its main primary.
func (m *Map) updateCompanyColours() {
	for i := range m.CompanyColourSchemes {
		if i < len(m.CompanyColours) {
			m.CompanyColours[i] = m.CompanyColourSchemes[i].Main[0]
		}
	}
	if int(companyIDNeutral) < len(m.CompanyColours) {
		m.CompanyColours[companyIDNeutral] = colourGrey
	}
}
