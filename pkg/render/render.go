package render

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/hajimehoshi/ebiten/v2"
)

var spriteErrorLogged = make(map[int]bool)
var objectSpriteWarned = make(map[string]bool)

// Renderer holds rendering state and sprite data
type Renderer struct {
	Screen *ebiten.Image
	Atlas  *Atlas
	G1     *assets.G1File
	ObjMgr *objects.ObjectManager

	// Cache for decoded G1 sprites (no palette remap)
	spriteCache map[int]*ebiten.Image

	// Cache for palette-remapped sprites, keyed by (spriteIdx<<8)|colourIdx
	colouredSpriteCache map[int]*ebiten.Image

	// Cache for decoded object sprites (keyed by "objectName:localIndex")
	objectSpriteCache map[string]*ebiten.Image
}

func NewRenderer() *Renderer {
	r := &Renderer{
		spriteCache:         make(map[int]*ebiten.Image),
		colouredSpriteCache: make(map[int]*ebiten.Image),
		objectSpriteCache:   make(map[string]*ebiten.Image),
	}
	// attempt to load atlas from default extracted assets directory; ignore errors
	if at, err := LoadAtlasFromDir("assets/extracted"); err == nil {
		r.Atlas = at
	}
	return r
}

func (r *Renderer) SetScreen(s *ebiten.Image) {
	r.Screen = s
}

// GetSprite returns an ebiten image for the given G1 sprite index
// Results are cached for performance
func (r *Renderer) GetSprite(index int) *ebiten.Image {
	if r.G1 == nil {
		return nil
	}

	// Check cache first
	if img, ok := r.spriteCache[index]; ok {
		return img
	}

	// Decode sprite
	rgba, err := r.G1.DecodeSprite(index)
	if err != nil {
		if !spriteErrorLogged[index] {
			log.Printf("[Render] Failed to decode sprite %d: %v", index, err)
			spriteErrorLogged[index] = true
		}
		return nil
	}

	// Convert to ebiten image and cache
	img := ebiten.NewImageFromImage(rgba)
	r.spriteCache[index] = img

	if len(r.spriteCache) <= 5 {
		log.Printf("[Render] Cached sprite %d: %dx%d", index, rgba.Bounds().Dx(), rgba.Bounds().Dy())
	}

	return img
}

// GetSpriteInfo returns the dimensions and offsets for a sprite
func (r *Renderer) GetSpriteInfo(index int) (width, height, xOff, yOff int16, ok bool) {
	if r.G1 == nil || index < 0 || index >= len(r.G1.Elements) {
		return 0, 0, 0, 0, false
	}
	elem := &r.G1.Elements[index]
	return elem.Width, elem.Height, elem.XOffset, elem.YOffset, true
}

// GetSpriteColoured returns a palette-remapped sprite for the given Colour index
// (0=black, 1=grey, 2=white, … matching OpenLoco's Colour enum).
// Results are cached per (spriteIndex, colourIndex) pair.
//
// OpenLoco reference: src/OpenLoco/src/Graphics/PaletteMap.cpp getForColour()
//   Colour N → G1[2170+N] is a 256-byte index remap table.
func (r *Renderer) GetSpriteColoured(spriteIndex, colourIndex int) *ebiten.Image {
	if r.G1 == nil {
		return nil
	}

	cacheKey := (spriteIndex << 8) | (colourIndex & 0xFF)
	if img, ok := r.colouredSpriteCache[cacheKey]; ok {
		return img
	}

	palMap, err := r.G1.GetPaletteMap(colourIndex)
	if err != nil {
		log.Printf("[Render] PaletteMap colour %d: %v", colourIndex, err)
		return r.GetSprite(spriteIndex) // fall back to unremapped
	}

	rgba, err := r.G1.DecodeSpriteMapped(spriteIndex, palMap)
	if err != nil {
		log.Printf("[Render] DecodeSpriteMapped sprite=%d colour=%d: %v", spriteIndex, colourIndex, err)
		return r.GetSprite(spriteIndex)
	}

	img := ebiten.NewImageFromImage(rgba)
	r.colouredSpriteCache[cacheKey] = img
	return img
}

// Clear fills the screen with transparent black (palette index 0).
//
// OpenLoco reference: src/OpenLoco/src/Graphics/SoftwareDrawingContext.cpp
//
//	SoftwareDrawingContext::clear(uint32_t fill)
func (r *Renderer) Clear() {
	if r.Screen != nil {
		r.Screen.Fill(color.RGBA{0, 0, 0, 0})
	}
}

