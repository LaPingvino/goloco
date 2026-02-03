package assets

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// Minimal DAT/G1 loader skeleton.
// This file provides a tiny API to open a .dat file, discover embedded G1
// headers and parse a G1 image table from an in-memory byte slice.
//
// We add `ParseG1` which parses a G1 from a byte slice and improve
// `FindG1Header` to do a slightly more robust discovery check.

// ErrG1NotFound is returned when no G1 header is found in the file.
var ErrG1NotFound = errors.New("G1 header not found")

// Note: G1Header and related types are defined in g1.go

// LoadDatFile opens the provided path and returns its contents as a reader.
// Future functions will parse the reader to extract G1 image tables.
func LoadDatFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	sz := fi.Size()
	if sz == 0 {
		return nil, io.EOF
	}

	buf := make([]byte, sz)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// FindG1Header searches the provided data for a G1 header signature.
// It returns the offset where the header starts or an error if not found.
//
// The search is tolerant and verifies there is enough room for a header
// once the magic is found (8 bytes: NumEntries + TotalSize).
func FindG1Header(data []byte) (int, error) {
	// G1 files in OpenLoco typically begin with ASCII "G1" followed by a 0.
	// We search for the sequence and ensure there are at least 8 bytes
	// following the magic to contain the header (two uint32 fields).
	magic := []byte{'G', '1', 0}
	for i := 0; i <= len(data)-len(magic); i++ {
		if data[i] == magic[0] && data[i+1] == magic[1] && data[i+2] == magic[2] {
			// Make sure header fits
			if i+8 <= len(data) {
				return i, nil
			}
			// If not enough room for header, continue searching
		}
	}
	return 0, ErrG1NotFound
}

