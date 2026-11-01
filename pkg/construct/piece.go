// Package construct holds the construction window's piece logic, ported
// from OpenLoco's Ui/Windows/Construction/ConstructionTab.cpp.
package construct

// TrackPiece is the construction window's piece choice
// (ConstructionTab.cpp namespace TrackPiece).
const (
	Straight = iota
	LeftCurveVerySmall
	RightCurveVerySmall
	LeftCurveSmall
	RightCurveSmall
	LeftCurve
	RightCurve
	LeftCurveLarge
	RightCurveLarge
	SBendLeft
	SBendRight
	SBendToDualTrack
	SBendToSingleTrack
	Turnaround
	PieceNull = 0xFF
)

// TrackGradient is the gradient choice (namespace TrackGradient).
const (
	Level          = 0
	SlopeUp        = 2
	SteepSlopeUp   = 4
	SlopeDown      = 6
	SteepSlopeDown = 8
)

// TrackPieceID is a track id (World::Track::TrackId) with its rotation.
type TrackPieceID struct {
	ID       uint8
	Rotation uint8
}

// GetTrackPieceID is getTrackPieceId (ConstructionTab.cpp): the track id
// for a piece and gradient at the construction rotation (0–3 straight on,
// 4–7 and 8–11 the two diagonal sets, 12–15 diagonal); ok false when the
// piece can't be built there. One-shot pieces (index ≥ 9: s-bends and the
// turnaround) go back to straight after being built (constructTrack).
func GetTrackPieceID(piece, gradient, rotation uint8) (TrackPieceID, bool) {
	var id uint8
	switch piece {
	case Straight:
		if rotation >= 12 {
			id = 1
			if gradient != Level {
				return TrackPieceID{}, false
			}
		} else if rotation >= 8 {
			switch gradient {
			case Level:
				id = 27
			case SteepSlopeUp:
				id = 35
			case SteepSlopeDown:
				id = 37
			default:
				return TrackPieceID{}, false
			}
		} else if rotation >= 4 {
			switch gradient {
			case Level:
				id = 26
			case SteepSlopeUp:
				id = 34
			case SteepSlopeDown:
				id = 36
			default:
				return TrackPieceID{}, false
			}
		} else {
			switch gradient {
			case Level:
				id = 0
			case SlopeUp:
				id = 14
			case SlopeDown:
				id = 15
			case SteepSlopeUp:
				id = 16
			case SteepSlopeDown:
				id = 17
			default:
				return TrackPieceID{}, false
			}
		}
	case LeftCurveVerySmall, RightCurveVerySmall:
		if gradient != Level || rotation >= 12 {
			return TrackPieceID{}, false
		}
		ids := [3]uint8{2, 28, 29} // rotation 0–3, 4–7, 8–11
		if piece == RightCurveVerySmall {
			ids = [3]uint8{3, 30, 31}
		}
		id = ids[rotation/4]
	case LeftCurveSmall, RightCurveSmall:
		if rotation >= 4 {
			return TrackPieceID{}, false
		}
		var r uint8
		if piece == RightCurveSmall {
			r = 1
		}
		switch gradient {
		case Level:
			id = 4 + r
		case SlopeUp:
			id = 18 + r
		case SlopeDown:
			id = 20 + r
		case SteepSlopeUp:
			id = 22 + r
		case SteepSlopeDown:
			id = 24 + r
		default:
			return TrackPieceID{}, false
		}
	case LeftCurve, RightCurve:
		if rotation >= 4 || gradient != Level {
			return TrackPieceID{}, false
		}
		id = 6
		if piece == RightCurve {
			id = 7
		}
	case LeftCurveLarge, RightCurveLarge:
		if gradient != Level {
			return TrackPieceID{}, false
		}
		var r uint8
		if piece == RightCurveLarge {
			r = 1
		}
		switch {
		case rotation >= 12:
			id = 10 + r
		case rotation >= 4:
			return TrackPieceID{}, false
		default:
			id = 8 + r
		}
	case SBendLeft:
		if gradient != Level || rotation >= 12 || (rotation >= 4 && rotation < 8) {
			return TrackPieceID{}, false
		}
		id = 12
		if rotation >= 8 {
			id = 33
		}
	case SBendRight:
		if gradient != Level || rotation >= 8 {
			return TrackPieceID{}, false
		}
		id = 13
		if rotation >= 4 {
			id = 32
		}
	case SBendToDualTrack:
		if gradient != Level || rotation >= 8 {
			return TrackPieceID{}, false
		}
		id = 38
		if rotation >= 4 {
			id = 40
		}
	case SBendToSingleTrack:
		if gradient != Level || rotation >= 12 || (rotation >= 4 && rotation < 8) {
			return TrackPieceID{}, false
		}
		id = 39
		if rotation >= 8 {
			id = 41
		}
	case Turnaround:
		if gradient != Level || rotation >= 12 || (rotation >= 4 && rotation < 8) {
			return TrackPieceID{}, false
		}
		id = 42
		if rotation >= 8 {
			id = 43
		}
	default:
		return TrackPieceID{}, false
	}
	if rotation < 12 {
		rotation &= 3
	}
	return TrackPieceID{ID: id, Rotation: rotation}, true
}

