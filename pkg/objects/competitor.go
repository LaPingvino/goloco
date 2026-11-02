package objects

import (
	"encoding/binary"
	"fmt"
	"sort"
)

// CompetitorObject is CompetitorObject.h (0x38 bytes): the company owner's
// name and their face images, one per Emotion the object has.
type CompetitorObject struct {
	Header *ObjectHeader
	Name   string
	// Emotions (0x0C) is the set of Emotion bits the object has faces for;
	// Images are images[9] after load: the image table's base, plus 2 for each
	// present emotion before it (absent ones keep the base).
	Emotions uint32
	Images   [9]uint32
	// Intelligence, Aggressiveness, Competitiveness (0x34–0x36): 1–9, the
	// AI's character (ObjectIndex display data).
	Intelligence, Aggressiveness, Competitiveness uint8
}

// ParseCompetitorObjectWithG1 is CompetitorObject::load: the two string
// tables (name, availableNamePrefixes), then the image table.
func ParseCompetitorObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*CompetitorObject, error) {
	const fixedSize = 0x38
	if len(data) < fixedSize {
		return nil, fmt.Errorf("competitor object too small: %d bytes", len(data))
	}
	c := &CompetitorObject{Header: header, Emotions: binary.LittleEndian.Uint32(data[0x0C:]),
		Intelligence: data[0x34], Aggressiveness: data[0x35], Competitiveness: data[0x36]}
	offset := fixedSize
	offset, c.Name = walkStringTableLang(data, offset)
	offset, _ = walkStringTableLang(data, offset)
	if g1 != nil && offset < len(data) {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("competitor image table: %w", err)
		}
		c.setImages(imgRes.ImageOffset)
	}
	return c, nil
}

// setImages fills images[] from the table's base as load does.
func (c *CompetitorObject) setImages(base uint32) {
	off := uint32(0)
	for i := range c.Images {
		c.Images[i] = base
		if c.Emotions&(1<<i) != 0 {
			c.Images[i] += off
			off += 2
		}
	}
}

// InstalledCompetitors is ObjectIndex::getAvailableObjects(competitor): every
// installed competitor object, in the index's order (sorted by display name,
// ObjectIndex.cpp createIndex).
func (m *ObjectManager) InstalledCompetitors() []*CompetitorObject {
	var out []*CompetitorObject
	for _, lo := range m.Objects {
		if c, ok := lo.Object.(*CompetitorObject); ok {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
