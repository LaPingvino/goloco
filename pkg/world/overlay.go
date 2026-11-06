package world

import (
	"fmt"
	"image"
	"image/color"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// overlay.go — the label/feedback overlay pass drawn on top of the composited
// world view: station name plaques, town name labels, and rising income
// ("+£N") money floats.
//
// This pass draws directly to the screen after all tile/vehicle passes, so it
// can freely use the ui bitmap-text helpers. (pkg/ui does not import pkg/world,
// so importing ui here creates no cycle.)
//
// OpenLoco reference: station/town name plaques are drawn from
// src/OpenLoco/src/Ui/Windows/Main.cpp (drawStationNames / drawTownNames);
// money floats mirror src/OpenLoco/src/Effects/MoneyEffect.cpp.

// ── Money floats ────────────────────────────────────────────────────────────

const (
	// moneyFloatLifeTicks is how long a "+£N" float lives before expiring.
	moneyFloatLifeTicks = 90
	// moneyFloatRisePerTick is how many screen pixels the float rises each tick.
	moneyFloatRisePerTick = 0.5
	// moneyFloatFadeTicks is the number of ticks over which the float fades out
	// at the end of its life.
	moneyFloatFadeTicks = 30
)

// moneyFloat is one rising "+£N" income indicator anchored to a world tile.
type moneyFloat struct {
	tileX, tileY int
	baseZ        uint8
	amount       int64
	age          int // ticks since spawn
}

// SpawnMoneyFloat creates a rising "+£amount" float anchored above tile
// (tileX,tileY) at height baseZ. Nothing spawns one since the invented
// delivery loop was retired; upstream's MoneyEffect (on payment) is unported.
func (w *World) SpawnMoneyFloat(tileX, tileY int, baseZ uint8, amount int64) {
	if amount <= 0 {
		return
	}
	w.moneyFloats = append(w.moneyFloats, moneyFloat{
		tileX: tileX, tileY: tileY, baseZ: baseZ, amount: amount,
	})
}

// tickMoneyFloats ages every float by one tick and drops expired ones. Called
// from TickVehicles (gameplay-only, per sim step).
func (w *World) tickMoneyFloats() {
	if len(w.moneyFloats) == 0 {
		return
	}
	live := w.moneyFloats[:0]
	for _, m := range w.moneyFloats {
		m.age++
		if m.age < moneyFloatLifeTicks {
			live = append(live, m)
		}
	}
	w.moneyFloats = live
}

// ── Overlay draw pass ───────────────────────────────────────────────────────

// drawOverlay renders station plaques, town labels and money floats over the
// composited world. Gated to zoom 0-1 (labels are illegible when zoomed out).
func (w *World) drawOverlay(screen *ebiten.Image) {
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()
	if !w.TitleMode { // Viewport::paint: stations (up to stationNamesMinScale), then towns
		if w.zoom <= stationNamesMinScale {
			w.drawStationNames(screen, sw, sh)
		}
		w.drawTownNames(screen, sw, sh)
	}
	if w.zoom > 1 {
		return
	}

	// Money floats: green "+£N" rising and fading near end of life.
	for _, m := range w.moneyFloats {
		sx, sy := w.TileToScreenPx(m.tileX, m.tileY)
		if sx < -200 || sx > sw+200 || sy < -200 || sy > sh+100 {
			continue
		}
		label := fmt.Sprintf("+£%d", m.amount)
		ui.MarkForeign(image.Rect(sx-60, sy-60, sx+60, sy+10)) // the rising float (draw8's screen copy)
		lw, _ := ui.MeasureText(label)
		ty := sy - 18 - int(float64(m.age)*moneyFloatRisePerTick)
		// Fade the last moneyFloatFadeTicks of life.
		alpha := 1.0
		if rem := moneyFloatLifeTicks - m.age; rem < moneyFloatFadeTicks {
			alpha = float64(rem) / float64(moneyFloatFadeTicks)
		}
		green := scaleColorPremul(color.RGBA{60, 220, 60, 255}, alpha)
		outline := scaleColorPremul(color.RGBA{0, 40, 0, 255}, alpha)
		drawOutlinedLabelColored(screen, label, sx-lw/2, ty, green, outline)
	}
}

// drawStationPlaque draws a filled orange plaque with a 1px dark border and
// white text, horizontally centred on sx and floating just above sy (the
// station tile's top corner).
func drawStationPlaque(screen *ebiten.Image, name string, sx, sy int) {
	tw, _ := ui.MeasureText(name)
	const padX, boxH = 4, 14
	boxW := tw + padX*2
	boxX := sx - boxW/2
	boxY := sy - 8 - boxH // hover above the tile top corner

	border := color.RGBA{40, 22, 4, 255}
	orange := color.RGBA{196, 100, 16, 255}
	// 1px dark border via a slightly larger backing rect, then the fill.
	fillRectOverlay(screen, boxX-1, boxY-1, boxW+2, boxH+2, border)
	fillRectOverlay(screen, boxX, boxY, boxW, boxH, orange)
	ui.DrawText(screen, name, boxX+padX, boxY+(boxH-10)/2, color.RGBA{255, 255, 255, 255})
}

// drawOutlinedLabel draws text with a 1px dark outline (8-way) for legibility
// over arbitrary terrain.
func drawOutlinedLabel(screen *ebiten.Image, s string, x, y int, clr color.Color) {
	drawOutlinedLabelColored(screen, s, x, y, clr, color.RGBA{0, 0, 0, 200})
}

// drawOutlinedLabelColored is drawOutlinedLabel with an explicit outline colour.
func drawOutlinedLabelColored(screen *ebiten.Image, s string, x, y int, clr, outline color.Color) {
	for _, d := range [][2]int{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}} {
		ui.DrawText(screen, s, x+d[0], y+d[1], outline)
	}
	ui.DrawText(screen, s, x, y, clr)
}

