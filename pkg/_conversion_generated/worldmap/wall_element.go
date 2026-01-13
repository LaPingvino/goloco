package worldmap

// WallElement represents a wall on a tile edge
type WallElement TileElement

// Rotation returns the wall rotation (which edge: 0-3)
func (e *WallElement) Rotation() uint8 {
	return e.typ & 0x03
}

// SetRotation sets the wall rotation
func (e *WallElement) SetRotation(rotation uint8) {
	e.typ = (e.typ & 0xFC) | (rotation & 0x03)
}

// ObjectId returns the wall object ID
func (e *WallElement) ObjectId() uint8 {
	return e.data[0]
}

// SetObjectId sets the wall object ID
func (e *WallElement) SetObjectId(id uint8) {
	e.data[0] = id
}

// PrimaryColour returns the primary colour
func (e *WallElement) PrimaryColour() uint8 {
	return e.data[1] & 0x1F
}

// SetPrimaryColour sets the primary colour
func (e *WallElement) SetPrimaryColour(colour uint8) {
	e.data[1] = (e.data[1] & 0xE0) | (colour & 0x1F)
}

// SecondaryColour returns the secondary colour
func (e *WallElement) SecondaryColour() uint8 {
	return e.data[2] & 0x1F
}

// SetSecondaryColour sets the secondary colour
func (e *WallElement) SetSecondaryColour(colour uint8) {
	e.data[2] = (e.data[2] & 0xE0) | (colour & 0x1F)
}
