package ui

import (
	"image"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// The Station window (StationWindow.cpp), its Station tab: frame and caption
// (the station's name), the three tabs (station, cargo, cargo ratings — the
// vehicle tabs come with the vehicle types that called), the viewport on the
// station (zoom half) with the centre-view button in its corner, and the
// status bar (getStatusString).
// clickable(StationWindow.cpp): close_button centre_on_viewport
// The Cargo tab: the cargo waiting (Cargo::drawScroll) and "Accepts:" with
// the accepted cargo's sprites (Cargo::draw).
// The vehicle tabs (VehiclesStopping, 400×200): one per vehicle type with a
// vehicle calling here (the others hidden: leftAlignTabs), each listing those
// vehicles by name, drawn inline with their status; a row opens the vehicle.
// The Cargo tab's list scrolls (getScrollSize: 12 a cargo, 10 more with an
// origin).
// PARTIAL: resizing, the viewport's rotation, the status bar's tooltip.
const stationWinW, stationWinH = 223, 136 // Station::kWindowSize

type StationWindow struct {
	WinBase
	Title, Status       string
	TitleColour, Colour uint8
	Tab                 int   // 0 Station, 1 Cargo, 2 Cargo ratings
	SkinImage           int   // the interface skin's image base (the tab icons)
	FrameNo             int   // the window's frameNo (the cargo tab animates)
	CompanyColour       uint8 // the station owner's colour (the station tab's icon)
	Ratings             []StationRating
	Cargo               []StationCargoLine
	Accepted            []int // the accepted cargo's unitInlineSprite, in cargo order
	// OnCatchment is the Cargo tab's station_catchment (onMouseUp:
	// showStationCatchment, toggling); Catchment is whether this station's
	// catchment shows (the button activated).
	OnCatchment func()
	Catchment   bool
	// OnRename is the caption\'s click (renameStationPrompt), nil while the
	// caption is disabled (enableRenameByCaption: not a player\'s station).
	OnRename     func()
	View         func(w, h int) *gfx8.RenderTarget
	OnCentreView func()
	viewImg      *ebiten.Image
	// Labels draws the station / town names over the viewport (nil: none).
	Labels ViewLabels

	// VehicleTypes are the types with a vehicle calling here
	// (setVehicleTypeAvailable); Vehicles(t) those of type t, by name
	// (populateVehicleList / sortVehicleList); OnVehicle a row's click.
	VehicleTypes [6]bool
	Vehicles     func(t int) []VehicleRow
	OnVehicle    func(headID uint16)
	vScroll      int
	vHover       int           // rowHover's row (-1 none)
	hold         L.VScrollHold // a press on the list's scrollbar held
	cScroll      int           // the Cargo tab's scroll offset
	cHold        L.VScrollHold // a press on the Cargo tab's scrollbar held
}

// The vehicle tabs (Tab 3 + VehicleType): VehiclesStopping::kWindowSize.
const stationVehiclesW, stationVehiclesH = 400, 200

// stationTabs are the window's nine tabs laid out (leftAlignTabs): station,
// cargo, ratings, then the six vehicle types, those without a vehicle calling
// hidden.
func (w *StationWindow) stationTabs() []int {
	return tabLayout(9, func(i int) bool { return i >= 3 && !w.VehicleTypes[i-3] })
}

// vehicleList is the open vehicle tab's list rectangle (3, 44 to width − 4,
// height − 14) and its rows' height (the type's row height).
func (w *StationWindow) vehicleList() (l, t, wd, ht, rowH int) {
	return 3, 44, stationVehiclesW - 4 - 3 + 1, stationVehiclesH - 14 - 44 + 1, vehicleListRowHeights[w.Tab-3]
}

// NewStationWindow places the window centred (createWindow's placement:
// PARTIAL).
func NewStationWindow(sw, sh int, title, status string, titleColour, colour uint8) *StationWindow {
	return &StationWindow{WinBase: WinBase{WinType: WindowTypeStation, X: (sw - stationWinW) / 2, Y: (sh - stationWinH) / 2, W: stationWinW, H: stationWinH}, Title: title, Status: status,
		TitleColour: titleColour, Colour: colour, vHover: -1}
}

// stationViewport is prepareDraw's: 3,44 to width − 4, height − 14.
func stationViewport() (l, t, r, b int) { return 3, 44, stationWinW - 4, stationWinH - 14 }

// width is the current tab's kWindowSize width (Cargo ratings: 249).
func (w *StationWindow) width() int {
	switch {
	case w.Tab == 2:
		return 249
	case w.Tab >= 3:
		return stationVehiclesW
	}
	return stationWinW
}

func (w *StationWindow) height() int {
	if w.Tab >= 3 {
		return stationVehiclesH
	}
	return stationWinH
}

// PrepareDraw is the tabs' onResize (PARTIAL: each tab's kWindowSize, fixed).
func (w *StationWindow) PrepareDraw() { w.W, w.H = w.width(), w.height() }

// MouseOver is VehiclesStopping's onScrollMouseOver: the row under the mouse.
func (w *StationWindow) MouseOver(px, py int) {
	if w.Tab < 3 || w.Vehicles == nil {
		return
	}
	l, t, wd, ht, rowH := w.vehicleList()
	rx, ry := px-w.X, py-w.Y
	if rx < l || rx >= l+wd-L.ScrollbarSize-L.ScrollbarMargin || ry < t || ry >= t+ht {
		return
	}
	i := (ry - (t + 1) + w.vScroll) / rowH
	if i >= len(w.Vehicles(w.Tab-3)) {
		i = -1
	}
	w.vHover = i
}

// HandleWheel scrolls a vehicle tab's list or the Cargo tab's.
func (w *StationWindow) HandleWheel(dy float64) {
	if w.Tab == 1 {
		_, _, _, ht, contentH := w.cargoList()
		w.cScroll = min(max(w.cScroll-int(dy*10), 0), max(contentH-ht, 0))
		return
	}
	if w.Tab < 3 || w.Vehicles == nil {
		return
	}
	_, _, _, ht, rowH := w.vehicleList()
	maxScroll := max(len(w.Vehicles(w.Tab-3))*rowH-ht, 0)
	w.vScroll = min(max(w.vScroll-int(dy*float64(rowH)), 0), maxScroll)
}

// HandleClick is Station::onMouseUp: close, centre_on_viewport (the
// viewport's bottom-right 24 × 24).
func (w *StationWindow) HandleClick(px, py int) {
	rx, ry := px-w.X, py-w.Y
	if rx >= w.width()-15 && rx < w.width()-2 && ry >= 2 && ry < 15 {
		w.Close()
		return
	}
	if rx >= 1 && rx < w.width()-15 && ry >= 1 && ry <= 13 && w.OnRename != nil { // Common::onMouseUp caption
		w.OnRename()
		return
	}
	if t := tabAt(w.stationTabs(), rx, ry); t >= 0 { // Common::switchTab
		if t != 1 && w.Catchment && w.OnCatchment != nil { // leaving for another tab: this station's catchment goes
			w.OnCatchment()
		}
		w.Tab, w.vScroll, w.vHover = t, 0, -1
		return
	}
	if w.Tab >= 3 { // onScrollMouseDown: the vehicle's window; the scrollbar's buttons and track page
		w.vehicleClick(rx, ry)
		return
	}
	if w.Tab == 1 && w.cargoBar(rx, ry) { // the Cargo list's scrollbar
		_, t, _, ht, contentH := w.cargoList()
		w.cScroll = w.cHold.Begin(ry-t, py, ht, contentH, w.cScroll)
		return
	}
	if w.Tab == 1 && rx >= stationWinW-25 && rx <= stationWinW-2 && ry >= 44 && ry < 44+24 && w.OnCatchment != nil {
		w.OnCatchment()
		return
	}
	if w.Tab != 0 {
		return
	}
	_, _, r, b := stationViewport()
	if rx >= r-24 && rx <= r-1 && ry >= b-24 && ry <= b-1 && w.OnCentreView != nil {
		w.OnCentreView()
	}
}

// StationCargoLine is one cargo waiting (Cargo::drawScroll): its
// unitInlineSprite, the units drawn (min(quantity, 400) + 9) / 10, the
// quantity line ("N passengers"), and the origin line when it came from
// another station ("" when it starts here).
type StationCargoLine struct {
	Sprite, Units int
	Text, Origin  string
}

// Draw is Station::draw.
func (w *StationWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	if w.Tab == 1 {
		w.drawCargo(screen)
		return
	}
	if w.Tab == 2 {
		w.drawRatings(screen)
		return
	}
	if w.Tab >= 3 {
		w.drawVehicles(screen)
		return
	}
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	l, t, r, b := stationViewport()
	DrawWindowFrame(screen, w.X, w.Y, stationWinW, stationWinH, w.TitleColour, true)
	draw8(screen, w.X, w.Y, stationWinW, stationWinH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, stationWinW-1, stationWinH-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, stationWinW-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawStationTabs(rt, sec, 0)
		gfx8.FillRectInset(rt, l, t, r, b, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
	})
	if w.View != nil {
		if rt := w.View(r-l-1, b-t-1); rt != nil {
			drawViewport(screen, &w.viewImg, rt, w.X+l+1, w.Y+t+1, w.Labels, true, w.X+r-24, w.Y+b-24, sec)
		}
	}
	// status_bar: black_stringid with the status, left-clipped at (3, height − 12) − 1.
	black := g1.GetPalette()[paletteBlack0]
	DrawTextFont(screen, w.Status, G1FontMediumNormal, w.X+3-1, w.Y+stationWinH-12-1, black)
	DrawWindowCaptionStation(screen, w.X+1, w.Y+1, stationWinW-2, w.Title, w.TitleColour)
}

// drawCargo is Cargo::draw and drawScroll: the scroll view (3, 44 to
// width − 26, height − 14) cleared to the secondary colour's shade 4, each
// cargo's unit sprites 10 apart from (1, y) and its line right-aligned at
// the view's width − 14 in WINDOW_2 with the black outline (y + 10; the
// origin line likewise; then 2 more) — "Nothing waiting" when none —; the
// catchment button; and the status bar's "Accepts: " with the accepted
// cargo's sprites 12 apart after it, or "Nothing".
func (w *StationWindow) drawCargo(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	pal := g1.GetPalette()
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	l, t, lw, lh, contentH := w.cargoList()
	w.cScroll = min(w.cScroll, max(contentH-lh, 0))
	var cx, cy, cw, ch int
	var bar bool
	DrawWindowFrame(screen, w.X, w.Y, stationWinW, stationWinH, w.TitleColour, true)
	draw8(screen, w.X, w.Y, stationWinW, stationWinH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, stationWinW-1, stationWinH-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, stationWinW-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawStationTabs(rt, sec, 1)
		cx, cy, cw, ch, bar = drawVScrollViewPressed(rt, l, t, lw, lh, gfx8.InsetColour{Map: sec}, contentH, w.cScroll, w.cHold.PressedPart())
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 4))
		if w.Catchment { // activated: the image button drawn pressed
			gfx8.FillRectInset(rt, stationWinW-25, 44, stationWinW-2, 44+23, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset)
		}
		if sp, err := gfx8.Decode(&g1.Elements[2379]); err == nil { // show_station_catchment at (width − 25, 44)
			gfx8.DrawSpriteNoise(rt, sp, stationWinW-25, 44, gfx8.OpFor(sp, true, false), sec, nil)
		}
		sub := gfx8.NewRenderTarget(cw, ch) // the scroll view's own target (its clip)
		for yy := range ch {
			copy(sub.Bits[yy*cw:(yy+1)*cw], rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw])
		}
		y := 1 - w.cScroll
		for _, c := range w.Cargo {
			for i, x := 0, 1; i < c.Units && x < cw; i, x = i+1, x+10 {
				drawG1At(sub, c.Sprite, x, y)
			}
			y += 10
			if c.Origin != "" {
				y += 10
			}
			y += 2
		}
		for yy := range ch {
			copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
		}
	})
	if bar {
		drawVScrollArrows(screen, w.X+l, w.Y+t, lw, lh)
	}
	clip := screen.SubImage(image.Rect(w.X+cx, w.Y+cy, w.X+cx+cw, w.Y+cy+ch)).(*ebiten.Image)
	y := w.Y + cy + 1 - w.cScroll
	xr := w.X + cx + lw - 14
	for _, c := range w.Cargo {
		tw, _ := MeasureTextBold(c.Text)
		DrawTextOutlined(clip, c.Text, xr-tw, y, win2, black)
		y += 10
		if c.Origin != "" {
			ow, _ := MeasureTextBold(c.Origin)
			DrawTextOutlined(clip, c.Origin, xr-ow, y, win2, black)
			y += 10
		}
		y += 2
	}
	if len(w.Cargo) == 0 { // nothing_waiting at (1, 0)
		DrawTextFont(clip, i18n.Str(i18n.NothingWaiting), G1FontMediumNormal, w.X+cx+1, w.Y+cy, black)
	}
	// The status bar (3, height − 12): accepted_cargo_separator, then the sprites.
	label := stripCodes(i18n.Str(i18n.AcceptedCargoSeparator))
	sx, sy := w.X+3, w.Y+stationWinH-12
	DrawTextFont(screen, label, G1FontMediumNormal, sx-1, sy-1, black)
	labelW := MeasureTextFont(label, G1FontMediumNormal)
	ox := sx + labelW + 2
	if len(w.Accepted) == 0 {
		DrawTextFont(screen, i18n.Str(i18n.CargoNothingAccepted), G1FontMediumNormal, ox, sy, black)
	} else {
		draw8(screen, ox, sy-1, 12*len(w.Accepted)+12, 12, func(rt *gfx8.RenderTarget) {
			for i, sp := range w.Accepted {
				drawG1At(rt, sp, 12*i, 0)
			}
		})
	}
	DrawWindowCaptionStation(screen, w.X+1, w.Y+1, stationWinW-2, w.Title, w.TitleColour)
}