// GetObjectSprite returns a decoded sprite from a land object
func (r *Renderer) GetObjectSprite(land *objects.LandObject, localIndex int) *ebiten.Image {
	if land == nil {
		return nil
	}

	// Check cache - use a proper string key
	cacheKey := fmt.Sprintf("%s:%d", land.Name, localIndex)
	if img, ok := r.objectSpriteCache[cacheKey]; ok {
		return img
	}

	// If ImageOffset is set, use G1 dynamic sprites
	if land.ImageOffset > 0 && r.G1 != nil {
		g1Index := int(land.ImageOffset) + localIndex
		rgba, err := r.G1.DecodeSprite(g1Index)
		if err != nil {
			if !objectSpriteWarned[cacheKey] {
				log.Printf("[Render] Failed to decode G1 sprite %d for %s:%d: %v", g1Index, land.Name, localIndex, err)
				objectSpriteWarned[cacheKey] = true
			}
			return nil
		}
		if rgba == nil {
			return nil
		}

		img := ebiten.NewImageFromImage(rgba)
		r.objectSpriteCache[cacheKey] = img
		return img
	}

	// Fallback: use embedded sprites
	if localIndex < 0 || localIndex >= len(land.Sprites) {
		return nil
	}

	sprite := land.Sprites[localIndex]
	if sprite == nil || len(sprite.Data) == 0 {
		if !objectSpriteWarned[cacheKey] {
			log.Printf("[Render] Sprite %s:%d has no data", land.Name, localIndex)
			objectSpriteWarned[cacheKey] = true
		}
		return nil
	}

	// Decode the sprite using the G1 palette
	rgba, err := r.decodeObjectSprite(sprite)
	if err != nil {
		if !objectSpriteWarned[cacheKey] {
			log.Printf("[Render] Failed to decode sprite %s:%d: %v", land.Name, localIndex, err)
			objectSpriteWarned[cacheKey] = true
		}
		return nil
	}
	if rgba == nil {
		if !objectSpriteWarned[cacheKey] {
			log.Printf("[Render] Decoded sprite %s:%d is nil", land.Name, localIndex)
			objectSpriteWarned[cacheKey] = true
		}
		return nil
	}

	img := ebiten.NewImageFromImage(rgba)
	r.objectSpriteCache[cacheKey] = img

	// Log first few cached sprites with pixel count
	if len(r.objectSpriteCache) <= 5 {
		// Count non-transparent pixels
		bounds := rgba.Bounds()
		nonTransparent := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, a := rgba.At(x, y).RGBA()
				if a > 0 {
					nonTransparent++
				}
			}
		}
		log.Printf("[Render] Cached object sprite %s:%d (%dx%d) - %d visible pixels", land.Name, localIndex, sprite.Width, sprite.Height, nonTransparent)
	}

	return img
}

// decodeObjectSprite decodes a sprite element using the G1 palette
func (r *Renderer) decodeObjectSprite(sprite *objects.SpriteElement) (*image.RGBA, error) {
	if r.G1 == nil {
		return nil, nil
	}

	width := int(sprite.Width)
	height := int(sprite.Height)
	if width <= 0 || height <= 0 {
		return nil, nil
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Get palette from G1
	palette := r.G1.GetPalette()
	if palette == nil {
		// Use a simple fallback palette
		palette = make([]color.RGBA, 256)
		for i := range palette {
			v := uint8(i)
			palette[i] = color.RGBA{v, v, v, 255}
		}
	}

	// Check sprite flags (from OpenLoco):
	// 0x01 = hasTransparency - has transparent sections
	// 0x04 = isRLECompressed - uses RLE encoding
	// 0x10 = hasZoomSprites - has zoom variants
	isRLE := sprite.Flags&0x04 != 0
	_ = sprite.Flags&0x01 != 0 // hasTransparency - used for drawing logic

	if isRLE {
		// RLE format: row offsets table followed by RLE data
		// First 2*height bytes are uint16 offsets to each row's data
		if len(sprite.Data) < height*2 {
			return nil, nil
		}

		for y := 0; y < height; y++ {
			rowOffset := int(sprite.Data[y*2]) | int(sprite.Data[y*2+1])<<8
			if rowOffset == 0 || rowOffset >= len(sprite.Data) {
				continue
			}

			x := 0
			pos := rowOffset
			for pos < len(sprite.Data) && x < width {
				if pos >= len(sprite.Data) {
					break
				}

				chunk := sprite.Data[pos]
				pos++

				if chunk == 0 {
					// End of row
					break
				}

				// High bit set = skip pixels (transparent)
				// Otherwise = draw pixels
				isTransparent := chunk&0x80 != 0
				count := int(chunk & 0x7F)

				if isTransparent {
					x += count
				} else {
					// Read 'count' palette indices
					for i := 0; i < count && x < width && pos < len(sprite.Data); i++ {
						palIdx := sprite.Data[pos]
						pos++
						if palIdx < uint8(len(palette)) {
							c := palette[palIdx]
							if palIdx != 0 { // Index 0 is typically transparent
								img.SetRGBA(x, y, c)
							}
						}
						x++
					}
				}
			}
		}
	} else {
		// Raw palette-indexed format
		pixelsSet := 0
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				idx := y*width + x
				if idx >= len(sprite.Data) {
					break
				}
				palIdx := sprite.Data[idx]
				if palIdx != 0 && int(palIdx) < len(palette) {
					c := palette[palIdx]
					// Only skip if the palette color itself is transparent
					if c.A > 0 {
						img.SetRGBA(x, y, c)
						pixelsSet++
					}
				}
			}
		}
		// Warn once if no pixels were set (indicates palette or data issue)
		if pixelsSet == 0 && len(sprite.Data) > 0 {
			cacheKey := fmt.Sprintf("rawdec:%d:%d", sprite.Width, sprite.Height)
			if !objectSpriteWarned[cacheKey] {
				log.Printf("[Render] Raw decode: 0 pixels set! Data len=%d, first 10 bytes: %v", len(sprite.Data), sprite.Data[:min(10, len(sprite.Data))])
				objectSpriteWarned[cacheKey] = true
			}
		}
	}

	return img, nil
}

