package worldmap

// TrackElement represents a rail track on a tile
type TrackElement TileElement

// Rotation returns the track rotation (0-3)
func (e *TrackElement) Rotation() uint8 {
	return e.typ & 0x03
}

// SetRotation sets the track rotation
func (e *TrackElement) SetRotation(rotation uint8) {
	e.typ = (e.typ & 0xFC) | (rotation & 0x03)
}

// TrackId returns the track piece ID
func (e *TrackElement) TrackId() uint8 {
	return e.data[0] & 0x3F
}

// SetTrackId sets the track piece ID
func (e *TrackElement) SetTrackId(id uint8) {
	e.data[0] = (e.data[0] & 0xC0) | (id & 0x3F)
}

// HasSignal returns true if this track has a signal
func (e *TrackElement) HasSignal() bool {
	return e.data[0]&0x40 != 0
}

// SetHasSignal sets whether the track has a signal
func (e *TrackElement) SetHasSignal(state bool) {
	if state {
		e.data[0] |= 0x40
	} else {
		e.data[0] &^= 0x40
	}
}

// HasStationElement returns true if this track has a station
func (e *TrackElement) HasStationElement() bool {
	return e.data[0]&0x80 != 0
}

// SetHasStationElement sets whether the track has a station
func (e *TrackElement) SetHasStationElement(state bool) {
	if state {
		e.data[0] |= 0x80
	} else {
		e.data[0] &^= 0x80
	}
}

// SequenceIndex returns the sequence index for multi-tile tracks
func (e *TrackElement) SequenceIndex() uint8 {
	return e.data[1] & 0x0F
}

// SetSequenceIndex sets the sequence index
func (e *TrackElement) SetSequenceIndex(index uint8) {
	e.data[1] = (e.data[1] & 0xF0) | (index & 0x0F)
}

// TrackObjectId returns the track object type ID
func (e *TrackElement) TrackObjectId() uint8 {
	return (e.data[1] >> 4) & 0x0F
}

// SetTrackObjectId sets the track object type ID
func (e *TrackElement) SetTrackObjectId(id uint8) {
	e.data[1] = (e.data[1] & 0x0F) | ((id & 0x0F) << 4)
}

// Owner returns the owning company ID
func (e *TrackElement) Owner() CompanyId {
	return CompanyId(e.data[2] & 0x0F)
}

// SetOwner sets the owning company ID
func (e *TrackElement) SetOwner(owner CompanyId) {
	e.data[2] = (e.data[2] & 0xF0) | (uint8(owner) & 0x0F)
}

// HasBridge returns true if this track has a bridge
func (e *TrackElement) HasBridge() bool {
	return e.data[3]&0x80 != 0
}

// BridgeObjectId returns the bridge object ID
func (e *TrackElement) BridgeObjectId() uint8 {
	return e.data[3] & 0x0F
}

// SetBridgeObjectId sets the bridge object ID
func (e *TrackElement) SetBridgeObjectId(id uint8) {
	e.data[3] = (e.data[3] & 0xF0) | (id & 0x0F)
	e.data[3] |= 0x80 // Set has bridge flag
}
