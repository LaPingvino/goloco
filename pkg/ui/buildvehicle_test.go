package ui

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/sim"
)

// TestBuildVehicleLayout: the window at its opening size (500 × 400) with a
// disabled tab and two track tabs, against prepareDraw's arithmetic.
func TestBuildVehicleLayout(t *testing.T) {
	if w, h := BuildVehicleSize(500, 400, 2); w != 500 || h != 400 {
		t.Errorf("size %d×%d", w, h)
	}
	if w, h := BuildVehicleSize(100, 100, 8); w != 443 || h != 233 {
		t.Errorf("minimum with 8 tabs: %d×%d, want 443×233", w, h)
	}
	l := BuildVehicleLayout(500, 400, 1<<4, 2) // aircraft disabled
	check := func(name string, got BVWidget, L, T, R, B int) {
		if got.L != L || got.T != T || got.R != R || got.B != B {
			t.Errorf("%s: %+v, want %d,%d–%d,%d", name, got, L, T, R, B)
		}
	}
	check("list", l[bvList], 3, 102, 313, 386)
	check("preview", l[bvPreview], 316, 44, 496, 109)
	check("close", l[bvClose], 485, 2, 497, 14)
	check("search", l[bvSearchBox], 4, 72, 271, 85)
	check("clear", l[bvSearchClear], 273, 72, 313, 85)
	// the list is 311 wide: thirds of 103
	check("sort", l[bvSortLabel], 210, 87, 312, 98)
	check("cargo", l[bvCargoLabel], 107, 87, 209, 98)
	check("filter", l[bvFilterLabel], 3, 87, 106, 98)
	check("filter ▼", l[bvFilterDropdown], 94, 88, 105, 97)
	// tabs: trains 3–33, buses 34–64, trucks 65–95, trams 96–126, aircraft hidden, ships 127–157
	check("trains tab", l[bvTabTrains], 3, 15, 33, 41)
	check("ships tab", l[bvTabTrains+5], 127, 15, 157, 41)
	if !l[bvTabTrains+4].Hidden || l[bvTabTrack0+1].Hidden || !l[bvTabTrack0+2].Hidden {
		t.Error("hidden tabs wrong")
	}
}

// The wheel over the list scrolls it 17 px a notch, within its content; off
// the list it isn't taken.
func TestBuildVehicleWheel(t *testing.T) {
	w := &BuildVehicleWindow{X: 0, Y: 0, W: 500, H: 400}
	for i := 0; i < 40; i++ {
		w.Rows = append(w.Rows, BVRow{})
	}
	list := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)[bvList]
	mx, my := (list.L+list.R)/2, (list.T+list.B)/2
	if !w.HandleWheel(mx, my, -1) || w.Scroll != 17 {
		t.Fatalf("one notch down: %d", w.Scroll)
	}
	w.HandleWheel(mx, my, 1)
	w.HandleWheel(mx, my, 1)
	if w.Scroll != 0 {
		t.Errorf("up past the top: %d", w.Scroll)
	}
	for i := 0; i < 500; i++ {
		w.HandleWheel(mx, my, -1)
	}
	if want := 40*sim.BuildVehicleRowHeight[0] - (list.B - list.T - 1); w.Scroll != want {
		t.Errorf("bottom: %d, want %d", w.Scroll, want)
	}
	if w.HandleWheel(list.L-5, 2, -1) {
		t.Error("took the wheel off the list")
	}
}
