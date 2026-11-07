package ui

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

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

// VrExpressMode is the express button's index in VehicleRouteLayout (for diags).
const VrExpressMode = vrExpressMode

// Route::widx (after the common widgets).
const (
	vrTool = vwTabRoute + 1 + iota
	vrLocalMode
	vrExpressMode
	vrRouteList
	vrForceUnload
	vrWait
	vrSkip
	vrDelete
	vrUp
	vrDown
	vrReverse
	vrWidgets
)

// Route's right-bar images (ImageIds).
const (
	imgRubbishBin       = 2363 // ImageIds::rubbish_bin (Details' remove)
	imgRedArrowUp       = 2338
	imgRedArrowDown     = 2339
	imgRouteDelete      = 2375
	imgRouteSkip        = 2376
	imgRouteWait        = 2377
	imgRouteForceUnload = 2378
)

// VehicleRouteMinHeight is Route::kMinWindowSize's height.
const VehicleRouteMinHeight = 202

// VehicleRouteLayout is Route::widgets after its prepareDraw for a window
// of width × height: the list growing with the window, the right bar at its
// edge; another company's vehicle (other) has no order buttons and a list 22
// wider; local / express share the list's width.
func VehicleRouteLayout(width, height int, hiddenTabs uint8, other bool) (w [vrWidgets]BVWidget) {
	vehicleCommonWidgets(w[:])
	w[vrLocalMode] = bvW(3, 44, 118, 12)
	w[vrExpressMode] = bvW(121, 44, 119, 12)
	w[vrRouteList] = bvW(3, 58, 237, 120)
	w[vrForceUnload] = bvW(240, 44, 24, 24)
	w[vrWait] = bvW(240, 68, 24, 24)
	w[vrSkip] = bvW(240, 92, 24, 24)
	w[vrDelete] = bvW(240, 116, 24, 24)
	w[vrUp] = bvW(240, 140, 24, 12)
	w[vrDown] = bvW(240, 152, 24, 12)
	w[vrReverse] = bvW(240, 164, 24, 24)
	w[vrTool].Hidden = true
	w[vwFrame].R, w[vwFrame].B = width-1, height-1
	w[vwPanel].R, w[vwPanel].B = width-1, height-1
	w[vwCaption].R = width - 2
	w[vwClose].L, w[vwClose].R = width-15, width-3
	w[vrRouteList].R, w[vrRouteList].B = width-26, height-14
	for _, i := range []int{vrForceUnload, vrWait, vrSkip, vrDelete, vrUp, vrDown, vrReverse} {
		w[i].L, w[i].R = width-25, width-2
		w[i].Hidden = other // isControllingCompany (named the other way round upstream)
	}
	if other {
		w[vrRouteList].R += 22
	}
	w[vrExpressMode].R = w[vrRouteList].R
	w[vrExpressMode].L = (w[vrExpressMode].R-3)/2 + 3
	w[vrLocalMode].R = w[vrExpressMode].L - 1
	leftAlignTabs(w[:], hiddenTabs)
	return w
}

// RouteLine is one row of the Route list: its text (the order as
// orderString formats it), whether it is the current order (the ▶ marker).
type RouteLine struct {
	Text    string
	Current bool
}

// Main::widx (after the common widgets).
const (
	vmViewport = vwTabRoute + 1 + iota
	vmStatus
	vmSpeedControl
	vmStopStart
	vmPickup
	vmPassSignal
	vmChangeDirection
	vmCentreViewport
	vmWidgets
)

// Main's right-bar images (ImageIds).
const (
	ImgRedFlag    = 2369
	ImgGreenFlag  = 2370
	ImgYellowFlag = 2371
	imgTurnaround = 2359 // construction_right_turnaround (change direction)
	imgCentreView = 2364
	imgPassSignal = 2374
)

// VehicleMainState is what Main::prepareDraw reads from the vehicle.
type VehicleMainState struct {
	Own        bool // CompanyManager::isPlayerCompany(head->owner)
	Manual     bool // VehicleFlags::manualControl
	HiddenTabs uint8
}

