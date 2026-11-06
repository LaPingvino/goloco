package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
	"strings"
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
	NoOverlay    bool   // docks / airports: the image alone (no preview-windows image)
	Company      uint8  // the player company's primary colour
	Cost         int64  // cState.stationCost: the ghost's cost (0: none, or it failed)
	// While a ghost stands: its station's name (or "New station") and the
	// cargo its catchment would accept and produce (name, inline sprite).
	Ghost              bool
	GhostName          string
	Accepted, Produced []CargoLine
}

// CargoLine is one cargo of the catchment lists.
type CargoLine struct {
	Name  string
	Image int // unitInlineSprite (-1 none)
}

// StationTabHeight is prepareDraw's height: with a ghost, 11 more per cargo
// line (at least one each, for "nothing").
func StationTabHeight(st StationTabState) int {
	if !st.Ghost {
		return StationTabH
	}
	return StationTabH + (max(1, len(st.Accepted))+max(1, len(st.Produced)))*11
}

// DrawStationTab is Station::draw (0x0049DE40) without a ghost: the station
// dropdown (3, 45), the preview at the image widget (35, 60): the station
// recoloured to the company colour, then its windows drawn translucent in
// getTranslucent(company) (ExtColour 0x2E when not recolourable); and the
// separator at the image's bottom + 16.
// The cost line: build_cost centred at x + 69, the image's bottom + 4, when
// the ghost has a cost.
// PARTIAL: airports and docks; with a ghost placed, the
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
			if sp, err := gfx8.Decode(&g1.Elements[st.Image+1]); err == nil && winMap != nil && !st.NoOverlay {
				gfx8.DrawSpriteNoise(rt, sp, imgL, imgT, gfx8.OpTransparent|gfx8.OpDst, winMap, nil)
			}
		}
		// drawRectInset(x + 3, image.bottom + 16, width − 4, 1, borderInset)
		y := imgB + 16
		gfx8.FillRectInset(rt, 3, y, 3+(ConstructionW-4)-1, y, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillNone)
	})
	DrawTextBold(screen, st.Name, f.X+3+1, f.Y+45+1, black)
	if st.Ghost { // the station's name (new_station_buffer) and the catchment lists
		nw, _ := MeasureTextBold(st.GhostName)
		DrawTextBold(screen, st.GhostName, f.X+69-nw/2, f.Y+imgB+18, black)
		y := imgB + 29
		list := func(title string, lines []CargoLine) {
			// drawStringLeft: {NEWLINE} starts a line below ("Catchment
			// Area{NEWLINE}Accepts:"); the list starts 11 below the last
			parts := strings.Split(title, "{NEWLINE}")
			for i, p := range parts {
				DrawTextBold(screen, p, f.X+2, f.Y+y+i*10, black)
			}
			y += (len(parts)-1)*10 + 11
			if len(lines) == 0 {
				DrawTextBold(screen, i18n.Label(i18n.CatchmentAreaNothing), f.X+14, f.Y+y, black)
				y += 11
				return
			}
			for _, c := range lines {
				if c.Image >= 0 && c.Image < len(g1.Elements) {
					if sp, err := gfx8.Decode(&g1.Elements[c.Image]); err == nil {
						draw8(screen, f.X+14, f.Y+y, 12, 11, func(rt *gfx8.RenderTarget) {
							gfx8.DrawSpriteNoise(rt, sp, 0, 0, gfx8.OpFor(sp, false, false), nil, nil)
						})
					}
				}
				DrawTextBold(screen, c.Name, f.X+14+12, f.Y+y+1, black)
				y += 11
			}
		}
		list(i18n.Label(i18n.CatchmentAreaAccepts), st.Accepted)
		list(i18n.Label(i18n.CatchmentAreaProduces), st.Produced)
	}
	if st.Cost != 0 { // "{COLOUR WINDOW_2}Cost: {COLOUR BLACK}{CURRENCY32}", drawStringCentred
		text := pal[gfx8.Shade(sec, 9)]
		label, amount := "Cost: ", L.Currency48(st.Cost)
		lw, _ := MeasureTextBold(label)
		aw, _ := MeasureTextBold(amount)
		x := f.X + 69 - (lw+aw)/2
		DrawTextBold(screen, label, x, f.Y+imgB+4, text)
		DrawTextBold(screen, amount, x+lw, f.Y+imgB+4, black)
	}
	arrow := string(rune(170))
	aw, _ := MeasureTextBold(arrow)
	DrawTextBold(screen, arrow, f.X+3+132-12+(11-1)/2-aw/2, f.Y+45+1, black)
}