// fillRectOverlay fills an integer-pixel rectangle on screen (no anti-alias).
func fillRectOverlay(dst *ebiten.Image, x, y, w, h int, c color.Color) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

// scaleColorPremul returns c scaled by alpha factor f in premultiplied form, so
// ebiten's ColorScale.ScaleWithColor composites it as a faded, source-over draw.
func scaleColorPremul(c color.RGBA, f float64) color.RGBA {
	if f < 0 {
		f = 0
	} else if f > 1 {
		f = 1
	}
	return color.RGBA{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
		A: uint8(float64(c.A) * f),
	}
}

// kZoomToTownFonts (Town.cpp): medium bold at zoom 0 and 1, medium normal at
// 2 and 3.
var kZoomToTownFonts = [4]int{ui.G1FontMediumBold, ui.G1FontMediumBold, ui.G1FontMediumNormal, ui.G1FontMediumNormal}

// drawTownNames is Viewport.cpp drawTownNames with Town::drawLabel: in the
// unzoomed render target (the view's position >> zoom), every town whose
// saved labelFrame at this zoom overlaps the screen gets its name at the
// frame's left + 1, top + 1, in that zoom's font, white with the outline.
// PARTIAL: the frames are the save's — a town that grows or is renamed keeps
// its old frame until Town::updateLabel is ported.
func (w *World) drawTownNames(screen *ebiten.Image, sw, sh int) {
	z := max(0, min(3, w.zoom))
	ox, oy := int(w.camX)>>z, int(w.camY)>>z
	white, black := ui.TextColours(colourWhite)
	for i := range w.Map.Towns {
		t := &w.Map.Towns[i]
		f := t.LabelFrame
		if !f.Contains(ox, oy, ox+sw-1, oy+sh-1, z) {
			continue
		}
		x, y := int(f.Left[z])+1-ox, int(f.Top[z])+1-oy
		name := w.Map.TownName(t.Name, i)
		tw := ui.DrawTextOutlinedFont(screen, name, kZoomToTownFonts[z], x, y, white, black)
		ui.MarkForeign(image.Rect(x-1, y-1, x+tw+1, y+12)) // draw8's screen copy
	}
}