// StationRating is one cargo with an origin (CargoRatings::drawScroll): its
// name and rating (0 … 200).
type StationRating struct {
	Name   string
	Rating int
}

// drawRatings is CargoRatings::draw and drawScroll: 249 wide, the scroll
// view (3, 44 to width − 4, height − 14) cleared to shade 4; per rated cargo
// the name (wcolour2_stringid, clipped to 98) at (1, y), the bar at (100, y)
// — drawRatingBar: a 100 × 10 inset frame, filled 2 in to (rating × 327 /
// 256) × 96 / 256 in green, yellow under 100 or red under 50 — and
// "N%" (rating / 2) at (201, y); 10 apart.
func (w *StationWindow) drawRatings(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	pal := g1.GetPalette()
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	ww := w.width()
	l, t, r, b := 3, 44, ww-4, stationWinH-14
	DrawWindowFrame(screen, w.X, w.Y, ww, stationWinH, w.TitleColour, true)
	draw8(screen, w.X, w.Y, ww, stationWinH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, ww-1, stationWinH-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, ww-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawStationTabs(rt, sec, 2)
		gfx8.FillRectInset(rt, l, t, r, b, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
		cx, cy := l+1, t+1
		gfx8.FillRect(rt, cx, cy, r-1, b-1, gfx8.Shade(sec, 4))
		for i, rr := range w.Ratings {
			y := cy + 10*i
			if y+9 > b-1 {
				break
			}
			x := cx + 100
			gfx8.FillRectInset(rt, x, y, x+99, y+9, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillNone)
			colour := 13 // Colour::green
			if rr.Rating < 100 {
				colour = 16 // yellow
				if rr.Rating < 50 {
					colour = 27 // red
				}
			}
			amount := (rr.Rating * 327 / 256) & 0xFF // uint8_t
			if bar := amount * 96 / 256; bar > 2 {
				if cm, err := g1.GetPaletteMap(colour); err == nil {
					gfx8.FillRectInset(rt, x+2, y+2, x+1+bar, y+8, gfx8.InsetColour{Map: cm}, 0)
				}
			}
		}
	})
	clip := screen.SubImage(image.Rect(w.X+l+1, w.Y+t+1, w.X+r, w.Y+b)).(*ebiten.Image)
	for i, rr := range w.Ratings {
		y := w.Y + t + 1 + 10*i
		name := rr.Name
		for len(name) > 0 && MeasureTextFont(name, G1FontMediumNormal) > 98 {
			name = name[:len(name)-1]
		}
		DrawTextFont(clip, name, G1FontMediumNormal, w.X+l+1+1, y, win2)
		DrawTextFont(clip, strconv.Itoa(rr.Rating/2)+"%", G1FontMediumNormal, w.X+l+1+201, y, black)
	}
	DrawWindowCaptionStation(screen, w.X+1, w.Y+1, ww-2, w.Title, w.TitleColour)
}