/*
ParseG1 parses a G1 image table directly from the provided byte slice.
It supports a few discovery strategies to make parsing real DAT files
more robust: try parsing at offset 0, use the tolerant FindG1Header, and
as a final fallback scan for the ASCII sequence 'G','1' and attempt to
parse at those offsets. The core parsing is performed by the inner
parseAt(offset) helper which returns (g1, nil) on success.
*/
func ParseG1(data []byte) (*G1File, error) {
	const elementHeaderSize = 16 // size of G1Element32 on-disk

	// parseAt attempts to parse a G1 starting exactly at headerOffset.
	parseAt := func(headerOffset int) (*G1File, error) {
		// Need at least 8 bytes for header (NumEntries + TotalSize)
		if headerOffset+8 > len(data) {
			return nil, io.ErrUnexpectedEOF
		}
		// Accept either files that include the ASCII 'G','1' magic or files
		// that expose a header-only layout (no leading magic). We no longer
		// fail early just because the magic isn't present; instead we rely on
		// the subsequent sanity checks (numEntries, header/element ranges)
		// to detect false positives. This permits parsing of DAT files that
		// embed a G1 header without the ASCII magic.
		//
		// If a strict magic is required in future, add an option to enforce it.

		numEntries := int(binary.LittleEndian.Uint32(data[headerOffset : headerOffset+4]))
		totalSize := int(binary.LittleEndian.Uint32(data[headerOffset+4 : headerOffset+8]))

		// Sanity checks
		if numEntries <= 0 {
			return nil, fmt.Errorf("invalid G1 entry count %d", numEntries)
		}

		headersOffset := headerOffset + 8
		headersSize := numEntries * elementHeaderSize
		if headersOffset+headersSize > len(data) {
			return nil, io.ErrUnexpectedEOF
		}

		// Parse element headers into temporary array
		type tmpHeader struct {
			Offset     uint32
			Width      int16
			Height     int16
			XOffset    int16
			YOffset    int16
			Flags      uint16
			ZoomOffset int16
		}

		tmpHeaders := make([]tmpHeader, numEntries)
		for i := 0; i < numEntries; i++ {
			hOff := headersOffset + i*elementHeaderSize
			if hOff+elementHeaderSize > len(data) {
				return nil, io.ErrUnexpectedEOF
			}
			tmpHeaders[i].Offset = binary.LittleEndian.Uint32(data[hOff : hOff+4])
			tmpHeaders[i].Width = int16(binary.LittleEndian.Uint16(data[hOff+4 : hOff+6]))
			tmpHeaders[i].Height = int16(binary.LittleEndian.Uint16(data[hOff+6 : hOff+8]))
			tmpHeaders[i].XOffset = int16(binary.LittleEndian.Uint16(data[hOff+8 : hOff+10]))
			tmpHeaders[i].YOffset = int16(binary.LittleEndian.Uint16(data[hOff+10 : hOff+12]))
			tmpHeaders[i].Flags = binary.LittleEndian.Uint16(data[hOff+12 : hOff+14])
			tmpHeaders[i].ZoomOffset = int16(binary.LittleEndian.Uint16(data[hOff+14 : hOff+16]))
		}

		// Element data block follows element headers
		elementDataOffset := headersOffset + headersSize
		if elementDataOffset > len(data) {
			return nil, io.ErrUnexpectedEOF
		}
		elementData := data[elementDataOffset:]
		availDataLen := len(elementData)
		if totalSize > availDataLen {
			// tolerate truncated element data
			totalSize = availDataLen
		}

		g1 := &G1File{
			Header:   G1Header{NumEntries: uint32(numEntries), TotalSize: uint32(totalSize)},
			Elements: make([]G1Element, numEntries),
		}

		// Convert headers into elements
		for i := 0; i < numEntries; i++ {
			h := tmpHeaders[i]
			elem := &g1.Elements[i]
			elem.Width = h.Width
			elem.Height = h.Height
			elem.XOffset = h.XOffset
			elem.YOffset = h.YOffset
			elem.Flags = G1ElementFlags(h.Flags)
			elem.ZoomOffset = h.ZoomOffset

			startOff := int(h.Offset)
			var endOff int
			if i+1 < numEntries {
				endOff = int(tmpHeaders[i+1].Offset)
			} else {
				endOff = totalSize
			}

			// Debug: log offsets for specific sprites (disabled for now)
			_ = startOff
			_ = endOff

			if startOff >= 0 && endOff >= 0 && startOff < endOff && endOff <= totalSize {
				elem.Data = elementData[startOff:endOff]
			} else {
				// invalid ranges are allowed (leave nil), but still return g1 -
				// many real g1s can be parsed even with some malformed entries.
				elem.Data = nil
			}
		}

		// Attempt to load palette (may fail; caller has fallback logic)
		_ = g1.loadPalette() // ignore error here; higher-level code handles fallback

		return g1, nil
	}

	// Strategy 1: try parsing with header at offset 0 (some DATs embed G1 at start)
	if g1, err := parseAt(0); err == nil {
		return g1, nil
	}

	// Strategy 2: try using the tolerant FindG1Header (looks for 'G','1',0)
	if off, err := FindG1Header(data); err == nil {
		if g1, err := parseAt(off); err == nil {
			return g1, nil
		}
	}

	// Strategy 3: scan for ASCII 'G','1' occurrences and try parsing at each.
	// Some DATs place the header with different padding or without the zero.
	for i := 0; i+1 < len(data); i++ {
		if data[i] == 'G' && data[i+1] == '1' {
			// skip if this looks like the same offset we already tried
			if i == 0 {
				continue
			}
			if off, _ := FindG1Header(data[i:]); off == 0 {
				// If FindG1Header succeeds on the slice, it indicates a 'G1\0' here;
				// parseAt using absolute offset i
				if g1, err := parseAt(i); err == nil {
					return g1, nil
				}
			} else {
				// Even if there's no immediate zero byte, attempt parseAt(i)
				if g1, err := parseAt(i); err == nil {
					return g1, nil
				}
			}
		}
	}

	return nil, ErrG1NotFound
}

// readUint32LE reads a uint32 from data at offset using little-endian.
func readUint32LE(data []byte, offset int) (uint32, error) {
	if offset+4 > len(data) {
		return 0, io.ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint32(data[offset : offset+4]), nil
}

// readUint16LE reads a uint16 from data at offset using little-endian.
func readUint16LE(data []byte, offset int) (uint16, error) {
	if offset+2 > len(data) {
		return 0, io.ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint16(data[offset : offset+2]), nil
}
