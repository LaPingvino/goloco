package gfx8

import (
	"bytes"
	"math/rand/v2"
	"testing"
)

// The drawing loops as they were before they were specialised (2026-11-04),
// kept as the reference the faster ones must match byte for byte.

func refDrawSpriteNoise(rt *RenderTarget, s *Sprite, x, y int, op BlendOp, palMap []uint8, noise *Sprite) {
	if noise != nil && !s.RLE {
		op |= OpNoiseMask
	} else {
		op &^= OpNoiseMask
	}
	x0, y0 := x+s.XOffset, y+s.YOffset
	// Clip the sprite rectangle to the target once (same pixels as testing every
	// pixel against the bounds).
	sxMin, sxMax := max(0, -x0), min(s.Width, rt.Width-x0)
	syMin, syMax := max(0, -y0), min(s.Height, rt.Height-y0)
	if sxMin >= sxMax || syMin >= syMax {
		return
	}
	// Fast path for the commonest op, a transparent copy: blitPixel with
	// OpTransparent alone writes src wherever src != 0.
	if op == OpTransparent {
		for sy := syMin; sy < syMax; sy++ {
			src := s.Pix[sy*s.Width+sxMin : sy*s.Width+sxMax]
			dst := rt.Bits[(y0+sy)*rt.Width+x0+sxMin:]
			for i, p := range src {
				if p != Transparent {
					dst[i] = p
				}
			}
		}
		return
	}
	for sy := syMin; sy < syMax; sy++ {
		row := (y0 + sy) * rt.Width
		for sx := sxMin; sx < sxMax; sx++ {
			nm := uint8(0xFF)
			if op&OpNoiseMask != 0 {
				if i := s.Width*sy + sx; i < len(noise.Pix) {
					nm = noise.Pix[i]
				} else {
					nm = 0
				}
			}
			blitPixel(op, s.Pix[sy*s.Width+sx], &rt.Bits[row+x0+sx], palMap, nm)
		}
	}
}

func refDrawSpriteMasked(rt *RenderTarget, img, mask *Sprite, x, y int) {
	x0, y0 := x+img.XOffset, y+img.YOffset
	for sy := 0; sy < img.Height; sy++ {
		ty := y0 + sy
		if ty < 0 || ty >= rt.Height {
			continue
		}
		for sx := 0; sx < img.Width; sx++ {
			tx := x0 + sx
			if tx < 0 || tx >= rt.Width {
				continue
			}
			i := sy*img.Width + sx
			if i >= len(mask.Pix) {
				continue
			}
			if v := mask.Pix[i] & img.Pix[i]; v != 0 {
				rt.Bits[ty*rt.Width+tx] = v
			}
		}
	}
}

func TestFastDrawMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewPCG(4, 11))
	sprite := func() *Sprite {
		w, h := 1+rng.IntN(40), 1+rng.IntN(40)
		s := &Sprite{Width: w, Height: h, XOffset: -rng.IntN(20), YOffset: -rng.IntN(20), Pix: make([]uint8, w*h), HasTransparency: true}
		for i := range s.Pix {
			if rng.IntN(3) > 0 {
				s.Pix[i] = uint8(rng.IntN(256))
			}
		}
		if rng.IntN(2) == 0 {
			s.spans = rowSpans(s) // as Decode makes them
		}
		return s
	}
	ops := []BlendOp{OpNone, OpTransparent, OpTransparent | OpSrc, OpTransparent | OpDst, OpTransparent | OpSrc | OpDst, OpSrc, OpDst}
	for n := 0; n < 20000; n++ {
		w, h := 30+rng.IntN(40), 30+rng.IntN(40)
		a, b := &RenderTarget{Width: w, Height: h, Bits: make([]uint8, w*h)}, &RenderTarget{Width: w, Height: h, Bits: make([]uint8, w*h)}
		for i := range a.Bits {
			a.Bits[i] = uint8(rng.IntN(256))
		}
		copy(b.Bits, a.Bits)
		s := sprite()
		x, y := rng.IntN(w+40)-20, rng.IntN(h+40)-20
		if n%3 == 0 {
			m := sprite()
			m.Width, m.Height = s.Width, s.Height
			m.Pix = m.Pix[:min(len(m.Pix), len(s.Pix)-rng.IntN(len(s.Pix)/2+1))]
			refDrawSpriteMasked(a, s, m, x, y)
			DrawSpriteMasked(b, s, m, x, y)
		} else {
			op := ops[rng.IntN(len(ops))]
			pal := make([]uint8, 256*(1+rng.IntN(3)))
			for i := range pal {
				pal[i] = uint8(rng.IntN(256))
				if rng.IntN(8) == 0 {
					pal[i] = 0
				}
			}
			if op&OpSrc != 0 && op&OpDst != 0 || op&OpTransparent == 0 && op != OpNone {
				for i := range s.Pix { // blitPixel's src-1 row needs src ≥ 1 there (as upstream's data)
					if s.Pix[i] == 0 {
						s.Pix[i] = 1
					}
				}
				if s.spans != nil {
					s.spans = rowSpans(s)
				}
			}
			var noise *Sprite
			if n%5 == 0 {
				noise = sprite()
			}
			refDrawSpriteNoise(a, s, x, y, op, pal, noise)
			DrawSpriteNoise(b, s, x, y, op, pal, noise)
		}
		if !bytes.Equal(a.Bits, b.Bits) {
			t.Fatalf("case %d differs", n)
		}
	}
}
