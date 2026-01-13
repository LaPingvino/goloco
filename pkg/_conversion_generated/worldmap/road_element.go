package worldmap

// RoadElement represents a road on a tile
type RoadElement TileElement

// Rotation returns the road rotation (0-3)
func (e *RoadElement) Rotation() uint8 {
	return e.typ & 0x03
}

// SetRotation sets the road rotation
func (e *RoadElement) SetRotation(rotation uint8) {
	e.typ = (e.typ & 0xFC) | (rotation & 0x03)
}

// RoadId returns the road piece ID
func (e *RoadElement) RoadId() uint8 {
	return e.data[0] & 0x0F
}

// SetRoadId sets the road piece ID
func (e *RoadElement) SetRoadId(id uint8) {
	e.data[0] = (e.data[0] & 0xF0) | (id & 0x0F)
}

// HasStationElement returns true if this road has a station
func (e *RoadElement) HasStationElement() bool {
	return e.data[0]&0x80 != 0
}

// SetHasStationElement sets whether the road has a station
func (e *RoadElement) SetHasStationElement(state bool) {
	if state {
		e.data[0] |= 0x80
	} else {
		e.data[0] &^= 0x80
	}
}

// SequenceIndex returns the sequence index for multi-tile roads
func (e *RoadElement) SequenceIndex() uint8 {
	return e.data[1] & 0x03
}

// SetSequenceIndex sets the sequence index
func (e *RoadElement) SetSequenceIndex(index uint8) {
	e.data[1] = (e.data[1] & 0xFC) | (index & 0x03)
}

// RoadObjectId returns the road object type ID
func (e *RoadElement) RoadObjectId() uint8 {
	return (e.data[1] >> 4) & 0x0F
}

// SetRoadObjectId sets the road object type ID
func (e *RoadElement) SetRoadObjectId(id uint8) {
	e.data[1] = (e.data[1] & 0x0F) | ((id & 0x0F) << 4)
}

// Owner returns the owning company ID
func (e *RoadElement) Owner() CompanyId {
	return CompanyId(e.data[2] & 0x0F)
}

// SetOwner sets the owning company ID
func (e *RoadElement) SetOwner(owner CompanyId) {
	e.data[2] = (e.data[2] & 0xF0) | (uint8(owner) & 0x0F)
}

// HasLevelCrossing returns true if this road has a level crossing
func (e *RoadElement) HasLevelCrossing() bool {
	return e.data[2]&0x10 != 0
}

// SetHasLevelCrossing sets whether the road has a level crossing
func (e *RoadElement) SetHasLevelCrossing(state bool) {
	if state {
		e.data[2] |= 0x10
	} else {
		e.data[2] &^= 0x10
	}
}

// LevelCrossingObjectId returns the level crossing object ID
func (e *RoadElement) LevelCrossingObjectId() uint8 {
	return (e.data[2] >> 5) & 0x07
}

// SetLevelCrossingObjectId sets the level crossing object ID
func (e *RoadElement) SetLevelCrossingObjectId(id uint8) {
	e.data[2] = (e.data[2] & 0x1F) | ((id & 0x07) << 5)
}

// HasBridge returns true if this road has a bridge
func (e *RoadElement) HasBridge() bool {
	return e.data[3]&0x80 != 0
}

// BridgeObjectId returns the bridge object ID
func (e *RoadElement) BridgeObjectId() uint8 {
	return e.data[3] & 0x0F
}

// SetBridgeObjectId sets the bridge object ID
func (e *RoadElement) SetBridgeObjectId(id uint8) {
	e.data[3] = (e.data[3] & 0xF0) | (id & 0x0F)
	e.data[3] |= 0x80 // Set has bridge flag
}

// Mods returns the road modifications bitmask
func (e *RoadElement) Mods() uint8 {
	return (e.data[3] >> 4) & 0x07
}

// SetMods sets the road modifications bitmask
func (e *RoadElement) SetMods(mods uint8) {
	e.data[3] = (e.data[3] & 0x8F) | ((mods & 0x07) << 4)
}