// drawStationTabs is Common::drawTabs: the station tab (the skin's
// toolbar_menu_stations in the owner's colour), the cargo tab
// (tab_cargo_delivered_frame0, animated by frameNo / 8 when current) and the
// cargo ratings tab (tab_cargo_ratings) with up to four of the ratings over
// it — from (left + 4, top + 14), 3 apart: a translucent ExtColour::unk30
// strip 23 × 2, then rating × 30 / 256 in its colour's shade 6.
func (w *StationWindow) drawStationTabs(rt *gfx8.RenderTarget, sec []uint8, current int) {
	g1 := globalG1Font.g1
	icons := [3]int{tabNoIcon, tabNoIcon, tabNoIcon}
	var maps [3][]uint8
	if w.SkinImage > 0 {
		icons[0] = w.SkinImage + 190
		maps[0], _ = g1.GetPaletteMap(int(w.CompanyColour))
		icons[1] = w.SkinImage + 198
		if current == 1 {
			icons[1] += (w.FrameNo / 8) % 4
		}
		icons[2] = w.SkinImage + 48
	}
	xs := w.stationTabs()
	for i := 0; i < 3; i++ {
		drawTab(rt, xs[i], 15, i == current, sec, icons[i], maps[i])
	}
	cm, _ := g1.GetPaletteMap(int(w.CompanyColour))
	for t := 0; t < 6; t++ { // the vehicle tabs: vehicle_*_frame_0, frameNo / 2 when current, in the company colour
		if xs[3+t] < 0 {
			continue
		}
		icon := tabNoIcon
		if w.SkinImage > 0 {
			icon = w.SkinImage + stationVehicleTabFrames[t]
			if current == 3+t {
				icon += (w.FrameNo / 2) % 8
			}
		}
		drawTab(rt, xs[3+t], 15, current == 3+t, sec, icon, cm)
	}
	m30, err := g1.GetPaletteMap(0x30)
	if err != nil {
		return
	}
	x, y := 3+31*2+4, 15+14
	for i, r := range w.Ratings {
		if i >= 4 {
			break
		}
		gfx8.FillRectTranslucent(rt, x, y, x+22, y+1, m30)
		colour := 13
		if r.Rating < 100 {
			colour = 16
			if r.Rating < 50 {
				colour = 27
			}
		}
		if n := r.Rating * 30 / 256; n > 0 {
			if cm, err := g1.GetPaletteMap(colour); err == nil {
				gfx8.FillRect(rt, x, y, x-1+n, y+1, gfx8.Shade(cm, 6))
			}
		}
		y += 3
	}
}

