package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// blueprint.go — the construction window's copy / paste tools' blueprint: the
// company's track, signals and train stations in a rectangle of tiles,
// copied as the placement arguments that build them again (relative to the
// rectangle's lowest corner), turned, and placed or taken up elsewhere — as
// a ghost under the paste tool, for real on its click.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/Blueprints.cpp
// (CopiedTrack, copyTrackToBlueprint, placeBlueprint, removeBlueprint,
// rotateBlueprint); ConstructionTab.cpp removeBlueprintGhosts.

// Blueprint is CopiedTrack: the placements, each position relative to the
// blueprint's pivot.
type Blueprint struct {
	Tracks   []TrackPlacement
	Signals  []SignalPlacement
	Stations []TrainStationPlacement
}

// blueprintGhostFlags are removeBlueprintGhosts' / the paste tool's ghost
// placement flags (apply | noErrorWindow | noPayment | ghost).
const blueprintGhostFlags = CmdApply | CmdNoErrorWindow | CmdNoPayment | CmdGhost

// piece0 is TrackData::getTrackPiece(trackId)[0]'s offset.
func piece0(trackID uint8) (int32, int32, int32) {
	if int(trackID) < len(TrackPieces) && len(TrackPieces[trackID]) > 0 {
		p := TrackPieces[trackID][0]
		return int32(p.X), int32(p.Y), int32(p.Z)
	}
	return 0, 0, 0
}

// CopyTrackToBlueprint is copyTrackToBlueprint: every tile from a to b
// (tiles, both included), its elements in order. A track element of the
// controlling company, not a ghost nor AI-allocated, at sequence index 0 is
// copied (its position less its piece's first offset) and becomes the
// processed track — any other track element clears it; a signal or train
// station after a processed track, not a ghost nor AI-allocated, is copied
// with that track's piece (a signal's two sides as one placement, or two
// when their signal types differ). Then every position is made relative to
// the pivot: the lowest x and y of the track copied (at most b's corner),
// at the height of the surface there.
func (m *Map) CopyTrackToBlueprint(a, b [2]int) Blueprint {
	dx, dy := 1, 1
	if b[0]-a[0] <= 0 {
		dx = -1
	}
	if b[1]-a[1] <= 0 {
		dy = -1
	}
	var bp Blueprint
	for ty := a[1]; ty != b[1]+dy; ty += dy {
		for tx := a[0]; tx != b[0]+dx; tx += dx {
			t := m.TileAt(tx, ty)
			if t == nil {
				continue
			}
			x, y := int32(tx)*tileSize, int32(ty)*tileSize
			var processed *scenario.TrackElement
			for _, r := range t.Order {
				switch r.Kind {
				case scenario.KindTrack:
					if int(r.Index) >= len(t.Tracks) {
						continue
					}
					el := &t.Tracks[r.Index]
					if el.Ghost || el.AiAllocated || el.Owner != m.ControllingCompany || el.SequenceIndex != 0 {
						processed = nil
						continue
					}
					px, py, pz := piece0(el.TrackID)
					bridge := uint8(0xFF)
					if el.HasBridge {
						bridge = el.BridgeID
					}
					bp.Tracks = append(bp.Tracks, TrackPlacement{X: x - px, Y: y - py, Z: int32(el.BaseZ)*smallZStep - pz,
						Rotation: el.Rotation, TrackID: el.TrackID, TrackObjectID: el.TrackObjectID, Bridge: bridge, Mods: el.Mods})
					processed = el
				case scenario.KindSignal:
					if processed == nil || int(r.Index) >= len(t.Signals) {
						continue
					}
					el := &t.Signals[r.Index]
					if el.Ghost || el.AiAllocated {
						continue
					}
					px, py, pz := piece0(processed.TrackID)
					s := SignalPlacement{X: x - px, Y: y - py, Z: int32(el.BaseZ)*smallZStep - pz, Index: 0,
						Rotation: el.Rotation, TrackID: processed.TrackID, TrackObjectID: processed.TrackObjectID, Type: 0xFF}
					left, right := el.Sides[0], el.Sides[1]
					if left.HasSignal() && !el.LeftGhost {
						s.Type, s.Left = left.ObjectID(), true
					}
					if right.HasSignal() && !el.RightGhost {
						// both sides with different signal types: two placements
						if s.Left && s.Type != right.ObjectID() {
							bp.Signals = append(bp.Signals, s)
							s.Left = false
						}
						s.Type, s.Right = right.ObjectID(), true
					}
					bp.Signals = append(bp.Signals, s)
				case scenario.KindStation:
					if processed == nil || int(r.Index) >= len(t.Stations) {
						continue
					}
					el := &t.Stations[r.Index]
					if el.Ghost || el.AiAllocated {
						continue
					}
					px, py, pz := piece0(processed.TrackID)
					bp.Stations = append(bp.Stations, TrainStationPlacement{X: x - px, Y: y - py, Z: int32(el.BaseZ)*smallZStep - pz,
						Rotation: el.Rotation, TrackID: processed.TrackID, Index: 0, TrackObjectID: processed.TrackObjectID, Type: el.ObjectID})
				}
			}
		}
	}
	minX, minY := int32(b[0])*tileSize, int32(b[1])*tileSize
	for _, t := range bp.Tracks {
		minX, minY = min(minX, t.X), min(minY, t.Y)
	}
	var minZ int32
	if t := m.TileAt(int(minX/tileSize), int(minY/tileSize)); t != nil {
		minZ = int32(t.BaseZ) * smallZStep
	}
	for i := range bp.Tracks {
		bp.Tracks[i].X, bp.Tracks[i].Y, bp.Tracks[i].Z = bp.Tracks[i].X-minX, bp.Tracks[i].Y-minY, bp.Tracks[i].Z-minZ
	}
	for i := range bp.Signals {
		bp.Signals[i].X, bp.Signals[i].Y, bp.Signals[i].Z = bp.Signals[i].X-minX, bp.Signals[i].Y-minY, bp.Signals[i].Z-minZ
	}
	for i := range bp.Stations {
		bp.Stations[i].X, bp.Stations[i].Y, bp.Stations[i].Z = bp.Stations[i].X-minX, bp.Stations[i].Y-minY, bp.Stations[i].Z-minZ
	}
	return bp
}

