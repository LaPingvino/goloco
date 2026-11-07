package sim

import "github.com/LaPingvino/goloco/pkg/core"

// State is everything the simulation owns: the map, plus the player's
// construction head and undo history. pkg/world embeds it and draws it.
type State struct {
	Map
	ConsHead      ConstructionHead // shared track/road construction head
	consStack     []placedPiece    // undo stack for placed track pieces
	consRoadStack []placedPiece    // undo stack for placed road pieces
	// TrackGhost is the construction ghost while positioning (trackghost.go);
	// ghostPos / ghostCost its last position (ghostTrackPos) and cost.
	TrackGhost *TrackGhost
	ghostPos   [3]int32
	ghostCost  int32

	// Waves (wave.go): the wave table, the generator they draw from (upstream
	// gPrng2) and the tile loop's start offset within its 16×16 block.
	Waves                    [MaxWaves]Wave
	Rng2                     core.Prng
	tileUpdateX, tileUpdateY int
}
