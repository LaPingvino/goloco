package graphics

import (
	"image/color"
	"log"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/gfx"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// DrawingContext wraps an Ebiten image and provides OpenLoco-style drawing operations
type DrawingContext struct {
	target        *ebiten.Image
	clipStack     []gfx.Rect
	saveStack     []drawingState
	lastError     error
	font          font.Face
	palette       [256]color.RGBA // active palette (copied from G1 when available)
	paletteLoaded bool            // whether palette field is initialized
}

// Track which sprites have been logged (log each sprite ID once)
var spriteLoggedForId = make(map[uint32]bool)
var drawStringCenteredLogged bool

type drawingState struct {
	clip gfx.Rect
}

// RectInsetFlags for FillRectInset operations
const (
	RectInsetNone        = 0
	RectInsetBorder      = 1 << 0
	RectInsetSecondary   = 1 << 1
	RectInsetTransparent = 1 << 2
)

// package-level global palette variables copied from G1 when available.
// Code elsewhere (e.g. after parsing G1) should call SetGlobalPalette(g1).
var globalPalette [256]color.RGBA
var globalPaletteLoaded bool

// SetGlobalPalette copies the palette from the provided G1 file into the
// package-level global palette. Returns true if a usable palette was found.
// Additionally enforce convention that palette index 0 is transparent.
func SetGlobalPalette(g1 *assets.G1File) bool {
	if g1 == nil {
		return false
	}
	globalPalette = g1.Palette

	// Ensure palette index 0 is transparent per convention.
	globalPalette[0] = color.RGBA{0, 0, 0, 0}

	// Verify palette is non-empty (at least one non-zero entry).
	ok := false
	for i := 0; i < 256; i++ {
		p := globalPalette[i]
		if p.A != 0 || p.R != 0 || p.G != 0 || p.B != 0 {
			ok = true
			break
		}
	}
	globalPaletteLoaded = ok
	return ok
}

// NewDrawingContext creates a new drawing context wrapping an Ebiten image
// If a global palette (from G1) has been set via SetGlobalPalette, the new
// context will copy it and use it for UI drawing.
func NewDrawingContext(target *ebiten.Image) *DrawingContext {
	bounds := target.Bounds()
	dc := &DrawingContext{
		target: target,
		clipStack: []gfx.Rect{{
			X:      0,
			Y:      0,
			Width:  int16(bounds.Dx()),
			Height: int16(bounds.Dy()),
		}},
		font: basicfont.Face7x13,
	}
	// If a global palette was loaded earlier, copy it into this context.
	if globalPaletteLoaded {
		dc.palette = globalPalette
		dc.paletteLoaded = true
	}
	return dc
}

// UseRendererPalette copies the active palette from a parsed G1 file (if available)
// into the drawing context so UI rendering uses the game palette. It looks for a
// palette copied by the G1 parsing code or, as a fallback, searches G1 elements
// for an element flagged as an R8G8B8 palette. Returns true if a palette was
// successfully loaded.
func (dc *DrawingContext) UseRendererPalette(g1 *assets.G1File) bool {
	if dc == nil || g1 == nil {
		return false
	}

	// If the G1File already has a palette populated, copy it.
	// This is the preferred path: ParseG1 / LoadG1 should fill G1File.Palette.
	dc.palette = g1.Palette

	// Check whether the copied palette looks valid (any non-zero RGB/A).
	loaded := false
	for i := 0; i < 256; i++ {
		p := dc.palette[i]
		if p.A != 0 || p.R != 0 || p.G != 0 || p.B != 0 {
			loaded = true
			break
		}
	}
	if loaded {
		dc.paletteLoaded = true
		return true
	}

	// Fallback: try to find a palette element within the G1 elements (flagged as R8G8B8).
	for i := range g1.Elements {
		e := &g1.Elements[i]
		if e == nil {
			continue
		}
		// Check for the R8G8B8 palette flag on the element.
		if (e.Flags&assets.G1FlagIsR8G8B8Palette) != 0 && len(e.Data) >= 3 {
			startIdx := int(e.XOffset)
			count := int(e.Width)
			if count <= 0 || startIdx < 0 || startIdx >= 256 {
				continue
			}
			// Ensure we don't read past element data bounds.
			maxCount := count
			if len(e.Data) < count*3 {
				maxCount = len(e.Data) / 3
			}
			for j := 0; j < maxCount && (startIdx+j) < 256; j++ {
				// Data is B, G, R order (Windows BITMAPINFO style), same as
				// assets.(*G1File).loadPalette.
				dc.palette[startIdx+j] = color.RGBA{
					R: e.Data[j*3+2],
					G: e.Data[j*3+1],
					B: e.Data[j*3+0],
					A: 255,
				}
			}
			// Make palette index 0 transparent per convention.
			dc.palette[0] = color.RGBA{0, 0, 0, 0}
			dc.paletteLoaded = true
			return true
		}
	}

	// No usable palette found.
	return false
}

// fillPixel is a shared 1x1 white image scaled and tinted by FillRect, so a
// fill does not allocate a new GPU image per call.
var fillPixel = func() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)
	return img
}()

