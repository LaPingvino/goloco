package objects

import (
	"encoding/binary"
	"fmt"
)

// BuildingObjectFlags define special behavior for building types.
//
// OpenLoco reference: src/OpenLoco/src/Objects/BuildingObject.h
type BuildingObjectFlags uint8

const (
	BuildingFlagLargeTile      BuildingObjectFlags = 1 << 0 // 2x2 tile
	BuildingFlagMiscBuilding   BuildingObjectFlags = 1 << 1
	BuildingFlagIndestructible BuildingObjectFlags = 1 << 2
	BuildingFlagIsHeadquarters BuildingObjectFlags = 1 << 3
	BuildingFlagHasShadows     BuildingObjectFlags = 1 << 4
)

// BuildingObject represents a building type (house, factory, etc.)
//
// OpenLoco reference: src/OpenLoco/src/Objects/BuildingObject.h
// struct BuildingObject — sizeof == 0xBE (190 bytes)
type BuildingObject struct {
	Header *ObjectHeader

	Name          string
	NumParts      uint8               // 0x06
	NumVariations uint8               // 0x07
	Flags         BuildingObjectFlags // 0x98
	Colours       uint32              // 0x90

	// Building parts: list of part indices per variation, terminated by 0xFF
	VariationParts [32][]uint8 // variationPartsOffset decoded
	PartHeights    []uint8     // height per part

	// Calculated from image table during load
	ImageOffset uint32 // Base sprite index in global G1 pool
}

// ParseBuildingObject parses a building object from decompressed data (no G1 pool).
func ParseBuildingObject(header *ObjectHeader, data []byte) (*BuildingObject, error) {
	return ParseBuildingObjectWithG1(header, data, nil)
}

// ParseBuildingObjectWithG1 parses a building object and loads sprites into the G1 pool.
//
// OpenLoco reference: src/OpenLoco/src/Objects/BuildingObject.cpp  BuildingObject::load()
func ParseBuildingObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*BuildingObject, error) {
	if len(data) < 0xBE {
		return nil, fmt.Errorf("building object data too small: %d bytes (need 0xBE)", len(data))
	}

	bldg := &BuildingObject{
		Header:        header,
		NumParts:      data[0x06],
		NumVariations: data[0x07],
		Colours:       binary.LittleEndian.Uint32(data[0x90:0x94]),
		Flags:         BuildingObjectFlags(data[0x98]),
	}

	// Skip past fixed structure (0xBE bytes)
	offset := 0xBE

	// Parse string table
	for offset < len(data) && data[offset] != 0xFF {
		offset++ // skip language ID
		strStart := offset
		for offset < len(data) && data[offset] != 0 {
			offset++
		}
		if bldg.Name == "" {
			bldg.Name = string(data[strStart:offset])
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF end marker
	}

	// Parse part heights (numParts bytes)
	if offset+int(bldg.NumParts) > len(data) {
		return nil, fmt.Errorf("building part heights truncated")
	}
	bldg.PartHeights = make([]uint8, bldg.NumParts)
	copy(bldg.PartHeights, data[offset:offset+int(bldg.NumParts)])
	offset += int(bldg.NumParts)

	// Skip part animations (numParts * 2 bytes)
	offset += int(bldg.NumParts) * 2

	// Parse variation parts (each variation is a list of uint8 terminated by 0xFF)
	for v := 0; v < int(bldg.NumVariations) && v < 32; v++ {
		var parts []uint8
		for offset < len(data) && data[offset] != 0xFF {
			parts = append(parts, data[offset])
			offset++
		}
		if offset < len(data) {
			offset++ // skip 0xFF terminator
		}
		bldg.VariationParts[v] = parts
	}

	// Skip cargo headers (2 produced + 2 required = 4 * 16 bytes)
	offset += 4 * 16

	// Skip elevator sequences
	numElevators := int(data[0xAD]) // numElevatorSequences at 0xAD
	for i := 0; i < numElevators && offset+2 <= len(data); i++ {
		size := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		offset += 2 + size
	}

	// Load image table into G1
	if g1 != nil && offset < len(data) {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("loading building image table to G1: %w", err)
		}
		bldg.ImageOffset = imgRes.ImageOffset
	}

	return bldg, nil
}

// HasFlags checks if the building object has the given flags.
func (b *BuildingObject) HasFlags(flags BuildingObjectFlags) bool {
	return b.Flags&flags != 0
}

// GetBuildingParts returns the part list for a given variation.
func (b *BuildingObject) GetBuildingParts(variation uint8) []uint8 {
	if int(variation) >= len(b.VariationParts) {
		return nil
	}
	return b.VariationParts[variation]
}
