// Package paint is a port of OpenLoco's paint session (Paint/Paint.cpp): painters
// add sprites with world-space bounding boxes, the session buckets them into
// quadrants, sorts them by bounding box (arrangeStructs) and draws them in that
// order into an 8-bit render target (pkg/gfx8). Ported line for line; the
// track/road priority lists, string structs, culling and interaction lookup are
// not ported yet.
package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// MaxQuadrants is kMaxPaintQuadrants (Paint.h).
const MaxQuadrants = 1024

// TileSize is World::kTileSize.
const TileSize = 32

// Quadrant flags (Paint.h QuadrantFlags).
const (
	qNone       uint8 = 0
	qPending    uint8 = 1 << 0
	qNeighbour  uint8 = 1 << 1
	qOutside    uint8 = 1 << 7
	noBackIndex       = int(^uint32(0) >> 1)
)

// Pos3 is a world position (World::Pos3).
type Pos3 struct{ X, Y, Z int }

// Image is what a struct draws: a decoded sprite and how to blit it (the
// resolved ImageId: blend op and palette map).
type Image struct {
	Sprite *gfx8.Sprite
	Op     gfx8.BlendOp
	PalMap []uint8
	Mask   *gfx8.Sprite // drawn masked (PaintStructFlags::hasMaskedImage) when set
	Noise  *gfx8.Sprite // noise image (ImageId withNoiseMask), non-RLE sprites only
}

// Struct is a PaintStruct.
type Struct struct {
	Img        Image
	VpX, VpY   int
	Mins, Maxs Pos3
	quadrant   int
	qFlags     uint8
	next       *Struct // nextQuadrantPS
	children   *Struct
	attached   *Attached
}

// Attached is an AttachedPaintStruct: an image drawn at its parent's position + offset.
type Attached struct {
	Img  Image
	X, Y int
	next *Attached
}

// Session is a PaintSession for one view.
type Session struct {
	Rotation           uint8
	spriteX, spriteY   int // world position of the tile being painted
	rtX, rtY, rtW, rtH int // render target in viewport coordinates
	quadrants          [MaxQuadrants]*Struct
	back, front        int
	last               *Struct
	head               *Struct

	trackRoad [5]*Struct // _trackRoadPaintStructs, by priority

	// Road merging (PaintRoad.cpp): set by merge pieces, drawn by FinalisePaintRoad.
	roadExits       uint8
	mergeRoadImage  int // absolute image of the merge set's first sprite
	mergeRoadHeight int
	mergeRoadPal    []uint8
}

// NewSession starts a session for a render target covering viewport rectangle
// (x, y, w, h).
func NewSession(rotation uint8, x, y, w, h int) *Session {
	return &Session{Rotation: rotation, rtX: x, rtY: y, rtW: w, rtH: h, back: noBackIndex}
}

// SetSpritePosition sets the world position painters' offsets are relative to.
func (s *Session) SetSpritePosition(x, y int) { s.spriteX, s.spriteY = x, y }

// rotate is Math::Vector::rotate.
func rotate(x, y int, dir uint8) (int, int) {
	switch dir & 3 {
	case 1:
		return y, -x
	case 2:
		return -x, -y
	case 3:
		return -y, x
	}
	return x, y
}

// directionFlipXAxis: { 0, 3, 2, 1 }.
func directionFlipXAxis(d uint8) uint8 { return (d * 3) % 4 }

// GameToScreen is World::gameToScreen.
func GameToScreen(p Pos3, rotation uint8) (int, int) {
	x, y := rotate(p.X, p.Y, rotation)
	return y - x, ((y + x) >> 1) - p.Z
}

func rotateBoundBoxSize(sz Pos3, rotation uint8) Pos3 {
	out := sz
	switch rotation {
	case 0:
		out.X--
		out.Y--
	case 1:
		out.X--
		out.X, out.Y = rotate(out.X, out.Y, 3)
	case 2:
		out.X, out.Y = rotate(out.X, out.Y, 2)
	case 3:
		out.Y--
		out.X, out.Y = rotate(out.X, out.Y, 1)
	}
	return out
}

func (s *Session) imageWithinRT(vpX, vpY int, sp *gfx8.Sprite) bool {
	left, bottom := vpX+sp.XOffset, vpY+sp.YOffset
	right, top := left+sp.Width, bottom+sp.Height
	return !(right <= s.rtX || top <= s.rtY || left >= s.rtX+s.rtW || bottom >= s.rtY+s.rtH)
}

