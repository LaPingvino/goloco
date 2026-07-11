package world

import (
	"fmt"
	"image/color"

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

	// passengerFare mirrors game.cargoFare(CargoSlotPass) — £6 per passenger
	// unit. Duplicated here (pkg/world does not import pkg/game) purely to label
	// the money float; the authoritative credit still runs through
	// game.CreditDelivery via OnCargoDelivered.
	passengerFare int64 = 6
)

// moneyFloat is one rising "+£N" income indicator anchored to a world tile.
type moneyFloat struct {
	tileX, tileY int
	baseZ        uint8
	amount       int64
	age          int // ticks since spawn
}

// SpawnMoneyFloat creates a rising "+£amount" float anchored above tile
// (tileX,tileY) at height baseZ. Called from the delivery path in station.go.
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
	if w.zoom > 1 {
		return
	}
	sw := screen.Bounds().Dx()
	sh := screen.Bounds().Dy()

	// Station name plaques: orange box, 1px dark border, white 10px text,
	// centred above the station's origin tile.
	for i, st := range w.StationList() {
		sx, sy := w.TileToScreenPx(st.TileX, st.TileY)
		if sx < -200 || sx > sw+200 || sy < -100 || sy > sh+100 {
			continue // cull off-screen
		}
		name := fmt.Sprintf("Station %d", i+1)
		drawStationPlaque(screen, name, sx, sy)
	}

	// Town name labels: white text with a 1px dark outline, no box.
	for _, t := range w.TownList() {
		sx, sy := w.TileToScreenPx(t.TileX, t.TileY)
		if sx < -200 || sx > sw+200 || sy < -100 || sy > sh+100 {
			continue
		}
		lw, _ := ui.MeasureText(t.Name)
		drawOutlinedLabel(screen, t.Name, sx-lw/2, sy-6, color.RGBA{255, 255, 255, 255})
	}

	// Money floats: green "+£N" rising and fading near end of life.
	for _, m := range w.moneyFloats {
		sx, sy := w.TileToScreenPx(m.tileX, m.tileY)
		if sx < -200 || sx > sw+200 || sy < -200 || sy > sh+100 {
			continue
		}
		label := fmt.Sprintf("+£%d", m.amount)
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
