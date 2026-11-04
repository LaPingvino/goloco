package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// RaiseLand / LowerLand on one whole tile (adjustSurfaceHeight): a flat tile
// goes a step up and back down; not below the map's lowest height; not from
// under a road without a bridge; water under ground raised above it is gone.
// And a town road's end a step above the ground raises it (appendToRoadEnd).
func TestRaiseLowerLand(t *testing.T) {
	c := townRoadCmd()
	if raiseSurfaceCornerMapping[mapSelectionFull][0] != slopeRequiresHeightAdjustment || lowerSurfaceCornerMapping[mapSelectionFull][0] != slopeRequiresHeightAdjustment {
		t.Fatalf("upstream's full-tile rows for a flat tile: %d / %d", raiseSurfaceCornerMapping[mapSelectionFull][0], lowerSurfaceCornerMapping[mapSelectionFull][0])
	}
	m := townRoadMap(8, 8)
	if _, ok := m.raiseLand(c, 3, 3, 3, 3, mapSelectionFull, CmdApply); !ok || m.TileAt(3, 3).BaseZ != 12 || m.TileAt(3, 3).Slope != 0 {
		t.Fatalf("raise: ok %v, baseZ %d slope %d", ok, m.TileAt(3, 3).BaseZ, m.TileAt(3, 3).Slope)
	}
	if m.TileAt(4, 3).BaseZ != 8 {
		t.Error("a neighbour moved")
	}
	if _, ok := m.lowerLand(c, 3, 3, 3, 3, mapSelectionFull, CmdApply); !ok || m.TileAt(3, 3).BaseZ != 8 {
		t.Errorf("lower: ok %v, baseZ %d", ok, m.TileAt(3, 3).BaseZ)
	}

	low := townRoadMap(8, 4)
	if _, ok := low.lowerLand(c, 3, 3, 3, 3, mapSelectionFull, CmdApply); ok || low.TileAt(3, 3).BaseZ != 4 {
		t.Error("below the lowest height: refused, unchanged")
	}

	road := townRoadMap(8, 8)
	rt := road.TileAt(3, 3)
	rt.Roads = []scenario.RoadElement{{BaseZ: 8, ClearZ: 16}}
	rt.insertRef(scenario.KindRoad, 0, 8)
	if _, ok := road.lowerLand(c, 3, 3, 3, 3, mapSelectionFull, CmdApply); ok || rt.BaseZ != 8 {
		t.Error("lowering the ground from under a road with no bridge: refused")
	}

	wet := townRoadMap(8, 4)
	wet.TileAt(3, 3).WaterLevel = 2 // water at small z 8
	if _, ok := wet.raiseLand(c, 3, 3, 3, 3, mapSelectionFull, CmdApply); !ok || wet.TileAt(3, 3).WaterLevel != 0 {
		t.Errorf("raised above the water: ok %v, water %d", ok, wet.TileAt(3, 3).WaterLevel)
	}

	town := townRoadMap(16, 4)
	town.RoadAnyTypeCompatible = 1
	town.appendToRoadEnd(c, 0, 5*32, 5*32, 32, 2, 0, 0, false) // the road's end at z 32, the ground at 16
	if town.TileAt(5, 5).BaseZ != 8 || len(town.TileAt(5, 5).Roads) != 0 {
		t.Errorf("town road a step up: ground %d, roads %d (the ground raised, no road yet)", town.TileAt(5, 5).BaseZ, len(town.TileAt(5, 5).Roads))
	}
}
