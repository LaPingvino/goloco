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
	// Rivers (OpenLoco's options; none in vanilla scenarios).
	NumRiverbeds, MinRiverWidth, MaxRiverWidth, RiverbankWidth, RiverMeanderRate uint8
	LandDistributionPatterns                                                     [32]uint8 // per land object slot
	// Trees.
	NumberOfForests, NumberRandomTrees                                                                             uint16
	MinForestRadius, MaxForestRadius, MinForestDensity, MaxForestDensity, MinAltitudeForTrees, MaxAltitudeForTrees uint8
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

// GenerateRivers is generateRivers: each riverbed a band carved to just below
// sea level (banks sloping), north–south or east–west, meandering every 4 rows.
func (m *Map) GenerateRivers(o TerrainOptions, hm *HeightMap) {
	for range int(o.NumRiverbeds) {
		eastWest := m.Rng1.RandBool()
		riverWidth := int(m.Rng1.RandNextRange(int32(o.MinRiverWidth), int32(o.MaxRiverWidth)))
		bed := uint8(1)
		if o.SeaLevel > 0 {
			bed = o.SeaLevel - 1
		}
		bed = max(bed, o.MinLandHeight)
		bank := o.RiverbankWidth // auto from a uint8_t: it wraps as it meanders
		total := riverWidth + 2*int(bank)
		eastOffset := riverWidth + int(bank)
		xStart := int(m.Rng1.RandNextRange(int32(0.15*float64(hm.Width)), int32(0.85*float64(hm.Width))))
		for yPos := 0; yPos < hm.Height; yPos++ {
			for xOff := 0; xOff < total; xOff++ {
				px, py := xStart+xOff, yPos
				if !eastWest {
					px, py = py, px
				}
				if int16(px) < 0 || int(int16(px)) >= hm.Width || int16(py) < 0 || int(int16(py)) >= hm.Height { // validCoords (tile_coord_t)
					break
				}
				px, py = int(int16(px)), int(int16(py))
				h := hm.At(px, py)
				switch {
				case bank > 0 && xOff < int(bank):
					*h = max(bed, uint8(int(*h)*(int(bank)-xOff)/int(bank)))
				case bank > 0 && xOff > eastOffset:
					*h = max(bed, uint8(int(*h)*(xOff-eastOffset)/int(bank)))
				default:
					*h = bed
				}
			}
			if rate := int(o.RiverMeanderRate); rate > 0 && yPos%4 == 0 {
				half := rate / 2
				off := int(int8(m.Rng1.RandNextRange(0, int32(rate)) - int32(half)))
				xStart += off
				if o.RiverbankWidth > 0 && half != 0 {
					bank += uint8(off / half)
					eastOffset += off / half
					total += off / half * 2
				}
			}
		}
	}
}

// GenerateWater is generateWater: every drawable surface below sea level
// (baseZ < seaLevel × 4) gets water at sea level.
func (m *Map) GenerateWater(seaLevel uint8) {
	for y := 1; y <= m.Height-2; y++ {
		for x := 1; x <= m.Width-2; x++ {
			if t := m.TileAt(x, y); t != nil && int(t.BaseZ) < int(seaLevel)<<2 {
				t.WaterLevel = seaLevel & 0x1F // setWater keeps the update timer bits; goloco stores only the level
			}
		}
	}
}

// Land distribution patterns (Scenario::LandDistributionPattern).
const (
	landEverywhere = iota
	landNowhere
	landFarFromWater
	landNearWater
	landOnMountains
	landFarFromMountains
	landSmallRandomAreas
	landLargeRandomAreas
	landAroundCliffs
)

const (
	cliffTerrainHeightDiff = 4  // kCliffTerrainHeightDiff
	mountainTerrainHeight  = 26 // kMountainTerrainHeight
)

