package sim

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// clearRoad's helpers (GameCommands/Road/CreateRoad.cpp 201–254, and
// GameCommands.cpp's checkCompanyCompatibility), item 5.9f3. The error texts
// are not kept (as elsewhere in goloco's game commands).

// roadTraitJunction is RoadTraitFlags::junction.
const roadTraitJunction = 1 << 6

// roadObject is the road object in a slot, or nil.
func (c *Cmd) roadObject(id uint8) *objects.RoadObject {
	if c.Objects != nil && int(id) < len(c.Objects.RoadObjects) {
		return c.Objects.RoadObjects[id]
	}
	return nil
}

// companyCompatible is checkCompanyCompatibility(Common): a neutral owner, or
// the updating company owns it, or the updating company is neutral (a town).
func (c *Cmd) companyCompatible(owner uint8) bool {
	return owner == CompanyNeutral || c.Company == owner || c.Company == CompanyNeutral
}

// aiUnkCheck is CreateRoad.cpp's aiUnkCheck.
func (c *Cmd) aiUnkCheck(e *scenario.RoadElement, a roadClearArgs) bool {
	if a.UnkFlags&(1<<4|1<<5) == 0 || a.Flags&CmdAiAllocated == 0 || e.AiAllocated {
		return false
	}
	return a.UnkFlags&(1<<4) != 0 || e.Owner == c.Company
}

// compatibleRoads is a road object's compatibleRoads (its compatible names
// resolved to slots).
func (c *Cmd) compatibleRoads(r *objects.RoadObject) uint16 {
	if r == nil || c.Objects == nil {
		return 0
	}
	_, roads := c.Objects.CompatibleSets(r.CompatibleNames)
	return roads
}

// roadTraitAndCompatibilityCheck is CreateRoad.cpp's
// roadTraitAndCompatbilityCheck: the same road object needs the junction
// trait; different ones must be compatible one way or the other.
func (c *Cmd) roadTraitAndCompatibilityCheck(e *scenario.RoadElement, a roadClearArgs, target, placed *objects.RoadObject) bool {
	if e.RoadObjectID == a.RoadObjectID {
		return target != nil && target.RoadPieces&roadTraitJunction != 0
	}
	return c.compatibleRoads(target)&(1<<a.RoadObjectID) != 0 || c.compatibleRoads(placed)&(1<<e.RoadObjectID) != 0
}

// finalJunctionChecks is CreateRoad.cpp's: a junction only between
// single-tile pieces, then the trait and compatibility check.
func (c *Cmd) finalJunctionChecks(e *scenario.RoadElement, a roadClearArgs, target, placed *objects.RoadObject) bool {
	if len(RoadPieces[a.RoadID]) > 1 || len(RoadPieces[e.RoadID]) > 1 {
		return false
	}
	return c.roadTraitAndCompatibilityCheck(e, a, target, placed)
}

// CommonTraitFlags slope / steepSlope (TrackEnum.h).
const (
	commonTraitSlope      = 1 << 0
	commonTraitSteepSlope = 1 << 1
)

// roadStationAt is CreateRoad.cpp's getRoadStationElement: the tile's road
// station element at height z (world units), or nil.
func (m *Map) roadStationAt(x, y, z int32) *scenario.StationElement {
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if t == nil {
		return nil
	}
	for i := range t.Stations {
		if s := &t.Stations[i]; int32(s.BaseZ) == z/4 && s.StationType == 1 {
			return s
		}
	}
	return nil
}

func nameIn(names []string, h *objects.ObjectHeader) bool {
	if h == nil {
		return false
	}
	for _, n := range names {
		if strings.EqualFold(n, h.GetName()) {
			return true
		}
	}
	return false
}

