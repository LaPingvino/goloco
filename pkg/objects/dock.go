package objects

import (
	"encoding/binary"
	"fmt"
)

// DockObject — a port of OpenLoco Objects/DockObject.cpp load(), for painting.
// Layout constants come from the compiler (layout_gen.go).
type DockObject struct {
	Header         *ObjectHeader
	Name           string
	Flags          uint16
	PartHeights    []uint8
	PartAnimations []BuildingPartAnimation
	Variations     [][]uint8 // numBuildingVariations is 0 or 1
	Image          uint32
	BuildingImage  uint32 // image + 1 (+ variations*4 shadow images when hasShadows)
}

// ParseDockObjectWithG1 parses a dock object (g1 may be nil: no images).
func ParseDockObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*DockObject, error) {
	if len(data) < DockObjectSize {
		return nil, fmt.Errorf("dock object too small: %d bytes", len(data))
	}
	o := &DockObject{Header: header, Flags: binary.LittleEndian.Uint16(data[DockFlagsOffset:])}
	parts, vars := int(data[DockNumBuildingPartsOffset]), int(data[DockNumBuildingVariationsOffset])
	off := DockObjectSize
	n, name := parseStringTable(data[off:])
	o.Name = name
	off += n
	if off+parts*(1+BuildingPartAnimationSize) > len(data) {
		return nil, fmt.Errorf("dock %q truncated", name)
	}
	o.PartHeights = append([]uint8(nil), data[off:off+parts]...)
	off += parts
	o.PartAnimations = make([]BuildingPartAnimation, parts)
	for i := range o.PartAnimations {
		o.PartAnimations[i] = BuildingPartAnimation{NumFrames: data[off+2*i], AnimationSpeed: data[off+2*i+1]}
	}
	off += parts * BuildingPartAnimationSize
	for range vars {
		start := off
		for off < len(data) && data[off] != 0xFF {
			off++
		}
		if off >= len(data) {
			return nil, fmt.Errorf("dock %q: variation parts past the end", name)
		}
		o.Variations = append(o.Variations, append([]uint8(nil), data[start:off]...))
		off++
	}
	if g1 != nil {
		res, err := g1.LoadImageTable(data[off:])
		if err != nil {
			return nil, fmt.Errorf("dock %q image table: %w", name, err)
		}
		o.Image = res.ImageOffset
	}
	o.BuildingImage = o.Image + 1
	if o.Flags&DockFlagHasShadows != 0 {
		o.BuildingImage = o.Image + uint32(vars)*4 + 1
	}
	return o, nil
}
