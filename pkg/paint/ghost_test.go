package paint

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// Ghosts are drawn through applyGhostToImage's palette map (ExtColour::unk2C)
// instead of the element's own colour: these tests tell the maps apart by
// identity.

func ghostSprite(int) *gfx8.Sprite { return &gfx8.Sprite{Width: 1, Height: 1, Pix: []uint8{1}} }

func samePal(a, b []uint8) bool { return len(a) > 0 && len(b) > 0 && &a[0] == &b[0] }

// palsOf are the palette maps of the structs added since n.
func palsOf(s *Session, n int) [][]uint8 {
	var out [][]uint8
	for i := n; i < len(s.structs); i++ {
		out = append(out, s.structs[i].Img.PalMap)
	}
	return out
}

// PaintBuilding.cpp paintBuilding: a ghost's baseColour is
// applyGhostToImage(0), and its scaffolding applyGhostToImage(scaffObj->image)
// — not the element's colour nor the object's scaffoldingColour.
func TestBuildingGhostColour(t *testing.T) {
	own, scaff, ghost := make([]uint8, 256), make([]uint8, 256), make([]uint8, 256)
	obj := BuildingObj{Image: 500, NumParts: 2, Parts: []uint8{0, 1}, PartHeights: []uint8{24, 8},
		PartAnimFrames: []uint8{0, 0}, PartAnimSpeed: []uint8{0, 0},
		Scaffold: &Scaffold{Image: 9000, Type: 1, SegmentHeight: 16, Colour: scaff}, GhostMap: ghost}
	for _, ghostEl := range []bool{false, true} {
		s := NewSession(0, -10000, -10000, 20000, 20000)
		s.SetSpritePosition(320, 320)
		n := len(s.structs)
		s.PaintBuilding(BuildingEl{BaseZ: 4, ClearZ: 20, Age: 0, Ghost: ghostEl}, obj, own, 10, 10, 0, ghostSprite)
		pals := palsOf(s, n)
		if len(pals) < 4 { // scaffolding back, part, front, roof at least
			t.Fatalf("ghost %v: %d structs", ghostEl, len(pals))
		}
		for i, p := range pals {
			if ghostEl && !samePal(p, ghost) {
				t.Errorf("ghost building: struct %d not in the ghost colour", i)
			}
			if !ghostEl && samePal(p, ghost) {
				t.Errorf("built building: struct %d in the ghost colour", i)
			}
		}
		if !ghostEl && !samePal(pals[0], scaff) {
			t.Error("built building: the scaffolding should keep scaffoldingColour")
		}
		if want := ItemNoInteraction; ghostEl && s.ItemType != want {
			t.Errorf("ghost building item type %v, want noInteraction", s.ItemType)
		}
	}
}

// PaintIndustry.cpp paintIndustryBuilding: a ghost's scaffolding is
// applyGhostToImage(scaffObj->image) too.
func TestIndustryGhostScaffolding(t *testing.T) {
	ghost, scaff := make([]uint8, 256), make([]uint8, 256)
	obj := IndustryObj{BuildingImageIDs: 500, Parts: []uint8{0, 1}, PartHeights: []uint8{24, 8},
		Scaffold: &Scaffold{Image: 9000, Type: 1, SegmentHeight: 16, Colour: scaff}, GhostMap: ghost}
	s := NewSession(0, -10000, -10000, 20000, 20000)
	s.SetSpritePosition(320, 320)
	n := len(s.structs)
	s.PaintIndustry(IndustryEl{Var6Low: 0, BaseZ: 4, ClearZ: 20, Ghost: true}, obj, make([]uint8, 256), 10, 10, 0, ghostSprite)
	for i, p := range palsOf(s, n) {
		if !samePal(p, ghost) {
			t.Errorf("ghost industry: struct %d not in the ghost colour", i)
		}
	}
	if samePal(obj.Scaffold.Colour, ghost) {
		t.Error("the shared scaffold was changed")
	}
}

// PaintTrack.cpp paintTrack: a mod the element has is in the element's
// colour; a ghost mod (hasGhostMods with the mod selected) in
// applyGhostToImage's.
func TestTrackGhostMod(t *testing.T) {
	own, ghost := make([]uint8, 256), make([]uint8, 256)
	for _, m := range []TrackMod{{Has: true, Image: 1000}, {Has: true, Image: 1000, Ghost: ghost}} {
		s := NewSession(0, -10000, -10000, 20000, 20000)
		s.SetSpritePosition(320, 320)
		s.PaintTrackMods(TrackEl{TrackID: 0, Rotation: 1, BaseZ: 8, Remap: own}, [4]TrackMod{m}, ghostSprite)
		pals := palsOf(s, 0)
		if len(pals) != 1 {
			t.Fatalf("%d structs, want 1", len(pals))
		}
		want := own
		if m.Ghost != nil {
			want = ghost
		}
		if !samePal(pals[0], want) {
			t.Errorf("mod ghost=%v: wrong colour", m.Ghost != nil)
		}
	}
}

// PaintAirport.cpp paintAirport: the parts in baseColour — the caller's
// Remap, applyGhostToImage(0) for a ghost (scene stationBuildingRemap).
func TestAirportRemap(t *testing.T) {
	ghost := make([]uint8, 256)
	s := NewSession(0, -10000, -10000, 20000, 20000)
	s.SetSpritePosition(320, 320)
	s.ElGhost = true
	s.PaintAirport(0, 0, 4, AirportObj{BuildingImage: 100, Parts: []uint8{0}, PartHeights: []uint8{8},
		PartAnimFrames: []uint8{1}, PartAnimSpeed: []uint8{0}, Remap: ghost}, 10, 10, 0, ghostSprite)
	pals := palsOf(s, 0)
	if len(pals) != 1 || !samePal(pals[0], ghost) {
		t.Errorf("airport part palettes %d, want the one given", len(pals))
	}
	if s.ItemType != ItemNoInteraction {
		t.Errorf("ghost airport item type %v", s.ItemType)
	}
}
