package ui

// BuildVehicle — the Build Vehicle window (Ui/Windows/BuildVehicle.cpp, item
// 5P.f): its widgets (_widgets 0x5231D0) laid out for a window size as
// prepareDraw and onResize do. The list it shows is sim.BuildableVehicles,
// its tabs sim.BuildVehicleTrackTabs / BuildVehicleDisabledTabs.
// PARTIAL: drawing and input come next (5P.f part 4).

// Build Vehicle widget indices (widx).
const (
	bvFrame = iota
	bvCaption
	bvClose
	bvPanel
	bvSearchBox
	bvSearchClear
	bvFilterLabel
	bvFilterDropdown
	bvCargoLabel
	bvCargoDropdown
	bvSortLabel
	bvSortDropdown
	bvTabTrains // … bvTabTrains+5: buses, trucks, trams, aircraft, ships
	bvTabTrack0 = bvTabTrains + 6
	bvList      = bvTabTrack0 + 8
	bvPreview   = bvList + 1
	bvWidgets   = bvPreview + 1
)

// BVWidget is a widget's rectangle (inclusive edges) and whether it is hidden.
type BVWidget struct {
	L, T, R, B int
	Hidden     bool
}

func bvW(x, y, w, h int) BVWidget { return BVWidget{x, y, x + w - 1, y + h - 1, false} }

// buildVehicleWidgets is _widgets as declared (a dropdown is its combo box and
// an 11×10 button 12 px from its right).
func buildVehicleWidgets() [bvWidgets]BVWidget {
	var w [bvWidgets]BVWidget
	w[bvFrame] = bvW(0, 0, 380, 233)
	w[bvCaption] = bvW(1, 1, 378, 13)
	w[bvClose] = bvW(365, 2, 13, 13)
	w[bvPanel] = bvW(0, 41, 380, 192)
	w[bvSearchBox] = bvW(4, 72, 246, 14)
	w[bvSearchClear] = bvW(50, 72, 38, 14)
	dd := func(label, button, x, y int) {
		w[label] = bvW(x, y, 90, 12)
		w[button] = bvW(x+90-12, y+1, 11, 10)
	}
	dd(bvFilterLabel, bvFilterDropdown, 3, 87)
	dd(bvCargoLabel, bvCargoDropdown, 48, 87)
	dd(bvSortLabel, bvSortDropdown, 93, 87)
	for i := range 6 {
		w[bvTabTrains+i] = bvW(3, 15, 31, 27)
	}
	for i := range 8 {
		w[bvTabTrack0+i] = bvW(5+31*i, 43, 31, 27)
	}
	w[bvList] = bvW(3, 102, 374, 146)
	w[bvPreview] = bvW(250, 44, 180, 66)
	return w
}

// BuildVehicleSize is onResize's size: at least max(tabs·31 + 195, 380) × 233,
// at most 700 × 600.
func BuildVehicleSize(width, height, numTrackTabs int) (int, int) {
	minW := max(numTrackTabs*31+195, 380)
	return min(max(width, minW), 700), min(max(height, 233), 600)
}

// BuildVehicleLayout is the widgets for a window of width × height as
// prepareDraw places them: the right-hand preview, the list's right and
// bottom, the search box and its clear button, the three filter dropdowns in
// thirds of the list's width, the vehicle-type tabs left-aligned past the
// disabled ones (leftAlignTabs, 30 px), the track-type tabs past the last
// one hidden (setTrackTypeTabs).
func BuildVehicleLayout(width, height int, disabledTabs uint8, numTrackTabs int) [bvWidgets]BVWidget {
	w := buildVehicleWidgets()
	w[bvFrame].R, w[bvFrame].B = width-1, height-1
	w[bvPanel].R, w[bvPanel].B = width-1, height-1
	w[bvCaption].R = width - 2
	w[bvClose].L, w[bvClose].R = width-15, width-3
	w[bvPreview].R, w[bvPreview].L = width-4, width-184
	list := &w[bvList]
	list.R, list.B = width-187, height-14
	listW := list.R - list.L + 1 // Widget::width()
	w[bvSearchClear].R, w[bvSearchClear].L = list.R, list.R-40
	w[bvSearchBox].R = list.R - 42
	w[bvSortLabel].R, w[bvSortLabel].L = list.R-1, list.R-listW/3
	w[bvSortDropdown].R, w[bvSortDropdown].L = list.R-2, list.R-12
	sortL := w[bvSortLabel].L
	w[bvCargoLabel].R, w[bvCargoLabel].L = sortL-1, sortL-listW/3
	w[bvCargoDropdown].R, w[bvCargoDropdown].L = sortL-2, sortL-12
	w[bvFilterLabel].R = w[bvCargoLabel].L - 1
	w[bvFilterDropdown].R = w[bvCargoLabel].L - 2
	w[bvFilterDropdown].L = w[bvFilterDropdown].R - 11
	x := w[bvTabTrains].L // leftAlignTabs
	for i := range 6 {
		t := &w[bvTabTrains+i]
		if disabledTabs&(1<<i) != 0 {
			t.Hidden = true
			continue
		}
		t.L, t.R = x, x+30
		x = t.R + 1
	}
	for i := numTrackTabs; i < 8; i++ {
		w[bvTabTrack0+i].Hidden = true
	}
	return w
}