// GetObjectSpriteInfo returns the dimensions and offsets for an object sprite
func (r *Renderer) GetObjectSpriteInfo(land *objects.LandObject, localIndex int) (width, height, xOff, yOff int16, ok bool) {
	if land == nil {
		return 0, 0, 0, 0, false
	}

	// If ImageOffset is set, get info from G1
	if land.ImageOffset > 0 && r.G1 != nil {
		g1Index := int(land.ImageOffset) + localIndex
		if g1Index < 0 || g1Index >= len(r.G1.Elements) {
			return 0, 0, 0, 0, false
		}
		elem := &r.G1.Elements[g1Index]
		return elem.Width, elem.Height, elem.XOffset, elem.YOffset, true
	}

	// Fallback: use embedded sprites
	if localIndex < 0 || localIndex >= len(land.Sprites) {
		return 0, 0, 0, 0, false
	}
	sprite := land.Sprites[localIndex]
	return sprite.Width, sprite.Height, sprite.XOffset, sprite.YOffset, true
}

// GetInterfaceSkinSprite returns a decoded sprite from the InterfaceSkin object.
//
// OpenLoco reference: src/OpenLoco/src/Objects/InterfaceSkinObject.h
//
//	Sprites are accessed via: interface->img + InterfaceSkin::ImageIds::toolbar_*
//
// The imageID is the InterfaceSkin::ImageIds constant (e.g., ISToolbarLoadsave).
func (r *Renderer) GetInterfaceSkinSprite(imageID uint32) *ebiten.Image {
	if r.ObjMgr == nil {
		log.Printf("[Render] GetInterfaceSkinSprite(%d): ObjMgr is nil", imageID)
		return nil
	}
	if r.ObjMgr.InterfaceSkin == nil {
		log.Printf("[Render] GetInterfaceSkinSprite(%d): InterfaceSkin is nil", imageID)
		return nil
	}

	skin := r.ObjMgr.InterfaceSkin

	if int(imageID) >= len(skin.Sprites) {
		log.Printf("[Render] GetInterfaceSkinSprite(%d): imageID out of range (max %d)", imageID, len(skin.Sprites)-1)
		return nil
	}

	// Check cache
	cacheKey := fmt.Sprintf("interface:%d", imageID)
	if img, ok := r.objectSpriteCache[cacheKey]; ok {
		return img
	}

	sprite := skin.Sprites[imageID]
	if sprite == nil {
		log.Printf("[Render] InterfaceSkin sprite %d is nil", imageID)
		return nil
	}
	if len(sprite.Data) == 0 {
		log.Printf("[Render] InterfaceSkin sprite %d has no data (width=%d, height=%d, flags=0x%04x)",
			imageID, sprite.Width, sprite.Height, sprite.Flags)
		return nil
	}

	// Decode the sprite
	rgba, err := r.decodeObjectSprite(sprite)
	if err != nil || rgba == nil {
		if !objectSpriteWarned[cacheKey] {
			log.Printf("[Render] Failed to decode InterfaceSkin sprite %d: %v", imageID, err)
			objectSpriteWarned[cacheKey] = true
		}
		return nil
	}

	img := ebiten.NewImageFromImage(rgba)
	r.objectSpriteCache[cacheKey] = img

	// Log first few cached sprites
	if len(r.objectSpriteCache) <= 10 {
		log.Printf("[Render] Cached InterfaceSkin sprite %d (%dx%d)", imageID, sprite.Width, sprite.Height)
	}

	return img
}
