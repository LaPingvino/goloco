package sim

import (
	"log"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Track construction: an OpenLoco-style construction head that pieces connect
// to, instead of free placement under the cursor.
//
// The head is the world position, height and heading where the NEXT piece
// begins. Placing a piece appends its elements to every sequence tile and
// advances the head by the piece's end transform (TrackCoordinates).
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/Construction.cpp
// and Map/Track/TrackData.cpp.

// ConstructionHead is the "next piece starts here" cursor.
type ConstructionHead struct {
	Active   bool
	X, Y     int   // tile coordinates of the piece origin
	BaseZ    uint8 // SmallZ height of the piece origin
	Rotation uint8 // heading 0-3 (diagonal headings 12-15 not yet buildable)
}

// placedPiece remembers one build action so it can be undone.
type placedPiece struct {
	trackID  int
	rotation uint8
	tiles    [][2]int // tile coords that received elements
	BaseZ    []uint8  // element BaseZ per tile
}

// rotOffset rotates a rotation-0 world offset to the given rotation.
// Matches OpenLoco Math::Vector::rotate.
func rotOffset(x, y int16, rotation uint8) (int16, int16) {
	switch rotation & 3 {
	case 1:
		return y, -x
	case 2:
		return -x, -y
	case 3:
		return -y, x
	}
	return x, y
}

// SetConstructionHead places the head on a tile (click-to-start). The head
// takes the tile's surface height and keeps its current heading.
func (s *State) SetConstructionHead(tileX, tileY int) {
	t := s.TileAt(tileX, tileY)
	if t == nil {
		return
	}
	s.ConsHead.Active = true
	s.ConsHead.X = tileX
	s.ConsHead.Y = tileY
	s.ConsHead.BaseZ = t.BaseZ
}

// ConstructionHeadState returns the current head (value copy).
func (s *State) ConstructionHeadState() ConstructionHead { return s.ConsHead }

// RotateConstructionHead turns the head 90° clockwise.
func (s *State) RotateConstructionHead() {
	s.ConsHead.Rotation = (s.ConsHead.Rotation + 1) & 3
}

// ClearConstructionHead deactivates the head (leave build mode).
func (s *State) ClearConstructionHead() { s.ConsHead.Active = false }

// CanPlaceTrackAtHead reports whether trackID fits on the map at the head.
// Which piece goes with the heading is getTrackPieceId's (pkg/construct);
// upstream's CreateTrack doesn't compare the piece's rotationBegin with the
// heading (rotationBegin serves following connected track, Track.cpp), so
// neither does this. (An earlier goloco check did, and refused e.g. the
// turnaround's track 42 where getTrackPieceId gives it.)
func (s *State) CanPlaceTrackAtHead(trackID int) bool {
	h := s.ConsHead
	if !h.Active || trackID < 0 || trackID<<3 >= len(TrackCoordinates) {
		return false
	}
	_, _, ok := s.TrackFootprint(trackID)
	return ok
}

// TrackFootprint returns the sequence-tile placements for trackID at the
// head: tile coords and BaseZ per sequence tile. ok=false when the piece
// would leave the map or the data table has no entry.
func (s *State) TrackFootprint(trackID int) (tiles [][2]int, baseZ []uint8, ok bool) {
	if trackID < 0 || trackID >= len(TrackPieceTiles) || len(TrackPieceTiles[trackID]) == 0 {
		return nil, nil, false
	}
	h := s.ConsHead
	for _, off := range TrackPieceTiles[trackID] {
		dx, dy := rotOffset(off[0], off[1], h.Rotation)
		tx := h.X + int(dx)/32
		ty := h.Y + int(dy)/32
		if tx < 0 || ty < 0 || tx >= s.Width || ty >= s.Height {
			return nil, nil, false
		}
		z := int(h.BaseZ) + int(off[2])/4 // piece z is world units; BaseZ is SmallZ
		if z < 0 || z > 255 {
			return nil, nil, false
		}
		tiles = append(tiles, [2]int{tx, ty})
		baseZ = append(baseZ, uint8(z))
	}
	return tiles, baseZ, true
}

// PlaceTrackAtHead appends trackID at the construction head and advances the
// head along the piece's end transform. Returns false if the piece cannot be
// placed (inactive head, unsupported heading, off-map).
func (s *State) PlaceTrackAtHead(trackID int, trackObjID uint8) bool {
	h := s.ConsHead
	if !s.CanPlaceTrackAtHead(trackID) {
		return false
	}
	tiles, baseZ, ok := s.TrackFootprint(trackID)
	if !ok {
		return false
	}
	_ = h

	placed := placedPiece{trackID: trackID, rotation: s.ConsHead.Rotation, tiles: tiles, BaseZ: baseZ}
	for i, tc := range tiles {
		s.AddTrack(tc[0], tc[1], scenario.TrackElement{
			TrackObjectID: trackObjID,
			TrackID:       uint8(trackID),
			Rotation:      s.ConsHead.Rotation & 3,
			SequenceIndex: uint8(i),
			BaseZ:         baseZ[i],
			ClearZ:        baseZ[i] + 8,
			IsLast:        i == len(tiles)-1, // CreateTrack: setFlag6 on the piece's last tile
		})
	}
	s.consStack = append(s.consStack, placed)
	log.Printf("[Cons] placed track id=%d objSlot=%d rot=%d tiles=%v z=%v", trackID, trackObjID, s.ConsHead.Rotation&3, tiles, baseZ)

	// Advance the head: trackAndDirection = trackID<<3 | rotation (forward).
	c := TrackCoordinates[trackID<<3|int(s.ConsHead.Rotation&3)]
	s.ConsHead.X += int(c.DX) / 32
	s.ConsHead.Y += int(c.DY) / 32
	s.ConsHead.BaseZ = uint8(int(s.ConsHead.BaseZ) + int(c.DZ)/4)
	s.ConsHead.Rotation = c.RotEnd // may become diagonal (12-15); UI restricts pieces then
	return true
}

// UndoLastTrack removes the most recently placed piece and moves the head
// back to its origin.
func (s *State) UndoLastTrack() bool {
	if len(s.consStack) == 0 {
		return false
	}
	p := s.consStack[len(s.consStack)-1]
	s.consStack = s.consStack[:len(s.consStack)-1]
	for i, tc := range p.tiles {
		s.removeTrackElement(tc[0], tc[1], uint8(p.trackID), p.rotation, uint8(i), p.BaseZ[i])
	}
	s.ConsHead.X = p.tiles[0][0]
	s.ConsHead.Y = p.tiles[0][1]
	s.ConsHead.BaseZ = p.BaseZ[0]
	s.ConsHead.Rotation = p.rotation
	s.ConsHead.Active = true
	return true
}

func (s *State) removeTrackElement(x, y int, trackID, rotation, seq, baseZ uint8) {
	t := s.TileAt(x, y)
	if t == nil {
		return
	}
	for i, te := range t.Tracks {
		if te.TrackID == trackID && te.Rotation == rotation &&
			te.SequenceIndex == seq && te.BaseZ == baseZ {
			t.Tracks = append(t.Tracks[:i], t.Tracks[i+1:]...)
			t.removeRef(scenario.KindTrack, i)
			return
		}
	}
}

// --- Road construction (mirrors the track functions above) ----------------
//
// Roads reuse the same construction head. The 80-entry RoadCoordinates table
// and RoadPieceTiles footprints are generated from the same upstream file.
// Roads never produce diagonal end rotations, so the head stays cardinal.

// CanPlaceRoadAtHead reports whether roadID can start at the head's exact
// heading: the piece's begin rotation (RoadCoordinates) must equal the head
// heading.
func (s *State) CanPlaceRoadAtHead(roadID int) bool {
	h := s.ConsHead
	if !h.Active || roadID < 0 || roadID<<3 >= len(RoadCoordinates) {
		return false
	}
	// No rotationBegin check: CreateRoad has none (see CanPlaceTrackAtHead).
	_, _, ok := s.RoadFootprint(roadID)
	return ok
}

// RoadFootprint returns the sequence-tile placements for roadID at the head:
// tile coords and BaseZ per sequence tile. ok=false when the piece would leave
// the map or the data table has no entry.
func (s *State) RoadFootprint(roadID int) (tiles [][2]int, baseZ []uint8, ok bool) {
	if roadID < 0 || roadID >= len(RoadPieceTiles) || len(RoadPieceTiles[roadID]) == 0 {
		return nil, nil, false
	}
	h := s.ConsHead
	for _, off := range RoadPieceTiles[roadID] {
		dx, dy := rotOffset(off[0], off[1], h.Rotation)
		tx := h.X + int(dx)/32
		ty := h.Y + int(dy)/32
		if tx < 0 || ty < 0 || tx >= s.Width || ty >= s.Height {
			return nil, nil, false
		}
		z := int(h.BaseZ) + int(off[2])/4 // piece z is world units; BaseZ is SmallZ
		if z < 0 || z > 255 {
			return nil, nil, false
		}
		tiles = append(tiles, [2]int{tx, ty})
		baseZ = append(baseZ, uint8(z))
	}
	return tiles, baseZ, true
}

// PlaceRoadAtHead appends roadID at the construction head and advances the head
// along the piece's end transform. Returns false if the piece cannot be placed.
func (s *State) PlaceRoadAtHead(roadID int, roadObjID uint8) bool {
	if !s.CanPlaceRoadAtHead(roadID) {
		return false
	}
	tiles, baseZ, ok := s.RoadFootprint(roadID)
	if !ok {
		return false
	}

	placed := placedPiece{trackID: roadID, rotation: s.ConsHead.Rotation, tiles: tiles, BaseZ: baseZ}
	for i, tc := range tiles {
		s.AddRoad(tc[0], tc[1], scenario.RoadElement{
			RoadObjectID:  roadObjID,
			RoadID:        uint8(roadID),
			Rotation:      s.ConsHead.Rotation & 3,
			SequenceIndex: uint8(i),
			BaseZ:         baseZ[i],
			ClearZ:        baseZ[i] + 8,
			IsLast:        i == len(tiles)-1, // CreateRoad: setFlag6 on the piece's last tile
		})
	}
	s.consRoadStack = append(s.consRoadStack, placed)

	// Advance the head: roadAndDirection = roadID<<3 | rotation (forward).
	c := RoadCoordinates[roadID<<3|int(s.ConsHead.Rotation&3)]
	s.ConsHead.X += int(c.DX) / 32
	s.ConsHead.Y += int(c.DY) / 32
	s.ConsHead.BaseZ = uint8(int(s.ConsHead.BaseZ) + int(c.DZ)/4)
	s.ConsHead.Rotation = c.RotEnd
	return true
}

// UndoLastRoad removes the most recently placed road piece and moves the head
// back to its origin.
func (s *State) UndoLastRoad() bool {
	if len(s.consRoadStack) == 0 {
		return false
	}
	p := s.consRoadStack[len(s.consRoadStack)-1]
	s.consRoadStack = s.consRoadStack[:len(s.consRoadStack)-1]
	for i, tc := range p.tiles {
		s.removeRoadElement(tc[0], tc[1], uint8(p.trackID), p.rotation, uint8(i), p.BaseZ[i])
	}
	s.ConsHead.X = p.tiles[0][0]
	s.ConsHead.Y = p.tiles[0][1]
	s.ConsHead.BaseZ = p.BaseZ[0]
	s.ConsHead.Rotation = p.rotation
	s.ConsHead.Active = true
	return true
}

// removeRoadElement deletes one exactly-matching road element from a tile.
func (s *State) removeRoadElement(x, y int, roadID, rotation, seq, baseZ uint8) {
	t := s.TileAt(x, y)
	if t == nil {
		return
	}
	for i, re := range t.Roads {
		if re.RoadID == roadID && re.Rotation == rotation &&
			re.SequenceIndex == seq && re.BaseZ == baseZ {
			t.Roads = append(t.Roads[:i], t.Roads[i+1:]...)
			t.removeRef(scenario.KindRoad, i)
			return
		}
	}
}

// DrawRoadConstructionGhost renders the selected road piece translucently at
// the head, covering every sequence tile. Uses the Style0 single-image-per-
// rotation offsets (roads 0-2 base merge sprites are valid direct images).

// BuildTrackAtHead builds trackID at the construction head the way the
// construction window does (Construction/ConstructionTab constructTrack): the
// createTrack game command for the player's company — checked, then charged
// (Construction) — and on success the head moves on along the piece. The
// cost; false when the command refuses (no money, in the way, too high …).
func (s *State) BuildTrackAtHead(c *Cmd, trackID int, trackObjID uint8) (int32, bool) {
	h := s.ConsHead
	if !h.Active || trackID < 0 || trackID >= len(TrackPieces) || h.Rotation > 3 {
		return 0, false
	}
	// constructTrack: removeConstructionGhosts first (the ghost stands where
	// the piece goes).
	s.RemoveTrackGhost(c)
	a := headTrackArgs(h, trackID, trackObjID, s.bridge()) // getTrackPlacementArgs: lastSelectedBridge
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	var placed []trackPiecePlan
	cost, ok := s.DoCommandFor(c, CmdApply, ExpenditureConstruction, func(flags uint8) (int32, bool) {
		if flags&CmdApply == 0 {
			return s.createTrack(c, a, flags)
		}
		r, ok := s.createTrackCheck(c, a, flags)
		if !ok || !s.createTrackApply(c, a, flags, &r) {
			return 0, false
		}
		placed = r.Pieces
		return r.Cost, true
	})
	if !ok {
		return 0, false
	}
	p := placedPiece{trackID: trackID, rotation: h.Rotation}
	for _, pl := range placed {
		p.tiles = append(p.tiles, [2]int{int(pl.X / 32), int(pl.Y / 32)})
		p.BaseZ = append(p.BaseZ, pl.BaseZ)
	}
	s.consStack = append(s.consStack, p)
	tc := TrackCoordinates[trackID<<3|int(h.Rotation&3)]
	s.ConsHead.X += int(tc.DX) / 32
	s.ConsHead.Y += int(tc.DY) / 32
	s.ConsHead.BaseZ = uint8(int(s.ConsHead.BaseZ) + int(tc.DZ)/4)
	s.ConsHead.Rotation = tc.RotEnd
	return cost, true
}
