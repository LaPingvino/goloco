package sim

// State is everything the simulation owns: the map, plus the construction
// window's state. pkg/world embeds it and draws it.
type State struct {
	Map
	// Cons is the construction window's state (construction.go).
	Cons ConstructionState
	// ConsArrow is the construction arrow on the map (setConstructionArrow +
	// enableConstructionArrow): steady under the cursor while hovering,
	// blinking at the position while building.
	ConsArrow ConstructionArrow
	// TrackGhost / RoadGhost is the construction ghost on the map
	// (constructionghost.go); ghostPos its last position under the cursor
	// (ghostTrackPos).
	TrackGhost *TrackGhost
	RoadGhost  *RoadGhost
	ghostPos   [3]int32

	// Waves (wave.go): the wave table (they draw from Map.Rng2).
	Waves [MaxWaves]Wave
}
