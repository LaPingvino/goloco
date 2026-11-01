package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// StationTabH is the station tab's window height before a ghost is placed
// (makeCommonWidgets(138, 190)).
const StationTabH = 190

// StationTabState is what the station tab shows (Construction/StationTab.cpp)
// for a train or road station.
type StationTabState struct {
	Name         string // the station object's name (the dropdown's text)
	Image        int    // its image (+0 preview, +1 preview windows)
	Recolourable bool   // TrainStationFlags/RoadStationFlags::recolourable
	Company      uint8  // the player company's primary colour
}

// DrawStationTab is Station::draw (0x0049DE40) without a ghost: the station
// dropdown (3, 45), the preview at the image widget (35, 60): the station
// recoloured to the company colour, then its windows drawn translucent in
// getTranslucent(company) (ExtColour 0x2E when not recolourable); and the
// separator at the image's bottom + 16.
// PARTIAL: airports and docks; the cost line; with a ghost placed, the
// station name and the catchment's accepted/produced cargo (and the window
// growing for them); the dropdown list; placing (the tool).
func DrawStationTab(screen *ebiten.Image, f *ConstructionFrame, st StationTabState) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(f.SkinColour))
	if err != nil {
		return
	}
	pal := g1.GetPalette()
	black := pal[paletteBlack0]
	prim, _ := g1.GetPaletteMap(int(st.Company))
	win := 0x2E // ExtColour::unk2E
	if st.Recolourable {
		win = int(paintTranslucent(st.Company, 1)) // getTranslucent(colour) = shade 1
	}
	winMap, _ := g1.GetPaletteMap(win)
	const imgL, imgT, imgB = 35, 60, 60 + 68 - 1
	draw8(screen, f.X, f.Y, ConstructionW, f.Height(), func(rt *gfx8.RenderTarget) {
		optDropdown(rt, 3, 45, 132, sec)
		if st.Image >= 0 && st.Image+1 < len(g1.Elements) {
			if sp, err := gfx8.Decode(&g1.Elements[st.Image]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, imgL, imgT, gfx8.OpFor(sp, prim != nil, false), prim, nil)
			}
			if sp, err := gfx8.Decode(&g1.Elements[st.Image+1]); err == nil && winMap != nil {
				gfx8.DrawSpriteNoise(rt, sp, imgL, imgT, gfx8.OpTransparent|gfx8.OpDst, winMap, nil)
			}
		}
		// drawRectInset(x + 3, image.bottom + 16, width − 4, 1, borderInset)
		y := imgB + 16
		gfx8.FillRectInset(rt, 3, y, 3+(ConstructionW-4)-1, y, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillNone)
	})
	DrawTextBold(screen, st.Name, f.X+3+1, f.Y+45+1, black)
	arrow := string(rune(170))
	aw, _ := MeasureTextBold(arrow)
	DrawTextBold(screen, arrow, f.X+3+132-12+(11-1)/2-aw/2, f.Y+45+1, black)
}
