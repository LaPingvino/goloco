package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Map generation — Map/MapGenerator (item 5.9). This file: HeightMap and
// generateLand (the height map's corner heights → each surface's baseZ and
// slope). Not yet ported: the height map itself (OriginalTerrainGenerator:
// hill shapes blitted from the HillShapesObject's G1 images, gPrng1),
// generateRivers, generateWater, generateTerrain (surface styles),
// generateSurfaceVariation, trees, towns, industries, misc buildings;
// pkg/scenario/generate.go stays goloco's stand-in until then.

const (
	heightMapMarked   = 1 << 7 // kHeightmapMarkedFlag
	microToSmallZStep = 4      // kMicroToSmallZStep
	smallZStep        = 4      // kSmallZStep
	slopeCornerNorth  = 1 << 0 // SurfaceSlope::CornerUp
	slopeCornerEast   = 1 << 1
	slopeCornerSouth  = 1 << 2
	slopeCornerWest   = 1 << 3
	slopeCornersAll   = 0x0F
	slopeDoubleHeight = 1 << 4
)

// HeightMap is MapGenerator's HeightMap: one MicroZ per tile corner, the top
// bit a marker.
type HeightMap struct {
	Width, Height int
	data          []uint8
}

// NewHeightMap makes a zeroed height map.
func NewHeightMap(w, h int) *HeightMap { return &HeightMap{w, h, make([]uint8, w*h)} }

// At is operator[] (the raw byte, marker included).
func (hm *HeightMap) At(x, y int) *uint8 { return &hm.data[y*hm.Width+x] }

// GetHeight is getHeight: the height without the marker.
func (hm *HeightMap) GetHeight(x, y int) uint8 { return *hm.At(x, y) &^ heightMapMarked }

// IsMarkerSet, SetMarker, UnsetMarker, ResetMarkerFlags: the marker bit.
func (hm *HeightMap) IsMarkerSet(x, y int) bool { return *hm.At(x, y)&heightMapMarked != 0 }
func (hm *HeightMap) SetMarker(x, y int)        { *hm.At(x, y) |= heightMapMarked }
func (hm *HeightMap) UnsetMarker(x, y int)      { *hm.At(x, y) &^= heightMapMarked }
func (hm *HeightMap) ResetMarkerFlags() {
	for i := range hm.data {
		hm.data[i] &^= heightMapMarked
	}
}

// SurfaceClearZ is a surface's clearZ as generateLand (and the terraform
// commands) set it: baseZ, one step up for any raised corner, one more when
// double height.
func (t *Tile) SurfaceClearZ() uint8 {
	z := t.BaseZ
	if t.Slope&slopeCornersAll != 0 {
		z += smallZStep
	}
	if t.Slope&slopeDoubleHeight != 0 {
		z += smallZStep
	}
	return z
}

// GenerateLand is generateLand: over the drawable tiles (1 … size − 2), each
// surface takes the lowest of its four corner heights as baseZ and a raised
// corner for each higher one; a single lowered corner two or more below the
// rest makes the slope double height. Corners: q00 = (x−1, y−1) south,
// q01 = (x, y−1) east, q10 = (x−1, y) west, q11 = (x, y) north.
func (m *Map) GenerateLand(hm *HeightMap) {
	for y := 1; y <= m.Height-2; y++ {
		for x := 1; x <= m.Width-2; x++ {
			t := m.TileAt(x, y)
			if t == nil {
				continue
			}
			q00, q01 := int(*hm.At(x-1, y-1)), int(*hm.At(x, y-1))
			q10, q11 := int(*hm.At(x-1, y)), int(*hm.At(x, y))
			base := min(q00, q01, q10, q11)
			t.BaseZ = uint8(base * microToSmallZStep)
			var slope uint8
			if q00 > base {
				slope |= slopeCornerSouth
			}
			if q01 > base {
				slope |= slopeCornerEast
			}
			if q10 > base {
				slope |= slopeCornerWest
			}
			if q11 > base {
				slope |= slopeCornerNorth
			}
			if (slope == slopeCornersAll&^slopeCornerNorth && q00-base >= 2) ||
				(slope == slopeCornersAll&^slopeCornerWest && q01-base >= 2) ||
				(slope == slopeCornersAll&^slopeCornerEast && q10-base >= 2) ||
				(slope == slopeCornersAll&^slopeCornerSouth && q11-base >= 2) {
				slope |= slopeDoubleHeight
			}
			t.Slope = slope
		}
	}
}

