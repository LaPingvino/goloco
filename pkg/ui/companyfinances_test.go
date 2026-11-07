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

// Common::switchTab: a tab click switches and resizes, keeping the place.
func TestCompanyWindowSwitchTab(t *testing.T) {
	w := &CompanyFinancesWindow{X: 82, Y: 140, W: 636, H: 319, Tab: CompanyTabFinances, Data: CompanyFinances{Own: true}}
	w.HandleClick(100, 168) // the status tab (3–33, 15–41)
	if w.Tab != CompanyTabStatus || w.W != 270 || w.H != 182 || w.X != 82 {
		t.Errorf("status: tab %d %d×%d at %d", w.Tab, w.W, w.H, w.X)
	}
	w.HandleClick(82+3+31*4+5, 168) // cargo delivered
	if w.Tab != CompanyTabCargoDelivered || w.W != 240 {
		t.Errorf("cargo: tab %d %d×%d", w.Tab, w.W, w.H)
	}
}

// showColour: 31 colours (all but one) in getItemsPerRow(31) = 8 columns,
// 4 rows of 16 px; the selected one highlighted; a cell picks its colour.
func TestColourDropdown(t *testing.T) {
	d := NewColourDropdown(300, 200, 16, 0x7FFFFFFF, 5, 1, 800, 600)
	if len(d.Colours) != 31 || d.Columns != 8 || d.W != 8*16+4 || d.H != 4*16+4 || d.Highlighted != 5 {
		t.Fatalf("grid %d colours %d cols %d×%d hl %d", len(d.Colours), d.Columns, d.W, d.H, d.Highlighted)
	}
	if d.Y != 200+17 {
		t.Errorf("below the widget: y %d", d.Y)
	}
	var got uint8 = 255
	d.OnSelect = func(c uint8) { got = c }
	d.HandleClick(d.X+2+16*2+3, d.Y+2+16*1+3) // row 1, column 2: colour 10
	if got != 10 {
		t.Errorf("picked %d", got)
	}
}

// The caption moves the window; an unmoved click on it renames.
func TestCompanyWindowPositioning(t *testing.T) {
	renamed := 0
	w := &CompanyFinancesWindow{X: 82, Y: 140, W: 636, H: 319, Tab: CompanyTabFinances, OnRename: func() { renamed++ }}
	w.HandleClick(300, 146)
	w.PositionMove(320, 156, 600)
	w.PositionEnd(330, 160, 600)
	if w.X != 112 || w.Y != 154 || renamed != 0 {
		t.Errorf("moved to %d,%d renamed %d", w.X, w.Y, renamed)
	}
	w.HandleClick(300, w.Y+6)
	w.PositionEnd(300, w.Y+6, 600)
	if renamed != 1 {
		t.Errorf("unmoved click: renamed %d", renamed)
	}
}

// sortCompanyList's orders: performance (the default) and value highest
// first, name by strcmp; a header click sorts by it.
func TestCompanyListSort(t *testing.T) {
	w := &CompanyListWindow{W: 640, H: 272, SortMode: 2, Rows: []CompanyListRow{
		{ID: 0, Name: "Bravo", Performance: 50, Value: 900}, {ID: 1, Name: "Alpha", Performance: 70, Value: 100}}}
	w.Sort()
	if w.Rows[0].ID != 1 {
		t.Errorf("performance: %d first", w.Rows[0].ID)
	}
	w.HandleClick(w.X+534+5, w.Y+45) // the value header
	if w.SortMode != 3 || w.Rows[0].ID != 0 {
		t.Errorf("value: mode %d, %d first", w.SortMode, w.Rows[0].ID)
	}
	w.HandleClick(w.X+4+5, w.Y+45) // the name header
	if w.Rows[0].Name != "Alpha" {
		t.Errorf("name: %s first", w.Rows[0].Name)
	}
}

// The loan buttons repeat while held: a step on the press, then every 4th
// tick from 16, the step ×10 every 100 ticks; letting go stops it.
func TestLoanClickRepeat(t *testing.T) {
	var steps []int
	w := &CompanyFinancesWindow{X: 82, Y: 140, W: 636, H: 319, Tab: CompanyTabFinances, Data: CompanyFinances{Own: true},
		OnLoan: func(dir, step int) { steps = append(steps, dir*step) }}
	b := CompanyFinancesLayout(w.W, w.H)[cfLoanIncrease]
	x, y := w.X+(b.L+b.R)/2, w.Y+(b.T+b.B)/2
	w.HandleClick(x, y)
	for i := 0; i < 103; i++ { // ticks 2 … 104
		w.UpdateHeld(x, y, true)
	}
	// 1 (press); 16, 20 … 96: 21 × 1; 100, 104: 2 × 10
	if len(steps) != 1+21+2 || steps[0] != 1 || steps[1] != 1 || steps[22] != 10 || steps[23] != 10 {
		t.Fatalf("steps %v", steps)
	}
	w.UpdateHeld(x, y, false)
	w.UpdateHeld(x, y, true)
	if len(steps) != 24 {
		t.Fatalf("repeated after release: %v", steps)
	}
}

// The Finances bar: a held right button steps 3 each tick while on it, off
// it it stops and unpresses, back on it resumes; the release clears it.
func TestFinancesBarHold(t *testing.T) {
	w := &CompanyFinancesWindow{X: 82, Y: 140, W: 636, H: 319, Tab: CompanyTabFinances, Data: CompanyFinances{Own: true, NumYears: 6}}
	sc := CompanyFinancesLayout(w.W, w.H)[cfScroll]
	rx, ry := w.X+sc.R-5, w.Y+sc.B-5 // the right button
	w.HandleClick(rx, ry)
	if w.Scroll != 3 || !w.hPressed {
		t.Fatalf("press: scroll %d pressed %v", w.Scroll, w.hPressed)
	}
	w.UpdateDrag(rx, ry, true)
	w.UpdateDrag(rx, ry, true)
	if w.Scroll != 9 {
		t.Fatalf("held: %d", w.Scroll)
	}
	w.UpdateDrag(rx, ry-40, true) // off the bar
	if w.Scroll != 9 || w.hPressed {
		t.Fatalf("off the part: %d %v", w.Scroll, w.hPressed)
	}
	w.UpdateDrag(rx, ry, true)
	if w.Scroll != 12 || !w.hPressed {
		t.Fatalf("back on: %d %v", w.Scroll, w.hPressed)
	}
	w.UpdateDrag(rx, ry, false)
	w.UpdateDrag(rx, ry, true)
	if w.Scroll != 12 || w.hPart != "" {
		t.Fatalf("after release: %d %q", w.Scroll, w.hPart)
	}
}
