package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// ExportVeh1Move and ExportVeh1MoveNoMove expose Vehicle1's piece moves for
// tests outside the package (pkg/scene's save-driven checks).
func ExportVeh1Move(m *Map, v *Veh1, n uint8) int32       { return m.veh1MoveByPieces(v, n) }
func ExportVeh1MoveNoMove(m *Map, v *Veh1, n uint8) int32 { return m.veh1MoveByPiecesNoMove(v, n) }

// ExportTrainStation exposes Tile::trainStation to pkg/scene's checks.
func ExportTrainStation(t *Tile, trackID, rotation, baseZ uint8) *scenario.StationElement {
	return t.trainStation(trackID, rotation, baseZ)
}

// FindRoadExtentForTest exposes findRoadExtent to other packages' tests.
func FindRoadExtentForTest(m *Map, i int, om *objects.ObjectManager) (roadExtent, bool) {
	return m.findRoadExtent(i, om)
}

// IdealTownRoadIDForTest exposes getIdealTownRoadId (−1 when none).
func IdealTownRoadIDForTest(m *Map, i int, om *objects.ObjectManager) int {
	if id, ok := idealTownRoadID(m.Towns[i].Size, om); ok {
		return int(id)
	}
	return -1
}

// TileClearFunctionForTest exposes tileClearFunction (default collision).
func TileClearFunctionForTest(m *Map, c *Cmd, t *Tile, k int, px, py int32, flags uint8, cost *int32) ClearResult {
	return m.tileClearFunction(c, t, k, px, py, RemovedBuildings{}, flags, cost, true)
}