// InitialiseTiles is TileManager::initialise: every tile a single surface of
// the last land option (else land object 0) at baseZ 4.
func (m *Map) InitialiseTiles() {
	land := m.LastLandOption
	if land == 0xFF {
		land = 0
	}
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			m.Tiles[y][x] = Tile{TerrainIndex: land, BaseZ: 4, Order: []scenario.ElemRef{{Kind: scenario.KindSurface}}}
		}
	}
	m.EntriesEnd = m.Width * m.Height // one surface element each
}

// GenerateLandscape is MapGenerator::generate as far as it is ported:
// initialiseDate, TileManager::initialise, the height map (original
// generator), rivers, generateLand, water, terrain styles, the snow line,
// surface variation, trees, towns, industries, misc buildings. PARTIAL:
// Scenario::reset, sub_4969E0 / sub_4748D4.
func (m *Map) GenerateLandscape(startYear uint16, o TerrainOptions, g1 G1Raw, om *objects.ObjectManager) {
	m.InitialiseDate(startYear)
	m.InitialiseSnowLine(om)
	m.InitialiseTiles()
	hm := NewHeightMap(m.Width, m.Height)
	m.GenerateTerrainHeightMap(o, om.HillShapes, g1, hm)
	m.GenerateRivers(o, hm)
	m.GenerateLand(hm)
	m.GenerateWater(o.SeaLevel)
	m.GenerateTerrain(hm, o.LandDistributionPatterns, o.SeaLevel, om)
	m.GenerateSurfaceVariation(om)
	m.GenerateTrees(o, om)
	m.GenerateTowns(o, g1, om)
	m.GenerateIndustries(o, om)
	m.GenerateMiscBuildings(om)
	m.GenerateSurfaceVariation(om)
}

// Surface slope shapes (SurfaceSlope).
const (
	slopeSideNE    = slopeCornerNorth | slopeCornerEast
	slopeSideSE    = slopeCornerSouth | slopeCornerEast
	slopeSideNW    = slopeCornerNorth | slopeCornerWest
	slopeSideSW    = slopeCornerSouth | slopeCornerWest
	slopeValleyNS  = slopeCornerNorth | slopeCornerSouth
	slopeValleyWE  = slopeCornerWest | slopeCornerEast
	slopeDownWest  = slopeCornersAll &^ slopeCornerWest
	slopeDownSouth = slopeCornersAll &^ slopeCornerSouth
	slopeDownEast  = slopeCornersAll &^ slopeCornerEast
	slopeDownNorth = slopeCornersAll &^ slopeCornerNorth
	tileSize       = 32
	microZStep     = 16 // kMicroZStep
)

// GetHeight is TileManager::getHeight: the land height (world Z) at a point
// on the map, following the surface's slope, and the water height; off the
// map (land 16, water 0).
func (m *Map) GetHeight(px, py int32) (land, water int16) {
	if uint32(px) >= uint32(m.Width*tileSize-1) || uint32(py) >= uint32(m.Height*tileSize-1) {
		return 16, 0
	}
	t := m.TileAt(int(px/tileSize), int(py/tileSize))
	if t == nil {
		return 16, 0
	}
	water = int16(t.WaterLevel) * microZStep
	land = int16(t.BaseZ) * smallZStep
	return land + surfaceLandHeight(t.Slope, int16(px&0x1F), int16(py&0x1F)), water
}

