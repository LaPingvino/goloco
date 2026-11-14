package sim

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// GameCommands/Track/CreateSignal.cpp: putting a signal on the left and / or
// right side of a track piece, on every tile of the piece.

// SignalPlacement is SignalPlacementArgs: a tile of the piece (world units,
// its base height), its rotation, the piece, which of its tiles, the signal
// object, the track object, and the sides (Left = bit 15, Right = bit 14).
type SignalPlacement struct {
	X, Y, Z       int32
	Rotation      uint8
	TrackID       uint8
	Index         uint8
	Type          uint8
	TrackObjectID uint8
	Left, Right   bool
}

func (a SignalPlacement) removal() SignalRemoval {
	return SignalRemoval{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, TrackID: a.TrackID, Index: a.Index, TrackObjectID: a.TrackObjectID}
}

// validateTrackIsSignalCompatible is Track::validateTrackIsSignalCompatible
// (04e7900; shared with the auto placement): every tile of the piece starting
// at (sx, sy, sz) is there and none has a station (both: StringIds::empty, as
// upstream's "TODO: need a better error message"), and no other (non-ghost)
// track at the same height shares a tile edge with it (a junction). ok false
// with the error text to set.
func (m *Map) validateTrackIsSignalCompatible(sx, sy, sz int32, rotation, trackID, trackObj uint8) (errText int, ok bool) {
	if int(trackID) >= len(TrackPieces) {
		return i18n.Empty, false
	}
	r := SignalRemoval{Rotation: rotation, TrackID: trackID, TrackObjectID: trackObj}
	for _, pc := range TrackPieces[trackID] {
		ox, oy := rotOffset(pc.X, pc.Y, rotation)
		t, ti, _ := m.signalTrackAt(r, sx+int32(ox), sy+int32(oy), sz+int32(pc.Z), pc.Index)
		if t == nil {
			return i18n.Empty, false
		}
		el := &t.Tracks[ti]
		if el.HasStation {
			return i18n.Empty, false
		}
		connect := pc.ConnectFlags[el.Rotation&3]
		for oi := range t.Tracks {
			o := &t.Tracks[oi]
			if oi == ti || o.BaseZ != el.BaseZ || o.Ghost || int(o.TrackID) >= len(TrackPieces) || int(o.SequenceIndex) >= len(TrackPieces[o.TrackID]) {
				continue
			}
			if TrackPieces[o.TrackID][o.SequenceIndex].ConnectFlags[o.Rotation&3]&connect != 0 {
				return i18n.SignalsCannotBeBuiltOnAJunction, false
			}
		}
	}
	return 0, true
}

// signalCost: per side asked, the signal's inflation-adjusted cost; a side
// that already has this signal costs nothing, one with another signal adds
// that one's sell price (read from the piece's first tile).
func (m *Map) signalCost(c *Cmd, a SignalPlacement, p0 PreviewTrack, sx, sy, sz int32) int32 {
	dx, dy := rotOffset(p0.X, p0.Y, a.Rotation)
	t, ti, pos := m.signalTrackAt(a.removal(), sx+int32(dx), sy+int32(dy), sz+int32(p0.Z), p0.Index)
	if t == nil {
		return 0
	}
	obj := func(id uint8) *trainSignalCost {
		if c.Objects == nil || int(id) >= len(c.Objects.TrainSignalObjects) || c.Objects.TrainSignalObjects[id] == nil {
			return nil
		}
		o := c.Objects.TrainSignalObjects[id]
		return &trainSignalCost{o.CostFactor, o.SellCostFactor, o.CostIndex}
	}
	so := obj(a.Type)
	if so == nil {
		return 0
	}
	base := m.InflationAdjustedCost(so.cost, so.index, 10)
	var sig *scenario.SignalElement
	if t.Tracks[ti].HasSignal {
		if si := signalAfter(t, pos); si >= 0 {
			sig = &t.Signals[si]
		}
	}
	var total int32
	side := func(s int) {
		total += base
		if sig == nil || !sig.Sides[s].HasSignal() {
			return
		}
		if sig.Sides[s].ObjectID() == a.Type {
			total -= base
		} else if ex := obj(sig.Sides[s].ObjectID()); ex != nil {
			total += m.InflationAdjustedCost(ex.sell, ex.index, 10)
		}
	}
	if a.Left {
		side(0)
	}
	if a.Right {
		side(1)
	}
	return total
}

type trainSignalCost struct {
	cost, sell int16
	index      uint8
}

// insertSignalAfter is insertElementAfterNoReorg for a signal element:
// directly after the track at order position k, with its occupied quarters.
func (t *Tile) insertSignalAfter(k int, el scenario.SignalElement, quarters uint8) int {
	t.Signals = append(t.Signals, el)
	f := quarters & 0xF
	if el.Ghost {
		f |= elemFlagGhost
	}
	if el.AiAllocated {
		f |= elemFlagAiAllocated
	}
	t.Order = append(t.Order, scenario.ElemRef{})
	copy(t.Order[k+2:], t.Order[k+1:])
	t.Order[k+1] = scenario.ElemRef{Kind: scenario.KindSignal, Index: uint16(len(t.Signals) - 1), Flags: f}
	return len(t.Signals) - 1
}

