package objects

import "strings"

// RegionObject is RegionObject.h (0x12 bytes, the name string table, the
// cargo influence objects' headers, the headers of objects it loads up to
// 0xFF, then the image table): what towns read of it.
type RegionObject struct {
	Header *ObjectHeader
	Name   string
	// CargoInfluenceNames are the cargo objects (DAT names) whose slots are
	// cargoInfluenceObjectIds after load; CargoInfluenceTownFilter is each
	// one's town filter (0 all towns, 1 may snow, 2 in desert).
	CargoInfluenceNames      []string
	CargoInfluenceTownFilter [4]uint8
}

// ParseRegionObject reads the fields towns use (no images).
func ParseRegionObject(header *ObjectHeader, data []byte) *RegionObject {
	r := &RegionObject{Header: header}
	if len(data) < 0x12 {
		return r
	}
	n := int(data[0x08])
	copy(r.CargoInfluenceTownFilter[:], data[0x0D:0x11])
	off, name := walkStringTableLang(data, 0x12)
	r.Name = name
	for i := 0; i < n && off+HeaderSize <= len(data); i++ {
		r.CargoInfluenceNames = append(r.CargoInfluenceNames, strings.TrimRight(string(data[off+4:off+12]), " \x00"))
		off += HeaderSize
	}
	return r
}

// CargoInfluenceObjectIDs resolves the cargo influence objects to the
// scenario's cargo slots (RegionObject::load's findObjectHandle); 0 when a
// header is not loaded (upstream leaves the id as it was).
func (m *ObjectManager) CargoInfluenceObjectIDs(r *RegionObject) []uint8 {
	ids := make([]uint8, len(r.CargoInfluenceNames))
	for i, name := range r.CargoInfluenceNames {
		for slot, c := range m.CargoObjects {
			if c != nil && c.Header != nil && strings.EqualFold(strings.TrimRight(string(c.Header.Name[:]), " \x00"), name) {
				ids[i] = uint8(slot)
				break
			}
		}
	}
	return ids
}

// ReorderRegion sets the scenario's region object (its one slot).
func (m *ObjectManager) ReorderRegion(order []string) {
	m.Region = nil
	if len(order) > 0 && order[0] != "" {
		if obj := m.GetObject(order[0]); obj != nil {
			m.Region, _ = obj.Object.(*RegionObject)
		}
	}
}
