package objects

import (
	"encoding/binary"
	"fmt"
)

// AirportObject — a port of OpenLoco Objects/AirportObject.cpp load(), as far
// as painting needs it (building parts, heights, animations, images). The
// movement nodes and edges are skipped, not parsed (aircraft aren't ported).
// Layout constants come from the compiler (layout_gen.go).
type AirportObject struct {
	Header         *ObjectHeader
	Name           string
	Flags          uint16
	NumSpriteSets  uint8
	NumTiles       uint8
	LargeTiles     uint32 // bit per building variation: 2×2
	PartHeights    []uint8
	PartAnimations []BuildingPartAnimation
	Variations     [][]uint8 // buildingVariationParts, per building type
	Image          uint32    // image table offset
	BuildingImage  uint32    // image + 1 (+ numTiles*4 shadow images when hasShadows)
}

func (o *AirportObject) HasShadows() bool { return o.Flags&AirportFlagHasShadows != 0 }

// ParseAirportObjectWithG1 parses an airport object (g1 may be nil: no images).
func ParseAirportObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*AirportObject, error) {
	if len(data) < AirportObjectSize {
		return nil, fmt.Errorf("airport object too small: %d bytes", len(data))
	}
	o := &AirportObject{
		Header:        header,
		Flags:         binary.LittleEndian.Uint16(data[AirportFlagsOffset:]),
		NumSpriteSets: data[AirportNumSpriteSetsOffset],
		NumTiles:      data[AirportNumTilesOffset],
		LargeTiles:    binary.LittleEndian.Uint32(data[AirportLargeTilesOffset:]),
	}
	numNodes, numEdges := int(data[AirportNumMovementNodesOffset]), int(data[AirportNumMovementEdgesOffset])
	off := AirportObjectSize
	n, name := parseStringTable(data[off:])
	o.Name = name
	off += n
	need := func(k int) error {
		if off+k > len(data) {
			return fmt.Errorf("airport %q truncated at %d (+%d)", name, off, k)
		}
		return nil
	}
	if err := need(int(o.NumSpriteSets)); err != nil {
		return nil, err
	}
	o.PartHeights = append([]uint8(nil), data[off:off+int(o.NumSpriteSets)]...)
	off += int(o.NumSpriteSets)
	if err := need(int(o.NumSpriteSets) * BuildingPartAnimationSize); err != nil {
		return nil, err
	}
	o.PartAnimations = make([]BuildingPartAnimation, o.NumSpriteSets)
	for i := range o.PartAnimations {
		o.PartAnimations[i] = BuildingPartAnimation{NumFrames: data[off+2*i], AnimationSpeed: data[off+2*i+1]}
	}
	off += int(o.NumSpriteSets) * BuildingPartAnimationSize
	for range int(o.NumTiles) {
		start := off
		for off < len(data) && data[off] != 0xFF {
			off++
		}
		if err := need(1); err != nil {
			return nil, err
		}
		o.Variations = append(o.Variations, append([]uint8(nil), data[start:off]...))
		off++
	}
	for off < len(data) && data[off] != 0xFF { // buildingPositions
		off += AirportBuildingSize
	}
	off++
	off += numNodes*AirportMovementNodeSize + numEdges*AirportMovementEdgeSize
	if off > len(data) {
		return nil, fmt.Errorf("airport %q: movement data past the end", name)
	}
	if g1 != nil {
		res, err := g1.LoadImageTable(data[off:])
		if err != nil {
			return nil, fmt.Errorf("airport %q image table: %w", name, err)
		}
		o.Image = res.ImageOffset
	}
	o.BuildingImage = o.Image + 1
	if o.HasShadows() {
		o.BuildingImage += uint32(o.NumTiles) * 4
	}
	return o, nil
}
