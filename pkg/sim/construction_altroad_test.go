package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/construct"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// constructRoad's retry (sub_4A18D4): a straight of road A refused across a
// straight of road B (unable_to_cross_or_create_junction_with_string, B the
// alternateRoadObjectId) switches the window to B when both are
// any-road-type compatible and B is available — then B's junction is built.
// Without the flag on A nothing switches and the refusal stands.
func TestConstructRoadAlternateRoad(t *testing.T) {
	hdr := func(name string) *objects.ObjectHeader {
		h := &objects.ObjectHeader{}
		copy(h.Name[:], name)
		return h
	}
	for _, anyA := range []bool{true, false} {
		a := &objects.RoadObject{Header: hdr("ROADA"), Name: "Road A", BuildCostFactor: 100, Flags: roadFlagAnyRoadTypeCompatible}
		if !anyA {
			a.Flags = 0
		}
		b := &objects.RoadObject{Header: hdr("ROADB"), Name: "Road B", BuildCostFactor: 100, Flags: roadFlagAnyRoadTypeCompatible, RoadPieces: roadTraitJunction}
		om := &objects.ObjectManager{RoadObjects: []*objects.RoadObject{a, b}}
		s := &State{}
		s.Width, s.Height = 8, 8
		s.Tiles = make([][]Tile, 8)
		for y := range s.Tiles {
			s.Tiles[y] = make([]Tile, 8)
			for x := range s.Tiles[y] {
				s.Tiles[y][x].BaseZ = 4
				s.Tiles[y][x].Order = []scenario.ElemRef{{Kind: scenario.KindSurface}}
			}
		}
		c := &Cmd{Company: CompanyNeutral, Objects: om}
		if _, ok := s.CreateRoad(c, CmdApply, RoadPlacement{X: 96, Y: 96, Z: 16, Rotation: 1, RoadObjectID: 1, Bridge: 0xFF}); !ok {
			t.Fatalf("road B: error %d", c.ErrorText)
		}
		var switched []uint8
		s.SwitchRoadType = func(obj uint8) { switched = append(switched, obj) }
		s.Cons = ConstructionState{X: 96, Y: 96, Z: 16, Piece: construct.Straight, Gradient: construct.Level}
		c = &Cmd{Company: CompanyNeutral, Objects: om}
		_, ok := s.ConstructRoad(c, 0)
		if anyA {
			if !ok || len(switched) != 1 || switched[0] != 1 {
				t.Fatalf("any-road-type A over B: built %v, switched %v, error %d", ok, switched, c.ErrorText)
			}
			if n := len(s.TileAt(3, 3).Roads); n != 2 {
				t.Errorf("the junction: %d road elements on the tile, want 2", n)
			}
			continue
		}
		if ok || len(switched) != 0 || c.ErrorText != i18n.UnableToCrossOrCreateJunctionWithString || c.ErrorArg != "Road B" {
			t.Errorf("A without the flag: built %v, switched %v, error %d %q", ok, switched, c.ErrorText, c.ErrorArg)
		}
	}
}
