package palblend

import (
	"image"
	"image/color"
	"testing"
)

func TestBlendPixelsMapsThroughPaletteMap(t *testing.T) {
	var pal [256]color.RGBA
	for i := range pal {
		pal[i] = color.RGBA{uint8(i), uint8(255 - i), uint8(i / 2), 255}
	}
	idx := NewIndex(&pal)
	palMap := make([]byte, 256)
	for i := range palMap {
		palMap[i] = byte((i + 100) % 256)
	}
	palMap[7] = 0 // maps to transparent: pixel must be left alone

	const w, h = 4, 2
	frame := make([]byte, w*h*4)
	set := func(x, y, i int) {
		o := (y*w + x) * 4
		frame[o], frame[o+1], frame[o+2], frame[o+3] = pal[i].R, pal[i].G, pal[i].B, 255
	}
	set(0, 0, 10)
	set(1, 0, 7)
	set(2, 0, 20)
	set(3, 0, 30)
	op := Op{Mask: []byte{255, 255, 0, 255}, Rect: image.Rect(0, 0, 4, 1), PalMap: palMap}
	out := make([]byte, w*h*4)
	Apply(frame, out, w, h, []Op{op}, &pal, idx)

	px := func(x, y int) color.RGBA {
		o := (y*w + x) * 4
		return color.RGBA{out[o], out[o+1], out[o+2], out[o+3]}
	}
	if got := px(0, 0); got != pal[110] {
		t.Errorf("index 10 → %v, want palette[110] %v", got, pal[110])
	}
	if got := px(1, 0); got.A != 0 {
		t.Errorf("index 7 maps to 0 (transparent) and must be skipped, got %v", got)
	}
	if got := px(2, 0); got.A != 0 {
		t.Errorf("unmasked pixel was written: %v", got)
	}
	if got := px(3, 0); got != pal[130] {
		t.Errorf("index 30 → %v, want palette[130] %v", got, pal[130])
	}
}
