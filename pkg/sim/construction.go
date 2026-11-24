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
	// signalGhost is the signal tab's ghost (GhostVisibilityFlags::signal and
	// cState.signalGhost…): the placement it was made with.
	signalGhost *SignalsPlacementAuto
	// X, Y, Z are cState.x, y and constructionZ (world units): where the next
	// piece starts — while hovering, where the ghost under the cursor stands.
	X, Y, Z int32
	// Rotation is constructionRotation: 0–3 straight on, 4–7 and 8–11 beside
	// the other track of a dual track, 12–15 diagonal.
	Rotation uint8
	// Hover is constructionHover: the window's tool is on the map, choosing
	// where to start; false while building on from X, Y, Z.
	Hover bool
	// Road is cState.trackType's bit 7: the window builds road (the window
	// sets it).
	Road bool
	// Piece and Gradient are lastSelectedTrackPiece and
	// lastSelectedTrackGradient (construct.Straight …, construct.Level …).
	Piece, Gradient uint8
	// Mods is lastSelectedMods: the Overhead tab's ticked extras.
	Mods uint8
	// Bridge is lastSelectedBridge + 1 (0: none), the bridge type a piece is
	// built with where it needs one.
	Bridge uint8
	// BridgeList is cState.bridgeList: the bridges available for this track
	// or road (getAvailableCompatibleBridges), each slot + 1 as Bridge (0:
	// the list's end, upstream's 0xFF).
	BridgeList [9]uint8
	// Cost is trackCost / roadCost: what the ghost standing would cost, or
	// the piece just built did (0: none — GameCommands::kFailure).
	Cost int32

	// ghostFlags is byte_1136076: flags_1136073 of the last ghost placed
	// (bit 0: it needs a bridge), 0 after a build or removal.
	ghostFlags uint8
	// placeUnk is byte_113607E: 0 while constructionLoop tries, 1 once a
	// piece of it stands, 0 again with the red cross; it goes to createTrack
	// as TrackPlacementArgs::unk, which nothing reads (CreateTrack.cpp).
	placeUnk uint8
	// pieceChanged is byte_113603A: a piece button was pressed since the
	// last ghost placement (changeTrackPiece sets it, every placeTrackGhost /
	// placeRoadGhost clears it) — the ghost may then try another bridge.
	pieceChanged uint8

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
// behind (Common::createConstructionWindow, then trackWindow's rotate_90):
// at (0x1800, 0x1800, 0x100) — where activateSelectedConstructionWidgets
// outlines the piece until the tool reaches the map —, hovering, rotation
// 0, straight and level, byte_113603A clear; the bridge and mods stay the
// window's.
func (s *State) ResetConstruction() {
	s.Cons = ConstructionState{X: 0x1800, Y: 0x1800, Z: 0x100, Hover: true, Mods: s.Cons.Mods, Bridge: s.Cons.Bridge, BridgeList: s.Cons.BridgeList, Road: s.Cons.Road}
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
	s.RemoveSignalGhost(c)
}

// signalGhostFlags are placeSignalGhost's / removeSignalGhost's doCommand flags.
const signalGhostFlags = CmdApply | CmdPreventBuildingClearing | CmdNoErrorWindow | CmdNoPayment | CmdGhost

// PlaceSignalGhost is SignalTab.cpp placeSignalGhost: the placement as a
// ghost (auto when its step isn't 0), remembered when made; its cost.
func (s *State) PlaceSignalGhost(c *Cmd, a SignalsPlacementAuto) (int32, bool) {
	cost, ok := s.CreateSignalsAuto(c, signalGhostFlags, a)
	if ok {
		g := a
		s.Cons.signalGhost = &g
	}
	return cost, ok
}

// SignalGhostShown is hasGhostVisibilityFlag(GhostVisibilityFlags::signal).
func (s *State) SignalGhostShown() bool { return s.Cons.signalGhost != nil }

