package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"image"
	"image/color"
	"os"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

// Locomotion's own bitmap fonts, straight from g1.DAT.
//
// OpenLoco reference: src/OpenLoco/src/Graphics/Gfx.cpp
// initialiseCharacterWidths / getImageIdForCharacter:
//
//	glyph sprite = 1116 (characters_medium_normal_space) + font offset + chr-32
//	fonts: medium_normal +0, medium_bold +224, small +448, large +672
//	advance = glyph width - 1 (large: +1); chars 123-150 are unused (width 0)
const (
	g1FontBase = 1116

	G1FontMediumNormal = 0
	G1FontMediumBold   = 1
	G1FontSmall        = 2
	G1FontLarge        = 3
)

type g1Font struct {
	g1       *assets.G1File
	glyphs   map[int]*ebiten.Image // white glyph shapes, tinted at draw time
	outlines map[int]*ebiten.Image // the glyphs' textRemap1/2 pixels (outline style)
	widths   [4][224]int8
}

var globalG1Font *g1Font

// InitG1Font enables authentic Locomotion bitmap text using the g1 sprite
// pool. Once called, DrawText/DrawTextBold render glyph sprites instead of
// TTF fallback fonts.
func InitG1Font(g1 *assets.G1File) {
	if g1 == nil || os.Getenv("GOLOCO_FONT") == "ttf" {
		return
	}
	f := &g1Font{g1: g1, glyphs: make(map[int]*ebiten.Image), outlines: make(map[int]*ebiten.Image)}
	fudge := [4]int8{-1, -1, -1, 1}
	for fi := 0; fi < 4; fi++ {
		for i := 0; i < 224; i++ {
			idx := g1FontBase + fi*224 + i
			if idx >= len(g1.Elements) {
				continue
			}
			w := int8(g1.Elements[idx].Width) + fudge[fi]
			chr := i + 32
			// Characters 123-150 are unused in vanilla
			if chr >= 123 && chr <= 150 {
				w = 0
			}
			f.widths[fi][i] = w
		}
	}
	globalG1Font = f
}

