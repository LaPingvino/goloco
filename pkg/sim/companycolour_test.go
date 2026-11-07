package sim

import "testing"

// changeCompanyColour: the main primary colour taken by a competitor is
// refused; applied, the main colour and the company colour change; a vehicle
// type's own scheme switched on and given a colour.
func TestChangeCompanyColourScheme(t *testing.T) {
	m := &Map{}
	m.ensureCompanySlots()
	m.CompanyNames[0], m.CompanyNames[1] = 2, 2 // two live companies
	m.CompanyColourSchemes[0].Main = [2]uint8{5, 6}
	m.CompanyColourSchemes[1].Main = [2]uint8{9, 10}
	m.CompanyColours[0], m.CompanyColours[1] = 5, 9
	c := &Cmd{Company: 0}
	if mask := m.competingColourMask(0); mask&(1<<9) != 0 {
		if _, ok := m.ChangeCompanyColourScheme(c, 0, ColourSchemeChange{Company: 0, Value: 9, IsPrimary: true}, nil); ok {
			t.Error("took a competitor's colour")
		}
	}
	if _, ok := m.ChangeCompanyColourScheme(c, CmdApply, ColourSchemeChange{Company: 0, Value: 20, IsPrimary: true}, nil); !ok ||
		m.CompanyColourSchemes[0].Main[0] != 20 || m.CompanyColours[0] != 20 {
		t.Errorf("main primary: %+v colour %d", m.CompanyColourSchemes[0].Main, m.CompanyColours[0])
	}
	m.ChangeCompanyColourScheme(c, CmdApply, ColourSchemeChange{Company: 0, ColourType: 3, Value: 1, SetMode: true}, nil)
	m.ChangeCompanyColourScheme(c, CmdApply, ColourSchemeChange{Company: 0, ColourType: 3, Value: 14, IsPrimary: false}, nil)
	if cs := m.CompanyColourSchemes[0]; cs.CustomSet&(1<<3) == 0 || cs.Vehicle[2][1] != 14 {
		t.Errorf("vehicle type 3: %+v", cs)
	}
	if got := m.vehicleColourScheme(0, 3); got != m.CompanyColourSchemes[0].Vehicle[2] {
		t.Errorf("type 3 scheme %v", got)
	}
}
