package assets

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
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
	Header   G1Header
	Elements []G1Element
	Palette  [256]color.RGBA
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

	// Debug logging for first few sprites
	if decodeLogCount < 5 {
		fmt.Printf("DecodeSprite(%d): %dx%d, flags=0x%X, dataLen=%d, isRLE=%v\n",
			index, w, h, elem.Flags, len(elem.Data), elem.Flags&G1FlagIsRLECompressed != 0)
		// Print first 20 bytes of data
		if len(elem.Data) > 20 {
			fmt.Printf("  First 20 bytes: %v\n", elem.Data[:20])
		}
		decodeLogCount++
	}

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

var decodeLogCount = 0

// decodeRaw decodes uncompressed palette index data
func (g1 *G1File) decodeRaw(elem *G1Element, img *image.RGBA) error {
	w := int(elem.Width)
	h := int(elem.Height)

	if len(elem.Data) < w*h {
		return fmt.Errorf("raw data too small: %d < %d", len(elem.Data), w*h)
	}

	// Debug: track palette indices used
	idxCount := make(map[uint8]int)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := elem.Data[y*w+x]
			idxCount[idx]++
			img.SetRGBA(x, y, g1.Palette[idx])
		}
	}

	// Log first few sprites' palette usage
	if len(idxCount) > 0 && rawDecodeLogCount < 3 {
		fmt.Printf("Raw sprite palette indices used: %v\n", idxCount)
		rawDecodeLogCount++
	}

	return nil
}

var rawDecodeLogCount = 0

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