// randomTerrainVariation is getRandomTerrainVariation (gPrng1): none on
// water / industrial / unknown land; 0 on slopes or land without variations;
// else a variation by the land object's likelihood.
func (m *Map) randomTerrainVariation(t *Tile, om *objects.ObjectManager) (uint8, bool) {
	if t.WaterLevel != 0 || t.IsIndustrial || int(t.TerrainIndex) >= len(om.LandObjects) || om.LandObjects[t.TerrainIndex] == nil {
		return 0, false
	}
	lo := om.LandObjects[t.TerrainIndex]
	if lo.NumVariations == 0 || t.Slope&0x1F != 0 { // slope(): the snow coverage bits excluded
		return 0, true
	}
	r := uint16(m.Rng1.RandNext())
	if uint16(lo.VariationLikelihood) <= r>>8 {
		return 0, true
	}
	return uint8(uint32(r&0xFF) * uint32(lo.NumVariations) >> 8), true
}

func (m *Map) setTerrainWithVariation(t *Tile, style uint8, om *objects.ObjectManager) {
	t.TerrainIndex = style
	if v, ok := m.randomTerrainVariation(t, om); ok {
		t.Variation = v
	}
}

// markAround marks every tile within r of (x, y), clamped to the map
// (getClampedRange).
func markAround(hm *HeightMap, x, y, r int) {
	for my := max(0, y-r); my <= min(hm.Height-1, y+r); my++ {
		for mx := max(0, x-r); mx <= min(hm.Width-1, x+r); mx++ {
			hm.SetMarker(mx, my)
		}
	}
}

// applySurfaceStyleToMarkedTiles over the drawable tiles.
func (m *Map) applySurfaceStyleToMarkedTiles(hm *HeightMap, style uint8, requireMark bool, om *objects.ObjectManager) {
	for y := 1; y <= m.Height-2; y++ {
		for x := 1; x <= m.Width-2; x++ {
			if hm.IsMarkerSet(x, y) == requireMark {
				m.setTerrainWithVariation(m.TileAt(x, y), style, om)
			}
		}
	}
}

// markWhere marks around every world tile whose height passes test.
func markWhere(hm *HeightMap, r int, test func(h uint8) bool) {
	hm.ResetMarkerFlags()
	for y := range hm.Height {
		for x := range hm.Width {
			if test(hm.GetHeight(x, y)) {
				markAround(hm, x, y, r)
			}
		}
	}
}

// generateTerrainInRandomAreas: 80–143 random walks of minTiles … maxTiles−1
// steps, each step above sea level taking the style.
func (m *Map) generateTerrainInRandomAreas(style uint8, minTiles, maxTiles int32, seaLevel uint8, om *objects.ObjectManager) {
	n := m.Rng1.RandNextRange(80, 143)
	for range n {
		r := m.Rng1.RandNext()
		px := int32(((r & 0xFFFF) * uint32(m.Width)) >> 16 * 32) // toWorldSpace
		py := int32(((r >> 16) * uint32(m.Height)) >> 16 * 32)
		steps := m.Rng1.RandNextRange(minTiles, maxTiles-1)
		for range steps {
			if px >= 0 && px < int32(m.Width*32) && py >= 0 && py < int32(m.Height*32) {
				t := m.TileAt(int(px/32), int(py/32))
				if t != nil && t.SurfaceClearZ()/microToSmallZStep > seaLevel { // getSurfaceCornerHeight / 4
					m.setTerrainWithVariation(t, style, om)
				}
			}
			d := RotationOffset[m.Rng1.RandNextBound(3)]
			px += int32(d[0])
			py += int32(d[1])
		}
	}
}

