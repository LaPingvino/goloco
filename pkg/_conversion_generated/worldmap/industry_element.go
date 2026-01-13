package worldmap

// IndustryElement represents an industry building on a tile
type IndustryElement TileElement

// Rotation returns the industry rotation (0-3)
func (e *IndustryElement) Rotation() uint8 {
	return e.typ & 0x03
}

// SetRotation sets the industry rotation
func (e *IndustryElement) SetRotation(rotation uint8) {
	e.typ = (e.typ & 0xFC) | (rotation & 0x03)
}

// IndustryId returns the industry ID
func (e *IndustryElement) IndustryId() IndustryId {
	return IndustryId(e.data[0])
}

// SetIndustryId sets the industry ID
func (e *IndustryElement) SetIndustryId(id IndustryId) {
	e.data[0] = uint8(id)
}

// SequenceIndex returns the sequence index for multi-tile industries
func (e *IndustryElement) SequenceIndex() uint8 {
	return e.data[1] & 0x03
}

// SetSequenceIndex sets the sequence index
func (e *IndustryElement) SetSequenceIndex(index uint8) {
	e.data[1] = (e.data[1] & 0xFC) | (index & 0x03)
}

// BuildingType returns the building type within the industry
func (e *IndustryElement) BuildingType() uint8 {
	return (e.data[1] >> 2) & 0x07
}

// SetBuildingType sets the building type
func (e *IndustryElement) SetBuildingType(buildingType uint8) {
	e.data[1] = (e.data[1] & 0xE3) | ((buildingType & 0x07) << 2)
}

// IsConstructed returns true if the industry building is fully constructed
func (e *IndustryElement) IsConstructed() bool {
	return e.typ&0x80 != 0
}

// SetConstructed sets the constructed flag
func (e *IndustryElement) SetConstructed(state bool) {
	if state {
		e.typ |= 0x80
	} else {
		e.typ &^= 0x80
	}
}

// Variation returns the building variation
func (e *IndustryElement) Variation() uint8 {
	return e.data[2]
}

// SetVariation sets the building variation
func (e *IndustryElement) SetVariation(variation uint8) {
	e.data[2] = variation
}