// glyph returns the white-shaped glyph image for (font, chr), or nil.
// Glyph pixels are forced to white so ColorScale tints them to any colour.
func (f *g1Font) glyph(font, chr int) *ebiten.Image {
	if chr < 32 || chr > 255 {
		return nil
	}
	idx := g1FontBase + font*224 + (chr - 32)
	if img, ok := f.glyphs[idx]; ok {
		return img
	}
	// Glyph pixels are palette-REMAP indices (the game recolours them via the
	// text palette map), so a palette lookup would leave them transparent.
	// Treat any non-zero index as part of the glyph shape.
	e := &f.g1.Elements[idx]
	w, h := int(e.Width), int(e.Height)
	if w <= 0 || h <= 0 {
		f.glyphs[idx] = nil
		return nil
	}
	indices := e.Data
	if e.Flags&assets.G1FlagIsRLECompressed != 0 {
		decoded, err := assets.DecodeRLE(e.Data, w, h)
		if err != nil {
			f.glyphs[idx] = nil
			return nil
		}
		indices = decoded
	}
	if len(indices) < w*h {
		f.glyphs[idx] = nil
		return nil
	}
	// Glyphs use three remap indices (textRemap0/1/2 = 0x01/0x02/0x03). For
	// normal text only textRemap0 is coloured; 1 and 2 are transparent (they
	// exist for the inset/outline text styles).
	// OpenLoco reference: TextRenderer.cpp setTextColours.
	white := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if indices[y*w+x] == 0x01 {
				white.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
	img := ebiten.NewImageFromImage(white)
	f.glyphs[idx] = img
	return img
}

// supports reports whether every rune in str has a vanilla bitmap glyph.
// Chars 123-150 are unused in the game's fonts (includes {|}~), and anything
// outside Latin-1 needs the TTF fallback.
func (f *g1Font) supports(str string) bool {
	for _, r := range str {
		if r < 32 || r > 255 || (r >= 123 && r <= 150) {
			return false
		}
	}
	return true
}

// drawString renders str at (x, y top-left) in the given font and colour and
// returns the advance in pixels. dst may be nil to only measure.
func (f *g1Font) drawString(dst *ebiten.Image, str string, font, x, y int, scale float64, clr color.Color) int {
	penX := 0
	for _, r := range str {
		chr := int(r)
		if chr < 32 || chr > 255 {
			chr = '?'
		}
		w := int(f.widths[font][chr-32])
		if dst != nil {
			if g := f.glyph(font, chr); g != nil {
				idx := g1FontBase + font*224 + (chr - 32)
				e := &f.g1.Elements[idx]
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(
					float64(x)+(float64(penX)+float64(e.XOffset))*scale,
					float64(y)+float64(e.YOffset)*scale)
				op.ColorScale.ScaleWithColor(clr)
				dst.DrawImage(g, op)
			}
		}
		penX += w
	}
	return int(float64(penX) * scale)
}

// g1FontHeight returns the nominal line height for a font index.
func g1FontHeight(font int) int {
	switch font {
	case G1FontSmall:
		return 6
	case G1FontLarge:
		return 18
	default:
		return 10
	}
}

// outline returns the glyph's textRemap1/2 pixels (0x02, 0x03) as a white
// shape: TextRenderer's outline style colours them (setTextColours pal2/pal3).
func (f *g1Font) outline(font, chr int) *ebiten.Image {
	if chr < 32 || chr > 255 {
		return nil
	}
	idx := g1FontBase + font*224 + (chr - 32)
	if img, ok := f.outlines[idx]; ok {
		return img
	}
	e := &f.g1.Elements[idx]
	w, h := int(e.Width), int(e.Height)
	indices := e.Data
	if e.Flags&assets.G1FlagIsRLECompressed != 0 {
		decoded, err := assets.DecodeRLE(e.Data, w, h)
		if err != nil {
			f.outlines[idx] = nil
			return nil
		}
		indices = decoded
	}
	if w <= 0 || h <= 0 || len(indices) < w*h {
		f.outlines[idx] = nil
		return nil
	}
	m := image.NewRGBA(image.Rect(0, 0, w, h))
	any := false
	for i, p := range indices[:w*h] {
		if p == 0x02 || p == 0x03 {
			m.SetRGBA(i%w, i/w, color.RGBA{255, 255, 255, 255})
			any = true
		}
	}
	var img *ebiten.Image
	if any {
		img = ebiten.NewImageFromImage(m)
	}
	f.outlines[idx] = img
	return img
}

// DrawTextOutlined draws str in medium bold with the outline text style
// (AdvancedColour::outline): the glyphs' outline pixels in outlineClr (upstream
// black0), then the body in clr. Returns the width.
func DrawTextOutlined(screen *ebiten.Image, str string, x, y int, clr, outlineClr color.Color) int {
	f := globalG1Font
	if f == nil {
		return 0
	}
	penX := 0
	for _, r := range str {
		chr := int(r)
		if chr < 32 || chr > 255 {
			chr = '?'
		}
		idx := g1FontBase + G1FontMediumBold*224 + (chr - 32)
		e := &f.g1.Elements[idx]
		for _, plane := range [2]struct {
			img *ebiten.Image
			c   color.Color
		}{{f.outline(G1FontMediumBold, chr), outlineClr}, {f.glyph(G1FontMediumBold, chr), clr}} {
			if plane.img == nil {
				continue
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x+penX+int(e.XOffset)), float64(y+int(e.YOffset)))
			op.ColorScale.ScaleWithColor(plane.c)
			screen.DrawImage(plane.img, op)
		}
		penX += int(f.widths[G1FontMediumBold][chr-32])
	}
	return penX
}

// TextColours returns the body and outline colours of the outline text style
// for a colour: getShade(colour, 9) and PaletteIndex::black0.
func TextColours(colour int) (body, outline color.Color) {
	if globalG1Font == nil {
		return color.White, color.Black
	}
	pal := globalG1Font.g1.GetPalette()
	m, err := globalG1Font.g1.GetPaletteMap(colour)
	if err != nil {
		return color.White, pal[paletteBlack0]
	}
	return pal[gfx8.Shade(m, 9)], pal[paletteBlack0]
}
