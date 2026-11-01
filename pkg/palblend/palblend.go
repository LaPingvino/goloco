// Package palblend implements OpenLoco's palette-map blend of the destination
// (Graphics/DrawSprite.cpp getDrawBlendOp, DrawBlendOp::transparent|dst;
// DrawSpriteHelper.hpp: dst = paletteMap[dst], skipped when that is index 0)
// on an RGBA frame: each masked pixel's palette index is recovered from its
// exact colour, mapped, and written out. No rendering dependencies.
package palblend

import (
	"image"
	"image/color"
)

// Op is one palette-map blend: a mask placed on the screen and the map to apply.
type Op struct {
	Mask   []byte          // alpha per pixel of the mask sprite, row-major
	Rect   image.Rectangle // screen-space region of the mask
	PalMap []byte          // 256-entry palette map
}

// Index maps an exact palette colour back to its index.
type Index map[[3]byte]uint8

func NewIndex(p *[256]color.RGBA) Index {
	idx := make(Index, 256)
	for i := 255; i >= 1; i-- { // lowest index wins on duplicates; 0 is transparent
		idx[[3]byte{p[i].R, p[i].G, p[i].B}] = uint8(i)
	}
	return idx
}

// Apply applies ops to frame (RGBA, w×h) and writes the mapped pixels into
// out (RGBA, same size; untouched pixels stay as they are in out). A pixel whose
// colour is not a palette colour, or whose mapped index is 0, is left alone.
func Apply(frame, out []byte, w, h int, ops []Op, pal *[256]color.RGBA, idx Index) {
	for _, op := range ops {
		mw := op.Rect.Dx()
		for my := 0; my < op.Rect.Dy(); my++ {
			y := op.Rect.Min.Y + my
			if y < 0 || y >= h {
				continue
			}
			for mx := 0; mx < mw; mx++ {
				x := op.Rect.Min.X + mx
				if x < 0 || x >= w || op.Mask[my*mw+mx] == 0 {
					continue
				}
				o := (y*w + x) * 4
				i, ok := idx[[3]byte{frame[o], frame[o+1], frame[o+2]}]
				if !ok {
					continue
				}
				m := op.PalMap[i]
				if m == 0 {
					continue
				}
				c := pal[m]
				out[o], out[o+1], out[o+2], out[o+3] = c.R, c.G, c.B, 255
				// later ops read the blended result, as upstream's in-place blit does
				frame[o], frame[o+1], frame[o+2] = c.R, c.G, c.B
			}
		}
	}
}