// VehicleMainLayout is Main::widgets after Main::prepareDraw for a window of
// width × height: the viewport and status line growing with the window, the
// right bar (stop / start, pickup, pass signal, change direction) at its
// right edge, the speed slider only in manual control, another company's
// vehicle without buttons, the centre-view button in the viewport's corner.
func VehicleMainLayout(width, height int, s VehicleMainState) (w [vmWidgets]BVWidget) {
	vehicleCommonWidgets(w[:])
	w[vmViewport] = bvW(3, 44, 237, 120)
	w[vmStatus] = bvW(3, 155, 237, 21)
	w[vmSpeedControl] = bvW(240, 46, 24, 115)
	w[vmStopStart] = bvW(240, 44, 24, 24)
	w[vmPickup] = bvW(240, 68, 24, 24)
	w[vmPassSignal] = bvW(240, 92, 24, 24)
	w[vmChangeDirection] = bvW(240, 116, 24, 24)
	w[vmSpeedControl].Hidden = true
	w[vwFrame].R, w[vwFrame].B = width-1, height-1
	w[vwPanel].R, w[vwPanel].B = width-1, height-1
	w[vwCaption].R = width - 2
	w[vwClose].L, w[vwClose].R = width-15, width-3
	viewportRight := width - 26
	if s.Manual && s.Own {
		viewportRight -= 27
		w[vmSpeedControl].Hidden = false
	}
	w[vmViewport].R = viewportRight
	w[vmViewport].B = height - 1 - 13
	w[vmStatus].T, w[vmStatus].B, w[vmStatus].R = height-1-13+2, height-1-13+2+9, width-14
	for _, i := range []int{vmStopStart, vmPickup, vmPassSignal, vmChangeDirection} {
		w[i].L, w[i].R = width-2-23, width-2
	}
	w[vmSpeedControl].L, w[vmSpeedControl].R = width-2-23-26, width-2-23-26+23
	if !s.Own {
		for _, i := range []int{vmStopStart, vmPickup, vmPassSignal, vmChangeDirection} {
			w[i].Hidden = true
		}
		w[vmViewport].R += 22
	}
	v := w[vmViewport]
	w[vmCentreViewport] = BVWidget{v.R - 1 - 23, v.B - 1 - 23, v.R - 1, v.B - 1, false}
	leftAlignTabs(w[:], s.HiddenTabs)
	return w
}

