package ui

import (
	"image"
	"image/color"

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
// PARTIAL: the search box, the filter / cargo / sort dropdowns.

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

// rect is the widget as left, top, width, height (WinBase.Button).
func (b BVWidget) rect() [4]int { return [4]int{b.L, b.T, b.R - b.L + 1, b.B - b.T + 1} }

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
// The caption is build_trains + currentTab whether or not it adds to a
// vehicle (prepareDraw); a target shows in the bottom-left message.
// PARTIAL (5P.f part 4): the search box and the three dropdowns' texts, the
// cargo-filter line (draw's _cargoSupportedFilter), the tooltips.
type BuildVehicleWindow struct {
	WinBase
	TitleColour, Colour uint8
	Tab, TrackTab       int
	Disabled            uint8
	NumTrackTabs        int
	// Target is _buildTargetVehicle: the vehicle a row adds a car to, or
	// sim's EntityIDNull (0xFFFF) for a new vehicle.
	Target uint16
	// TabIcon / TrackTabIcon are drawTransportTypeTabs' and
	// drawTrackTypeTabs' images for a tab (the active one animated by the
	// caller) and their colour (-1: none).
	TabIcon, TrackTabIcon func(i int) (img, colour int)
	Rows                  []BVRow
	// RowHover is window.rowHover: the hovered vehicle's object id (not a
	// row index), -1 none. onResize sets the first row's when none.
	RowHover int
	// TargetName is _buildTargetVehicle's name (vehicle->name with its
	// ordinalNumber, formatted), "" for none: the bottom-left message
	// (select_vehicle_to_add_to_string_id) and the empty list's
	// (no_compatible_vehicles_available).
	TargetName string
	// Details is draw's buffer_1250 for a vehicle object (VehicleStats).
	Details func(id uint16) string
	// CargoSprites are the vehicle's usable cargoes' unitInlineSprite
	// images, by cargo type (drawScroll's vehicle_preview).
	CargoSprites func(id uint16) []int
	// PreviewFrame is WindowManager's vehiclePreviewRotationFrame.
	PreviewFrame       func() int
	shownHover         int // rowHover as the last Draw showed it
	Scroll             int
	hold               uilayout.VScrollHold // a press on the list\'s scrollbar held (stateScrollLeft)
	Primary, Secondary uint8                // the controlling company's colours for the vehicles
	OnTab              func(int)            // a vehicle-type tab
	OnBuy              func(id uint16)      // a row clicked (onScrollMouseDown): buy that vehicle
}

// OnResize is onResize (0x4C3929): the window resizable within
// BuildVehicleSize's bounds (at least max(track tabs · 31 + 195, 380) ×
// 233, at most kMaxWindowSize 700 × 600), and the list's scroll no further
// than its rows reach (contentHeight − (bottom − top)).
// Then, with no row hovered and rows to show, rowHover is the first row's
// vehicle. Upstream calls it from open (callOnResize) — onUpdate's refresh
// (sub_4B92A5) clears rowHover again at the next update, so the details
// show while the pointer is on a row.
func (w *BuildVehicleWindow) OnResize() {
	w.resizeBounds()
	if w.RowHover == -1 && len(w.Rows) > 0 {
		w.RowHover = int(w.Rows[0].ID)
	}
}

// resizeBounds is onResize's size bounds and scroll clamp.
func (w *BuildVehicleWindow) resizeBounds() {
	w.WinFlags |= WindowFlagResizable
	w.SetSize(max(w.NumTrackTabs*31+195, 380), 233, 700, 600)
	list := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)[bvList]
	w.Scroll = min(w.Scroll, max(0, len(w.Rows)*sim.BuildVehicleRowHeight[w.Tab]-(list.B-list.T)))
}

// Refresh is sub_4B92A5 on the open window: the target, the list (and what
// goes with it) from n, a window built afresh for it; rowHover −1. The
// window's place, size, scroll and held scrollbar stay.
func (w *BuildVehicleWindow) Refresh(n *BuildVehicleWindow) {
	base, scroll, hold, shown := w.WinBase, w.Scroll, w.hold, w.shownHover
	*w = *n
	w.WinBase, w.Scroll, w.hold, w.shownHover = base, scroll, hold, shown
	w.RowHover = -1
	w.resizeBounds()
}

// listContent is the vehicle list's scrolled view in screen coordinates:
// the widget less its 1 px border, less the bar when the rows overflow.
func (w *BuildVehicleWindow) listContent() (x, y, cw, ch int) {
	list := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)[bvList]
	x, y = w.X+list.L+uilayout.ScrollbarMargin, w.Y+list.T+uilayout.ScrollbarMargin
	cw, ch = list.R-list.L+1-2*uilayout.ScrollbarMargin, list.B-list.T+1-2*uilayout.ScrollbarMargin
	if len(w.Rows)*sim.BuildVehicleRowHeight[w.Tab] > list.B-list.T+1 {
		cw -= uilayout.ScrollbarSize
	}
	return
}

