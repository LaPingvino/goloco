package objects

import (
	"encoding/binary"
	"fmt"
)

// HillShapesObject is HillShapesObject.h (0x0E bytes, then the name string
// table and the image table): the height-map images the original terrain
// generator stamps onto the map.
type HillShapesObject struct {
	Header        *ObjectHeader
	Name          string
	HillCount     uint8  // 0x02 hillHeightMapCount
	MountainCount uint8  // 0x03 mountainHeightMapCount
	Flags         uint16 // 0x0C HillShapeFlags (bit 0 isHeightMap)
	Image         uint32 // the image table's base (G1 pool index)
}

// HillShapeIsHeightMap is HillShapeFlags::isHeightMap: one image is the
// whole map's height map.
const HillShapeIsHeightMap = 1 << 0

// ParseHillShapesObjectWithG1 is HillShapesObject::load.
func ParseHillShapesObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*HillShapesObject, error) {
	const fixedSize = 0x0E
	if len(data) < fixedSize {
		return nil, fmt.Errorf("hill shapes object too short (%d bytes)", len(data))
	}
	h := &HillShapesObject{Header: header, HillCount: data[0x02], MountainCount: data[0x03], Flags: binary.LittleEndian.Uint16(data[0x0C:])}
	offset, name := walkStringTableLang(data, fixedSize)
	h.Name = name
	if g1 != nil && offset < len(data) {
		res, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("hill shapes images: %w", err)
		}
		h.Image = res.ImageOffset
	}
	return h, nil
}

// ReorderHillShapes sets the scenario's hill shapes object (its one slot).
func (m *ObjectManager) ReorderHillShapes(order []string) {
	m.HillShapes = nil
	if len(order) > 0 && order[0] != "" {
		if obj := m.GetObject(order[0]); obj != nil {
			m.HillShapes, _ = obj.Object.(*HillShapesObject)
		}
	}
}
