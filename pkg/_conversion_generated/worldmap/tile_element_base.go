package worldmap

// ElementType represents the type of a tile element
type ElementType uint8

const (
	ElementTypeSurface  ElementType = 0x00
	ElementTypeTrack    ElementType = 0x04
	ElementTypeStation  ElementType = 0x08
	ElementTypeSignal   ElementType = 0x0C
	ElementTypeBuilding ElementType = 0x10
	ElementTypeTree     ElementType = 0x14
	ElementTypeWall     ElementType = 0x18
	ElementTypeRoad     ElementType = 0x1C
	ElementTypeIndustry ElementType = 0x20
)

// ElementFlags are flags for tile elements
const (
	ElementFlagGhost       uint8 = 1 << 4
	ElementFlagAiAllocated uint8 = 1 << 5
	ElementFlagFlag6       uint8 = 1 << 6
	ElementFlagLast        uint8 = 1 << 7
)

// TileElementBase is the base structure for all tile elements (8 bytes)
type TileElementBase struct {
	typ    uint8 // bits 0-1: rotation/data, bits 2-5: element type, bits 6-7: flags
	flags  uint8 // bits 0-3: occupied quarter, bits 4-7: element flags
	baseZ  uint8 // base height / 4
	clearZ uint8 // clear height / 4
}

// Type returns the element type
func (e *TileElementBase) Type() ElementType {
	return ElementType((e.typ >> 2) & 0x0F << 2)
}

// SetType sets the element type
func (e *TileElementBase) SetType(t ElementType) {
	e.typ = uint8(t)
}

// Flags returns the element flags
func (e *TileElementBase) Flags() uint8 {
	return e.flags
}

// BaseZ returns the base height in small Z units
func (e *TileElementBase) BaseZ() SmallZ {
	return e.baseZ
}

// BaseHeight returns the base height in world units
func (e *TileElementBase) BaseHeight() int16 {
	return int16(e.baseZ) * SmallZStep
}

// ClearZ returns the clear height in small Z units
func (e *TileElementBase) ClearZ() SmallZ {
	return e.clearZ
}

// ClearHeight returns the clear height in world units
func (e *TileElementBase) ClearHeight() int16 {
	return int16(e.clearZ) * SmallZStep
}

// SetBaseZ sets the base height
func (e *TileElementBase) SetBaseZ(z uint8) {
	e.baseZ = z
}

// SetClearZ sets the clear height
func (e *TileElementBase) SetClearZ(z uint8) {
	e.clearZ = z
}

// OccupiedQuarter returns which quarter of the tile is occupied
func (e *TileElementBase) OccupiedQuarter() uint8 {
	return e.flags & 0x0F
}

// SetOccupiedQuarter sets which quarter of the tile is occupied
func (e *TileElementBase) SetOccupiedQuarter(val uint8) {
	e.flags = (e.flags & 0xF0) | (val & 0x0F)
}

// IsGhost returns true if this is a ghost element
func (e *TileElementBase) IsGhost() bool {
	return e.flags&ElementFlagGhost != 0
}

// SetGhost sets the ghost flag
func (e *TileElementBase) SetGhost(state bool) {
	if state {
		e.flags |= ElementFlagGhost
	} else {
		e.flags &^= ElementFlagGhost
	}
}

// IsAiAllocated returns true if AI has allocated this element
func (e *TileElementBase) IsAiAllocated() bool {
	return e.flags&ElementFlagAiAllocated != 0
}

// SetAiAllocated sets the AI allocated flag
func (e *TileElementBase) SetAiAllocated(state bool) {
	if state {
		e.flags |= ElementFlagAiAllocated
	} else {
		e.flags &^= ElementFlagAiAllocated
	}
}

// IsLast returns true if this is the last element on the tile
func (e *TileElementBase) IsLast() bool {
	return e.flags&ElementFlagLast != 0
}

// SetLastFlag sets the last element flag
func (e *TileElementBase) SetLastFlag(state bool) {
	if state {
		e.flags |= ElementFlagLast
	} else {
		e.flags &^= ElementFlagLast
	}
}

// Rotation returns the rotation value (0-3) from the type byte
func (e *TileElementBase) Rotation() uint8 {
	return e.typ & 0x03
}

// SetRotation sets the rotation value (0-3)
func (e *TileElementBase) SetRotation(rotation uint8) {
	e.typ = (e.typ & 0xFC) | (rotation & 0x03)
}
