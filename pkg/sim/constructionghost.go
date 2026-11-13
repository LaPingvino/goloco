package sim

import "github.com/LaPingvino/goloco/pkg/construct"

// The construction window's ghost and its tool on the map (ConstructionTab
// onToolUpdateSingle 0x004A1968, constructionGhostLoop, onToolUpSingle
// 0x0049DC97's single part, placeTrackGhost 0x0049FB63, placeRoadGhost and
// removeTrackGhosts 0x004A006C): while hovering, the selected piece stands
// as a ghost under the cursor in the construction rotation, taken up again
// (as a ghost removal) before the next one; a click starts building there.

// TrackGhost is the track ghost on the map (ghostRemovalTrack*), nil none.
type TrackGhost struct {
	X, Y, Z       int32
	Rotation      uint8
	TrackID       uint8
	TrackObjectID uint8
}

// RoadGhost is the road ghost on the map (ghostRemovalTrack* for a road), nil none.
type RoadGhost struct {
	X, Y, Z      int32
	Rotation     uint8
	RoadID       uint8
	RoadObjectID uint8
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

// RemoveTrackGhost is removeTrackGhosts for a track ghost: TrackRemoval at
// the ghost's first tile (its height plus the piece's first z, index 0) with
// apply, no error window, no payment, ghost; then forgotten. (The piece's
// first z was left out before: a ghost that starts above its position — a
// slope going down — was never found and stayed on the map.)
func (s *State) RemoveTrackGhost(c *Cmd) {
	g := s.TrackGhost
	if g == nil {
		return
	}
	s.TrackGhost = nil
	z := g.Z
	if int(g.TrackID) < len(TrackPieces) && len(TrackPieces[g.TrackID]) > 0 {
		z += int32(TrackPieces[g.TrackID][0].Z)
	}
	s.RemoveTrack(c, CmdApply|CmdNoErrorWindow|CmdNoPayment|CmdGhost, TrackRemoval{X: g.X, Y: g.Y, Z: z, Rotation: g.Rotation & 3, TrackID: g.TrackID, Index: 0, TrackObjectID: g.TrackObjectID}, nil)
}

// RemoveRoadGhost is removeTrackGhosts for a road ghost, as RemoveTrackGhost.
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

// placeTrackGhost is placeTrackGhost: the old ghost removed, then
// createTrack with apply, prevent building clearing, no error window, no
// payment and ghost; placed, it is remembered for removal and its cost is
// the state's (trackCost). PARTIAL: the
// bridge retry (bridge_type_unsuitable…: the next bridge in the list) and
// the viewport's underground view.
func (s *State) placeTrackGhost(c *Cmd, a trackPlacementArgs) (int32, bool) {
	s.RemoveTrackGhost(c)
	cost, ok := s.DoCommandFor(c, CmdApply|CmdPreventBuildingClearing|CmdNoErrorWindow|CmdNoPayment|CmdGhost, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.createTrack(c, a, f)
	})
	s.Cons.Cost = 0
	if ok {
		s.Cons.Cost = cost
		s.TrackGhost = &TrackGhost{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, TrackID: a.TrackID, TrackObjectID: a.TrackObjectID}
	}
	return cost, ok
}

// placeRoadGhost is placeRoadGhost, as placeTrackGhost through createRoad.
func (s *State) placeRoadGhost(c *Cmd, a roadPlacementArgs) (int32, bool) {
	s.RemoveRoadGhost(c)
	cost, ok := s.DoCommandFor(c, CmdApply|CmdPreventBuildingClearing|CmdNoErrorWindow|CmdNoPayment|CmdGhost, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.createRoad(c, a, f)
	})
	s.Cons.Cost = 0
	if ok {
		s.Cons.Cost = cost
		s.RoadGhost = &RoadGhost{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, RoadID: a.RoadID, RoadObjectID: a.RoadObjectID}
	}
	return cost, ok
}

// selectedPiece is the tool's getPieceId / getPiece: the selected piece's id
// in the construction rotation and its tiles, a track's or a road's.
func (s *State) selectedPiece(road bool) (construct.TrackPieceID, []PreviewTrack, bool) {
	cs := &s.Cons
	if road {
		p, ok := construct.GetRoadPieceID(cs.Piece, cs.Gradient, cs.Rotation)
		if !ok || int(p.ID) >= len(RoadPieces) {
			return p, nil, false
		}
		return p, RoadPieces[p.ID], true
	}
	p, ok := construct.GetTrackPieceID(cs.Piece, cs.Gradient, cs.Rotation)
	if !ok || int(p.ID) >= len(TrackPieces) {
		return p, nil, false
	}
	return p, TrackPieces[p.ID], true
}

// maxSelectionHeight is getMaxConstructHeightFromExistingSelection over
// setMapSelectedTilesFromPiece's tiles: the highest construction height
// under the piece laid at the tile in the construction rotation (its
// diagonal tiles left out), water counting 16 higher.
func (s *State) maxSelectionHeight(tx, ty int, pieces []PreviewTrack) int32 {
	var maxH int32
	for _, p := range pieces {
		if p.Flags&(1<<7) != 0 { // PreviewTrackFlags::diagonal: not selected
			continue
		}
		dx, dy := rotOffset(p.X, p.Y, s.Cons.Rotation)
		if l, w, ok := s.constructionHeight(tx+int(dx)/32, ty+int(dy)/32); ok {
			maxH = max(maxH, l)
			if w != 0 {
				maxH = max(maxH, w+16)
			}
		}
	}
	return maxH
}

