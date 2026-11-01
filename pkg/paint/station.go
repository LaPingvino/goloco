package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Train stations — a port of Paint/PaintTrainStation.cpp (paint style 0:
// straight and diagonal track). PARTIAL: cargo on the platforms
// (paintStationCargo needs the station's cargo stats) and ghosts. Company colour
// (StationEl.Remap) and the translucent canopy layer (StationEl.Translucent) are
// ported; road stations below, airports and docks in industry.go.

// StationEl is what paintTrainStation reads from a station element and the
// track element before it.
type StationEl struct {
	Rotation, BaseZ   uint8
	TrackID, TrackSeq uint8
	Remap             []uint8 // company colour of the track's owner (elTrack->owner())
	Translucent       []uint8 // palette map of the canopy's translucent colour (nil: skip)
}

// attachTranslucent is session.attachToPrevious(imageTranslucentBase + off):
// a blended image without primary colour remaps what it covers.
func (s *Session) attachTranslucent(sp *gfx8.Sprite, pal []uint8) {
	if sp == nil || pal == nil {
		return
	}
	s.AttachToPrevious(Image{Sprite: sp, Op: gfx8.OpFor(sp, false, true), PalMap: pal}, 0, 0)
}

type stationDetail struct {
	img         int
	bbOff, bbSz Pos3
}

var (
	stationNE = [3]stationDetail{
		{Station0StraightBackNE, Pos3{2, 2, 8}, Pos3{28, 4, 3}},
		{Station0StraightFrontNE, Pos3{2, 24, 8}, Pos3{28, 4, 3}},
		{Station0StraightCanopyNE, Pos3{2, 2, 26}, Pos3{28, 28, 1}},
	}
	stationSE = [3]stationDetail{
		{Station0StraightBackSE, Pos3{2, 2, 8}, Pos3{4, 28, 3}},
		{Station0StraightFrontSE, Pos3{24, 2, 8}, Pos3{4, 28, 3}},
		{Station0StraightCanopySE, Pos3{2, 2, 26}, Pos3{28, 28, 1}},
	}
)

// PaintTrainStation paints a style-0 train station element at the current
// sprite position; base is stationObj->imageOffsets[sequenceIndex].
func (s *Session) PaintTrainStation(el StationEl, base int, sprite func(int) *gfx8.Sprite) {
	rot := int((s.Rotation + el.Rotation) & 3)
	ho := Pos3{0, 0, int(el.BaseZ) * 4}
	img := func(off int) Image {
		return Colour(sprite(base+off), el.Remap)
	}
	at := func(p Pos3) Pos3 { return Pos3{p.X, p.Y, p.Z + ho.Z} }
	switch el.TrackID {
	case 0: // paintTrainStationStyle0StraightTrack
		d := stationNE
		if rot&1 == 1 {
			d = stationSE
		}
		for _, part := range d[:2] {
			if i := img(part.img); i.Sprite != nil {
				s.AddAsParent(i, ho, at(part.bbOff), part.bbSz) // addToPlotList4FD150
			}
		}
		if i := img(d[2].img); i.Sprite != nil {
			s.AddTrackRoadAddition(i, 1, ho, at(d[2].bbOff), d[2].bbSz)
		}
		tr := Station0StraightCanopyTranslucentNE
		if rot&1 == 1 {
			tr = Station0StraightCanopyTranslucentSE
		}
		s.attachTranslucent(sprite(base+tr), el.Translucent)
		s.AddPendingSegments(SegmentsAll) // set525CF8(| all)
	case 1: // paintTrainStationStyle0DiagonalTrack
		seq := int(el.TrackSeq)
		if rot&2 != 0 {
			seq = 3 - seq
		}
		add := func(off int, bbOff, bbSz Pos3) {
			if i := img(off); i.Sprite != nil {
				s.AddAsParent(i, ho, at(bbOff), bbSz)
			}
		}
		if rot&1 == 1 {
			switch seq {
			case 1:
				add(Station0DiagonalSE1, Pos3{28, 34, 8}, Pos3{2, 2, 3})
			case 2:
				add(Station0DiagonalSE2, Pos3{34, 28, 8}, Pos3{2, 2, 3})
			case 3:
				if i := img(Station0DiagonalSE3); i.Sprite != nil {
					s.AddTrackRoadAddition(i, 1, ho, at(Pos3{0, 0, 26}), Pos3{30, 30, 1})
				}
				s.attachTranslucent(sprite(base+Station0DiagonalCanopyTranslucentSE3), el.Translucent)
			}
		} else {
			switch seq {
			case 0:
				add(Station0DiagonalNE0, Pos3{2, 2, 8}, Pos3{2, 2, 3})
			case 1:
				add(Station0DiagonalNE1, Pos3{6, 6, 8}, Pos3{2, 2, 11})
				add(Station0DiagonalCanopyNE1, Pos3{6, 6, 26}, Pos3{2, 2, 1})
				s.attachTranslucent(sprite(base+Station0DiagonalCanopyTranslucentNE1), el.Translucent)
			case 3:
				add(Station0DiagonalNE3, Pos3{2, 2, 8}, Pos3{2, 2, 3})
			}
		}
	}
}

// Road stations — a port of Paint/PaintRoadStation.cpp (paint style 0, straight
// road only: upstream's "simplified copy of train station paint"). Same PARTIAL
// set as train stations: no cargo, no ghosts.
// Company colour by station->owner (remap). Image indices are RoadStation::ImageIds::Style0 (0..7).
var (
	roadStationNE = [3]stationDetail{
		{0, Pos3{2, 2, 8}, Pos3{28, 4, 3}},   // straightBackNE
		{1, Pos3{2, 24, 8}, Pos3{28, 4, 3}},  // straightFrontNE
		{2, Pos3{2, 2, 26}, Pos3{28, 28, 1}}, // straightCanopyNE
	}
	roadStationSE = [3]stationDetail{
		{4, Pos3{2, 2, 8}, Pos3{4, 28, 3}},   // straightBackSE
		{5, Pos3{24, 2, 8}, Pos3{4, 28, 3}},  // straightFrontSE
		{6, Pos3{2, 2, 26}, Pos3{28, 28, 1}}, // straightCanopySE
	}
)

// PaintRoadStation paints a style-0 road station element; base is
// stationObj->imageOffsets[1] (upstream always uses index 1), roadID the road
// element before it.
func (s *Session) PaintRoadStation(rotation, baseZ, roadID uint8, base int, remap, translucent []uint8, sprite func(int) *gfx8.Sprite) {
	if roadID != 0 {
		return // paintRoadStationStyle0: only straight road
	}
	rot := (s.Rotation + rotation) & 3
	ho := Pos3{0, 0, int(baseZ) * 4}
	d := roadStationNE
	if rot&1 == 1 {
		d = roadStationSE
	}
	for _, part := range d[:2] {
		if sp := sprite(base + part.img); sp != nil {
			s.AddAsParent(Colour(sp, remap), ho, Pos3{part.bbOff.X, part.bbOff.Y, part.bbOff.Z + ho.Z}, part.bbSz)
		}
	}
	if sp := sprite(base + d[2].img); sp != nil {
		s.AddTrackRoadAddition(Colour(sp, remap), 1, ho, Pos3{d[2].bbOff.X, d[2].bbOff.Y, d[2].bbOff.Z + ho.Z}, d[2].bbSz)
	}
	tr := 3 // straightCanopyTranslucentNE
	if rot&1 == 1 {
		tr = 7 // straightCanopyTranslucentSE
	}
	s.attachTranslucent(sprite(base+tr), translucent)
	s.AddPendingSegments(SegmentsAll) // set525CF8(| all)
}