// SignalGhostIs reports whether the signal ghost shown was made with a
// (onToolUpdate's "same as before": nothing to do).
func (s *State) SignalGhostIs(a SignalsPlacementAuto) bool {
	return s.Cons.signalGhost != nil && *s.Cons.signalGhost == a
}

// RemoveSignalGhost is removeSignalGhost (0x0049FEF6): the ghost's removal
// along the same walk, then no ghost.
func (s *State) RemoveSignalGhost(c *Cmd) {
	g := s.Cons.signalGhost
	if g == nil {
		return
	}
	s.Cons.signalGhost = nil
	s.RemoveSignalsAuto(c, signalGhostFlags, SignalsRemovalAuto{SignalRemoval: SignalRemoval{X: g.X, Y: g.Y, Z: g.Z, Rotation: g.Rotation,
		TrackID: g.TrackID, Index: g.Index, TrackObjectID: g.TrackObjectID, Left: g.Left, Right: g.Right}, Step: g.Step})
}

// trackPlacement is getTrackPlacementArgs at the state's position: the
// selected piece's track id (getTrackPieceId) with the window's mods and
// bridge; ok false when the piece can't be built in this rotation.
func (s *State) trackPlacement(objID uint8) (TrackPlacement, bool) {
	cs := &s.Cons
	p, ok := construct.GetTrackPieceID(cs.Piece, cs.Gradient, cs.Rotation)
	if !ok {
		return TrackPlacement{}, false
	}
	return TrackPlacement{X: cs.X, Y: cs.Y, Z: cs.Z, Rotation: p.Rotation & 3, TrackID: p.ID,
		Mods: cs.Mods, Bridge: s.bridge(), TrackObjectID: objID, Unk: cs.placeUnk&1 != 0}, true
}

// roadPlacement is getRoadPlacementArgs, as trackPlacement.
func (s *State) roadPlacement(objID uint8) (RoadPlacement, bool) {
	cs := &s.Cons
	p, ok := construct.GetRoadPieceID(cs.Piece, cs.Gradient, cs.Rotation)
	if !ok {
		return RoadPlacement{}, false
	}
	return RoadPlacement{X: cs.X, Y: cs.Y, Z: cs.Z, Rotation: p.Rotation & 3, RoadID: p.ID,
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
	cost, ok := s.constructTrack(c, objID)
	s.ActivateSelection() // constructTrackOrRoad
	return cost, ok
}

func (s *State) constructTrack(c *Cmd, objID uint8) (int32, bool) {
	s.Cons.Cost, s.Cons.ghostFlags = 0, 0
	s.RemoveConstructionGhosts(c)
	a, ok := s.trackPlacement(objID)
	if !ok {
		return 0, false
	}
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	cost, ok := s.DoCommandFor(c, CmdApply, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return s.CreateTrack(c, f, a)
	})
	if !ok {
		return 0, false
	}
	s.Cons.Cost = cost
	s.advance(TrackCoordinates[int(a.TrackID)<<3|int(a.Rotation)])
	return cost, true
}

// ConstructRoad is constructRoad, as ConstructTrack through createRoad;
// objID is the window's road object (cState.trackType). Refused with
// unable_to_cross_or_create_junction_with_string over another road
// (alternateRoadObjectId), the window may switch to that road
// (sub_4A18D4, SwitchRoadType) and the piece is tried once more — with the
// road type the window has then, even when it didn't change (as upstream).
func (s *State) ConstructRoad(c *Cmd, objID uint8) (int32, bool) {
	cost, ok := s.constructRoad(c, &objID)
	s.ActivateSelection() // constructTrackOrRoad
	return cost, ok
}

