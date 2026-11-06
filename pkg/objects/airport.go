package objects

import (
	"encoding/binary"
	"fmt"
)

// AirportObject — a port of OpenLoco Objects/AirportObject.cpp load(), as far
// as painting and the aircraft need it (building parts, heights, animations,
// the buildings' positions, the movement nodes and edges planes follow,
// images). Layout constants come from the compiler (layout_gen.go).
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
	Buildings      []AirportBuilding
	MovementNodes  []AirportMovementNode
	MovementEdges  []AirportMovementEdge
}

// AirportBuilding is AirportBuilding: a building of the airport (its type
// index), its rotation and tile offset.
type AirportBuilding struct {
	Index, Rotation uint8
	X, Y            int8
}

// AirportMovementNode is AirportObject::MovementNode: a point a plane moves
// through (relative to the airport, in world units), with its
// AirportMovementNodeFlags.
type AirportMovementNode struct {
	X, Y, Z int16
	Flags   uint16
}

// AirportMovementNodeFlags.
const (
	AirportNodeTerminal         = 1 << 0
	AirportNodeTakeoffEnd       = 1 << 1
	AirportNodeFlag2            = 1 << 2
	AirportNodeTaxiing          = 1 << 3
	AirportNodeInFlight         = 1 << 4
	AirportNodeHeliTakeoffBegin = 1 << 5
	AirportNodeTakeoffBegin     = 1 << 6
	AirportNodeHeliTakeoffEnd   = 1 << 7
	AirportNodeTouchdown        = 1 << 8
)

// AirportMovementEdge is AirportObject::MovementEdge: from curNode to
// nextNode, usable when every edge of MustBeClear and at least one of
// AtLeastOneClear (when any) are clear; Var00 / Var03 as stored.
type AirportMovementEdge struct {
	Var00, CurNode, NextNode, Var03 uint8
	MustBeClear                     uint32
	AtLeastOneClear                 uint32
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
		if err := need(AirportBuildingSize); err != nil {
			return nil, err
		}
		o.Buildings = append(o.Buildings, AirportBuilding{Index: data[off], Rotation: data[off+1], X: int8(data[off+2]), Y: int8(data[off+3])})
		off += AirportBuildingSize
	}
	off++
	if err := need(numNodes*AirportMovementNodeSize + numEdges*AirportMovementEdgeSize); err != nil {
		return nil, fmt.Errorf("airport %q: movement data past the end", name)
	}
	le16 := func(o int) int16 { return int16(binary.LittleEndian.Uint16(data[o:])) }
	for range numNodes {
		o.MovementNodes = append(o.MovementNodes, AirportMovementNode{X: le16(off), Y: le16(off + 2), Z: le16(off + 4), Flags: uint16(le16(off + 6))})
		off += AirportMovementNodeSize
	}
	for range numEdges {
		o.MovementEdges = append(o.MovementEdges, AirportMovementEdge{Var00: data[off], CurNode: data[off+1], NextNode: data[off+2], Var03: data[off+3],
			MustBeClear: binary.LittleEndian.Uint32(data[off+4:]), AtLeastOneClear: binary.LittleEndian.Uint32(data[off+8:])})
		off += AirportMovementEdgeSize
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