// leftAlignTabs is Widget::leftAlignTabs over the five tabs.
func leftAlignTabs(w []BVWidget, hidden uint8) {
	x := w[vwTabMain].L
	for i := vwTabMain; i <= vwTabRoute; i++ {
		if hidden&(1<<(i-vwTabMain)) != 0 {
			w[i].Hidden = true
			continue
		}
		w[i].L, w[i].R = x, x+30
		x += 31
	}
}

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
func vehicleCommonWidgets(w []BVWidget) {
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
	vehicleCommonWidgets(w[:])
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
	leftAlignTabs(w[:], s.HiddenTabs)
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

// vehicleRowHeights is Main::open's rowHeights by vehicle type (all 22).
var vehicleRowHeights = [6]int{22, 22, 22, 22, 22, 22}

// VehicleWindow is the Vehicle window on its Details tab (5P.g): the
// caption (the vehicle's name), the tabs, the car list (drawScroll) and the
// right bar. PARTIAL: the other tabs (Main, Cargo, Finances, Route), the
// tab icons, the right bar's images (build-new, pickup, rubbish bin, paint
// brush), the three text lines (power / weight, speed / reliability,
// length), hover and drag, the scroll bar's input, createWindow's placement.
type VehicleWindow struct {
	X, Y, W, H          int
	TitleColour, Colour uint8
	Title               string
	VehicleType         int
	State               VehicleDetailsState
	Cars                []TrainCarDraw
	Names               []string // each car's vehicle object name
	Scroll              int
	hold                L.VScrollHold // a press on the car list\'s scrollbar held (stateScrollLeft)
	OnClose             func()
	OnPickup            func() // Common::onPickup: lift the vehicle, or take the placing tool
	PickupActive        bool   // the placing tool is this window's (the button shown pressed)
	PickupImage         int    // getPickupImageIdandTooltip's image (0: none)
	BuildNewImage       int    // Details' buildNew: skin + build_additional_<type> (0: none)
	// Positioning: the window follows the pointer (MouseInput's
	// positioningWindow state, begun on the caption).
	Positioning        bool
	posLastX, posLastY int
	posMoved           bool
	// OnBuildNew is Details::onMouseDown(buildNew): the button's screen rect,
	// for its dropdown (Modify vehicle / Clone vehicle).
	OnBuildNew     func(x, y, w, h int)
	Tab            int // 0 Main, 1 Details (the other tabs are not ported yet)
	Main           VehicleMainState
	StopStartImage int                  // Main: the red / green / yellow flag
	OnStopStart    func(x, y, w, h int) // Main: stopStartOpen at the button's screen rectangle
	OnRefit        func(x, y, w, h int) // Cargo: onRefitButton at the button's screen rectangle
	OnRename       func()               // every tab's caption: renameVehicle (nil: disabled — another company's)
	OnSell         func()               // Details: the remove button (sell the whole vehicle)
	// Details' car drag (onScrollMouseDown → DragVehiclePart, scrollDragEnd):
	// Dragging while a car (DragCar, its row) follows the mouse (DragX, DragY);
	// dropped on the bin OnDropSell(row), on the list OnDropMove(row, dest row;
	// −1: below the last car — before the tail).
	Dragging     bool
	DragCar      int
	DragX, DragY int
	// Hover is rowHover in the Details / Cargo list: the car row under the
	// mouse (−1 none; while dragging len(Cars) is the tail, below the last).
	Hover         int
	OnDropSell    func(row int)
	OnDropMove    func(row, dest int)
	dragImg       *ebiten.Image
	OnTab         func(tab int)       // switchTab
	Route         []RouteLine         // Route: the order list's rows, the last the end of the list
	Finances      VehicleFinances     // Finances: its lines, formatted
	Cargo         VehicleCargo        // Cargo: its lines, formatted
	SkinImage     int                 // the interface skin's image base (tab icons; 0: none)
	FrameNo       int                 // the window's frameNo (counts updates: the current tab's icon animates)
	RouteSel      int                 // Route: orderTableIndex, the selected row (-1: none)
	OnRouteSelect func(row int)       // Route: a row clicked
	OnRoute       func(action string) // Route: "delete", "up", "down", "skip", "reverse", "local", "express"
	// OnOrderDropdown is Route::onMouseDown orderForceUnload ("unload") /
	// orderWait ("wait"): createOrderDropdown at the button's screen rect.
	OnOrderDropdown func(kind string, x, y, w, h int)
	Express         bool // Route: veh1 Flags48::expressMode (express pressed, else local)
	// View is Main's viewport, 8-bit, of the given size (nil: none — the
	// vehicle not placed).
	View         func(w, h int) *gfx8.RenderTarget
	OnCentreView func() // Main: centreViewport — the main view moved to the vehicle
	viewImg      *ebiten.Image
}

// Draw draws the current tab: Main (drawMain) or Details (Details::draw and
// drawScroll).
func (w *VehicleWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	if w.Tab == 0 {
		w.drawMain(screen)
		return
	}
	if w.Tab == 4 {
		w.drawRoute(screen)
		return
	}
	if w.Tab == 3 {
		w.drawFinances(screen)
		return
	}
	if w.Tab == 2 {
		w.drawCargo(screen)
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw, _ := VehicleDetailsLayout(w.W, w.H, w.State)
	rowH := vehicleRowHeights[max(0, min(5, w.VehicleType))]
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	list := lw[vdCarList]
	var nameX []int
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[vwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[vwClose].L, lw[vwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawVehicleTabs(rt, lw[:], vwTabDetails, sec, prim)
		// The pickup button (ImageButton::draw): inset while its tool is
		// active, then its image. PARTIAL: hover, the disabled (solid) look,
		// Gfx::recolour's exact colouring (the company colour here), the
		// other right-bar buttons.
		if b := lw[vdPickup]; !b.Hidden && w.PickupImage > 0 && w.PickupImage < len(g1.Elements) {
			if w.PickupActive {
				gfx8.FillRectInset(rt, b.L, b.T, b.R, b.B, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset)
			}
			if sp, err := gfx8.Decode(&g1.Elements[w.PickupImage]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, b.L, b.T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			}
		}
		if b := lw[vdBuildNew]; !b.Hidden && w.BuildNewImage > 0 && w.BuildNewImage < len(g1.Elements) { // recolour(buildImage, company colour)
			if sp, err := gfx8.Decode(&g1.Elements[w.BuildNewImage]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, b.L, b.T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			}
		}
		if b := lw[vdRemove]; !b.Hidden { // ImageButton rubbish_bin (tooltip_sell_or_drag_vehicle)
			if sp, err := gfx8.Decode(&g1.Elements[imgRubbishBin]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, b.L, b.T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			}
		}
		cx, cy, cw, ch, _ := drawVScrollViewPressed(rt, list.L, list.T, list.R-list.L+1, list.B-list.T+1, gfx8.InsetColour{Map: sec}, len(w.Cars)*rowH, w.Scroll, w.hold.PressedPart())
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 4)) // clearSingle(getShade(secondary, 4))
		if cw <= 0 || ch <= 0 {
			return
		}
		sub := gfx8.NewRenderTarget(cw, ch)
		for yy := range ch {
			copy(sub.Bits[yy*cw:(yy+1)*cw], rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw])
		}
		w.drawRowHover(sub, rowH, cw, true)
		for i, c := range w.Cars { // each car: its bogies, then its bodies; the name after it
			y := i*rowH - w.Scroll + (rowH-22)/2
			if c.Object == nil {
				nameX = append(nameX, 0)
				continue
			}
			drawCarInline(sub, c, 0, y, false)
			nameX = append(nameX, drawCarInline(sub, c, 0, y, true)+2)
		}
		for yy := range ch {
			copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
		}
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, w.Title, int(w.TitleColour))
	black := pal[paletteBlack0]
	cx, cy := w.X+list.L+1, w.Y+list.T+1
	// The names clipped to the scroll view's content, as its drawing context.
	clip := screen.SubImage(image.Rect(cx, cy, w.X+list.R, w.Y+list.B)).(*ebiten.Image)
	for i, n := range w.Names {
		y := i*rowH - w.Scroll + rowH/2 - 6
		if i >= len(nameX) {
			continue
		}
		if i == w.Hover && !w.Dragging { // carStr = wcolour2_stringid
			DrawTextBold(clip, n, cx+nameX[i], cy+y, pal[gfx8.Shade(mustMap(int(w.Colour)), 9)])
			continue
		}
		DrawTextBold(clip, n, cx+nameX[i], cy+y, black)
	}
}

