package gfx8

import (
	"bytes"
	"math/rand"
	"testing"
)

func upscaleWH(s *Sprite, f int) *Sprite {
	u := &Sprite{Width: s.Width * f, Height: s.Height * f, XOffset: s.XOffset * f, YOffset: s.YOffset * f, Pix: make([]uint8, s.Width*s.Height*f*f)}
	for y := range u.Height {
		for x := range u.Width {
			u.Pix[y*u.Width+x] = s.Pix[(y/f)*s.Width+x/f]
		}
	}
	return u
}

// drawMaskedZoom samples every 2^zoom-th pixel: an image and mask upscaled
// 2^z drawn at zoom z (positions on the grid) equal the originals at zoom 0.
func TestMaskedZoomDownsample(t *testing.T) {
	r := rand.New(rand.NewSource(9))
	for c := 0; c < 300; c++ {
		zoom := 1 + r.Intn(3)
		f := 1 << zoom
		img := randSpriteWH(r, 1+r.Intn(12), 1+r.Intn(12), false)
		mask := &Sprite{Width: img.Width, Height: img.Height, Pix: make([]uint8, len(img.Pix))}
		for i := range mask.Pix {
			mask.Pix[i] = uint8(r.Intn(256))
		}
		x, y := r.Intn(30)-10, r.Intn(30)-10
		ox, oy := r.Intn(9)-4, r.Intn(9)-4
		a := NewRenderTarget(20, 20)
		a.OriginX, a.OriginY = ox, oy
		DrawSpriteMasked(a, img, mask, x-ox, y-oy)
		b := NewRenderTarget(20, 20)
		b.OriginX, b.OriginY = ox*f, oy*f
		DrawSpriteMaskedZoom(b, upscaleWH(img, f), upscaleWH(mask, f), nil, x*f, y*f, zoom)
		if !bytes.Equal(a.Bits, b.Bits) {
			t.Fatalf("case %d (zoom %d, %dx%d at %d,%d, origin %d,%d): differ", c, zoom, img.Width, img.Height, x, y, ox, oy)
		}
	}
}

// The zoom-sprite recursion: image and mask with zoom sprites at zoom 1 draw
// as their half-size pair at zoom 0 (origin and position halved).
func TestMaskedZoomRecursion(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	for c := 0; c < 100; c++ {
		small := randSpriteWH(r, 1+r.Intn(10), 1+r.Intn(10), false)
		smask := &Sprite{Width: small.Width, Height: small.Height, Pix: make([]uint8, len(small.Pix))}
		for i := range smask.Pix {
			smask.Pix[i] = uint8(r.Intn(256))
		}
		small.Index, smask.Index = 3, 4
		big := upscaleWH(small, 2)
		big.HasZoomSprites, big.ZoomOffset, big.Index = true, 7, 10
		bmask := upscaleWH(smask, 2)
		bmask.HasZoomSprites, bmask.ZoomOffset, bmask.Index = true, 7, 11
		table := map[int]*Sprite{3: small, 4: smask, 10: big, 11: bmask}
		x, y, ox, oy := r.Intn(40)-10, r.Intn(40)-10, r.Intn(20)-10, r.Intn(20)-10
		a := NewRenderTarget(20, 20)
		a.OriginX, a.OriginY = ox, oy
		DrawSpriteMaskedZoom(a, big, bmask, func(id int) *Sprite { return table[id] }, x, y, 1)
		b := NewRenderTarget(20, 20)
		b.OriginX, b.OriginY = ox>>1, oy>>1
		DrawSpriteMaskedZoom(b, small, smask, nil, x>>1, y>>1, 0)
		if !bytes.Equal(a.Bits, b.Bits) {
			t.Fatalf("case %d: the recursion differs", c)
		}
	}
}
