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
	// Zoom is the render target's zoomLevel (0-3); SpriteFor looks an image
	// up for the hasZoomSprites recursion (needed when Zoom > 0).
	Zoom        int
	SpriteFor   func(id int) *gfx8.Sprite
	quadrants   [MaxQuadrants]*Struct
	back, front int
	last        *Struct
	head        *Struct

	trackRoad    [5]*Struct // _trackRoadPaintStructs, by priority
	trackRoadAdd [2]*Struct // _trackRoadAdditionsPaintStructs, by priority

	// Road merging (PaintRoad.cpp): set by merge pieces, drawn by FinalisePaintRoad.
	roadExits       uint8
	mergeRoadImage  int // absolute image of the merge set's first sprite
	mergeRoadHeight int
	mergeRoadPal    []uint8
	// mergeRoadStreetlight is setMergeRoadStreetlight's style (0: none) and
	// streetlightImage the StreetLight object's image it uses.
	mergeRoadStreetlight uint8
	streetlightImage     int

	// Slabs for paint structs and attachments (reused across frames by Reset).
	structs         []Struct
	tunnels         [4][]tunnelEntry
	bridge          BridgeEntry
	support         SupportHeight
	segments        [9]SupportHeight
	pendingSegments uint16
	waterHeight     int
	surfaceHeight   int
	surfaceSlope    uint8
	waterHeight2    int
	attachedSlab    []Attached
	chunks          [][]Struct
	attachedChunks  [][]Attached
}

// NewSession starts a session for a render target covering viewport rectangle
// (x, y, w, h).
func NewSession(rotation uint8, x, y, w, h int) *Session {
	return &Session{Rotation: rotation, rtX: x, rtY: y, rtW: w, rtH: h, back: noBackIndex, bridge: BridgeEntry{Height: -1}}
}

// Colour returns an image of sp remapped to a company colour (ImageId(index,
// primary colour)): remap nil means no colour, a plain transparent copy.
func Colour(sp *gfx8.Sprite, remap []uint8) Image {
	if remap == nil {
		return Image{Sprite: sp, Op: gfx8.OpTransparent}
	}
	return Image{Sprite: sp, Op: gfx8.OpTransparent | gfx8.OpSrc, PalMap: remap}
}

// Reset prepares the session for a new frame, keeping its allocations (the
// struct and attachment slabs) so a frame allocates nothing in steady state.
func (s *Session) Reset(rotation uint8, x, y, w, h int) {
	structs, attached, tunnels := s.structs[:0], s.attachedSlab[:0], s.tunnels
	zoom, spriteFor := s.Zoom, s.SpriteFor
	*s = Session{Rotation: rotation, rtX: x, rtY: y, rtW: w, rtH: h, back: noBackIndex, Zoom: zoom, SpriteFor: spriteFor,
		structs: structs, attachedSlab: attached, tunnels: tunnels, bridge: BridgeEntry{Height: -1}}
	s.ResetTunnels()
}

// newStruct returns a zeroed struct from the slab (upstream allocatePaintStruct).
// The slab grows in chunks; earlier chunks stay valid because a chunk is never
// reallocated once handed out.
func (s *Session) newStruct() *Struct {
	if len(s.structs) == cap(s.structs) {
		s.chunks = append(s.chunks, s.structs)
		s.structs = make([]Struct, 0, max(1024, 2*cap(s.structs)))
	}
	s.structs = append(s.structs, Struct{})
	return &s.structs[len(s.structs)-1]
}

func (s *Session) newAttached() *Attached {
	if len(s.attachedSlab) == cap(s.attachedSlab) {
		s.attachedChunks = append(s.attachedChunks, s.attachedSlab)
		s.attachedSlab = make([]Attached, 0, max(256, 2*cap(s.attachedSlab)))
	}
	s.attachedSlab = append(s.attachedSlab, Attached{})
	return &s.attachedSlab[len(s.attachedSlab)-1]
}

// SetSpritePosition sets the world position painters' offsets are relative to.
func (s *Session) SetSpritePosition(x, y int) { s.spriteX, s.spriteY = x, y }

// SupportHeight is the session's general support height (Paint.h _support:
// height, slope); 0xFFFF height = nothing to support from.
type SupportHeight struct {
	Height int
	Slope  uint8
}

// SetGeneralSupportHeight is setGeneralSupportHeight.
func (s *Session) SetGeneralSupportHeight(height int, slope uint8) {
	s.support = SupportHeight{height, slope}
}

