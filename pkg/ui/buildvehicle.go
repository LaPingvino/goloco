package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// BuildVehicle — the Build Vehicle window (Ui/Windows/BuildVehicle.cpp, item
// 5P.f): its widgets (_widgets 0x5231D0) laid out for a window size as
// prepareDraw and onResize do. The list it shows is sim.BuildableVehicles,
// its tabs sim.BuildVehicleTrackTabs / BuildVehicleDisabledTabs.
// PARTIAL: drawing and input come next (5P.f part 4).

// PaletteIndex::mutedDarkRed3 / 5 (Colour.h): a locked vehicle's row.
const (
	paletteMutedDarkRed3 = 0x3D
	paletteMutedDarkRed5 = 0x3F
)

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

// BVRow is one vehicle of the list.
type BVRow struct {
	ID      uint16 // the vehicle object's id (rowInfo)
	Vehicle *objects.VehicleObject
	Name    string
	Locked  bool // !isVehicleIndexUnlocked && !Config::buildLockedVehicles
}

// BuildVehicleWindow is the Build Vehicle window: 500×400 at first, the
// lighter frame, the skin's titlebar colour and the player colour.
// PARTIAL (5P.f part 4): the tabs' icons (drawTransportTypeTabs /
// drawTrackTypeTabs), the search box and the three dropdowns' texts, the
// preview (drawVehicleOverview, the cargo icons), the cargo-filter line,
// the scroll bar, input, and createWindow's free-spot placement.
type BuildVehicleWindow struct {
	WinBase
	TitleColour, Colour uint8
	Tab, TrackTab       int
	Disabled            uint8
	NumTrackTabs        int
	Rows                []BVRow
	Hover               int // a row index, or -1
	Scroll              int
	hold                uilayout.VScrollHold // a press on the list\'s scrollbar held (stateScrollLeft)
	Primary, Secondary  uint8                // the controlling company's colours for the vehicles
	Message             string
	OnTab               func(int)       // a vehicle-type tab
	OnBuy               func(id uint16) // a row clicked (onScrollMouseDown): buy that vehicle
}

// Draw is draw + drawScroll(vehicle_selection).
func (w *BuildVehicleWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)
	rowH := sim.BuildVehicleRowHeight[w.Tab]
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	list := lw[bvList]
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[bvPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[bvClose].L, lw[bvClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for i := range 6 {
			if t := lw[bvTabTrains+i]; !t.Hidden {
				drawTab(rt, t.L, t.T, i == w.Tab, sec, tabNoIcon, nil)
			}
		}
		for i := range 8 {
			if t := lw[bvTabTrack0+i]; !t.Hidden {
				drawTab(rt, t.L, t.T, i == w.TrackTab, sec, tabNoIcon, nil)
			}
		}
		cx, cy, cw, ch, _ := drawVScrollViewPressed(rt, list.L, list.T, list.R-list.L+1, list.B-list.T+1, gfx8.InsetColour{Map: sec}, len(w.Rows)*rowH, w.Scroll, w.hold.PressedPart())
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 4)) // clear(getShade(secondary, 4))
		if cw <= 0 || ch <= 0 {
			return
		}
		sub := gfx8.NewRenderTarget(cw, ch)
		for yy := range ch {
			copy(sub.Bits[yy*cw:(yy+1)*cw], rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw])
		}
		m30, _ := g1.GetPaletteMap(0x30)
		for i, r := range w.Rows {
			y := i*rowH - w.Scroll
			if y+rowH+30 <= 0 || y >= ch+30 {
				continue
			}
			switch {
			case i == w.Hover && r.Locked:
				fillRectCrossHatch(sub, 0, y, cw-1, y+rowH-1, paletteMutedDarkRed3)
			case i == w.Hover:
				gfx8.FillRectTranslucent(sub, 0, max(y, 0), cw-1, min(y+rowH-1, ch-1), m30)
			case r.Locked:
				fillRectCrossHatch(sub, 0, y, cw-1, y+rowH-1, paletteMutedDarkRed5)
			}
			if r.Vehicle != nil {
				DrawVehicleInline(sub, r.Vehicle, 0, y+(rowH-22)/2, 0, w.Primary, w.Secondary)
			}
		}
		for yy := range ch {
			copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
		}
	})
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, i18n.Label(i18n.BuildTrains+w.Tab), int(w.TitleColour))
	cx, cy := w.X+list.L+1, w.Y+list.T+1
	for i, r := range w.Rows {
		y := i*rowH - w.Scroll
		if y < 0 || y+rowH > list.B-list.T-1 {
			continue
		}
		x := 0
		if r.Vehicle != nil {
			_, total := r.Vehicle.InlineDrawItems(40, 0)
			x = int(total)
		}
		c := black
		if i == w.Hover {
			c = win2
		}
		DrawTextBold(screen, r.Name, cx+x+3, cy+y+(rowH-10)/2, c)
	}
	msg := w.Message
	if msg == "" {
		msg = i18n.Label(i18n.SelectNewVehicle)
	}
	DrawTextBold(screen, msg, w.X+2, w.Y+w.H-13, black)
}

