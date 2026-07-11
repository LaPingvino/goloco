package assets

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// G1 file format constants
const (
	G1ExpectedCountDisc  = 0x101A // and GOG (upstream Gfx.h G1ExpectedCount::kDisc)
	G1ExpectedCountSteam = 0x0F38
	DefaultPaletteIndex  = 304
)

// G1ElementFlags for sprite properties
type G1ElementFlags uint16

const (
	G1FlagNone            G1ElementFlags = 0
	G1FlagHasTransparency G1ElementFlags = 1 << 0
	G1FlagUnk1            G1ElementFlags = 1 << 1
	G1FlagIsRLECompressed G1ElementFlags = 1 << 2
	G1FlagIsR8G8B8Palette G1ElementFlags = 1 << 3
	G1FlagHasZoomSprites  G1ElementFlags = 1 << 4
	G1FlagNoZoomDraw      G1ElementFlags = 1 << 5
	G1FlagDuplicatePrev   G1ElementFlags = 1 << 6
)

// G1Header is the file header
type G1Header struct {
	NumEntries uint32
	TotalSize  uint32
}

// G1Element32 is the on-disk element header (16 bytes)
type G1Element32 struct {
	Offset     uint32
	Width      int16
	Height     int16
	XOffset    int16
	YOffset    int16
	Flags      G1ElementFlags
	ZoomOffset int16
}

// G1Element is a parsed sprite element with data pointer
type G1Element struct {
	Width      int16
	Height     int16
	XOffset    int16
	YOffset    int16
	Flags      G1ElementFlags
	ZoomOffset int16
	Data       []byte // raw pixel/RLE data
}

// G1File holds all parsed G1 data
type G1File struct {
	Header       G1Header
	Elements     []G1Element
	Palette      [256]color.RGBA
	totalSprites uint32 // Track total sprite count for dynamic loading
}

// ToRGBA converts a G1Element to an RGBA image using the provided palette
func (e *G1Element) ToRGBA(palette [256]color.RGBA) *image.RGBA {
	width := int(e.Width)
	height := int(e.Height)
	if width <= 0 || height <= 0 {
		return nil
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	var indices []byte
	if e.Flags&G1FlagIsRLECompressed != 0 {
		decoded, err := decodeRLEElement(e.Data, width, height)
		if err != nil {
			return nil
		}
		indices = decoded
	} else {
		indices = e.Data
	}

	if len(indices) < width*height {
		return nil
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := indices[y*width+x]
			c := palette[idx]
			img.SetRGBA(x, y, c)
		}
	}

	return img
}

