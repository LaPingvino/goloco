package worldmap

// SignalElement represents a rail signal on a tile
type SignalElement TileElement

// Rotation returns the signal rotation (0-3)
func (e *SignalElement) Rotation() uint8 {
	return e.typ & 0x03
}

// SetRotation sets the signal rotation
func (e *SignalElement) SetRotation(rotation uint8) {
	e.typ = (e.typ & 0xFC) | (rotation & 0x03)
}

// LeftSignalObjectId returns the left signal object ID
func (e *SignalElement) LeftSignalObjectId() uint8 {
	return e.data[0] & 0x0F
}

// SetLeftSignalObjectId sets the left signal object ID
func (e *SignalElement) SetLeftSignalObjectId(id uint8) {
	e.data[0] = (e.data[0] & 0xF0) | (id & 0x0F)
}

// RightSignalObjectId returns the right signal object ID
func (e *SignalElement) RightSignalObjectId() uint8 {
	return (e.data[0] >> 4) & 0x0F
}

// SetRightSignalObjectId sets the right signal object ID
func (e *SignalElement) SetRightSignalObjectId(id uint8) {
	e.data[0] = (e.data[0] & 0x0F) | ((id & 0x0F) << 4)
}

// HasLeftSignal returns true if there is a left signal
func (e *SignalElement) HasLeftSignal() bool {
	return e.data[1]&0x80 != 0
}

// SetHasLeftSignal sets whether there is a left signal
func (e *SignalElement) SetHasLeftSignal(state bool) {
	if state {
		e.data[1] |= 0x80
	} else {
		e.data[1] &^= 0x80
	}
}

// HasRightSignal returns true if there is a right signal
func (e *SignalElement) HasRightSignal() bool {
	return e.data[1]&0x40 != 0
}

// SetHasRightSignal sets whether there is a right signal
func (e *SignalElement) SetHasRightSignal(state bool) {
	if state {
		e.data[1] |= 0x40
	} else {
		e.data[1] &^= 0x40
	}
}

// LeftSignalState returns the left signal state
func (e *SignalElement) LeftSignalState() uint8 {
	return e.data[2] & 0x03
}

// SetLeftSignalState sets the left signal state
func (e *SignalElement) SetLeftSignalState(state uint8) {
	e.data[2] = (e.data[2] & 0xFC) | (state & 0x03)
}

// RightSignalState returns the right signal state
func (e *SignalElement) RightSignalState() uint8 {
	return (e.data[2] >> 2) & 0x03
}

// SetRightSignalState sets the right signal state
func (e *SignalElement) SetRightSignalState(state uint8) {
	e.data[2] = (e.data[2] & 0xF3) | ((state & 0x03) << 2)
}
