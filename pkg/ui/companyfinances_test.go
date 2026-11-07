package ui

import "testing"

// Finances::widgets after prepareDraw at 636 × 319: the company select at the
// right, the scroll view and the loan stepper where makeWidgets puts them.
func TestCompanyFinancesLayout(t *testing.T) {
	lw := CompanyFinancesLayout(636, 319)
	if s := lw[cwCompanySelect]; s.L != 608 || s.R != 633 {
		t.Errorf("company select %d–%d", s.L, s.R)
	}
	if s := lw[cfScroll]; s.L != 133 || s.T != 45 || s.R != 631 || s.B != 259 {
		t.Errorf("scroll %+v", s)
	}
	if s := lw[cfCurrentLoan]; s.L != 87 || s.T != 264 || s.R != 186 {
		t.Errorf("loan %+v", s)
	}
	if s := lw[cwClose]; s.L != 621 || s.R != 633 {
		t.Errorf("close %+v", s)
	}
}
