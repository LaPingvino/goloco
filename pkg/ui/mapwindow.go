package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// MapWindow — a port of Ui/Windows/MapWindow.cpp (PORT-PLANS A, steps 1a–c):
// 350×272, resizable 229×176 … 888×828, the skin's windowTitlebarColour and
// windowMapColour; the caption, the close button, the panel and the scroll
// view ({3, 44} to the window's width − 108 and height − 14) showing the map
// buffer (sim.MapBuffer: the map turned 45°) at the scroll offset, with the
// main view's frame on it (drawViewportPosition). It opens centred on the main
// view (centerOnViewPoint), and a press or a drag on the map centres the main
// view there (scrollMouseDown / scrollMouseDrag).
//
// clickable(MapWindow.cpp): close, the map (centres the main view)
//
// PARTIAL (PORT-PLANS A 1a, 1b, 1d): the scroll bars (drawn and clicked; the
// right drag works), the remembered size, the tabs, the graph key and its
// flashing items (_flashingItems: always none here), the status bar, the
// other overlays (vehicles, routes, town names).

const (
	mapWinW, mapWinH       = 350, 272
	mapWinMinW, mapWinMinH = 229, 176
	mapWinMaxW, mapWinMaxH = sim.MapRenderedWidth + 120, sim.MapRenderedWidth + 60
	// mapContentSize is the scroll area's contentWidth / contentHeight:
	// getScrollSize's kRenderedMapWidth + 1 (Window::initScrollWidgets).
	mapContentSize = sim.MapRenderedWidth + 1
)

// MapMainView is what the map window reads of the main window's viewport
// (WindowManager::getMainWindow()->viewports[0]): viewX / viewY, viewWidth /
// viewHeight (viewport coordinates: the screen size << zoom) and the view's
// rotation (getCurrentRotation). OK false: no main viewport.
type MapMainView struct {
	X, Y, W, H int
	Rotation   uint8
	OK         bool
}

// MapWindow is the window.
type MapWindow struct {
	WinBase
	TitleColour, Colour uint8
	// Buf is the map's pixels (filled by the game); ScrollX, ScrollY the
	// scroll view's content offset.
	Buf              *sim.MapBuffer
	ScrollX, ScrollY int
	// Rotation is var_846: the view's rotation the buffer was filled for.
	Rotation uint8
	// FrameNo is mapFrameNumber (0x00F2541D): counted by onUpdate, 0 at the
	// open; bit 2 picks the flashing buffer and the view frame's corners.
	FrameNo uint16
	// MainView reads the main viewport; OnCentre is moveMainViewToMapView
	// (0x0046B9D4): the main view centred on a world position.
	MainView func() MapMainView
	OnCentre func(x, y int32)
	// viewHeld: a left press on the map is held (Input::State::scrollLeft
	// with _currentScrollArea view).
	viewHeld bool
}

// NewMapWindow is MapWindow::open's window (createWindow: the stack places
// it); the opener sets MainView and OnCentre, then calls CentreOnViewPoint.
func NewMapWindow(buf *sim.MapBuffer, titleColour, colour uint8) *MapWindow {
	return &MapWindow{WinBase: WinBase{WinType: WindowTypeMap, W: mapWinW, H: mapWinH, MinW: mapWinMinW, MinH: mapWinMinH,
		MaxW: mapWinMaxW, MaxH: mapWinMaxH, WinFlags: WindowFlagResizable}, TitleColour: titleColour, Colour: colour, Buf: buf}
}

// scrollView is the scroll view widget as prepareDraw lays it out (x, y,
// w, h, window-relative; w, h are Widget::width() / height()).
func (w *MapWindow) scrollView() (x, y, ww, hh int) {
	return 3, 44, w.W - 108 - 3 + 1, w.H - 14 - 44 + 1
}

// bars are the scroll view's bars (both flagged visible: Scrollbars::
// horizontal | vertical) as getPart and ScrollView::draw see them: shown
// while the content is wider / taller than the widget.
func (w *MapWindow) bars() (h, v bool) {
	_, _, ww, hh := w.scrollView()
	return mapContentSize > ww, mapContentSize > hh
}

