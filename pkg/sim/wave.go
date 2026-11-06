package sim

// Waves on open water — a port of OpenLoco Map/WaveManager.cpp and the water
// branch of the surface tile update (Map/SurfaceElement.cpp update, Map/
// TileManager.cpp update). A wave lives on one water tile whose four
// neighbours are water too, runs 16 frames (one per 4 scenario ticks), and
// marks its tile with WaveFlag (upstream: SurfaceElement flag 6) while it lives;
// the surface painter draws wave frame +60 only on flagged tiles.

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// MaxWaves is the size of the wave table (OpenLoco S5/Limits.h kMaxWaves).
const MaxWaves = 64

// locationNull marks an empty slot (OpenLoco Location::null = int16 0x8000).
const locationNull = -32768

// TileSize is one tile in world units.
const TileSize = 32

// Wave is one wave slot (OpenLoco Map/Wave.h): world position and frame.
type Wave struct {
	X, Y  int // world units; X == locationNull: empty
	Frame uint16
}

// Empty reports whether the slot is free.
func (w Wave) Empty() bool { return w.X == locationNull }

// WaveIndex is the slot for a tile (OpenLoco WaveManager.h getWaveIndex).
func WaveIndex(tileX, tileY int) int { return (tileX & 7) | ((tileY & 7) << 3) }

// ResetWaves empties every slot (OpenLoco WaveManager::reset, 0x004C4BC0).
func (s *State) ResetWaves() {
	for i := range s.Waves {
		s.Waves[i].X = locationNull
	}
}

// createWave tries to start a wave on the water tile at (tileX, tileY)
// (OpenLoco WaveManager::createWave, 0x0046956E). visible reports whether the
// tile is shown in a viewport; upstream only starts waves where someone looks.
func (s *State) createWave(t *Tile, tileX, tileY int, visible func(tileX, tileY int) bool) {
	w := &s.Waves[WaveIndex(tileX, tileY)]
	if !w.Empty() {
		return
	}
	if visible != nil && !visible(tileX, tileY) {
		return
	}
	if s.Rng2.RandNext()&0xFFFF > 0x1745 {
		return
	}
	// Every orthogonal neighbour must be a water tile.
	for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		n := s.TileAt(tileX+d[0], tileY+d[1])
		if n == nil || n.WaterLevel == 0 {
			return
		}
	}
	w.X, w.Y, w.Frame = tileX*TileSize, tileY*TileSize, 0
	t.WaveFlag = true
}

// UpdateWaves advances every wave one frame every 4 scenario ticks and frees
// it after 16 frames or when its tile has no water (WaveManager::update,
// 0x004C56F6).
func (s *State) UpdateWaves(scenarioTicks uint32) {
	if scenarioTicks&3 != 0 {
		return
	}
	for i := range s.Waves {
		w := &s.Waves[i]
		if w.Empty() {
			continue
		}
		t := s.TileAt(w.X/TileSize, w.Y/TileSize)
		if t == nil {
			w.X = locationNull
			continue
		}
		if t.WaterLevel != 0 {
			w.Frame++
			if w.Frame < 16 {
				continue
			}
		}
		w.X = locationNull
		t.WaveFlag = false
	}
}

// UpdateTiles runs one pass of the tile loop: every 16th tile in both
// directions from a start offset that moves through a 16×16 block, so each tile
// is visited once per 256 ticks (OpenLoco TileManager::update, 0x004612EC).
//
// Each visited tile's elements are updated in their stored order (ghosts
// skipped), stopping at one that changed the tile: the surface (PARTIAL: of
// SurfaceElement::update only the water branch — a wave on a water tile that
// has none; snow, the type-6 counter and growth not yet), buildings
// (BuildingElement::update, buildingupdate.go) and industry elements (their
// stations in range, industryprod.go). PARTIAL: trees, roads, and the
// industry elements' construction and animation.
func (s *State) UpdateTiles(visible func(tileX, tileY int) bool, om *objects.ObjectManager) {
	for ty := s.tileUpdateY; ty < s.Height; ty += 16 {
		for tx := s.tileUpdateX; tx < s.Width; tx += 16 {
			t := s.TileAt(tx, ty)
			if t == nil {
				continue
			}
		elements:
			for _, r := range t.Order {
				if r.Flags&0x10 != 0 { // ghost
					continue
				}
				switch r.Kind {
				case scenario.KindSurface:
					if t.WaterLevel != 0 && !t.WaveFlag {
						s.createWave(t, tx, ty, visible)
					}
				case scenario.KindBuilding:
					if int(r.Index) < len(t.Buildings) && !s.updateBuilding(&t.Buildings[r.Index], false, int32(tx)*32, int32(ty)*32, om) {
						break elements // the tile's elements changed
					}
				case scenario.KindIndustry:
					if int(r.Index) < len(t.Industries) {
						s.industryElementStations(&t.Industries[r.Index], int32(tx)*32, int32(ty)*32, om)
					}
				}
			}
		}
	}
	shift := uint8(s.tileUpdateY<<4 + s.tileUpdateX + 9)
	s.tileUpdateX, s.tileUpdateY = int(shift&0xF), int(shift>>4)
	if shift == 0 { // a whole round: IndustryManager::updateProducedCargoStats
		s.industriesUpdateProducedCargoStats(om)
	}
}

// NewWaveRng returns the generator the wave code draws from (OpenLoco gPrng2 =
// GameState unkRng; seeded from the scenario when the loader reads it).
func NewWaveRng() core.Prng { return *core.NewPrngSeed(0, 0) }
