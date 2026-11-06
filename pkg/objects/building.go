package objects

import (
	"encoding/binary"
	"fmt"
	"strings"
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

	PartAnimations       []BuildingPartAnimation // per part (BuildingCommon.h)
	NumElevatorSequences uint8                   // 0xAD
	ProducedQuantity     [2]uint8                // 0xA0 (passengers: the building\'s population)
	// ProducedCargoNames are the two produced cargoes' object names (after
	// the variation parts; "" for none — a 0xFF first byte); ProducedCargoIDs
	// resolves them to the scenario's cargo slots (load's findObjectHandle).
	ProducedCargoNames      [2]string
	TownAmenityCategory     uint8     // 0xAC (0xFF none)
	ClearCostIndex          uint8     // 0x99
	ClearCostFactor         int16     // 0x9A
	DemolishRatingReduction int16     // 0xAA
	ElevatorSequences       [][]uint8 // heights; size is a power of 2
	ScaffoldingSegmentType  uint8     // 0x9C (0xFF = none)
	ScaffoldingColour       uint8     // 0x9D

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

	// Part animations (numParts × {numFrames, animationSpeed})
	bldg.PartAnimations = make([]BuildingPartAnimation, bldg.NumParts)
	for i := range bldg.PartAnimations {
		if offset+2*i+1 < len(data) {
			bldg.PartAnimations[i] = BuildingPartAnimation{NumFrames: data[offset+2*i], AnimationSpeed: data[offset+2*i+1]}
		}
	}
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

	// Cargo headers: 2 produced + 2 required, 16 bytes each (also for none:
	// a 0xFF first byte — load always skips sizeof(ObjectHeader)).
	for i := 0; i < 2 && offset+HeaderSize <= len(data); i++ {
		if data[offset+i*HeaderSize] != 0xFF {
			bldg.ProducedCargoNames[i] = strings.TrimRight(string(data[offset+i*HeaderSize+4:offset+i*HeaderSize+12]), " \x00")
		}
	}
	offset += 4 * 16

	// Elevator sequences: uint16 size, then size height bytes each
	bldg.NumElevatorSequences = data[0xAD]
	copy(bldg.ProducedQuantity[:], data[0xA0:0xA2])
	bldg.TownAmenityCategory = data[0xAC]
	bldg.ClearCostIndex = data[0x99]
	bldg.ClearCostFactor = int16(uint16(data[0x9A]) | uint16(data[0x9B])<<8)
	bldg.DemolishRatingReduction = int16(uint16(data[0xAA]) | uint16(data[0xAB])<<8)
	bldg.ScaffoldingSegmentType = data[0x9C]
	bldg.ScaffoldingColour = data[0x9D]
	for i := 0; i < int(bldg.NumElevatorSequences) && offset+2 <= len(data); i++ {
		size := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		if offset+2+size <= len(data) {
			bldg.ElevatorSequences = append(bldg.ElevatorSequences, append([]uint8(nil), data[offset+2:offset+2+size]...))
		}
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

// BuildingPartAnimation is OpenLoco BuildingCommon.h BuildingPartAnimation.
type BuildingPartAnimation struct {
	NumFrames      uint8 // power of 2 (0 = no part animation)
	AnimationSpeed uint8 // bit 7: position-modified
}

// ProducedCargoIDs is the building's producedCargoType: its produced cargoes'
// scenario cargo slots, 0xFF for none (or not loaded, as upstream leaves it).
func (m *ObjectManager) ProducedCargoIDs(b *BuildingObject) [2]uint8 {
	ids := [2]uint8{0xFF, 0xFF}
	for i, name := range b.ProducedCargoNames {
		if name == "" {
			continue
		}
		for slot, c := range m.CargoObjects {
			if c != nil && c.Header != nil && strings.EqualFold(strings.TrimRight(string(c.Header.Name[:]), " \x00"), name) {
				ids[i] = uint8(slot)
				break
			}
		}
	}
	return ids
}
