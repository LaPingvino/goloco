package world

import (
	"github.com/LaPingvino/goloco/pkg/sim"
	"image"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
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
// (Money: upstream's MoneyEffect entity is unported — LOOP item ME.)

// drawOverlay renders station plaques, town labels and money floats over the
// composited world. Gated to zoom 0-1 (labels are illegible when zoomed out).
func (w *World) drawOverlay(screen *ebiten.Image) {
	if !w.TitleMode {
		z := max(0, min(3, w.zoom))
		w.drawNames(screen, int(w.camX)>>z, int(w.camY)>>z, z, true)
	}
	w.drawMoneyEffects(screen)
}

// drawNames is Viewport::paint's names for a view whose top-left is (ox, oy)
// in the zoom's unzoomed pixels, drawn on dst (its Bounds: the view's screen
// rectangle, which clips them): the station names up to stationNamesMinScale
// unless hidden (ViewportFlags::hideStationNames), then the town names.
func (w *World) drawNames(dst *ebiten.Image, ox, oy, z int, stations bool) {
	if stations && z <= stationNamesMinScale {
		w.drawStationNames(dst, ox, oy, z)
	}
	w.drawTownNames(dst, ox, oy, z)
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
func (w *World) drawTownNames(screen *ebiten.Image, ox, oy, z int) {
	b := screen.Bounds()
	sw, sh := b.Dx(), b.Dy()
	ox, oy = ox-b.Min.X, oy-b.Min.Y // view pixels → dst's screen coordinates
	white, black := ui.TextColours(colourWhite)
	for i := range w.Map.Towns {
		t := &w.Map.Towns[i]
		f := t.LabelFrame
		if !f.Contains(ox+b.Min.X, oy+b.Min.Y, ox+b.Min.X+sw-1, oy+b.Min.Y+sh-1, z) {
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
func (w *World) drawStationNames(screen *ebiten.Image, ox, oy, z int) {
	bd := screen.Bounds()
	vl, vt, vr, vb := ox, oy, ox+bd.Dx()-1, oy+bd.Dy()-1 // the view's rectangle in its pixels
	ox, oy = ox-bd.Min.X, oy-bd.Min.Y                    // view pixels → dst's screen coordinates
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
		if st.Name == 0xFFFF || st.Flags&(1<<5) != 0 || !f.Contains(vl, vt, vr, vb, z) {
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

// kWiggleYOffsets is PaintEffectEntity.cpp's (440 entries): the same 20-step
// wave repeated, each glyph of a money popup offset by it.
var kWiggleYOffsets = func() [440]int8 {
	wave := [20]int8{0, 1, 2, 2, 3, 3, 3, 3, 2, 2, 1, 0, -1, -2, -2, -3, -3, -2, -2, -1}
	var o [440]int8
	for i := range o {
		o[i] = wave[i%20]
	}
	return o
}()

// drawMoneyEffects is paintWindowCurrencyEntity through the string plot list
// (PaintSession::addToStringPlotList / drawStringStructs): each money effect,
// up to zoom 1 (Config cashPopupRendering, on by default), "+ amount" /
// "- amount" (format_currency_income / expense_in_company_colour) in the
// company's colour (WINDOW_1: shade 7, outlined), centred on its position
// (offsetX: minus half the medium bold width), each glyph raised by
// kWiggleYOffsets from the effect's wiggle; medium bold at zoom 0, small at 1.
// PARTIAL: the redGreen subtype (made by nothing in OpenLoco) is not painted;
// drawn after the town names rather than interleaved with the scene's paint.
func (w *World) drawMoneyEffects(screen *ebiten.Image) {
	if len(w.Map.MoneyEffects) == 0 || w.zoom > 1 || w.Map.EntityLists == nil {
		return
	}
	z := w.zoom
	font := ui.G1FontMediumBold
	if z == 1 {
		font = ui.G1FontSmall
	}
	l := w.Map.EntityLists
	for id := l.Heads[sim.EntityListMisc]; id != 0xFFFF; id = l.Links[id].Next {
		e := w.Map.MoneyEffects[id]
		if e == nil || e.SubType != 2 {
			continue
		}
		amount := int64(e.Amount)
		text := "+ " + ui.FormatCurrency(amount)
		if amount < 0 {
			text = "- " + ui.FormatCurrency(-amount)
		}
		offsetX := -ui.MeasureTextFont(text, ui.G1FontMediumBold) / 2
		vx, vy := w.gameToView(int(e.X), int(e.Y), int(e.Z))
		sx := (vx+offsetX)>>z - int(w.camX)>>z
		sy := vy>>z - int(w.camY)>>z
		colour := 0
		if int(e.Company) < len(w.CompanyColours) {
			colour = int(w.CompanyColours[e.Company])
		}
		body, outline := ui.WindowColour1Text(colour)
		tw := ui.DrawTextYOffsets(screen, text, font, sx, sy, kWiggleYOffsets[e.Wiggle:], body, outline)
		ui.MarkForeign(image.Rect(sx-1, sy-5, sx+tw+1, sy+16)) // draw8's screen copy
	}
}
