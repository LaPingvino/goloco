package ui

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// ViewLabels draws a viewport's names over its picture: dst is the screen
// clipped to the viewport (its Bounds are the viewport's screen rectangle).
// Drawn on the screen rather than into the 8-bit picture, so the names take
// the same text path as the rest of the interface (the TTF fallback too).
type ViewLabels func(dst *ebiten.Image)

// drawViewport is a window's viewport widget: the scene's picture rt at
// (x, y) (Viewport::paint's frame), the station and town names over it
// (labels; nil none), then the centre-view button on top at screen (cx, cy)
// in the colour map cmap (a widget, drawn after the viewport; centre false:
// none). img keeps the window's texture between frames.
func drawViewport(screen *ebiten.Image, img **ebiten.Image, rt *gfx8.RenderTarget, x, y int, labels ViewLabels, centre bool, cx, cy int, cmap []uint8) {
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	pix := make([]byte, rt.Width*rt.Height*4)
	for k, ix := range rt.Bits[:rt.Width*rt.Height] {
		c := pal[ix]
		pix[k*4], pix[k*4+1], pix[k*4+2], pix[k*4+3] = c.R, c.G, c.B, 255
	}
	if *img == nil || (*img).Bounds().Dx() != rt.Width || (*img).Bounds().Dy() != rt.Height {
		*img = ebiten.NewImage(rt.Width, rt.Height)
	}
	(*img).WritePixels(pix)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(*img, op)
	r := image.Rect(x, y, x+rt.Width, y+rt.Height)
	MarkForeign(r)
	clip := screen.SubImage(r).(*ebiten.Image)
	if labels != nil {
		labels(clip)
	}
	if !centre {
		return
	}
	if sp, err := gfx8.Decode(&g1.Elements[imgCentreView]); err == nil {
		draw8(clip, cx, cy, sp.Width, sp.Height, func(rt *gfx8.RenderTarget) {
			gfx8.DrawSpriteNoise(rt, sp, 0, 0, gfx8.OpTransparent|gfx8.OpSrc, cmap, nil)
		})
	}
}
