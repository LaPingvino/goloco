package ui

import "testing"

// Details::prepareDraw at the default 265×177: the route tab moves next to
// finances (leftAlignTabs), the right bar's buttons, the list above three
// text lines; another company's vehicle: no buttons, the list to the edge;
// the pickup button stays for an own placed vehicle (upstream un-hides it).
func TestVehicleDetailsLayout(t *testing.T) {
	w, _ := VehicleDetailsLayout(265, 177, VehicleDetailsState{Own: true, Placed: true})
	want := map[int]BVWidget{
		vwTabRoute:   {127, 15, 157, 41, false},
		vwClose:      {250, 2, 262, 14, false},
		vdCarList:    {3, 44, 239, 139, false},
		vdBuildNew:   {240, 44, 263, 67, false},
		vdPickup:     {240, 68, 263, 91, false},
		vdPaintBrush: {240, 115, 263, 139, false},
	}
	for i, x := range want {
		if w[i] != x {
			t.Errorf("widget %d: %+v, want %+v", i, w[i], x)
		}
	}
	if !w[vdPaintPrimary].Hidden {
		t.Error("paint colours shown without the paint tool")
	}
	o, _ := VehicleDetailsLayout(300, 200, VehicleDetailsState{})
	if !o[vdBuildNew].Hidden || !o[vdRemove].Hidden || !o[vdPaintBrush].Hidden || o[vdCarList].R != 296 || o[vdCarList].B != 162 {
		t.Errorf("another company's: list %+v, buttons hidden %v %v %v", o[vdCarList], o[vdBuildNew].Hidden, o[vdRemove].Hidden, o[vdPaintBrush].Hidden)
	}
	p, _ := VehicleDetailsLayout(265, 182, VehicleDetailsState{Own: true, PaintTool: true})
	if p[vdPaintPrimary] != (BVWidget{226, 149, 242, 165, false}) || p[vdPaintSecondary] != (BVWidget{244, 149, 260, 165, false}) {
		t.Errorf("paint colours %+v %+v", p[vdPaintPrimary], p[vdPaintSecondary])
	}
}

// The caption moves the window (statePositioningWindow: the pointer's y
// clamped to 29 … height − 29) and is a rename click only when unmoved.
func TestVehicleWindowPositioning(t *testing.T) {
	renamed := 0
	w := &VehicleWindow{X: 100, Y: 100, W: 265, H: 177, OnRename: func() { renamed++ }}
	w.HandleClick(150, 105)
	if !w.Positioning {
		t.Fatal("a press on the caption doesn't start moving")
	}
	w.PositionMove(170, 125, 600)
	w.PositionEnd(200, 5, 600)                                    // y clamped to 29
	if w.X != 150 || w.Y != 24 || w.Positioning || renamed != 0 { // 100+20, then to y 29 from 125: −96
		t.Errorf("moved to %d,%d positioning %v renamed %d", w.X, w.Y, w.Positioning, renamed)
	}
	w.HandleClick(160, w.Y+5)
	w.PositionEnd(160, w.Y+5, 600)
	if renamed != 1 {
		t.Errorf("an unmoved caption click renames: %d", renamed)
	}
}

// Details' car list bar: a press on its down arrow scrolls (and starts no
// car drag); held, it repeats; released, it stops.
func TestVehicleCarListBar(t *testing.T) {
	w := &VehicleWindow{X: 100, Y: 100, W: 265, H: 177, Tab: 1, Cars: make([]TrainCarDraw, 20),
		State: VehicleDetailsState{Own: true}, OnDropMove: func(int, int) {}}
	lw, _ := VehicleDetailsLayout(w.W, w.H, w.State)
	list := lw[vdCarList]
	x, y := w.X+list.R-4, w.Y+list.B-4
	if !w.carBar(x, y) {
		t.Fatal("the down arrow is not on the bar")
	}
	w.HandleClick(x, y)
	if w.Scroll != 3 || w.Dragging {
		t.Fatalf("press: scroll %d dragging %v", w.Scroll, w.Dragging)
	}
	w.ScrollDrag(x, y, true)
	w.ScrollDrag(x, y, false)
	w.ScrollDrag(x, y, true)
	if w.Scroll != 6 {
		t.Fatalf("hold / release: %d", w.Scroll)
	}
}
