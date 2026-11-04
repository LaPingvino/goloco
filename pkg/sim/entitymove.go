package sim

import (
	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// InvalidateEntity is called with an entity's sprite bounds (viewport
// coordinates) where upstream calls EntityBase::invalidateSprite →
// ViewportManager::invalidate(entity, ZoomLevel::eighth): the world maps them
// onto its screen's invalidation grid. nil: nothing to invalidate.
type InvalidateEntity func(left, top, right, bottom int16)

// moveTo is EntityBase::moveTo: the old sprite area invalidated (when the
// entity was placed), the position set, and the sprite bounds recomputed from
// gameToScreen (rotation 0 — PARTIAL: the view's rotation, LOOP item R) and
// the entity's sprite extents, then the new area invalidated. Location::null
// in x leaves it unplaced (spriteLeft null). (moveSpatialEntry — the entity
// manager's tile lists — PARTIAL: goloco's renderer indexes by position.)
func (m *Map) moveTo(e *scenario.VehicleEntity, x, y, z int32) {
	if e.PosX != locationNull {
		m.invalidateSprite(e)
	}
	e.PosX, e.PosY, e.PosZ = coords.WorldUnits(x), coords.WorldUnits(y), int16(z)
	if x == locationNull {
		e.SpriteLeft = locationNull
		return
	}
	vx, vy := y-x, ((y+x)>>1)-z // World::gameToScreen at rotation 0
	e.SpriteLeft, e.SpriteRight = int16(vx-int32(e.SpriteWidth)), int16(vx+int32(e.SpriteWidth))
	e.SpriteTop, e.SpriteBottom = int16(vy-int32(e.SpriteHeightNegative)), int16(vy+int32(e.SpriteHeightPositive))
	m.invalidateSprite(e)
}

// invalidateSprite is EntityBase::invalidateSprite.
func (m *Map) invalidateSprite(e *scenario.VehicleEntity) {
	if m.InvalidateEntity == nil || e.SpriteLeft == locationNull {
		return
	}
	m.InvalidateEntity(e.SpriteLeft, e.SpriteTop, e.SpriteRight, e.SpriteBottom)
}

// Zoom levels (ZoomLevel): ViewportManager::invalidate skips viewports zoomed
// out further than the level it is given.
const (
	ZoomFull uint8 = iota
	ZoomHalf
	ZoomQuarter
	ZoomEighth
)

// InvalidateTile is called where upstream calls ViewportManager::invalidate(
// pos, zMin, zMax, zoom, radius): the world turns the tile's column between
// zMin and zMax (pos + 16, gameToScreen, grown by radius) into a screen rect.
type InvalidateTile func(x, y, zMin, zMax int32, zoom uint8, radius int32)

// defaultRadius is ViewportManager::invalidate's default radius (ViewportManager.h:
// int radius = 32) — what every call that names none gets.
const defaultRadius = 32

// invalidateTile is ViewportManager::invalidate(pos, zMin, zMax, zoom, radius).
func (m *Map) invalidateTile(x, y, zMin, zMax int32, zoom uint8, radius int32) {
	if m.InvalidateTile != nil {
		m.InvalidateTile(x, y, zMin, zMax, zoom, radius)
	}
}

// mapInvalidateTileFull is TileManager::mapInvalidateTileFull: the tile from
// height 0 to 1120 at ZoomLevel::eighth.
func (m *Map) mapInvalidateTileFull(x, y int32) {
	m.invalidateTile(x, y, 0, 1120, ZoomEighth, defaultRadius)
}
