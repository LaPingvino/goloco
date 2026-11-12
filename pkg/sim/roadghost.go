package sim

import "github.com/LaPingvino/goloco/pkg/i18n"

// The construction window's road ghost at the head (ConstructionTab
// updateConstruction for roads: getRoadPlacementArgs, placeRoadGhost and
// removeConstructionGhosts' road part), as trackghost.go does for track.

// RoadGhost is the road ghost on the map (ghostRemovalTrack* for a road), nil none.
type RoadGhost struct {
	X, Y, Z      int32
	Rotation     uint8
	RoadID       uint8
	RoadObjectID uint8
}

// RemoveRoadGhost is removeConstructionGhosts' road part: RoadRemoval at the
// ghost's first tile (its height plus the piece's first z, sequence 0) with
// apply, no error window, no payment, ghost; then forgotten.
func (s *State) RemoveRoadGhost(c *Cmd) {
	g := s.RoadGhost
	if g == nil {
		return
	}
	s.RoadGhost = nil
	z := g.Z
	if int(g.RoadID) < len(RoadPieces) && len(RoadPieces[g.RoadID]) > 0 {
		z += int32(RoadPieces[g.RoadID][0].Z)
	}
	s.RemoveRoad(c, CmdApply|CmdNoErrorWindow|CmdNoPayment|CmdGhost, RoadRemoval{X: g.X, Y: g.Y, Z: z, Rotation: g.Rotation & 3, RoadID: g.RoadID, Index: 0, RoadObjectID: g.RoadObjectID}, nil)
}

// placeRoadGhost is placeRoadGhost (ConstructionTab): the old ghost removed,
// then createRoad with apply, prevent building clearing, no error window, no
// payment and ghost; placed, it is remembered for removal.
// PARTIAL: the bridge retry (as placeTrackGhost's).
func (s *State) placeRoadGhost(c *Cmd, a roadPlacementArgs) (int32, bool) {
	s.RemoveRoadGhost(c)
	cost, ok := s.DoCommandFor(c, CmdApply|CmdPreventBuildingClearing|CmdNoErrorWindow|CmdNoPayment|CmdGhost, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.createRoad(c, a, f)
	})
	if ok {
		s.RoadGhost = &RoadGhost{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, RoadID: a.RoadID, RoadObjectID: a.RoadObjectID}
	}
	return cost, ok
}

// BuildRoadAtHead is constructRoad (ConstructionTab): the ghost taken up, then
// the createRoad game command at the head (checked and charged under
// Construction, the element flags set as upstream's insertion does), the head
// advanced along the piece and the piece kept for undo. Before, the player's
// road went in through AddRoad — no checks, no charge, and refs without the
// occupied quarters the clearance checks read (LOOP 5.9g). The cost and
// whether it was built. PARTIAL: the bridge retry.
func (s *State) BuildRoadAtHead(c *Cmd, roadID int, roadObjID uint8) (int32, bool) {
	h := s.ConsHead
	if !h.Active || roadID < 0 || roadID >= len(RoadPieces) || h.Rotation > 3 {
		return 0, false
	}
	tiles, baseZ, ok := s.RoadFootprint(roadID)
	if !ok {
		return 0, false
	}
	s.RemoveRoadGhost(c)
	a := roadPlacementArgs{X: int32(h.X) * 32, Y: int32(h.Y) * 32, Z: int32(h.BaseZ) * 4, Rotation: h.Rotation & 3,
		RoadID: uint8(roadID), RoadObjectID: roadObjID, Bridge: 0xFF}
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	cost, ok := s.DoCommandFor(c, CmdApply, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.createRoad(c, a, f)
	})
	if !ok {
		return 0, false
	}
	s.consRoadStack = append(s.consRoadStack, placedPiece{trackID: roadID, rotation: h.Rotation, tiles: tiles, BaseZ: baseZ})
	rc := RoadCoordinates[roadID<<3|int(h.Rotation&3)]
	s.ConsHead.X += int(rc.DX) / 32
	s.ConsHead.Y += int(rc.DY) / 32
	s.ConsHead.BaseZ = uint8(int(s.ConsHead.BaseZ) + int(rc.DZ)/4)
	s.ConsHead.Rotation = rc.RotEnd
	return cost, true
}