// rowAt is the list row under a screen point (y / rowHeight over the
// scrolled content), -1 off the view or past the last row.
func (w *BuildVehicleWindow) rowAt(x, y int) int {
	cx, cy, cw, ch := w.listContent()
	if x < cx || x >= cx+cw || y < cy || y >= cy+ch {
		return -1
	}
	if row := (y - cy + w.Scroll) / sim.BuildVehicleRowHeight[w.Tab]; row < len(w.Rows) {
		return row
	}
	return -1
}

// MouseOver is onScrollMouseOver (0x4C3802) over the vehicle list: the
// row under the pointer's vehicle becomes rowHover; past the last row, or
// off the list, rowHover stays.
func (w *BuildVehicleWindow) MouseOver(x, y int) {
	if r := w.rowAt(x, y); r >= 0 {
		w.RowHover = int(w.Rows[r].ID)
	}
}

// PrepareDraw runs onResize's bounds and scroll clamp: the stack calls it
// as the size changes (not its rowHover part: see OnResize).
func (w *BuildVehicleWindow) PrepareDraw() { w.resizeBounds() }

// OnFrame: no widget but the frame and the panel at the point
// (findWidgetAt): a press there moves the window, or in the corner
// resizes it.
func (w *BuildVehicleWindow) OnFrame(x, y int) bool {
	lw := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)
	return bareFrame(lw[:], bvPanel, w.X, w.Y, x, y)
}

// bareFrame reports a point on the window whose last widget there
// (findWidgetAt) is the frame or the panel — the widgets up to and with
// index panel are the frame, caption, close button and panel.
func bareFrame(lw []BVWidget, panel, wx, wy, x, y int) bool {
	at := -1
	for i, c := range lw {
		if !c.Hidden && x >= wx+c.L && x <= wx+c.R && y >= wy+c.T && y <= wy+c.B {
			at = i
		}
	}
	return at == 0 || at == panel
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
	w.shownHover = w.RowHover
	lw := BuildVehicleLayout(w.W, w.H, w.Disabled, w.NumTrackTabs)
	rowH := sim.BuildVehicleRowHeight[w.Tab]
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	list := lw[bvList]
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[bvPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		drawResizeHandle(rt, &w.WinBase, sec) // PanelWidget: the handle on a resizable window
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[bvClose].L, lw[bvClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		icon := func(f func(int) (int, int), i int) (int, []uint8) {
			if f == nil {
				return tabNoIcon, nil
			}
			img, c := f(i)
			if img < 0 || img >= len(g1.Elements) {
				return tabNoIcon, nil
			}
			var m []uint8
			if c >= 0 {
				m, _ = g1.GetPaletteMap(c)
			}
			return img, m
		}
		for i := range 6 { // drawTransportTypeTabs
			if t := lw[bvTabTrains+i]; !t.Hidden {
				img, m := icon(w.TabIcon, i)
				drawTab(rt, t.L, t.T, i == w.Tab, sec, img, m)
			}
		}
		// drawTrackTypeTabs: the two lines bounding the track tabs (shade 7),
		// the active tab's underline (shade 5), then each tab.
		gfx8.FillRect(rt, 0, 69, w.W-187, 69, gfx8.Shade(sec, 7))
		gfx8.FillRect(rt, w.W-187, 41, w.W-187, 41+27, gfx8.Shade(sec, 7))
		for i := range 8 {
			if t := lw[bvTabTrack0+i]; !t.Hidden {
				if i == w.TrackTab {
					gfx8.FillRect(rt, t.L+1, t.T+26, t.L+1+29, t.T+26, gfx8.Shade(sec, 5))
				}
				img, m := icon(w.TrackTabIcon, i)
				drawTab(rt, t.L, t.T, i == w.TrackTab, sec, img, m)
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
			hover := int(r.ID) == w.RowHover
			y := i*rowH - w.Scroll
			if y+rowH+30 <= 0 || y >= ch+30 {
				continue
			}
			switch {
			case hover && r.Locked:
				fillRectCrossHatch(sub, 0, y, cw-1, y+rowH-1, paletteMutedDarkRed3)
			case hover:
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
		w.drawPreview(rt, lw[bvPreview], sec)
	})
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, i18n.Label(i18n.BuildTrains+w.Tab), int(w.TitleColour))
	// The rows' names, clipped to the scroll view's content (drawScroll's
	// render target): a row the view cuts shows the part inside.
	cx, cy, vw, vh := w.listContent()
	clip := screen.SubImage(image.Rect(cx, cy, cx+vw-uilayout.ScrollbarMargin, cy+vh-uilayout.ScrollbarMargin)).(*ebiten.Image)
	for i, r := range w.Rows {
		y := i*rowH - w.Scroll
		if y+rowH <= 0 || y >= vh {
			continue
		}
		x := 0
		if r.Vehicle != nil {
			_, total := r.Vehicle.InlineDrawItems(40, 0)
			x = int(total)
		}
		c := black
		if int(r.ID) == w.RowHover { // wcolour2_stringid
			c = win2
		}
		DrawTextBold(clip, r.Name, cx+x+3, cy+y+(rowH-10)/2, c)
	}
	drawVScrollArrows(screen, w.X+list.L, w.Y+list.T, list.R-list.L+1, list.B-list.T+1)
	if len(w.Rows) == 0 { // drawScroll: no_vehicles_available / no_compatible_vehicles_available at (3, (rowHeight − 10) / 2)
		msg := i18n.Str(i18n.NoVehiclesAvailable)
		if w.TargetName != "" {
			msg = FillArgs(i18n.Str(i18n.NoCompatibleVehiclesAvailable), w.TargetName)
		}
		drawCodedWrapped(screen, msg, cx+3, cy+(rowH-10)/2, list.R-list.L-17, win2, black)
	}
	// draw: the bottom-left message, clipped to width − 186.
	msg := i18n.Label(i18n.SelectNewVehicle)
	if w.TargetName != "" {
		msg = FillArgs(i18n.Label(i18n.SelectVehicleToAddToStringId), w.TargetName)
	}
	DrawTextBold(screen, ClipString(msg, G1FontMediumBold, w.W-186), w.X+2, w.Y+w.H-13, black)
	w.drawPreviewName(screen, lw[bvPreview], win2)
	// draw: the hovered vehicle's buffer_1250, wrapped to 180 at
	// (selection.right + 2, preview.bottom + 2).
	if w.RowHover >= 0 && w.Details != nil {
		drawCodedWrapped(screen, w.Details(uint16(w.RowHover)), w.X+list.R+2, w.Y+lw[bvPreview].B+2, 180, win2, black)
	}
}

// ShownHover is rowHover as the last frame drew it (the details shown for
// that vehicle; -1 none) — a script's check, between frames.
func (w *BuildVehicleWindow) ShownHover() int { return w.shownHover }

// hoveredVehicle is rowHover's vehicle object among the rows (nil none).
func (w *BuildVehicleWindow) hoveredVehicle() *objects.VehicleObject {
	for _, r := range w.Rows {
		if int(r.ID) == w.RowHover {
			return r.Vehicle
		}
	}
	return nil
}

// drawPreview is drawScroll's vehicle_preview (a scroll view without bars):
// cleared to the secondary colour's shade 0, then the hovered vehicle's
// drawVehicleOverview at (90, 37) turned by the preview rotation frame (yaw
// frame & 0x3F, roll ((frame + 2) / 4) & 0x3F).
func (w *BuildVehicleWindow) drawPreview(rt *gfx8.RenderTarget, p BVWidget, sec []uint8) {
	cx, cy, cw, ch, _ := drawVScrollView(rt, p.L, p.T, p.R-p.L+1, p.B-p.T+1, gfx8.InsetColour{Map: sec}, 0, 0)
	if cw <= 0 || ch <= 0 {
		return
	}
	gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 0))
	v := w.hoveredVehicle()
	if v == nil {
		return
	}
	frame := 0
	if w.PreviewFrame != nil {
		frame = w.PreviewFrame()
	}
	sub := gfx8.NewRenderTarget(cw, ch)
	for yy := range ch {
		copy(sub.Bits[yy*cw:(yy+1)*cw], rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw])
	}
	DrawVehicleOverview(sub, v, 90, 37, uint8(frame&0x3F), uint8(((frame+2)/4)&0x3F), w.Primary, w.Secondary)
	for yy := range ch {
		copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
	}
}

