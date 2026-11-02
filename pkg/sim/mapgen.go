package sim

import (
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
}

// GenerateLandscape is MapGenerator::generate as far as it is ported:
// initialiseDate, TileManager::initialise, the height map (original
// generator), rivers, generateLand, water. PARTIAL: Scenario::reset, the snow
// line, terrain styles, surface variation, trees, towns, industries, misc
// buildings, sub_4969E0 / sub_4748D4 — so a generated map is bare land of one
// type for now.
func (m *Map) GenerateLandscape(startYear uint16, o TerrainOptions, g1 G1Raw, om *objects.ObjectManager) {
	m.InitialiseDate(startYear)
	m.InitialiseTiles()
	hm := NewHeightMap(m.Width, m.Height)
	m.GenerateTerrainHeightMap(o, om.HillShapes, g1, hm)
	m.GenerateRivers(o, hm)
	m.GenerateLand(hm)
	m.GenerateWater(o.SeaLevel)
}
