package worldmap

// TreeElement represents a tree on a tile
type TreeElement TileElement

// Rotation returns the tree rotation (0-3)
func (e *TreeElement) Rotation() uint8 {
	return e.typ & 0x03
}

// SetRotation sets the tree rotation
func (e *TreeElement) SetRotation(rotation uint8) {
	e.typ = (e.typ & 0xFC) | (rotation & 0x03)
}

// ObjectId returns the tree object ID
func (e *TreeElement) ObjectId() uint8 {
	return e.data[0] & 0x3F
}

// SetObjectId sets the tree object ID
func (e *TreeElement) SetObjectId(id uint8) {
	e.data[0] = (e.data[0] & 0xC0) | (id & 0x3F)
}

// Colour returns the tree colour
func (e *TreeElement) Colour() uint8 {
	return ((e.data[0] >> 6) & 0x03) | ((e.data[1] & 0x03) << 2)
}

// SetColour sets the tree colour
func (e *TreeElement) SetColour(colour uint8) {
	e.data[0] = (e.data[0] & 0x3F) | ((colour & 0x03) << 6)
	e.data[1] = (e.data[1] & 0xFC) | ((colour >> 2) & 0x03)
}

// Season returns the tree season (0-3)
func (e *TreeElement) Season() uint8 {
	return (e.data[1] >> 6) & 0x03
}

// SetSeason sets the tree season
func (e *TreeElement) SetSeason(season uint8) {
	e.data[1] = (e.data[1] & 0x3F) | ((season & 0x03) << 6)
}

// Unk2l returns the lower bits of data[2]
func (e *TreeElement) Unk2l() uint8 {
	return e.data[2] & 0x07
}

// Snow returns the snow amount on the tree
func (e *TreeElement) Snow() uint8 {
	return (e.data[2] >> 3) & 0x03
}

// SetSnow sets the snow amount
func (e *TreeElement) SetSnow(snow uint8) {
	e.data[2] = (e.data[2] & 0xE7) | ((snow & 0x03) << 3)
}

// HasShadow returns true if the tree casts a shadow
func (e *TreeElement) HasShadow() bool {
	return e.data[2]&0x40 != 0
}

// SetHasShadow sets whether the tree casts a shadow
func (e *TreeElement) SetHasShadow(state bool) {
	if state {
		e.data[2] |= 0x40
	} else {
		e.data[2] &^= 0x40
	}
}
