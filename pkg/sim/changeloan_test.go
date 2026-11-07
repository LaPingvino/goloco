package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// changeLoan: pay back with the cash for it, not past the cash; borrow up to
// the inflation-adjusted maximum; cash moves by the difference.
func TestChangeLoan(t *testing.T) {
	m := &Map{Companies: make([]scenario.CompanyEconomy, 1), MaxLoanSize: 100}
	m.Companies[0].Cash, m.Companies[0].CurrentLoan = 5000, 20000
	c := &Cmd{Company: 0}
	if _, ok := m.ChangeLoan(c, CmdApply, 19000); !ok || m.Companies[0].CurrentLoan != 19000 || m.Companies[0].Cash != 4000 {
		t.Fatalf("pay back 1000: %+v", m.Companies[0])
	}
	if _, ok := m.ChangeLoan(c, CmdApply, 0); ok || c.ErrorText != i18n.NotEnoughCashAvailable {
		t.Errorf("paid back past the cash (%d)", c.ErrorText)
	}
	max := m.InflationAdjustedCost(100, 0, 8)
	if _, ok := m.ChangeLoan(c, CmdApply, max+1); ok {
		t.Errorf("borrowed past the maximum %d", max)
	}
	if _, ok := m.ChangeLoan(c, CmdApply, -1); ok {
		t.Error("a negative loan")
	}
	if _, ok := m.ChangeLoan(c, 0, 18500); !ok || m.Companies[0].CurrentLoan != 19000 { // a check (no apply): allowed, nothing changed
		t.Error("a check changed the loan")
	}
}
