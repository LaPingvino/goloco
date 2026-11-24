package sim

import (
	"github.com/LaPingvino/goloco/pkg/construct"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
)

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
// the state's (trackCost), and the main view follows it
// (ghostVisibility). Refused with bridge_type_unsuitable_for_this_configuration
// just after a piece button (byte_113603A), the window takes the first
// bridge of its list that suits the piece (none: the selected one comes
// first, nothing changes) and the ghost is tried again — upstream retries
// with the old args (placeTrackGhost(args), not newArgs), so that try fails
// too and the next update places the ghost with the new bridge.
func (s *State) placeTrackGhost(c *Cmd, a TrackPlacement) (int32, bool) {
	s.RemoveTrackGhost(c)
	cost, ok := s.DoCommandFor(c, CmdApply|CmdPreventBuildingClearing|CmdNoErrorWindow|CmdNoPayment|CmdGhost, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.CreateTrack(c, f, a)
	})
	s.Cons.Cost = 0
	if !ok && s.retryBridge(c, TrackMiscFlags[a.TrackID]) {
		return s.placeTrackGhost(c, a)
	}
	if ok {
		s.Cons.Cost = cost
		s.TrackGhost = &TrackGhost{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, TrackID: a.TrackID, TrackObjectID: a.TrackObjectID}
		s.ghostVisibility(c)
	}
	s.Cons.pieceChanged = 0
	return cost, ok
}

// retryBridge is placeTrackGhost's / placeRoadGhost's bridge retry: with
// byte_113603A set and the ghost refused for its bridge, byte_113603A is
// cleared and the first bridge of the list whose disabledTrackCfg allows
// the piece (flags: its CommonTraitFlags) becomes the window's — unless it
// is the selected one already. True: a bridge was chosen, try again.
func (s *State) retryBridge(c *Cmd, flags uint16) bool {
	cs := &s.Cons
	if cs.pieceChanged == 0 || c.ErrorText != i18n.BridgeTypeUnsuitableForThisConfiguration {
		return false
	}
	cs.pieceChanged = 0
	for _, b := range cs.BridgeList {
		if b == 0 {
			break
		}
		if o := c.bridge(b - 1); o == nil || o.DisabledTrackCfg&flags != 0 {
			continue
		}
		if b == cs.Bridge {
			break
		}
		cs.Bridge = b
		return true
	}
	return false
}

// ghostVisibility is what placeTrackGhost / placeRoadGhost do to the main
// view once the ghost is placed (WindowManager::viewportSetVisibility): the
// underground view while some of the ghost is under the ground
// (flags_1136072's underground), the overground view otherwise; the height
// marks on tracks with a gradient selected.
func (s *State) ghostVisibility(c *Cmd) {
	if s.GhostVisibility != nil {
		s.GhostVisibility(c.PosFlags&PosUnderground != 0, s.Cons.Gradient != 0)
	}
}

// placeRoadGhost is placeRoadGhost (0x0049FC60), as placeTrackGhost through
// createRoad.
func (s *State) placeRoadGhost(c *Cmd, a RoadPlacement) (int32, bool) {
	s.RemoveRoadGhost(c)
	cost, ok := s.DoCommandFor(c, CmdApply|CmdPreventBuildingClearing|CmdNoErrorWindow|CmdNoPayment|CmdGhost, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.CreateRoad(c, f, a)
	})
	s.Cons.Cost = 0
	if !ok && s.retryBridge(c, RoadMiscFlags[a.RoadID]) {
		return s.placeRoadGhost(c, a)
	}
	if ok {
		s.Cons.Cost = cost
		s.RoadGhost = &RoadGhost{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, RoadID: a.RoadID, RoadObjectID: a.RoadObjectID}
		s.ghostVisibility(c)
	}
	s.Cons.pieceChanged = 0
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

// ConstructionSelection is the map selection's free-form tiles
// (_mapSelectedFreeFormTiles, world x and y) with
// MapSelectionFlags::enableConstruct (On): the tiles under the construction
// tool's piece, outlined on the map (PaintSurface.cpp).
type ConstructionSelection struct {
	On bool
	// Unk03 is MapSelectionFlags::unk_03: the outline in ExtColour::unk2B
	// (building on), not unk25 (the tool's, hovering).
	Unk03 bool
	Tiles [][2]int32
}

// setSelectedTilesFromPiece is setMapSelectedTilesFromPiece: the piece's
// tiles laid at origin in the rotation, its diagonal tiles left out.
func (s *State) setSelectedTilesFromPiece(pieces []PreviewTrack, x, y int32, rotation uint8) {
	s.ConsSelection.Tiles = s.ConsSelection.Tiles[:0]
	for _, p := range pieces {
		if p.Flags&(1<<7) != 0 { // PreviewTrackFlags::diagonal
			continue
		}
		dx, dy := rotOffset(p.X, p.Y, rotation)
		s.ConsSelection.Tiles = append(s.ConsSelection.Tiles, [2]int32{x + int32(dx), y + int32(dy)})
	}
}