// LoadG1 loads and parses a G1.DAT file using the in-memory parser.
// This implementation centralizes parsing by reading the file bytes with
// `LoadDatFile` and delegating the heavy-lifting to `ParseG1`.
func LoadG1(path string) (*G1File, error) {
	// Read whole file into memory
	data, err := LoadDatFile(path)
	if err != nil {
		return nil, fmt.Errorf("load dat: %w", err)
	}

	// Parse the G1 table from the in-memory buffer
	g1, err := ParseG1(data)
	if err != nil {
		return nil, fmt.Errorf("parse g1: %w", err)
	}

	// Handle elements marked as duplicate of previous element: the element is
	// a full copy of the previous one with its own x/y offsets added as deltas.
	// OpenLoco reference: src/OpenLoco/src/Objects/ObjectImageTable.cpp:36-41
	for i := 1; i < len(g1.Elements); i++ {
		if (g1.Elements[i].Flags & G1FlagDuplicatePrev) != 0 {
			dx, dy := g1.Elements[i].XOffset, g1.Elements[i].YOffset
			g1.Elements[i] = g1.Elements[i-1]
			g1.Elements[i].XOffset += dx
			g1.Elements[i].YOffset += dy
		}
	}

	// The Steam G1.DAT is missing two localised tutorial icons and a smaller
	// font variant. Realign to the disc layout so all sprite indices ≥3549
	// (title logo, globe animation, tabs, cursors, …) match ImageIds.
	// OpenLoco reference: src/OpenLoco/src/Graphics/Gfx.cpp loadG1
	if g1.Header.NumEntries == G1ExpectedCountSteam {
		// Temporarily convert zoom offsets to absolute indices so the shuffle
		// below cannot break the relative links.
		for i := range g1.Elements {
			if g1.Elements[i].Flags&G1FlagHasZoomSprites != 0 {
				g1.Elements[i].ZoomOffset = int16(i) - g1.Elements[i].ZoomOffset
			}
		}

		els := make([]G1Element, G1ExpectedCountDisc)
		copy(els, g1.Elements[:3549])
		// Shift everything from 3549 up by two; fill the two missing tutorial
		// icons with the closest variant.
		copy(els[3551:], g1.Elements[3549:])
		els[3549] = els[3551]
		els[3550] = els[3551]
		// Extra font variant: duplicate 223 glyphs from 1788.
		copy(els[3898:3898+223], els[1788:1788+223])
		g1.Elements = els
		g1.Header.NumEntries = G1ExpectedCountDisc

		// Restore relative zoom offsets.
		for i := range g1.Elements {
			if g1.Elements[i].Flags&G1FlagHasZoomSprites != 0 {
				g1.Elements[i].ZoomOffset = int16(i) - g1.Elements[i].ZoomOffset
			}
		}
		fmt.Printf("G1: Steam layout realigned to disc layout (%d entries)\n", g1.Header.NumEntries)
	}

	fmt.Printf("G1: %d entries, %d bytes total\n", g1.Header.NumEntries, g1.Header.TotalSize)

	// Validate entry count (preserve original warning behavior)
	if g1.Header.NumEntries != G1ExpectedCountDisc && g1.Header.NumEntries != G1ExpectedCountSteam {
		fmt.Printf("Warning: unexpected G1 entry count %d (expected %d or %d)\n",
			g1.Header.NumEntries, G1ExpectedCountDisc, G1ExpectedCountSteam)
	}

	// Ensure we have a usable palette. ParseG1 attempts to load the palette
	// but may fail silently; keep the original fallback behavior to a
	// grayscale palette when the parsed palette appears empty.
	needFallback := true
	nonZeroCount := 0
	for i := 1; i < 256; i++ {
		if g1.Palette[i].A != 0 || g1.Palette[i].R != 0 || g1.Palette[i].G != 0 || g1.Palette[i].B != 0 {
			needFallback = false
			nonZeroCount++
		}
	}
	fmt.Printf("Palette check: %d non-zero entries, needFallback=%v\n", nonZeroCount, needFallback)
	if needFallback {
		fmt.Printf("Warning: failed to load palette, using grayscale fallback\n")
		for i := 0; i < 256; i++ {
			g1.Palette[i] = color.RGBA{uint8(i), uint8(i), uint8(i), 255}
		}
		// make index 0 transparent
		g1.Palette[0] = color.RGBA{0, 0, 0, 0}
	}
	// Debug: print a few palette entries
	fmt.Printf("Palette samples: [1]=%v [10]=%v [50]=%v [100]=%v [200]=%v [255]=%v\n",
		g1.Palette[1], g1.Palette[10], g1.Palette[50], g1.Palette[100], g1.Palette[200], g1.Palette[255])

	// Initialize sprite counter for dynamic loading
	g1.totalSprites = g1.Header.NumEntries
	fmt.Printf("Initialized G1 sprite pool: %d base sprites\n", g1.totalSprites)

	return g1, nil
}

// loadPalette extracts the palette from the default palette element
func (g1 *G1File) loadPalette() error {
	if DefaultPaletteIndex >= len(g1.Elements) {
		return errors.New("palette index out of range")
	}

	elem := &g1.Elements[DefaultPaletteIndex]
	if elem.Flags&G1FlagIsR8G8B8Palette == 0 {
		return errors.New("palette element doesn't have R8G8B8 flag")
	}

	// Palette data is sequential R, G, B bytes
	// xOffset indicates starting palette index, width indicates count
	startIdx := int(elem.XOffset)
	count := int(elem.Width)

	if len(elem.Data) < count*3 {
		return fmt.Errorf("palette data too small: %d < %d", len(elem.Data), count*3)
	}

	for i := 0; i < count; i++ {
		idx := startIdx + i
		if idx >= 0 && idx < 256 {
			// Data is B, G, R order (Windows BITMAPINFO style)
			g1.Palette[idx] = color.RGBA{
				R: elem.Data[i*3+2],
				G: elem.Data[i*3+1],
				B: elem.Data[i*3+0],
				A: 255,
			}
		}
	}

	// Make index 0 transparent (typically the transparent color)
	g1.Palette[0] = color.RGBA{0, 0, 0, 0}

	// Indices 7-9 and 246-254 are "primary remap" slots — intentionally absent from
	// the base palette. They are filled at draw time by PaletteMap::getForColour()
	// which redirects them to actual palette entries for the chosen colour.
	// PaletteMap is now implemented in GetPaletteMap() + DecodeSpriteMapped().
	//
	// For world sprites (buildings, vehicles) that still use GetSprite() without a
	// colour remap, we provide a visible placeholder (dark blue) so they are not
	// invisible. Proper company-colour support requires calling GetSpriteColoured()
	// with the correct company colour index.
	//
	// OpenLoco reference: src/OpenLoco/src/Graphics/PaletteMap.cpp
	//   primaryRemap5=0xF8(248) … primaryRemapB=0xFE(254)
	g1.Palette[248] = color.RGBA{R: 20, G: 30, B: 60, A: 255}
	g1.Palette[249] = color.RGBA{R: 30, G: 50, B: 90, A: 255}
	g1.Palette[250] = color.RGBA{R: 40, G: 70, B: 120, A: 255}
	g1.Palette[251] = color.RGBA{R: 50, G: 90, B: 150, A: 255}
	g1.Palette[252] = color.RGBA{R: 60, G: 110, B: 180, A: 255}

	// Index 255 — G1 sprites 3746+ use it as a remap template marker; those are
	// never rendered directly so this value is only a safety placeholder.
	g1.Palette[255] = color.RGBA{R: 71, G: 175, B: 39, A: 255}

	return nil
}

