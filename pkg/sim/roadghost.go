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
