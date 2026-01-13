package worldmap

// SurfaceSlope constants for terrain slopes
const (
	SlopeFlatFlag     uint8 = 0x00
	SlopeCornerNorth  uint8 = 1 << 0
	SlopeCornerEast   uint8 = 1 << 1
	SlopeCornerSouth  uint8 = 1 << 2
	SlopeCornerWest   uint8 = 1 << 3
	SlopeCornerAll    uint8 = 0x0F
	SlopeDoubleHeight uint8 = 1 << 4
)

// SurfaceElement represents the ground surface on a tile
// Memory layout: [type|flags|baseZ|clearZ|slope|water|terrain|variation]
type SurfaceElement TileElement

// Slope returns the slope value (0-31)
func (e *SurfaceElement) Slope() uint8 {
	return e.data[0] & 0x1F
}

// SetSlope sets the slope value (0-31)
func (e *SurfaceElement) SetSlope(slope uint8) {
	e.data[0] = (e.data[0] & 0xE0) | (slope & 0x1F)
}

// SlopeCorners returns just the corner flags (0-15)
func (e *SurfaceElement) SlopeCorners() uint8 {
	return e.data[0] & 0x0F
}

// IsSlopeDoubleHeight returns true if the slope has double height
func (e *SurfaceElement) IsSlopeDoubleHeight() bool {
	return e.data[0]&SlopeDoubleHeight != 0
}

// SnowCoverage returns the snow coverage level (0-7)
func (e *SurfaceElement) SnowCoverage() uint8 {
	return e.data[0] >> 5
}

// SetSnowCoverage sets the snow coverage level (0-7)
func (e *SurfaceElement) SetSnowCoverage(coverage uint8) {
	e.data[0] = (e.data[0] & 0x1F) | (coverage << 5)
}

// Water returns the water level in micro Z units
func (e *SurfaceElement) Water() MicroZ {
	return e.data[1] & 0x1F
}

// WaterHeight returns the water height in world units
func (e *SurfaceElement) WaterHeight() int16 {
	return int16(e.data[1]&0x1F) * MicroZStep
}

// SetWater sets the water level in micro Z units
func (e *SurfaceElement) SetWater(level MicroZ) {
	e.data[1] = (e.data[1] & 0xE0) | (level & 0x1F)
}

// Terrain returns the terrain type
func (e *SurfaceElement) Terrain() uint8 {
	return e.data[2] & 0x1F
}

// SetTerrain sets the terrain type
func (e *SurfaceElement) SetTerrain(terrain uint8) {
	e.data[2] = (e.data[2] & 0xE0) | (terrain & 0x1F)
}

// GrowthStage returns the growth stage of vegetation
func (e *SurfaceElement) GrowthStage() uint8 {
	return e.data[2] >> 5
}

// SetGrowthStage sets the growth stage
func (e *SurfaceElement) SetGrowthStage(stage uint8) {
	e.data[2] = (e.data[2] & 0x1F) | (stage << 5)
}

// Variation returns the terrain variation
func (e *SurfaceElement) Variation() uint8 {
	return e.data[3]
}

// SetVariation sets the terrain variation
func (e *SurfaceElement) SetVariation(variation uint8) {
	e.data[3] = variation
}

// IndustryId returns the industry ID if this surface is industrial
func (e *SurfaceElement) IndustryId() IndustryId {
	return IndustryId(e.data[3])
}

// SetIndustry sets the industry ID
func (e *SurfaceElement) SetIndustry(id IndustryId) {
	e.data[3] = uint8(id)
}

// IsIndustrial returns true if this surface is part of an industry
func (e *SurfaceElement) IsIndustrial() bool {
	return e.typ&0x80 != 0
}

// SetIsIndustrial sets the industrial flag
func (e *SurfaceElement) SetIsIndustrial(state bool) {
	if state {
		e.typ |= 0x80
	} else {
		e.typ &^= 0x80
	}
}

// NewSurfaceElement creates a new surface element
func NewSurfaceElement(baseZ, clearZ SmallZ, terrain uint8) *SurfaceElement {
	elem := &SurfaceElement{}
	elem.SetType(ElementTypeSurface)
	elem.SetBaseZ(baseZ)
	elem.SetClearZ(clearZ)
	elem.SetTerrain(terrain)
	return elem
}
