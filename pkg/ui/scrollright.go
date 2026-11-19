package ui

import (
	"github.com/LaPingvino/goloco/pkg/sim"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
)

// ScrollArea is one scroll view widget as the right-button drag sees it
// (MouseInput.cpp scrollDragBegin / stateScrollRight): the widget's rectangle
// on the screen (W, H: Widget::width() / height()), what getScrollSize gives
// (the area's contentWidth / contentHeight are one more), the bars the widget
// has (its Scrollbars content: hscrollbarVisible / vscrollbarVisible — set
// whether the content overflows or not) and its offsets.
type ScrollArea struct {
	X, Y, W, H       int
	ScrollW, ScrollH int
	HBar, VBar       bool
	OffX, OffY       *int
}

// ScrollAreas is a window with scroll views: its scroll areas in widget order
// (the index is Window::getScrollDataIndex's).
type ScrollAreas interface{ ScrollAreas() []ScrollArea }

// Contains reports a screen point on the scroll view widget, its bars too
// (findWidgetAt).
func (a ScrollArea) Contains(x, y int) bool {
	return x >= a.X && y >= a.Y && x < a.X+a.W && y < a.Y+a.H
}

// DragFollow is stateScrollRight's step: horizontalDragFollow by dx, then
// verticalDragFollow by dy.
func (a ScrollArea) DragFollow(dx, dy int) {
	if a.OffX != nil {
		*a.OffX = L.DragFollow(a.HBar, a.VBar, a.W, a.ScrollW+1, *a.OffX, dx)
	}
	if a.OffY != nil {
		*a.OffY = L.DragFollow(a.VBar, a.HBar, a.H, a.ScrollH+1, *a.OffY, dy)
	}
}

// ScrollAreaAt is the scroll view widget of a window at a screen point: its
// index, or −1 (stateNormalRight's WidgetType::scrollview case).
func ScrollAreaAt(w Win, x, y int) int {
	s, ok := w.(ScrollAreas)
	if !ok {
		return -1
	}
	for i, a := range s.ScrollAreas() {
		if a.Contains(x, y) {
			return i
		}
	}
	return -1
}

// vArea is a scroll view with a vertical bar only (Scrollbars::vertical) at
// window-relative r (x, y, width, height) of a window at wx, wy.
func vArea(wx, wy int, r [4]int, scrollH int, off *int) ScrollArea {
	return ScrollArea{X: wx + r[0], Y: wy + r[1], W: r[2], H: r[3], ScrollH: scrollH, VBar: true, OffY: off}
}

// bvArea is a BVWidget's rectangle as vArea's r.
func bvArea(b BVWidget) [4]int { return [4]int{b.L, b.T, b.R - b.L + 1, b.B - b.T + 1} }

// The windows' scroll views (each window's widget table's ScrollView entries;
// getScrollSize as the window's own drawing uses it).

// ScrollAreas: TownList / IndustryList / StationList / VehicleList's scrollview.
func (w *TownListWindow) ScrollAreas() []ScrollArea {
	return []ScrollArea{vArea(w.X, w.Y, w.List, len(w.Rows)*w.RowHeight, &w.Scroll)}
}

// ScrollAreas: MessageWindow's messages tab (the settings tab has none).
func (w *MessagesWindow) ScrollAreas() []ScrollArea {
	if w.Tab != 0 {
		return nil
	}
	return []ScrollArea{vArea(w.X, w.Y, w.list(), w.contentHeight(), &w.scroll)}
}

// ScrollAreas: MusicSelection's list.
func (w *MusicSelectionWindow) ScrollAreas() []ScrollArea {
	return []ScrollArea{vArea(w.X, w.Y, w.list(), w.contentHeight(), &w.scroll)}
}

// ScrollAreas: KeyboardShortcuts' list.
func (w *KeyboardShortcutsWindow) ScrollAreas() []ScrollArea {
	return []ScrollArea{vArea(w.X, w.Y, w.list(), w.contentHeight(), &w.scroll)}
}

// ScrollAreas: CompanyList's list tab (3,56 to width − 4, height − 14).
func (w *CompanyListWindow) ScrollAreas() []ScrollArea {
	if w.Tab != 0 {
		return nil
	}
	return []ScrollArea{vArea(w.X, w.Y, [4]int{3, 56, w.W - 4 - 3 + 1, w.H - 14 - 56 + 1}, len(w.Rows)*25, &w.Scroll)}
}

// ScrollAreas: BuildVehicle's vehicle selection and its preview (no bars:
// a drag there scrolls nothing).
func (w *BuildVehicleWindow) ScrollAreas() []ScrollArea {
	lw := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)
	list := vArea(w.X, w.Y, bvArea(lw[bvList]), len(w.Rows)*sim.BuildVehicleRowHeight[w.Tab], &w.Scroll)
	p := bvArea(lw[bvPreview])
	return []ScrollArea{list, {X: w.X + p[0], Y: w.Y + p[1], W: p[2], H: p[3]}}
}

// ScrollAreas: TerraForm's tree and wall lists.
func (w *TerraformWindow) ScrollAreas() []ScrollArea {
	if w.Tab != TerraformTabTrees && w.Tab != TerraformTabWalls {
		return nil
	}
	return []ScrollArea{vArea(w.X, w.Y, w.list(), w.contentHeight(), &w.Scroll)}
}

// ScrollAreas: the Station window's Cargo tab and its vehicle tabs.
// PARTIAL: the Cargo Ratings tab's scroll view (goloco draws no list there).
func (w *StationWindow) ScrollAreas() []ScrollArea {
	switch {
	case w.Tab == 1:
		l, t, wd, ht, contentH := w.cargoList()
		return []ScrollArea{vArea(w.X, w.Y, [4]int{l, t, wd, ht}, contentH, &w.cScroll)}
	case w.Tab >= 3 && w.Vehicles != nil:
		l, t, wd, ht, rowH := w.vehicleList()
		return []ScrollArea{vArea(w.X, w.Y, [4]int{l, t, wd, ht}, len(w.Vehicles(w.Tab-3))*rowH, &w.vScroll)}
	}
	return nil
}

// ScrollAreas: the Vehicle window's car list (Details) and cargo list.
func (w *VehicleWindow) ScrollAreas() []ScrollArea {
	list, content, ok := w.barList()
	if !ok || list.Hidden {
		return nil
	}
	return []ScrollArea{vArea(w.X, w.Y, bvArea(list), content, &w.Scroll)}
}

// ScrollAreas: the Company window's Finances tab (Scrollbars::horizontal;
// getScrollSize numExpenditureYears × the column width).
func (w *CompanyFinancesWindow) ScrollAreas() []ScrollArea {
	if w.Tab != CompanyTabFinances {
		return nil
	}
	sc := CompanyFinancesLayout(w.W, w.H)[cfScroll]
	return []ScrollArea{{X: w.X + sc.L, Y: w.Y + sc.T, W: sc.R - sc.L + 1, H: sc.B - sc.T + 1,
		ScrollW: w.financesContentW() - 1, HBar: true, OffX: &w.Scroll}}
}

// ScrollAreas: the map (both bars; kRenderedMapWidth × kRenderedMapHeight).
func (w *MapWindow) ScrollAreas() []ScrollArea {
	x, y, ww, hh := w.scrollView()
	return []ScrollArea{{X: w.X + x, Y: w.Y + y, W: ww, H: hh, ScrollW: sim.MapRenderedWidth, ScrollH: sim.MapRenderedWidth,
		HBar: true, VBar: true, OffX: &w.ScrollX, OffY: &w.ScrollY}}
}