// stationVehicleTabFrames are kTabAnimations' first frames by vehicle type:
// trains, buses, trucks, trams, aircraft, ships (vehicle_*_frame_0).
var stationVehicleTabFrames = [6]int{418, 434, 450, 442, 426, 458}

// vehicleClick is a click in a vehicle tab's list: on its scrollbar the
// buttons, track and thumb (held: Drag); on a row its vehicle opens.
func (w *StationWindow) vehicleClick(rx, ry int) {
	if w.Vehicles == nil {
		return
	}
	rows := w.Vehicles(w.Tab - 3)
	l, t, wd, ht, rowH := w.vehicleList()
	if rx < l || rx >= l+wd || ry < t || ry >= t+ht {
		return
	}
	if rx >= l+wd-L.ScrollbarSize-L.ScrollbarMargin && len(rows)*rowH > ht { // ScrollView's buttons, track and thumb
		w.vScroll = w.hold.Begin(ry-t, w.Y+ry, ht, len(rows)*rowH, w.vScroll)
		return
	}
	if i := (ry - (t + 1) + w.vScroll) / rowH; i >= 0 && i < len(rows) && w.OnVehicle != nil {
		w.OnVehicle(rows[i].HeadID)
	}
}

// drawVehicles is VehiclesStopping::draw and drawScroll: the caption
// (stringid_trains … with the station's name), the list on the secondary
// colour's shade 1, the hover row in shade 0, each vehicle drawn inline at
// its row's bottom and "name - status" outlined at (1, y + 2) (the status
// red where it says so), and the footer ("N trains" …).
func (w *StationWindow) drawVehicles(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	pal := g1.GetPalette()
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	var rows []VehicleRow
	if w.Vehicles != nil {
		rows = w.Vehicles(w.Tab - 3)
	}
	W, H := stationVehiclesW, stationVehiclesH
	l, t, wd, ht, rowH := w.vehicleList()
	w.vScroll = min(w.vScroll, max(len(rows)*rowH-ht, 0))
	var cx, cy, cw, ch int
	var bar bool
	DrawWindowFrame(screen, w.X, w.Y, W, H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, W, H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, W-1, H-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawStationTabs(rt, sec, w.Tab)
		cx, cy, cw, ch, bar = drawVScrollViewPressed(rt, l, t, wd, ht, gfx8.InsetColour{Map: sec}, len(rows)*rowH, w.vScroll, w.hold.PressedPart())
		if cw <= 0 || ch <= 0 {
			return
		}
		sub := gfx8.NewRenderTarget(cw, ch)
		for i := range sub.Bits {
			sub.Bits[i] = gfx8.Shade(sec, 1) // clearSingle(shade 1)
		}
		for i, v := range rows {
			y := i*rowH - w.vScroll
			if y+rowH <= 0 || y >= ch {
				continue
			}
			if i == w.vHover {
				gfx8.FillRect(sub, 0, max(y, 0), cw-1, min(y+rowH-1, ch-1), gfx8.Shade(sec, 0))
			}
			DrawTrainInline(sub, v.Cars, 0, y+(rowH-28)/2+6)
		}
		for yy := range ch {
			copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
		}
	})
	if bar {
		drawVScrollArrows(screen, w.X+l, w.Y+t, wd, ht)
	}
	clip := screen.SubImage(image.Rect(w.X+cx, w.Y+cy, w.X+cx+cw, w.Y+cy+ch)).(*ebiten.Image)
	for i, v := range rows {
		y := w.Y + cy + i*rowH - w.vScroll + 2
		if y+10 < w.Y+cy || y > w.Y+cy+ch {
			continue
		}
		x := w.X + cx + 1
		name := v.Name
		if v.Status != "" {
			name += " - "
		}
		x += DrawTextOutlined(clip, name, x, y, win2, black)
		if v.Status != "" {
			c := win2
			if v.StatusRed {
				if rm, err := g1.GetPaletteMap(27); err == nil { // Colour::red
					c = pal[gfx8.Shade(rm, 9)]
				}
			}
			DrawTextOutlined(clip, v.Status, x, y, c, black)
		}
	}
	DrawTextBold(screen, vehicleListFooter(w.Tab-3, len(rows)), w.X+3, w.Y+H-13, black)
	caption := strings.Replace(i18n.Label(i18n.StringidTrains+w.Tab-3), "{STRINGID}", w.Title, 1)
	DrawWindowCaptionStation(screen, w.X+1, w.Y+1, W-2, caption, w.TitleColour)
}

