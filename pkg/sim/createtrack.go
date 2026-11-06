package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// createTrack (GameCommands/Track/CreateTrack.cpp 321–585), PL.1 (e): the
// game command a player builds track with (the construction window's "Build
// this…"). Ported in parts, as createRoad was; this is its start: the
// arguments, the track extras that fit the piece, the height rule and the
// cost of the track and its extras.
// PARTIAL: the per-piece checks (height, bridges, clearance with the clear
// function, position flags), the apply step (inserting the elements), the
// bridge / tunnel costs, the company observation and the sound. Until then
// the window still builds with PlaceTrackAtHead (no cost — LOOP PL.1 e).

// trackPlacementArgs is TrackPlacementArgs.
type trackPlacementArgs struct {
	X, Y, Z       int32
	Rotation      uint8 // 0–3
	TrackID       uint8 // 0–43
	Mods          uint8 // the track extras asked for (bits 0–3)
	Bridge        uint8 // a bridge object slot, 0xFF none
	TrackObjectID uint8
	UnkFlags      uint8
}

// trackObject is the track object in a slot, or nil.
func (c *Cmd) trackObject(id uint8) *objects.TrackObject {
	if c.Objects != nil && int(id) < len(c.Objects.TrackObjects) {
		return c.Objects.TrackObjects[id]
	}
	return nil
}

// createTrackCost is createTrack's start (0x0049BB98 up to the pieces): the
// extras asked for that fit this piece (its compatibleFlags all in the
// extra's trackPieces), the height (a multiple of 16: no Loco piece has
// startsAtHalfHeight, the RCT2 leftover), and the cost — the track's and each
// valid extra's inflation-adjusted build cost (divisor 10) times the piece's
// cost factor / 256.
func (m *Map) createTrackCost(c *Cmd, a trackPlacementArgs) (validMods uint8, cost int32, ok bool) {
	track := c.trackObject(a.TrackObjectID)
	if track == nil || int(a.TrackID) >= len(TrackCompatibleFlags) {
		return 0, 0, false
	}
	compat := TrackCompatibleFlags[a.TrackID]
	validMods = a.Mods
	extra := func(i int) *objects.ExtraObject {
		if i >= len(track.ModNames) || c.Objects == nil {
			return nil
		}
		return c.Objects.Extra(track.ModNames[i])
	}
	for i := range 4 {
		if a.Mods&(1<<i) == 0 {
			continue
		}
		if ex := extra(i); ex == nil || ex.Pieces&compat != compat {
			validMods &^= 1 << i
		}
	}
	if a.Z&0xF != 0 {
		return validMods, 0, false
	}
	factor := int32(TrackCostFactor[a.TrackID])
	cost = m.InflationAdjustedCost(track.BuildCostFactor, track.CostIndex, 10) * factor / 256
	for i := range 4 {
		if validMods&(1<<i) != 0 {
			ex := extra(i)
			cost += m.InflationAdjustedCost(ex.BuildCostFactor, ex.CostIndex, 10) * factor / 256
		}
	}
	return validMods, cost, true
}
