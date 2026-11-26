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
	// ConsSelection is the construction tool's selected tiles
	// (enableConstruct with the free-form tiles), outlined on the map.
	ConsSelection ConstructionSelection
	// GhostVisibility is told when a construction ghost is placed: whether
	// it is under the ground and whether a gradient is selected — for the
	// view's underground mode and height marks (the world sets it; nil: no view).
	GhostVisibility func(underground, sloped bool)
	// SwitchRoadType is Common::sub_4A3A50 after constructRoad's retry chose
	// another road type (sub_4A18D4): the window takes the road object obj
	// (cState.trackType) with its stations, bridge (Cons.Bridge) and mods
	// (Cons.Mods). nil: the type can't change (no window).
	SwitchRoadType func(obj uint8)
	// TrackGhost / RoadGhost is the construction ghost on the map
	// (constructionghost.go); ghostPos its last position under the cursor
	// (ghostTrackPos).
	TrackGhost *TrackGhost
	RoadGhost  *RoadGhost
	ghostPos   [3]int32
	// blueprint is the copy tool's _copiedTrack, blueprintGhost where its
	// ghost stands (GhostVisibilityFlags::blueprint, ghostRemovalTrackPos;
	// nil none), blueprintG1 the station commands' G1 (blueprint.go).
	blueprint      *Blueprint
	blueprintGhost *[3]int32
	blueprintG1    G1Raw

	// Waves (wave.go): the wave table (they draw from Map.Rng2).
	Waves [MaxWaves]Wave
}