// contentSize is ScrollView::draw's crop for drawScroll (ScrollViewWidget.cpp
// draw): the widget less its 1 px border, the bars, and one more margin at
// the right and the bottom.
func (w *MapWindow) contentSize() (cw, ch int) {
	_, _, ww, hh := w.scrollView()
	hBar, vBar := w.bars()
	cw, ch = ww-2*L.ScrollbarMargin, hh-2*L.ScrollbarMargin
	if hBar {
		ch -= L.ScrollbarSize
	}
	if vBar {
		cw -= L.ScrollbarSize
	}
	return cw - L.ScrollbarMargin, ch - L.ScrollbarMargin
}

// PrepareDraw: onResize (0x0046B8E6 … l. 239) is called by open alone, before
// initScrollWidgets zeroes the offsets — nothing keeps the offset within the
// map afterwards (centerOnViewPoint may leave it up to 11 px past
// contentWidth − widget width), so nothing is done here.
func (w *MapWindow) PrepareDraw() {}

// CentreOnViewPoint is MapWindow::centerOnViewPoint (0x0046B5C0, l. 2452):
// the scroll offset that puts the main view's centre in the middle of the
// map's visible part (the widget less a bar), kept off the content's far
// edges the way upstream does. Called by open and after a rotation
// (Shortcuts.cpp rotateView, MouseInput.cpp's wheel, ToolbarTop.cpp
// rotateMenuDropdown).
func (w *MapWindow) CentreOnViewPoint() {
	if w.MainView == nil {
		return
	}
	v := w.MainView()
	if !v.OK {
		return
	}
	w.PrepareDraw() // callPrepareDraw: the scroll view sized for the window
	vpCentreX := int(int16(((v.W / 2) + v.X) / 32))
	vpCentreY := int(int16(((v.H / 2) + v.Y) / 16))

	_, _, ww, hh := w.scrollView()
	miniMapWidth := int(int16(ww - L.ScrollbarSize))
	miniMapHeight := int(int16(hh - L.ScrollbarSize))
	visibleMapWidth := int(int16(mapContentSize - miniMapWidth))
	visibleMapHeight := int(int16(mapContentSize - miniMapHeight))

	off := sim.MapViewFrameOffsets[v.Rotation&3]
	centreX := int(int16(max(vpCentreX+off[0]-(miniMapWidth/2), 0)))
	centreY := int(int16(max(vpCentreY+off[1]-(miniMapHeight/2), 0)))
	if visibleMapWidth < centreX {
		centreX = int(int16(max(centreX+(visibleMapWidth-centreX), 0)))
	}
	if visibleMapHeight < centreY {
		centreY = int(int16(max(centreY+(visibleMapHeight-centreY), 0)))
	}
	w.ScrollX, w.ScrollY = centreX, centreY
}

// mapPart is ScrollView::getPart (ScrollView.cpp l. 167) for the map's
// scroll view at a screen point on the widget: the point in the content
// (scrollviewLoc) and whether it is the view part — not a bar, not the
// border's top or left pixel.
func (w *MapWindow) mapPart(px, py int) (cx, cy int, view bool) {
	sx, sy, ww, hh := w.scrollView()
	left, top := w.X+sx, w.Y+sy
	right, bottom := left+ww-1, top+hh-1
	if px < left || px > right || py < top || py > bottom { // findWidgetAt: not on the widget
		return 0, 0, false
	}
	hBar, vBar := w.bars()
	if hBar && py >= bottom-L.ScrollbarSize || vBar && px >= right-L.ScrollbarSize {
		return 0, 0, false // a bar's part (or the deadzone)
	}
	cx, cy = px-(left+1), py-(top+1)
	if cx < 0 || cy < 0 {
		return 0, 0, false
	}
	return cx + w.ScrollX, cy + w.ScrollY, true
}

