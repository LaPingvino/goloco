package ui

import (
	"strings"
	"testing"
)

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

// formatChallengeArguments' sentences: the objective with its rating name and
// the time limit, the cargo one with its unit name.
func TestChallengeObjectiveText(t *testing.T) {
	got := stripCodes(ChallengeObjectiveText(ChallengeObjective{Type: 2, Flags: 4, PerformanceIndex: 70, UntilYear: 1960}))
	if !strings.Contains(got, "70.0%") || !strings.Contains(got, "1960") {
		t.Errorf("performance: %q", got)
	}
	got = stripCodes(ChallengeObjectiveText(ChallengeObjective{Type: 3, CargoUnitPlural: "{COMMA32} passengers", CargoAmount: 5000, Flags: 1}))
	if !strings.Contains(got, "5,000 passengers") {
		t.Errorf("cargo: %q", got)
	}
}
