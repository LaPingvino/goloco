package sim

import (
	"github.com/LaPingvino/goloco/pkg/construct"
	"github.com/LaPingvino/goloco/pkg/i18n"
)

// The construction window's state and what its Construction tab does with
// it: the position the next piece starts at, building there, taking the
// piece behind it up again, and the ghost and arrow that stand there.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/ConstructionTab.cpp
// (constructTrack 0x0049F93A, constructRoad 0x0049FA10, removeTrack
// 0x004A012E, removeRoad 0x004A02F2, onMouseUp 0x0049D3F6, constructionLoop,
// updateConstruction 0x0049FD66) and Common.cpp (removeConstructionGhosts
// 0x0049FEC7).

// ConstructionState is the ported part of Construction's ConstructionState
// (cState, Ui/Windows/Construction.h).
type ConstructionState struct {
	// X, Y, Z are cState.x, y and constructionZ (world units): where the next
	// piece starts — while hovering, where the ghost under the cursor stands.
	X, Y, Z int32
	// Rotation is constructionRotation: 0–3 straight on, 4–7 and 8–11 beside
	// the other track of a dual track, 12–15 diagonal.
	Rotation uint8
	// Hover is constructionHover: the window's tool is on the map, choosing
	// where to start; false while building on from X, Y, Z.
	Hover bool
	// Piece and Gradient are lastSelectedTrackPiece and
	// lastSelectedTrackGradient (construct.Straight …, construct.Level …).
	Piece, Gradient uint8
	// Mods is lastSelectedMods: the Overhead tab's ticked extras.
	Mods uint8
	// Bridge is lastSelectedBridge + 1 (0: none), the bridge type a piece is
	// built with where it needs one.
	Bridge uint8
	// Cost is trackCost / roadCost: what the ghost standing would cost, or
	// the piece just built did (0: none — GameCommands::kFailure).
	Cost int32

	arrowFrame uint8 // constructionArrowFrameNum
	arrowBlink bool  // GhostVisibilityFlags::constructArrow
}

// ConstructionArrow is World::setConstructionArrow's arrow with
// MapSelectionFlags::enableConstructionArrow (On).
type ConstructionArrow struct {
	On        bool
	X, Y, Z   int32
	Direction uint8
}

