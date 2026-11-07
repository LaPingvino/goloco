package ui

import (
	"image"
	"strconv"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
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
// PARTIAL: the Cargo ratings tab and the vehicle tabs (and the tabs'
// images), renaming by the caption, resizing, the viewport's rotation, the
// cargo list's scrolling, the catchment button (showStationCatchment), the
// status bar's tooltip.
const stationWinW, stationWinH = 223, 136 // Station::kWindowSize

type StationWindow struct {
	X, Y                int
	Title, Status       string
	TitleColour, Colour uint8
	Tab                 int   // 0 Station, 1 Cargo, 2 Cargo ratings
	SkinImage           int   // the interface skin's image base (the tab icons)
	FrameNo             int   // the window's frameNo (the cargo tab animates)
	CompanyColour       uint8 // the station owner's colour (the station tab's icon)
	Ratings             []StationRating
	Cargo               []StationCargoLine
	Accepted            []int // the accepted cargo's unitInlineSprite, in cargo order
	View                func(w, h int) *gfx8.RenderTarget
	OnCentreView        func()
	OnClose             func()
	viewImg             *ebiten.Image
}

// NewStationWindow places the window centred (createWindow's placement:
// PARTIAL).
func NewStationWindow(sw, sh int, title, status string, titleColour, colour uint8) *StationWindow {
	return &StationWindow{X: (sw - stationWinW) / 2, Y: (sh - stationWinH) / 2, Title: title, Status: status,
		TitleColour: titleColour, Colour: colour}
}

// stationViewport is prepareDraw's: 3,44 to width − 4, height − 14.
func stationViewport() (l, t, r, b int) { return 3, 44, stationWinW - 4, stationWinH - 14 }

// width is the current tab's kWindowSize width (Cargo ratings: 249).
func (w *StationWindow) width() int {
	if w.Tab == 2 {
		return 249
	}
	return stationWinW
}

func (w *StationWindow) Contains(px, py int) bool {
	return px >= w.X && py >= w.Y && px < w.X+w.width() && py < w.Y+stationWinH
}

// HandleClick is Station::onMouseUp: close, centre_on_viewport (the
// viewport's bottom-right 24 × 24).
func (w *StationWindow) HandleClick(px, py int) {
	rx, ry := px-w.X, py-w.Y
	if rx >= w.width()-15 && rx < w.width()-2 && ry >= 2 && ry < 15 {
		if w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	for t := 0; t < 3; t++ { // Common::switchTab
		if rx >= 3+31*t && rx < 3+31*t+31 && ry >= 15 && ry < 42 {
			w.Tab = t
			return
		}
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
			if sp, err := gfx8.Decode(&g1.Elements[imgCentreView]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, (r-24)-(l+1), (b-24)-(t+1), gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
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
			op.GeoM.Translate(float64(w.X+l+1), float64(w.Y+t+1))
			screen.DrawImage(w.viewImg, op)
			MarkForeign(image.Rect(w.X+l+1, w.Y+t+1, w.X+r, w.Y+b))
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
	l, t, r, b := 3, 44, stationWinW-26, stationWinH-14
	DrawWindowFrame(screen, w.X, w.Y, stationWinW, stationWinH, w.TitleColour, true)
	draw8(screen, w.X, w.Y, stationWinW, stationWinH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, stationWinW-1, stationWinH-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, stationWinW-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawStationTabs(rt, sec, 1)
		gfx8.FillRectInset(rt, l, t, r, b, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
		gfx8.FillRect(rt, l+1, t+1, r-1, b-1, gfx8.Shade(sec, 4))
		if sp, err := gfx8.Decode(&g1.Elements[2379]); err == nil { // show_station_catchment at (width − 25, 44)
			gfx8.DrawSpriteNoise(rt, sp, stationWinW-25, 44, gfx8.OpFor(sp, true, false), sec, nil)
		}
		y := 1
		for _, c := range w.Cargo {
			for i, x := 0, 1; i < c.Units && x < r-l-1; i, x = i+1, x+10 {
				drawG1At(rt, c.Sprite, l+1+x, t+1+y)
			}
			y += 10
			if c.Origin != "" {
				y += 10
			}
			y += 2
		}
	})
	clip := screen.SubImage(image.Rect(w.X+l+1, w.Y+t+1, w.X+r, w.Y+b)).(*ebiten.Image)
	y := w.Y + t + 1 + 1
	xr := w.X + l + 1 + (r - l + 1) - 14
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
		DrawTextFont(clip, i18n.Str(i18n.NothingWaiting), G1FontMediumNormal, w.X+l+1+1, w.Y+t+1, black)
	}
	// The status bar (3, height − 12): accepted_cargo_separator, then the sprites.
	label := stripCodes(i18n.Str(i18n.AcceptedCargoSeparator))
	sx, sy := w.X+3, w.Y+stationWinH-12
	DrawTextFont(screen, label, G1FontMediumNormal, sx-1, sy-1, black)
	lw := MeasureTextFont(label, G1FontMediumNormal)
	ox := sx + lw + 2
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
	for i := 0; i < 3; i++ {
		drawTab(rt, 3+31*i, 15, i == current, sec, icons[i], maps[i])
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
