package ui

import "testing"

// The vehicle list's filter types (af4d69c): all vehicles keeps everything;
// transports cargo looks at what the cars carry, has cargo order at the
// orders; without a cargo chosen neither filters.
func TestVehicleListFilterModes(t *testing.T) {
	v := VehicleRow{Carried: []int{3}, Cargos: []int{5}}
	for _, c := range []struct {
		mode, cargo int
		want        bool
	}{
		{0, 3, true}, {1, -1, true}, {1, 3, true}, {1, 5, false}, {2, 5, true}, {2, 3, false},
	} {
		if got := vlKeeps(v, c.mode, c.cargo); got != c.want {
			t.Errorf("mode %d cargo %d: %v, want %v", c.mode, c.cargo, got, c.want)
		}
	}
}

// VehicleList's onResize: at 550 wide the name column ends at 318 (its right
// edge set to max(310, 550 − 232)), the others follow one past it — as the
// real OpenLoco draws the headers (Monthly Profit at 319, Age at 420,
// Reliability at 486 to 546); wider, the name column takes the rest. The
// filter boxes keep to the bottom right: cargo 380–530, filter 258–378 at
// 550 × 213.
func TestVehicleListColumns(t *testing.T) {
	w := NewVehicleList(800, 0, nil, 0, 1, nil)
	w.PrepareDraw()
	want := [][2]int{{4, 318}, {319, 419}, {420, 485}, {486, 546}}
	for i, c := range w.Columns {
		if c.HeaderX != want[i][0] || c.HeaderX+c.HeaderW-1 != want[i][1] || c.TextX != want[i][0]-3 || c.TextW != c.HeaderW-2 {
			t.Errorf("550: column %d at %d…%d, text %d+%d", i, c.HeaderX, c.HeaderX+c.HeaderW-1, c.TextX, c.TextW)
		}
	}
	f, cg, shown := vlBoxes(w)
	if !shown || f != [4]int{258, 199, 121, 12} || cg != [4]int{380, 199, 151, 12} {
		t.Errorf("filter boxes %v %v", f, cg)
	}
	w.W = 800
	w.PrepareDraw()
	if c := w.Columns[0]; c.HeaderX+c.HeaderW-1 != 568 || w.Columns[3].HeaderX+w.Columns[3].HeaderW-1 != 796 {
		t.Errorf("800: name to %d, reliability to %d", c.HeaderX+c.HeaderW-1, w.Columns[3].HeaderX+w.Columns[3].HeaderW-1)
	}
	w.W = 300 // narrow: every edge within width − 4; the filters hidden
	w.PrepareDraw()
	if c := w.Columns[3]; c.HeaderX != 296 || c.HeaderW != 1 {
		t.Errorf("300: reliability at %d width %d", c.HeaderX, c.HeaderW)
	}
	if _, _, shown := vlBoxes(w); shown {
		t.Error("300: the filters shown")
	}
}