// kRotationOffset (Map/Tile.h): a step in each of the sixteen rotations.
var kRotationOffset = [16][2]int32{
	{-32, 0}, {0, 32}, {32, 0}, {0, -32},
	{-32, 0}, {0, 32}, {32, 0}, {0, -32},
	{-32, 0}, {0, 32}, {32, 0}, {0, -32},
	{-32, 32}, {32, 32}, {32, -32}, {-32, -32},
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

// ResetConstruction is the state a construction window opens with and leaves
// behind (Common's open, then trackWindow's rotate_90): hovering, rotation
// 0, straight and level; the bridge and mods stay the window's.
func (s *State) ResetConstruction() {
	s.Cons = ConstructionState{Hover: true, Mods: s.Cons.Mods, Bridge: s.Cons.Bridge}
	s.ghostPos = [3]int32{-1, -1, -1}
	s.ConsArrow.On = false
}

// bridge is the selected bridge type for the placement arguments (0xFF none).
func (s *State) bridge() uint8 { return s.Cons.Bridge - 1 }

// RemoveConstructionGhosts is removeConstructionGhosts' track and road part:
// the blinking arrow off, the ghost taken up.
func (s *State) RemoveConstructionGhosts(c *Cmd) {
	if s.Cons.arrowBlink {
		s.Cons.arrowBlink = false
		s.ConsArrow.On = false
	}
	s.RemoveTrackGhost(c)
	s.RemoveRoadGhost(c)
}

// trackPlacement is getTrackPlacementArgs at the state's position: the
// selected piece's track id (getTrackPieceId) with the window's mods and
// bridge; ok false when the piece can't be built in this rotation.
// PARTIAL: args.unk (byte_113607E).
func (s *State) trackPlacement(objID uint8) (trackPlacementArgs, bool) {
	cs := &s.Cons
	p, ok := construct.GetTrackPieceID(cs.Piece, cs.Gradient, cs.Rotation)
	if !ok {
		return trackPlacementArgs{}, false
	}
	return trackPlacementArgs{X: cs.X, Y: cs.Y, Z: cs.Z, Rotation: p.Rotation & 3, TrackID: p.ID,
		Mods: cs.Mods, Bridge: s.bridge(), TrackObjectID: objID}, true
}

// roadPlacement is getRoadPlacementArgs, as trackPlacement.
func (s *State) roadPlacement(objID uint8) (roadPlacementArgs, bool) {
	cs := &s.Cons
	p, ok := construct.GetRoadPieceID(cs.Piece, cs.Gradient, cs.Rotation)
	if !ok {
		return roadPlacementArgs{}, false
	}
	return roadPlacementArgs{X: cs.X, Y: cs.Y, Z: cs.Z, Rotation: p.Rotation & 3, RoadID: p.ID,
		Mods: cs.Mods, Bridge: s.bridge(), RoadObjectID: objID}, true
}

// advance moves the state past a built piece: by getUnkTrack / getUnkRoad's
// pos, turned to its rotationEnd; the ghost and arrow start afresh
// (resetGhostVisibilityFlags, constructionArrowFrameNum = 0) and a one-shot
// piece (≥ 9: s-bends, the turnaround) goes back to straight.
func (s *State) advance(tc TrackCoord) {
	cs := &s.Cons
	cs.X += int32(tc.DX)
	cs.Y += int32(tc.DY)
	cs.Z += int32(tc.DZ)
	cs.Rotation = tc.RotEnd
	cs.arrowBlink, cs.arrowFrame = false, 0
	if cs.Piece >= 9 {
		cs.Piece = construct.Straight
	}
}

// ConstructTrack is constructTrack (the Construct button on a track): the
// ghosts removed, then the selected piece built at the state's position by
// the createTrack game command — checked, and charged under Construction —
// and the state moved past it. The cost; false when no piece is selected
// that fits the rotation or the command refuses (c.ErrorText says why).
func (s *State) ConstructTrack(c *Cmd, objID uint8) (int32, bool) {
	s.Cons.Cost = 0
	s.RemoveConstructionGhosts(c)
	a, ok := s.trackPlacement(objID)
	if !ok {
		return 0, false
	}
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	cost, ok := s.DoCommandFor(c, CmdApply, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.createTrack(c, a, f)
	})
	if !ok {
		return 0, false
	}
	s.Cons.Cost = cost
	s.advance(TrackCoordinates[int(a.TrackID)<<3|int(a.Rotation)])
	return cost, true
}

// ConstructRoad is constructRoad, as ConstructTrack through createRoad.
// PARTIAL: the retry with the alternate road object (sub_4A18D4, a road of
// another any-road-type-compatible type in the way).
func (s *State) ConstructRoad(c *Cmd, objID uint8) (int32, bool) {
	s.Cons.Cost = 0
	s.RemoveConstructionGhosts(c)
	a, ok := s.roadPlacement(objID)
	if !ok {
		return 0, false
	}
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	cost, ok := s.DoCommandFor(c, CmdApply, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.createRoad(c, a, f)
	})
	if !ok {
		return 0, false
	}
	s.Cons.Cost = cost
	s.advance(RoadCoordinates[int(a.RoadID)<<3|int(a.Rotation)])
	return cost, true
}

// retreat moves the state back over a removed piece (removeTrack's /
// removeRoad's end): to its start, in its rotationBegin, straight and level
// selected.
func (s *State) retreat(tc TrackCoord) {
	cs := &s.Cons
	cs.X -= int32(tc.DX)
	cs.Y -= int32(tc.DY)
	cs.Z -= int32(tc.DZ)
	cs.Rotation = tc.RotBegin
	cs.Piece, cs.Gradient = construct.Straight, construct.Level
}

