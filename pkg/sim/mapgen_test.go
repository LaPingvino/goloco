package sim

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// GenerateLand against upstream's generateLand body compiled on 40 random
// 24×24 height maps (tools/gen-paintdata/gen_generateland.cpp): every
// drawable tile's baseZ, slope and clearZ.
func TestGenerateLandVectors(t *testing.T) {
	f, err := os.Open("testdata/generateland.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	const w, h = 24, 24
	maps, tiles := 0, 0
	for sc.Scan() {
		hmLine := strings.Fields(sc.Text())
		if !sc.Scan() {
			t.Fatal("truncated")
		}
		want := strings.Fields(sc.Text())
		hm := NewHeightMap(w, h)
		for i, s := range hmLine {
			v, _ := strconv.Atoi(s)
			hm.data[i] = uint8(v)
		}
		var m Map
		m.Width, m.Height = w, h
		m.Tiles = make([][]Tile, h)
		for y := range m.Tiles {
			m.Tiles[y] = make([]Tile, w)
		}
		m.GenerateLand(hm)
		k := 0
		for y := 1; y <= h-2; y++ {
			for x := 1; x <= w-2; x++ {
				tl := m.TileAt(x, y)
				wb, _ := strconv.Atoi(want[k])
				ws, _ := strconv.Atoi(want[k+1])
				wc, _ := strconv.Atoi(want[k+2])
				k += 3
				if int(tl.BaseZ) != wb || int(tl.Slope) != ws || int(tl.SurfaceClearZ()) != wc {
					t.Fatalf("map %d tile (%d,%d): baseZ %d slope %#x clearZ %d, want %d %#x %d", maps, x, y, tl.BaseZ, tl.Slope, tl.SurfaceClearZ(), wb, ws, wc)
				}
				tiles++
			}
		}
		maps++
	}
	if maps != 40 {
		t.Fatalf("%d maps", maps)
	}
	t.Logf("%d maps, %d tiles match", maps, tiles)
}
