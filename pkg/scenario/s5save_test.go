package scenario

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// gsSection names the GameState part an offset falls in, and the offset in
// its record.
func gsSection(o int) string {
	type sec struct {
		name        string
		start, size int
	}
	secs := []sec{{"general", 0, 1}, {"company", companiesGameStateOffset, companyRecordSize}, {"town", townsGameStateOffset, townRecordSize},
		{"industry", industriesGameStateOffset, industryRecordSize}, {"station", stationsGameStateOffset, stationRecordSize},
		{"entity", EntityTableOffset, EntitySize}, {"animation", animationsOffset, 6}, {"wave", wavesOffset, 6},
		{"userstring", userStringsOffset, userStringSize}, {"routing", gameStateRoutings, 2}, {"orders", gameStateOrders, 1}}
	for i := len(secs) - 1; i >= 0; i-- {
		if o >= secs[i].start {
			if secs[i].name == "general" || secs[i].name == "orders" {
				return fmt.Sprintf("%s %#x", secs[i].name, o-secs[i].start)
			}
			return fmt.Sprintf("%s +%#x", secs[i].name, (o-secs[i].start)%secs[i].size)
		}
	}
	return "?"
}

// TestEncodeGameState_identity: a loaded game's state written back unchanged
// is the file's own, byte for byte, but for upstream's save normalisations
// (normalised here on the file's bytes too).
func TestEncodeGameState_identity(t *testing.T) {
	n := 0
	for _, f := range gameFiles(t) {
		sc, err := LoadScenarioData(f)
		if err != nil || sc.Raw == nil || sc.Header.Flags&(assets.HeaderFlagIsRaw|assets.HeaderFlagIsDump) != 0 {
			continue
		}
		n++
		got, want := sc.EncodeGameState(), sc.Raw.GameState
		if sc.Header.Type == assets.S5TypeSavedGame {
			want = upstreamSaveNormalised(want)
		}
		diffs := map[string]int{}
		for i := range want {
			if got[i] != want[i] {
				diffs[gsSection(i)]++
			}
		}
		var keys []string
		for k := range diffs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 40 {
				t.Errorf("%s: … %d more", filepath.Base(f), len(keys)-i)
				break
			}
			t.Errorf("%s: %s ×%d", filepath.Base(f), k, diffs[k])
		}
	}
	if n == 0 {
		t.Skip("no game data")
	}
}

// upstreamSaveNormalised is a GameState as upstream's save would leave it
// unchanged: EntityManager::zeroUnused (free slots zeroed but their links and
// id), every entity's sprite bounds Location::null (exportEntityBase),
// StationManager::zeroUnused, OrderManager::zeroUnusedOrderTable.
func upstreamSaveNormalised(raw []byte) []byte {
	gs := append([]byte(nil), raw...)
	for id := range MaxEntities {
		e := gs[EntityTableOffset+id*EntitySize : EntityTableOffset+(id+1)*EntitySize]
		switch e[ebBaseType] {
		case EntityBaseTypeFree:
			b, next, prev, off := e[0], e[4:6], e[6:8], e[8]
			nx, pv := append([]byte(nil), next...), append([]byte(nil), prev...)
			clear(e)
			e[0], e[8] = b, off
			copy(e[4:], nx)
			copy(e[6:], pv)
			put16(e, ebID, uint16(id))
		case EntityBaseTypeVehicle, EntityBaseTypeEffect:
			for _, o := range []int{ebSpriteLeft, ebSpriteTop, ebSpriteRight, ebSpriteBottom} {
				put16(e, o, 0x8000)
			}
		}
	}
	for i := range stationMaxCount {
		r := gs[stationsGameStateOffset+i*stationRecordSize : stationsGameStateOffset+(i+1)*stationRecordSize]
		if r[0] == 0xFF && r[1] == 0xFF {
			clear(r)
			put16(r, 0, stationNameNull)
			for c := range 32 {
				put16(r, 0x2E+c*13+2, 0xFFFF)
			}
			continue
		}
		n := int(r[0x1CE]) | int(r[0x1CF])<<8
		for k := n; k < 80; k++ {
			clear(r[0x1D0+6*k : 0x1D0+6*k+6])
		}
	}
	ol := int(gs[0x1A0]) | int(gs[0x1A1])<<8 | int(gs[0x1A2])<<16 | int(gs[0x1A3])<<24
	clear(gs[gameStateOrders+ol : gameStateOrders+maxOrders])
	return gs
}
