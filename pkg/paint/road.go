package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Roads — a port of Paint/PaintRoad.cpp paintRoad (Style02::paintRoadPP
// standard and multi-tile merge, Style1::paintRoadPP) and finalisePaintRoad,
// with the tables generated from upstream (roaddata_gen.go). PARTIAL: company
// colour remap (needs the saved company colours), bridges and
// tunnels (paintRoadPCP), level crossings, road extras, height markers and
// one-way arrows are not ported yet.

// RoadEl is what paintRoad reads from a road element.
type RoadEl struct {
	RoadID, SequenceIndex, Rotation uint8
	BaseZ                           uint8   // SmallZ
	Remap                           []uint8 // company colour of elRoad.owner() (nil: none)
	Tunnel                          int     // roadObj->tunnel's image base (-1: none)
	HasBridge                       bool
	Bridge                          uint8 // elRoad.bridge(): bridge object slot
	// Streetlight is elRoad.streetLightStyle() (0: none) when the road has no
	// level crossing, signal or station element (zero otherwise);
	// StreetlightImage is the StreetLight object's image.
	Streetlight      uint8
	StreetlightImage int
	// Mods are the road object's two mods (TrackMod: the element's hasMod,
	// the road extra's image and paint style); AnyRoadTypeCompatible is the
	// road object's flag (then no mods are drawn).
	Mods                  [2]TrackMod
	AnyRoadTypeCompatible bool
	// Crossing is the level crossing on this road (nil: none): its object's
	// image and animation, the element's frame (unk6l) and the scenario ticks.
	Crossing *LevelCrossing
}

// LevelCrossing is what paintLevelCrossing reads.
type LevelCrossing struct {
	Image                                       int
	AnimationSpeed, ClosingFrames, ClosedFrames uint8
	Frame                                       uint8 // elRoad.unk6l(): 0 open, 15 animating
	Ticks                                       uint32
}

// paintLevelCrossing (0x00475DDF): four images (the gates at the corners)
// for the crossing's frame (frame 15 animates through the closing frames by
// the ticks), the fourth's box depending on the rotation while animating.
func (s *Session) paintLevelCrossing(el RoadEl, rot int, sprite func(int) *gfx8.Sprite) {
	c := el.Crossing
	frame := uint32(c.Frame)
	if frame == 15 {
		frame = (c.Ticks/(1<<c.AnimationSpeed))&(uint32(c.ClosingFrames)-1) + uint32(c.ClosedFrames) + 1
	}
	img0 := c.Image + (rot&1)*4 + int(frame)*8
	h := int(el.BaseZ) * 4
	ho := Pos3{0, 0, h}
	add := func(i int, bbOff, bbSize Pos3) {
		if sp := sprite(img0 + i); sp != nil {
			s.AddAsParent(Colour(sp, el.Remap), ho, Pos3{bbOff.X, bbOff.Y, bbOff.Z + h}, bbSize) // addToPlotList4FD150
		}
	}
	add(0, Pos3{2, 2, 1}, Pos3{1, 1, 8})
	add(1, Pos3{2, 30, 1}, Pos3{1, 1, 8})
	add(2, Pos3{30, 2, 1}, Pos3{1, 1, 8})
	switch {
	case c.Frame != 15:
		add(3, Pos3{30, 30, 1}, Pos3{1, 1, 8})
	case rot&1 != 0:
		add(3, Pos3{17, 30, 1}, Pos3{14, 1, 8})
	default:
		add(3, Pos3{30, 17, 1}, Pos3{1, 14, 8})
	}
}

// kStreetlightOffsets / BoundingBoxOffsets / Size (PaintRoad.cpp), by edge.
var (
	streetlightOffsets   = [4]Pos3{{2, 15, 0}, {15, 28, 0}, {28, 15, 0}, {15, 2, 0}}
	streetlightBBOffsets = [4]Pos3{{2, 15, 6}, {15, 28, 6}, {28, 15, 6}, {15, 2, 6}}
	streetlightBBSize    = Pos3{1, 1, 6}
)

// paintRoadStreetlight: the style's image for this edge (Streetlight::ImageIds
// kStyle<n><dir> = style·4 + edge) at baseHeight + streetlightHeight.
func (s *Session) paintRoadStreetlight(image, baseHeight, style, streetlightHeight, edge int, sprite func(int) *gfx8.Sprite) {
	sp := sprite(image + style*4 + edge)
	if sp == nil {
		return
	}
	h := baseHeight + streetlightHeight
	o, b := streetlightOffsets[edge], streetlightBBOffsets[edge]
	s.AddAsParent(Image{Sprite: sp, Op: gfx8.OpFor(sp, false, false)}, Pos3{o.X, o.Y, o.Z + h}, Pos3{b.X, b.Y, b.Z + h}, streetlightBBSize)
}

var (
	roadExitsToMergeID = [16]int{0, 0, 1, 5, 0, 0, 2, 8, 1, 4, 1, 7, 3, 6, 9, 10}
	roadMergeBBOff     = [11]Pos3{{2, 5, 0}, {5, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}}
	roadMergeBBSize    = [11]Pos3{{28, 22, 0}, {22, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}}
)

// PaintRoad paints one road element at the current sprite position. image is
// the road object's image base, paintStyle its paint style (0, 1, 2).
func (s *Session) PaintRoad(el RoadEl, image int, paintStyle uint8, sprite func(int) *gfx8.Sprite) {
	s.paintRoadPiece(el, image, paintStyle, sprite)
	// paintRoad, after the piece: beyond zoom 1 nothing more; the level crossing.
	if s.Zoom > 1 {
		return
	}
	rot := int((s.Rotation + el.Rotation) & 3)
	if el.Crossing != nil {
		s.paintLevelCrossing(el, rot, sprite)
	}
	// The road extras (zoom 0, not for anyRoadTypeCompatible roads): each mod
	// the element has, AdditionStyle1's piece (the null piece — image 3 = 0 —
	// skipped) — mergeable into track/road addition priority 0, else as a
	// parent (addToPlotList4FD150). PARTIAL: the pieces' supports
	// (setAdditionSupport / paintSupports) and ghost mods.
	if s.Zoom > 0 || el.AnyRoadTypeCompatible {
		return
	}
	rid, seq := int(el.RoadID), int(el.SequenceIndex)
	h := Pos3{0, 0, int(el.BaseZ) * 4}
	for _, m := range el.Mods {
		if !m.Has || m.PaintStyle != 1 || rid >= len(RoadAdditionParts) || seq >= len(RoadAdditionParts[rid]) {
			continue
		}
		p := RoadAdditionParts[rid][seq]
		if p.ImageIDs[3] == 0 {
			continue
		}
		sp := sprite(m.Image + int(p.ImageIDs[rot]))
		if sp == nil {
			continue
		}
		img := Colour(sp, el.Remap)
		bbOff := Pos3{p.BBOffsets[rot].X, p.BBOffsets[rot].Y, p.BBOffsets[rot].Z + h.Z}
		if p.IsMergeable {
			s.AddTrackRoadAddition(img, 0, h, bbOff, p.BBSizes[rot])
		} else {
			s.AddAsParent(img, h, bbOff, p.BBSizes[rot])
		}
		if sup := p.Supports; sup != nil { // AdditionStyle1::paintSupport
			s.setAdditionSupport(h.Z+int(sup.Height), sup.Segments[rot], sup.ImageIDs[rot][:], true, sup.Frequencies[rot], m.Image, el.Remap)
		}
	}
}

// paintRoadPiece is paintRoad's piece part (paintRoadPP / PCP).
func (s *Session) paintRoadPiece(el RoadEl, image int, paintStyle uint8, sprite func(int) *gfx8.Sprite) {
	rid, seq := int(el.RoadID), int(el.SequenceIndex)
	if rid >= len(RoadCommonParts) || seq >= len(RoadCommonParts[rid]) {
		return
	}
	rot := int((s.Rotation + el.Rotation) & 3)
	height := int(el.BaseZ) * 4
	rpcp := RoadCommonParts[rid][seq]
	ho := Pos3{0, 0, height}
	bbOff := rpcp.BBOffsets[rot]
	bbOff.Z += height
	img := func(off uint32) Image {
		return Colour(sprite(image+int(off)), el.Remap)
	}
	// paintRoadPCP runs after every piece (merge pieces too); its only effect
	// ported here is insertTunnels, so it can go first.
	if el.HasBridge {
		s.setBridge(height, rpcp.BridgeType[rot], rpcp.BridgeEdges[rot], rpcp.BridgeQuarters[rot], el.Bridge, el.Remap)
	}
	s.InsertTunnels(rpcp.TunnelHeights[rot], height, el.Tunnel)
	s.AddPendingSegments(rpcp.Segments[rot])
	switch paintStyle {
	case 0, 2:
		parts := RoadStyle0Parts
		if paintStyle == 2 {
			parts = RoadStyle2Parts
		}
		if seq >= len(parts[rid]) {
			return
		}
		rpp := parts[rid][seq]
		if rpp.MergeType[0] != 0 {
			// paintRoadPPMultiTileMerge (not hit test): collect exits; drawn in FinalisePaintRoad.
			s.roadExits |= rpcp.BridgeEdges[rot]
			s.mergeRoadImage = image + int(RoadMergeBaseImage[rpp.MergeType[rot]-1])
			s.mergeRoadHeight = height
			s.mergeRoadPal = el.Remap // the merge base image carries the road's colour
			if s.Zoom == 0 && el.Streetlight != 0 {
				s.mergeRoadStreetlight, s.streetlightImage = el.Streetlight, el.StreetlightImage
			}
			return
		}
		if i := img(rpp.ImageIndexOffsets[rot]); i.Sprite != nil {
			s.AddTrackRoad(i, 2, ho, bbOff, rpcp.BBSizes[rot])
		}
		if s.Zoom == 0 && el.Streetlight != 0 { // paintRoadStreetlights
			for edge, sh := range rpp.StreetlightHeights[rot] {
				if sh != -1 { // kNoStreetlight
					s.paintRoadStreetlight(el.StreetlightImage, height, int(el.Streetlight)-1, int(sh), edge, sprite)
				}
			}
		}
	case 1:
		if seq >= len(RoadStyle1Parts[rid]) {
			return
		}
		rpp := RoadStyle1Parts[rid][seq]
		for layer, prio := range [3]int{0, 1, 3} {
			if i := img(rpp.ImageIndexOffsets[rot][layer]); i.Sprite != nil {
				s.AddTrackRoad(i, prio, ho, bbOff, rpcp.BBSizes[rot])
			}
		}
	}
}

// FinalisePaintRoad is finalisePaintRoad: when merge pieces set exits on this
// tile, draw the one merged sprite (straight, curve, junction, crossroads).
func (s *Session) FinalisePaintRoad(sprite func(int) *gfx8.Sprite) {
	if s.roadExits == 0 {
		return
	}
	exits := s.roadExits & 0xF
	id := roadExitsToMergeID[exits]
	light := s.mergeRoadStreetlight
	s.roadExits, s.mergeRoadStreetlight = 0, 0
	ho := Pos3{0, 0, s.mergeRoadHeight}
	bb := roadMergeBBOff[id]
	bb.Z += s.mergeRoadHeight
	if sp := sprite(s.mergeRoadImage + id); sp != nil {
		s.AddTrackRoad(Colour(sp, s.mergeRoadPal), 2, ho, bb, roadMergeBBSize[id])
	}
	if light != 0 { // a streetlight on every edge without an exit
		for edge := range 4 {
			if exits&(1<<edge) == 0 {
				s.paintRoadStreetlight(s.streetlightImage, s.mergeRoadHeight, int(light)-1, 0, edge, sprite)
			}
		}
	}
}
