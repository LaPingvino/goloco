package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// createTrack's clear function (GameCommands/Track/CreateTrack.cpp 50–320),
// PL.1 (e), createTrack part 2. This is its start: the state and arguments,
// and clearTrack (0x0049C4FF) — new track meeting track already there.
// PARTIAL: clearRoad (0x0049C401, track crossing road), clearFunction
// (0x0049C275, the dispatch with buildings and trees), and createTrack's
// per-piece checks and apply step that use them.

// commonTraitOneSided is CommonTraitFlags::oneSided (TrackEnum.h; slope and
// steepSlope are in roadclearhelp.go).
const commonTraitOneSided = 1 << 9

// trackTraitJunction is TrackTraitFlags::junction (a track object's pieces).
const trackTraitJunction = 1 << 10

// trackCmdState is createTrack's legacy return state (flags_1136072/73,
// byte_1136074/75).
type trackCmdState struct {
	PosFlags      uint8 // flags_1136072
	Flags73       uint8 // flags_1136073: bit 0/1 bridge, 3 a junction made, 7 over water
	MaxBridgeDiff uint8 // byte_1136074
	Bridge        uint8 // byte_1136075 (0xFF none)
}

// trackClearArgs is CreateTrack.cpp's ClearFunctionArgs.
type trackClearArgs struct {
	X, Y, Z       int32
	Rotation      uint8
	TrackID       uint8
	UnkFlags      uint8
	BridgeID      uint8
	TrackObjectID uint8
	Index         uint8
	IsLastIndex   bool
	Flags         uint8
}

// trackClearTrack is clearTrack: new track over a track element. Pieces that
// don't share a connection pass by; otherwise only a level junction of the
// same height, owner-compatible, without signal or station, of compatible
// objects (or a junction-capable one), with a matching junction-capable
// bridge, and not the very piece already built (or its reverse).
func (m *Map) trackClearTrack(c *Cmd, st *trackCmdState, e *scenario.TrackElement, a trackClearArgs) ClearResult {
	if e.HasBridge {
		st.Bridge = e.BridgeID
	}
	if int(e.TrackID) >= len(TrackPieces) || int(e.SequenceIndex) >= len(TrackPieces[e.TrackID]) ||
		int(a.TrackID) >= len(TrackPieces) || int(a.Index) >= len(TrackPieces[a.TrackID]) {
		return ClearCollisionErrorSet
	}
	target := TrackPieces[e.TrackID][e.SequenceIndex]
	piece := TrackPieces[a.TrackID][a.Index]
	if target.ConnectFlags[e.Rotation&3]&piece.ConnectFlags[a.Rotation&3] == 0 {
		return ClearNoCollision
	}
	if a.UnkFlags&(1<<2) != 0 { // junctions_not_possible
		return ClearCollisionErrorSet
	}
	if !c.companyCompatible(e.Owner) {
		return ClearCollisionErrorSet
	}
	if TrackMiscFlags[e.TrackID]&(commonTraitSlope|commonTraitSteepSlope) != 0 ||
		TrackMiscFlags[a.TrackID]&(commonTraitSlope|commonTraitSteepSlope) != 0 { // junction_must_be_entirely_level
		return ClearCollisionErrorSet
	}
	if int32(e.BaseZ)*4 != a.Z { // wrong height for a junction
		return ClearCollisionErrorSet
	}
	if e.HasSignal || e.HasStation {
		return ClearCollisionErrorSet
	}
	if (a.Rotation^e.Rotation)&1 == 0 && (TrackMiscFlags[e.TrackID]^TrackMiscFlags[a.TrackID])&commonTraitOneSided != 0 {
		return ClearCollisionErrorSet // track_combination_not_possible
	}
	if e.HasBridge {
		if e.BridgeID != a.BridgeID { // bridge_types_must_match
			return ClearCollisionErrorSet
		}
		if b := c.bridge(a.BridgeID); b != nil && b.DisabledTrackCfg&commonTraitJunction != 0 {
			return ClearCollisionErrorSet // bridge_not_suitable_for_junction
		}
	}
	if e.TrackObjectID == a.TrackObjectID {
		if e.TrackID == a.TrackID && e.Rotation == a.Rotation && e.SequenceIndex == a.Index {
			return ClearCollisionErrorSet // already_built_here
		}
		if TrackReverseTrackID[e.TrackID] == a.TrackID && (TrackReverseRotation[e.TrackID]+e.Rotation)&3 == a.Rotation &&
			a.IsLastIndex && e.SequenceIndex == 0 {
			return ClearCollisionErrorSet // already_built_here (reversed)
		}
	}
	om := c.Objects
	var targetObj, newObj = c.trackObject(e.TrackObjectID), c.trackObject(a.TrackObjectID)
	if e.TrackObjectID == a.TrackObjectID {
		if targetObj == nil || targetObj.TrackPieces&trackTraitJunction == 0 { // junctions_not_possible
			return ClearCollisionErrorSet
		}
	} else {
		var targetTracks, newTracks uint16
		if targetObj != nil && om != nil {
			targetTracks, _ = om.CompatibleSets(targetObj.CompatibleNames)
		}
		if newObj != nil && om != nil {
			newTracks, _ = om.CompatibleSets(newObj.CompatibleNames)
		}
		if targetTracks&(1<<a.TrackObjectID) == 0 && newTracks&(1<<e.TrackObjectID) == 0 {
			return ClearCollisionErrorSet // unable_to_cross_or_create_junction_with
		}
	}
	st.Flags73 |= 1 << 3
	return ClearNoCollision
}