// Drag is the vehicle list's scrollbar press held each tick
// (VScrollHold.Continue): the thumb follows, a held button repeats.
func (w *StationWindow) Drag(px, py int, held bool) {
	if w.Tab == 1 {
		_, t, _, ht, contentH := w.cargoList()
		w.cScroll = w.cHold.Continue(held, w.cargoBar(px-w.X, py-w.Y), py-w.Y-t, py, ht, contentH, w.cScroll)
		return
	}
	if w.Tab < 3 || w.Vehicles == nil {
		return
	}
	l, t, wd, ht, rowH := w.vehicleList()
	rx, ry := px-w.X, py-w.Y
	onBar := rx >= l+wd-L.ScrollbarSize-L.ScrollbarMargin && rx < l+wd && ry >= t && ry < t+ht
	w.vScroll = w.hold.Continue(held, onBar, ry-t, py, ht, len(w.Vehicles(w.Tab-3))*rowH, w.vScroll)
}

// cargoList is the Cargo tab's scroll view (3, 44 to width − 26, height − 14:
// x, y, width, height) and its content height (getScrollSize).
func (w *StationWindow) cargoList() (l, t, wd, ht, contentH int) {
	for _, c := range w.Cargo {
		contentH += 12
		if c.Origin != "" {
			contentH += 10
		}
	}
	return 3, 44, stationWinW - 26 - 3 + 1, stationWinH - 14 - 44 + 1, contentH
}