// drawPreviewName is the preview's last line: wcolour2_stringid of the
// vehicle's name and, for each usable cargo, a space and its
// unitInlineSprite — drawStringCentredClipped at (89, 52), width 177.
func (w *BuildVehicleWindow) drawPreviewName(screen *ebiten.Image, p BVWidget, win2 color.Color) {
	v := w.hoveredVehicle()
	if v == nil || globalG1Font == nil {
		return
	}
	var sprites []int
	if w.CargoSprites != nil {
		sprites = w.CargoSprites(uint16(w.RowHover))
	}
	drawInlineCentredClipped(screen, v.DisplayName, sprites, w.X+p.L+1+89, w.Y+p.T+1+52, 177, win2)
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
		w.PressWidget(lw[bvClose].rect())
		w.Close()
		return
	}
	for i := 0; i < 6; i++ {
		if in(lw[bvTabTrains+i]) { // a disabled tab is hidden (leftAlignTabs)
			w.PressWidget(lw[bvTabTrains+i].rect())
			if i != w.Tab && w.OnTab != nil {
				w.OnTab(i)
			}
			return
		}
	}
	// onScrollMouseDown: the row under the pointer in the list's scroll view
	// (y / rowHeight over the scrolled content). The clickDown sound is the caller's (buyVehicle).
	list := lw[bvList]
	if w.onBar(x, y) { // scrollLeftBegin on the bar
		w.Scroll = w.hold.Begin(y-(w.Y+list.T), y, list.B-list.T+1, len(w.Rows)*sim.BuildVehicleRowHeight[w.Tab], w.Scroll)
		return
	}
	if row := w.rowAt(x, y); row >= 0 && w.OnBuy != nil {
		w.OnBuy(w.Rows[row].ID)
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
