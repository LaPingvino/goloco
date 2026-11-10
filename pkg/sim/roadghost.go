package sim

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
