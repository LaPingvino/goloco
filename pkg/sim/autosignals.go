package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Signal auto placement (upstream 04e7900, bec1570):
// GameCommands/Track/AutoSignals.h, CreateSignalsAuto.cpp, RemoveSignalsAuto.cpp
// and Map/Track/TrackIteration.h — from the clicked piece, walk the track
// both ways to the next junction or dead end, placing (removing) a signal
// every step pieces.

// SignalsPlacementAuto is SignalsPlacementAutoArgs: a signal placement and
// the step (0: only the clicked piece, as createSignal).
type SignalsPlacementAuto struct {
	SignalPlacement
	Step uint8
}

// SignalsRemovalAuto is SignalsRemovalAutoArgs.
type SignalsRemovalAuto struct {
	SignalRemoval
	Step uint8
}

// signal sides as upstream's (Left = bit 15, Right = bit 14).
const (
	signalSideLeft  uint16 = 1 << 15
	signalSideRight uint16 = 1 << 14
	signalSidesBoth        = signalSideLeft | signalSideRight
)

func signalSides(left, right bool) uint16 {
	var s uint16
	if left {
		s |= signalSideLeft
	}
	if right {
		s |= signalSideRight
	}
	return s
}

// iterateTrackToJunction is Track::iterateTrackToJunction: from the piece
// starting at (x, y, z) in direction tad, f gets each piece's start (a
// reversed piece's from its far end) until f says stop, the piece's end has
// no single connection (a junction or nothing), or the walk comes back to
// where it began.
func (m *Map) iterateTrackToJunction(x, y, z int32, tad uint16, trackObj, company uint8, f func(x, y, z int32, tad uint16) bool) {
	sx, sy, sz := x, y, z
	for {
		cx, cy, cz := x, y, z
		if tad&(1<<2) != 0 && int(tad) < len(TrackCoordinates) {
			ts := TrackCoordinates[tad]
			cx, cy, cz = cx+int32(ts.DX), cy+int32(ts.DY), cz+int32(ts.DZ)
			if ts.RotEnd < 12 {
				cx -= kRotationOffset[ts.RotEnd][0]
				cy -= kRotationOffset[ts.RotEnd][1]
			}
		}
		if !f(cx, cy, cz, tad) {
			return
		}
		end, ok := TrackConnectionEnd(x, y, z, tad)
		if !ok {
			return
		}
		tc := m.TrackConnections(end.X, end.Y, end.Z, end.NextRotation, company, trackObj, 0, 0)
		if len(tc.Connections) != 1 {
			return
		}
		x, y, z = end.X, end.Y, end.Z
		tad = tc.Connections[0] & TaDBasicMask
		if x == sx && y == sy && z == sz {
			return
		}
	}
}

// autoSignalAction places or removes a signal on one piece (index 0 of the
// piece at x, y, z in direction tad) — the nested createSignal / removeSignal.
type autoSignalAction func(x, y, z int32, tad, sides uint16, trackObj, index, flags uint8) (int32, bool)

// autoPlaceSignals is AutoSignals.h autoPlaceSignals: along the walk, every
// step-th piece (counting from initialStep) without a level crossing gets the
// action — its sides swapped on a reversed piece unless both; skip stops the
// walk (createSignal: a piece that already has a signal), and so does a piece
// that isn't signal compatible (its error text kept). failure: a missing piece
// or a refused action.
func (m *Map) autoPlaceSignals(c *Cmd, x, y, z int32, tad uint16, trackObj uint8, sides uint16, step, initialStep, flags uint8,
	skip func(*scenario.TrackElement) bool, action autoSignalAction) (cost int32, failure, placed bool) {
	current := int(initialStep)
	if current >= int(step) {
		current = 0
	}
	m.iterateTrackToJunction(x, y, z, tad, trackObj, c.Company, func(px, py, pz int32, tad uint16) bool {
		rotation, trackID := uint8(tad&3), uint8(tad>>3)&0x3F
		if int(trackID) >= len(TrackPieces) {
			failure = true
			return false
		}
		p0 := TrackPieces[trackID][0]
		ox, oy := rotOffset(p0.X, p0.Y, rotation)
		sx, sy, sz := px+int32(ox), py+int32(oy), pz+int32(p0.Z)
		t, ti, _ := m.signalTrackAt(SignalRemoval{Rotation: rotation, TrackID: trackID, TrackObjectID: trackObj}, sx, sy, sz, 0)
		if t == nil {
			failure = true
			return false
		}
		el := &t.Tracks[ti]
		if skip(el) {
			return false
		}
		if text, ok := m.validateTrackIsSignalCompatible(px, py, pz, rotation, trackID, trackObj); !ok {
			c.ErrorText = text
			return false
		}
		if current == 0 && !el.HasLevelCrossing {
			s := sides
			if tad&(1<<2) != 0 && sides != signalSidesBoth {
				s ^= signalSidesBoth
			}
			v, ok := action(sx, sy, sz, tad, s, trackObj, 0, flags)
			if !ok {
				failure = true
				return false
			}
			placed = true
			cost += v
		}
		current++
		if current >= int(step) {
			current = 0
		}
		return true
	})
	return cost, failure, placed
}

