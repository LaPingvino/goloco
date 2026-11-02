package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// ExportVeh1Move and ExportVeh1MoveNoMove expose Vehicle1's piece moves for
// tests outside the package (pkg/scene's save-driven checks).
func ExportVeh1Move(m *Map, v *Veh1, n uint8) int32       { return m.veh1MoveByPieces(v, n) }
func ExportVeh1MoveNoMove(m *Map, v *Veh1, n uint8) int32 { return m.veh1MoveByPiecesNoMove(v, n) }

// ExportTrainStation exposes Tile::trainStation to pkg/scene's checks.
func ExportTrainStation(t *Tile, trackID, rotation, baseZ uint8) *scenario.StationElement {
	return t.trainStation(trackID, rotation, baseZ)
}
