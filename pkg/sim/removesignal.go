package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// GameCommands/Track/RemoveSignal.cpp: taking signals off a track piece — the
// left and / or right side, on every tile of the piece.

// SignalRemoval is SignalRemovalArgs: a tile of the piece (world units, its
// base height), its rotation, the piece, which of its tiles, the track object,
// and which sides go (Left = flags bit 15, Right = bit 14).
type SignalRemoval struct {
	X, Y, Z       int32
	Rotation      uint8
	TrackID       uint8
	Index         uint8
	TrackObjectID uint8
	Left, Right   bool
}

// signalTrackAt is RemoveSignal's getElTrackAt: the track element at this
// height with the removal's rotation, sequence index, track object and piece —
// ghost or not, any owner — and the order position of its ref (-1 none).
func (m *Map) signalTrackAt(a SignalRemoval, x, y, z int32, seq uint8) (*Tile, int, int) {
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if t == nil {
		return nil, -1, -1
	}
	for pos, r := range t.Order {
		if r.Kind != scenario.KindTrack || int(r.Index) >= len(t.Tracks) {
			continue
		}
		el := &t.Tracks[r.Index]
		if int32(el.BaseZ)*smallZStep != z || el.Rotation != a.Rotation || el.SequenceIndex != seq ||
			el.TrackObjectID != a.TrackObjectID || el.TrackID != a.TrackID {
			continue
		}
		return t, int(r.Index), pos
	}
	return nil, -1, -1
}

// signalAfter is trackEntry->next()->as<SignalElement>(): the signal element
// right after the track's ref in the tile's order (-1 none).
func signalAfter(t *Tile, pos int) int {
	if pos+1 < len(t.Order) && t.Order[pos+1].Kind == scenario.KindSignal && int(t.Order[pos+1].Index) < len(t.Signals) {
		return int(t.Order[pos+1].Index)
	}
	return -1
}

// signalRemoveCost: the refund of the sides removed, read from the piece's
// first tile (0 without a signal; false when the track says signal but none follows).
func (m *Map) signalRemoveCost(c *Cmd, a SignalRemoval, p0 PreviewTrack, sx, sy, sz int32) (int32, bool) {
	dx, dy := rotOffset(p0.X, p0.Y, a.Rotation)
	t, ti, pos := m.signalTrackAt(a, sx+int32(dx), sy+int32(dy), sz+int32(p0.Z), p0.Index)
	if t == nil || !t.Tracks[ti].HasSignal {
		return 0, true
	}
	si := signalAfter(t, pos)
	if si < 0 {
		return 0, false
	}
	sig := &t.Signals[si]
	var total int32
	refund := func(side scenario.SignalSide) {
		id := side.ObjectID()
		if c.Objects != nil && int(id) < len(c.Objects.TrainSignalObjects) {
			if o := c.Objects.TrainSignalObjects[id]; o != nil {
				total += m.InflationAdjustedCost(o.SellCostFactor, o.CostIndex, 10)
			}
		}
	}
	if a.Left && sig.Sides[0].HasSignal() {
		refund(sig.Sides[0])
	}
	if a.Right && sig.Sides[1].HasSignal() {
		refund(sig.Sides[1])
	}
	return total, true
}

// RemoveSignal is the removeSignal game command (0x004891E4), the position its
// tile's centre: refused when the track isn't there or its owner isn't
// compatible; the refund from the first tile; then on each tile (all must be
// there) the asked sides are switched off — for a ghost removal only ghost
// sides — and the element goes when no side is left (the track's signal flag
// cleared); after a real (not ghost or AI) removal the block's signals are
// re-evaluated (sub_4A2AD7, UpdateBlockSignals). The cost.
// PARTIAL: the viewport invalidation (goloco redraws).
func (m *Map) RemoveSignal(c *Cmd, flags uint8, a SignalRemoval) (int32, bool) {
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	if int(a.TrackID) >= len(TrackPieces) {
		return 0, false
	}
	t, ti, _ := m.signalTrackAt(a, a.X, a.Y, a.Z, a.Index)
	if t == nil || !c.companyCompatible(t.Tracks[ti].Owner) {
		return 0, false
	}
	pieces := TrackPieces[a.TrackID]
	if int(a.Index) >= len(pieces) {
		return 0, false
	}
	p := pieces[a.Index]
	dx, dy := rotOffset(p.X, p.Y, a.Rotation)
	sx, sy, sz := a.X-int32(dx), a.Y-int32(dy), a.Z-int32(p.Z)
	cost, ok := m.signalRemoveCost(c, a, pieces[0], sx, sy, sz)
	if !ok {
		return 0, false
	}
	for _, pc := range pieces {
		ox, oy := rotOffset(pc.X, pc.Y, a.Rotation)
		pt, pti, pos := m.signalTrackAt(a, sx+int32(ox), sy+int32(oy), sz+int32(pc.Z), pc.Index)
		if pt == nil {
			return 0, false
		}
		if flags&CmdApply == 0 {
			continue
		}
		if !pt.Tracks[pti].HasSignal {
			return 0, true // upstream: return 0 (the rest of the piece unvisited)
		}
		si := signalAfter(pt, pos)
		if si < 0 {
			return 0, false
		}
		sig := &pt.Signals[si]
		ghost := flags&CmdGhost != 0
		if a.Left && (!ghost || sig.LeftGhost) {
			sig.Sides[0].SetHasSignal(false)
			sig.Sides[0].SetAllLights(0)
			sig.Sides[0].SetFrame(0)
			sig.LeftGhost, sig.HasLeftSignal, sig.LeftFrame = false, false, 0
		}
		if a.Right && (!ghost || sig.RightGhost) {
			sig.Sides[1].SetHasSignal(false)
			sig.Sides[1].SetAllLights(0)
			sig.Sides[1].SetFrame(0)
			sig.RightGhost, sig.HasRightSignal, sig.RightFrame = false, false, 0
		}
		m.invalidateTile(sx+int32(ox), sy+int32(oy), int32(sig.BaseZ)*smallZStep, int32(sig.BaseZ)*smallZStep+32, ZoomHalf, defaultRadius) // ViewportManager::invalidate
		sig.Ghost = false // upstream: "Why??"
		if !sig.Sides[0].HasSignal() && !sig.Sides[1].HasSignal() {
			pt.Tracks[pti].HasSignal = false
			pt.Signals = append(pt.Signals[:si], pt.Signals[si+1:]...)
			pt.removeRef(scenario.KindSignal, si)
		}
	}
	if flags&CmdApply != 0 && flags&(CmdAiAllocated|CmdGhost) == 0 {
		// sub_4A2AD7: the block's signals re-set occupied / free.
		m.UpdateBlockSignals(sx, sy, sz, uint16(a.TrackID)<<3|uint16(a.Rotation&3), c.Company, a.TrackObjectID)
	}
	return cost, true
}