// Rotate is rotateBlueprint: every placement turned by rotation quarter
// turns about the pivot (Math::Vector::rotate), its own rotation with it.
func (bp *Blueprint) Rotate(rotation uint8) {
	turn := func(x, y *int32, r *uint8) {
		*r = (*r + rotation) & 3
		nx, ny := rotOffset(int16(*x), int16(*y), rotation)
		*x, *y = int32(nx), int32(ny)
	}
	for i := range bp.Tracks {
		turn(&bp.Tracks[i].X, &bp.Tracks[i].Y, &bp.Tracks[i].Rotation)
	}
	for i := range bp.Signals {
		turn(&bp.Signals[i].X, &bp.Signals[i].Y, &bp.Signals[i].Rotation)
	}
	for i := range bp.Stations {
		turn(&bp.Stations[i].X, &bp.Stations[i].Y, &bp.Stations[i].Rotation)
	}
}

// BlueprintResult is what PlaceBlueprint did beyond its cost: Built, some
// placement was made (placeBlueprint's builtAnything: the error window
// closes); Failed, the command state of the last placement refused after
// the trial (its error window, opened quietly — setErrorSound(false) — is
// the one left when nothing was built); nil, none was.
type BlueprintResult struct {
	Built  bool
	Failed *Cmd
}

// PlaceBlueprint is placeBlueprint at pos: first a trial of every track
// placement without apply (one refused: the whole paste is refused, c holds
// why); then the track, the signals and the stations, each a top-level
// command with the flags, a refusal skipped. The blueprint itself is left
// as it was (upstream works on a copy).
func (s *State) PlaceBlueprint(c *Cmd, bp *Blueprint, pos [3]int32, flags uint8) (int32, BlueprintResult, bool) {
	var res BlueprintResult
	tracks := make([]TrackPlacement, len(bp.Tracks))
	for i, t := range bp.Tracks {
		t.X, t.Y, t.Z = t.X+pos[0], t.Y+pos[1], t.Z+pos[2]
		tracks[i] = t
		if _, ok := s.DoCommandFor(c, flags&^CmdApply, ExpenditureConstruction, func(f uint8) (int32, bool) { return s.CreateTrack(c, f, t) }); !ok {
			return 0, res, false
		}
	}
	var total int32
	do := func(run func(f uint8) (int32, bool)) {
		cost, ok := s.DoCommandFor(c, flags, ExpenditureConstruction, run)
		if ok {
			res.Built = true
			total += cost
		} else {
			failed := *c
			res.Failed = &failed
		}
	}
	for _, t := range tracks {
		do(func(f uint8) (int32, bool) { return s.CreateTrack(c, f, t) })
	}
	for _, sg := range bp.Signals {
		sg.X, sg.Y, sg.Z = sg.X+pos[0], sg.Y+pos[1], sg.Z+pos[2]
		do(func(f uint8) (int32, bool) { return s.CreateSignal(c, f, sg) })
	}
	g1 := s.blueprintG1
	for _, st := range bp.Stations {
		st.X, st.Y, st.Z = st.X+pos[0], st.Y+pos[1], st.Z+pos[2]
		do(func(f uint8) (int32, bool) { return s.CreateTrainStation(c, f, st, g1) })
	}
	return total, res, true
}

