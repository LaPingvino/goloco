package objects

import (
	"compress/gzip"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

// testdata/body_image_index.txt is upstream's getBodyImageIndex and
// getBrakingImageIndex (VehicleDraw.cpp, compiled) for four sprite setups ×
// pitch 0–14 × yaw 0–255: "body braking," per case, one line per setup.
func TestBodyImageIndexMatchesUpstream(t *testing.T) {
	data, err := os.ReadFile("testdata/body_image_index.txt")
	if err != nil {
		t.Fatal(err)
	}
	setups := []struct{ flags, flatAcc, slAcc, nfpr uint8 }{{1, 1, 0, 3}, {1 | 8 | 16, 3, 2, 5}, {1 | 2 | 8, 2, 1, 2}, {1 | 8 | 16, 4, 3, 7}}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != len(setups) {
		t.Fatalf("%d lines", len(lines))
	}
	for si, st := range setups {
		s := VehicleBodySprite{Flags: st.flags, FlatYawAccuracy: st.flatAcc, SlopedYawAccuracy: st.slAcc, NumFramesPerRotation: st.nfpr, FlatImageID: 1000, GentleImageID: 2000, SteepImageID: 3000}
		cases := strings.Split(strings.TrimSuffix(lines[si], ","), ",")
		bad := 0
		for i, c := range cases {
			p, yaw := Pitch(i/256), uint8(i%256)
			f := strings.Fields(c)
			body, _ := strconv.Atoi(f[0])
			brake, _ := strconv.Atoi(f[1])
			if g := s.BodyImageIndex(p, yaw, 0, 0); g != uint32(body) {
				if bad++; bad < 5 {
					t.Errorf("setup %d pitch %d yaw %d: body %d, upstream %d", si, p, yaw, g, body)
				}
			}
			if g := s.BrakingImageIndex(p, yaw); g != uint32(brake) {
				if bad++; bad < 5 {
					t.Errorf("setup %d pitch %d yaw %d: braking %d, upstream %d", si, p, yaw, g, brake)
				}
			}
		}
	}
}

// testdata/draw_items.txt.gz is upstream's getDrawItemsForVehicle (object
// variant, VehicleDraw.cpp compiled) for 200 pseudo-random vehicle objects ×
// 13 yaws × 8 rolls: "total image:dist:isBody …" per line. The objects come
// from the same LCG in the C++ harness and here.
func TestDrawItemsMatchesUpstream(t *testing.T) {
	f, err := os.Open("testdata/draw_items.txt.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(zr)
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	st := uint32(12345)
	rnd := func(n uint32) uint32 { st = st*1103515245 + 12345; return (st >> 16) % n }
	line, bad := 0, 0
	for c := 0; c < 200; c++ {
		var v VehicleObject
		v.Mode = TransportMode(rnd(3))
		v.Flags = uint16(rnd(4) << 2)
		v.NumCarComponents = uint8(1 + rnd(4))
		for i := range 4 {
			b := &v.BodySprites[i]
			b.HalfLength = uint8(rnd(20))
			f1 := rnd(2) << 1
			f6 := rnd(2) << 6
			b.Flags = uint8(1 | f1 | f6)
			b.NumAnimationFrames = uint8(1 << rnd(3))
			b.NumRollFrames = 1
			if rnd(2) != 0 {
				b.NumRollFrames = 3
			}
			b.FlatYawAccuracy = uint8(1 + rnd(4))
			b.NumFramesPerRotation = uint8(1 + rnd(9))
			b.FlatImageID = rnd(500)
		}
		for i := range 2 {
			g := &v.BogieSprites[i]
			g.Flags = uint8(1 | rnd(2)<<1)
			g.NumFramesPerRotation = uint8(1 << rnd(3))
			g.FlatImageIDs = rnd(500)
		}
		pick := func() uint8 {
			if rnd(3) == 0 {
				return 0xFF
			}
			return uint8(rnd(2))
		}
		for i := range 4 {
			k := &v.CarComponents[i]
			k.FrontBogiePosition = uint8(rnd(16))
			k.BackBogiePosition = uint8(rnd(16))
			k.FrontBogieSpriteIdx = pick()
			k.BackBogieSpriteIdx = pick()
			if rnd(5) == 0 {
				k.BodySpriteIdx = 0xFF
			} else {
				lo := rnd(4)
				k.BodySpriteIdx = uint8(lo | rnd(2)<<7)
			}
		}
		for yaw := 0; yaw < 64; yaw += 5 {
			for roll := range 8 {
				items, total := v.DrawItems(uint8(yaw), uint8(roll))
				got := strconv.Itoa(int(total))
				for _, it := range items {
					b := 0
					if it.IsBody {
						b = 1
					}
					got += " " + strconv.Itoa(int(it.Image)) + ":" + strconv.Itoa(int(it.Dist)) + ":" + strconv.Itoa(b)
				}
				if got != lines[line] {
					if bad++; bad < 5 {
						t.Errorf("object %d yaw %d roll %d:\n got %s\nwant %s", c, yaw, roll, got, lines[line])
					}
				}
				line++
			}
		}
	}
	if bad > 0 {
		t.Errorf("%d of %d cases differ", bad, line)
	}
}

// testdata/screen_distance.txt is toScreenDistDrawItems' getScreenDistance
// (computeXYVector, Pos3, gameToScreen at rotation 0 — compiled) for yaw
// 0–63 × distance −300…300 step 7.
func TestScreenDistanceMatchesUpstream(t *testing.T) {
	data, err := os.ReadFile("testdata/screen_distance.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Split(strings.TrimSuffix(strings.TrimSpace(string(data)), ","), ",")
	i := 0
	for yaw := 0; yaw < 64; yaw++ {
		for d := -300; d <= 300; d += 7 {
			if g := strconv.Itoa(int(screenDistance(int32(d), uint8(yaw)))); g != want[i] {
				t.Errorf("yaw %d dist %d: %s, upstream %s", yaw, d, g, want[i])
			}
			i++
		}
	}
}

// A car whose parts mirror its object's components (front bogie turned as
// the body, back bogie the other way, no animation) draws as the object at
// roll 0 — the object variant is checked against upstream above.
func TestCarDrawItemsMatchesObjectVariant(t *testing.T) {
	st := uint32(777)
	rnd := func(n uint32) uint32 { st = st*1103515245 + 12345; return (st >> 16) % n }
	for c := 0; c < 300; c++ {
		var v VehicleObject
		v.Mode = TransportMode(rnd(3))
		v.NumCarComponents = uint8(1 + rnd(4))
		for i := range 4 {
			b := &v.BodySprites[i]
			b.HalfLength, b.Flags = uint8(rnd(20)), uint8(1|rnd(2)<<1)
			b.NumAnimationFrames, b.NumRollFrames = 1, 1
			b.FlatYawAccuracy, b.NumFramesPerRotation, b.FlatImageID = uint8(1+rnd(4)), uint8(1+rnd(9)), rnd(500)
		}
		for i := range 2 {
			g := &v.BogieSprites[i]
			g.Flags, g.NumFramesPerRotation, g.FlatImageIDs = uint8(1|rnd(2)<<1), uint8(1<<rnd(3)), rnd(500)
		}
		var comps []CarComponent
		for i := range int(v.NumCarComponents) {
			k := &v.CarComponents[i]
			k.FrontBogiePosition, k.BackBogiePosition = uint8(rnd(16)), uint8(rnd(16))
			k.FrontBogieSpriteIdx, k.BackBogieSpriteIdx = uint8(rnd(2)), uint8(rnd(2))
			k.BodySpriteIdx = uint8(rnd(4) | rnd(2)<<7)
			rev := k.BodySpriteIdx&SpriteIndexIsReversed != 0
			comps = append(comps, CarComponent{
				Front: CarPart{SpriteType: k.FrontBogieSpriteIdx, Reversed: rev},
				Back:  CarPart{SpriteType: k.BackBogieSpriteIdx, Reversed: !rev},
				Body:  CarPart{SpriteType: k.BodySpriteIdx &^ SpriteIndexIsReversed, Reversed: rev},
			})
		}
		for yaw := 0; yaw < 64; yaw += 3 {
			want, wt := v.DrawItems(uint8(yaw), 0)
			got, gt := v.CarDrawItems(uint8(yaw), comps, false, true, false)
			if wt != gt || len(want) != len(got) {
				t.Fatalf("object %d yaw %d: %v/%d vs %v/%d", c, yaw, got, gt, want, wt)
			}
			for i := range want {
				if want[i] != got[i] {
					t.Fatalf("object %d yaw %d item %d: %+v, object variant %+v", c, yaw, i, got[i], want[i])
				}
			}
		}
	}
}

// A reversed two-component car, animated, with cargo and brake lights:
// components from the last, bogie positions swapped, animation and cargo
// frames added, the braking frame over the body.
func TestCarDrawItemsReversedAnimated(t *testing.T) {
	var v VehicleObject
	v.Mode = TransportModeRail
	v.NumCarComponents = 2
	v.BodySprites[0] = VehicleBodySprite{Flags: BodyFlagHasSprites | BodyFlagHasBrakingLights, HalfLength: 10, FlatYawAccuracy: 3, NumFramesPerRotation: 4, FlatImageID: 100}
	v.BodySprites[1] = VehicleBodySprite{Flags: BodyFlagHasSprites, HalfLength: 6, FlatYawAccuracy: 3, NumFramesPerRotation: 2, FlatImageID: 300}
	v.BogieSprites[0] = VehicleBogieSprite{Flags: BogieFlagHasSprites, NumFramesPerRotation: 2, FlatImageIDs: 50}
	v.CarComponents[0] = VehicleCarComponent{FrontBogiePosition: 2, BackBogiePosition: 3, FrontBogieSpriteIdx: 0, BackBogieSpriteIdx: 0, BodySpriteIdx: 0}
	v.CarComponents[1] = VehicleCarComponent{FrontBogiePosition: 4, BackBogiePosition: 5, FrontBogieSpriteIdx: 0, BackBogieSpriteIdx: 0xFF, BodySpriteIdx: 1}
	col := [2]uint8{7, 2}
	comps := []CarComponent{
		{Front: CarPart{SpriteType: 0, Reversed: true, Animation: 1, Colour: col}, Back: CarPart{SpriteType: 0xFF}, Body: CarPart{SpriteType: 1, Reversed: true, Animation: 1, CargoFrame: 0, Colour: col}},
		{Front: CarPart{SpriteType: 0, Animation: 1, Colour: col}, Back: CarPart{SpriteType: 0, Reversed: true, Colour: col}, Body: CarPart{SpriteType: 0, Animation: 2, CargoFrame: 1, Colour: col}},
	}
	got, total := v.CarDrawItems(8, comps, true, true, true)
	// yaw 8: bogie (8^32)/2 = 20 → 2*20+1+50 = 91; unreversed (8/2) = 4 → 2*4+1+50 = 59, reversed back 2*20+0+50 = 90.
	// component 1 first (reversed): front pos = back 5, back pos = front 4, length 12; body (8^32)>>1 = 20 → 20*2+300+1 = 341 at (5+12-4)/2 = 6.
	// component 0: front = 3, back = 2, length 20, at total 12; body 8>>1 = 4 → 4*4+100+2+1 = 119 at 12+(3+20-2)/2 = 22; braking 4*4+100+3 = 119.
	want := []DrawItem{
		{91, 5, false, col}, {341, 6, true, col},
		{59, 15, false, col}, {90, 30, false, col}, {119, 22, true, col}, {119, 22, true, col},
	}
	if total != 32 || len(got) != len(want) {
		t.Fatalf("got %+v total %d", got, total)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("item %d: %+v, want %+v", i, got[i], want[i])
		}
	}
}
