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

// Autopay: the controlling company's loan paid back daily as far as its cash
// goes; off, nothing.
func TestLoanAutorepay(t *testing.T) {
	m := &Map{Companies: make([]scenario.CompanyEconomy, 1), CompanyFaces: make([]scenario.CompanyFace, 1)}
	m.Companies[0].Cash, m.Companies[0].CurrentLoan = 7000, 20000
	m.updateLoanAutorepay(0)
	if m.Companies[0].CurrentLoan != 20000 {
		t.Fatal("repaid without autopay")
	}
	m.CompanyFaces[0].ChallengeFlags |= CompanyFlagAutopayLoan
	m.updateLoanAutorepay(0)
	if c := m.Companies[0]; c.CurrentLoan != 13000 || c.Cash != 0 {
		t.Errorf("after autopay: loan %d cash %d", c.CurrentLoan, c.Cash)
	}
}
