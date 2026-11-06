package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// ConstructionFrame is the frame of the construction window
// (Ui/Windows/Construction/Common.cpp makeCommonWidgets, createConstructionWindow):
// 138×276, lighter frame in the owner company's colour (WindowManager draws an
// owned window's primary colour as the company's), a colourText caption with
// the track's name, the close button, the panel in the skin's
// windowConstructionColour, and the four tabs (construction, station, signal,
// overhead) drawn by drawTrackTabs.
// PARTIAL: the station, signal and overhead tabs' previews (zoomed station
// image, animated signal); tabs hidden by setDisabledWidgets/repositionTabs;
// road, airport and dock tab sets; the tab contents (ConstructionTab etc.).
type ConstructionFrame struct {
	X, Y          int
	Title         string
	CompanyColour uint8
	SkinColour    uint8 // InterfaceSkin windowConstructionColour
	TrackImage    int   // the track object's image (TrackObject::image)
	Tab           int   // 0 construction, 1 station, 2 signal, 3 overhead
	FrameNo       int   // window frameNo (counts updates; animates the tab)
	// TabFrames is how many preview frames the construction tab cycles through
	// (drawTrackTabs: 16, drawRoadTabs: 32).
	TabFrames int
	// Disabled tabs (setDisabledWidgets) are hidden and the others move left
	// (repositionTabs): index 0 construction, 1 station, 2 signal, 3 overhead.
	Disabled [4]bool
	// H is the window height for the current tab (0: ConstructionH).
	H int
	// StationTabIcon is the station tab's image (drawDockTabs / drawAirportTabs:
	// the skin's toolbar_menu_ship_port / _airport; 0: none, as drawTrackTabs).
	StationTabIcon int
}

// Height is the window's height (each tab sets its own size).
func (w *ConstructionFrame) Height() int {
	if w.H > 0 {
		return w.H
	}
	return ConstructionH
}

// ConstructionW/H is Construction::kWindowSize.
const ConstructionW, ConstructionH = 138, 276

// Draw draws the frame and tabs.
func (w *ConstructionFrame) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	prim, _ := g1.GetPaletteMap(int(w.CompanyColour))
	sec, err := g1.GetPaletteMap(int(w.SkinColour))
	if err != nil {
		return
	}
	h := w.Height()
	DrawWindowFrame(screen, w.X, w.Y, ConstructionW, h, w.CompanyColour, true)
	draw8(screen, w.X, w.Y, ConstructionW, h, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, ConstructionW-1, h-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, ConstructionW-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		// Construction tab: the track image, animated through kTrackPreviewImages
		// (+0..15, frameNo / 4) while it is the current tab; recoloured to the
		// company colour.
		icon := w.TrackImage
		if w.Tab == 0 {
			icon += (w.FrameNo / 4) % max(w.TabFrames, 1)
		}
		for i, x := range w.TabX() {
			if x < 0 {
				continue
			}
			if i == 0 {
				drawTab(rt, x, 15, w.Tab == 0, sec, icon, prim)
			} else if i == 1 && w.StationTabIcon > 0 {
				drawTab(rt, x, 15, w.Tab == i, sec, w.StationTabIcon, nil)
			} else {
				drawTab(rt, x, 15, w.Tab == i, sec, tabNoIcon, nil)
			}
		}
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, ConstructionW-2, w.Title, int(w.CompanyColour))
}

// TabX is repositionTabs: each tab's x, -1 for a hidden (disabled) one; the
// enabled tabs sit side by side from x 3, 31 apart.
func (w *ConstructionFrame) TabX() [4]int {
	var xs [4]int
	x := 3
	for i := range xs {
		if w.Disabled[i] {
			xs[i] = -1
			continue
		}
		xs[i] = x
		x += 31
	}
	return xs
}

// TabAt returns the tab under (rx, ry) relative to the window, or -1.
func (w *ConstructionFrame) TabAt(rx, ry int) int {
	if ry < 15 || ry >= 15+27 {
		return -1
	}
	for i, x := range w.TabX() {
		if x >= 0 && rx >= x && rx < x+31 {
			return i
		}
	}
	return -1
}
