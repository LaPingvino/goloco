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
	GrowthStage  uint8 // terrain growth/season stage 0-7 (byte 6 bits [7:5]); selects sprite set
	Variation    uint8 // per-tile flat-grass variation (byte 7); 0 = plain flat image
	IsIndustrial bool  // surface owned by an industry; TerrainIndex is not a land slot
	Trees        []scenario.TreeElement
	Buildings    []scenario.BuildingElement
	Walls        []scenario.WallElement
	Tracks       []scenario.TrackElement
	Roads        []scenario.RoadElement
	Stations     []scenario.StationElement
	Signals      []scenario.SignalElement
	Industries   []scenario.IndustryElement
}
