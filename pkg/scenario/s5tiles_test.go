package scenario

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// gameFiles is the title game and every scenario of the game data, when
// present (the tests skip without it).
func gameFiles(t *testing.T) []string {
	files := []string{"../../../locomotion/Data/title.dat"}
	sc, _ := filepath.Glob("../../../locomotion/Scenarios/*.SC5")
	files = append(files, sc...)
	sv, _ := filepath.Glob("../../../locomotion/Single Player Saved Games/*.SV5")
	files = append(files, sv...)
	return files
}

// TestEncodeTileElements_bytes: the tile elements written back from a loaded
// map are the file's own, byte for byte, but for the bits toSaveElement
// doesn't carry (a surface's type bits 0–1, a station's unk4SLR4, a wall's
// byte 7).
func TestEncodeTileElements_bytes(t *testing.T) {
	n := 0
	for _, f := range gameFiles(t) {
		sc, err := LoadScenarioData(f)
		if err != nil || sc.Raw == nil || sc.Raw.TileElements == nil || sc.Header.Flags&(assets.HeaderFlagIsRaw|assets.HeaderFlagIsDump) != 0 {
			continue // a dump (DEBUG_AUTO_DUMP.SV5) keeps every empty element slot: upstream refuses it
		}
		n++
		got, want := EncodeTileElements(sc.Tiles), sc.Raw.TileElements
		if len(got) != len(want) {
			t.Errorf("%s: %d bytes, want %d", filepath.Base(f), len(got), len(want))
			continue
		}
		diffs := map[string]int{}
		for i := 0; i < len(want); i += tileElementSize {
			kind := (want[i] & elementTypeMask) >> elementTypeShift
			for b := range tileElementSize {
				x := got[i+b] ^ want[i+b]
				if x == 0 || upstreamDrops(kind, b, x) {
					continue
				}
				diffs[fmt.Sprintf("kind %d byte %d mask %02x", kind, b, x)]++
			}
		}
		var keys []string
		for k := range diffs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			t.Errorf("%s: %s ×%d", filepath.Base(f), k, diffs[k])
		}
	}
	if n == 0 {
		t.Skip("no game data")
	}
}

// upstreamDrops: bits upstream's load (loadTileElements) or save
// (toSaveElement) doesn't carry.
func upstreamDrops(kind uint8, b int, x uint8) bool {
	switch {
	case kind == KindSurface && b == 0:
		return x&^0x03 == 0
	case kind == KindStation && b == 4:
		return x&^0xF0 == 0
	case kind == KindWall && b == 7:
		return true
	}
	return false
}
