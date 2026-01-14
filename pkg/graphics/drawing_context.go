package graphics

import (
	"image/color"

	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/gfx"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// DrawingContext wraps an Ebiten image and provides OpenLoco-style drawing operations
type DrawingContext struct {
	target     *ebiten.Image
	clipStack  []gfx.Rect
	saveStack  []drawingState
	lastError  error
	font       font.Face
}

type drawingState struct {
	clip gfx.Rect
}

// NewDrawingContext creates a new drawing context wrapping an Ebiten image
func NewDrawingContext(target *ebiten.Image) *DrawingContext {
	bounds := target.Bounds()
	return &DrawingContext{
		target: target,
		clipStack: []gfx.Rect{{
			X:      0,
			Y:      0,
			Width:  int16(bounds.Dx()),
			Height: int16(bounds.Dy()),
		}},
		font: basicfont.Face7x13,
	}
}

// FillRect fills a rectangle with a solid color
func (dc *DrawingContext) FillRect(x, y, width, height int16, col uint8) error {
	if dc.target == nil {
		return nil
	}

	// Create a temporary image for the rectangle
	rect := ebiten.NewImage(int(width), int(height))
	c := paletteColor(col)
	rect.Fill(c)

	// Draw it at the specified position
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x), float64(y))
	dc.target.DrawImage(rect, opts)

	return nil
}

// FillRectInset draws a 3D beveled rectangle
func (dc *DrawingContext) FillRectInset(x, y, width, height int16, col gfx.AdvancedColor, flags uint8) error {
	// For now, just draw a solid rectangle
	// TODO: Add proper inset/outset beveling
	return dc.FillRect(x, y, width, height, uint8(col.Color))
}

// DrawImage draws a sprite at the specified position
func (dc *DrawingContext) DrawImage(x, y int16, imageId uint32) error {
	// Try to get sprite from atlas
	sprite := render.GetGlobalSprite(imageId)
	if sprite == nil {
		// Fallback: draw placeholder rectangle
		return dc.FillRect(x, y, 16, 16, 5)
	}

	// Draw the sprite
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x), float64(y))
	dc.target.DrawImage(sprite, opts)
	return nil
}

// DrawString draws text at the specified position
func (dc *DrawingContext) DrawString(x, y int16, str string, col uint8) error {
	if dc.target == nil || dc.font == nil {
		return nil
	}

	c := paletteColor(col)
	text.Draw(dc.target, str, dc.font, int(x), int(y)+10, c) // +10 for baseline
	return nil
}

// DrawStringCentered draws text centered at the specified position
func (dc *DrawingContext) DrawStringCentered(x, y int16, str string, col uint8) error {
	// Simple center alignment - measure text and offset
	// TODO: Proper text measurement
	width := len(str) * 7 // Approximate character width
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

// paletteColor converts a palette index to an RGBA color
// TODO: Use actual OpenLoco palette
func paletteColor(index uint8) color.Color {
	// Simple grayscale palette for now
	if index == 0 {
		return color.Black
	}
	if index == 1 {
		return color.White
	}
	// Generate a color based on index
	r := uint8((int(index) * 7) % 256)
	g := uint8((int(index) * 13) % 256)
	b := uint8((int(index) * 23) % 256)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}
