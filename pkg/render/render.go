package render

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"runtime/debug"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/hajimehoshi/ebiten/v2"
)

var spriteErrorLogged = make(map[int]bool)
var objectSpriteWarned = make(map[string]bool)

// ── Sprite atlas ──────────────────────────────────────────────────────────────
//
// SpriteAtlas packs decoded G1 sprites into a single large *ebiten.Image.
// Sub-images from the same parent share one GPU texture, so Ebiten can batch
// all DrawImage calls in a single GPU draw call instead of one per sprite.
//
// Packing is a simple left-to-right, top-to-bottom row strip.  Sprites are
// 1-pixel-separated to avoid bilinear bleed (not that Ebiten uses it, but it
// is good hygiene).
//
// The atlas is lazily initialised on the first GetSprite call so that
// ebiten.NewImage is always called from within the game loop.
const (
	spriteAtlasW = 4096
	spriteAtlasH = 4096
	spriteAtlasP = 1 // 1-px gap between packed sprites
)

type SpriteAtlas struct {
	backing      *ebiten.Image
	regions      map[int]image.Rectangle
	packX, packY int
	rowH         int
	full         bool
}

func newSpriteAtlas() *SpriteAtlas {
	return &SpriteAtlas{
		backing: ebiten.NewImage(spriteAtlasW, spriteAtlasH),
		regions: make(map[int]image.Rectangle),
	}
}

// pack copies img into the atlas and returns the sub-image.
// Returns nil when the atlas is full; callers must fall back to the raw image.
func (sa *SpriteAtlas) pack(key int, img *ebiten.Image) *ebiten.Image {
	if sa.full {
		return nil
	}
	iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
	if iw <= 0 || ih <= 0 {
		return nil
	}
	if sa.packX+iw > spriteAtlasW {
		sa.packX = 0
		sa.packY += sa.rowH + spriteAtlasP
		sa.rowH = 0
	}
	if sa.packY+ih > spriteAtlasH {
		sa.full = true
		log.Printf("[Atlas] Sprite atlas full after %d sprites", len(sa.regions))
		return nil
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(sa.packX), float64(sa.packY))
	sa.backing.DrawImage(img, op)
	rect := image.Rect(sa.packX, sa.packY, sa.packX+iw, sa.packY+ih)
	sa.regions[key] = rect
	if ih > sa.rowH {
		sa.rowH = ih
	}
	sa.packX += iw + spriteAtlasP
	return sa.backing.SubImage(rect).(*ebiten.Image)
}

// get returns the sub-image for key, or nil if not yet packed.
func (sa *SpriteAtlas) get(key int) *ebiten.Image {
	rect, ok := sa.regions[key]
	if !ok {
		return nil
	}
	return sa.backing.SubImage(rect).(*ebiten.Image)
}

// ── Renderer ──────────────────────────────────────────────────────────────────

// Renderer holds rendering state and sprite data
type Renderer struct {
	Screen *ebiten.Image
	Atlas  *Atlas
	G1     *assets.G1File
	ObjMgr *objects.ObjectManager

	// mainAtlas packs all raw G1 sprites (and palette-remapped variants) into
	// one GPU texture so Ebiten can batch consecutive DrawImage calls.
	// Lazily created on first GetSprite call (ebiten.NewImage requires game loop).
	mainAtlas *SpriteAtlas

	// Cache for decoded G1 sprites (no palette remap) — fallback when atlas full
	spriteCache map[int]*ebiten.Image

	// Cache for palette-remapped sprites, keyed by (spriteIdx<<8)|colourIdx
	colouredSpriteCache map[int]*ebiten.Image

	// Cache for two-colour remapped sprites, keyed by (idx<<16)|(c1<<8)|c2.
	coloured2SpriteCache map[int]*ebiten.Image

	// Cache for raw-palette-map sprites (water blend etc.), keyed by spriteIdx.
	// Separate from colouredSpriteCache: its raw spriteIndex keys would collide
	// with that map's shifted keys.
	paletteMapSpriteCache map[int]*ebiten.Image

	// Cache for decoded object sprites (keyed by "objectName:localIndex")
	objectSpriteCache map[string]*ebiten.Image

	// Cache for blend-mask sprites (palette index 0 → opaque white, others → transparent)
	maskSpriteCache map[int]*ebiten.Image

	// Cache for cliff-edge masked sprites keyed by spriteID*100000+maskID
	maskedSpriteCache map[int]*ebiten.Image
}

func NewRenderer() *Renderer {
	r := &Renderer{
		spriteCache:           make(map[int]*ebiten.Image),
		colouredSpriteCache:   make(map[int]*ebiten.Image),
		coloured2SpriteCache:  make(map[int]*ebiten.Image),
		paletteMapSpriteCache: make(map[int]*ebiten.Image),
		objectSpriteCache:     make(map[string]*ebiten.Image),
		maskSpriteCache:       make(map[int]*ebiten.Image),
		maskedSpriteCache:     make(map[int]*ebiten.Image),
	}
	// attempt to load atlas from default extracted assets directory; ignore errors
	if at, err := LoadAtlasFromDir("assets/extracted"); err == nil {
		r.Atlas = at
	}
	return r
}

