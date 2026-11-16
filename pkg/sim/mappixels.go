package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The map window's overall tab, per tile (MapWindow.cpp setMapPixelsOverall,
// 0x0046C5E5; PORT-PLANS A step 1): the two palette colours its rendered
// pixel pair takes, and the pair while flashing (the graph key's hovered
// row). The window (not yet ported) lays the pairs out rotated 45° and fills
// them a row a tick.

// Palette indices the map uses (Colour.h PaletteIndex).
const (
	palBlack0        = 0x0A
	palBlack2        = 0x0C
	palBlack7        = 0x11
	palBlackB        = 0x15
	palMutedDarkRed7 = 0x41
	palGreen6        = 0x64
	palMutedPurple7  = 0x7D
	palOrange8       = 0xBA
)

// MapFlash bits (_flashingItems): which of the overall tab's key rows flash.
const (
	MapFlashBuildings = 1 << iota
	MapFlashIndustries
	MapFlashRoads
	MapFlashRailways
	MapFlashStations
	MapFlashTrees
)

// mapFlashColour is kFlashColours (0x004FDC5C): black0, or blackB for the
// four black0–3 indices.
func mapFlashColour(c uint8) uint8 {
	if c >= 10 && c < 14 {
		return palBlackB
	}
	return palBlack0
}

// Track and road object flags the map reads.
const (
	trackFlagIsRoad = 1 << 2 // TrackObjectFlags::isRoad
	roadFlagIsRail  = 1 << 1 // RoadObjectFlags::isRail
)

// MapPixelsOverall is setMapPixelsOverall's body for the tile at tile
// coordinates (tx, ty): the surface's colours by its height from the land
// (or water) object's map pixel image, then each element over it — track and
// road (rail black7, road black2), stations orange, buildings muted red,
// industries muted purple, trees green in the second pixel; ghosts and AI
// placements left out, walls and signals skipped. pixel reads a G1
// element's bytes (getG1Element(image)->offset[i]). ok is false on the
// map's edge, which upstream leaves untouched.
func (m *Map) MapPixelsOverall(tx, ty int, om *objects.ObjectManager, pixel func(image uint32, i int) uint8, flashing uint32) (c0, c1, f0, f1 uint8, ok bool) {
	if tx <= 0 || ty <= 0 || tx >= m.Width-1 || ty >= m.Height-1 || om == nil {
		return 0, 0, 0, 0, false
	}
	t := m.TileAt(tx, ty)
	if t == nil {
		return 0, 0, 0, 0, false
	}
	both := func(c uint8, bit uint32) { // a colour over both pixels, flashing its key row
		c0, c1, f0, f1 = c, c, c, c
		if flashing&bit != 0 {
			f0 = mapFlashColour(c)
			f1 = f0
		}
	}
	for _, r := range t.Order {
		ghost, ai := r.Flags&elemFlagGhost != 0, r.Flags&elemFlagAiAllocated != 0
		switch r.Kind {
		case scenario.KindSurface:
			if t.WaterLevel == 0 {
				if lo := om.GetLandObjectByIndex(int(t.TerrainIndex)); lo != nil {
					off := int(t.BaseZ) / 4 * 2
					f0, f1 = pixel(lo.MapPixelImage, off), pixel(lo.MapPixelImage, off+1)
				}
			} else if om.WaterObj != nil {
				off := (int(t.WaterLevel)*4 - int(t.BaseZ)) / 2
				img := om.WaterObj.ImageOffset + 40 // WaterObject::mapPixelImage
				f0, f1 = pixel(img, off-2), pixel(img, off-1)
			}
			c0, c1 = f0, f1
		case scenario.KindTrack:
			if ghost || ai || int(r.Index) >= len(t.Tracks) {
				continue
			}
			if to := om.GetTrackObjectByIndex(int(t.Tracks[r.Index].TrackObjectID)); to != nil && to.Flags&trackFlagIsRoad != 0 {
				both(palBlack2, MapFlashRoads)
			} else {
				both(palBlack7, MapFlashRailways)
			}
		case scenario.KindRoad:
			if ghost || ai || int(r.Index) >= len(t.Roads) {
				continue
			}
			if ro := om.GetRoadObjectByIndex(int(t.Roads[r.Index].RoadObjectID)); ro != nil && ro.Flags&roadFlagIsRail != 0 {
				both(palBlack7, MapFlashRailways)
			} else {
				both(palBlack2, MapFlashRoads)
			}
		case scenario.KindStation:
			if !ghost && !ai {
				both(palOrange8, MapFlashStations)
			}
		case scenario.KindBuilding:
			if !ghost {
				both(palMutedDarkRed7, MapFlashBuildings)
			}
		case scenario.KindIndustry:
			if !ghost {
				both(palMutedPurple7, MapFlashIndustries)
			}
		case scenario.KindTree:
			if !ghost {
				c1, f1 = palGreen6, palGreen6
				if flashing&MapFlashTrees != 0 {
					f1 = palBlack0
				}
			}
		}
	}
	return c0, c1, f0, f1, true
}

// The map window's buffers (MapWindow.cpp): kMapColumns × 2 square, a pair of
// pixels per tile along diagonal rows — the map turned 45°.
const (
	mapColumns       = 384 // kMapColumns
	MapRenderedWidth = mapColumns * 2
	mapRenderedSize  = MapRenderedWidth * MapRenderedWidth
)

// MapBuffer is _mapPixels and _mapAltPixels (the flashing frame), and
// _drawMapRowIndex, the row the next update fills.
type MapBuffer struct {
	Pixels, Alt []uint8
	Row         int
}

// NewMapBuffer is the buffers as clearMap (0x0046B69C) leaves them: black0.
func NewMapBuffer() *MapBuffer {
	b := &MapBuffer{Pixels: make([]uint8, mapRenderedSize), Alt: make([]uint8, mapRenderedSize)}
	for i := range b.Pixels {
		b.Pixels[i], b.Alt[i] = palBlack0, palBlack0
	}
	return b
}

// FillMapRowOverall is setMapPixels (the overall tab) for the buffer's next
// row: from the rotation's corner a column of tiles (the tile loop's first
// column for rotation 0), each tile's pair a row down and a pixel right of
// the last; then on to the next row, round to 0 after kMapColumns.
func (m *Map) FillMapRowOverall(b *MapBuffer, rotation int, om *objects.ObjectManager, pixel func(image uint32, i int) uint8, flashing uint32) {
	r := b.Row
	off := r*(MapRenderedWidth-1) + (mapColumns - 1)
	var x, y, dx, dy int
	switch rotation & 3 {
	case 0:
		x, y, dx, dy = r, 0, 0, 1
	case 1:
		x, y, dx, dy = mapColumns-1, r, -1, 0
	case 2:
		x, y, dx, dy = mapColumns-1-r, mapColumns-1, 0, -1
	case 3:
		x, y, dx, dy = 0, mapColumns-1-r, 1, 0
	}
	for range mapColumns {
		if c0, c1, f0, f1, ok := m.MapPixelsOverall(x, y, om, pixel, flashing); ok && off >= 0 && off+1 < mapRenderedSize {
			b.Pixels[off], b.Pixels[off+1] = c0, c1
			b.Alt[off], b.Alt[off+1] = f0, f1
		}
		x, y = x+dx, y+dy
		off += MapRenderedWidth + 1
	}
	b.Row++
	if b.Row > mapColumns {
		b.Row = 0
	}
}
