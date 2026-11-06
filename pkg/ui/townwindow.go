package ui

import (
	"image"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/hajimehoshi/ebiten/v2"
)

// The Town window (TownWindow.cpp), its Town tab: frame and caption (the
// town's name), the three tabs, the viewport on the town (zoom quarter), the
// centre-view button in its corner, and the status line "<size> population N".
// Outside the editor / sandbox the expand and demolish buttons are hidden and
// the viewport takes their 22 px.
// clickable(TownWindow.cpp): close_button centre_on_viewport
// PARTIAL: the Population and Company ratings tabs (and the tabs' images),
// renaming by the caption, the expand / demolish buttons (editor / sandbox),
// resizing, the viewport's own scrolling.
const townWinW, townWinH = 223, 161

type TownWindow struct {
	X, Y                int
	Title               string // the town's name
	SizeName            string // getTownSizeString
	Population          int
	TitleColour, Colour uint8
	// View renders the viewport (World.TownView); OnCentreView moves the main
	// view to the town; OnClose closes the window.
	View         func(w, h int) *gfx8.RenderTarget
	OnCentreView func()
	OnClose      func()
	viewImg      *ebiten.Image
}

// NewTownWindow places the window as upstream's open (WindowManager's
// default position: here centred on the screen).
func NewTownWindow(sw, sh int, title, size string, population int, titleColour, colour uint8) *TownWindow {
	return &TownWindow{X: (sw - townWinW) / 2, Y: (sh - townWinH) / 2, Title: title, SizeName: size, Population: population,
		TitleColour: titleColour, Colour: colour}
}

// townViewport is prepareDraw's viewport outside the editor: 3,44 to
// width-26+22, height-14 (inclusive edges).
func townViewport() (l, t, r, b int) { return 3, 44, townWinW - 26 + 22, townWinH - 14 }

func (w *TownWindow) Contains(px, py int) bool {
	return px >= w.X && py >= w.Y && px < w.X+townWinW && py < w.Y+townWinH
}

// HandleClick is onMouseUp: close, centre_on_viewport.
func (w *TownWindow) HandleClick(px, py int) {
	rx, ry := px-w.X, py-w.Y
	if rx >= townWinW-15 && rx < townWinW-2 && ry >= 2 && ry < 15 {
		if w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	_, _, r, b := townViewport()
	if rx >= r-24 && rx <= r-1 && ry >= b-24 && ry <= b-1 && w.OnCentreView != nil {
		w.OnCentreView()
	}
}

// Draw is the Town tab's draw.
func (w *TownWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	l, t, r, b := townViewport()
	DrawWindowFrame(screen, w.X, w.Y, townWinW, townWinH, w.TitleColour, true)
	draw8(screen, w.X, w.Y, townWinW, townWinH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, townWinW-1, townWinH-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, townWinW-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for i := 0; i < 3; i++ {
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
	// status_town_population: "{STRINGID} population {INT32}", left-clipped
	// at the status bar (3, height-12) − 1.
	status := strings.Replace(i18n.Label(i18n.StatusTownPopulation), "{STRINGID}", w.SizeName, 1)
	status = strings.Replace(status, "{INT32}", strconv.Itoa(w.Population), 1)
	black := g1.GetPalette()[paletteBlack0]
	DrawTextFont(screen, status, G1FontMediumNormal, w.X+3-1, w.Y+townWinH-12-1, black)
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, townWinW-2, w.Title, colourWhite)
}