// clearRoad is CreateRoad.cpp's clearRoad (0x004771C0 … 0x0047744C): placing
// road over a road element — no collision when their connections don't meet;
// otherwise a junction, an overlay (another road type on the same piece) or a
// collision, by the rules of junctions, heights, signals, road stations,
// bridges and the pieces' sequences. st takes the state the game command reads
// afterwards (flags_1136073 bits 5 / 6, the bridge, the alternate road object).
func (m *Map) clearRoad(c *Cmd, st *roadCmdState, e *scenario.RoadElement, a roadClearArgs) roadClearResult {
	if !e.Ghost && !e.AiAllocated {
		st.Flags73 |= 1 << 6
	}
	if e.HasBridge {
		st.Bridge = e.BridgeID
	}
	tp, np := RoadPieces[e.RoadID], RoadPieces[a.RoadID]
	if int(e.SequenceIndex) >= len(tp) || int(a.Index) >= len(np) ||
		tp[e.SequenceIndex].ConnectFlags[e.Rotation&3]&np[a.Index].ConnectFlags[a.Rotation&3] == 0 {
		return roadRes(ClearNoCollision)
	}
	target, placed := c.roadObject(e.RoadObjectID), c.roadObject(a.RoadObjectID)
	if a.Flags&CmdAiAllocated != 0 && a.UnkFlags&(1<<4) != 0 && target != nil && target.Flags&1 != 0 { // isOneWay
		return roadRes(ClearCollisionErrorSet)
	}
	if m.RoadUsableByAll&(1<<e.RoadObjectID) == 0 || m.RoadUsableByAll&(1<<a.RoadObjectID) == 0 {
		if !c.companyCompatible(e.Owner) {
			return roadRes(ClearCollisionErrorSet)
		}
	}
	slope := func(id uint8) bool { return RoadMiscFlags[id]&(commonTraitSlope|commonTraitSteepSlope) != 0 }
	confirmedOverlay := false
	aiUnk := c.aiUnkCheck(e, a)
	if int32(e.BaseZ)*4 == a.Z && slope(e.RoadID) && isPotentialOverlayRoad(e, a) {
		st.Flags73 |= 1 << 5
		if e.RoadObjectID != a.RoadObjectID || aiUnk {
			confirmedOverlay = true
		}
	}
	if !confirmedOverlay {
		if slope(e.RoadID) || slope(a.RoadID) { // junction_must_be_entirely_level
			return roadRes(ClearCollisionErrorSet)
		}
		if int32(e.BaseZ)*4 != a.Z { // in the way: wrong height for a junction
			return roadRes(ClearCollisionErrorSet)
		}
	}
	res := roadRes(ClearNoCollision)
	if e.HasLevelCrossing {
		res.LevelCrossingObjID = e.LevelCrossingObjectID
	}
	st.AltRoadObject = e.RoadObjectID
	if e.HasSignalElement {
		return roadRes(ClearCollisionErrorSet)
	}
	if e.HasStationElement {
		if a.RoadID != 0 || (a.Rotation != e.Rotation && a.Rotation != e.Rotation^2) {
			return roadRes(ClearCollisionErrorSet)
		}
		if s := m.roadStationAt(a.X, a.Y, a.Z); s != nil {
			res.StationID = s.StationID
			var so *objects.RoadStationObject
			if c.Objects != nil {
				so = c.Objects.GetRoadStationObjectByIndex(int(s.ObjectID))
			}
			// the road lists the station, or the station the road
			if placed != nil && so != nil && (nameIn(placed.StationNames, so.Header) || nameIn(so.CompatibleNames, placed.Header)) {
				res.HasStation = true
			} else {
				return roadRes(ClearCollisionErrorSet)
			}
		}
	}
	isExact := a.RoadID == e.RoadID && a.Rotation == e.Rotation && a.Index == e.SequenceIndex
	isReverseExact := !isExact && RoadReverseRoadID[e.RoadID] == a.RoadID &&
		(RoadReverseRotation[e.RoadID]+e.Rotation)&3 == a.Rotation && RoadReverseLane[e.RoadID] == RoadReverseLane[a.RoadID]
	isConnectableSequence := (e.SequenceIndex == 0 && a.IsLastIndex) ||
		(e.SequenceIndex != 0 && e.IsLast && a.Index == 0) ||
		(e.SequenceIndex != 0 && !e.IsLast && e.SequenceIndex == a.Index)
	if e.HasBridge {
		if e.BridgeID != a.BridgeID { // bridge types must match
			return roadRes(ClearCollisionErrorSet)
		}
		if b := c.bridge(a.BridgeID); b != nil && b.DisabledTrackCfg&commonTraitJunction != 0 {
			if a.RoadObjectID == e.RoadObjectID || (!isExact && (!isReverseExact || !isConnectableSequence)) {
				return roadRes(ClearCollisionErrorSet) // bridge not suitable for a junction
			}
		}
	}
	junction := func() roadClearResult { // 0x004773FE
		if !c.finalJunctionChecks(e, a, target, placed) {
			return roadRes(ClearCollisionErrorSet)
		}
		if e.RoadObjectID == a.RoadObjectID {
			res.RoadIDUnk[e.RoadID] |= 1 << e.Rotation
			res.RoadIDUnk[0] |= 1 << 7
		}
		return res
	}
	sameOrCompatible := func() roadClearResult { // 0x004773A5 / the exact overlay
		st.Flags73 |= 1 << 5
		if e.RoadObjectID == a.RoadObjectID {
			if aiUnk {
				return junction()
			}
			return roadRes(ClearCollisionErrorSet) // already built here
		}
		if !c.roadTraitAndCompatibilityCheck(e, a, target, placed) {
			return roadRes(ClearCollisionErrorSet)
		}
		return res
	}
	switch {
	case isExact:
		return sameOrCompatible()
	case isReverseExact:
		if !isConnectableSequence {
			return junction()
		}
		return sameOrCompatible()
	}
	return junction()
}