// Contains reports whether the point is on the window.
// HandleClick is onMouseUp's close and vehicle-type tabs (BuildVehicle.cpp):
// close → WindowManager::close; a type tab → switchTab (here: OnTab, which
// opens the window afresh on that tab). PARTIAL: the track tabs, the list
// (buying), the search box and the dropdowns.
//
// clickable(BuildVehicle.cpp): close_button tab_build_new_trains tab_build_new_buses tab_build_new_trucks tab_build_new_trams tab_build_new_aircraft tab_build_new_ships
func (w *BuildVehicleWindow) HandleClick(x, y int) {
	lw := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)
	in := func(b BVWidget) bool {
		return !b.Hidden && x >= w.X+b.L && x <= w.X+b.R && y >= w.Y+b.T && y <= w.Y+b.B
	}
	if in(lw[bvClose]) {
		w.Close()
		return
	}
	for i := 0; i < 6; i++ {
		if in(lw[bvTabTrains+i]) && i != w.Tab && w.OnTab != nil {
			w.OnTab(i)
			return
		}
	}
	// onScrollMouseDown: the row under the pointer in the list's scroll view
	// (y / rowHeight over the scrolled content). PARTIAL: the clickDown sound.
	list := lw[bvList]
	if w.onBar(x, y) { // scrollLeftBegin on the bar
		w.Scroll = w.hold.Begin(y-(w.Y+list.T), y, list.B-list.T+1, len(w.Rows)*sim.BuildVehicleRowHeight[w.Tab], w.Scroll)
		return
	}
	cx, cy := w.X+list.L+uilayout.ScrollbarMargin, w.Y+list.T+uilayout.ScrollbarMargin
	cw, ch := list.R-list.L+1-2*uilayout.ScrollbarMargin, list.B-list.T+1-2*uilayout.ScrollbarMargin
	if x >= cx && x < cx+cw && y >= cy && y < cy+ch && w.OnBuy != nil {
		if row := (y - cy + w.Scroll) / sim.BuildVehicleRowHeight[w.Tab]; row >= 0 && row < len(w.Rows) {
			w.OnBuy(w.Rows[row].ID)
		}
	}
}

// onBar reports whether a point is on the list's vertical bar (shown when
// the rows overflow the view).
func (w *BuildVehicleWindow) onBar(x, y int) bool {
	list := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)[bvList]
	h := list.B - list.T + 1
	return !list.Hidden && len(w.Rows)*sim.BuildVehicleRowHeight[w.Tab] > h &&
		y >= w.Y+list.T && y <= w.Y+list.B && x <= w.X+list.R &&
		x >= w.X+list.R+1-uilayout.ScrollbarSize-uilayout.ScrollbarMargin
}

// Drag is the list bar's held press each tick (VScrollHold.Continue).
func (w *BuildVehicleWindow) Drag(x, y int, held bool) {
	list := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)[bvList]
	w.Scroll = w.hold.Continue(held, w.onBar(x, y), y-(w.Y+list.T), y, list.B-list.T+1, len(w.Rows)*sim.BuildVehicleRowHeight[w.Tab], w.Scroll)
}

// HandleWheel is WindowManager::wheelInput → windowScrollWheelInput on the
// vehicle list when the pointer is over it: 17 px a notch (MouseInput's
// wheel ∓17, down scrolls down), clamped to the content's height less the
// view's (bottom − top − 1). ok false elsewhere on the window.
func (w *BuildVehicleWindow) HandleWheel(x, y int, dy float64) bool {
	lw := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)
	list := lw[bvList]
	if list.Hidden || x < w.X+list.L || x > w.X+list.R || y < w.Y+list.T || y > w.Y+list.B {
		return false
	}
	wheel := 0
	if dy < 0 {
		wheel = 17
	} else if dy > 0 {
		wheel = -17
	}
	size := max(0, len(w.Rows)*sim.BuildVehicleRowHeight[w.Tab]-(list.B-list.T-1))
	w.Scroll = min(max(w.Scroll+wheel, 0), size)
	return true
}
