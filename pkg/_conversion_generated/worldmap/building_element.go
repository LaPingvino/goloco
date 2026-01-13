package worldmap

// BuildingElement represents a building on a tile
type BuildingElement TileElement

// Rotation returns the building rotation (0-3)
func (e *BuildingElement) Rotation() uint8 {
	return e.typ & 0x03
}

// SetRotation sets the building rotation
func (e *BuildingElement) SetRotation(rotation uint8) {
	e.typ = (e.typ & 0xFC) | (rotation & 0x03)
}

// IsMiscBuilding returns true if this is a miscellaneous building
func (e *BuildingElement) IsMiscBuilding() bool {
	return e.typ&0x40 != 0
}

// SetIsMiscBuilding sets the miscellaneous building flag
func (e *BuildingElement) SetIsMiscBuilding(state bool) {
	if state {
		e.typ |= 0x40
	} else {
		e.typ &^= 0x40
	}
}

// IsConstructed returns true if the building is fully constructed
func (e *BuildingElement) IsConstructed() bool {
	return e.typ&0x80 != 0
}

// SetConstructed sets the constructed flag
func (e *BuildingElement) SetConstructed(state bool) {
	if state {
		e.typ |= 0x80
	} else {
		e.typ &^= 0x80
	}
}

// ObjectId returns the building object ID
func (e *BuildingElement) ObjectId() uint8 {
	return e.data[0]
}

// SetObjectId sets the building object ID
func (e *BuildingElement) SetObjectId(id uint8) {
	e.data[0] = id
}

// SequenceIndex returns the sequence index for multi-tile buildings
func (e *BuildingElement) SequenceIndex() uint8 {
	return e.data[1] & 0x03
}

// SetSequenceIndex sets the sequence index
func (e *BuildingElement) SetSequenceIndex(index uint8) {
	e.data[1] = (e.data[1] & 0xFC) | (index & 0x03)
}

// Age returns the building age
func (e *BuildingElement) Age() uint8 {
	return (e.data[2] | (uint8(e.data[3]) << 8)) & 0x3F
}

// SetAge sets the building age
func (e *BuildingElement) SetAge(age uint8) {
	val := uint16(e.data[2]) | (uint16(e.data[3]) << 8)
	val = (val & 0xFFC0) | uint16(age&0x3F)
	e.data[2] = uint8(val)
	e.data[3] = uint8(val >> 8)
}

// Variation returns the building variation
func (e *BuildingElement) Variation() uint8 {
	val := uint16(e.data[2]) | (uint16(e.data[3]) << 8)
	return uint8((val >> 6) & 0x1F)
}

// SetVariation sets the building variation
func (e *BuildingElement) SetVariation(variation uint8) {
	val := uint16(e.data[2]) | (uint16(e.data[3]) << 8)
	val = (val & 0xF83F) | (uint16(variation&0x1F) << 6)
	e.data[2] = uint8(val)
	e.data[3] = uint8(val >> 8)
}

// Colour returns the building colour
func (e *BuildingElement) Colour() uint8 {
	return uint8((uint16(e.data[2]) | (uint16(e.data[3]) << 8)) >> 11)
}

// SetColour sets the building colour
func (e *BuildingElement) SetColour(colour uint8) {
	val := uint16(e.data[2]) | (uint16(e.data[3]) << 8)
	val = (val & 0x07FF) | (uint16(colour) << 11)
	e.data[2] = uint8(val)
	e.data[3] = uint8(val >> 8)
}
