package sim

// The construction window's track ghost while positioning (ConstructionTab
// onToolUpdateSingle 0x004A1968, constructionGhostLoop, placeTrackGhost
// 0x0049FA45 and removeTrackGhosts): the selected piece placed as a ghost
// under the cursor in the construction rotation, taken up again (as a ghost
// removal) before the next one.

// TrackGhost is the ghost on the map (ghostRemovalTrack*), nil none.
type TrackGhost struct {
	X, Y, Z       int32
	Rotation      uint8
	TrackID       uint8
	TrackObjectID uint8
}

// constructionHeight is getConstructionHeight: the tile's surface height,
// + 16 when sloped, + 16 more on a double-height slope; its water height.
func (m *Map) constructionHeight(tx, ty int) (land, water int32, ok bool) {
	t := m.TileAt(tx, ty)
	if t == nil {
		return 0, 0, false
	}
	land = int32(t.BaseZ) * smallZStep
	if t.Slope&0x0F != 0 {
		land += 16
	}
	if t.Slope&0x10 != 0 {
		land += 16
	}
	return land, int32(t.WaterLevel) * 16, true
}

// RemoveTrackGhost is removeTrackGhosts: the ghost taken up (TrackRemoval
// with apply, no error window, no payment, ghost) and forgotten.
func (s *State) RemoveTrackGhost(c *Cmd) {
	g := s.TrackGhost
	if g == nil {
		return
	}
	s.TrackGhost = nil
	s.RemoveTrack(c, CmdApply|CmdNoErrorWindow|CmdNoPayment|CmdGhost, TrackRemoval{X: g.X, Y: g.Y, Z: g.Z, Rotation: g.Rotation, TrackID: g.TrackID, TrackObjectID: g.TrackObjectID}, nil)
}

// placeTrackGhost is placeTrackGhost: the old ghost removed, then
// createTrack with apply, prevent building clearing, no error window, no
// payment and ghost; placed, it is remembered for removal. PARTIAL: the
// bridge retry (bridge_type_unsuitable…: the next bridge in the list) and
// the viewport's underground view.
func (s *State) placeTrackGhost(c *Cmd, a trackPlacementArgs) (int32, bool) {
	s.RemoveTrackGhost(c)
	cost, ok := s.DoCommandFor(c, CmdApply|CmdPreventBuildingClearing|CmdNoErrorWindow|CmdNoPayment|CmdGhost, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.createTrack(c, a, f)
	})
	if ok {
		s.TrackGhost = &TrackGhost{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, TrackID: a.TrackID, TrackObjectID: a.TrackObjectID}
	}
	return cost, ok
}

// UpdateTrackGhost is onToolUpdateSingle for track without a junction under
// the cursor: at the tile (tx, ty) — getConstructionPos: its land height,
// at least its water's — the piece trackID in rotation rot, its tiles' highest
// construction height (water + 16 over water) if higher, less the piece's
// lowest point, then 16 down; constructionGhostLoop: unchanged from the last
// position, nothing; else the ghost is tried there and, refused, once more 16
// higher (two tries; shift's deeper search is not ported). The ghost's cost
// and whether one stands. PARTIAL: tryMakeTrackJunctionAtLoc (building off
// existing track), the construction arrow and the selected tiles drawn.
func (s *State) UpdateTrackGhost(c *Cmd, tx, ty int, trackID, rot, objID uint8, bridge uint8) (int32, bool) {
	land, water, ok := s.constructionHeight(tx, ty)
	if !ok || int(trackID) >= len(TrackPieces) {
		s.RemoveTrackGhost(c)
		return 0, false
	}
	h := max(land, water)
	x, y := int32(tx)*32, int32(ty)*32
	pieces := TrackPieces[trackID]
	var maxH int32 // getMaxConstructHeightFromExistingSelection over setMapSelectedTilesFromPiece's tiles
	minZ := int32(pieces[0].Z)
	for _, p := range pieces {
		minZ = min(minZ, int32(p.Z))
		if p.Flags&(1<<7) != 0 { // PreviewTrackFlags::diagonal: not selected
			continue
		}
		dx, dy := rotOffset(p.X, p.Y, rot)
		px, py := int(x+int32(dx))/32, int(y+int32(dy))/32
		if l, w, ok := s.constructionHeight(px, py); ok {
			maxH = max(maxH, l)
			if w != 0 {
				maxH = max(maxH, w+16)
			}
		}
	}
	h = max(maxH, h) - minZ - 16
	pos := [3]int32{x, y, h}
	if s.TrackGhost != nil && s.ghostPos == pos {
		return s.ghostCost, true
	}
	s.ghostPos = pos
	a := trackPlacementArgs{X: x, Y: y, Z: h, Rotation: rot, TrackID: trackID, TrackObjectID: objID, Bridge: bridge}
	for tries := 2; ; {
		cost, ok := s.placeTrackGhost(c, a)
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

// UpdateHeadTrackGhost is updateConstruction's track ghost (ConstructionTab
// 0x0049FD66): with the head placed, the selected piece trackID stands as a
// ghost at the head — getTrackPlacementArgs at (cState.x, cState.y,
// constructionZ), the arguments BuildTrackAtHead builds with — through
// placeTrackGhost. Upstream places it on each update while
// GhostVisibilityFlags::track is clear (removeConstructionGhosts clears it on
// every change of head or piece); here: when the head or the piece differ from
// the ghost standing, or none stands (a refused one is tried again, as there).
// trackID < 0 (nothing buildable selected) or no head: the ghost removed.
// PARTIAL: the bridge retry (placeTrackGhost's), the cost shown.
func (s *State) UpdateHeadTrackGhost(c *Cmd, trackID int, objID uint8) {
	h := s.ConsHead
	if !h.Active || trackID < 0 || trackID >= len(TrackPieces) || h.Rotation > 3 {
		s.RemoveTrackGhost(c)
		return
	}
	a := headTrackArgs(h, trackID, objID)
	if g := s.TrackGhost; g != nil && g.X == a.X && g.Y == a.Y && g.Z == a.Z && g.Rotation == a.Rotation && g.TrackID == a.TrackID && g.TrackObjectID == a.TrackObjectID {
		return
	}
	s.placeTrackGhost(c, a)
}

// headTrackArgs is the createTrack arguments for trackID at the head.
func headTrackArgs(h ConstructionHead, trackID int, objID uint8) trackPlacementArgs {
	return trackPlacementArgs{X: int32(h.X) * 32, Y: int32(h.Y) * 32, Z: int32(h.BaseZ) * 4, Rotation: h.Rotation & 3,
		TrackID: uint8(trackID), TrackObjectID: objID, Bridge: 0xFF}
}
