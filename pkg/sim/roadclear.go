package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// createRoad's clear function (GameCommands/Road/CreateRoad.cpp, item 5.9f3).

// roadCmdState is the legacy return state createRoad and its clear function
// share (flags_1136072/73, byte_1136074/75, alternateRoadObjectId).
type roadCmdState struct {
	PosFlags      uint8 // flags_1136072
	Flags73       uint8 // flags_1136073: bit 0/1 bridge, 2 level crossing, 4 building cleared, 7 over water
	MaxBridgeDiff uint8 // byte_1136074
	Bridge        uint8 // byte_1136075 (0xFF none)
	AltRoadObject uint8 // alternateRoadObjectId (0xFF none)
}

// roadClearArgs is CreateRoad.cpp's ClearFunctionArgs.
type roadClearArgs struct {
	X, Y, Z      int32
	Rotation     uint8
	RoadID       uint8
	BridgeID     uint8
	RoadObjectID uint8
	Index        uint8
	IsLastIndex  bool
	Flags        uint8
	UnkFlags     uint8
}

// roadClearResult is RoadClearFunctionResult.
type roadClearResult struct {
	Res                ClearResult
	HasLevelCrossing   bool
	LevelCrossingObjID uint8 // 0xFF none
	HasStation         bool
	StationID          uint16 // 0xFFFF none
	RoadIDUnk          [16]uint8
}

func roadRes(r ClearResult) roadClearResult {
	return roadClearResult{Res: r, LevelCrossingObjID: 0xFF, StationID: 0xFFFF}
}

const commonTraitJunction = 1 << 11

// clearTrack is CreateRoad.cpp's clearTrack: a road may cross track only as a
// level crossing — straight road over straight track at right angles, the same
// height, compatible objects, no signal, station or mismatched / junction-less
// bridge — and the crossing is marked on the track when applied.
func (m *Map) clearTrack(c *Cmd, st *roadCmdState, e *scenario.TrackElement, a roadClearArgs) roadClearResult {
	if e.HasBridge {
		st.Bridge = e.BridgeID
		if b := c.bridge(e.BridgeID); b != nil && b.DisabledTrackCfg&commonTraitJunction != 0 {
			return roadRes(ClearCollisionErrorSet)
		}
	}
	om := c.Objects
	var trackRoads, roadTracks uint16
	if int(e.TrackObjectID) < len(om.TrackObjects) && om.TrackObjects[e.TrackObjectID] != nil {
		_, trackRoads = om.CompatibleSets(om.TrackObjects[e.TrackObjectID].CompatibleNames)
	}
	if int(a.RoadObjectID) < len(om.RoadObjects) && om.RoadObjects[a.RoadObjectID] != nil {
		roadTracks, _ = om.CompatibleSets(om.RoadObjects[a.RoadObjectID].CompatibleNames)
	}
	if trackRoads&(1<<a.RoadObjectID) == 0 && roadTracks&(1<<e.TrackObjectID) == 0 {
		return roadRes(ClearCollisionErrorSet)
	}
	if e.HasSignal || e.HasStation {
		return roadRes(ClearCollisionErrorSet)
	}
	if e.HasBridge && e.BridgeID != a.BridgeID {
		return roadRes(ClearCollisionErrorSet)
	}
	if e.TrackID != 0 || a.RoadID != 0 || (e.Rotation-a.Rotation)&1 == 0 || int32(e.BaseZ)*4 != a.Z {
		return roadRes(ClearCollisionErrorSet)
	}
	res := roadRes(ClearNoCollision)
	res.HasLevelCrossing = true
	st.Flags73 |= 1 << 2
	if a.Flags&(CmdAiAllocated|CmdGhost) == 0 && a.Flags&CmdApply != 0 {
		e.HasLevelCrossing = true
	}
	return res
}

// isPotentialOverlayRoad is CreateRoad.cpp's: the same piece (id, rotation,
// sequence) or its reverse (reverse road id, rotation ^ 2, sequence counted
// from the other end).
func isPotentialOverlayRoad(e *scenario.RoadElement, a roadClearArgs) bool {
	if e.RoadID == a.RoadID && e.Rotation == a.Rotation && e.SequenceIndex == a.Index {
		return true
	}
	return RoadReverseRoadID[e.RoadID] == a.RoadID && e.Rotation^2 == a.Rotation &&
		RoadReverseLane[e.RoadID]-1-e.SequenceIndex == a.Index
}

// roadClearFunction is CreateRoad.cpp's clearFunction (0x0047744C).
//
// PARTIAL: the road branch (clearRoad: overlays, junctions, station roads) is
// a collision until clearRoad is ported (next in 5.9f3).
func (m *Map) roadClearFunction(c *Cmd, st *roadCmdState, t *Tile, k int, cost *int32, removed RemovedBuildings, a roadClearArgs) roadClearResult {
	r := t.Order[k]
	switch r.Kind {
	case scenario.KindTrack:
		return m.clearTrack(c, st, &t.Tracks[r.Index], a)
	case scenario.KindStation:
		if t.Stations[r.Index].StationType == 1 { // roadStation
			return roadRes(ClearNoCollision)
		}
		return roadRes(ClearCollision)
	case scenario.KindSignal:
		return roadRes(ClearNoCollision)
	case scenario.KindBuilding:
		st.Flags73 |= 1 << 4
		return roadRes(m.clearBuildingCollision(c, t, k, a.X, a.Y, removed, a.Flags, cost))
	case scenario.KindTree:
		return roadRes(m.clearTreeCollision(c, t, k, a.X, a.Y, a.Flags, cost))
	case scenario.KindRoad:
		return roadRes(ClearCollision) // PARTIAL: clearRoad
	}
	return roadRes(ClearCollision)
}

// bridge is the bridge object in a slot, or nil.
func (c *Cmd) bridge(id uint8) *objects.BridgeObject {
	if int(id) < len(c.Objects.BridgeObjects) {
		return c.Objects.BridgeObjects[id]
	}
	return nil
}