// PaletteMapG1Base is the G1 index of the first colour palette map (Colour::black = 0).
// The map for Colour N is at G1 index (PaletteMapG1Base + N).
// Each element is 256×1 raw bytes: outputIndex = map[inputIndex].
//
// OpenLoco reference: src/OpenLoco/src/Graphics/PaletteMap.cpp
//   kDefaultPaletteToG1Offset: ExtColour 0 → ImageIds::paletteMapBlack = 2170
//   ExtColour 1 → paletteMapGrey = 2171, …
const PaletteMapG1Base = 2170

// GetPaletteMap loads the 256-byte index remap table for the given colour index
// (0 = black, 1 = grey, 2 = white, …) from the G1 file.
// Returns a 256-element slice where result[srcIdx] = destIdx in the base palette.
//
// OpenLoco reference: src/OpenLoco/src/Graphics/PaletteMap.cpp getForColour()
func (g1 *G1File) GetPaletteMap(colourIdx int) ([]byte, error) {
	g1Index := PaletteMapG1Base + colourIdx
	if g1Index < 0 || g1Index >= len(g1.Elements) {
		return nil, fmt.Errorf("palette map index %d (g1[%d]) out of range", colourIdx, g1Index)
	}
	e := &g1.Elements[g1Index]
	if int(e.Width)*int(e.Height) != 256 {
		return nil, fmt.Errorf("palette map g1[%d] is not 256 bytes (w=%d h=%d)", g1Index, e.Width, e.Height)
	}
	if len(e.Data) < 256 {
		return nil, fmt.Errorf("palette map g1[%d] data too short: %d", g1Index, len(e.Data))
	}
	result := make([]byte, 256)
	copy(result, e.Data[:256])
	return result, nil
}

// GetPaletteMapAt loads a 256-byte index remap table from any G1 element index directly.
// Unlike GetPaletteMap, this takes a raw G1 index rather than an ExtColour offset.
// Used for water blend palette maps stored inside object sprite pools (e.g. WaterObject+41).
//
// OpenLoco reference: src/OpenLoco/src/Objects/ObjectManager.cpp updateWaterPalette()
//   paletteImageId = waterObj->image + Water::ImageIds::kColourPalette (= 41)
func (g1 *G1File) GetPaletteMapAt(g1Index int) ([]byte, error) {
	if g1Index < 0 || g1Index >= len(g1.Elements) {
		return nil, fmt.Errorf("palette map g1[%d] out of range", g1Index)
	}
	e := &g1.Elements[g1Index]
	if len(e.Data) < 256 {
		return nil, fmt.Errorf("palette map g1[%d] data too short: %d bytes", g1Index, len(e.Data))
	}
	result := make([]byte, 256)
	copy(result, e.Data[:256])
	return result, nil
}

