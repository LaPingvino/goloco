package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// OverheadTabH is the overhead tab's window height (makeCommonWidgets(138, 192)).
const OverheadTabH = 192

// OverheadTabState is what the overhead tab shows (Construction/OverheadTab.cpp).
type OverheadTabState struct {
	// Names are the available mods' names (modList; "" = none, its checkbox
	// hidden): up to 4 for track, 2 for road.
	Names [4]string
	// Selected is lastSelectedMods (bit i: checkbox i ticked).
	Selected uint8
}

// DrawOverheadTab is Overhead::prepareDraw + draw (0x0049EA3E) after the
// frame: a checkbox per available mod (3, 45 + 12·i, 132×12), ticked by
// lastSelectedMods; with a mod ticked, the "Single section" dropdown at
// (3, 95) and the preview area (Wt3, 35,110 66×66).
// PARTIAL: the preview (drawTrack/drawRoad with the mods), the cost line,
// the dropdown list, upgrading track by clicking it (the tool).
func DrawOverheadTab(screen *ebiten.Image, f *ConstructionFrame, st OverheadTabState) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(f.SkinColour))
	if err != nil {
		return
	}
	pal := g1.GetPalette()
	text := pal[gfx8.Shade(sec, 9)] // the label's colour: the window colour, opaque
	black := pal[paletteBlack0]
	any := st.Selected&0xF != 0
	draw8(screen, f.X, f.Y, ConstructionW, f.Height(), func(rt *gfx8.RenderTarget) {
		for i, n := range st.Names {
			if n != "" {
				y := 45 + 12*i
				gfx8.FillRectInset(rt, 3, y, 3+10-1, y+10-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			}
		}
		if any {
			optDropdown(rt, 3, 95, 132, sec)
			gfx8.FillRectInset(rt, 35, 110, 35+66-1, 110+66-1, gfx8.InsetColour{Map: sec}, 0)
		}
	})
	tick := string(rune(0xAC))
	for i, n := range st.Names {
		if n == "" {
			continue
		}
		y := 45 + 12*i
		if st.Selected&(1<<i) != 0 {
			DrawTextBold(screen, tick, f.X+3, f.Y+y, text)
		}
		DrawTextBold(screen, n, f.X+3+10+4, f.Y+y, text)
	}
	if any {
		DrawTextBold(screen, "Single section", f.X+3+1, f.Y+95+1, black) // modString[lastSelectedTrackModSection]
		arrow := string(rune(170))
		aw, _ := MeasureTextBold(arrow)
		DrawTextBold(screen, arrow, f.X+3+132-12+(11-1)/2-aw/2, f.Y+95+1, black)
	}
}

// OverheadHit returns the checkbox index (0–3) under (rx, ry), or -1.
func OverheadHit(rx, ry int, st OverheadTabState) int {
	for i, n := range st.Names {
		y := 45 + 12*i
		if n != "" && rx >= 3 && rx < 3+132 && ry >= y && ry < y+12 {
			return i
		}
	}
	return -1
}
