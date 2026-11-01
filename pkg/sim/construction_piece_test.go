package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/construct"
)

// The pieces construct.GetTrackPieceID gives for a head heading straight
// (rotation 0–3) can be built there and move the head: the s-bends, the
// s-bends to and from dual track, and the slopes on small curves (pieces
// goloco couldn't build before).
// OPEN: the turnaround — getTrackPieceId gives track 42 at rotation 0–3, but
// kTrackCoordinates has 42 begin at 4–7 and the button only shows at 4–11;
// goloco's CanPlaceTrackAtHead (rotationBegin == heading) is not upstream's
// check, which lives in the unported CreateTrack command (LOOP 3o).
func TestGetTrackPieceIDPiecesPlace(t *testing.T) {
	pieces := []struct {
		piece, gradient uint8
	}{
		{construct.SBendLeft, construct.Level}, {construct.SBendRight, construct.Level},
		{construct.SBendToDualTrack, construct.Level}, {construct.SBendToSingleTrack, construct.Level},
		{construct.LeftCurveSmall, construct.SlopeUp}, {construct.RightCurveSmall, construct.SteepSlopeDown},
	}
	for _, pc := range pieces {
		for rot := uint8(0); rot < 4; rot++ {
			p, ok := construct.GetTrackPieceID(pc.piece, pc.gradient, rot)
			if !ok {
				t.Errorf("piece %d gradient %d rot %d: no track id", pc.piece, pc.gradient, rot)
				continue
			}
			s := flatState(30, 64)
			s.ConsHead = ConstructionHead{Active: true, X: 15, Y: 15, BaseZ: 64, Rotation: rot}
			if !s.PlaceTrackAtHead(int(p.ID), 0) {
				t.Errorf("piece %d (track %d) rot %d: can't be built", pc.piece, p.ID, rot)
				continue
			}
			if h := s.ConsHead; h.X == 15 && h.Y == 15 && h.BaseZ == 64 && h.Rotation == rot {
				t.Errorf("piece %d (track %d) rot %d: the head didn't move", pc.piece, p.ID, rot)
			}
		}
	}
}

// The road pieces construct.GetRoadPieceID gives can be built from a head on
// flat ground in every rotation 0–3 and move it: the curves and the
// turnaround (which goloco couldn't build before).
func TestGetRoadPieceIDPiecesPlace(t *testing.T) {
	for _, piece := range []uint8{construct.Straight, construct.LeftCurveVerySmall, construct.RightCurveVerySmall,
		construct.LeftCurveSmall, construct.RightCurveSmall, construct.Turnaround} {
		for rot := uint8(0); rot < 4; rot++ {
			p, ok := construct.GetRoadPieceID(piece, construct.Level, rot)
			if !ok {
				t.Errorf("piece %d rot %d: no road id", piece, rot)
				continue
			}
			s := flatState(30, 64)
			s.ConsHead = ConstructionHead{Active: true, X: 15, Y: 15, BaseZ: 64, Rotation: rot}
			if !s.PlaceRoadAtHead(int(p.ID), 0) {
				t.Errorf("piece %d (road %d) rot %d: can't be built", piece, p.ID, rot)
				continue
			}
			if h := s.ConsHead; h.X == 15 && h.Y == 15 && h.Rotation == rot {
				t.Errorf("piece %d (road %d) rot %d: the head didn't move or turn", piece, p.ID, rot)
			}
		}
	}
}