// autoSignalsWalk is AutoSignals.h autoSignalsWalk: step 0 acts on the
// clicked piece alone; else a forward walk from the piece, and — when the
// piece's start has a single connection behind it — a backward walk from
// there (sides swapped unless both, starting at step 1 so the clicked piece
// isn't counted twice). Nothing acted on is a failure; the command's position
// is the clicked tile's centre.
func (m *Map) autoSignalsWalk(c *Cmd, x, y, z int32, trackID, rotation, index, trackObj uint8, sides uint16, step, flags uint8,
	skip func(*scenario.TrackElement) bool, action autoSignalAction) (int32, bool) {
	if step == 0 {
		return action(x, y, z, uint16(rotation&3)|uint16(trackID)<<3, sides, trackObj, index, flags)
	}
	if int(trackID) >= len(TrackPieces) || len(TrackPieces[trackID]) < int(index) || int(index) >= len(TrackPieces[trackID]) {
		return 0, false
	}
	p := TrackPieces[trackID][index]
	ox, oy := rotOffset(p.X, p.Y, rotation)
	sx, sy, sz := x-int32(ox), y-int32(oy), z-int32(p.Z)
	startTad := uint16(rotation) | uint16(trackID)<<3
	total, failure, placed := m.autoPlaceSignals(c, sx, sy, sz, startTad, trackObj, sides, step, 0, flags, skip, action)
	if failure {
		return 0, false
	}
	rx, ry, rz := sx, sy, sz
	ts := TrackCoordinates[startTad]
	if ts.RotBegin < 12 {
		rx -= kRotationOffset[ts.RotBegin][0]
		ry -= kRotationOffset[ts.RotBegin][1]
	}
	tc := m.TrackConnections(rx, ry, rz, ReverseRotation[ts.RotBegin], c.Company, trackObj, 0, 0)
	if len(tc.Connections) == 1 {
		reverseSides := sides
		if sides != signalSidesBoth {
			reverseSides ^= signalSidesBoth
		}
		cost, fail, pl := m.autoPlaceSignals(c, rx, ry, rz, tc.Connections[0]&TaDBasicMask, trackObj, reverseSides, step, 1, flags, skip, action)
		if fail {
			return 0, false
		}
		total += cost
		placed = placed || pl
	}
	if !placed {
		return 0, false
	}
	c.Pos = &[3]int32{x + 16, y + 16, z}
	return total, true
}

// CreateSignalsAuto is GameCommands::createSignalsAuto: signals along the
// track every Step pieces, stopping at a piece that already has one.
func (m *Map) CreateSignalsAuto(c *Cmd, flags uint8, a SignalsPlacementAuto) (int32, bool) {
	sides := signalSides(a.Left, a.Right)
	return m.autoSignalsWalk(c, a.X, a.Y, a.Z, a.TrackID, a.Rotation, a.Index, a.TrackObjectID, sides, a.Step, flags,
		func(el *scenario.TrackElement) bool { return el.HasSignal },
		func(x, y, z int32, tad, s uint16, trackObj, index, f uint8) (int32, bool) {
			return m.CreateSignal(c, f, SignalPlacement{X: x, Y: y, Z: z, Rotation: uint8(tad & 3), TrackID: uint8(tad>>3) & 0x3F, Index: index,
				Type: a.Type, TrackObjectID: trackObj, Left: s&signalSideLeft != 0, Right: s&signalSideRight != 0})
		})
}

// RemoveSignalsAuto is GameCommands::removeSignalsAuto: the signals along
// the track every Step pieces.
func (m *Map) RemoveSignalsAuto(c *Cmd, flags uint8, a SignalsRemovalAuto) (int32, bool) {
	sides := signalSides(a.Left, a.Right)
	return m.autoSignalsWalk(c, a.X, a.Y, a.Z, a.TrackID, a.Rotation, a.Index, a.TrackObjectID, sides, a.Step, flags,
		func(*scenario.TrackElement) bool { return false },
		func(x, y, z int32, tad, s uint16, trackObj, index, f uint8) (int32, bool) {
			return m.RemoveSignal(c, f, SignalRemoval{X: x, Y: y, Z: z, Rotation: uint8(tad & 3), TrackID: uint8(tad>>3) & 0x3F, Index: index,
				TrackObjectID: trackObj, Left: s&signalSideLeft != 0, Right: s&signalSideRight != 0})
		})
}