// createNormalPaintStruct (Paint.cpp).
func (s *Session) create(img Image, off, bbOff, bbSize Pos3) *Struct {
	if img.Sprite == nil {
		return nil
	}
	sw := directionFlipXAxis(s.Rotation)
	ox, oy := rotate(off.X, off.Y, sw)
	vpX, vpY := GameToScreen(Pos3{ox + s.spriteX, oy + s.spriteY, off.Z}, s.Rotation)
	if !s.imageWithinRT(vpX, vpY, img.Sprite) {
		return nil
	}
	bx, by := rotate(bbOff.X, bbOff.Y, sw)
	rs := rotateBoundBoxSize(bbSize, s.Rotation)
	mins := Pos3{bx + s.spriteX, by + s.spriteY, bbOff.Z}
	return &Struct{Img: img, VpX: vpX, VpY: vpY, Mins: mins,
		Maxs: Pos3{mins.X + rs.X, mins.Y + rs.Y, mins.Z + rs.Z}}
}

// remapPositionToQuadrant (Paint.cpp).
func remapPositionToQuadrant(ps *Struct, rotation uint8) int {
	const rangeMax = MaxQuadrants * TileSize
	const center = rangeMax / 2
	x, y := ps.Mins.X, ps.Mins.Y
	switch rotation & 3 {
	case 0:
		return x + y
	case 1:
		return (y - x) + center
	case 2:
		return -(y + x) + rangeMax
	case 3:
		return (x - y) + center
	}
	return 0
}

func (s *Session) addToQuadrant(ps *Struct) {
	i := remapPositionToQuadrant(ps, s.Rotation) / TileSize
	i = min(max(i, 0), MaxQuadrants-1)
	ps.quadrant = i
	ps.next = s.quadrants[i]
	s.quadrants[i] = ps
	s.back = min(s.back, i)
	s.front = max(s.front, i)
}

// AddAsParent is addToPlotListAsParent.
func (s *Session) AddAsParent(img Image, off, bbOff, bbSize Pos3) *Struct {
	s.last = nil
	ps := s.create(img, off, bbOff, bbSize)
	if ps != nil {
		s.last = ps
		s.addToQuadrant(ps)
	}
	return ps
}

// AddAsChild is addToPlotListAsChild.
func (s *Session) AddAsChild(img Image, off, bbOff, bbSize Pos3) *Struct {
	if s.last == nil {
		return s.AddAsParent(img, off, bbOff, bbSize)
	}
	ps := s.create(img, off, bbOff, bbSize)
	if ps == nil {
		return nil
	}
	s.last.children = ps
	s.last = ps
	return ps
}

// AttachToPrevious is attachToPrevious.
func (s *Session) AttachToPrevious(img Image, x, y int) *Attached {
	if s.last == nil {
		return nil
	}
	a := &Attached{Img: img, X: x, Y: y, next: s.last.attached}
	s.last.attached = a
	return a
}

// checkBoundingBox (Paint.cpp, per rotation).
func checkBoundingBox(r uint8, a, b *Struct) bool {
	i, c := a, b
	switch r {
	case 0:
		return i.Maxs.Z >= c.Mins.Z && i.Maxs.Y >= c.Mins.Y && i.Maxs.X >= c.Mins.X &&
			!(i.Mins.Z < c.Maxs.Z && i.Mins.Y < c.Maxs.Y && i.Mins.X < c.Maxs.X)
	case 1:
		return i.Maxs.Z >= c.Mins.Z && i.Maxs.Y >= c.Mins.Y && i.Maxs.X < c.Mins.X &&
			!(i.Mins.Z < c.Maxs.Z && i.Mins.Y < c.Maxs.Y && i.Mins.X >= c.Maxs.X)
	case 2:
		return i.Maxs.Z >= c.Mins.Z && i.Maxs.Y < c.Mins.Y && i.Maxs.X < c.Mins.X &&
			!(i.Mins.Z < c.Maxs.Z && i.Mins.Y >= c.Maxs.Y && i.Mins.X >= c.Maxs.X)
	case 3:
		return i.Maxs.Z >= c.Mins.Z && i.Maxs.Y < c.Mins.Y && i.Maxs.X >= c.Mins.X &&
			!(i.Mins.Z < c.Maxs.Z && i.Mins.Y >= c.Maxs.Y && i.Mins.X < c.Maxs.X)
	}
	return false
}

// arrangeHelper is arrangeStructsHelperRotation.
func arrangeHelper(psNext *Struct, qi int, flag uint8, r uint8) *Struct {
	var ps *Struct
	for {
		ps = psNext
		psNext = psNext.next
		if psNext == nil {
			return ps
		}
		if !(qi > psNext.quadrant) {
			break
		}
	}
	entry := ps
	tmp := ps
	for {
		ps = ps.next
		if ps == nil {
			break
		}
		switch {
		case ps.quadrant > qi+1:
			ps.qFlags = qOutside
		case ps.quadrant == qi+1:
			ps.qFlags = qNeighbour | qPending
		case ps.quadrant == qi:
			ps.qFlags = flag | qPending
		}
		if !(ps.quadrant <= qi+1) {
			break
		}
	}
	ps = tmp
	for {
		for {
			psNext = ps.next
			if psNext == nil || psNext.qFlags&qOutside != 0 {
				return entry
			}
			if psNext.qFlags&qPending != 0 {
				break
			}
			ps = psNext
		}
		psNext.qFlags &^= qPending
		tmp = ps
		initial := psNext
		for {
			ps = psNext
			psNext = psNext.next
			if psNext == nil || psNext.qFlags&qOutside != 0 {
				break
			}
			if psNext.qFlags&qNeighbour == 0 {
				continue
			}
			if checkBoundingBox(r, initial, psNext) {
				ps.next = psNext.next
				t2 := tmp.next
				tmp.next = psNext
				psNext.next = t2
				psNext = ps
			}
		}
		ps = tmp
	}
}

