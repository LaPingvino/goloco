package objects

import "fmt"

// LevelCrossingObject represents a level crossing (road-over-rail barrier) object.
//
// OpenLoco reference: src/OpenLoco/src/Objects/LevelCrossingObject.h
// sizeof(LevelCrossingObject) == 0x12
//
// Sprite layout (8 sprites per frame, (rotation&1)*4 selects NE or SE set):
//
//	Frame 0 (gates open): imageOffset + 0..3 for NE, imageOffset + 4..7 for SE
//	Frame N (animated):   imageOffset + N*8 + (rotation&1)*4 + 0..3
//
// The 4 sprites per rotation are the 4 gate positions at tile corners:
//
//	+0: gate at tile pos {2, 2}    (north corner)
//	+1: gate at tile pos {2, 30}   (east corner)
//	+2: gate at tile pos {30, 2}   (west corner)
//	+3: gate at tile pos {30, 30}  (south corner)
type LevelCrossingObject struct {
	Header        *ObjectHeader
	Name          string
	AnimSpeed     uint8
	ClosingFrames uint8
	ClosedFrames  uint8

	// ImageOffset is the base G1 pool index set by LoadImageTable.
	ImageOffset uint32
}

// ParseLevelCrossingObject parses without loading sprites.
func ParseLevelCrossingObject(header *ObjectHeader, data []byte) (*LevelCrossingObject, error) {
	return ParseLevelCrossingObjectWithG1(header, data, nil)
}

// ParseLevelCrossingObjectWithG1 parses a LevelCrossingObject from decompressed DAT data.
//
// Layout:
//
//	0x00–0x11: fixed struct (18 bytes)
//	0x12:      string table [langID][str\0]...[0xFF]
//	image table (no compatible object headers)
//
// OpenLoco reference: src/OpenLoco/src/Objects/LevelCrossingObject.cpp load()
func ParseLevelCrossingObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*LevelCrossingObject, error) {
	const fixedSize = 0x12
	if len(data) < fixedSize {
		return nil, fmt.Errorf("level crossing object data too small: %d bytes (need %d)", len(data), fixedSize)
	}

	obj := &LevelCrossingObject{
		Header:        header,
		AnimSpeed:     data[0x07],
		ClosingFrames: data[0x08],
		ClosedFrames:  data[0x09],
	}

	offset := fixedSize

	// Walk string table: [langID][str\0]...[0xFF]
	for offset < len(data) && data[offset] != 0xFF {
		offset++ // skip langID
		for offset < len(data) && data[offset] != 0 {
			offset++ // skip string bytes
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF end marker
	}

	if offset > len(data) {
		return nil, fmt.Errorf("level crossing object: offset %d > data len %d", offset, len(data))
	}

	// Load image table
	if g1 != nil {
		result, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("level crossing object: loading image table: %w", err)
		}
		obj.ImageOffset = result.ImageOffset
	}

	return obj, nil
}
