package ui

import (
	"image"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/hajimehoshi/ebiten/v2"
)

// ViewWindow is the first tab both the Town window (TownWindow.cpp) and the
// Industry window (IndustryWindow.cpp) open on: frame and caption (the
// place's name), the tabs, the viewport on the place (zoom quarter), the
// centre-view button in its corner, and a status line (the town's size and
// population, the industry's getStatusString). Outside the editor / sandbox
// the expand / demolish buttons are hidden and the viewport takes their 22 px.
// clickable(TownWindow.cpp, IndustryWindow.cpp): close_button centre_on_viewport
// PARTIAL: the other tabs (town: Population, Company ratings; industry:
// Production, Production 2, Transported) and the tabs' images, renaming by the
// caption, the editor's buttons, resizing, the viewport's own scrolling, the
// status's clipping (drawStringLeftClipped).
type ViewWindow struct {
	X, Y, W, H          int
	Tabs                int    // 3 (town) or 4 (industry)
	Title               string // the place's name
	Status              string // the status bar's line
	TitleColour, Colour uint8
	// View renders the viewport; OnCentreView moves the main view to the
	// place; OnClose closes the window.
	View         func(w, h int) *gfx8.RenderTarget
	OnCentreView func()
	OnClose      func()
	viewImg      *ebiten.Image
}

// NewTownWindow is Town::open's window (223×161, three tabs), its status
// status_town_population "{STRINGID} population {INT32}". Placed as
// WindowManager's default position (here centred on the screen).
func NewTownWindow(sw, sh int, title, size string, population int, titleColour, colour uint8) *ViewWindow {
	status := strings.Replace(i18n.Label(i18n.StatusTownPopulation), "{STRINGID}", size, 1)
	status = strings.Replace(status, "{INT32}", strconv.Itoa(population), 1)
	return newViewWindow(sw, sh, 223, 161, 3, title, status, titleColour, colour)
}

// NewIndustryWindow is IndustryWindow open's window (223×137, four tabs), its
// status the industry's getStatusString.
func NewIndustryWindow(sw, sh int, title, status string, titleColour, colour uint8) *ViewWindow {
	return newViewWindow(sw, sh, 223, 137, 4, title, status, titleColour, colour)
}

func newViewWindow(sw, sh, w, h, tabs int, title, status string, titleColour, colour uint8) *ViewWindow {
	return &ViewWindow{X: (sw - w) / 2, Y: (sh - h) / 2, W: w, H: h, Tabs: tabs, Title: title, Status: status,
		TitleColour: titleColour, Colour: colour}
}

// viewport is prepareDraw's viewport outside the editor: 3,44 to
// width-26+22, height-14 (inclusive edges).
func (w *ViewWindow) viewport() (l, t, r, b int) { return 3, 44, w.W - 26 + 22, w.H - 14 }

func (w *ViewWindow) Contains(px, py int) bool {
	return px >= w.X && py >= w.Y && px < w.X+w.W && py < w.Y+w.H
}

// HandleClick is onMouseUp: close, centre_on_viewport.
func (w *ViewWindow) HandleClick(px, py int) {
	rx, ry := px-w.X, py-w.Y
	if rx >= w.W-15 && rx < w.W-2 && ry >= 2 && ry < 15 {
		if w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	_, _, r, b := w.viewport()
	if rx >= r-24 && rx <= r-1 && ry >= b-24 && ry <= b-1 && w.OnCentreView != nil {
		w.OnCentreView()
	}
}

// Draw is the first tab's draw.
func (w *ViewWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	l, t, r, b := w.viewport()
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for i := 0; i < w.Tabs; i++ {
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
	// the status, left-clipped at the status bar (3, height-12) − 1
	black := g1.GetPalette()[paletteBlack0]
	DrawTextFont(screen, w.Status, G1FontMediumNormal, w.X+3-1, w.Y+w.H-12-1, black)
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, w.Title, colourWhite)
}
