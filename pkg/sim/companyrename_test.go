package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/i18n"
)

// changeCompanyName: renamed (a user string), a rival's name refused, the
// check alone changing nothing.
func TestRenameCompany(t *testing.T) {
	m := &Map{}
	m.ensureCompanySlots()
	m.CompanyNames[0], m.CompanyNames[1] = stringIDNewCompany, stringIDNewCompany
	c := &Cmd{Company: 0}
	if _, ok := m.RenameCompany(c, CmdApply, 0, "Cloudburst Railways"); !ok || m.CompanyName(0) != "Cloudburst Railways" {
		t.Fatalf("renamed: %q", m.CompanyName(0))
	}
	if _, ok := m.RenameCompany(c, CmdApply, 1, "Cloudburst Railways"); ok || c.ErrorText != i18n.ChosenNameInUse {
		t.Errorf("a rival's name taken (%d)", c.ErrorText)
	}
	if _, ok := m.RenameCompany(c, 0, 0, "Other"); !ok || m.CompanyName(0) != "Cloudburst Railways" {
		t.Errorf("a check renamed: %q", m.CompanyName(0))
	}
}