// drawMain is Main::draw: the frame, the tabs (Main active), the viewport's
// place, the right-bar image buttons. PARTIAL: the viewport itself (an inset
// panel for now), the status line, the speed slider, hover.
func (w *VehicleWindow) drawMain(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw := VehicleMainLayout(w.W, w.H, w.Main)
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[vwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[vwClose].L, lw[vwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawVehicleTabs(rt, lw[:], vwTabMain, sec, prim)
		v := lw[vmViewport]
		gfx8.FillRectInset(rt, v.L, v.T, v.R, v.B, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
		for _, b := range []struct {
			i, img int
		}{{vmStopStart, w.StopStartImage}, {vmPickup, w.PickupImage}, {vmPassSignal, imgPassSignal}, {vmChangeDirection, imgTurnaround}, {vmCentreViewport, imgCentreView}} {
			if lw[b.i].Hidden || b.img <= 0 || b.img >= len(g1.Elements) {
				continue
			}
			if b.i == vmPickup && w.PickupActive {
				gfx8.FillRectInset(rt, lw[b.i].L, lw[b.i].T, lw[b.i].R, lw[b.i].B, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset)
			}
			if sp, err := gfx8.Decode(&g1.Elements[b.img]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, lw[b.i].L, lw[b.i].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			}
		}
	})
	// The viewport (Viewport widget: the picture inside its 1-px border),
	// with the centre-view button painted into it, as upstream's frame has
	// both, then put on screen once.
	if v := lw[vmViewport]; w.View != nil {
		if rt := w.View(v.R-v.L-1, v.B-v.T-1); rt != nil {
			if c := lw[vmCentreViewport]; !c.Hidden {
				if sp, err := gfx8.Decode(&g1.Elements[imgCentreView]); err == nil {
					gfx8.DrawSpriteNoise(rt, sp, c.L-(v.L+1), c.T-(v.T+1), gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
				}
			}
			pal := g1.GetPalette()
			pix := make([]byte, rt.Width*rt.Height*4)
			for k, ix := range rt.Bits[:rt.Width*rt.Height] {
				c := pal[ix]
				pix[k*4], pix[k*4+1], pix[k*4+2], pix[k*4+3] = c.R, c.G, c.B, 255
			}
			if w.viewImg == nil || w.viewImg.Bounds().Dx() != rt.Width || w.viewImg.Bounds().Dy() != rt.Height {
				w.viewImg = ebiten.NewImage(rt.Width, rt.Height)
			}
			w.viewImg.WritePixels(pix)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(w.X+v.L+1), float64(w.Y+v.T+1))
			screen.DrawImage(w.viewImg, op)
			MarkForeign(image.Rect(w.X+v.L+1, w.Y+v.T+1, w.X+v.R, w.Y+v.B))
		}
	}
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, w.Title, int(w.TitleColour))
}

