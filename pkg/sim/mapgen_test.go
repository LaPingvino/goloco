package sim

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

// GetHeight against upstream's getHeight and slope helpers compiled on a stub
// surface (tools/gen-paintdata/gen_getheight.cpp): all 32 slope bytes × base
// heights 0 / 4 / 40 × water 0 / 3 × every point of the tile, plus off the
// map — the same text, compared by SHA-256. (The helpers' kTileSize is
// TileManager.cpp's own 31; the vectors were first made with 32, the tile's
// size, and GetHeight was a unit off on slopes until 2026-11-13.)
func TestGetHeightMatchesUpstream(t *testing.T) {
	var b strings.Builder
	var m Map
	m.Width, m.Height = 384, 384
	m.Tiles = make([][]Tile, 384)
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, 384)
	}
	for sl := 0; sl < 32; sl++ {
		for _, bz := range []int{0, 4, 40} {
			for _, w := range []int{0, 3} {
				tl := m.TileAt(2, 3)
				tl.BaseZ, tl.Slope, tl.WaterLevel = uint8(bz), uint8(sl), uint8(w)
				for y := 0; y < 32; y++ {
					for x := 0; x < 32; x++ {
						land, water := m.GetHeight(int32(64+x), int32(96+y))
						fmt.Fprintf(&b, "%d %d %d %d %d %d %d\n", sl, bz, w, x, y, land, water)
					}
				}
			}
		}
	}
	land, water := m.GetHeight(-1, 5)
	fmt.Fprintf(&b, "off %d %d\n", land, water)
	sum := sha256.Sum256([]byte(b.String()))
	if got := hex.EncodeToString(sum[:]); got != "1af84ee6d59a2f63c8758623fd2deee8c12db75944b9849ea7350917a1f94541" {
		t.Fatalf("sha256 %s, upstream 1af84ee6d59a2f63c8758623fd2deee8c12db75944b9849ea7350917a1f94541", got)
	}
}
