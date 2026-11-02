package objects

import (
	"encoding/binary"
	"fmt"
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
}

// ParseCompetitorObjectWithG1 is CompetitorObject::load: the two string
// tables (name, availableNamePrefixes), then the image table.
func ParseCompetitorObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*CompetitorObject, error) {
	const fixedSize = 0x38
	if len(data) < fixedSize {
		return nil, fmt.Errorf("competitor object too small: %d bytes", len(data))
	}
	c := &CompetitorObject{Header: header, Emotions: binary.LittleEndian.Uint32(data[0x0C:])}
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