// FillRect fills a rectangle with a solid color
func (dc *DrawingContext) FillRect(x, y, width, height int16, col uint8) error {
	if dc.target == nil || width <= 0 || height <= 0 {
		return nil
	}

	c := dc.paletteColor(col)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(float64(width), float64(height))
	opts.GeoM.Translate(float64(x), float64(y))
	opts.ColorScale.ScaleWithColor(c)
	dc.target.DrawImage(fillPixel, opts)

	return nil
}

// FillRectInset draws a 3D beveled rectangle
func (dc *DrawingContext) FillRectInset(x, y, width, height int16, col gfx.AdvancedColor, flags uint8) error {
	// Draw base/background
	bgIdx := uint8(col.Color)
	_ = dc.FillRect(x, y, width, height, bgIdx)

	// If no border flag requested, return early
	if flags&RectInsetBorder == 0 {
		return nil
	}

	// Compute light/dark variants for bevel using simple index offsets.
	// This is a heuristic: real OpenLoco uses palette maps; here we choose
	// nearby indices to simulate highlights and shadows.
	var lightIdx uint8
	var darkIdx uint8
	if bgIdx <= 252 {
		lightIdx = bgIdx + 3
	} else {
		lightIdx = bgIdx
	}
	if bgIdx >= 3 {
		darkIdx = bgIdx - 3
	} else {
		darkIdx = bgIdx
	}

	// Top border (light)
	_ = dc.FillRect(x, y, width, 1, lightIdx)
	// Left border (light)
	_ = dc.FillRect(x, y, 1, height, lightIdx)
	// Bottom border (dark)
	if height > 0 {
		_ = dc.FillRect(x, y+height-1, width, 1, darkIdx)
	}
	// Right border (dark)
	if width > 0 {
		_ = dc.FillRect(x+width-1, y, 1, height, darkIdx)
	}

	// If a secondary flag is provided, draw a second inner border to enhance 3D look
	if flags&RectInsetSecondary != 0 {
		innerLight := lightIdx
		innerDark := darkIdx
		if width > 2 && height > 2 {
			_ = dc.FillRect(x+1, y+1, width-2, 1, innerLight)
			_ = dc.FillRect(x+1, y+1, 1, height-2, innerLight)
			_ = dc.FillRect(x+1, y+height-2, width-2, 1, innerDark)
			_ = dc.FillRect(x+width-2, y+1, 1, height-2, innerDark)
		}
	}

	return nil
}

// DrawImage draws a sprite at the specified position
func (dc *DrawingContext) DrawImage(x, y int16, imageId uint32) error {
	// Try to get sprite from atlas
	sprite := render.GetGlobalSprite(imageId)
	if sprite == nil {
		if !spriteLoggedForId[imageId] {
			log.Printf("[Graphics] DrawImage: id=%d spriteFound=false x=%d y=%d", imageId, x, y)
			spriteLoggedForId[imageId] = true
		}
		// Fallback: draw placeholder rectangle
		return dc.FillRect(x, y, 16, 16, 5)
	}

	// Log found sprite (size helps debugging scaling) - log once per sprite ID
	spriteLogged := spriteLoggedForId[imageId]
	if !spriteLogged {
		log.Printf("[Graphics] DrawImage: id=%d spriteFound=true x=%d y=%d size=%dx%d", imageId, x, y, sprite.Bounds().Dx(), sprite.Bounds().Dy())
		spriteLoggedForId[imageId] = true
	}

	// Draw the sprite
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(2, 2) // Scale up to 32x32 for tile size
	opts.GeoM.Translate(float64(x), float64(y))
	dc.target.DrawImage(sprite, opts)
	return nil
}

// DrawString draws text at the specified position
func (dc *DrawingContext) DrawString(x, y int16, str string, col uint8) error {
	if dc.target == nil || dc.font == nil {
		return nil
	}

	c := dc.paletteColor(col)
	text.Draw(dc.target, str, dc.font, int(x), int(y)+10, c) // +10 for baseline
	return nil
}

// DrawStringCentered draws text centered at the specified position.
// Stub: the width calculation uses a fixed per-character approximation
// instead of querying the actual font metrics.
//
// OpenLoco reference: src/OpenLoco/src/Graphics/DrawingContext.cpp
//
//	DrawingContext::drawStringCentred(...)
//	DrawingContext::measureString(...)
//
// Replace the len*7 heuristic with a proper call to font.Metrics or
// golang.org/x/image/font.MeasureString once font sizing is wired up.
func (dc *DrawingContext) DrawStringCentered(x, y int16, str string, col uint8) error {
	if !drawStringCenteredLogged {
		log.Println("[Graphics] DrawStringCentered: stub — using approximate character width (7px per char)")
		drawStringCenteredLogged = true
	}
	width := len(str) * 7 // approximate; needs real font measurement
	return dc.DrawString(x-int16(width/2), y, str, col)
}

