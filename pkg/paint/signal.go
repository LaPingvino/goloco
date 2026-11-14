package paint

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// Signals — a port of Paint/PaintSignal.cpp paintSignal / paintSignalSide.
// PARTIAL: one-way arrows (a view flag) and ghost signals are not ported.

// SignalSide is one side of a signal element.
type SignalSide struct {
	Has    bool
	Object int // signal object index
	Frame  uint8
	Ghost  bool // the side is a ghost (isLeftGhost / isRightGhost): no interaction
}

// SignalObj is what paintSignalSide reads from a train signal object.
type SignalObj struct {
	Image  int
	IsLeft bool // TrainSignalObjectFlags::isLeft
}

type signalOffsets struct{ off, bb [2]int }

var (
	signalOffs4FE830 = [16]signalOffsets{
		{[2]int{24, 4}, [2]int{24, 4}}, {[2]int{4, 8}, [2]int{4, 8}}, {[2]int{8, 28}, [2]int{8, 28}}, {[2]int{28, 24}, [2]int{28, 24}},
		{[2]int{24, 1}, [2]int{24, 1}}, {[2]int{1, 8}, [2]int{1, 8}}, {[2]int{8, 31}, [2]int{8, 31}}, {[2]int{31, 24}, [2]int{31, 24}},
		{[2]int{24, 16}, [2]int{24, 16}}, {[2]int{16, 8}, [2]int{16, 8}}, {[2]int{8, 16}, [2]int{8, 16}}, {[2]int{16, 24}, [2]int{16, 24}},
		{[2]int{6, 10}, [2]int{0, 0}}, {[2]int{10, 26}, [2]int{10, 26}}, {[2]int{26, 22}, [2]int{26, 22}}, {[2]int{22, 6}, [2]int{22, 6}},
	}
	signalOffs4FE870 = [16]signalOffsets{
		{[2]int{24, 28}, [2]int{24, 28}}, {[2]int{28, 8}, [2]int{28, 8}}, {[2]int{8, 4}, [2]int{8, 4}}, {[2]int{4, 24}, [2]int{4, 24}},
		{[2]int{24, 16}, [2]int{24, 16}}, {[2]int{16, 8}, [2]int{16, 8}}, {[2]int{8, 16}, [2]int{8, 16}}, {[2]int{16, 24}, [2]int{16, 24}},
		{[2]int{24, 31}, [2]int{24, 31}}, {[2]int{31, 8}, [2]int{31, 8}}, {[2]int{8, 1}, [2]int{8, 1}}, {[2]int{31, 24}, [2]int{31, 24}},
		{[2]int{22, 26}, [2]int{22, 26}}, {[2]int{26, 10}, [2]int{26, 10}}, {[2]int{10, 6}, [2]int{0, 0}}, {[2]int{6, 22}, [2]int{6, 22}},
	}
)

func (s *Session) paintSignalSide(side SignalSide, obj SignalObj, isRight bool, trackID, rotation int, height int, sprite func(int) *gfx8.Sprite) {
	if !side.Has {
		return
	}
	s.ItemType = ItemSignal // setItemType, setTrackModId(isRight ? 1 : 0) (PaintSignal.cpp:149)
	s.TrackModID = uint8(b2i(isRight))
	if side.Ghost {
		s.ItemType = ItemNoInteraction
	}
	dir := trackID<<3 | rotation
	if isRight {
		dir = trackID<<3 | (rotation + 4)
	}
	if dir >= len(sim.TrackCoordinates) {
		return
	}
	trackRot := int(sim.TrackCoordinates[dir].RotBegin) // getUnkTrack(...).rotationBegin
	offs := signalOffs4FE830
	if obj.IsLeft {
		offs = signalOffs4FE870
	}
	o := offs[trackRot&15]
	imgRot := (trackRot&3)<<1 | b2i(trackRot >= 12)
	sp := sprite(imgRot + obj.Image + int(side.Frame)<<3)
	if sp == nil {
		return
	}
	hOff := int(TrackSignalHeightOffset[trackID][0])
	if isRight {
		hOff = int(TrackSignalHeightOffset[trackID][1])
	}
	off := Pos3{o.off[0], o.off[1], hOff + height}
	s.AddAsParent(Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false)}, off, Pos3{o.bb[0], o.bb[1], off.Z + 4}, Pos3{1, 1, 14})
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// PaintSignal paints a signal element on the track element it belongs to
// (upstream: the element before it). trackSeq/trackIsLast are that track's
// sequence index and flag_6.
func (s *Session) PaintSignal(rotation uint8, baseZ uint8, trackID, trackSeq int, trackIsLast bool, left, right SignalSide, objOf func(int) (SignalObj, bool), sprite func(int) *gfx8.Sprite) {
	if trackID >= len(TrackSignalHeightOffset) {
		return
	}
	if s.Zoom > 1 { // paintSignal: no signals beyond zoom 1
		return
	}
	rot := int((s.Rotation + rotation) & 3)
	height := int(baseZ) * 4
	if trackSeq == 0 {
		if o, ok := objOf(left.Object); ok {
			s.paintSignalSide(left, o, false, trackID, rot, height, sprite)
		}
	}
	if !trackIsLast {
		return
	}
	if o, ok := objOf(right.Object); ok {
		s.paintSignalSide(right, o, true, trackID, rot, height, sprite)
	}
}