// CreateSignal is the createSignal game command (0x00488BDB), the position
// its tile's centre: refused when the track isn't there, has a level
// crossing, belongs to an incompatible company, or the piece isn't signal
// compatible (a station, a junction); the cost; then, applying, on every tile
// a signal element after the track when there is none (rotation, ghost / AI,
// clearZ; the track's flag set), and the asked sides set to the signal — a
// ghost only on free sides, and a ghost placement skips the sides the piece
// already has —, an animation for a real one. After a real placement the
// blocks at both ends of the piece are re-evaluated (sub_4A2AD7).
// PARTIAL: checkFreeElementsAndReorganise (goloco's tiles have no element
// limit); the viewport invalidation (goloco redraws).
func (m *Map) CreateSignal(c *Cmd, flags uint8, a SignalPlacement) (int32, bool) {
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	if int(a.TrackID) >= len(TrackPieces) {
		return 0, false
	}
	r := a.removal()
	t, ti, _ := m.signalTrackAt(r, a.X, a.Y, a.Z, a.Index)
	if t == nil {
		return 0, false
	}
	if t.Tracks[ti].HasLevelCrossing {
		c.ErrorText = i18n.LevelCrossingInTheWay
		return 0, false
	}
	if !c.companyCompatible(t.Tracks[ti].Owner) {
		return 0, false
	}
	pieces := TrackPieces[a.TrackID]
	if int(a.Index) >= len(pieces) {
		return 0, false
	}
	p := pieces[a.Index]
	dx, dy := rotOffset(p.X, p.Y, a.Rotation)
	sx, sy, sz := a.X-int32(dx), a.Y-int32(dy), a.Z-int32(p.Z)
	if text, ok := m.validateTrackIsSignalCompatible(sx, sy, sz, a.Rotation, a.TrackID, a.TrackObjectID); !ok {
		c.ErrorText = text
		return 0, false
	}
	total := m.signalCost(c, a, pieces[0], sx, sy, sz)
	ghost, ai := flags&CmdGhost != 0, flags&CmdAiAllocated != 0
	left, right := a.Left, a.Right // a ghost placement drops the sides already signalled
	for _, pc := range pieces {
		ox, oy := rotOffset(pc.X, pc.Y, a.Rotation)
		x, y, z := sx+int32(ox), sy+int32(oy), sz+int32(pc.Z)
		pt, pti, pos := m.signalTrackAt(r, x, y, z, pc.Index)
		if pt == nil {
			return 0, false
		}
		if ghost && pt.Tracks[pti].HasSignal {
			if si := signalAfter(pt, pos); si >= 0 {
				if pt.Signals[si].Sides[0].HasSignal() {
					left = false
				}
				if pt.Signals[si].Sides[1].HasSignal() {
					right = false
				}
			}
		}
		if flags&CmdApply == 0 {
			continue
		}
		el := &pt.Tracks[pti]
		if !el.HasSignal {
			pt.insertSignalAfter(pos, scenario.SignalElement{
				Rotation: el.Rotation, BaseZ: el.BaseZ, ClearZ: el.ClearZ, Ghost: ghost, AiAllocated: ai,
			}, pt.Order[pos].Flags)
			el.HasSignal = true
		}
		si := signalAfter(pt, pos)
		if si < 0 {
			return 0, false
		}
		sig := &pt.Signals[si]
		if left && (!ghost || !sig.Sides[0].HasSignal()) {
			sig.Sides[0].SetHasSignal(true)
			sig.LeftGhost = ghost
			sig.Sides[0].B4 = sig.Sides[0].B4&0xF0 | a.Type&0x0F
			sig.Sides[0].SetFrame(0)
			sig.Sides[0].SetAllLights(0)
			sig.HasLeftSignal, sig.SignalObjectID, sig.LeftFrame = true, a.Type, 0
		}
		if right && (!ghost || !sig.Sides[1].HasSignal()) {
			sig.Sides[1].SetHasSignal(true)
			sig.RightGhost = ghost
			sig.Sides[1].B4 = sig.Sides[1].B4&0xF0 | a.Type&0x0F
			sig.Sides[1].SetFrame(0)
			sig.Sides[1].SetAllLights(0)
			sig.HasRightSignal, sig.RightSignalObjectID, sig.RightFrame = true, a.Type, 0
		}
		if !ghost {
			m.CreateAnimation(0, x, y, sig.BaseZ)
		}
	}
	if flags&CmdApply != 0 && flags&(CmdAiAllocated|CmdGhost) == 0 {
		tad := uint16(a.Rotation&3) | uint16(a.TrackID)<<3
		if end, ok := TrackConnectionEnd(sx, sy, sz, tad); ok {
			if tc := m.TrackConnections(end.X, end.Y, end.Z, end.NextRotation, c.Company, a.TrackObjectID, 0, 0); len(tc.Connections) > 0 {
				m.UpdateBlockSignals(end.X, end.Y, end.Z, tc.Connections[0]&(TaDBasicMask|TaDHasSignal), c.Company, a.TrackObjectID)
			}
		}
		tc := TrackCoordinates[tad]
		nx, ny, nz := sx+int32(tc.DX), sy+int32(tc.DY), sz+int32(tc.DZ)
		if tc.RotEnd < 12 {
			nx -= int32(RotationOffset[tc.RotEnd][0])
			ny -= int32(RotationOffset[tc.RotEnd][1])
		}
		if cs := m.TrackConnections(nx, ny, nz, ReverseRotation[tc.RotEnd], c.Company, a.TrackObjectID, 0, 0); len(cs.Connections) > 0 {
			m.UpdateBlockSignals(nx, ny, nz, cs.Connections[0]&(TaDBasicMask|TaDHasSignal), c.Company, a.TrackObjectID)
		}
	}
	return total, true
}