// UpdateHeadRoadGhost is updateConstruction for a road: with the head placed,
// the selected road piece roadID stands as a ghost at the head
// (getRoadPlacementArgs at cState.x, y, constructionZ), placed again when the
// head or piece change (as UpdateHeadTrackGhost). roadID < 0 or no head: removed.
func (s *State) UpdateHeadRoadGhost(c *Cmd, roadID int, objID uint8) {
	h := s.ConsHead
	if !h.Active || roadID < 0 || roadID >= len(RoadPieces) || h.Rotation > 3 {
		s.RemoveRoadGhost(c)
		return
	}
	a := roadPlacementArgs{X: int32(h.X) * 32, Y: int32(h.Y) * 32, Z: int32(h.BaseZ) * 4, Rotation: h.Rotation & 3,
		RoadID: uint8(roadID), RoadObjectID: objID, Bridge: 0xFF}
	if g := s.RoadGhost; g != nil && g.X == a.X && g.Y == a.Y && g.Z == a.Z && g.Rotation == a.Rotation && g.RoadID == a.RoadID && g.RoadObjectID == a.RoadObjectID {
		return
	}
	s.placeRoadGhost(c, a)
}

// UpdateRoadGhost is onToolUpdateSingle for a road without a junction under
// the cursor, as UpdateTrackGhost: the piece roadID in rotation rot at the
// tile's position height (piecePositionHeight over the road piece's tiles),
// unchanged from the last position nothing, else placed as a ghost and,
// refused, once more 16 higher. PARTIAL: tryMakeRoadJunctionAtLoc (building
// off an existing road), shift's deeper search.
func (s *State) UpdateRoadGhost(c *Cmd, tx, ty int, roadID, rot, objID uint8) (int32, bool) {
	if int(roadID) >= len(RoadPieces) {
		s.RemoveRoadGhost(c)
		return 0, false
	}
	x, y, h, ok := s.piecePositionHeight(tx, ty, RoadPieces[roadID], rot)
	if !ok {
		s.RemoveRoadGhost(c)
		return 0, false
	}
	pos := [3]int32{x, y, h}
	if s.RoadGhost != nil && s.ghostPos == pos {
		return s.ghostCost, true
	}
	s.ghostPos = pos
	a := roadPlacementArgs{X: x, Y: y, Z: h, Rotation: rot, RoadID: roadID, RoadObjectID: objID, Bridge: 0xFF}
	for tries := 2; ; {
		cost, ok := s.placeRoadGhost(c, a)
		if ok {
			s.ghostCost = cost
			return cost, true
		}
		tries--
		if tries == 0 {
			return 0, false
		}
		if a.Z -= 16; a.Z < 0 {
			return 0, false
		}
		a.Z += 32
	}
}

// BuildRoadAtTile is onToolUpSingle + constructionLoop for a road, as
// BuildTrackAtTile: the head set at the tile at the road piece's position
// height and the piece built there (constructRoad); refused, once more 16
// higher unless only above ground is allowed, then the rotation turns 90°.
func (s *State) BuildRoadAtTile(c *Cmd, tx, ty, roadID int, objID uint8) (int32, bool) {
	rot := s.ConsHead.Rotation & 3
	if roadID < 0 || roadID >= len(RoadPieces) {
		return 0, false
	}
	_, _, h, ok := s.piecePositionHeight(tx, ty, RoadPieces[roadID], rot)
	if !ok {
		return 0, false
	}
	h = max(0, h) // onToolUpSingle: "Height should never go negative"
	for tries := 2; ; {
		s.ConsHead = ConstructionHead{Active: true, X: tx, Y: ty, BaseZ: uint8(h / 4), Rotation: rot}
		if cost, ok := s.BuildRoadAtHead(c, roadID, objID); ok {
			return cost, true
		}
		if c.ErrorText == i18n.ErrorCanOnlyBuildAboveGround {
			break
		}
		if tries--; tries == 0 {
			break
		}
		if h -= 16; h < 0 {
			break
		}
		h += 32
	}
	s.ConsHead.Rotation = (rot + 1) & 3
	return 0, false
}
