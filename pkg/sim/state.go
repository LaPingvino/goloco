package sim

// State is everything the simulation owns: the map, plus the player's
// construction head and undo history. pkg/world embeds it and draws it.
type State struct {
	Map
	ConsHead      ConstructionHead // shared track/road construction head
	consStack     []placedPiece    // undo stack for placed track pieces
	consRoadStack []placedPiece    // undo stack for placed road pieces
}