// ensureAtlas lazily creates the main sprite atlas on first use.
// Must be called from within the Ebiten game loop (Draw/Update).
func (r *Renderer) ensureAtlas() *SpriteAtlas {
	if r.mainAtlas == nil {
		r.mainAtlas = newSpriteAtlas()
	}
	return r.mainAtlas
}

func (r *Renderer) SetScreen(s *ebiten.Image) {
	r.Screen = s
}

// GetSprite returns an ebiten image for the given G1 sprite index.
// The first call for each index decodes the sprite and packs it into the main
// sprite atlas so all subsequent DrawImage calls share one GPU texture and are
// batched by Ebiten's renderer.
func (r *Renderer) GetSprite(index int) *ebiten.Image {
	if r.G1 == nil {
		return nil
	}

	// Check atlas first (fast path — sub-image lookup only)
	atlas := r.ensureAtlas()
	if img := atlas.get(index); img != nil {
		return img
	}

	// Fallback cache (used when atlas is full)
	if img, ok := r.spriteCache[index]; ok {
		return img
	}

	// Decode sprite
	rgba, err := r.G1.DecodeSprite(index)
	if err != nil {
		if !spriteErrorLogged[index] {
			log.Printf("[Render] Failed to decode sprite %d: %v", index, err)
			spriteErrorLogged[index] = true
			if os.Getenv("GOLOCO_DEBUG_SPRITES") == "1" {
				log.Printf("[Render] caller stack:\n%s", debug.Stack())
			}
		}
		return nil
	}

	tmp := ebiten.NewImageFromImage(rgba)

	// Pack into atlas; if atlas is full fall back to individual sprite cache
	if sub := atlas.pack(index, tmp); sub != nil {
		return sub
	}
	r.spriteCache[index] = tmp
	return tmp
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
//
//	Colour N → G1[2170+N] is a 256-byte index remap table.
func (r *Renderer) GetSpriteColoured(spriteIndex, colourIndex int) *ebiten.Image {
	if r.G1 == nil {
		return nil
	}
	if spriteIndex < 0 || spriteIndex >= len(r.G1.Elements) {
		return nil
	}

	cacheKey := (spriteIndex << 8) | (colourIndex & 0xFF)
	if img, ok := r.colouredSpriteCache[cacheKey]; ok {
		return img // nil entries cached here prevent repeated failed decodes
	}

	palMap, err := r.G1.GetPaletteMap(colourIndex)
	if err != nil {
		log.Printf("[Render] PaletteMap colour %d: %v", colourIndex, err)
		return r.GetSprite(spriteIndex) // fall back to unremapped
	}

	rgba, err := r.G1.DecodeSpriteMapped(spriteIndex, palMap)
	if err != nil {
		if !spriteErrorLogged[spriteIndex] {
			log.Printf("[Render] DecodeSpriteMapped sprite=%d colour=%d: %v", spriteIndex, colourIndex, err)
			spriteErrorLogged[spriteIndex] = true
		}
		r.colouredSpriteCache[cacheKey] = nil // cache nil to stop repeated attempts
		return nil
	}

	img := ebiten.NewImageFromImage(rgba)
	r.colouredSpriteCache[cacheKey] = img
	return img
}

// GetSpriteColoured2 returns a sprite remapped with two Colour indices
// (primary + secondary), like OpenLoco's Gfx::recolour2. Used for e.g. the
// loading-screen train. Results are cached per (sprite, c1, c2).
func (r *Renderer) GetSpriteColoured2(spriteIndex, primary, secondary int) *ebiten.Image {
	if r.G1 == nil || spriteIndex < 0 || spriteIndex >= len(r.G1.Elements) {
		return nil
	}
	cacheKey := (spriteIndex << 16) | ((primary & 0xFF) << 8) | (secondary & 0xFF)
	if img, ok := r.coloured2SpriteCache[cacheKey]; ok {
		return img
	}
	palMap, err := r.G1.GetPaletteMap2(primary, secondary)
	if err != nil {
		return r.GetSprite(spriteIndex)
	}
	rgba, err := r.G1.DecodeSpriteMapped(spriteIndex, palMap)
	if err != nil {
		return r.GetSprite(spriteIndex)
	}
	img := ebiten.NewImageFromImage(rgba)
	r.coloured2SpriteCache[cacheKey] = img
	return img
}

// GetSpriteWithPaletteMap decodes a sprite applying a raw 256-byte palette remap table.
// Used for water blend sprites whose palette indices 1-4 map to water-blue shades
// via the WaterObject's remap table (G1[waterImageOffset+41]).
// Results are cached by spriteIndex (caller must pass the same palMap per spriteIndex).
func (r *Renderer) GetSpriteWithPaletteMap(spriteIndex int, palMap []byte) *ebiten.Image {
	if r.G1 == nil {
		return nil
	}
	if img, ok := r.paletteMapSpriteCache[spriteIndex]; ok {
		return img
	}
	rgba, err := r.G1.DecodeSpriteMapped(spriteIndex, palMap)
	if err != nil {
		return r.GetSprite(spriteIndex)
	}
	img := ebiten.NewImageFromImage(rgba)
	r.paletteMapSpriteCache[spriteIndex] = img
	return img
}

// GetSpriteMask returns a blend-mask image for the given G1 sprite index.
// Pixels with palette index 0 (the blend region in withBlend sprites) become
// opaque white; all other pixels are transparent.
// Used for water +35, glass, shadow, and any other withBlend() sprite.
//
// OpenLoco reference: src/OpenLoco/src/Graphics/DrawSpriteHelper.hpp (blend pixel mode)
func (r *Renderer) GetSpriteMask(index int) *ebiten.Image {
	if r.G1 == nil {
		return nil
	}
	if img, ok := r.maskSpriteCache[index]; ok {
		return img
	}
	rgba, err := r.G1.DecodeSpriteMaskOnly(index)
	if err != nil {
		return nil
	}
	img := ebiten.NewImageFromImage(rgba)
	r.maskSpriteCache[index] = img
	return img
}

// GetMaskedSprite returns a cliff-edge sprite clipped by maskSpriteID using
// DestinationIn compositing.  The result is cached so allocation only happens
// once per (spriteID, maskID) pair.
//
// OpenLoco reference: PaintSurface.cpp paintEdgeSection() — hasMaskedImage /
//
//	maskedImageId drawn with DrawSpriteHelper blend-mask logic.
func (r *Renderer) GetMaskedSprite(spriteID, maskID int) *ebiten.Image {
	key := spriteID*100000 + maskID
	if img, ok := r.maskedSpriteCache[key]; ok {
		return img
	}
	sprite := r.GetSprite(spriteID)
	if sprite == nil {
		return nil
	}
	mask := r.GetSprite(maskID)
	if mask == nil {
		// No mask available — return unmasked sprite so cliff is still visible.
		return sprite
	}

	// Create a temp image large enough for both the cliff and the mask.
	// Both sprites share the same in-world anchor, so they align at (0,0).
	sw, sh := sprite.Bounds().Dx(), sprite.Bounds().Dy()
	mw, mh := mask.Bounds().Dx(), mask.Bounds().Dy()
	w, h := sw, sh
	if mw > w {
		w = mw
	}
	if mh > h {
		h = mh
	}
	tmp := ebiten.NewImage(w, h)
	tmp.DrawImage(sprite, nil)

	mop := &ebiten.DrawImageOptions{}
	mop.Blend = ebiten.BlendDestinationIn
	tmp.DrawImage(mask, mop)

	r.maskedSpriteCache[key] = tmp
	return tmp
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

	// If ImageOffset is set, route through GetSprite so the sprite goes into the
	// main atlas and all land/object draws share the same GPU texture.
	if land.ImageOffset > 0 && r.G1 != nil {
		g1Index := int(land.ImageOffset) + localIndex
		img := r.GetSprite(g1Index)
		if img != nil {
			r.objectSpriteCache[cacheKey] = img // secondary cache so string-key lookup still works
		}
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
		// RLE format: per-row uint16 LE offset table, then chunks of
		// [size (bit7 = last chunk)][firstX][size pixel bytes] — decoded by
		// the same logic as g1.dat sprites.
		indices, err := assets.DecodeRLE(sprite.Data, width, height)
		if err != nil {
			return nil, nil
		}
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				palIdx := indices[y*width+x]
				if palIdx != 0 && int(palIdx) < len(palette) {
					img.SetRGBA(x, y, palette[palIdx])
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

// PrewarmSpriteRange decodes and packs G1 sprites [start, end) into the main
// atlas.  Call this during the loading phase so the render loop pays no decode
// cost.  Returns the number of sprites successfully packed.
func (r *Renderer) PrewarmSpriteRange(start, end int) int {
	if r.G1 == nil {
		return 0
	}
	n := 0
	for i := start; i < end; i++ {
		if r.GetSprite(i) != nil {
			n++
		}
	}
	return n
}

// PrewarmSpriteList decodes and packs the given G1 sprite IDs into the atlas.
func (r *Renderer) PrewarmSpriteList(ids []int) int {
	n := 0
	for _, id := range ids {
		if r.GetSprite(id) != nil {
			n++
		}
	}
	return n
}

// AtlasPackedCount returns the number of sprites currently in the main atlas.
func (r *Renderer) AtlasPackedCount() int {
	if r.mainAtlas == nil {
		return 0
	}
	return len(r.mainAtlas.regions)
}
