package objects

import (
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