// GetRoadPieceID is getRoadPieceId (ConstructionTab.cpp): the road id
// (World::Track::RoadId) for a piece and gradient at the construction
// rotation. Roads have no diagonals, normal or large curves, or s-bends; the
// turnaround (one-shot) is built at any heading below 12.
func GetRoadPieceID(piece, gradient, rotation uint8) (TrackPieceID, bool) {
	var id uint8
	switch piece {
	case Straight:
		if rotation >= 4 {
			return TrackPieceID{}, false
		}
		switch gradient {
		case Level:
			id = 0
		case SlopeUp:
			id = 5
		case SlopeDown:
			id = 6
		case SteepSlopeUp:
			id = 7
		case SteepSlopeDown:
			id = 8
		default:
			return TrackPieceID{}, false
		}
	case LeftCurveVerySmall, RightCurveVerySmall, LeftCurveSmall, RightCurveSmall:
		if gradient != Level || rotation >= 4 {
			return TrackPieceID{}, false
		}
		id = map[uint8]uint8{LeftCurveVerySmall: 1, RightCurveVerySmall: 2, LeftCurveSmall: 3, RightCurveSmall: 4}[piece]
	case Turnaround:
		if gradient != Level || rotation >= 12 {
			return TrackPieceID{}, false
		}
		id = 9
	default: // normal and large curves, s-bends
		return TrackPieceID{}, false
	}
	if rotation < 12 {
		rotation &= 3
	}
	return TrackPieceID{ID: id, Rotation: rotation}, true
}

// Construction button images (ImageIds.h) for the dual-track pair.
const (
	ImgSBendDualTrackLeft      = 2355
	ImgSBendDualTrackRight     = 2356
	ImgSBendToSingleTrackLeft  = 2357
	ImgSBendToSingleTrackRight = 2358
	ImgRightTurnaround         = 2359
	ImgLeftTurnaround          = 2360
)

// DualTrackButtons is activateSelectedTrackWidgets for a one-sided track's
// s_bend_dual_track_left/right buttons at a construction rotation: each
// button's image and the piece it chooses (onMouseDown reads the image).
// Heading straight or diagonal: the two dual-track s-bends (both choose
// s_bend_to_dual_track, as upstream); at 4–7: right turnaround and s-bend to
// single track (left image); at 8–11: s-bend to single track (right image)
// and left turnaround.
func DualTrackButtons(rotation uint8) (leftImg, rightImg int, leftPiece, rightPiece uint8) {
	switch {
	case rotation >= 4 && rotation < 8:
		return ImgRightTurnaround, ImgSBendToSingleTrackLeft, Turnaround, SBendToSingleTrack
	case rotation >= 8 && rotation < 12:
		return ImgSBendToSingleTrackRight, ImgLeftTurnaround, SBendToSingleTrack, Turnaround
	}
	return ImgSBendDualTrackLeft, ImgSBendDualTrackRight, SBendToDualTrack, SBendToDualTrack
}