// drawRoute is Route::draw and drawScroll: the frame, the tabs (Route
// active), the local / express buttons, the order list — each order on its
// row of 10, the current one marked ▶ — and the right bar's images. PARTIAL:
// selection and hover, the scroll bar, the order-number circles while the
// tool is on, the cargo sprites in unload / wait orders.
func (w *VehicleWindow) drawRoute(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw := VehicleRouteLayout(w.W, w.H, w.State.HiddenTabs, !w.Main.Own)
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	list := lw[vrRouteList]
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[vwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[vwClose].L, lw[vwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawVehicleTabs(rt, lw[:], vwTabRoute, sec, prim)
		for _, b := range []int{vrLocalMode, vrExpressMode} { // Button::draw: activated → borderInset
			flags := 0
			if (b == vrExpressMode) == w.Express {
				flags = gfx8.InsetBorderInset
			}
			gfx8.FillRectInset(rt, lw[b].L, lw[b].T, lw[b].R, lw[b].B, gfx8.InsetColour{Map: sec}, flags)
		}
		gfx8.FillRectInset(rt, list.L, list.T, list.R, list.B, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillNone)
		gfx8.FillRect(rt, list.L+1, list.T+1, list.R-1, list.B-1, gfx8.Shade(sec, 4)) // clearSingle(getShade(secondary, 4))
		for _, b := range []struct{ i, img int }{{vrForceUnload, imgRouteForceUnload}, {vrWait, imgRouteWait}, {vrSkip, imgRouteSkip},
			{vrDelete, imgRouteDelete}, {vrUp, imgRedArrowUp}, {vrDown, imgRedArrowDown}, {vrReverse, imgTurnaround}} {
			if lw[b.i].Hidden {
				continue
			}
			if sp, err := gfx8.Decode(&g1.Elements[b.img]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, lw[b.i].L, lw[b.i].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			}
		}
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, w.Title, int(w.TitleColour))
	pal := g1.GetPalette()
	black := pal[paletteBlack0]
	for _, b := range []struct {
		i    int
		text string
	}{{vrLocalMode, i18n.Label(i18n.LocalModeButton)}, {vrExpressMode, i18n.Label(i18n.ExpressModeButton)}} {
		tw, _ := MeasureTextBold(b.text)
		x, y := w.X+(lw[b.i].L+lw[b.i].R-tw)/2, w.Y+lw[b.i].T
		if !w.Main.Own { // disabled (another company's): the label in colour.inset()
			DrawTextBold(screen, b.text, x+1, y+1, pal[gfx8.Shade(sec, 4)])
			DrawTextBold(screen, b.text, x, y, pal[gfx8.Shade(sec, 6)])
			continue
		}
		DrawTextBold(screen, b.text, x, y, black)
	}
	cx, cy := w.X+list.L+1, w.Y+list.T+1
	clip := screen.SubImage(image.Rect(cx, cy, w.X+list.R, w.Y+list.B)).(*ebiten.Image)
	white := pal[gfx8.Shade(mustMap(colourWhite), 9)] // white_stringid's colour, as the prompt window has it
	for row, l := range w.Route {
		y := cy + row*10
		colour := black
		if row == w.RouteSel { // the selected row: a black bar, white text (black_stringid → white_stringid)
			vector.FillRect(clip, float32(cx), float32(y), float32(list.R-list.L-1), 10, black, false)
			colour = white
		}
		DrawTextBold(clip, l.Text, cx+8, y-1, colour)
		if l.Current {
			DrawTextBold(clip, "▶", cx+1, y-1, black)
		}
	}
}

// PickupButton is the pickup button's centre on screen; ok is false when it
// is hidden.
func (w *VehicleWindow) PickupButton() (x, y int, ok bool) {
	var b BVWidget
	if w.Tab == 0 {
		b = VehicleMainLayout(w.W, w.H, w.Main)[vmPickup]
	} else {
		lw, _ := VehicleDetailsLayout(w.W, w.H, w.State)
		b = lw[vdPickup]
	}
	return w.X + (b.L+b.R)/2, w.Y + (b.T+b.B)/2, !b.Hidden
}

// CentreViewButton is the centre-view button's bottom-left corner on screen
// (where its dropdown opens), on the Main tab.
func (w *VehicleWindow) CentreViewButton() (x, y int, ok bool) {
	b := VehicleMainLayout(w.W, w.H, w.Main)[vmCentreViewport]
	return w.X + b.L, w.Y + b.B + 1, w.Tab == 0 && !b.Hidden
}

// StopStartButton is the stop / start button's centre on the Main tab.
func (w *VehicleWindow) StopStartButton() (x, y int, ok bool) {
	b := VehicleMainLayout(w.W, w.H, w.Main)[vmStopStart]
	return w.X + (b.L+b.R)/2, w.Y + (b.T+b.B)/2, w.Tab == 0 && !b.Hidden
}

// Contains reports whether the point is on the window.
func (w *VehicleWindow) Contains(x, y int) bool {
	return x >= w.X && x < w.X+w.W && y >= w.Y && y < w.Y+w.H
}

// HandleClick is Details::onMouseUp: the close button, the pickup button
// (unless hidden or disabled). PARTIAL: the other tabs and buttons.
//
// clickable(Vehicle.cpp): closeButton tabMain tabDetails tabRoute tabCargo tabFinances stopStart centreViewport orderDelete orderUp orderDown orderSkip
// clickable-partial(Vehicle.cpp): pickup
func (w *VehicleWindow) HandleClick(x, y int) {
	// The caption (1, 1)–(width − 2, 13) left of the close button, on every
	// tab: a press starts moving the window (windowPositionBegin); released
	// without having moved it, it is the caption's click — renameVehicle
	// (setCaptionEnableState: the controlling company's only); see PositionEnd.
	if x >= w.X+1 && x <= w.X+w.W-16 && y >= w.Y+1 && y <= w.Y+13 {
		w.Positioning, w.posLastX, w.posLastY, w.posMoved = true, x, y, false
		return
	}
	in := func(c BVWidget) bool {
		return !c.Hidden && x >= w.X+c.L && x <= w.X+c.R && y >= w.Y+c.T && y <= w.Y+c.B
	}
	if w.Tab == 4 { // Route::onMouseUp and onScrollMouseDown. PARTIAL: the tool button
		lw := VehicleRouteLayout(w.W, w.H, w.State.HiddenTabs, !w.Main.Own)
		list := lw[vrRouteList]
		action := map[int]string{vrDelete: "delete", vrUp: "up", vrDown: "down", vrSkip: "skip", vrReverse: "reverse"}
		if w.Main.Own { // disabled for another company's vehicle
			action[vrLocalMode], action[vrExpressMode] = "local", "express"
		}
		switch {
		case in(lw[vwClose]) && w.OnClose != nil:
			w.OnClose()
		case in(lw[vwTabMain]) && w.OnTab != nil:
			w.OnTab(0)
		case in(lw[vwTabDetails]) && w.OnTab != nil:
			w.OnTab(1)
		case in(lw[vwTabFinances]) && w.OnTab != nil:
			w.OnTab(3)
		case in(lw[vwTabCargo]) && w.OnTab != nil:
			w.OnTab(2)
		case in(lw[vrForceUnload]) && w.OnOrderDropdown != nil:
			b := lw[vrForceUnload]
			w.OnOrderDropdown("unload", w.X+b.L, w.Y+b.T, b.R-b.L+1, b.B-b.T+1)
		case in(lw[vrWait]) && w.OnOrderDropdown != nil:
			b := lw[vrWait]
			w.OnOrderDropdown("wait", w.X+b.L, w.Y+b.T, b.R-b.L+1, b.B-b.T+1)
		case in(list) && w.OnRouteSelect != nil: // a row (lineHeight 10) — the end of the list too
			if row := (y - (w.Y + list.T + 1)) / 10; row >= 0 && row < len(w.Route) {
				w.OnRouteSelect(row)
			}
		default:
			for i, a := range action {
				if in(lw[i]) && w.OnRoute != nil {
					w.OnRoute(a)
					return
				}
			}
		}
		return
	}
	if w.Tab == 0 { // Main::onMouseUp / onMouseDown
		lw := VehicleMainLayout(w.W, w.H, w.Main)
		switch {
		case in(lw[vwClose]) && w.OnClose != nil:
			w.OnClose()
		case in(lw[vwTabDetails]) && w.OnTab != nil:
			w.OnTab(1)
		case in(lw[vwTabRoute]) && w.OnTab != nil:
			w.OnTab(4)
		case in(lw[vwTabFinances]) && w.OnTab != nil:
			w.OnTab(3)
		case in(lw[vwTabCargo]) && w.OnTab != nil:
			w.OnTab(2)
		case in(lw[vmStopStart]) && w.OnStopStart != nil:
			b := lw[vmStopStart]
			w.OnStopStart(w.X+b.L, w.Y+b.T, b.R-b.L+1, b.B-b.T+1)
		case in(lw[vmPickup]) && w.OnPickup != nil:
			w.OnPickup()
		case in(lw[vmCentreViewport]) && w.OnCentreView != nil:
			w.OnCentreView()
		}
		return
	}
	if w.Tab == 2 { // Cargo::onMouseUp / onMouseDown (refit). PARTIAL: renaming by the caption
		lw := VehicleCargoLayout(w.W, w.H, w.State.HiddenTabs, w.Cargo.CanRefit)
		for i, tab := range map[int]int{vwTabMain: 0, vwTabDetails: 1, vwTabFinances: 3, vwTabRoute: 4} {
			if in(lw[i]) && w.OnTab != nil {
				w.OnTab(tab)
				return
			}
		}
		if w.carBar(x, y) {
			w.barBegin(y)
			return
		}
		if b := lw[vcRefit]; in(b) && w.OnRefit != nil {
			w.OnRefit(w.X+b.L, w.Y+b.T, b.R-b.L+1, b.B-b.T+1)
			return
		}
		if in(lw[vwClose]) && w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	if w.Tab == 3 { // Finances::onMouseUp. PARTIAL: renaming by the caption
		lw := VehicleFinancesLayout(w.W, w.H, w.State.HiddenTabs)
		for i, tab := range map[int]int{vwTabMain: 0, vwTabDetails: 1, vwTabCargo: 2, vwTabRoute: 4} {
			if in(lw[i]) && w.OnTab != nil {
				w.OnTab(tab)
				return
			}
		}
		if in(lw[vwClose]) && w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	lw, pickupDisabled := VehicleDetailsLayout(w.W, w.H, w.State)
	switch {
	case in(lw[vwClose]) && w.OnClose != nil:
		w.OnClose()
	case in(lw[vwTabMain]) && w.OnTab != nil:
		w.OnTab(0)
	case in(lw[vwTabRoute]) && w.OnTab != nil:
		w.OnTab(4)
	case in(lw[vwTabFinances]) && w.OnTab != nil:
		w.OnTab(3)
	case in(lw[vwTabCargo]) && w.OnTab != nil:
		w.OnTab(2)
	case in(lw[vdPickup]) && !pickupDisabled && w.OnPickup != nil:
		w.OnPickup()
	case in(lw[vdBuildNew]) && w.OnBuildNew != nil:
		b := lw[vdBuildNew]
		w.OnBuildNew(w.X+b.L, w.Y+b.T, b.R-b.L+1, b.B-b.T+1)
	case in(lw[vdRemove]) && w.OnSell != nil: // Details::onMouseUp remove: VehicleSell(head)
		w.OnSell()
	case w.carBar(x, y): // scrollLeftBegin on the car list's bar
		w.barBegin(y)
	case in(lw[vdCarList]) && w.State.Own && w.OnDropMove != nil: // onScrollMouseDown: the car under it starts a drag
		list := lw[vdCarList]
		rowH := vehicleRowHeights[max(0, min(5, w.VehicleType))]
		if row := w.carRowAt(y-(w.Y+list.T+1)+w.Scroll, rowH); row >= 0 {
			w.Dragging, w.DragCar, w.DragX, w.DragY = true, row, x, y
		}
	}
}

// carRowAt is getCarFromScrollView: the first car whose rows reach y (scroll
// coordinates); −1 below the last.
func (w *VehicleWindow) carRowAt(y, rowH int) int {
	for i := range w.Cars {
		if y -= rowH; y <= 0 {
			return i
		}
	}
	return -1
}

// DragTo follows the mouse while a car is dragged.
func (w *VehicleWindow) DragTo(x, y int) { w.DragX, w.DragY = x, y }

// DragEnd is DragVehiclePart's onMove → Details::scrollDragEnd at (x, y):
// on the remove button the car is sold, on the car list it moves before the
// car at the middle of the row under the cursor (getCarFromScrollViewPos:
// none — the tail).
func (w *VehicleWindow) DragEnd(x, y int) {
	if !w.Dragging {
		return
	}
	w.Dragging = false
	lw, _ := VehicleDetailsLayout(w.W, w.H, w.State)
	in := func(c BVWidget) bool {
		return !c.Hidden && x >= w.X+c.L && x <= w.X+c.R && y >= w.Y+c.T && y <= w.Y+c.B
	}
	rowH := vehicleRowHeights[max(0, min(5, w.VehicleType))]
	switch list := lw[vdCarList]; {
	case in(lw[vdRemove]) && w.OnDropSell != nil:
		w.OnDropSell(w.DragCar)
	case in(list) && w.OnDropMove != nil:
		w.OnDropMove(w.DragCar, w.carRowAt(rowH/2+y-(w.Y+list.T+1)+w.Scroll, rowH))
	}
}

// DrawDrag is DragVehiclePart::draw: the dragged car, bogies then bodies, in
// a car-wide window 30 above the cursor and centred on it (open's position).
func (w *VehicleWindow) DrawDrag(screen *ebiten.Image) {
	if !w.Dragging || w.DragCar < 0 || w.DragCar >= len(w.Cars) || w.Cars[w.DragCar].Object == nil || globalG1Font == nil {
		return
	}
	c := w.Cars[w.DragCar]
	width := max(carInlineWidth(c), 1)
	// WindowFlags::transparent: only the car's own pixels (index 0 is
	// nothing) — not draw8, which repaints its whole box from the 8-bit copy
	// of the screen (the window's text isn't in it).
	rt := gfx8.NewRenderTarget(width, 60)
	drawCarInline(rt, c, 0, 0, false)
	drawCarInline(rt, c, 0, 0, true)
	pal := globalG1Font.g1.GetPalette()
	pix := make([]byte, 4*width*60)
	for i, v := range rt.Bits {
		if v != 0 {
			p := pal[v]
			pix[4*i], pix[4*i+1], pix[4*i+2], pix[4*i+3] = p.R, p.G, p.B, 0xFF
		}
	}
	if w.dragImg == nil || w.dragImg.Bounds().Dx() != width {
		w.dragImg = ebiten.NewImage(width, 60)
	}
	w.dragImg.WritePixels(pix)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(w.DragX-width/2), float64(w.DragY-30))
	screen.DrawImage(w.dragImg, op)
}

// drawRowHover is drawScroll's highlight of rowHover: the row in translucent
// ExtColour::unk30 — or, while a car is dragged (Details), a two-pixel line
// at the row's top, the tail's below the last row.
func (w *VehicleWindow) drawRowHover(sub *gfx8.RenderTarget, rowH, width int, details bool) {
	if w.Hover < 0 {
		return
	}
	m30, err := globalG1Font.g1.GetPaletteMap(0x30)
	if err != nil {
		return
	}
	top := w.Hover*rowH - w.Scroll
	if details && w.Dragging {
		gfx8.FillRectTranslucent(sub, 0, max(top-1, 0), width-1, top, m30)
		return
	}
	if w.Hover < len(w.Cars) {
		gfx8.FillRectTranslucent(sub, 0, max(top, 0), width-1, min(top+rowH-1, sub.Height-1), m30)
	}
}

// MouseOver is the car lists' scrollMouseOver (and, while dragging,
// scrollDrag): rowHover from the mouse — getCarFromScrollView at the
// cursor, or at the row's middle while dragging (none then: the tail).
func (w *VehicleWindow) MouseOver(x, y int) {
	w.Hover = -1
	var list BVWidget
	switch w.Tab {
	case 1:
		lw, _ := VehicleDetailsLayout(w.W, w.H, w.State)
		list = lw[vdCarList]
	case 2:
		list = VehicleCargoLayout(w.W, w.H, w.State.HiddenTabs, w.Cargo.CanRefit)[vcCargoList]
	default:
		return
	}
	if x < w.X+list.L || x > w.X+list.R || y < w.Y+list.T || y > w.Y+list.B {
		return
	}
	rowH := vehicleRowHeights[max(0, min(5, w.VehicleType))]
	sy := y - (w.Y + list.T + 1) + w.Scroll
	if w.Dragging {
		if w.Hover = w.carRowAt(rowH/2+sy, rowH); w.Hover < 0 {
			w.Hover = len(w.Cars)
		}
		return
	}
	w.Hover = w.carRowAt(sy, rowH)
}

// PositionMove is statePositioningWindow while the button is held: the
// pointer's y clamped to 29 … screen height − 29, the window moved by the
// pointer's travel since the last frame.
func (w *VehicleWindow) PositionMove(x, y, screenH int) {
	y = min(max(y, 29), screenH-29)
	if dx, dy := x-w.posLastX, y-w.posLastY; dx != 0 || dy != 0 {
		w.X += dx
		w.Y += dy
		w.posMoved = true
	}
	w.posLastX, w.posLastY = x, y
}

// PositionEnd is its leftReleased: the last move, then, if the window never
// moved, the caption's mouse-up (rename).
func (w *VehicleWindow) PositionEnd(x, y, screenH int) {
	w.PositionMove(x, y, screenH)
	w.Positioning = false
	if !w.posMoved && w.OnRename != nil {
		w.OnRename()
	}
}

// barList is the current tab's scrolled list and its content height:
// Details' cars, Cargo's rows (ok false on the other tabs).
func (w *VehicleWindow) barList() (list BVWidget, content int, ok bool) {
	rowH := vehicleRowHeights[max(0, min(5, w.VehicleType))]
	switch w.Tab {
	case 1:
		lw, _ := VehicleDetailsLayout(w.W, w.H, w.State)
		return lw[vdCarList], len(w.Cars) * rowH, true
	case 2:
		lw := VehicleCargoLayout(w.W, w.H, w.State.HiddenTabs, w.Cargo.CanRefit)
		return lw[vcCargoList], len(w.Cargo.Rows) * rowH, true
	}
	return BVWidget{}, 0, false
}

// carBar reports whether a point is on the current list's vertical bar
// (shown when the rows overflow the view).
func (w *VehicleWindow) carBar(x, y int) bool {
	list, content, ok := w.barList()
	return ok && !list.Hidden && content > list.B-list.T+1 &&
		y >= w.Y+list.T && y <= w.Y+list.B && x <= w.X+list.R &&
		x >= w.X+list.R+1-L.ScrollbarSize-L.ScrollbarMargin
}

// barBegin is scrollLeftBegin on the current list's bar.
func (w *VehicleWindow) barBegin(y int) {
	list, content, _ := w.barList()
	w.Scroll = w.hold.Begin(y-(w.Y+list.T), y, list.B-list.T+1, content, w.Scroll)
}

// ScrollDrag is the list bar's held press each tick (VScrollHold.Continue).
func (w *VehicleWindow) ScrollDrag(x, y int, held bool) {
	if w.hold.Part == "" {
		return
	}
	list, content, ok := w.barList()
	w.Scroll = w.hold.Continue(held && ok, w.carBar(x, y), y-(w.Y+list.T), y, list.B-list.T+1, content, w.Scroll)
}