// RemoveBlueprint is removeBlueprint at pos: the stations, the signals and
// the track taken up — each at its piece's first tile, a top-level command
// with the flags, a refusal skipped; the total cost.
func (s *State) RemoveBlueprint(c *Cmd, bp *Blueprint, pos [3]int32, flags uint8) int32 {
	var total int32
	do := func(run func(f uint8) (int32, bool)) {
		if cost, ok := s.DoCommandFor(c, flags, ExpenditureConstruction, run); ok {
			total += cost
		}
	}
	at := func(x, y, z int32, trackID uint8) (int32, int32, int32) {
		px, py, pz := piece0(trackID)
		return x + pos[0] + px, y + pos[1] + py, z + pos[2] + pz
	}
	g1 := s.blueprintG1
	for _, st := range bp.Stations {
		x, y, z := at(st.X, st.Y, st.Z, st.TrackID)
		a := TrainStationRemoval{X: x, Y: y, Z: z, Rotation: st.Rotation & 3, Index: 0, TrackID: st.TrackID, TrackObjectID: st.TrackObjectID}
		do(func(f uint8) (int32, bool) { return s.RemoveTrainStation(c, f, a, g1) })
	}
	for _, sg := range bp.Signals {
		x, y, z := at(sg.X, sg.Y, sg.Z, sg.TrackID)
		a := SignalRemoval{X: x, Y: y, Z: z, Rotation: sg.Rotation & 3, Index: 0, TrackID: sg.TrackID, TrackObjectID: sg.TrackObjectID, Left: sg.Left, Right: sg.Right}
		do(func(f uint8) (int32, bool) { return s.RemoveSignal(c, f, a) })
	}
	for _, t := range bp.Tracks {
		x, y, z := at(t.X, t.Y, t.Z, t.TrackID)
		a := TrackRemoval{X: x, Y: y, Z: z, Rotation: t.Rotation & 3, Index: 0, TrackID: t.TrackID, TrackObjectID: t.TrackObjectID}
		do(func(f uint8) (int32, bool) { return s.RemoveTrack(c, f, a, g1) })
	}
	return total
}

// The paste tool's ghost (ConstructionTab.cpp): the copied blueprint
// (_copiedTrack — kept while the game runs, whichever window) and where its
// ghost stands (GhostVisibilityFlags::blueprint with ghostRemovalTrackPos).

// SetBlueprint is the copy tool's release (_copiedTrack = …) or the Copy
// button (= nullopt: bp nil), after the ghosts went.
func (s *State) SetBlueprint(bp *Blueprint) { s.blueprint = bp }

// Blueprint is _copiedTrack (nil: none).
func (s *State) Blueprint() *Blueprint { return s.blueprint }

// SetBlueprintG1 gives the station commands the G1 they read (a station's
// size); nil without one.
func (s *State) SetBlueprintG1(g1 G1Raw) { s.blueprintG1 = g1 }

// RemoveBlueprintGhosts is removeBlueprintGhosts: a standing blueprint ghost
// taken up where it was placed.
func (s *State) RemoveBlueprintGhosts(c *Cmd) {
	if s.blueprintGhost == nil || s.blueprint == nil {
		return
	}
	s.RemoveBlueprint(c, s.blueprint, *s.blueprintGhost, blueprintGhostFlags)
	s.blueprintGhost = nil
}

// BlueprintGhostShown is hasGhostVisibilityFlag(blueprint).
func (s *State) BlueprintGhostShown() bool { return s.blueprintGhost != nil }

// PasteToolUpdate is onToolUpdate's paste part at pos (getConstructionPos:
// the tile's world position and height): unless the ghost already stands
// there, the old one goes and the blueprint is placed there as a ghost —
// remembered when that worked.
func (s *State) PasteToolUpdate(c *Cmd, pos [3]int32) {
	if s.blueprint == nil {
		return
	}
	if s.blueprintGhost != nil && *s.blueprintGhost == pos {
		return
	}
	s.RemoveBlueprintGhosts(c)
	if _, _, ok := s.PlaceBlueprint(c, s.blueprint, pos, blueprintGhostFlags); ok {
		p := pos
		s.blueprintGhost = &p
	}
}

// PasteToolUp is onToolUp's paste part at pos: the ghost goes and the
// blueprint is built (apply); ok false when the trial refused it.
func (s *State) PasteToolUp(c *Cmd, pos [3]int32) (int32, BlueprintResult, bool) {
	s.RemoveBlueprintGhosts(c)
	return s.PlaceBlueprint(c, s.blueprint, pos, CmdApply)
}

// ConstructionPos is getConstructionPos(x, y) for the tile (tx, ty): its
// world position at its land height, at least its water's; ok false off the
// map.
func (s *State) ConstructionPos(tx, ty int) ([3]int32, bool) {
	land, water, ok := s.constructionHeight(tx, ty)
	if !ok {
		return [3]int32{}, false
	}
	return [3]int32{int32(tx) * tileSize, int32(ty) * tileSize, max(land, water)}, true
}
