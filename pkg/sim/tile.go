// Package sim holds goloco's simulation state — the map and what lives on it —
// with no rendering dependencies, so it can be tested headless. Drawing lives in
// pkg/world, which embeds this state.
//
// OpenLoco keeps the same split: Map/ and Vehicles/ hold state and rules,
// Paint/ and Ui/ draw them.
package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Tile is the state of one map tile, copied from the scenario's tile elements
// (OpenLoco Map/TileElement.h: SurfaceElement plus the elements stacked on it).
type Tile struct {
	TerrainIndex uint8 // raw LandObject slot index from the scenario
	BaseZ        uint8 // height in SmallZ units (OpenLoco: 1 unit = 4px)
	Slope        uint8 // slope corners and flags from byte 4 of surface element
	WaterLevel   uint8 // water level from surface element byte 5 bits [4:0]; 0 = no water
	UpdateTimer  uint8 // byte 5 bits [7:5]: an industrial field's update timer (getUpdateTimer)
	GrowthStage  uint8 // terrain growth/season stage 0-7 (byte 6 bits [7:5]); selects sprite set
	Variation    uint8 // per-tile flat-grass variation (byte 7); 0 = plain flat image
	IsIndustrial bool  // surface owned by an industry; TerrainIndex is not a land slot
	Type6Flag    bool  // SurfaceElement::hasType6Flag: water a ship route needs (adjustSurfaceHeight keeps it)
	WaveFlag     bool  // a wave is running on this tile (OpenLoco SurfaceElement flag 6; see wave.go)
	Trees        []scenario.TreeElement
	Buildings    []scenario.BuildingElement
	Walls        []scenario.WallElement
	Tracks       []scenario.TrackElement
	Roads        []scenario.RoadElement
	Stations     []scenario.StationElement
	Signals      []scenario.SignalElement
	Industries   []scenario.IndustryElement
	Order        []scenario.ElemRef // elements in stored (height) order; kept in step by insertRef/removeRef
}

// baseZOf is the height of the element ref points at (the surface: BaseZ).
func (t *Tile) baseZOf(r scenario.ElemRef) uint8 {
	switch r.Kind {
	case scenario.KindTrack:
		return t.Tracks[r.Index].BaseZ
	case scenario.KindStation:
		return t.Stations[r.Index].BaseZ
	case scenario.KindSignal:
		return t.Signals[r.Index].BaseZ
	case scenario.KindBuilding:
		return t.Buildings[r.Index].BaseZ
	case scenario.KindTree:
		return t.Trees[r.Index].BaseZ
	case scenario.KindWall:
		return t.Walls[r.Index].BaseZ
	case scenario.KindRoad:
		return t.Roads[r.Index].BaseZ
	case scenario.KindIndustry:
		return t.Industries[r.Index].BaseZ
	}
	return t.BaseZ
}

// insertRef records a newly appended element (index = its slice position) where
// TileManager::insertElement puts it: after every element whose baseZ is <= its own.
func (t *Tile) insertRef(kind uint8, index int, baseZ uint8) {
	at := len(t.Order)
	for i, r := range t.Order {
		if baseZ < t.baseZOf(r) {
			at = i
			break
		}
	}
	t.Order = append(t.Order, scenario.ElemRef{})
	copy(t.Order[at+1:], t.Order[at:])
	t.Order[at] = scenario.ElemRef{Kind: kind, Index: uint16(index)}
}

// removeRef drops the ref of an element removed from its slice at index and
// renumbers the later elements of that kind.
func (t *Tile) removeRef(kind uint8, index int) {
	out := t.Order[:0]
	for _, r := range t.Order {
		if r.Kind == kind {
			if int(r.Index) == index {
				continue
			}
			if int(r.Index) > index {
				r.Index--
			}
		}
		out = append(out, r)
	}
	t.Order = out
}

// InsertRef and RemoveRef are insertRef/removeRef for packages that edit tiles
// (pkg/world construction and undo).
func (t *Tile) InsertRef(kind uint8, index int, baseZ uint8) { t.insertRef(kind, index, baseZ) }
func (t *Tile) RemoveRef(kind uint8, index int)              { t.removeRef(kind, index) }
