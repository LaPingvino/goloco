package world

import (
	"github.com/LaPingvino/goloco/pkg/palblend"
	"github.com/hajimehoshi/ebiten/v2"
)

// Palette-map blending: the draw op OpenLoco uses for an image that is blended
// but has no primary colour — "only a transparency mask: just colour the
// pixels using the palette map" (Graphics/DrawSprite.cpp getDrawBlendOp,
// DrawBlendOp::transparent|dst; DrawSpriteHelper.hpp: dst = paletteMap[dst],
// skipped when the result is index 0). Water uses it with the water object's
// palette map (image +41, WaterObject::kColourPalette).
//
// goloco draws in RGBA, so the frame is read back once, each masked pixel's
// palette index is recovered from its exact colour (zoom 0 draws only palette
// colours), mapped, and written to blendBuf. Pure pixel work: blendPixels.
// PERF NOTE: one full-frame ReadPixels per frame when water is on screen.

// applyPaletteBlends runs the queued ops: reads src (the frame drawn so far)
// once, blends, and writes the result into dst (transparent elsewhere).
func (w *World) applyPaletteBlends(src, dst *ebiten.Image) {
	if len(w.blendOps) == 0 || w.renderer == nil || w.renderer.G1 == nil {
		return
	}
	b := src.Bounds()
	n := b.Dx() * b.Dy() * 4
	if len(w.blendFrame) != n {
		w.blendFrame = make([]byte, n)
		w.blendOut = make([]byte, n)
	}
	if w.palIndex == nil {
		w.palIndex = palblend.NewIndex(&w.renderer.G1.Palette)
	}
	src.ReadPixels(w.blendFrame)
	clear(w.blendOut)
	palblend.Apply(w.blendFrame, w.blendOut, b.Dx(), b.Dy(), w.blendOps, &w.renderer.G1.Palette, w.palIndex)
	dst.WritePixels(w.blendOut)
}

// maskAlpha returns the alpha channel of a sprite (cached).
func (w *World) maskAlpha(spriteID int) (alpha []byte, wd, ht int, ok bool) {
	if a, ok := w.maskCache[spriteID]; ok {
		return a.alpha, a.w, a.h, a.alpha != nil
	}
	if w.maskCache == nil {
		w.maskCache = map[int]maskEntry{}
	}
	// A mask is every non-zero palette index of the sprite. DecodeSprite would
	// drop indices whose palette entry has alpha 0 — the water mask uses 1-4.
	rgba, err := w.renderer.G1.DecodeSpriteMaskOnly(spriteID)
	if err != nil {
		w.maskCache[spriteID] = maskEntry{}
		return nil, 0, 0, false
	}
	bw, bh := rgba.Bounds().Dx(), rgba.Bounds().Dy()
	a := make([]byte, bw*bh)
	for i := range a {
		a[i] = rgba.Pix[i*4+3]
	}
	w.maskCache[spriteID] = maskEntry{a, bw, bh}
	return a, bw, bh, true
}

type maskEntry struct {
	alpha []byte
	w, h  int
}