// maxSelectionHeight is getMaxConstructHeightFromExistingSelection: the
// highest construction height under the selected tiles, water counting 16
// higher; 0 with no selection.
func (s *State) maxSelectionHeight() int32 {
	var maxH int32
	if !s.ConsSelection.On {
		return 0
	}
	for _, t := range s.ConsSelection.Tiles {
		if !s.validCoords(t[0], t[1]) {
			continue
		}
		if l, w, ok := s.constructionHeight(int(t[0]/32), int(t[1]/32)); ok {
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

// shiftRetries is maxRetries with shift held (0x80000008): the search goes
// down 16 at a time until below zero.
const shiftRetries uint32 = 0x80000008

// ConstructionToolUpdate is onToolUpdateSingle (0x004A1968: the Construction
// tab's tool moved over the map while hovering), for a track or a road. The
// cursor is over the tile (tx, ty): with junction >= 0 on existing track /
// road at that height (tryMakeTrackJunctionAtLoc / tryMakeRoadJunctionAtLoc),
// else on the ground (getConstructionPos: its land height, at least its
// water's). The arrow stands there in the construction rotation and the
// tile is selected (enableConstruct); with a piece selected (none: the
// ghosts go) its tiles are, and it starts at that height — over the ground
// at least as high as the highest selected tile — less the piece's lowest
// point; the arrow moves there, and constructionGhostLoop puts the ghost 16
// below it: the state's position follows; with a ghost standing and the
// position unchanged nothing happens; else the ghost is tried there and,
// refused, once more 16 higher. With shift held the ghost starts 16 lower
// still (the arrow too) and the tries go down, 16 at a time, to zero.
func (s *State) ConstructionToolUpdate(c *Cmd, tx, ty int, junction int32, road bool, objID uint8, shift bool) {
	cs := &s.Cons
	s.ConsArrow.On, s.ConsSelection.On = false, false
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
	s.ConsSelection.On, s.ConsSelection.Unk03 = true, false
	s.ConsSelection.Tiles = append(s.ConsSelection.Tiles[:0], [2]int32{x, y})
	_, pieces, ok := s.selectedPiece(road)
	if !ok {
		s.RemoveConstructionGhosts(c)
		return
	}
	s.setSelectedTilesFromPiece(pieces, x, y, cs.Rotation)
	if junction < 0 {
		h = max(s.maxSelectionHeight(), h)
	}
	h -= minPieceZ(pieces)
	s.ConsArrow.Z = h
	h -= 16
	retries := uint32(2)
	if shift {
		retries = shiftRetries
		h -= 16
		s.ConsArrow.Z = h
	}
	s.constructionGhostLoop(c, x, y, h, retries, road, objID, shift)
}

// constructionGhostLoop is constructionGhostLoop: the state's position is
// (x, y, z); with a ghost standing there already nothing happens; else the
// ghost is placed there (the bridge it meets taken, sub_4A193B) and, refused
// and tries left, 16 higher — with shift, 16 lower — until one stands or
// the height would go below zero.
func (s *State) constructionGhostLoop(c *Cmd, x, y, z int32, retries uint32, road bool, objID uint8, shift bool) {
	cs := &s.Cons
	cs.X, cs.Y, cs.Z = x, y, z
	pos := [3]int32{x, y, z}
	if (s.TrackGhost != nil || s.RoadGhost != nil) && s.ghostPos == pos {
		return
	}
	s.ghostPos = pos
	var ra RoadPlacement
	var ta TrackPlacement
	var ok bool
	if road {
		ra, ok = s.roadPlacement(objID)
		s.RemoveTrackGhost(c)
	} else {
		ta, ok = s.trackPlacement(objID)
		s.RemoveRoadGhost(c)
	}
	if !ok {
		return
	}
	for {
		var placed bool
		if road {
			_, placed = s.placeRoadGhost(c, ra)
		} else {
			_, placed = s.placeTrackGhost(c, ta)
		}
		cs.ghostFlags = c.Flags73
		s.adoptBridge(c)
		if !placed {
			if retries--; retries != 0 {
				ta.Z -= 16
				ra.Z -= 16
				if (road && ra.Z >= 0) || (!road && ta.Z >= 0) {
					if !shift {
						ta.Z += 32
						ra.Z += 32
					}
					continue
				}
			}
		}
		s.ActivateSelection()
		return
	}
}

// ConstructionToolUp is onToolUpSingle (the click on the map while
// hovering): the ghosts removed; with a piece selected that fits the
// rotation, building starts at the tile (tx, ty) through ConstructionLoop —
// on existing track / road (junction >= 0) at exactly that height, one try;
// else (or with shift held) at the height the tool's ghost stood at: the
// highest tile the tool selected (getMaxConstructHeightFromExistingSelection)
// or the ground's — land no higher than that, water at least —, less the
// piece's lowest point, 16 down (shift: 32, and the tries go down to zero),
// not below zero, two tries. The selection and the arrow go, and the tool
// (toolCancel); a refusal brings it back (ConstructionLoop). built: a piece
// stands and the state has moved past it; refused: the command said no
// (c.ErrorText) and the state hovers again.
func (s *State) ConstructionToolUp(c *Cmd, tx, ty int, junction int32, road bool, objID uint8, shift bool) (cost int32, built, refused bool) {
	s.RemoveConstructionGhosts(c)
	_, pieces, ok := s.selectedPiece(road)
	if !ok {
		return 0, false, false
	}
	h := s.maxSelectionHeight()
	maxH := h // cState.constructionMaxHeight
	s.ConsArrow.On, s.ConsSelection.On = false, false
	if junction < 0 {
		// getConstructionPos(x, y, constructionMaxHeight)
		land, water, ok := s.constructionHeight(tx, ty)
		if !ok {
			return 0, false, false
		}
		h = max(h, max(min(land, maxH), water))
	}
	var retries uint32
	if shift || junction < 0 {
		h -= minPieceZ(pieces) + 16
		retries = 2
		if shift {
			retries = shiftRetries
			h -= 16
		}
	} else {
		retries, h = 1, junction
	}
	h = max(0, h) // "Height should never go negative"
	cost, built = s.ConstructionLoop(c, int32(tx)*32, int32(ty)*32, h, retries, road, objID, shift)
	return cost, built, !built
}

// AirportToolTiles is getAirportPlacementArgs' stationMinPos …
// stationMaxPos: the airport's extents (getAirportExtents) round the tile at
// world (x, y) in the rotation, in tiles; ok false without the object.
func AirportToolTiles(om *objects.ObjectManager, typ uint8, x, y int32, rotation uint8) (x0, y0, x1, y1 int, ok bool) {
	if om == nil || int(typ) >= len(om.AirportObjects) || om.AirportObjects[typ] == nil {
		return 0, 0, 0, 0, false
	}
	x0, y0, x1, y1 = airportExtents(om.AirportObjects[typ], int(x)>>5, int(y)>>5, rotation)
	return x0, y0, x1, y1, true
}

// DockToolTiles is getDockPlacementArgs' stationMinPos … stationMaxPos: the
// tile at world (x, y) and the one past it in x and y — the dock's 2 × 2.
func DockToolTiles(x, y int32) (x0, y0, x1, y1 int) {
	return int(x) >> 5, int(y) >> 5, int(x)>>5 + 1, int(y)>>5 + 1
}

// SetStationToolSelection is onToolUpdateAirport / onToolUpdateDock once the
// placement is found (StationTab.cpp 0x004A4F3B, 0x004A5158):
// enableConstruct and enableConstructionArrow on, unk_03 off, the arrow at
// the placement in its rotation (setConstructionArrow) and the selected
// tiles the station's tiles x0, y0 … x1, y1 (setMapSelectedTilesFromRange
// of getClampedRange: each corner clamped to the map).
func (s *State) SetStationToolSelection(x0, y0, x1, y1 int, arrow ConstructionArrow) {
	clamp := func(v, n int) int { return min(max(v, 0), n-1) } // clampTileCoord
	x0, x1 = clamp(x0, s.Width), clamp(x1, s.Width)
	y0, y1 = clamp(y0, s.Height), clamp(y1, s.Height)
	arrow.On = true
	s.ConsArrow = arrow
	s.ConsSelection.On, s.ConsSelection.Unk03 = true, false
	s.ConsSelection.Tiles = s.ConsSelection.Tiles[:0]
	for ty := y0; ty <= y1; ty++ { // TilePosRangeView: x fastest, both ends included
		for tx := x0; tx <= x1; tx++ {
			s.ConsSelection.Tiles = append(s.ConsSelection.Tiles, [2]int32{int32(tx) * 32, int32(ty) * 32})
		}
	}
}

// ClearStationToolSelection is onToolUpdateAirport / onToolUpdateDock's
// first step: resetMapSelectionFlag(enable | enableConstruct |
// enableConstructionArrow).
func (s *State) ClearStationToolSelection() {
	s.ConsArrow.On, s.ConsSelection.On = false, false
}