// DecodeSpriteMapped decodes a sprite applying a PaletteMap remap before the palette lookup.
// For each pixel: finalColour = palette[paletteMap[pixelIndex]].
// Pass a nil paletteMap to use the identity (same as DecodeSprite).
//
// OpenLoco reference: src/OpenLoco/src/Graphics/DrawSpriteHelper.hpp blitPixel (remap mode)
func (g1 *G1File) DecodeSpriteMapped(index int, paletteMap []byte) (*image.RGBA, error) {
	if index < 0 || index >= len(g1.Elements) {
		return nil, fmt.Errorf("sprite index %d out of range", index)
	}
	if len(paletteMap) < 256 {
		return g1.DecodeSprite(index)
	}

	elem := &g1.Elements[index]
	if elem.Width <= 0 || elem.Height <= 0 {
		return nil, fmt.Errorf("invalid sprite dimensions: %dx%d", elem.Width, elem.Height)
	}

	w, h := int(elem.Width), int(elem.Height)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	if elem.Flags&G1FlagIsRLECompressed != 0 {
		if err := g1.decodeRLEMapped(elem, img, paletteMap); err != nil {
			return nil, fmt.Errorf("decode RLE mapped: %w", err)
		}
	} else {
		if err := g1.decodeRawMapped(elem, img, paletteMap); err != nil {
			return nil, fmt.Errorf("decode raw mapped: %w", err)
		}
	}
	return img, nil
}

// decodeRawMapped decodes uncompressed data with palette remap.
func (g1 *G1File) decodeRawMapped(elem *G1Element, img *image.RGBA, palMap []byte) error {
	w, h := int(elem.Width), int(elem.Height)
	if len(elem.Data) < w*h {
		return fmt.Errorf("raw data too small: %d < %d", len(elem.Data), w*h)
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			src := elem.Data[y*w+x]
			dst := palMap[src]
			img.SetRGBA(x, y, g1.Palette[dst])
		}
	}
	return nil
}

// decodeRLEMapped decodes RLE data with palette remap.
func (g1 *G1File) decodeRLEMapped(elem *G1Element, img *image.RGBA, palMap []byte) error {
	w, h := int(elem.Width), int(elem.Height)
	data := elem.Data
	if len(data) < h*2 {
		return errors.New("RLE data too small for line offsets")
	}
	for y := 0; y < h; y++ {
		lineOffset := int(binary.LittleEndian.Uint16(data[y*2 : y*2+2]))
		if lineOffset >= len(data) {
			continue
		}
		p := lineOffset
		for {
			if p+2 > len(data) {
				break
			}
			dataSize := int(data[p])
			firstX := int(data[p+1])
			p += 2
			isLast := (dataSize & 0x80) != 0
			dataSize &= 0x7F
			if p+dataSize > len(data) {
				break
			}
			for i := 0; i < dataSize; i++ {
				x := firstX + i
				if x >= 0 && x < w {
					src := data[p+i]
					dst := palMap[src]
					img.SetRGBA(x, y, g1.Palette[dst])
				}
			}
			p += dataSize
			if isLast {
				break
			}
		}
	}
	return nil
}

// DecodeSpriteMaskOnly decodes a blend-mask sprite into a stencil image where
// every pixel that is explicitly present in the sprite data becomes opaque white
// and all absent (gap) positions remain transparent.
//
// For RLE-compressed blend sprites (e.g. water +35) the palette index value is
// irrelevant — OpenLoco's withBlend() mode applies the water colour remap to
// every pixel that appears in the RLE run, regardless of whether its index is 0
// or a small shading value (1-4, etc.).  We therefore use decodeRLEPresence so
// the full diamond shape is treated as the blend region.
//
// For uncompressed sprites the blend region is still defined by palette index 0
// (legacy behaviour preserved for non-water blend sprites).
//
// OpenLoco reference: src/OpenLoco/src/Graphics/DrawSpriteHelper.hpp (blend pixel mode)
func (g1 *G1File) DecodeSpriteMaskOnly(index int) (*image.RGBA, error) {
	if index < 0 || index >= len(g1.Elements) {
		return nil, fmt.Errorf("sprite index %d out of range", index)
	}
	elem := &g1.Elements[index]
	if elem.Width <= 0 || elem.Height <= 0 {
		return nil, fmt.Errorf("invalid sprite dimensions: %dx%d", elem.Width, elem.Height)
	}
	w, h := int(elem.Width), int(elem.Height)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	opaque := color.RGBA{255, 255, 255, 255}

	if elem.Flags&G1FlagIsRLECompressed != 0 {
		// For RLE blend sprites: every pixel present in the RLE stream is the
		// blend region.  Use the presence bitmap rather than checking index == 0.
		presence, err := decodeRLEPresence(elem.Data, w, h)
		if err != nil {
			return nil, fmt.Errorf("decode mask RLE presence: %w", err)
		}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if presence[y*w+x] {
					img.SetRGBA(x, y, opaque)
				}
			}
		}
	} else {
		// Uncompressed: palette index 0 marks the blend region.
		if len(elem.Data) < w*h {
			return nil, fmt.Errorf("mask raw data too small: %d < %d", len(elem.Data), w*h)
		}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if elem.Data[y*w+x] == 0 {
					img.SetRGBA(x, y, opaque)
				}
			}
		}
	}
	return img, nil
}

