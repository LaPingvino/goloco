package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// applyPaymentToCompany: cash −= payment; this month's expenditure of that
// type −= payment (income is a negative payment).
func TestApplyPaymentToCompany(t *testing.T) {
	m := Map{Companies: make([]scenario.CompanyEconomy, 2)}
	m.Companies[1].Cash = 1000
	m.ApplyPaymentToCompany(1, -250, 3)
	m.ApplyPaymentToCompany(1, 100, 4)
	if c := m.Companies[1]; c.Cash != 1150 || c.Expenditures[0][3] != 250 || c.Expenditures[0][4] != -100 {
		t.Errorf("cash %d, expenditures %d / %d", c.Cash, c.Expenditures[0][3], c.Expenditures[0][4])
	}
	m.ApplyPaymentToCompany(9, 100, 0) // no such company: nothing
}
