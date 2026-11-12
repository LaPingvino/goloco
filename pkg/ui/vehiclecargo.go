package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/objects"
)

// The Vehicle window's Cargo tab (Ui/Windows/Vehicle.cpp Cargo: prepareDraw,
// draw 0x004B3F0D, drawScroll 0x004B3F62), item 5P.g.

// VehicleCargoMinSize is Cargo::kMinWindowSize.
var VehicleCargoMinSize = [2]int{192, 142}

const (
	vcRefit             = vwTabRoute + 1 // Cargo::widx::refit (9)
	vcCargoList         = vwTabRoute + 2 // cargoList (10)
	vcWidgets           = vwTabRoute + 3
	imgRefitCargoButton = 2386 // ImageIds::refit_cargo_button
)

// VehicleCargoRow is one car's row: its picture and its cargo lines — each
// cargo with a quantity as "{unit} from {station}" (cargo_from), or "Empty"
// alone (Empty: drawn 5 lower); none at all when the body carries nothing.
type VehicleCargoRow struct {
	Car   TrainCarDraw
	Lines []string
}

// VehicleCargo is what Cargo::draw prints: the totals (total_stringid:
// generateCargoTotalString) and capacities (vehicle_capacity_stringid), the
// rows, and whether the refit button shows (canRefit).
type VehicleCargo struct {
	Total, Capacity string
	Rows            []VehicleCargoRow
	CanRefit        bool
}

// VehicleCargoLayout is Cargo::widgets after prepareDraw: the common widgets
// to the size, the cargo list to 26 from the right and 27 from the bottom
// (22 wider without the refit button), the refit button on the right bar.
func VehicleCargoLayout(width, height int, hiddenTabs uint8, canRefit bool) (w [vcWidgets]BVWidget) {
	vehicleCommonWidgets(w[:])
	w[vcRefit] = bvW(240, 44, 24, 24)
	w[vcCargoList] = bvW(3, 44, 259, 120)
	w[vwFrame].R, w[vwFrame].B = width-1, height-1
	w[vwPanel].R, w[vwPanel].B = width-1, height-1
	w[vwCaption].R = width - 2
	w[vwClose].L, w[vwClose].R = width-15, width-3
	w[vcCargoList].R, w[vcCargoList].B = width-26, height-27
	w[vcRefit].L, w[vcRefit].R = width-25, width-2
	if !canRefit {
		w[vcRefit].Hidden = true
		w[vcCargoList].R = width - 26 + 22
	}
	leftAlignTabs(w[:], hiddenTabs)
	return w
}

// carInlineWidth is getWidthVehicleInline: the car's screen length.
func carInlineWidth(c TrainCarDraw) int {
	if c.Object == nil {
		return 0
	}
	_, total := c.Object.CarDrawItems(40, c.Components, c.Reversed, true, c.Brake)
	return int(objects.ScreenDistance(total, 40))
}

// drawCargo is Cargo::draw and drawScroll: the frame and tabs (Cargo
// active), the refit button, the list — each car drawn to end at x 24, its
// cargo lines from there —, then the total 25 and the capacity 13 above the
// bottom. PARTIAL: the hovered row's highlight, drawStringLeftClipped's
// clipping of the two bottom lines.
func (w *VehicleWindow) drawCargo(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	c := &w.Cargo
	lw := VehicleCargoLayout(w.W, w.H, w.State.HiddenTabs, c.CanRefit)
	rowH := vehicleRowHeights[max(0, min(5, w.VehicleType))]
	list := lw[vcCargoList]
	const cargoX = 24 // kCargoXPos
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	var cx, cy int
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[vwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[vwClose].L, lw[vwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawVehicleTabs(rt, lw[:], vwTabCargo, sec, prim)
		if b := lw[vcRefit]; !b.Hidden {
			if sp, err := gfx8.Decode(&g1.Elements[imgRefitCargoButton]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, b.L, b.T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			}
		}
		var cw, ch int
		cx, cy, cw, ch, _ = drawVScrollViewPressed(rt, list.L, list.T, list.R-list.L+1, list.B-list.T+1, gfx8.InsetColour{Map: sec}, len(c.Rows)*rowH, w.Scroll, w.hold.PressedPart())
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 4)) // clearSingle(getShade(secondary, 4))
		if cw <= 0 || ch <= 0 {
			return
		}
		sub := gfx8.NewRenderTarget(cw, ch)
		for yy := range ch {
			copy(sub.Bits[yy*cw:(yy+1)*cw], rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw])
		}
		w.drawRowHover(sub, rowH, cw, false)
		for i, r := range c.Rows {
			if r.Car.Object == nil {
				continue
			}
			x, y := cargoX-carInlineWidth(r.Car), i*rowH-w.Scroll+(rowH-22)/2
			drawCarInline(sub, r.Car, x, y, false, nil)
			drawCarInline(sub, r.Car, x, y, true, nil)
		}
		for yy := range ch {
			copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
		}
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, w.Title, int(w.TitleColour))
	pal := g1.GetPalette()
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	sx, sy := w.X+cx, w.Y+cy
	clip := screen.SubImage(image.Rect(sx, sy, w.X+list.R, w.Y+list.B)).(*ebiten.Image)
	for i, r := range c.Rows {
		y := sy + i*rowH - w.Scroll + rowH/2 - (rowH-22)/2 - 10
		if len(r.Lines) == 1 {
			y += 5 // one cargo (or Empty): centred
		}
		col := black
		if i == w.Hover { // strFormat = wcolour2_stringid
			col = win2
		}
		for _, l := range r.Lines {
			drawFormatted(clip, sx+cargoX, y, l, col, col)
			y += 10
		}
	}
	drawFormatted(screen, w.X+3, w.Y+w.H-25, c.Total, win2, black)
	drawFormatted(screen, w.X+3, w.Y+w.H-13, c.Capacity, win2, black)
}
