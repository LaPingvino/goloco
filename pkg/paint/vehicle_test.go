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
