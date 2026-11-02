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
