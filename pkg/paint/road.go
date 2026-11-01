package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Roads — a port of Paint/PaintRoad.cpp paintRoad (Style02::paintRoadPP
// standard and multi-tile merge, Style1::paintRoadPP) and finalisePaintRoad,
// with the tables generated from upstream (roaddata_gen.go). PARTIAL: company
// colour remap (needs the saved company colours), streetlights, bridges and
// tunnels (paintRoadPCP), level crossings, road extras, height markers and
// one-way arrows are not ported yet.

// RoadEl is what paintRoad reads from a road element.
type RoadEl struct {
	RoadID, SequenceIndex, Rotation uint8
	BaseZ                           uint8   // SmallZ
	Remap                           []uint8 // company colour of elRoad.owner() (nil: none)
}

var (
	roadExitsToMergeID = [16]int{0, 0, 1, 5, 0, 0, 2, 8, 1, 4, 1, 7, 3, 6, 9, 10}
	roadMergeBBOff     = [11]Pos3{{2, 5, 0}, {5, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}, {2, 2, 0}}
	roadMergeBBSize    = [11]Pos3{{28, 22, 0}, {22, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}, {28, 28, 0}}
)

// PaintRoad paints one road element at the current sprite position. image is
// the road object's image base, paintStyle its paint style (0, 1, 2).
func (s *Session) PaintRoad(el RoadEl, image int, paintStyle uint8, sprite func(int) *gfx8.Sprite) {
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
			return
		}
		if i := img(rpp.ImageIndexOffsets[rot]); i.Sprite != nil {
			s.AddTrackRoad(i, 2, ho, bbOff, rpcp.BBSizes[rot])
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
	id := roadExitsToMergeID[s.roadExits&0xF]
	s.roadExits = 0
	ho := Pos3{0, 0, s.mergeRoadHeight}
	bb := roadMergeBBOff[id]
	bb.Z += s.mergeRoadHeight
	if sp := sprite(s.mergeRoadImage + id); sp != nil {
		s.AddTrackRoad(Colour(sp, s.mergeRoadPal), 2, ho, bb, roadMergeBBSize[id])
	}
}
