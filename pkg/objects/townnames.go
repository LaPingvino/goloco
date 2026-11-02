package objects

import "strings"

// TownNamesObject is TownNamesObject.h: six name categories (count, bias,
// offset), whose offsets point into the object's own data (after the 0x1A
// byte header and the name string table) — so Data keeps the decompressed
// object as it is.
type TownNamesObject struct {
	Header     *ObjectHeader
	Name       string
	Categories [6]struct {
		Count, Bias uint8
		Offset      uint16
	}
	Data []byte
}

// ParseTownNamesObject is TownNamesObject::load.
func ParseTownNamesObject(header *ObjectHeader, data []byte) *TownNamesObject {
	o := &TownNamesObject{Header: header, Data: append([]byte(nil), data...)}
	if len(data) < 0x1A {
		return o
	}
	for i := range o.Categories {
		c := &o.Categories[i]
		c.Count, c.Bias = data[2+4*i], data[3+4*i]
		c.Offset = uint16(data[4+4*i]) | uint16(data[5+4*i])<<8
	}
	_, o.Name = walkStringTableLang(data, 0x1A)
	return o
}

// Town name location flags (TownManager's LocationFlags).
const (
	TownNameAdjacentToLargeWater = 1 << 0
	TownNameNotMountainous       = 1 << 1
	TownNameAdjacentToSmallWater = 1 << 2
)

// NameFromRand is townNameFromNamesObject: a name built from up to six
// morphemes chosen by the bits of rand, and the location flags they carry.
func (o *TownNamesObject) NameFromRand(rand uint32) (string, uint8) {
	var b strings.Builder
	var flags uint8
	for _, c := range o.Categories {
		if c.Count == 0 {
			continue
		}
		ax, dx := uint32(uint16(rand)), uint32(uint16(int(c.Count)+int(c.Bias)))
		index := int16(int32(ax*dx>>16) - int32(c.Bias))
		if index >= 0 {
			s, f := o.copyName(int(c.Offset), int(index))
			b.WriteString(s)
			flags |= f
		}
		for shifts := int(c.Count) + int(c.Bias); shifts > 0; shifts >>= 1 {
			rand = rand>>1 | rand<<31 // std::rotr(rand, 1)
		}
	}
	return b.String(), flags
}

// copyName is copyTownNameToBuffer: the index-th string of the category at
// offset (an int16 table of offsets from it), and the flags byte after it.
func (o *TownNamesObject) copyName(off, index int) (string, uint8) {
	p := off + index*2
	if p+2 > len(o.Data) {
		return "", 0
	}
	src := off + int(int16(uint16(o.Data[p])|uint16(o.Data[p+1])<<8))
	if src < 0 || src >= len(o.Data) {
		return "", 0
	}
	end := src
	for end < len(o.Data) && o.Data[end] != 0 {
		end++
	}
	var f uint8
	if end+1 < len(o.Data) {
		f = o.Data[end+1]
	}
	return string(o.Data[src:end]), f
}

// ReorderTownNames sets the scenario's town names object (its one slot).
func (m *ObjectManager) ReorderTownNames(order []string) {
	m.TownNames = nil
	if len(order) > 0 && order[0] != "" {
		if obj := m.GetObject(order[0]); obj != nil {
			m.TownNames, _ = obj.Object.(*TownNamesObject)
		}
	}
}