// minPieceZ is getMinPieceHeight: the piece's lowest point.
func minPieceZ(pieces []PreviewTrack) int32 {
	m := int32(pieces[0].Z)
	for _, p := range pieces {
		m = min(m, int32(p.Z))
	}
	return m
}

// ConstructionToolUpdate is onToolUpdateSingle (the Construction tab's tool
// moved over the map while hovering), for a track or a road. The cursor is
// over the tile (tx, ty): with junction >= 0 on existing track / road at
// that height (tryMakeTrackJunctionAtLoc / tryMakeRoadJunctionAtLoc), else
// on the ground (getConstructionPos: its land height, at least its water's).
// The arrow stands there in the construction rotation; the selected piece
// (none: the ghosts go) starts at that height — over the ground at least as
// high as the highest tile under it — less the piece's lowest point, and
// constructionGhostLoop puts its ghost 16 below that: the state's position
// follows; with a ghost standing and the position unchanged nothing happens;
// else the ghost is tried there and, refused, once more 16 higher.
// PARTIAL: shift's search downwards, the selected tiles' outline.
func (s *State) ConstructionToolUpdate(c *Cmd, tx, ty int, junction int32, road bool, objID uint8) {
	cs := &s.Cons
	s.ConsArrow.On = false
	land, water, ok := s.constructionHeight(tx, ty)
	if !ok {
		return
	}
	x, y := int32(tx)*32, int32(ty)*32
	h := max(land, water)
	if junction >= 0 {
		h = junction
	}
	s.ConsArrow = ConstructionArrow{On: true, X: x, Y: y, Z: h, Direction: cs.Rotation}
	_, pieces, ok := s.selectedPiece(road)
	if !ok {
		s.RemoveConstructionGhosts(c)
		return
	}
	if junction < 0 {
		h = max(s.maxSelectionHeight(tx, ty, pieces), h)
	}
	h -= minPieceZ(pieces)
	s.ConsArrow.Z = h
	h -= 16

	// constructionGhostLoop
	cs.X, cs.Y, cs.Z = x, y, h
	pos := [3]int32{x, y, h}
	if (s.TrackGhost != nil || s.RoadGhost != nil) && s.ghostPos == pos {
		return
	}
	s.ghostPos = pos
	if road {
		a, ok := s.roadPlacement(objID)
		if !ok {
			return
		}
		s.RemoveTrackGhost(c)
		for tries := 2; ; a.Z += 32 {
			if _, ok := s.placeRoadGhost(c, a); ok {
				return
			}
			if tries--; tries == 0 {
				return
			}
			if a.Z -= 16; a.Z < 0 {
				return
			}
		}
	}
	a, ok := s.trackPlacement(objID)
	if !ok {
		return
	}
	s.RemoveRoadGhost(c)
	for tries := 2; ; a.Z += 32 {
		if _, ok := s.placeTrackGhost(c, a); ok {
			return
		}
		if tries--; tries == 0 {
			return
		}
		if a.Z -= 16; a.Z < 0 {
			return
		}
	}
}

// ConstructionToolUp is onToolUpSingle (the click on the map while
// hovering): the ghosts removed; with a piece selected that fits the
// rotation, building starts at the tile (tx, ty) through ConstructionLoop —
// on existing track / road (junction >= 0) at exactly that height, one try;
// else at the height the tool's ghost stood at (the ground or the highest
// tile under the piece, less the piece's lowest point, 16 down, not below
// zero), two tries. The tool goes (toolCancel); a refusal brings it back
// (ConstructionLoop). built: a piece stands and the state has moved past it;
// refused: the command said no (c.ErrorText) and the state hovers again.
// PARTIAL: shift.
func (s *State) ConstructionToolUp(c *Cmd, tx, ty int, junction int32, road bool, objID uint8) (cost int32, built, refused bool) {
	s.RemoveConstructionGhosts(c)
	_, pieces, ok := s.selectedPiece(road)
	if !ok {
		return 0, false, false
	}
	s.ConsArrow.On = false
	land, water, ok := s.constructionHeight(tx, ty)
	if !ok {
		return 0, false, false
	}
	h, tries := junction, 1
	if junction < 0 {
		// getConstructionPos(x, y, constructionMaxHeight): the land no higher
		// than the selection's highest, the water at least; then the higher
		// of that and the selection's.
		sel := s.maxSelectionHeight(tx, ty, pieces)
		h = max(sel, max(min(land, sel), water))
		h -= minPieceZ(pieces) + 16
		tries = 2
	}
	h = max(0, h) // "Height should never go negative"
	cost, built = s.ConstructionLoop(c, int32(tx)*32, int32(ty)*32, h, tries, road, objID)
	return cost, built, !built
}