// ConstructionRemoveTrack is removeTrack (the bulldozer button on a track):
// the ghosts removed; not while hovering. The piece behind the state's
// position is looked up — a straight (or, by the rotation, the dual-track
// straights 26 / 27, or the diagonal 1 one diagonal step on) laid backwards
// from here ends on the tile getTrackConnections searches for the player's
// track of this type — and the last connection found is the piece: taken up
// by the removeTrack game command (its sell price paid back under
// Construction; g1 for a station on it, RemoveTrainStation), and the state
// moves back to where it began.
//
// removed: a piece went. refused: the command said no (c.ErrorText); both
// false: nothing behind the position (as upstream, nothing happens — the
// ghost comes back with the next update).
func (s *State) ConstructionRemoveTrack(c *Cmd, objID uint8, g1 G1Raw) (removed, refused bool) {
	s.Cons.Cost = 0
	s.RemoveConstructionGhosts(c)
	cs := &s.Cons
	if cs.Hover {
		return false, false
	}
	x, y, z := cs.X, cs.Y, cs.Z
	var tad uint16
	switch {
	case cs.Rotation < 4:
		tad = 0
	case cs.Rotation < 8:
		tad = 26 << 3
	case cs.Rotation < 12:
		tad = 27 << 3
	default:
		tad = 1 << 3
		x += kRotationOffset[cs.Rotation][0]
		y += kRotationOffset[cs.Rotation][1]
	}
	tad |= 1<<2 | uint16(cs.Rotation&3)
	end, _ := TrackConnectionEnd(x, y, z, tad)
	tc := s.TrackConnections(end.X, end.Y, end.Z, end.NextRotation, c.Company, objID, 0, 0)
	if len(tc.Connections) == 0 {
		return false, false
	}
	tad2 := tc.Connections[len(tc.Connections)-1]&TaDBasicMask ^ 1<<2
	unk := TrackCoordinates[tad2]
	piece := TrackPieces[tad2>>3]
	a := TrackRemoval{X: cs.X - int32(unk.DX), Y: cs.Y - int32(unk.DY), Z: cs.Z - int32(unk.DZ),
		Rotation: uint8(tad2 & 3), TrackID: uint8(tad2 >> 3), TrackObjectID: objID}
	i := 0
	if tad2&(1<<2) != 0 {
		a.Z += int32(unk.DZ)
		i = len(piece) - 1
	}
	a.Z += int32(piece[i].Z)
	a.Index = piece[i].Index
	if _, ok := s.DoCommandFor(c, CmdApply, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.RemoveTrack(c, f, a, g1)
	}); !ok {
		return false, true
	}
	s.retreat(unk)
	return true, false
}

// roadObjectUnk02 is RoadObjectFlags::unk_02 (a road whose two lanes are
// pieces of their own: removing takes every connection found).
const roadObjectUnk02 = 1 << 2

// ConstructionRemoveRoad is removeRoad, as ConstructionRemoveTrack: the
// road connections behind the position (a straight laid backwards), a
// straight one first; only that first one unless the road object has
// unk_02; each taken up by the removeRoad game command — a refusal stops
// there, the position unmoved — and the state moves back over the last.
func (s *State) ConstructionRemoveRoad(c *Cmd, objID uint8, g1 G1Raw) (removed, refused bool) {
	s.Cons.Cost = 0
	s.RemoveConstructionGhosts(c)
	cs := &s.Cons
	if cs.Hover {
		return false, false
	}
	end, _ := RoadConnectionEnd(cs.X, cs.Y, cs.Z, 1<<2|uint16(cs.Rotation&3))
	rc := s.RoadConnections(end.X, end.Y, end.Z, end.NextRotation, c.Company, objID, 0, 0)
	conns := rc.Connections
	if len(conns) == 0 {
		return false, false
	}
	for i := range conns {
		if conns[i]&0x1F8 == 0 { // roadId 0: a straight
			conns[i], conns[0] = conns[0], conns[i]
		}
	}
	if obj := c.roadObject(objID); obj == nil || obj.Flags&roadObjectUnk02 == 0 {
		conns = conns[:1]
	}
	var unk TrackCoord
	for _, conn := range conns {
		tad2 := conn&TaDBasicMask ^ 1<<2
		unk = RoadCoordinates[tad2]
		piece := RoadPieces[tad2>>3]
		a := RoadRemoval{X: cs.X - int32(unk.DX), Y: cs.Y - int32(unk.DY), Z: cs.Z - int32(unk.DZ),
			Rotation: uint8(tad2 & 3), RoadID: uint8(tad2 >> 3), RoadObjectID: objID}
		i := 0
		if tad2&(1<<2) != 0 {
			a.Z += int32(unk.DZ)
			i = len(piece) - 1
		}
		a.Z += int32(piece[i].Z)
		a.Index = piece[i].Index
		if _, ok := s.DoCommandFor(c, CmdApply, ExpenditureConstruction, func(f uint8) (int32, bool) {
			return s.RemoveRoad(c, f, a, g1)
		}); !ok {
			return false, true
		}
	}
	s.retreat(unk)
	return true, false
}