// surfaceLandHeight is getHeight's slope switch and its four helpers.
func surfaceLandHeight(slopeByte uint8, xl, yl int16) int16 {
	slope := slopeByte & slopeCornersAll
	half := func(q int16) int16 {
		if q > 0 {
			return q / 2
		}
		return 0
	}
	switch slope {
	case slopeCornerNorth:
		return half(xl + yl - tileSize)
	case slopeCornerEast:
		return half(xl - yl)
	case slopeCornerSouth:
		return half(tileSize - yl - xl)
	case slopeCornerWest:
		return half(yl - xl)
	case slopeSideNE:
		return xl/2 + 1
	case slopeSideSE:
		return (tileSize - yl) / 2
	case slopeSideNW:
		return yl/2 + 1
	case slopeSideSW:
		return (tileSize - xl) / 2
	case slopeDownWest, slopeDownSouth, slopeDownEast, slopeDownNorth:
		var extra, q int16
		switch slope {
		case slopeDownWest:
			extra, q = xl+tileSize-yl, xl-yl
		case slopeDownSouth:
			extra, q = xl+yl, xl+yl-tileSize-1
		case slopeDownEast:
			extra, q = tileSize-xl+yl, yl-xl
		case slopeDownNorth:
			extra, q = (tileSize-xl)+(tileSize-yl), tileSize-yl-xl-1
		}
		if slopeByte&slopeDoubleHeight != 0 {
			return extra/2 + 1
		}
		return q/2 + 16
	case slopeValleyWE:
		if xl+yl > tileSize+1 {
			return half(tileSize - xl - yl)
		}
		return 0
	case slopeValleyNS:
		return half(xl - yl)
	}
	return 0
}

// EditorDefaultTerrainOptions is EditorController::init's landscape options
// (minLandHeight 2, mountains, hill density 50, sea level 4, the forest and
// tree counts, the river options) with resetLandDistributionPatterns: each
// land object's own pattern, everywhere (0) for an empty slot.
func EditorDefaultTerrainOptions(om *objects.ObjectManager) TerrainOptions {
	o := TerrainOptions{MinLandHeight: 2, TopographyStyle: 2 /* TopographyStyle::mountains */, HillDensity: 50, SeaLevel: 4,
		NumRiverbeds: 0, MinRiverWidth: 10, MaxRiverWidth: 20, RiverbankWidth: 5, RiverMeanderRate: 10,
		NumberOfForests: 100, MinForestRadius: 4, MaxForestRadius: 40, MinForestDensity: 1, MaxForestDensity: 7,
		NumberRandomTrees: 1000, MinAltitudeForTrees: 0, MaxAltitudeForTrees: 25,
		NumberOfTowns: 55, MaxTownSize: 3, NumberOfIndustries: 1}
	if om != nil {
		for i := range o.LandDistributionPatterns {
			if l := om.GetLandObjectByIndex(i); l != nil {
				o.LandDistributionPatterns[i] = l.DistributionPattern
			}
		}
	}
	return o
}

// NewGeneratedMap is a map of the vanilla size (kMapColumns × kMapRows,
// 384 × 384) with the landscape generated (GenerateLandscape) on an rng
// seeded with seed0 / seed1 — the open build's stand-in for title.dat.
func NewGeneratedMap(seed0, seed1 uint32, startYear uint16, o TerrainOptions, g1 G1Raw, om *objects.ObjectManager) *Map {
	m := &Map{Width: 384, Height: 384}
	m.Rng1 = *core.NewPrngSeed(seed0, seed1)
	if om != nil && om.HillShapes == nil { // no scenario chose one: the object selection's (the only one loaded, upstream)
		for _, h := range om.HillShapesObjects {
			if h != nil {
				om.HillShapes = h
				break
			}
		}
	}
	m.Tiles = make([][]Tile, m.Height)
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, m.Width)
	}
	if om != nil && om.Climate == nil {
		// OpenGraphics v0.1.8 has no climate object (its Climate/ folder is
		// empty): upstream can't run without one; the open build has no snow
		// line rather than one at 0 (every tile "above" it). Set before the
		// generation (the terrain takes its snow from it; InitialiseSnowLine
		// leaves it alone without a climate).
		m.CurrentSnowLine = 0xFF
	}
	m.GenerateLandscape(startYear, o, g1, om)
	return m
}