// constructRoad is ConstructRoad; *objID follows a switch of road type.
func (s *State) constructRoad(c *Cmd, objID *uint8) (int32, bool) {
	s.Cons.Cost, s.Cons.ghostFlags = 0, 0
	s.RemoveConstructionGhosts(c)
	a, ok := s.roadPlacement(*objID)
	if !ok {
		return 0, false
	}
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	place := func() (int32, bool) {
		return s.DoCommandFor(c, CmdApply, ExpenditureConstruction, func(f uint8) (int32, bool) {
			return s.CreateRoad(c, f, a)
		})
	}
	cost, ok := place()
	if !ok {
		alt := c.AltRoadObject
		if c.ErrorText != i18n.UnableToCrossOrCreateJunctionWithString || alt == 0xFF {
			return 0, false
		}
		s.switchToAlternateRoad(c, alt, objID)
		if a, ok = s.roadPlacement(*objID); !ok {
			return 0, false
		}
		// WindowManager::close(WindowType::error): the caller opens the
		// error window after the last try only.
		cost, ok = place()
		if !ok {
			return 0, false
		}
	}
	s.Cons.Cost = cost
	s.advance(RoadCoordinates[int(a.RoadID)<<3|int(a.Rotation)])
	return cost, true
}

// switchToAlternateRoad is sub_4A18D4 (0x004A18D4): the road met (alt)
// becomes the window's road type when it isn't already, both it and the
// current one are any-road-type compatible, and the controlling company has
// it (companyGetAvailableRoads); then Common::sub_4A3A50 (SwitchRoadType).
func (s *State) switchToAlternateRoad(c *Cmd, alt uint8, objID *uint8) {
	if alt == 0xFF || alt == *objID || c.Objects == nil {
		return
	}
	altObj, cur := c.roadObject(alt), c.roadObject(*objID)
	if altObj == nil || altObj.Flags&roadFlagAnyRoadTypeCompatible == 0 || cur == nil || cur.Flags&roadFlagAnyRoadTypeCompatible == 0 {
		return
	}
	for _, id := range s.companyAvailableRoads(s.ControllingCompany, c.Objects) {
		if id == alt|1<<7 {
			*objID = alt
			s.RemoveConstructionGhosts(c) // sub_4A3A50's removeConstructionGhosts
			if s.SwitchRoadType != nil {
				s.SwitchRoadType(alt)
			}
			s.ActivateSelection()
			return
		}
	}
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
	s.Cons.Cost, s.Cons.ghostFlags = 0, 0
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
	s.ActivateSelection()
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
	s.Cons.Cost, s.Cons.ghostFlags = 0, 0
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
	s.ActivateSelection()
	return true, false
}

// ChangeTrackPiece is changeTrackPiece (a piece or gradient button): the
// ghosts removed, the piece — or, slope, the gradient — selected.
func (s *State) ChangeTrackPiece(c *Cmd, v uint8, slope bool) {
	s.Cons.pieceChanged = 0xFF // byte_113603A = TrackPiece::null (the dual-track buttons: 0xFF)
	s.RemoveConstructionGhosts(c)
	s.Cons.Cost = 0
	if slope {
		s.Cons.Gradient = v
	} else {
		s.Cons.Piece = v
	}
	s.ActivateSelection()
}

// ChangeBridge is onDropdown on the bridge dropdown: the bridge chosen
// (lastSelectedBridge; slot), the ghosts removed so the ghost comes back
// with it, no cost.
func (s *State) ChangeBridge(c *Cmd, slot uint8) {
	s.Cons.Bridge = slot + 1
	s.RemoveConstructionGhosts(c)
	s.Cons.Cost = 0
	s.ActivateSelection()
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
		s.ActivateSelection()
		return
	}
	s.RemoveConstructionGhosts(c)
	cs.Hover = true
	cs.placeUnk = 0
	cs.Rotation &= 3
	s.ActivateSelection()
}

