package scenario

import (
	"fmt"
	"sort"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// s5compare.go — a written game compared with the file it was loaded from,
// for the tests (here and in pkg/sim): what differs, by GameState section and
// offset in its record, once the file's bytes are normalised as upstream's
// save would leave them unchanged.

// DiffFromFile lists (at most max) the places where gs (a GameState) and
// tiles (a TileElements chunk; nil: not compared) differ from sc.Raw's, as
// "section +offset ×count". loaded: the file as a load leaves it
// (TileManager::resetSurfaceClearance: a flat surface's clearZ its baseZ).
func (sc *Scenario) DiffFromFile(gs, tiles []byte, max int, loaded bool) []string {
	want := sc.Raw.GameState
	if sc.Header.Type == assets.S5TypeSavedGame {
		want = upstreamSaveNormalised(want)
	}
	diffs := map[string]int{}
	for i := range min(len(gs), len(want)) {
		if gs[i] != want[i] {
			diffs[gsSection(i)]++
		}
	}
	if tiles != nil {
		raw := sc.Raw.TileElements
		if loaded {
			raw = append([]byte(nil), raw...)
			for i := 0; i < len(raw); i += tileElementSize {
				if (raw[i]&elementTypeMask)>>elementTypeShift == KindSurface && raw[i+4]&0x1F == 0 {
					raw[i+3] = raw[i+2]
				}
			}
		}
		if len(tiles) != len(raw) {
			diffs[fmt.Sprintf("tile elements: %d bytes, want %d", len(tiles), len(raw))]++
		} else {
			for i := 0; i < len(raw); i += tileElementSize {
				kind := (raw[i] & elementTypeMask) >> elementTypeShift
				for b := range tileElementSize {
					if x := tiles[i+b] ^ raw[i+b]; x != 0 && !upstreamDrops(kind, b, x) {
						diffs[fmt.Sprintf("tile kind %d byte %d mask %02x", kind, b, x)]++
					}
				}
			}
		}
	}
	var keys []string
	for k := range diffs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []string
	for i, k := range keys {
		if i >= max {
			out = append(out, fmt.Sprintf("… %d more", len(keys)-i))
			break
		}
		out = append(out, fmt.Sprintf("%s ×%d", k, diffs[k]))
	}
	return out
}

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