// SegmentsAll is SegmentFlags::all; segment s is flag 1 << s (kSegmentOffsets).
const SegmentsAll = 0x1FF

// SetSegmentsSupportHeight is setSegmentsSupportHeight / setSegmentSupportHeight:
// the slope is only stored when the height is not 0xFFFF.
func (s *Session) SetSegmentsSupportHeight(segments uint16, height int, slope uint8) {
	for i := range s.segments {
		if segments&(1<<i) != 0 {
			s.segments[i].Height = height
			if height != 0xFFFF {
				s.segments[i].Slope = slope
			}
		}
	}
}

// SetSurfaceSegments is the end of paintSurface: every segment from
// kSlopeToSegmentsSupportHeights[slope] (heights as stored, no base added).
func (s *Session) SetSurfaceSegments(slope uint8) {
	for i, sh := range SlopeToSegmentsSupportHeights[slope&0x1F] {
		s.SetSegmentsSupportHeight(1<<i, sh.Height, sh.Slope)
	}
}

// AddPendingSegments ORs segments into the pending mask (session 525CF8).
func (s *Session) AddPendingSegments(segments uint16) { s.pendingSegments |= segments }

// FlushPendingSegments is the end of paintTileElementsEndLoop: the pending
// segments become unsupported (0xFFFF) and the mask is cleared.
func (s *Session) FlushPendingSegments() {
	if s.pendingSegments != 0 {
		s.SetSegmentsSupportHeight(s.pendingSegments, 0xFFFF, 0)
		s.pendingSegments = 0
	}
}

// SetWaterHeight is setWaterHeight (paintSurface, tiles with water only; it is
// not reset per tile upstream).
func (s *Session) SetWaterHeight(h int) { s.waterHeight = h }

// blockSupports is what 2×2 objects do on every footprint tile:
// setSegmentsSupportHeight(all, 0xFFFF, 0) and setGeneralSupportHeight(0xFFFF, 0).
func (s *Session) blockSupports() {
	s.SetSegmentsSupportHeight(SegmentsAll, 0xFFFF, 0)
	s.SetGeneralSupportHeight(0xFFFF, 0)
}

// GeneralSupportHeight is getGeneralSupportHeight.
func (s *Session) GeneralSupportHeight() SupportHeight { return s.support }

// SetSurface records what paintSurface tells the session for bridge shadows
// (setSurfaceHeight / setSurfaceSlope) and the tile's water (setWaterHeight2).
func (s *Session) SetSurface(height int, slope uint8, waterHeight2 int) {
	s.surfaceHeight, s.surfaceSlope, s.waterHeight2 = height, slope, waterHeight2
}

// BridgeEntry is Paint.h BridgeEntry: set by a track/road piece with a bridge,
// painted and cleared in paintTileElementsEndLoop. Height -1 = empty
// (kNullBridgeEntry). Remap is bridgeColoursBaseImageId's colour (the owner's).
type BridgeEntry struct {
	Height        int
	SubType       uint8
	EdgesQuarters uint8 // edges << 4 | quarters
	ObjectID      uint8
	Remap         []uint8
}

// setBridge is the hasBridge block of paintTrackPP* / paintRoadPCP: a new
// entry whose edges/quarters are ORed with the current one (merging pieces).
func (s *Session) setBridge(height int, subType, edges, quarters, objectID uint8, remap []uint8) {
	eq := edges<<4 | quarters
	if s.bridge.Height != -1 {
		eq |= s.bridge.EdgesQuarters
	}
	s.bridge = BridgeEntry{Height: height, SubType: subType, EdgesQuarters: eq, ObjectID: objectID, Remap: remap}
}

// TakeBridge returns the pending bridge entry and resets it to kNullBridgeEntry
// (the end of paintTileElementsEndLoop's bridge block); ok is false when empty.
func (s *Session) TakeBridge() (BridgeEntry, bool) {
	b := s.bridge
	s.bridge = BridgeEntry{Height: -1}
	return b, b.Height != -1
}

// Tunnels — Paint.cpp resetTunnels / insertTunnel / insertTunnels. Per edge a
// list of (MicroZ height, tunnel) in insertion order; reading past the end
// gives upstream's {0xFF, 0xFF} terminator. Tunnel is the tunnel object's image
// base (one per tunnel object, so equality matches upstream's type compare).
type tunnelEntry struct{ height, tunnel int }

