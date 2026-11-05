package ui

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
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
	OnClose             func()
	OnPickup            func() // Common::onPickup: lift the vehicle, or take the placing tool
	PickupActive        bool   // the placing tool is this window's (the button shown pressed)
	PickupImage         int    // getPickupImageIdandTooltip's image (0: none)
}

// Draw is Details::draw and drawScroll.
func (w *VehicleWindow) Draw(screen *ebiten.Image) {
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
		for i := vwTabMain; i <= vwTabRoute; i++ {
			if t := lw[i]; !t.Hidden {
				drawTab(rt, t.L, t.T, i == vwTabDetails, sec, tabNoIcon, nil)
			}
		}
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
		cx, cy, cw, ch, _ := drawVScrollView(rt, list.L, list.T, list.R-list.L+1, list.B-list.T+1, gfx8.InsetColour{Map: sec}, len(w.Cars)*rowH, w.Scroll)
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 4)) // clearSingle(getShade(secondary, 4))
		if cw <= 0 || ch <= 0 {
			return
		}
		sub := gfx8.NewRenderTarget(cw, ch)
		for yy := range ch {
			copy(sub.Bits[yy*cw:(yy+1)*cw], rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw])
		}
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
		DrawTextBold(clip, n, cx+nameX[i], cy+y, black)
	}
}

// PickupButton is the pickup button's centre on screen; ok is false when it
// is hidden.
func (w *VehicleWindow) PickupButton() (x, y int, ok bool) {
	lw, _ := VehicleDetailsLayout(w.W, w.H, w.State)
	b := lw[vdPickup]
	return w.X + (b.L+b.R)/2, w.Y + (b.T+b.B)/2, !b.Hidden
}

// Contains reports whether the point is on the window.
func (w *VehicleWindow) Contains(x, y int) bool {
	return x >= w.X && x < w.X+w.W && y >= w.Y && y < w.Y+w.H
}

// HandleClick is Details::onMouseUp: the close button, the pickup button
// (unless hidden or disabled). PARTIAL: the other tabs and buttons.
//
// clickable(Vehicle.cpp): closeButton
// clickable-partial(Vehicle.cpp): pickup
func (w *VehicleWindow) HandleClick(x, y int) {
	lw, pickupDisabled := VehicleDetailsLayout(w.W, w.H, w.State)
	in := func(c BVWidget) bool {
		return !c.Hidden && x >= w.X+c.L && x <= w.X+c.R && y >= w.Y+c.T && y <= w.Y+c.B
	}
	switch {
	case in(lw[vwClose]) && w.OnClose != nil:
		w.OnClose()
	case in(lw[vdPickup]) && !pickupDisabled && w.OnPickup != nil:
		w.OnPickup()
	}
}