// ChangeTrackPiece is changeTrackPiece (a piece or gradient button): the
// ghosts removed, the piece — or, slope, the gradient — selected.
func (s *State) ChangeTrackPiece(c *Cmd, v uint8, slope bool) {
	s.RemoveConstructionGhosts(c)
	s.Cons.Cost = 0
	if slope {
		s.Cons.Gradient = v
	} else {
		s.Cons.Piece = v
	}
}

// ConstructionRotate90 is the Construction tab's rotate_90 button
// (onMouseUp): while hovering it turns the construction rotation a quarter
// (the rotate arrows); while building it is the red cross — "new
// construction position": the ghosts removed and the window back to
// hovering, its tool on the map, the rotation's low two bits kept.
func (s *State) ConstructionRotate90(c *Cmd) {
	cs := &s.Cons
	if cs.Hover {
		cs.Rotation = (cs.Rotation + 1) & 3
		cs.Cost = 0
		return
	}
	s.RemoveConstructionGhosts(c)
	cs.Hover = true
	cs.Rotation &= 3
}

// ConstructionLoop is constructionLoop (0x004A0121's caller, the map click's
// end): building starts at the tile (x, y) (world units) at height — the
// state leaves hovering and the selected piece is built there as the
// Construct button would; refused, it is tried again 16 higher (maxRetries
// tries in all; not after error_can_only_build_above_ground, not from below
// zero). When every try is refused the loop presses rotate_90 — with the
// state no longer hovering that is the red cross: back to hovering, the
// rotation as it was (an earlier port turned it a quarter here). The cost
// and whether a piece was built. PARTIAL: shift's search downwards, the
// error sound, byte_113607E.
func (s *State) ConstructionLoop(c *Cmd, x, y, height int32, maxRetries int, road bool, objID uint8) (int32, bool) {
	cs := &s.Cons
	for {
		cs.Hover = false
		cs.X, cs.Y, cs.Z = x, y, height
		cs.arrowBlink, cs.arrowFrame = false, 0
		c.ErrorText = 0
		var cost int32
		var ok bool
		if road {
			cost, ok = s.ConstructRoad(c, objID)
		} else {
			cost, ok = s.ConstructTrack(c, objID)
		}
		if ok {
			return cost, true
		}
		if c.ErrorText != i18n.ErrorCanOnlyBuildAboveGround {
			if maxRetries--; maxRetries != 0 {
				if height -= 16; height >= 0 {
					height += 32
					continue
				}
			}
		}
		s.ConstructionRotate90(c)
		return 0, false
	}
}