// cargoBar reports window point (rx, ry) on the Cargo tab's scrollbar (shown
// when the cargo doesn't fit).
func (w *StationWindow) cargoBar(rx, ry int) bool {
	l, t, wd, ht, contentH := w.cargoList()
	return contentH > ht && rx >= l+wd-L.ScrollbarSize-L.ScrollbarMargin && rx < l+wd && ry >= t && ry < t+ht
}

// stationTabTooltips are the Station window's tab tooltips (makeCommonWidgets).
var stationTabTooltips = [9]int{i18n.TooltipStation, i18n.TooltipStationCargo, i18n.TooltipStationCargoRatings,
	i18n.TooltipTrains, i18n.TooltipBuses, i18n.TooltipTrucks, i18n.TooltipTrams, i18n.TooltipAircraft, i18n.TooltipShips}

// WidgetTooltip is the tooltip of the widget under a screen point: close, a
// tab, and on the Station tab the centre-view button.
func (w *StationWindow) WidgetTooltip(px, py int) string {
	rx, ry := px-w.X, py-w.Y
	if rx >= w.width()-15 && rx < w.width()-2 && ry >= 2 && ry < 15 {
		return i18n.Str(i18n.TooltipCloseWindow)
	}
	if i := tabAt(w.stationTabs(), rx, ry); i >= 0 {
		return i18n.Str(stationTabTooltips[i])
	}
	if w.Tab == 0 {
		if _, _, r, b := stationViewport(); rx >= r-24 && rx <= r-1 && ry >= b-24 && ry <= b-1 {
			return i18n.Str(i18n.MoveMainViewToShowThis)
		}
	}
	return ""
}