// DecodeSprite decodes a sprite element to an RGBA image
func (g1 *G1File) DecodeSprite(index int) (*image.RGBA, error) {
	if index < 0 || index >= len(g1.Elements) {
		return nil, fmt.Errorf("sprite index %d out of range", index)
	}

	elem := &g1.Elements[index]
	if elem.Width <= 0 || elem.Height <= 0 {
		return nil, fmt.Errorf("invalid sprite dimensions: %dx%d", elem.Width, elem.Height)
	}

	w := int(elem.Width)
	h := int(elem.Height)

	// Create output image
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Check if RLE compressed
	if elem.Flags&G1FlagIsRLECompressed != 0 {
		if err := g1.decodeRLE(elem, img); err != nil {
			return nil, fmt.Errorf("decode RLE: %w", err)
		}
	} else {
		// Raw palette indices
		if err := g1.decodeRaw(elem, img); err != nil {
			return nil, fmt.Errorf("decode raw: %w", err)
		}
	}

	return img, nil
}

// decodeRaw decodes uncompressed palette index data
func (g1 *G1File) decodeRaw(elem *G1Element, img *image.RGBA) error {
	w := int(elem.Width)
	h := int(elem.Height)

	if len(elem.Data) < w*h {
		return fmt.Errorf("raw data too small: %d < %d", len(elem.Data), w*h)
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := elem.Data[y*w+x]
			img.SetRGBA(x, y, g1.Palette[idx])
		}
	}

	return nil
}

// decodeRLE decodes RLE compressed sprite data
func (g1 *G1File) decodeRLE(elem *G1Element, img *image.RGBA) error {
	w := int(elem.Width)
	h := int(elem.Height)
	data := elem.Data

	// RLE format: first h*2 bytes are uint16 LE line offsets
	if len(data) < h*2 {
		return errors.New("RLE data too small for line offsets")
	}

	for y := 0; y < h; y++ {
		lineOffset := int(binary.LittleEndian.Uint16(data[y*2 : y*2+2]))
		if lineOffset >= len(data) {
			continue // Skip invalid lines
		}

		p := lineOffset
		for {
			if p+2 > len(data) {
				break
			}

			dataSize := int(data[p])
			firstX := int(data[p+1])
			p += 2

			isLast := (dataSize & 0x80) != 0
			dataSize &= 0x7F

			if p+dataSize > len(data) {
				break
			}

			// Copy pixel data
			for i := 0; i < dataSize; i++ {
				x := firstX + i
				if x >= 0 && x < w {
					idx := data[p+i]
					img.SetRGBA(x, y, g1.Palette[idx])
				}
			}
			p += dataSize

			if isLast {
				break
			}
		}
	}

	return nil
}

// GetSpriteCount returns the number of sprites
func (g1 *G1File) GetSpriteCount() int {
	return len(g1.Elements)
}

// GetPalette returns the color palette
func (g1 *G1File) GetPalette() []color.RGBA {
	return g1.Palette[:]
}

