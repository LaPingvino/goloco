package ui

// The Vehicle window (Ui/Windows/Vehicle.cpp), item 5P.g: its common widgets
// (makeCommonWidgets) and the Details tab's (Details::widgets, prepareDraw).
// Drawing and input follow.

// Common::widx and Details::widx.
const (
	vwFrame = iota
	vwCaption
	vwClose
	vwPanel
	vwTabMain
	vwTabDetails
	vwTabCargo
	vwTabFinances
	vwTabRoute
	vdBuildNew // Details::widx from here
	vdPickup
	vdRemove
	vdPaintBrush
	vdPaintPrimary
	vdPaintSecondary
	vdCarList
	vdWidgets
)

// Vehicle window sizes (Main::kWindowSize, Details' min / max).
var (
	VehicleWindowSize          = [2]int{265, 177}
	VehicleDetailsMinSize      = [2]int{192, 166}
	VehicleDetailsMinSizePaint = [2]int{192, 182} // the controlling company's own vehicle (onResize)
	VehicleDetailsMaxSize      = [2]int{400, 440}
)

// Details' text block under the car list: kVehicleDetailsOffset 2 + three
// lines of 12.
const vehicleDetailsTextHeight = 2 + 12*3

// vehicleCommonWidgets is makeCommonWidgets(265, 177, …).
func vehicleCommonWidgets(w *[vdWidgets]BVWidget) {
	w[vwFrame] = bvW(0, 0, 265, 177)
	w[vwCaption] = bvW(1, 1, 263, 13)
	w[vwClose] = bvW(250, 2, 13, 13)
	w[vwPanel] = bvW(0, 41, 265, 136)
	for i, x := range []int{3, 34, 65, 96, 158} {
		w[vwTabMain+i] = bvW(x, 15, 31, 27)
	}
}

// VehicleDetailsState is what Details::prepareDraw reads from the vehicle.
type VehicleDetailsState struct {
	Own        bool // head->owner == the controlling company
	Placed     bool // head->isPlaced()
	NoCars     bool // the train has no cars (pickup disabled)
	PaintTool  bool // isPaintToolActive
	HiddenTabs uint8
}

// VehicleDetailsLayout is Details::widgets after prepareDraw for a window of
// width × height: the frame, caption, close and panel to the size, the tabs
// left-aligned (leftAlignTabs: 31 wide, side by side), the car list to the
// right bar and above the text lines, the right-bar buttons (buildNew, pickup,
// remove, paintBrush at the bottom), the paint colours shown only with the
// paint tool; another company's vehicle has no buttons and a list to the
// right edge. The second value is disabledWidgets' pickup bit.
func VehicleDetailsLayout(width, height int, s VehicleDetailsState) (w [vdWidgets]BVWidget, pickupDisabled bool) {
	vehicleCommonWidgets(&w)
	w[vdBuildNew] = bvW(240, 44, 24, 24)
	w[vdPickup] = bvW(240, 68, 24, 24)
	w[vdRemove] = bvW(240, 96, 24, 24)
	w[vdPaintBrush] = bvW(240, 122, 24, 24)
	w[vdPaintPrimary] = bvW(240, 150, 16, 16)
	w[vdPaintSecondary] = bvW(258, 150, 16, 16)
	w[vdCarList] = bvW(3, 44, 237, 110)

	w[vwFrame].R, w[vwFrame].B = width-1, height-1
	w[vwPanel].R, w[vwPanel].B = width-1, height-1
	w[vwCaption].R = width - 2
	w[vwClose].L, w[vwClose].R = width-15, width-3
	x := w[vwTabMain].L // Widget::leftAlignTabs
	for i := vwTabMain; i <= vwTabRoute; i++ {
		if s.HiddenTabs&(1<<(i-vwTabMain)) != 0 {
			w[i].Hidden = true
			continue
		}
		w[i].L, w[i].R = x, x+30
		x += 31
	}
	w[vdCarList].R = width - 26
	w[vdCarList].B = height - vehicleDetailsTextHeight
	for _, i := range []int{vdBuildNew, vdPickup, vdRemove, vdPaintBrush} { // alignToRightBar
		w[i].L, w[i].R = width-25, width-2
	}
	if s.PaintTool {
		w[vdPaintPrimary] = BVWidget{width - 39, height - 33, width - 23, height - 17, false}
		w[vdPaintSecondary] = BVWidget{width - 21, height - 33, width - 5, height - 17, false}
	} else {
		w[vdPaintPrimary].Hidden = true
		w[vdPaintSecondary].Hidden = true
	}
	if s.Placed { // "Differs to main tab! Unsure why." (upstream)
		w[vdPickup].Hidden = true
	}
	if !s.Own {
		w[vdPaintPrimary].Hidden, w[vdPaintSecondary].Hidden = true, true
		w[vdBuildNew].Hidden, w[vdPickup].Hidden, w[vdRemove].Hidden = true, true, true
		w[vdCarList].R = width - 4
		w[vdPaintBrush].Hidden = true
	} else {
		w[vdBuildNew].Hidden, w[vdPickup].Hidden, w[vdRemove].Hidden = false, false, false
		w[vdPaintBrush].B = height - vehicleDetailsTextHeight
		w[vdPaintBrush].T = height - vehicleDetailsTextHeight - 24
	}
	return w, s.NoCars
}
