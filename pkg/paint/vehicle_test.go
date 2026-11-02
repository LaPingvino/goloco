package paint

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// testdata/body_bounds.txt is paintBody's bounding-box code (PaintVehicle.cpp,
// from the overhang offset on, with kBoundingBoxFactorsByYaw) compiled and
// run for 3000 cases of one LCG, which this test repeats.
func TestBodyBoundsMatchUpstream(t *testing.T) {
	data, err := os.ReadFile("testdata/body_bounds.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Split(strings.TrimSpace(string(data)), "\n")
	st := uint32(4242)
	rnd := func(n uint32) int { st = st*1103515245 + 12345; return int((st >> 16) % n) }
	for c := range 3000 {
		front, back, bodyRev, rev, half, yaw, z := rnd(32), rnd(32), rnd(2), rnd(2), rnd(40), rnd(64), rnd(2000)-500
		var v objects.VehicleObject
		v.CarComponents[0] = objects.VehicleCarComponent{FrontBogiePosition: uint8(front), BackBogiePosition: uint8(back), BodySpriteIdx: uint8(bodyRev << 7)}
		sp := objects.VehicleBodySprite{HalfLength: uint8(half)}
		off, size := bodyBounds(BodyEl{Reversed: rev == 1, Z: z}, &v, &sp, uint8(yaw))
		if got := fmt.Sprintf("%d %d %d %d %d %d", off.X, off.Y, off.Z, size.X, size.Y, size.Z); got != want[c] {
			t.Fatalf("case %d: %s, upstream %s", c, got, want[c])
		}
	}
}

// paintBogie's branches, worked by hand from PaintVehicle.cpp.
func TestBogieImage(t *testing.T) {
	sym := objects.VehicleBogieSprite{Flags: objects.BogieFlagRotationalSymmetry, NumFramesPerRotation: 2, FlatImageIDs: 100, GentleImageIDs: 200, SteepImageIDs: 300}
	big := objects.VehicleBogieSprite{Flags: objects.BogieFlagLargerBoundingBox, NumFramesPerRotation: 3, FlatImageIDs: 100, GentleImageIDs: 200, SteepImageIDs: 300}
	for i, c := range []struct {
		b      BogieEl
		sp     objects.VehicleBogieSprite
		img    uint32
		table  int
		bbSize Pos3
	}{
		// flat, yaw 40 → 20 & 15 = 4: 2·4 + 1 + 100
		{BogieEl{SpriteYaw: 40, AnimationIndex: 1}, sym, 109, 0, Pos3{12, 12, 1}},
		// reversed flat: yaw 40^32 = 8 → 4; larger box
		{BogieEl{SpriteYaw: 40, Reversed: true}, big, 112, 0, Pos3{18, 18, 5}},
		// down12 (6), not symmetric: yaw 10 → 5 ^ 32 = 37: 3·37 + 200
		{BogieEl{SpriteYaw: 10, SpritePitch: 6}, big, 311, 1, Pos3{16, 16, 1}},
		// down25 (8), symmetric: yaw 10 → 5 ^ 16 = 21: 2·21 + 300
		{BogieEl{SpriteYaw: 10, SpritePitch: 8}, sym, 342, 2, Pos3{12, 12, 1}},
		// up25 reversed → down25: yaw 10^32 = 42 → 21 ^ 16 = 5 (symmetric): 2·5 + 300
		{BogieEl{SpriteYaw: 10, SpritePitch: 4, Reversed: true}, sym, 310, 2, Pos3{12, 12, 1}},
	} {
		img, table, _, size := bogieImage(c.b, &c.sp, 0)
		if img != c.img || table != c.table || size != c.bbSize {
			t.Errorf("case %d: image %d table %d size %v, want %d %d %v", i, img, table, size, c.img, c.table, c.bbSize)
		}
	}
}
