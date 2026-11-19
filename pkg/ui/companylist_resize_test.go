package ui

import "testing"

// The Company list resizes (CompanyList::open: resizable, 300 × 272 … 800
// × 940); a graph tab is its own size and grows from there; the speed
// records are fixed; back on the list tab, tabReset's setSize(640 × 272)
// clamps to the graph tab's bounds first (660 wide after Cargo distance).
func TestCompanyListResize(t *testing.T) {
	s := &WindowStack{ScreenW: 1024, ScreenH: 768}
	w := NewCompanyList(0, 1, nil)
	w.X, w.Y = 100, 100
	s.Open(w)
	if !w.CanResize() || !w.HasPanelResizeHandle() || w.W != 640 || w.H != 272 {
		t.Fatalf("company list %dx%d", w.W, w.H)
	}
	gx, gy := w.X+w.W-2, w.Y+w.H-2
	s.Press(gx, gy)
	s.Held(gx-400, gy+100)
	s.Release(gx-400, gy+100)
	if w.W != 300 || w.H != 372 {
		t.Errorf("resized to %dx%d, want 300x372", w.W, w.H)
	}
	if w.OnFrame(w.X+w.W-10, w.Y+w.H-16) || !w.OnFrame(w.X+w.W-2, w.Y+w.H-2) {
		t.Error("the scroll view blocks the grip, the corner doesn't")
	}
	w.switchTab(3) // cargo distance: 660 × 272, to 800 × 940
	if w.W != 660 || w.H != 272 || w.MinW != 660 || w.MaxW != 800 || !w.CanResize() {
		t.Errorf("cargo distance tab %dx%d, %d…%d", w.W, w.H, w.MinW, w.MaxW)
	}
	w.switchTab(0)
	if w.W != 660 || w.H != 272 || w.MinW != 300 {
		t.Errorf("back on the list %dx%d, min %d", w.W, w.H, w.MinW)
	}
	w.switchTab(6)
	if w.CanResize() || w.W != 495 || w.H != 169 {
		t.Errorf("speed records %dx%d resizable %v", w.W, w.H, w.CanResize())
	}
}

// The Company window resizes on its Status tab only (Status::onResize:
// 270 × 182 … 640 × 400); every tab change sets the tab's fixed size first.
func TestCompanyStatusResize(t *testing.T) {
	s := &WindowStack{ScreenW: 1024, ScreenH: 768}
	w := &CompanyFinancesWindow{WinBase: WinBase{WinType: WindowTypeCompany, X: 100, Y: 100, W: 636, H: 319}, Tab: CompanyTabFinances}
	w.switchTab(CompanyTabFinances)
	s.Open(w)
	if w.CanResize() {
		t.Error("the Finances tab resizes")
	}
	w.switchTab(CompanyTabStatus)
	if !w.CanResize() || !w.HasPanelResizeHandle() || w.W != 270 || w.H != 182 || w.MaxW != 640 || w.MaxH != 400 {
		t.Fatalf("status tab %dx%d, max %dx%d", w.W, w.H, w.MaxW, w.MaxH)
	}
	gx, gy := w.X+w.W-2, w.Y+w.H-2
	if !w.OnFrame(gx, gy) || w.OnFrame(w.X+w.W-16, w.Y+w.H-5) {
		t.Error("the status label blocks the grip where it reaches; the corner is free")
	}
	s.Press(gx, gy)
	s.Held(gx+400, gy+250)
	s.Release(gx+400, gy+250)
	if w.W != 640 || w.H != 400 {
		t.Errorf("resized to %dx%d, want the 640 × 400 limit", w.W, w.H)
	}
	w.switchTab(CompanyTabDetails)
	if w.CanResize() || w.W != CompanyDetailsSize[0] {
		t.Errorf("details tab %dx%d resizable %v", w.W, w.H, w.CanResize())
	}
}