// scrollMouseDown is scrollMouseDown (0x0046B97C, l. 1171), also the
// scrollMouseDrag handler: the map point as a world position
// (mapWindowPosToLocation) and the main view moved there
// (moveMainViewToMapView).
func (w *MapWindow) scrollMouseDown(cx, cy int) {
	if w.MainView == nil || w.OnCentre == nil {
		return
	}
	v := w.MainView()
	x, y := sim.MapWindowPosToLocation(cx, cy, v.Rotation)
	w.OnCentre(x, y)
}

// HandleClick: the close button; on the map's view part a press centres the
// main view and holds (ScrollView::scrollLeftBegin, then the window's
// scrollMouseDown).
func (w *MapWindow) HandleClick(px, py int) {
	if w.Button([4]int{w.W - 15, 2, 13, 13}, px, py) { // closeButton (prepareDraw: width − 15)
		w.Close()
		return
	}
	if cx, cy, ok := w.mapPart(px, py); ok {
		w.viewHeld = true
		w.scrollMouseDown(cx, cy)
	}
}

// ScrollDrag is MouseInput.cpp stateScrollLeft each tick (l. 641): released,
// the hold ends; held with the pointer over this window's scroll view
// (onTop: the window findAt gives) and still on its view part
// (scrollLeftContinue's getPart), the window's scrollMouseDrag — the main view
// follows the pointer. Elsewhere it does nothing that tick.
func (w *MapWindow) ScrollDrag(px, py int, held, onTop bool) {
	if !w.viewHeld {
		return
	}
	if !held {
		w.viewHeld = false
		return
	}
	if !onTop {
		return
	}
	if cx, cy, ok := w.mapPart(px, py); ok {
		w.scrollMouseDown(cx, cy)
	}
}

// WidgetTooltip is the tooltip of the widget under a screen point.
func (w *MapWindow) WidgetTooltip(px, py int) string {
	x, y := px-w.X, py-w.Y
	if x >= w.W-15 && x < w.W-2 && y >= 2 && y < 15 {
		return i18n.Str(i18n.TooltipCloseWindow)
	}
	return ""
}

// mapTarget is drawScroll's render target: the crop at Org (in the window's
// target) of W × H, its rt.x / rt.y the scroll offset.
type mapTarget struct {
	rt               *gfx8.RenderTarget
	OrgX, OrgY, W, H int
	X, Y             int
}

// fillRect is drawRectImpl (SoftwareDrawingContext.cpp l. 680) on the scroll
// target: a plain fill, or (cross) RectFlags::crossHatching — every other
// pixel, its phase set by how far the rectangle is clipped at the target's
// left and top (crossPattern ^= leftX / topY) and flipped per row.
func (t mapTarget) fillRect(left, top, right, bottom int, colour uint8, cross bool) {
	if left > right || top > bottom || right < t.X || left >= t.X+t.W || bottom < t.Y || top >= t.Y+t.H {
		return
	}
	var crossPattern uint32
	leftX := left - t.X
	if leftX < 0 {
		crossPattern ^= uint32(int32(leftX))
		leftX = 0
	}
	rightX := min(right-t.X+1, t.W)
	topY := top - t.Y
	if topY < 0 {
		crossPattern ^= uint32(int32(topY))
		topY = 0
	}
	bottomY := min(bottom-t.Y+1, t.H)
	for y := topY; y < bottomY; y++ {
		p := crossPattern>>1 | crossPattern<<31 // std::rotr(crossPattern, 1)
		for x := leftX; x < rightX; x++ {
			p ^= 0x80000000
			if cross && p&0x80000000 == 0 {
				continue
			}
			sx, sy := t.OrgX+x, t.OrgY+y
			if sx >= 0 && sy >= 0 && sx < t.rt.Width && sy < t.rt.Height {
				t.rt.Bits[sy*t.rt.Width+sx] = colour
			}
		}
		crossPattern ^= 1
	}
}

// drawRectOnMap is drawRectOnMap (0x0046BE51 / 0x0046BE34, l. 1954): the
// corners put in order, then fillRect.
func (t mapTarget) drawRectOnMap(left, top, right, bottom int, colour uint8, cross bool) {
	if left > right {
		left, right = right, left
	}
	if top > bottom {
		top, bottom = bottom, top
	}
	t.fillRect(left, top, right, bottom, colour, cross)
}

