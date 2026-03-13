package render

import (
	_ "embed"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed shaders/palette_tint.kage
var paletteTintKage []byte

// PaletteTintShader applies a destination-based colour tint using a blend mask,
// approximating OpenLoco's withBlend(ExtColour) draw mode.
//
// Usage:
//  1. Render all opaque world sprites to worldBuf.
//  2. For each blend sprite, call Draw() with the relevant worldBuf sub-region
//     as terrain and the mask returned by Renderer.GetSpriteMask() as mask.
//  3. Composite worldBuf and blendBuf onto the screen.
//
// OpenLoco reference: src/OpenLoco/src/Graphics/DrawSpriteHelper.hpp (blend pixel mode)
type PaletteTintShader struct {
	shader *ebiten.Shader
}

// NewPaletteTintShader compiles the Kage shader. Call once at startup.
// Returns nil error on success; a non-nil error means Kage compilation failed
// (the caller should fall back to the solid-diamond approximation).
func NewPaletteTintShader() (*PaletteTintShader, error) {
	s, err := ebiten.NewShader(paletteTintKage)
	if err != nil {
		return nil, err
	}
	return &PaletteTintShader{shader: s}, nil
}

// Draw tints a terrain region toward tint wherever mask is opaque, writing the
// result into dst at position (x, y).
//
//   dst      — destination image (e.g. blendBuf)
//   terrain  — worldBuf.SubImage(tileRect).(*ebiten.Image) for this tile
//   mask     — Renderer.GetSpriteMask() result (opaque white = apply tint)
//   tint     — target colour (e.g. water blue)
//   strength — blend factor 0.0–1.0 (0 = full terrain, 1 = full tint colour)
//   x, y     — pixel position in dst to place the output
func (s *PaletteTintShader) Draw(dst, terrain, mask *ebiten.Image,
	tint color.RGBA, strength float32, x, y int) {
	if s == nil || s.shader == nil || terrain == nil || mask == nil {
		return
	}
	w, h := mask.Bounds().Dx(), mask.Bounds().Dy()
	if w <= 0 || h <= 0 {
		return
	}
	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = terrain
	opts.Images[1] = mask
	opts.Uniforms = map[string]any{
		"TintR":        float32(tint.R) / 255,
		"TintG":        float32(tint.G) / 255,
		"TintB":        float32(tint.B) / 255,
		"TintStrength": strength,
	}
	opts.GeoM.Translate(float64(x), float64(y))
	dst.DrawRectShader(w, h, s.shader, opts)
}
