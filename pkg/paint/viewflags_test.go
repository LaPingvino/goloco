package paint

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// The see-through classes are Paint.cpp's isTypeCullable… switches.
func TestShouldTryCull(t *testing.T) {
	classes := map[ViewFlags][]InteractionItem{
		ViewSeeThroughTracks:    {ItemTrack, ItemTrackExtra, ItemSignal, ItemTrainStation},
		ViewSeeThroughRoads:     {ItemRoadStation, ItemRoad, ItemRoadExtra},
		ViewSeeThroughTrees:     {ItemIndustryTree, ItemTree},
		ViewSeeThroughScenery:   {ItemWall},
		ViewSeeThroughBuildings: {ItemBuilding, ItemIndustry, ItemHeadquarterBuilding, ItemAirport, ItemDock},
		ViewSeeThroughBridges:   {ItemBridge},
	}
	owner := map[InteractionItem]ViewFlags{}
	for f, items := range classes {
		for _, it := range items {
			owner[it] = f
		}
	}
	for it := ItemNoInteraction; it <= ItemBuildingInfo; it++ {
		for f := range classes {
			if got, want := ShouldTryCull(it, f), owner[it] == f; got != want {
				t.Errorf("ShouldTryCull(%v, %#x) = %v, want %v", it, f, got, want)
			}
		}
		// the flags that are not see-through modes cull nothing
		if ShouldTryCull(it, ViewUnderground|ViewHeightMarksOnLand|ViewHeightMarksOnTracks|ViewOneWayDirectionArrows|ViewGridlines|ViewHideTownNames|ViewHideStationNames) {
			t.Errorf("ShouldTryCull(%v) without a see-through flag", it)
		}
	}
}

// drawStructs' see-through step: a see-through struct's image is its bare
// sprite translucent in unk30; in the underground view, or when it was
// blended already, it is not drawn; others are untouched.
func TestCulled(t *testing.T) {
	sp := &gfx8.Sprite{Width: 1, Height: 1, Pix: []uint8{5}}
	remap, unk30 := []uint8{1}, []uint8{2}
	s := &Session{ColourMap: func(c int) []uint8 {
		if c != 0x30 {
			t.Errorf("colour map %#x asked for, want 0x30", c)
		}
		return unk30
	}}
	plain := Image{Sprite: sp, Op: gfx8.OpTransparent | gfx8.OpSrc, PalMap: remap, Noise: sp}
	if img, ok := s.culled(ItemTree, plain); !ok || img.Op != plain.Op || &img.PalMap[0] != &remap[0] {
		t.Errorf("no flags: %+v %v, want the image as it is", img, ok)
	}
	s.ViewFlags = ViewSeeThroughTrees
	img, ok := s.culled(ItemTree, plain)
	if !ok || img.Op != gfx8.OpTransparent|gfx8.OpDst || &img.PalMap[0] != &unk30[0] || img.Noise != nil || img.Sprite != sp {
		t.Errorf("see-through tree: %+v %v, want the sprite translucent in unk30", img, ok)
	}
	if img, ok := s.culled(ItemBuilding, plain); !ok || img.Op != plain.Op {
		t.Errorf("a building with see-through trees: %+v %v, want it as it is", img, ok)
	}
	if _, ok := s.culled(ItemTree, Image{Sprite: sp, Op: gfx8.OpTransparent | gfx8.OpDst, PalMap: remap}); ok {
		t.Error("a blended image of a see-through struct (a shadow) is drawn")
	}
	s.ViewFlags |= ViewUnderground
	if _, ok := s.culled(ItemTree, plain); ok {
		t.Error("a see-through struct is drawn in the underground view")
	}
	if _, ok := s.culled(ItemBuilding, plain); !ok {
		t.Error("the underground view alone hides a building")
	}
}

// getHeightMarkerImage: height / 16 − seaLevel + 128 into the units' table.
// ImageIds.h: height_marker_neg_128_u = 2470, the feet 256 on, the metres 512 on.
func TestHeightMarkerImage(t *testing.T) {
	for _, c := range []struct{ height, sea, units, want int }{
		{0, 0, 0, 2470 + 128},
		{16 * 4, 4, 0, 2470 + 128}, // at sea level: "0"
		{16*4 + 3, 4, 0, 2470 + 128},
		{16 * 5, 4, 1, 2726 + 129},
		{16 * 3, 4, 2, 2982 + 127},
	} {
		if got, ok := HeightMarkerImage(c.height, c.sea, c.units); !ok || got != c.want {
			t.Errorf("HeightMarkerImage(%d, sea %d, units %d) = %d %v, want %d", c.height, c.sea, c.units, got, ok, c.want)
		}
	}
	if _, ok := HeightMarkerImage(16*200, 0, 0); ok {
		t.Error("a height past the table has a marker")
	}
}

// A blended image — the underground view's translucent ground — still counts
// under the cursor: isSpriteInteractedWith tests the sprite's pixel.
func TestHitsBlended(t *testing.T) {
	sp := &gfx8.Sprite{Width: 1, Height: 1, Pix: []uint8{5}, HasTransparency: true}
	s := NewSession(0, 0, 0, 1, 1)
	if !s.hits(Image{Sprite: sp, Op: gfx8.OpTransparent | gfx8.OpDst, PalMap: make([]uint8, 256)}, 0, 0) {
		t.Error("a translucent image over the pixel is not hit")
	}
}
