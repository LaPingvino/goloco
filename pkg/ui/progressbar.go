package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// ProgressBar.cpp: a 350×47 window centred on the screen (y at least 28),
// lighter black frame, white caption, black panel, and in the clip rect
// (2, 17, w-5, h-19) the track image with a train sliding in from the left
// at value - 255. The train alternates between two styles on each close.
const (
	progW, progH = 350, 47

	imgProgressTrack       = 2330
	imgProgressStyle0Frame = 2326 // ..2329
	imgProgressStyle1Frame = 2331 // ..2334
)

// DrawProgressBar draws the progress window for a caption at value 0..255.
func DrawProgressBar(screen *ebiten.Image, caption string, value uint8, style int) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	x := sw/2 - progW/2
	y := max(28, sh/2-progH/2)
	black, err := g1.GetPaletteMap(0)
	if err != nil {
		return
	}
	DrawWindowFrame(screen, x, y, progW, progH, 0, true)
	draw8(screen, x, y, progW, progH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 15, progW-1, progH-1, gfx8.InsetColour{Map: black}, 0) // Panel, secondary = black
		// clipRenderTarget(x+2, y+17, w-5, h-19): draw into a target of the clip
		// size seeded from the window, then copy it back.
		const cx, cy, cw, ch = 2, 17, progW - 5, progH - 19
		clip := gfx8.NewRenderTarget(cw, ch)
		for r := range ch {
			copy(clip.Bits[r*cw:(r+1)*cw], rt.Bits[(cy+r)*rt.Width+cx:])
		}
		defer func() {
			for r := range ch {
				copy(rt.Bits[(cy+r)*rt.Width+cx:(cy+r)*rt.Width+cx+cw], clip.Bits[r*cw:(r+1)*cw])
			}
		}()
		if sp, err := gfx8.Decode(&g1.Elements[imgProgressTrack]); err == nil {
			gfx8.DrawSprite(clip, sp, 0, 0, gfx8.OpFor(sp, false, false), nil)
		}
		base, c1, c2 := imgProgressStyle0Frame, 11, 25 // mutedSeaGreen, mutedDarkRed
		if style != 0 {
			base, c1, c2 = imgProgressStyle1Frame, 0, 12 // black, mutedGrassGreen
		}
		p, err1 := g1.GetPaletteMap(c1)
		s, err2 := g1.GetPaletteMap(c2)
		if err1 != nil || err2 != nil {
			return
		}
		frame := base + int(value/4)%4
		if sp, err := gfx8.Decode(&g1.Elements[frame]); err == nil {
			gfx8.DrawSprite(clip, sp, int(value)-255, 0, gfx8.OpFor(sp, true, false), gfx8.TwoColourMap(p, s))
		}
	})
	DrawWindowCaption(screen, x+1, y+1, progW-2, caption)
}
