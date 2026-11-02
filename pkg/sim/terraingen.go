package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// OriginalTerrainGenerator (Map/MapGenerator/OriginalTerrainGenerator.cpp):
// the height map from the hill shapes object — one whole-map height map
// image, or random hills / mountains stamped by gPrng1 — then the sea-level
// caps.

// G1Raw reads a G1 element's raw bytes (assets.G1File.ElementRaw).
type G1Raw interface {
	ElementRaw(i int) (w, h int, data []byte, ok bool)
}

// TerrainOptions is what the generator reads of Scenario::Options and the
// game state.
type TerrainOptions struct {
	MinLandHeight   uint8
	TopographyStyle uint8
	HillDensity     uint8
	HillsEdgeOfMap  bool // ScenarioFlags::hillsEdgeOfMap
	SeaLevel        uint8
}

const (
	topoHasHills     = 1 << 0
	topoHasMountains = 1 << 1
	mapPitch         = 512 // kMapPitch
)

var topographyStyleFlags = [10]uint8{0, 0, topoHasHills, topoHasHills, topoHasMountains | topoHasHills,
	topoHasMountains | topoHasHills, topoHasHills, topoHasMountains | topoHasHills, 0, topoHasMountains | topoHasHills}

// GenerateTerrainHeightMap is OriginalTerrainGenerator::generate.
func (m *Map) GenerateTerrainHeightMap(o TerrainOptions, hs *objects.HillShapesObject, g1 G1Raw, hm *HeightMap) {
	for i := range hm.data {
		hm.data[i] = o.MinLandHeight
	}
	if hs != nil && hs.Flags&objects.HillShapeIsHeightMap != 0 {
		if _, _, src, ok := g1.ElementRaw(int(hs.Image)); ok {
			k := 0
			for y := hm.Height - 1; y > 0; y-- {
				for x := hm.Width - 1; x > 0; x-- {
					if k < len(src) {
						*hm.At(x, y) = max(src[k], *hm.At(x, y))
					}
					k++
				}
				k += mapPitch
			}
		}
	} else if hs != nil {
		n := int(uint8(m.Rng1.RandNext()&7)) + int(o.HillDensity)*5 + 1
		for range uint16(n) {
			m.blitHill(o, hs, g1, hm)
		}
	}
	// capSeaLevels
	for y := 0; y < hm.Height-2; y++ {
		for x := 0; x < hm.Width-2; x++ {
			if *hm.At(x, y) == o.SeaLevel && *hm.At(x+1, y) == o.SeaLevel && *hm.At(x, y+1) == o.SeaLevel && *hm.At(x+1, y+1) == o.SeaLevel {
				*hm.At(x, y)++
				*hm.At(x+1, y)++
				*hm.At(x, y+1)++
				*hm.At(x+1, y+1)++
			}
		}
	}
}

// blitHill is blitHill: one random hill or mountain shape, maybe flipped,
// at a random place (max with the height already there).
func (m *Map) blitHill(o TerrainOptions, hs *objects.HillShapesObject, g1 G1Raw, hm *HeightMap) {
	r := m.Rng1.RandNext()
	flipLateral, flipVertical := r&0x100 != 0, r&0x200 != 0
	count := hs.HillCount + hs.MountainCount
	if count == 0 {
		return
	}
	idx := uint8(r&0xFF) % count
	fw, fh, src, ok := g1.ElementRaw(int(hs.Image) + int(idx))
	if !ok {
		return
	}
	fw, fh = fw&0xFF, fh&0xFF
	if flipLateral {
		fw, fh = fh, fw
	}
	randX := uint32(m.Rng1.RandNextBound(int32(hm.Width - 1)))
	randY := uint32(m.Rng1.RandNextBound(int32(hm.Height - 1)))
	if randX < 2 || randY < 2 {
		return
	}
	topo := uint32(o.TopographyStyle) << 1
	if randX > randY {
		topo |= 1
	}
	if int(topo) >= len(topographyStyleFlags) {
		return
	}
	f := topographyStyleFlags[topo]
	validHill := f&topoHasHills != 0 && idx < hs.HillCount
	validMountain := f&topoHasMountains != 0 && idx >= hs.HillCount
	if !validHill && !validMountain {
		return
	}
	if !o.HillsEdgeOfMap && (randX+uint32(fw) >= uint32(hm.Width-2) || randY+uint32(fh) >= uint32(hm.Height-2)) {
		return
	}
	k := 0
	put := func(x, y int) {
		if k < len(src) {
			*hm.At(x, y) = max(src[k], *hm.At(x, y))
		}
		k++
	}
	mw, mh := hm.Width-1, hm.Height-1
	switch {
	case flipLateral && flipVertical: // blitImageFlippedXFlippedY
		x := int(randX)
		for range fw {
			y := (fh + int(randY) - 1) % mh
			for range fh {
				put(x, y)
				y = (y - 1 + mh) % mh
			}
			x = (x + 1) % mw
		}
	case flipLateral: // blitImageFlippedXNormalY
		x := int(randX)
		for range fw {
			y := int(randY)
			for range fh {
				put(x, y)
				y = (y + 1) % mh
			}
			x = (x + 1) % mw
		}
	case flipVertical: // blitImageNormalXFlippedY
		y := int(randY)
		for range fh {
			x := (int(randX) + fw - 1) % mw
			for range fw {
				put(x, y)
				x = (x - 1 + mw) % mw
			}
			y = (y + 1) % mh
		}
	default: // blitImageNormalXNormalY
		y := int(randY)
		for range fh {
			x := int(randX)
			for range fw {
				put(x, y)
				x = (x + 1) % mw
			}
			y = (y + 1) % mh
		}
	}
}