// colourWhite is Colour::white (ui's too).
const colourWhite = 2

// stationNamesMinScale is Config::stationNamesMinScale (default 2).
const stationNamesMinScale = 2

// kZoomToStationBorder (Station.cpp): the curved ends' images and the end
// width, by zoom; kZoomToStationFonts: medium bold at 0–1, small at 2–3.
var kZoomToStationBorder = [4]struct{ left, right, width int }{
	{2315, 2316, 3}, {2315, 2316, 3}, {2319, 2320, 1}, {2319, 2320, 1},
}
var kZoomToStationFonts = [4]int{ui.G1FontMediumBold, ui.G1FontMediumBold, ui.G1FontSmall, ui.G1FontSmall}

// labelIcons is getTransportIconsFromStationFlags: label_icons_none … by the
// four transport-mode bits.
const labelIconsNone = 230

// drawStationNames is Viewport.cpp drawStationNames with Station.cpp
// drawStationName: every station but those not fully created (flag_5)
// whose saved labelFrame at this zoom overlaps the screen — the curved
// ends and the middle each drawn twice translucent (ExtColour::unk34, then
// the owner's colour, grey for none, getTranslucent(…, 1)), then in black
// the name with its town and the transport icons. PARTIAL: the hovered
// station's look (getTranslucent(…, 0)), Station::updateLabel (frames are the
// save's).
func (w *World) drawStationNames(screen *ebiten.Image, sw, sh int) {
	z := max(0, min(3, w.zoom))
	ox, oy := int(w.camX)>>z, int(w.camY)>>z
	g1 := w.renderer.G1
	if g1 == nil {
		return
	}
	m34, err := g1.GetPaletteMap(0x34)
	if err != nil {
		return
	}
	b := kZoomToStationBorder[z]
	pal := g1.GetPalette()
	for id := range w.Map.StationNames {
		st := &w.Map.StationNames[id]
		f := st.LabelFrame
		if st.Name == 0 || st.Flags&(1<<5) != 0 || !f.Contains(ox, oy, ox+sw-1, oy+sh-1, z) {
			continue
		}
		colour := 1 // Colour::grey
		if st.Owner != 0xFF && int(st.Owner) < len(w.CompanyColours) {
			colour = int(w.CompanyColours[st.Owner])
		}
		mc, err := g1.GetPaletteMap(int(paint.TranslucentColour[colour%len(paint.TranslucentColour)][1]))
		if err != nil {
			continue
		}
		l, t, r, btm := int(f.Left[z])-ox, int(f.Top[z])-oy, int(f.Right[z])-ox, int(f.Bottom[z])-oy
		wd, ht := r-l+1, btm-t+1
		left, lerr := gfx8.Decode(&g1.Elements[b.left])
		right, rerr := gfx8.Decode(&g1.Elements[b.right])
		drawLabelBox := func(rt *gfx8.RenderTarget) {
			for _, m := range [][]uint8{m34, mc} {
				if lerr == nil {
					gfx8.DrawSpriteNoise(rt, left, 0, 0, gfx8.OpFor(left, false, true), m, nil)
				}
				if rerr == nil {
					gfx8.DrawSpriteNoise(rt, right, wd-b.width, 0, gfx8.OpFor(right, false, true), m, nil)
				}
				gfx8.FillRectTranslucent(rt, b.width+1, 0, b.width+1+wd-2*b.width-1, ht-1, m)
			}
		}
		ui.DrawRect8(screen, l, t, wd, ht, drawLabelBox)
		name := w.Map.StationName(id)
		icons := i18n.Str(labelIconsNone + int(st.Flags&0xF))
		ui.DrawTextFont(screen, name+" "+icons, kZoomToStationFonts[z], l+b.width, t, pal[paletteBlack0])
		ui.MarkForeign(image.Rect(l, t, l+wd, t+ht))
	}
}

// paletteBlack0 is PaletteIndex::black0.
const paletteBlack0 = 10