// LoadImageTable loads sprites from a DAT object's image table into the global G1 pool.
// This appends the sprites from the object to the G1 sprite array and returns the
// offset where they were added.
//
// The data format is identical to G1.DAT format:
//   [G1Header: 8 bytes]
//   [G1Element32 array: numEntries * 16 bytes]
//   [Pixel data: referenced by element offsets]
//
// Elements with the DuplicatePrevious flag reuse the previous sprite's data but with
// adjusted x/y offsets.
//
// OpenLoco reference: src/OpenLoco/src/Objects/ObjectImageTable.cpp::loadImageTable
func (g1 *G1File) LoadImageTable(data []byte) (objects.ImageTableResult, error) {
	if len(data) < 8 {
		return objects.ImageTableResult{}, errors.New("data too small for G1 header")
	}

	// Parse header
	header := G1Header{
		NumEntries: binary.LittleEndian.Uint32(data[0:4]),
		TotalSize:  binary.LittleEndian.Uint32(data[4:8]),
	}

	// Calculate total length consumed. Arithmetic is done in int64: a malformed
	// header (huge NumEntries/TotalSize) must fail the size check below rather
	// than wrap around uint32 and pass it.
	headerSize := int64(8)
	elementTableSize := int64(header.NumEntries) * 16 // G1Element32 is 16 bytes
	tableLength := headerSize + elementTableSize + int64(header.TotalSize)

	if int64(len(data)) < headerSize+elementTableSize {
		return objects.ImageTableResult{}, fmt.Errorf("data too small: need %d bytes for elements",
			headerSize+elementTableSize)
	}

	// Remember starting index for this object's sprites
	imageOffset := g1.totalSprites

	// Parse elements
	elemData := data[headerSize:]
	pixelDataStart := headerSize + elementTableSize

	for i := uint32(0); i < header.NumEntries; i++ {
		offset := int64(i) * 16
		elem32 := G1Element32{
			Offset:     binary.LittleEndian.Uint32(elemData[offset+0 : offset+4]),
			Width:      int16(binary.LittleEndian.Uint16(elemData[offset+4 : offset+6])),
			Height:     int16(binary.LittleEndian.Uint16(elemData[offset+6 : offset+8])),
			XOffset:    int16(binary.LittleEndian.Uint16(elemData[offset+8 : offset+10])),
			YOffset:    int16(binary.LittleEndian.Uint16(elemData[offset+10 : offset+12])),
			Flags:      G1ElementFlags(binary.LittleEndian.Uint16(elemData[offset+12 : offset+14])),
			ZoomOffset: int16(binary.LittleEndian.Uint16(elemData[offset+14 : offset+16])),
		}

		// Convert to G1Element
		elem := G1Element{
			Width:      elem32.Width,
			Height:     elem32.Height,
			XOffset:    elem32.XOffset,
			YOffset:    elem32.YOffset,
			Flags:      elem32.Flags,
			ZoomOffset: elem32.ZoomOffset,
		}

		// DuplicatePrevious: full copy of the previous element with this
		// element's x/y offsets added as deltas.
		// OpenLoco reference: src/OpenLoco/src/Objects/ObjectImageTable.cpp:36-41
		if (elem.Flags&G1FlagDuplicatePrev) != 0 && len(g1.Elements) > 0 {
			prev := g1.Elements[len(g1.Elements)-1]
			elem = prev
			elem.XOffset = prev.XOffset + elem32.XOffset
			elem.YOffset = prev.YOffset + elem32.YOffset
		} else {
			// Extract pixel data. Size runs to the next element's offset (or
			// TotalSize for the last element); a non-monotonic or out-of-range
			// offset table yields dataSize <= 0 and the element keeps nil Data
			// instead of wrapping around and panicking on the slice bounds.
			dataOffset := pixelDataStart + int64(elem32.Offset)
			var dataSize int64
			if i+1 < header.NumEntries {
				nextOffset := binary.LittleEndian.Uint32(elemData[int64(i+1)*16 : int64(i+1)*16+4])
				dataSize = int64(nextOffset) - int64(elem32.Offset)
			} else {
				dataSize = int64(header.TotalSize) - int64(elem32.Offset)
			}

			if dataSize > 0 && dataOffset+dataSize <= int64(len(data)) {
				elem.Data = make([]byte, dataSize)
				copy(elem.Data, data[dataOffset:dataOffset+dataSize])
			}
		}

		g1.Elements = append(g1.Elements, elem)
		g1.totalSprites++
	}

	return objects.ImageTableResult{
		ImageOffset: imageOffset,
		TableLength: uint32(tableLength),
	}, nil
}

// GetTotalSprites returns the total number of sprites including dynamically loaded ones.
// This matches OpenLoco's ObjectManager::getTotalNumImages().
//
// OpenLoco reference: src/OpenLoco/src/Objects/ObjectImageTable.cpp::getTotalNumImages
func (g1 *G1File) GetTotalSprites() uint32 {
	return g1.totalSprites
}

// SetTotalSprites sets the sprite counter. Used to initialize after loading G1.DAT.
//
// OpenLoco reference: src/OpenLoco/src/Objects/ObjectImageTable.cpp::setTotalNumImages
func (g1 *G1File) SetTotalSprites(count uint32) {
	g1.totalSprites = count
}
