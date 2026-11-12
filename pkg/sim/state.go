package sim

import "github.com/LaPingvino/goloco/pkg/core"

// State is everything the simulation owns: the map, plus the player's
// construction head and undo history. pkg/world embeds it and draws it.
type State struct {
	Map
	ConsHead ConstructionHead // shared track/road construction head
	// ConsBridge is cState.lastSelectedBridge + 1 (0: none): the bridge type a
	// track or road piece is built with where it needs one (getTrackPlacementArgs
	// / getRoadPlacementArgs). The game sets it from the window's choice.
	ConsBridge    uint8
	consStack     []placedPiece // undo stack for placed track pieces
	consRoadStack []placedPiece // undo stack for placed road pieces
	// TrackGhost is the construction ghost while positioning (trackghost.go);
	// ghostPos / ghostCost its last position (ghostTrackPos) and cost.
	TrackGhost *TrackGhost
	// RoadGhost is the road construction ghost at the head (roadghost.go).
	RoadGhost *RoadGhost
	ghostPos  [3]int32
	ghostCost int32

	// Waves (wave.go): the wave table, the generator they draw from (upstream
	// gPrng2) and the tile loop's start offset within its 16×16 block.
	Waves                    [MaxWaves]Wave
	Rng2                     core.Prng
	tileUpdateX, tileUpdateY int
}
