package worldmap

// StationElement represents a station on a tile
type StationElement TileElement

// Rotation returns the station rotation (0-3)
func (e *StationElement) Rotation() uint8 {
	return e.typ & 0x03
}

// SetRotation sets the station rotation
func (e *StationElement) SetRotation(rotation uint8) {
	e.typ = (e.typ & 0xFC) | (rotation & 0x03)
}

// ObjectId returns the station object ID
func (e *StationElement) ObjectId() uint8 {
	return e.data[0] & 0x1F
}

// SetObjectId sets the station object ID
func (e *StationElement) SetObjectId(id uint8) {
	e.data[0] = (e.data[0] & 0xE0) | (id & 0x1F)
}

// StationType returns the station type
func (e *StationElement) StationType() uint8 {
	return (e.data[0] >> 5) & 0x07
}

// SetStationType sets the station type
func (e *StationElement) SetStationType(stationType uint8) {
	e.data[0] = (e.data[0] & 0x1F) | ((stationType & 0x07) << 5)
}

// StationId returns the station ID
func (e *StationElement) StationId() uint16 {
	return uint16(e.data[1]) | (uint16(e.data[2]&0x03) << 8)
}

// SetStationId sets the station ID
func (e *StationElement) SetStationId(id uint16) {
	e.data[1] = uint8(id)
	e.data[2] = (e.data[2] & 0xFC) | uint8((id>>8)&0x03)
}

// Owner returns the owning company ID
func (e *StationElement) Owner() CompanyId {
	return CompanyId((e.data[2] >> 4) & 0x0F)
}

// SetOwner sets the owning company ID
func (e *StationElement) SetOwner(owner CompanyId) {
	e.data[2] = (e.data[2] & 0x0F) | ((uint8(owner) & 0x0F) << 4)
}

// MultiTileIndex returns the multi-tile index
func (e *StationElement) MultiTileIndex() uint8 {
	return e.data[3] & 0x03
}

// SetMultiTileIndex sets the multi-tile index
func (e *StationElement) SetMultiTileIndex(index uint8) {
	e.data[3] = (e.data[3] & 0xFC) | (index & 0x03)
}
