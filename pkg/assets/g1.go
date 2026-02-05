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
	G1ExpectedCountDisc  = 0x0F4A
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
		indices = decompressRLE(e.Data, width*height)
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

// decompressRLE decompresses RLE compressed data
func decompressRLE(data []byte, expectedLen int) []byte {
	var out []byte
	i := 0
	for i < len(data) && len(out) < expectedLen {
		if i >= len(data) {
			break
		}
		b := data[i]
		i++
		if b&0x80 != 0 {
			count := int(b & 0x7F)
			if i+count > len(data) {
				break
			}
			out = append(out, data[i:i+count]...)
			i += count
		} else {
			count := int(b)
			if i >= len(data) {
				break
			}
			val := data[i]
			i++
			for j := 0; j < count && len(out) < expectedLen; j++ {
				out = append(out, val)
			}
		}
	}
	return out
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

	// Handle elements marked as duplicate of previous element.
	// Some G1s use the DuplicatePrev flag to indicate the element reuses the
	// previous element's image/data. Ensure such elements get a proper Data
	// slice and sensible defaults copied from the previous element when
	// fields are zero-valued.
	for i := 0; i < len(g1.Elements); i++ {
		if (g1.Elements[i].Flags&G1FlagDuplicatePrev) != 0 && i > 0 {
			prev := &g1.Elements[i-1]
			cur := &g1.Elements[i]
			if len(cur.Data) == 0 {
				cur.Data = prev.Data
			}
			if cur.Width == 0 {
				cur.Width = prev.Width
			}
			if cur.Height == 0 {
				cur.Height = prev.Height
			}
			if cur.XOffset == 0 {
				cur.XOffset = prev.XOffset
			}
			if cur.YOffset == 0 {
				cur.YOffset = prev.YOffset
			}
		}
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

	// TEMPORARY FIX: Indices 248-252 and 255 are remappable colors but not in default palette.
	// These are normally remapped at draw time based on context (company colors, terrain type, etc.)
	// For now we'll hardcode them to visible colors so sprites appear.
	// Proper solution requires implementing the full ImageId + PaletteMap system.
	//
	// OpenLoco reference: These indices are remappable "primary" colors that get replaced
	// at draw time based on company colors or other context. See:
	// - src/OpenLoco/src/Graphics/PaletteMap.cpp (color remapping)
	// - src/OpenLoco/src/Graphics/ImageId.h (kMaskPrimary, kFlagPrimary)
	//
	// For the logo, we'll use a dark blue gradient:
	g1.Palette[248] = color.RGBA{R: 20, G: 30, B: 60, A: 255}   // Very dark blue
	g1.Palette[249] = color.RGBA{R: 30, G: 50, B: 90, A: 255}   // Dark blue
	g1.Palette[250] = color.RGBA{R: 40, G: 70, B: 120, A: 255}  // Medium blue (most pixels)
	g1.Palette[251] = color.RGBA{R: 50, G: 90, B: 150, A: 255}  // Lighter blue
	g1.Palette[252] = color.RGBA{R: 60, G: 110, B: 180, A: 255} // Light blue

	// Index 255 is used by terrain template sprites at 3746+
	// Remap to grass green for now (should be based on LandObject color scheme)
	g1.Palette[255] = color.RGBA{R: 71, G: 175, B: 39, A: 255}  // Grass green (from palette index 100)

	return nil
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

	// Calculate total length consumed
	headerSize := uint32(8)
	elementTableSize := header.NumEntries * 16 // G1Element32 is 16 bytes
	tableLength := headerSize + elementTableSize + header.TotalSize

	if uint32(len(data)) < headerSize+elementTableSize {
		return objects.ImageTableResult{}, fmt.Errorf("data too small: need %d bytes for elements",
			headerSize+elementTableSize)
	}

	// Remember starting index for this object's sprites
	imageOffset := g1.totalSprites

	// Parse elements
	elemData := data[headerSize:]
	pixelDataStart := headerSize + elementTableSize

	for i := uint32(0); i < header.NumEntries; i++ {
		offset := i * 16
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

		// Handle DuplicatePrevious flag
		if (elem.Flags&G1FlagDuplicatePrev) != 0 && len(g1.Elements) > 0 {
			// Duplicate previous element but with adjusted offsets
			prevIdx := len(g1.Elements) - 1
			elem.Data = g1.Elements[prevIdx].Data
			elem.Width = g1.Elements[prevIdx].Width
			elem.Height = g1.Elements[prevIdx].Height
			elem.XOffset = g1.Elements[prevIdx].XOffset + elem32.XOffset
			elem.YOffset = g1.Elements[prevIdx].YOffset + elem32.YOffset
		} else {
			// Extract pixel data
			dataOffset := pixelDataStart + elem32.Offset
			if dataOffset < uint32(len(data)) {
				// Calculate data size (goes to next element's offset or end of data)
				var dataSize uint32
				if i+1 < header.NumEntries {
					// Next element's offset
					nextOffset := binary.LittleEndian.Uint32(elemData[(i+1)*16 : (i+1)*16+4])
					dataSize = nextOffset - elem32.Offset
				} else {
					// Last element: use remaining data
					dataSize = header.TotalSize - elem32.Offset
				}

				if dataOffset+dataSize <= uint32(len(data)) {
					elem.Data = make([]byte, dataSize)
					copy(elem.Data, data[dataOffset:dataOffset+dataSize])
				}
			}
		}

		g1.Elements = append(g1.Elements, elem)
		g1.totalSprites++
	}

	return objects.ImageTableResult{
		ImageOffset: imageOffset,
		TableLength: tableLength,
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
