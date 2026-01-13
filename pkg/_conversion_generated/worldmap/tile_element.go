package worldmap

// TileElement is a generic tile element (8 bytes total)
// The first 4 bytes are the base, the remaining 4 are element-specific
type TileElement struct {
	TileElementBase
	data [4]uint8 // Element-specific data
}

// Data returns the element-specific data bytes
func (e *TileElement) Data() []uint8 {
	return e.data[:]
}

// AsSurface returns this element as a SurfaceElement, or nil if not a surface
func (e *TileElement) AsSurface() *SurfaceElement {
	if e.Type() == ElementTypeSurface {
		return (*SurfaceElement)(e)
	}
	return nil
}

// AsTrack returns this element as a TrackElement, or nil if not a track
func (e *TileElement) AsTrack() *TrackElement {
	if e.Type() == ElementTypeTrack {
		return (*TrackElement)(e)
	}
	return nil
}

// AsBuilding returns this element as a BuildingElement, or nil if not a building
func (e *TileElement) AsBuilding() *BuildingElement {
	if e.Type() == ElementTypeBuilding {
		return (*BuildingElement)(e)
	}
	return nil
}

// AsTree returns this element as a TreeElement, or nil if not a tree
func (e *TileElement) AsTree() *TreeElement {
	if e.Type() == ElementTypeTree {
		return (*TreeElement)(e)
	}
	return nil
}

// AsWall returns this element as a WallElement, or nil if not a wall
func (e *TileElement) AsWall() *WallElement {
	if e.Type() == ElementTypeWall {
		return (*WallElement)(e)
	}
	return nil
}

// AsRoad returns this element as a RoadElement, or nil if not a road
func (e *TileElement) AsRoad() *RoadElement {
	if e.Type() == ElementTypeRoad {
		return (*RoadElement)(e)
	}
	return nil
}

// AsStation returns this element as a StationElement, or nil if not a station
func (e *TileElement) AsStation() *StationElement {
	if e.Type() == ElementTypeStation {
		return (*StationElement)(e)
	}
	return nil
}

// AsSignal returns this element as a SignalElement, or nil if not a signal
func (e *TileElement) AsSignal() *SignalElement {
	if e.Type() == ElementTypeSignal {
		return (*SignalElement)(e)
	}
	return nil
}

// AsIndustry returns this element as an IndustryElement, or nil if not an industry
func (e *TileElement) AsIndustry() *IndustryElement {
	if e.Type() == ElementTypeIndustry {
		return (*IndustryElement)(e)
	}
	return nil
}