// ResetTunnels is resetTunnels, called once per tile (paintTileElementsSetup).
func (s *Session) ResetTunnels() {
	for i := range s.tunnels {
		s.tunnels[i] = s.tunnels[i][:0]
	}
}

// InsertTunnels is insertTunnels: an entry per edge whose height is not -1, at
// (tunnelHeight + height) / kMicroZStep; a repeat of the edge's last entry is dropped.
func (s *Session) InsertTunnels(heights [4]int16, height, tunnel int) {
	if tunnel < 0 {
		return // no tunnel object resolved (upstream: tunnel 0xFF)
	}
	for edge, th := range heights {
		if th == -1 {
			continue
		}
		e := tunnelEntry{(int(th) + height) / microZStep, tunnel}
		if n := len(s.tunnels[edge]); n > 0 && s.tunnels[edge][n-1] == e {
			continue
		}
		s.tunnels[edge] = append(s.tunnels[edge], e)
	}
}

func (s *Session) tunnelAt(edge, i int) tunnelEntry {
	if i < len(s.tunnels[edge]) {
		return s.tunnels[edge][i]
	}
	return tunnelEntry{0xFF, 0xFF}
}

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
	ps := s.newStruct()
	ps.Img, ps.VpX, ps.VpY, ps.Mins = img, vpX, vpY, mins
	ps.Maxs = Pos3{mins.X + rs.X, mins.Y + rs.Y, mins.Z + rs.Z}
	return ps
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
	a := s.newAttached()
	a.Img, a.X, a.Y, a.next = img, x, y, s.last.attached
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
			if s.Zoom > 0 { // drawImageMaskedZoom<zoom>
				zt := *rt
				zt.OriginX, zt.OriginY = s.rtX, s.rtY
				gfx8.DrawSpriteMaskedZoom(&zt, img.Sprite, img.Mask, s.SpriteFor, x, y, s.Zoom)
				return
			}
			gfx8.DrawSpriteMasked(rt, img.Sprite, img.Mask, x-s.rtX, y-s.rtY)
			return
		}
		if s.Zoom > 0 { // drawImagePaletteSet at the target's zoom: rt.x/y = the session's
			zt := *rt
			zt.OriginX, zt.OriginY = s.rtX, s.rtY
			gfx8.DrawSpriteAtZoom(&zt, img.Sprite, s.SpriteFor, x, y, s.Zoom, img.Op, img.PalMap, img.Noise)
			return
		}
		gfx8.DrawSpriteNoise(rt, img.Sprite, x-s.rtX, y-s.rtY, img.Op, img.PalMap, img.Noise)
	}
	attached := func(ps *Struct) {
		for a := ps.attached; a != nil; a = a.next {
			x, y := ps.VpX+a.X, ps.VpY+a.Y
			if s.Zoom != 0 { // drawAttachStruct: floor2(pos, 2) when zoomed out
				x, y = x&^1, y&^1
			}
			draw(a.Img, x, y)
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

// AddTrackRoadAddition is addToPlotListTrackRoadAddition (second set of lists).
func (s *Session) AddTrackRoadAddition(img Image, priority int, off, bbOff, bbSize Pos3) *Struct {
	s.last = nil
	ps := s.create(img, off, bbOff, bbSize)
	if ps != nil {
		s.last = ps
		ps.children = s.trackRoadAdd[priority]
		s.trackRoadAdd[priority] = ps
	}
	return ps
}

// FinaliseTrackRoadOrdering is finaliseTrackRoadOrdering → finaliseOrdering.
func (s *Session) FinaliseTrackRoadOrdering() { s.finaliseOrdering(s.trackRoad[:]) }

// FinaliseTrackRoadAdditionsOrdering is finaliseTrackRoadAdditionsOrdering.
func (s *Session) FinaliseTrackRoadAdditionsOrdering() { s.finaliseOrdering(s.trackRoadAdd[:]) }

// finaliseOrdering chains the priority lists in order as one struct, gives it
// the min/max XY bounds of the chain (getMinMaxXYBounding) and adds it to a quadrant.
func (s *Session) finaliseOrdering(lists []*Struct) {
	var route, lastEntry *Struct
	for i, e := range lists {
		if e == nil {
			continue
		}
		if lastEntry == nil {
			route = e
		} else {
			lastEntry.children = e
		}
		lastEntry = lastChild(e)
		lists[i] = nil
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