// ConstructionLoop is constructionLoop (0x004A0121's caller, the map click's
// end): building starts at the tile (x, y) (world units) at height — the
// state leaves hovering and the selected piece is built there as the
// Construct button would (byte_113607E 0 for it, 1 once built); refused, it
// is tried again 16 higher — with shift held, 16 lower — while tries are
// left (maxRetries), not after error_can_only_build_above_ground, not from
// below zero. When every try is refused the loop presses rotate_90 — with
// the state no longer hovering that is the red cross: back to hovering, the
// rotation as it was (an earlier port turned it a quarter here) — and plays
// the error sound (the caller: Audio::playSound at the mouse's x). The cost
// and whether a piece was built.
func (s *State) ConstructionLoop(c *Cmd, x, y, height int32, maxRetries uint32, road bool, objID uint8, shift bool) (int32, bool) {
	cs := &s.Cons
	for {
		cs.Hover = false
		cs.placeUnk = 0
		cs.X, cs.Y, cs.Z = x, y, height
		cs.arrowBlink, cs.arrowFrame = false, 0
		s.ActivateSelection()
		c.ErrorText = 0
		var cost int32
		var ok bool
		if road {
			cost, ok = s.constructRoad(c, &objID)
		} else {
			cost, ok = s.constructTrack(c, objID)
		}
		s.ActivateSelection() // onMouseUp(construct): constructTrackOrRoad
		if ok {
			cs.placeUnk = 1
			return cost, true
		}
		if c.ErrorText != i18n.ErrorCanOnlyBuildAboveGround {
			if maxRetries--; maxRetries != 0 {
				if height -= 16; height >= 0 {
					if !shift {
						height += 32
					}
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
// After each placement the window takes the bridge the piece met (sub_4A193B).
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
			s.Cons.ghostFlags = c.Flags73
			s.adoptBridge(c)
			s.ActivateSelection()
		}
	} else {
		if a, ok := s.trackPlacement(objID); !ok {
			s.RemoveTrackGhost(c)
		} else if g := s.TrackGhost; g == nil || *g != (TrackGhost{a.X, a.Y, a.Z, a.Rotation, a.TrackID, a.TrackObjectID}) {
			s.RemoveRoadGhost(c)
			s.placeTrackGhost(c, a)
			s.Cons.ghostFlags = c.Flags73
			s.adoptBridge(c)
			s.ActivateSelection()
		}
	}
	s.updateConstructionArrow()
}

// ActivateSelection is activateSelectedConstructionWidgets' map part
// (activateSelectedTrackWidgets / activateSelectedRoadWidgets 0x0049F1B5,
// on the Construction tab): enableConstruct with unk_03 (the outline in
// ExtColour::unk2B) and the selected piece's tiles at the state's position
// in the piece's rotation — none selectable: piece 0 in the construction
// rotation.
func (s *State) ActivateSelection() {
	cs := &s.Cons
	rot, id := cs.Rotation, uint8(0)
	pieces := TrackPieces[:]
	p, ok := construct.GetTrackPieceID(cs.Piece, cs.Gradient, cs.Rotation)
	if cs.Road {
		pieces = RoadPieces[:]
		p, ok = construct.GetRoadPieceID(cs.Piece, cs.Gradient, cs.Rotation)
	}
	if ok {
		rot, id = p.Rotation, p.ID
	}
	if int(id) >= len(pieces) {
		return
	}
	s.ConsSelection.On, s.ConsSelection.Unk03 = true, true
	s.setSelectedTilesFromPiece(pieces[id], cs.X, cs.Y, rot&3)
}

// GhostNeedsBridge is byte_1136076 & 1: the last ghost placed needs a
// bridge (the window shows its bridge box while building on).
func (s *State) GhostNeedsBridge() bool { return s.Cons.ghostFlags&1 != 0 }

// adoptBridge is sub_4A193B (0x004A193B): the bridge of the bridged element
// the last placement met (byte_1136075), when it is in the window's list,
// becomes the window's bridge.
func (s *State) adoptBridge(c *Cmd) {
	for _, b := range s.Cons.BridgeList {
		if b == 0 {
			return
		}
		if c.Bridge == b-1 {
			s.Cons.Bridge = b
			return
		}
	}
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
	s.ConsSelection.On = false // toolCancel / onClose: enableConstruct off
	s.Cons.ghostFlags = 0      // switchTab: byte_1136076 = 0
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
	cs.Cost, cs.ghostFlags = 0, 0
	cs.arrowBlink, cs.arrowFrame = false, 0
	s.ActivateSelection()
}
