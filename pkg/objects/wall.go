package objects

import (
	"fmt"
)

// WallObjectFlags define special behavior for wall types
type WallObjectFlags uint8

const (
	WallFlagHasPrimaryColour   WallObjectFlags = 1 << 0
	WallFlagHasGlass           WallObjectFlags = 1 << 1
	WallFlagOnlyOnLevelLand    WallObjectFlags = 1 << 2
	WallFlagTwoSided           WallObjectFlags = 1 << 3
	WallFlagHasSecondaryColour WallObjectFlags = 1 << 6
)

// WallObject represents a wall/fence type (hedgerow, stone wall, etc.)
//
// OpenLoco reference: src/OpenLoco/src/Objects/WallObject.h
// Header size: 0x0A (10 bytes)
type WallObject struct {
	Header *ObjectHeader

	Name   string
	Sprite uint32          // Base sprite index in G1 pool
	Flags  WallObjectFlags // Feature flags (colour, glass, etc.)
	Height uint8           // Physical height in SmallZ units
}

// ParseWallObject parses a wall object from decompressed data
func ParseWallObject(header *ObjectHeader, data []byte) (*WallObject, error) {
	return ParseWallObjectWithG1(header, data, nil)
}

// ParseWallObjectWithG1 parses a wall object and loads sprites into the G1 pool.
//
// OpenLoco reference: src/OpenLoco/src/Objects/WallObject.cpp::load
//
// Header layout (10 bytes):
//   0x00: StringId name (2 bytes, placeholder)
//   0x02: uint32 sprite (4 bytes, set after image table load)
//   0x06: uint8 var_06 (tool cursor, unused)
//   0x07: uint8 flags (WallObjectFlags)
//   0x08: uint8 height (SmallZ units)
//   0x09: uint8 var_09 (flags2, unused)
func ParseWallObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*WallObject, error) {
	if len(data) < 0x0A {
		return nil, fmt.Errorf("wall object data too small: %d bytes", len(data))
	}

	wall := &WallObject{
		Header: header,
		Flags:  WallObjectFlags(data[0x07]),
		Height: data[0x08],
	}

	// String table starts after the fixed 10-byte header
	offset := 0x0A

	// Parse string table: (language_id + null-terminated string)*, ends with 0xFF
	for offset < len(data) && data[offset] != 0xFF {
		offset++ // skip language ID
		strStart := offset
		for offset < len(data) && data[offset] != 0 {
			offset++
		}
		if wall.Name == "" {
			wall.Name = string(data[strStart:offset])
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF end marker
	}

	// Load sprites into G1 pool if G1 loader is available
	if g1 != nil && offset < len(data) {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("loading wall image table to G1: %w", err)
		}
		wall.Sprite = imgRes.ImageOffset
	}

	return wall, nil
}