// DrawStringCenteredClipped draws centered text with clipping
func (dc *DrawingContext) DrawStringCenteredClipped(x, y int16, maxWidth int16, col gfx.AdvancedColor, str string) error {
	return dc.DrawStringCentered(x, y, str, uint8(col.Color))
}

// DrawStringCentred is a British spelling alias
func (dc *DrawingContext) DrawStringCentredClipped(x, y int16, maxWidth int16, col gfx.AdvancedColor, str string) error {
	return dc.DrawStringCenteredClipped(x, y, maxWidth, col, str)
}

// SetClip sets the clipping rectangle
func (dc *DrawingContext) SetClip(x, y, width, height int16) {
	if len(dc.clipStack) > 0 {
		dc.clipStack[len(dc.clipStack)-1] = gfx.Rect{X: x, Y: y, Width: width, Height: height}
	}
}

// Save pushes the current drawing state
func (dc *DrawingContext) Save() {
	var clip gfx.Rect
	if len(dc.clipStack) > 0 {
		clip = dc.clipStack[len(dc.clipStack)-1]
	}
	dc.saveStack = append(dc.saveStack, drawingState{clip: clip})
}

// Restore pops the drawing state
func (dc *DrawingContext) Restore() {
	if len(dc.saveStack) > 0 {
		state := dc.saveStack[len(dc.saveStack)-1]
		dc.saveStack = dc.saveStack[:len(dc.saveStack)-1]
		if len(dc.clipStack) > 0 {
			dc.clipStack[len(dc.clipStack)-1] = state.clip
		}
	}
}

// SubContext creates a sub-context with relative coordinates
func (dc *DrawingContext) SubContext(x, y, width, height int16) *DrawingContext {
	// Create a sub-region (for now, just return self with adjusted clip)
	sub := &DrawingContext{
		target:    dc.target,
		clipStack: make([]gfx.Rect, len(dc.clipStack)),
		font:      dc.font,
	}
	copy(sub.clipStack, dc.clipStack)
	sub.SetClip(x, y, width, height)
	return sub
}

// SetError sets the last error
func (dc *DrawingContext) SetError(err error) {
	dc.lastError = err
}

// GetError returns the last error
func (dc *DrawingContext) GetError() error {
	return dc.lastError
}

// paletteColor converts a palette index to an RGBA color using the active
// palette on the drawing context when available. If no palette is loaded,
// fall back to a synthetic mapping for visibility during development.
//
// When a package-level global palette is present (copied from a real G1),
// the function will use that palette for consistent UI coloring even when
// the context-specific palette isn't initialized.
func (dc *DrawingContext) paletteColor(index uint8) color.Color {
	// If the drawing context has a loaded palette, use it.
	if dc != nil && dc.paletteLoaded {
		return dc.palette[index]
	}

	// If a package-level global palette is available, use that.
	if globalPaletteLoaded {
		return globalPalette[index]
	}

	// Fallback behavior:
	// Treat 0 as transparent black (alpha 0) which matches many sprite systems.
	if index == 0 {
		return color.RGBA{0, 0, 0, 0}
	}
	// Index 1 default to white for visibility
	if index == 1 {
		return color.RGBA{255, 255, 255, 255}
	}

	// Synthetic mapping for development / when no real palette is available.
	r := uint8((int(index) * 7) % 256)
	g := uint8((int(index) * 13) % 256)
	b := uint8((int(index) * 23) % 256)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// MatchPaletteIndex finds the closest palette index in the global palette to the
// provided color using simple squared Euclidean distance on RGB. If no global
// palette exists, it returns 1 (a visible default).
func MatchPaletteIndex(c color.RGBA) uint8 {
	if !globalPaletteLoaded {
		return 1
	}
	bestIdx := uint8(1)
	bestDist := int(^uint(0) >> 1) // large number
	tr := int(c.R)
	tg := int(c.G)
	tb := int(c.B)
	// Start at 1: index 0 is forced transparent, so matching it would make
	// the caller draw nothing (dark colours would otherwise snap to it).
	for i := 1; i < 256; i++ {
		p := globalPalette[i]
		dr := tr - int(p.R)
		dg := tg - int(p.G)
		db := tb - int(p.B)
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			bestDist = dist
			bestIdx = uint8(i)
		}
	}
	return bestIdx
}

// GetGlobalColor returns the RGBA color for the given palette index from the
// package-level global palette. If the global palette is not loaded, it
// returns a sensible fallback color mapping used elsewhere in this package.
func GetGlobalColor(index uint8) color.RGBA {
	if globalPaletteLoaded {
		return globalPalette[index]
	}
	// Fallback behavior matches paletteColor's fallback so widgets remain visible.
	if index == 0 {
		return color.RGBA{0, 0, 0, 0}
	}
	if index == 1 {
		return color.RGBA{255, 255, 255, 255}
	}
	r := uint8((int(index) * 7) % 256)
	g := uint8((int(index) * 13) % 256)
	b := uint8((int(index) * 23) % 256)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// IsGlobalPaletteLoaded reports whether a valid global palette has been set.
func IsGlobalPaletteLoaded() bool {
	return globalPaletteLoaded
}