// GenerateTerrain is generateTerrain: every drawable surface the
// "everywhere" land (else lastLandOption), growth 0, a variation; then each
// pattern in turn for the land objects that use it.
func (m *Map) GenerateTerrain(hm *HeightMap, patterns [32]uint8, seaLevel uint8, om *objects.ObjectManager) {
	style, ok := uint8(0), false
	for i, lo := range om.LandObjects {
		if lo != nil && i < 32 && patterns[i] == landEverywhere {
			style, ok = uint8(i), true
			break
		}
	}
	if !ok && m.LastLandOption != 0xFF {
		style, ok = m.LastLandOption, true
	}
	if !ok {
		return
	}
	for y := 1; y <= m.Height-2; y++ {
		for x := 1; x <= m.Width-2; x++ {
			t := m.TileAt(x, y)
			t.TerrainIndex, t.GrowthStage = style, 0
			if v, ok := m.randomTerrainVariation(t, om); ok {
				t.Variation = v
			}
		}
	}
	for _, p := range []int{landFarFromWater, landNearWater, landOnMountains, landFarFromMountains, landSmallRandomAreas, landLargeRandomAreas, landAroundCliffs} {
		for i, lo := range om.LandObjects {
			if lo == nil || i >= 32 || int(patterns[i]) != p {
				continue
			}
			s := uint8(i)
			switch p {
			case landFarFromWater:
				markWhere(hm, 25, func(h uint8) bool { return h <= seaLevel })
				m.applySurfaceStyleToMarkedTiles(hm, s, false, om)
			case landNearWater:
				markWhere(hm, 25, func(h uint8) bool { return h >= seaLevel })
				m.applySurfaceStyleToMarkedTiles(hm, s, true, om)
			case landOnMountains:
				markWhere(hm, 12, func(h uint8) bool { return h > mountainTerrainHeight })
				m.applySurfaceStyleToMarkedTiles(hm, s, true, om)
			case landFarFromMountains:
				markWhere(hm, 25, func(h uint8) bool { return h >= mountainTerrainHeight })
				m.applySurfaceStyleToMarkedTiles(hm, s, false, om)
			case landSmallRandomAreas:
				m.generateTerrainInRandomAreas(s, 24, 56, seaLevel, om)
			case landLargeRandomAreas:
				m.generateTerrainInRandomAreas(s, 128, 388, seaLevel, om)
			case landAroundCliffs:
				// generateTerrainAroundCliffs compares the same position, (x, y+1),
				// with itself (A = B, C = D), so it never finds a cliff and marks
				// nothing — kept as upstream.
				hm.ResetMarkerFlags()
				m.applySurfaceStyleToMarkedTiles(hm, s, true, om)
			}
		}
	}
}

// Seasons (Scenario::Season): autumn, winter, spring, summer.
const seasonWinter = 1

// InitialiseSnowLine is initialiseSnowLine: the season for today's day of
// the year from the climate (updateSeason), and that season's snow line.
func (m *Map) InitialiseSnowLine(om *objects.ObjectManager) {
	c := om.Climate
	if c == nil {
		return
	}
	day := CalcDate(m.CurrentDay).DayOfYear
	season := c.FirstSeason
	for _, l := range c.SeasonLength {
		day -= int(l)
		if day < 0 {
			break
		}
		season = (season + 1) % 4 // nextSeason: autumn → winter → spring → summer → autumn
	}
	m.CurrentSeason = season
	if season == seasonWinter {
		m.CurrentSnowLine = c.WinterSnowLine
	} else {
		m.CurrentSnowLine = c.SummerSnowLine
	}
}

// GenerateSurfaceVariation is generateSurfaceVariation: growth stages for
// land that has them (0 under water above the surface, unless terrain 0;
// else the last stage) and snow coverage from the snow line (0–5, kept in
// the slope byte's top bits as upstream).
func (m *Map) GenerateSurfaceVariation(om *objects.ObjectManager) {
	snow := int(m.CurrentSnowLine) / microToSmallZStep
	for y := 1; y <= m.Height-2; y++ {
		for x := 1; x <= m.Width-2; x++ {
			t := m.TileAt(x, y)
			if !t.IsIndustrial && int(t.TerrainIndex) < len(om.LandObjects) && om.LandObjects[t.TerrainIndex] != nil {
				lo := om.LandObjects[t.TerrainIndex]
				if lo.Flags&objects.LandFlagHasGrowthStages != 0 {
					set := false
					if t.WaterLevel != 0 {
						wz := int(t.WaterLevel) * microToSmallZStep
						if t.Slope&0x1F != 0 {
							wz -= 4
						}
						if wz > int(t.BaseZ) && t.TerrainIndex != 0 {
							t.GrowthStage, set = 0, true
						}
					}
					if !set {
						t.GrowthStage = lo.NumGrowthStages - 1
					}
				}
			}
			cover := min(max(int(t.BaseZ)/microToSmallZStep+1-snow, 0), 5)
			t.Slope = t.Slope&^0xE0 | uint8(cover)<<5
		}
	}
}