// drawViewportPosition is drawViewportPosition (0x0046BAD5, l. 2006): the
// main view's frame on the map — its four edges cross-hatched in black
// (drawViewOnMap: a viewport x / 32 and y / 16 from the rotation's
// kViewFrameOffsetsByRotation) and, on the frames with mapFrameNumber's bit 2
// and no flashing graph-key items, its corners solid, 5 px along each edge
// (drawViewCornersOnMap).
func (w *MapWindow) drawViewportPosition(t mapTarget) {
	if w.MainView == nil {
		return
	}
	v := w.MainView()
	if !v.OK {
		return
	}
	off := sim.MapViewFrameOffsets[v.Rotation&3]
	viewOnMap := func(left, top, right, bottom int) { // drawViewOnMap (l. 1969)
		left, top, right, bottom = int(int16(left))/32, int(int16(top))/16, int(int16(right))/32, int(int16(bottom))/16
		t.drawRectOnMap(left+off[0], top+off[1], right+off[0], bottom+off[1], paletteBlack0, true)
	}
	l, tp, r, b := v.X, v.Y, v.X+v.W, v.Y+v.H
	viewOnMap(l, tp, r, tp)
	viewOnMap(l, b, r, b)
	viewOnMap(l, tp, l, b)
	viewOnMap(r, tp, r, b)

	if w.FrameNo&(1<<2) == 0 {
		return
	}
	// _flashingItems != 0 returns here: the graph key is not ported, none flash.
	const cornerSize = 5
	corner := func(left, top, lo, to, ro, bo int) { // drawViewCornersOnMap (l. 1987)
		left, top = int(int16(left))/32+off[0], int(int16(top))/16+off[1]
		right, bottom := left, top
		t.drawRectOnMap(left+lo, top+to, right+ro, bottom+bo, paletteBlack0, false)
	}
	corner(l, tp, 0, 0, cornerSize, 0)
	corner(r, tp, -cornerSize, 0, 0, 0)
	corner(l, tp, 0, 0, 0, cornerSize)
	corner(l, b, 0, -cornerSize, 0, 0)
	corner(l, b, 0, 0, cornerSize, 0)
	corner(r, b, -cornerSize, 0, 0, 0)
	corner(r, tp, 0, 0, 0, cornerSize)
	corner(r, b, 0, -cornerSize, 0, 0)
}

// Draw draws the window: frame, caption, panel, close button, and the
// scroll view (ScrollView::draw's inset, darker fill) with drawScroll: the
// buffer (_mapAltPixels on mapFrameNumber's bit 2) from the offset, then
// drawViewportPosition.
func (w *MapWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	sx, sy, sw, sh := w.scrollView()
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, false)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0) // Panel
		drawResizeHandle(rt, &w.WinBase, sec)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		gfx8.FillRectInset(rt, sx, sy, sx+sw-1, sy+sh-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker) // ScrollView::draw
		cw, ch := w.contentSize()
		if w.Buf == nil || cw <= 0 || ch <= 0 {
			return
		}
		t := mapTarget{rt: rt, OrgX: sx + L.ScrollbarMargin, OrgY: sy + L.ScrollbarMargin, W: cw, H: ch, X: w.ScrollX, Y: w.ScrollY}
		pix := w.Buf.Pixels
		if w.FrameNo&(1<<2) != 0 {
			pix = w.Buf.Alt
		}
		for yy := range ch {
			my := w.ScrollY + yy
			if my < 0 || my >= sim.MapRenderedWidth || t.OrgY+yy >= rt.Height {
				continue
			}
			for xx := range cw {
				mx := w.ScrollX + xx
				if mx < 0 || mx >= sim.MapRenderedWidth || t.OrgX+xx >= rt.Width {
					continue
				}
				rt.Bits[(t.OrgY+yy)*rt.Width+t.OrgX+xx] = pix[my*sim.MapRenderedWidth+mx]
			}
		}
		w.drawViewportPosition(t)
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, i18n.Label(i18n.TitleMap))
}