// Arrange is arrangeStructs: link the quadrants back to front, then sort.
func (s *Session) Arrange() {
	if s.back == noBackIndex {
		return
	}
	head := &Struct{}
	ps := head
	for qi := s.back; qi <= s.front; qi++ {
		if n := s.quadrants[qi]; n != nil {
			ps.next = n
			for n != nil {
				ps = n
				n = n.next
			}
		}
	}
	cache := arrangeHelper(head, s.back, qNeighbour, s.Rotation)
	for qi := s.back + 1; qi < s.front; qi++ {
		cache = arrangeHelper(cache, qi, qNone, s.Rotation)
	}
	s.head = head.next
}

// Order returns the structs in draw order (after Arrange), for tests.
func (s *Session) Order() []*Struct {
	var out []*Struct
	for ps := s.head; ps != nil; ps = ps.next {
		out = append(out, ps)
	}
	return out
}

// Draw is drawStructs: each struct, its child chain (with their attachments),
// then its own attachments. rt's (0,0) is viewport (s.rtX, s.rtY).
func (s *Session) Draw(rt *gfx8.RenderTarget) {
	draw := func(img Image, x, y int) {
		if img.Mask != nil {
			gfx8.DrawSpriteMasked(rt, img.Sprite, img.Mask, x-s.rtX, y-s.rtY)
			return
		}
		gfx8.DrawSpriteNoise(rt, img.Sprite, x-s.rtX, y-s.rtY, img.Op, img.PalMap, img.Noise)
	}
	attached := func(ps *Struct) {
		for a := ps.attached; a != nil; a = a.next {
			draw(a.Img, ps.VpX+a.X, ps.VpY+a.Y)
		}
	}
	for ps := s.head; ps != nil; ps = ps.next {
		draw(ps.Img, ps.VpX, ps.VpY)
		for c := ps.children; c != nil; c = c.children {
			draw(c.Img, c.VpX, c.VpY)
			attached(c)
		}
		attached(ps)
	}
}

// AddTrackRoad is addToPlotListTrackRoad: the struct goes into the priority
// list (newest first through children) instead of a quadrant.
func (s *Session) AddTrackRoad(img Image, priority int, off, bbOff, bbSize Pos3) *Struct {
	s.last = nil
	ps := s.create(img, off, bbOff, bbSize)
	if ps != nil {
		s.last = ps
		ps.children = s.trackRoad[priority]
		s.trackRoad[priority] = ps
	}
	return ps
}

func lastChild(ps *Struct) *Struct {
	for ps.children != nil {
		ps = ps.children
	}
	return ps
}

// FinaliseTrackRoadOrdering is finaliseTrackRoadOrdering → finaliseOrdering:
// chain the priority lists in order as one struct, give it the min/max XY
// bounds of the chain (getMinMaxXYBounding) and add it to a quadrant.
func (s *Session) FinaliseTrackRoadOrdering() {
	var route, lastEntry *Struct
	for i, e := range s.trackRoad {
		if e == nil {
			continue
		}
		if lastEntry == nil {
			route = e
		} else {
			lastEntry.children = e
		}
		lastEntry = lastChild(e)
		s.trackRoad[i] = nil
	}
	if route == nil {
		return
	}
	mins, maxs := route.Mins, route.Maxs
	for c := route.children; c != nil; c = c.children {
		switch s.Rotation {
		case 0:
			mins.X, mins.Y = min(mins.X, c.Mins.X), min(mins.Y, c.Mins.Y)
			maxs.X, maxs.Y = max(maxs.X, c.Maxs.X), max(maxs.Y, c.Maxs.Y)
		case 1:
			mins.X, mins.Y = max(mins.X, c.Mins.X), min(mins.Y, c.Mins.Y)
			maxs.X, maxs.Y = min(maxs.X, c.Maxs.X), max(maxs.Y, c.Maxs.Y)
		case 2:
			mins.X, mins.Y = max(mins.X, c.Mins.X), max(mins.Y, c.Mins.Y)
			maxs.X, maxs.Y = min(maxs.X, c.Maxs.X), min(maxs.Y, c.Maxs.Y)
		case 3:
			mins.X, mins.Y = min(mins.X, c.Mins.X), max(mins.Y, c.Mins.Y)
			maxs.X, maxs.Y = max(maxs.X, c.Maxs.X), min(maxs.Y, c.Maxs.Y)
		}
	}
	route.Mins, route.Maxs = mins, maxs
	s.addToQuadrant(route)
}
