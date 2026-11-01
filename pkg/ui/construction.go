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
	DrawWindowFrame(screen, w.X, w.Y, ConstructionW, ConstructionH, w.CompanyColour, true)
	draw8(screen, w.X, w.Y, ConstructionW, ConstructionH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, ConstructionW-1, ConstructionH-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, ConstructionW-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		// Construction tab: the track image, animated through kTrackPreviewImages
		// (+0..15, frameNo / 4) while it is the current tab; recoloured to the
		// company colour.
		icon := w.TrackImage
		if w.Tab == 0 {
			icon += (w.FrameNo / 4) % 16
		}
		drawTab(rt, 3, 15, w.Tab == 0, sec, icon, prim)
		for i := 1; i < 4; i++ {
			drawTab(rt, 3+31*i, 15, w.Tab == i, sec, tabNoIcon, nil)
		}
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, ConstructionW-2, w.Title, int(w.CompanyColour))
}
