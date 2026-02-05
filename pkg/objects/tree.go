package objects

import (
	"encoding/binary"
	"fmt"
)

// TreeObjectFlags define special behavior for tree types.
//
// OpenLoco reference: src/OpenLoco/src/Objects/TreeObject.h
type TreeObjectFlags uint16

const (
	TreeFlagHasSnowVariation TreeObjectFlags = 1 << 0
	TreeFlagUnk1             TreeObjectFlags = 1 << 1
	TreeFlagHighAltitude     TreeObjectFlags = 1 << 2
	TreeFlagLowAltitude      TreeObjectFlags = 1 << 3
	TreeFlagRequiresWater    TreeObjectFlags = 1 << 4
	TreeFlagUnk5             TreeObjectFlags = 1 << 5
	TreeFlagDroughtResistant TreeObjectFlags = 1 << 6
	TreeFlagHasShadow        TreeObjectFlags = 1 << 7
)

// TreeObject represents a tree type (oak, birch, etc.)
//
// OpenLoco reference: src/OpenLoco/src/Objects/TreeObject.h
// struct TreeObject — sizeof == 0x4C (76 bytes)
type TreeObject struct {
	Header *ObjectHeader

	Name          string
	InitialHeight uint8  // 0x02
	Height        uint8  // 0x03
	Var04         uint8  // 0x04
	Var05         uint8  // 0x05
	NumRotations  uint8  // 0x06 (1, 2, or 4)
	Growth        uint8  // 0x07 (number of tree size images)
	Flags         TreeObjectFlags // 0x08
	Var3C         uint8  // 0x3C — bitmask controlling which season variants exist
	SeasonState   uint8  // 0x3D
	Colours       uint32 // 0x44

	// Calculated from image table during load
	Sprites     [6]uint32 // Base image indices for 6 season variants
	SnowSprites [6]uint32 // Snow variant base indices
	ShadowImageOffset uint16

	ImageOffset uint32 // Base index in global G1 pool
}

// ParseTreeObject parses a tree object from decompressed data (no G1 pool).
func ParseTreeObject(header *ObjectHeader, data []byte) (*TreeObject, error) {
	return ParseTreeObjectWithG1(header, data, nil)
}

// ParseTreeObjectWithG1 parses a tree object and loads sprites into the G1 pool.
//
// OpenLoco reference: src/OpenLoco/src/Objects/TreeObject.cpp  TreeObject::load()
func ParseTreeObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*TreeObject, error) {
	if len(data) < 0x4C {
		return nil, fmt.Errorf("tree object data too small: %d bytes (need 0x4C)", len(data))
	}

	tree := &TreeObject{
		Header:        header,
		InitialHeight: data[0x02],
		Height:        data[0x03],
		Var04:         data[0x04],
		Var05:         data[0x05],
		NumRotations:  data[0x06],
		Growth:        data[0x07],
		Flags:         TreeObjectFlags(binary.LittleEndian.Uint16(data[0x08:0x0A])),
		Var3C:         data[0x3C],
		SeasonState:   data[0x3D],
		Colours:       binary.LittleEndian.Uint32(data[0x44:0x48]),
	}

	// Skip past fixed structure
	offset := 0x4C

	// Parse string table
	for offset < len(data) && data[offset] != 0xFF {
		offset++ // skip language ID
		strStart := offset
		for offset < len(data) && data[offset] != 0 {
			offset++
		}
		if tree.Name == "" {
			tree.Name = string(data[strStart:offset])
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF end marker
	}

	// Load sprites into G1 pool
	if g1 != nil {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("loading tree image table to G1: %w", err)
		}

		tree.ImageOffset = imgRes.ImageOffset

		// Initialize all sprite arrays to base offset
		for v := 0; v < 6; v++ {
			tree.Sprites[v] = imgRes.ImageOffset
			tree.SnowSprites[v] = imgRes.ImageOffset
		}

		// Assign sprite offsets per season variant based on var3C bitmask
		// This matches OpenLoco TreeObject.cpp:load() lines 16-28
		numVariantImages := uint32(tree.NumRotations) * uint32(tree.Growth)
		totalImageCount := uint32(0)

		for v := 0; v < 6; v++ {
			if tree.Var3C&(1<<v) == 0 {
				continue
			}
			tree.Sprites[v] = imgRes.ImageOffset + totalImageCount
			totalImageCount += numVariantImages
		}

		numPrimaryImages := totalImageCount

		// Fallback assignments matching OpenLoco TreeObject.cpp lines 30-44
		if tree.Var3C&(1<<5) == 0 && tree.Var3C&(1<<4) != 0 {
			tree.Sprites[5] = tree.Sprites[4]
		}
		if tree.Var3C&(1<<5) == 0 && tree.Var3C&(1<<1) != 0 {
			tree.Sprites[5] = tree.Sprites[1]
		}
		if tree.Var3C&(1<<4) == 0 && tree.Var3C&(1<<1) != 0 {
			tree.Sprites[4] = tree.Sprites[1]
		}
		if tree.Var3C&(1<<1) == 0 && tree.Var3C&(1<<0) != 0 {
			tree.Sprites[1] = tree.Sprites[0]
		}
		if tree.Var3C&(1<<0) == 0 && tree.Var3C&(1<<1) != 0 {
			tree.Sprites[0] = tree.Sprites[1]
		}

		// Snow variants
		if tree.Flags&TreeFlagHasSnowVariation != 0 {
			for v := 0; v < 6; v++ {
				tree.SnowSprites[v] = tree.Sprites[v] + numPrimaryImages
			}
			totalImageCount += numPrimaryImages
		}

		// Shadow offset
		if tree.Flags&TreeFlagHasShadow != 0 {
			tree.ShadowImageOffset = uint16(totalImageCount)
			totalImageCount += totalImageCount
		}
	}

	return tree, nil
}

// HasFlags checks if the tree object has the given flags.
func (t *TreeObject) HasFlags(flags TreeObjectFlags) bool {
	return t.Flags&flags != 0
}
