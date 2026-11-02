package paint

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// An industry under construction (1×1, one section built: age 0 counts one
// part, 24 high, segment height 16) gets its scaffolding: two back segments
// before the parts, two front segments and the roof after.
func TestIndustryScaffolding(t *testing.T) {
	var asked []int
	sprite := func(id int) *gfx8.Sprite {
		asked = append(asked, id)
		return &gfx8.Sprite{Width: 1, Height: 1, Pix: []uint8{1}}
	}
	s := NewSession(0, -10000, -10000, 20000, 20000)
	s.SetSpritePosition(320, 320)
	obj := IndustryObj{BuildingImageIDs: 500, Parts: []uint8{0, 1}, PartHeights: []uint8{24, 8},
		Scaffold: &Scaffold{Image: 9000, Type: 1, SegmentHeight: 16}}
	before := len(s.structs)
	s.PaintIndustry(IndustryEl{Var6Low: 0, IsConstructed: false, BaseZ: 4, ClearZ: 20}, obj, nil, 10, 10, 0, sprite)
	// type 1, 1x1: back 12, front 13, roof NE 14 (rotation 0); each looked up once
	count := map[int]int{}
	for _, id := range asked {
		count[id]++
	}
	if count[9012] != 1 || count[9013] != 1 || count[9014] != 1 {
		t.Errorf("scaffolding images asked %v", asked)
	}
	// structs: 2 back segments + the one built part + 2 front + roof = 6
	if n := len(s.structs) - before; n != 6 {
		t.Errorf("%d structs added, want 6", n)
	}
	// Constructed: no scaffolding.
	asked = nil
	s.PaintIndustry(IndustryEl{IsConstructed: true, BaseZ: 4, ClearZ: 20}, obj, nil, 10, 10, 0, sprite)
	for _, id := range asked {
		if id >= 9000 {
			t.Fatalf("constructed industry asked scaffolding image %d", id)
		}
	}
}
