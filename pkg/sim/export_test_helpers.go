package sim

// ExportVeh1Move and ExportVeh1MoveNoMove expose Vehicle1's piece moves for
// tests outside the package (pkg/scene's save-driven checks).
func ExportVeh1Move(m *Map, v *Veh1, n uint8) int32       { return m.veh1MoveByPieces(v, n) }
func ExportVeh1MoveNoMove(m *Map, v *Veh1, n uint8) int32 { return m.veh1MoveByPiecesNoMove(v, n) }
