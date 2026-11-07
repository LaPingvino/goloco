package ui

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// The Station window (StationWindow.cpp), its Station tab: frame and caption
// (the station's name), the three tabs (station, cargo, cargo ratings — the
// vehicle tabs come with the vehicle types that called), the viewport on the
// station (zoom half) with the centre-view button in its corner, and the
// status bar (getStatusString).
// clickable(StationWindow.cpp): close_button centre_on_viewport
// PARTIAL: the Cargo and Cargo ratings tabs and the vehicle tabs (and the
// tabs' images), renaming by the caption, resizing, the viewport's rotation.
const stationWinW, stationWinH = 223, 136 // Station::kWindowSize

type StationWindow struct {
	X, Y                int
	Title, Status       string
	TitleColour, Colour uint8
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

func (w *StationWindow) Contains(px, py int) bool {
	return px >= w.X && py >= w.Y && px < w.X+stationWinW && py < w.Y+stationWinH
}

// HandleClick is Station::onMouseUp: close, centre_on_viewport (the
// viewport's bottom-right 24 × 24).
func (w *StationWindow) HandleClick(px, py int) {
	rx, ry := px-w.X, py-w.Y
	if rx >= stationWinW-15 && rx < stationWinW-2 && ry >= 2 && ry < 15 {
		if w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	_, _, r, b := stationViewport()
	if rx >= r-24 && rx <= r-1 && ry >= b-24 && ry <= b-1 && w.OnCentreView != nil {
		w.OnCentreView()
	}
}

// Draw is Station::draw.
func (w *StationWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
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
		for i := 0; i < 3; i++ { // leftAlignTabs(tab_station … tab_cargo_ratings)
			drawTab(rt, 3+31*i, 15, i == 0, sec, tabNoIcon, nil)
		}
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