// UpdateConstruction is updateConstruction (the Construction tab's onUpdate,
// 0x0049FD66): nothing while hovering; else the selected piece stands as a
// ghost at the state's position (placeTrackGhost / placeRoadGhost with
// getTrackPlacementArgs / getRoadPlacementArgs; none when the piece doesn't
// fit the rotation) and the arrow blinks there. Upstream places the ghost
// whenever GhostVisibilityFlags::track is clear — every change of position
// or piece clears it through removeConstructionGhosts, and a refused ghost
// leaves it clear; here: whenever the ghost standing isn't the one wanted.
// PARTIAL: the bridge retry (sub_4A193B).
func (s *State) UpdateConstruction(c *Cmd, road bool, objID uint8) {
	if s.Cons.Hover {
		return
	}
	if road {
		if a, ok := s.roadPlacement(objID); !ok {
			s.RemoveRoadGhost(c)
		} else if g := s.RoadGhost; g == nil || *g != (RoadGhost{a.X, a.Y, a.Z, a.Rotation, a.RoadID, a.RoadObjectID}) {
			s.RemoveTrackGhost(c)
			s.placeRoadGhost(c, a)
		}
	} else {
		if a, ok := s.trackPlacement(objID); !ok {
			s.RemoveTrackGhost(c)
		} else if g := s.TrackGhost; g == nil || *g != (TrackGhost{a.X, a.Y, a.Z, a.Rotation, a.TrackID, a.TrackObjectID}) {
			s.RemoveRoadGhost(c)
			s.placeTrackGhost(c, a)
		}
	}
	s.updateConstructionArrow()
}

// updateConstructionArrow is updateConstructionArrow: every sixth update the
// arrow is aimed at the state's position and rotation and toggled.
func (s *State) updateConstructionArrow() {
	cs := &s.Cons
	cs.arrowFrame--
	if cs.arrowFrame == 0xFF {
		cs.arrowFrame = 5
		cs.arrowBlink = !cs.arrowBlink
		s.ConsArrow = ConstructionArrow{On: cs.arrowBlink, X: cs.X, Y: cs.Y, Z: cs.Z, Direction: cs.Rotation}
	}
}

// EndConstruction is the construction tool going (Common::onClose, a tab
// switch, the tool cancelled): the ghosts and the arrow off.
func (s *State) EndConstruction(c *Cmd) {
	s.RemoveConstructionGhosts(c)
	s.ConsArrow.On = false
}

// ConstructionEnd is one end of a piece as the construction state takes it
// (cState.nextTile / previousTile with their rotations).
type ConstructionEnd struct {
	X, Y, Z  int32
	Rotation uint8
}

// NextAndPreviousTile is Common::setNextAndPreviousTrackTile /
// setNextAndPreviousRoadTile (Construction/Common.cpp): for the track or road
// element (its id, sequence index and rotation) at world (x, y, z), where
// construction continues after the piece's end (next) and before its start
// (previous: the start itself turned round, a step back unless diagonal).
func NextAndPreviousTile(road bool, id, seq, rotation uint8, x, y, z int32) (next, previous ConstructionEnd, ok bool) {
	var pieces []PreviewTrack
	var size TrackCoord
	if road {
		if int(id) >= len(RoadPieces) {
			return next, previous, false
		}
		pieces, size = RoadPieces[id], RoadCoordinates[int(id)<<3|int(rotation&3)]
	} else {
		if int(id) >= len(TrackPieces) {
			return next, previous, false
		}
		pieces, size = TrackPieces[id], TrackCoordinates[int(id)<<3|int(rotation&3)]
	}
	if int(seq) >= len(pieces) {
		return next, previous, false
	}
	p := pieces[seq]
	dx, dy := rotOffset(p.X, p.Y, rotation)
	fx, fy, fz := x-int32(dx), y-int32(dy), z-int32(p.Z)
	next = ConstructionEnd{fx + int32(size.DX), fy + int32(size.DY), fz + int32(size.DZ), size.RotEnd}
	unk := ReverseRotation[size.RotBegin]
	previous = ConstructionEnd{fx, fy, fz, unk}
	if unk < 12 {
		previous.X += kRotationOffset[unk][0]
		previous.Y += kRotationOffset[unk][1]
	}
	return next, previous, true
}

// ConstructionOpenAt is openAtTrack / openAtRoad's state (Common.cpp
// 0x004A0EAD, 0x004A147F): building on from end, straight and level, with
// the element's mods.
func (s *State) ConstructionOpenAt(end ConstructionEnd, mods uint8) {
	cs := &s.Cons
	cs.Hover = false
	cs.X, cs.Y, cs.Z, cs.Rotation = end.X, end.Y, end.Z, end.Rotation
	cs.Piece, cs.Gradient = 0, 0
	cs.Mods = mods
	cs.Cost = 0
	cs.arrowBlink, cs.arrowFrame = false, 0
}
