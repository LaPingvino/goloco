package world

import (
	"image/color"
	"math"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/hajimehoshi/ebiten/v2"
)

// slopeToDisplaySlope maps raw slope values (0-31) to display slope indices (0-18)
// From OpenLoco PaintSurface.cpp kSlopeToDisplaySlope
// Slopes that can't be displayed are mapped to 0 (flat)
var slopeToDisplaySlope = [32]uint8{
	0, 2, 1, 3, 8, 10, 9, 11,
	4, 6, 5, 7, 12, 14, 13, 15,
	0, 0, 0, 0, 0, 0, 0, 17,
	0, 0, 0, 16, 0, 18, 15, 0,
}

// TileType represents different terrain types
type TileType int

const (
	TileGrass TileType = iota
	TileDirt
	TileWater
)

// tile stores per-tile render state
type tile struct {
	tileType     TileType
	terrainIndex uint8 // raw LandObject slot index from the scenario
	baseZ        uint8 // height in SmallZ units (OpenLoco: 1 unit = 4px)
	slope        uint8 // slope corners and flags from byte 4 of surface element
	waterLevel   uint8 // water level from surface element byte 5 bits [4:0]; 0 = no water
	growthStage  uint8 // terrain growth/season stage 0-7 (byte 6 bits [7:5]); selects sprite set
	trees        []scenario.TreeElement
	buildings    []scenario.BuildingElement
	walls        []scenario.WallElement
	tracks       []scenario.TrackElement
	roads        []scenario.RoadElement
	stations     []scenario.StationElement
	signals      []scenario.SignalElement
	industries   []scenario.IndustryElement
}

// DemoTrain is a simple animated train entity for the title sequence and early gameplay.
// It follows the tile grid, moving one tile at a time in direction (dx,dy).
type DemoTrain struct {
	tileX, tileY int     // current tile (leading car position)
	dx, dy       int     // movement direction in tile space (one axis only)
	progress     float64 // 0.0 = at tile centre, 1.0 = at next tile
	vehicleIdx   int     // index into renderer.ObjMgr.Vehicles
}

// World holds a tile grid for an isometric game world
type World struct {
	renderer *render.Renderer
	width    int
	height   int
	tiles    [][]tile
	trains   []*DemoTrain            // demo trains (title screen / scenarios without saved entities)
	entities []scenario.VehicleEntity // real vehicle entities from .SV5 saved-game files
	// isometric tile dimensions (standard 2:1 ratio)
	tileW int // tile width in pixels (64)
	tileH int // tile height in pixels (32)
	// cache colored diamond images per tile type (fallback)
	tileCache  map[TileType]*ebiten.Image
	waterImage *ebiten.Image // cached translucent water diamond overlay (fallback)
	// Cache for pre-rendered solid-blue water blend sprites, keyed by shape (0-4).
	// Each entry is the +35 water sprite forced to the game's actual water colour
	// (derived from the water palette map at WaterObject.ImageOffset+41).
	// This approximates OpenLoco's dst-based water blend without requiring screen readback.
	waterBlendCache map[int]*ebiten.Image
	// camera offset in pixels (in world-space before zoom is applied)
	camX float64
	camY float64
	// when true, Draw() does not recentre the camera —
	// an external caller (e.g. title sequence) is driving it via SetCamera.
	externalCamera bool
	// zoom level 0–3 matching OpenLoco convention:
	//   0 = full size (1×), 1 = half (2×), 2 = quarter (4×), 3 = eighth (8×)
	// The pixel scale factor is 1 << zoom applied in reverse:
	// world coordinates are divided by (1 << zoom) for screen placement.
	zoom int
}

func NewWorld(r *render.Renderer) *World {
	w := &World{
		renderer:  r,
		width:     20,
		height:    15,
		tileW:     64, // standard isometric tile width
		tileH:     32, // standard isometric tile height (64×32 = 2:1 ratio)
		tiles:     make([][]tile, 15),
		tileCache: make(map[TileType]*ebiten.Image),
		zoom:      0, // start at full zoom (1× scale)
	}

	// Initialize tiles with some variety
	for y := 0; y < w.height; y++ {
		w.tiles[y] = make([]tile, w.width)
		for x := 0; x < w.width; x++ {
			tt := TileGrass
			if x < 2 || y < 2 || x >= w.width-2 || y >= w.height-2 {
				tt = TileWater
			} else if (x+y)%7 == 0 {
				tt = TileDirt
			}
			w.tiles[y][x] = tile{tileType: tt}
		}
	}

	// Center camera on the map
	cx, cy := w.tileToScreen(w.width/2, w.height/2)
	w.camX = cx - 400 // half screen width
	w.camY = cy - 300 // half screen height
	return w
}

// tileToScreen converts tile coordinates to viewport pixel coordinates.
// Matches OpenLoco Map/Tile.cpp::gameToScreen (rotation 0):
//
//	vpX = (tileY - tileX) * kTileSize        where kTileSize = 32
//	vpY = (tileY + tileX) * (kTileSize / 2)  = (tileY + tileX) * 16
//
// Height is NOT included here; it is subtracted separately in Draw().
// Camera offset and zoom are also applied in Draw().
func (w *World) tileToScreen(tileX, tileY int) (float64, float64) {
	vpX := float64((tileY - tileX) * 32)
	vpY := float64((tileY + tileX) * 16)
	return vpX, vpY
}

func (w *World) LoadFromScenario(sc *scenario.Scenario) {
	if sc == nil || sc.Tiles == nil {
		return
	}
	w.width = sc.MapWidth
	w.height = sc.MapHeight
	w.tiles = make([][]tile, w.height)
	for y := 0; y < w.height; y++ {
		w.tiles[y] = make([]tile, w.width)
		for x := 0; x < w.width; x++ {
			st := sc.GetTile(x, y)
			if st == nil {
				w.tiles[y][x] = tile{tileType: TileGrass}
				continue
			}

			tt := TileGrass
			switch st.Surface {
			case scenario.SurfaceWater:
				tt = TileWater
			case scenario.SurfaceDirt, scenario.SurfaceSand, scenario.SurfaceRock:
				tt = TileDirt
			}
			w.tiles[y][x] = tile{
				tileType:     tt,
				terrainIndex: st.TerrainIndex,
				baseZ:        st.Height,
				slope:        st.Slope,
				waterLevel:   st.Water,
				growthStage:  st.GrowthStage,
				trees:        st.Trees,
				buildings:    st.Buildings,
				walls:        st.Walls,
				tracks:       st.Tracks,
				roads:        st.Roads,
				stations:     st.Stations,
				signals:      st.Signals,
				industries:   st.Industries,
			}
		}
	}

	// Center camera on the map
	cx, cy := w.tileToScreen(w.width/2, w.height/2)
	w.camX = cx - 400
	w.camY = cy - 300

	// Store real vehicle entities from saved games.  When entities are present
	// they are rendered by paintVehicleEntities; otherwise fall back to demo trains.
	w.entities = sc.Entities

	if len(w.entities) == 0 {
		// No real entities (scenario file or empty save) — spawn scripted demo trains
		// so the title screen and freshly-loaded scenarios show some movement.
		w.spawnDemoTrains()
	} else {
		w.trains = nil
	}
}

// spawnDemoTrains places a handful of trains on track tiles found in the map.
func (w *World) spawnDemoTrains() {
	w.trains = nil
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(w.renderer.ObjMgr.Vehicles) == 0 {
		return
	}

	// Find tiles with at least one track element. Spread spawns evenly.
	type trackPos struct{ x, y int }
	var trackTiles []trackPos
	step := 1
	if w.width*w.height > 1000 {
		step = 8 // sample every 8th tile on large maps
	}
	for y := 0; y < w.height; y += step {
		for x := 0; x < w.width; x += step {
			if len(w.tiles[y][x].tracks) > 0 {
				trackTiles = append(trackTiles, trackPos{x, y})
			}
		}
	}
	if len(trackTiles) == 0 {
		return
	}

	// Spawn up to 5 trains, evenly spaced in the found list.
	maxTrains := 5
	if len(trackTiles) < maxTrains {
		maxTrains = len(trackTiles)
	}
	for i := 0; i < maxTrains; i++ {
		idx := i * len(trackTiles) / maxTrains
		tp := trackTiles[idx]

		// Prefer a direction that has a neighbouring track tile; fall back to +Y.
		dx, dy := 0, 1
		for _, cand := range [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}} {
			nx, ny := tp.x+cand[0], tp.y+cand[1]
			if nx >= 0 && nx < w.width && ny >= 0 && ny < w.height &&
				len(w.tiles[ny][nx].tracks) > 0 {
				dx, dy = cand[0], cand[1]
				break
			}
		}

		vIdx := i % len(w.renderer.ObjMgr.Vehicles)
		w.trains = append(w.trains, &DemoTrain{
			tileX:      tp.x,
			tileY:      tp.y,
			dx:         dx,
			dy:         dy,
			progress:   float64(i) / float64(maxTrains), // stagger start positions
			vehicleIdx: vIdx,
		})
	}
}

// SetCamera overrides the camera position directly (used by the title
// sequence to pan across the map). Uses the same projection as tileToScreen.
func (w *World) SetCamera(tileX, tileY float64) {
	w.externalCamera = true
	// Mirror of tileToScreen: vpX = (tileY-tileX)*32, vpY = (tileY+tileX)*16
	worldX := (tileY - tileX) * 32.0
	worldY := (tileY + tileX) * 16.0
	halfW := 400.0 * float64(int(1)<<w.zoom)
	halfH := 300.0 * float64(int(1)<<w.zoom)
	w.camX = worldX - halfW
	w.camY = worldY - halfH
}

// ZoomIn decreases the zoom level (more detail, fewer tiles visible).
// Zoom 0 is full size; cannot go below 0.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Window.cpp
//
//	Window::viewportZoomIn(bool toCursor)
func (w *World) ZoomIn() {
	if w.zoom > 0 {
		w.zoom--
	}
}

// ZoomOut increases the zoom level (less detail, more tiles visible).
// Zoom 3 is the minimum size (1/8); cannot go above 3.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Window.cpp
//
//	Window::viewportZoomOut(bool toCursor)
func (w *World) ZoomOut() {
	if w.zoom < 3 {
		w.zoom++
	}
}

// SetZoom sets the zoom level directly (0=full, 1=half, 2=quarter, 3=eighth)
func (w *World) SetZoom(level int) {
	if level >= 0 && level <= 3 {
		w.zoom = level
	}
}

// GetMapSize returns the map dimensions in tiles
func (w *World) GetMapSize() (width, height int) {
	return w.width, w.height
}

// GetZoom returns the current zoom level (0=full, 1=half, 2=quarter, 3=eighth).
// The screen scale factor is 1/(1<<zoom).
func (w *World) GetZoom() int {
	return w.zoom
}

func (w *World) Update() {
	// Advance demo trains at ~2 tiles/second (60fps → 1/30 per frame).
	const speed = 1.0 / 30.0
	for _, tr := range w.trains {
		tr.progress += speed
		if tr.progress >= 1.0 {
			tr.progress -= 1.0
			// Move to the next tile
			nx := tr.tileX + tr.dx
			ny := tr.tileY + tr.dy
			// Bounce if out-of-bounds or no track at destination
			if nx < 0 || nx >= w.width || ny < 0 || ny >= w.height ||
				len(w.tiles[ny][nx].tracks) == 0 {
				tr.dx = -tr.dx
				tr.dy = -tr.dy
			} else {
				tr.tileX = nx
				tr.tileY = ny
			}
		}
	}
}

// PanCamera moves the camera by (dx, dy) pixels in world-space.
// Called each frame when WASD/arrow keys are held.
func (w *World) PanCamera(dx, dy float64) {
	w.externalCamera = false // user is now driving the camera
	w.camX += dx
	w.camY += dy
}

// getFallbackImage returns a cached colored diamond for a tile type.
// Used when no LandObject sprite is available.
func (w *World) getFallbackImage(tt TileType) *ebiten.Image {
	if img, ok := w.tileCache[tt]; ok {
		return img
	}
	img := w.createDiamondTile(tt)
	w.tileCache[tt] = img
	return img
}

// createDiamondTile creates a simple diamond-shaped tile image
func (w *World) createDiamondTile(tt TileType) *ebiten.Image {
	img := ebiten.NewImage(w.tileW, w.tileH)

	var mainColor, darkColor, lightColor color.RGBA
	switch tt {
	case TileGrass:
		mainColor = color.RGBA{80, 160, 80, 255}
		darkColor = color.RGBA{60, 140, 60, 255}
		lightColor = color.RGBA{100, 180, 100, 255}
	case TileDirt:
		mainColor = color.RGBA{139, 90, 43, 255}
		darkColor = color.RGBA{100, 65, 30, 255}
		lightColor = color.RGBA{170, 120, 70, 255}
	case TileWater:
		mainColor = color.RGBA{64, 120, 192, 255}
		darkColor = color.RGBA{40, 90, 160, 255}
		lightColor = color.RGBA{100, 160, 220, 255}
	}

	centerX := w.tileW / 2
	centerY := w.tileH / 2

	for y := 0; y < w.tileH; y++ {
		var halfWidth int
		if y < centerY {
			halfWidth = (y * centerX) / centerY
		} else {
			halfWidth = ((w.tileH - 1 - y) * centerX) / centerY
		}

		for x := centerX - halfWidth; x <= centerX+halfWidth; x++ {
			if x < 0 || x >= w.tileW {
				continue
			}

			var c color.RGBA
			if y < centerY {
				if x < centerX {
					c = lightColor
				} else {
					c = mainColor
				}
			} else {
				if x < centerX {
					c = mainColor
				} else {
					c = darkColor
				}
			}

			img.Set(x, y, c)
		}
	}

	return img
}

// getTile returns a pointer to the tile at (x, y), or nil if out of bounds
func (w *World) getTile(x, y int) *tile {
	if x < 0 || x >= w.width || y < 0 || y >= w.height {
		return nil
	}
	return &w.tiles[y][x]
}

// Edge directions for neighbor lookup
const (
	EdgeSW = 0 // South-West (bottom-left)
	EdgeSE = 1 // South-East (bottom-right)
	EdgeNW = 2 // North-West (top-left)
	EdgeNE = 3 // North-East (top-right)
)

// getNeighborTile returns the neighbor tile in the given edge direction.
// Offsets match OpenLoco PaintSurface.cpp::kNeighbourOffsets[rotation=0]:
//   SW: Pos2{+32,0}  → tileX+1
//   SE: Pos2{0,+32}  → tileY+1
//   NW: Pos2{0,-32}  → tileY-1
//   NE: Pos2{-32,0}  → tileX-1
func (w *World) getNeighborTile(x, y int, edge int) *tile {
	var dx, dy int
	switch edge {
	case EdgeSW: // world +x
		dx, dy = 1, 0
	case EdgeSE: // world +y
		dx, dy = 0, 1
	case EdgeNW: // world -y
		dx, dy = 0, -1
	case EdgeNE: // world -x
		dx, dy = -1, 0
	default:
		return nil
	}
	return w.getTile(x+dx, y+dy)
}

// CornerHeight represents the height of the 4 corners of a tile
// From OpenLoco CornerHeight struct
type CornerHeight struct {
	Top    uint8 // North corner
	Right  uint8 // East corner
	Bottom uint8 // South corner
	Left   uint8 // West corner
}

// cornerHeights lookup table from OpenLoco kCornerHeights
// Maps slope value (0-31) to corner heights (0, 1, or 2)
var cornerHeights = [32]CornerHeight{
	{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}, {0, 0, 1, 1},
	{1, 0, 0, 0}, {1, 0, 1, 0}, {1, 0, 0, 1}, {1, 0, 1, 1},
	{0, 1, 0, 0}, {0, 1, 1, 0}, {0, 1, 0, 1}, {0, 1, 1, 1},
	{1, 1, 0, 0}, {1, 1, 1, 0}, {1, 1, 0, 1}, {1, 1, 1, 1},
	{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 0, 1}, {0, 0, 1, 1},
	{1, 0, 0, 0}, {1, 0, 1, 0}, {1, 0, 0, 1}, {1, 0, 1, 2},
	{0, 1, 0, 0}, {0, 1, 1, 0}, {0, 1, 0, 1}, {0, 1, 2, 1},
	{1, 1, 0, 0}, {1, 2, 1, 0}, {2, 1, 0, 1}, {1, 1, 1, 1},
}

// getCornerHeights returns the absolute corner heights for a tile in MicroZ units.
// OpenLoco PaintSurface.cpp: microZ = baseZ / kMicroToSmallZStep (= baseZ / 4).
// 1 MicroZ = 16 viewport pixels at zoom-0.
func (w *World) getCornerHeights(t *tile) CornerHeight {
	if t == nil {
		return CornerHeight{0, 0, 0, 0}
	}
	// SmallZ → MicroZ: divide by kMicroToSmallZStep = 4
	microZ := uint8(t.baseZ) / 4
	slope := t.slope & 0x1F
	rel := cornerHeights[slope]
	return CornerHeight{
		Top:    microZ + rel.Top,
		Right:  microZ + rel.Right,
		Bottom: microZ + rel.Bottom,
		Left:   microZ + rel.Left,
	}
}

// EdgeHeight represents the heights of the two corners on an edge
type EdgeHeight struct {
	Self0      uint8 // First corner on self side
	Neighbor0  uint8 // First corner on neighbor side
	Self1      uint8 // Second corner on self side
	Neighbor1  uint8 // Second corner on neighbor side
}

// getEdgeHeights calculates the corner heights for a specific edge
func (w *World) getEdgeHeights(x, y int, edge int, selfCorners CornerHeight) EdgeHeight {
	neighbor := w.getNeighborTile(x, y, edge)
	if neighbor == nil {
		return EdgeHeight{0, 0, 0, 0}
	}

	neighborCorners := w.getCornerHeights(neighbor)

	// Map edge direction to corner pairs
	// From OpenLoco paintSurface edge height calculation
	var eh EdgeHeight
	switch edge {
	case EdgeSW: // 0
		eh.Self0 = selfCorners.Left
		eh.Neighbor0 = neighborCorners.Top
		eh.Self1 = selfCorners.Bottom
		eh.Neighbor1 = neighborCorners.Right
	case EdgeSE: // 1
		eh.Self0 = selfCorners.Right
		eh.Neighbor0 = neighborCorners.Top
		eh.Self1 = selfCorners.Bottom
		eh.Neighbor1 = neighborCorners.Left
	case EdgeNW: // 2
		eh.Self0 = selfCorners.Top
		eh.Neighbor0 = neighborCorners.Right
		eh.Self1 = selfCorners.Left
		eh.Neighbor1 = neighborCorners.Bottom
	case EdgeNE: // 3
		eh.Self0 = selfCorners.Top
		eh.Neighbor0 = neighborCorners.Left
		eh.Self1 = selfCorners.Right
		eh.Neighbor1 = neighborCorners.Bottom
	}

	return eh
}

// Edge factor offsets for cliff sprite variation per edge direction.
// OpenLoco reference: PaintSurface.cpp kEdgeFactorOffset
var kEdgeFactorOffset = [4]uint32{0, 16, 16, 0} // SW, SE, NW, NE

// paintCliffEdges renders cliff edge sprites for height transitions
// OpenLoco reference: PaintSurface.cpp paintSurfaceCliffEdge(), paintEdgeSection()
func (w *World) paintCliffEdges(screen *ebiten.Image, x, y int, t *tile, land *objects.LandObject, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.G1 == nil || land == nil {
		return
	}

	// drawY was computed with the tile height already subtracted (heightOffset = baseZ*4).
	// getCliffEdgeOffset returns offsets relative to the tile's FLAT (zero-height) viewport
	// position, so we need drawY without the height contribution.
	// drawYFlat = drawY + baseZ*4*scale restores the flat reference.
	drawYFlat := drawY + float64(t.baseZ)*4.0*scale

	// Get corner heights for this tile
	selfCorners := w.getCornerHeights(t)

	// Check all 4 edges
	edges := []int{EdgeSW, EdgeSE, EdgeNW, EdgeNE}

	for _, edge := range edges {
		neighbor := w.getNeighborTile(x, y, edge)
		if neighbor == nil {
			continue
		}

		edgeHeights := w.getEdgeHeights(x, y, edge, selfCorners)

		if edgeHeights.Self0 <= edgeHeights.Neighbor0 && edgeHeights.Self1 <= edgeHeights.Neighbor1 {
			continue
		}

		cliffEdgeImageBase := land.CliffEdgeImage

		minHeight := edgeHeights.Neighbor0
		if edgeHeights.Neighbor1 < minHeight {
			minHeight = edgeHeights.Neighbor1
		}
		maxHeight := edgeHeights.Self0
		if edgeHeights.Self1 > maxHeight {
			maxHeight = edgeHeights.Self1
		}

		// Checkerboard: OpenLoco uses world coords (tileX*32)^(tileY*32) & 0x20.
		// Equivalent with tile coords: ((tileX^tileY) & 1) * 32 — alternates every tile.
		checkerboard := uint32((x^y)&1) * 32
		factor := checkerboard + kEdgeFactorOffset[edge]

		for h := minHeight; h < maxHeight; h++ {
			spriteID := int(cliffEdgeImageBase) + int(factor) + int(h&0xF)

			if img := w.renderer.GetSprite(spriteID); img != nil {
				_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
				if ok {
					edgeOffsetX, edgeOffsetY := w.getCliffEdgeOffset(edge, h)

					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(scale, scale)
					op.GeoM.Translate(
						drawX+float64(xOff+edgeOffsetX)*scale,
						drawYFlat+float64(yOff+edgeOffsetY)*scale)
					screen.DrawImage(img, op)
				}
			}
		}
	}
}

// Tree quadrant offsets within a tile (in sub-tile coordinates).
// OpenLoco reference: PaintTree.cpp kTreeQuadrantOffset
var kTreeQuadrantOffset = [4][2]int16{
	{7, 7},   // quadrant 0
	{7, 23},  // quadrant 1
	{23, 23}, // quadrant 2
	{23, 7},  // quadrant 3
}

// paintWaterCliffEdges renders cliff edge sprites at the water-surface level
// for tiles where the adjacent tile has a different (lower) water height.
// This creates the visual "shore cliff" between water bodies at different levels.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintSurface.cpp
//   paintSurfaceWaterCliffEdge() — called for all 4 edges after paintWater
func (w *World) paintWaterCliffEdges(screen *ebiten.Image, x, y int, t *tile, land *objects.LandObject, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.G1 == nil || land == nil || land.CliffEdgeImage == 0 {
		return
	}
	if t.waterLevel == 0 {
		return
	}

	// drawYFlat is the Y position of this tile at terrain base (height zero reference).
	drawYFlat := drawY + float64(t.baseZ)*4.0*scale

	checkerboard := uint32((x^y)&1) * 32
	cliffEdgeImageBase := land.CliffEdgeImage
	waterMicroZ := t.waterLevel // waterLevel is stored in MicroZ (1 MicroZ = 16 px)

	edges := []int{EdgeSW, EdgeSE, EdgeNW, EdgeNE}
	for _, edge := range edges {
		neighbor := w.getNeighborTile(x, y, edge)
		if neighbor == nil || neighbor.waterLevel == t.waterLevel {
			continue
		}

		// Neighbor terrain corner heights in MicroZ.
		nc := w.getCornerHeights(neighbor)
		var neigh0, neigh1 uint8
		switch edge {
		case EdgeSW:
			neigh0, neigh1 = nc.Top, nc.Right
		case EdgeSE:
			neigh0, neigh1 = nc.Top, nc.Left
		case EdgeNW:
			neigh0, neigh1 = nc.Right, nc.Bottom
		case EdgeNE:
			neigh0, neigh1 = nc.Left, nc.Bottom
		}

		// Only draw where our water is above the neighbor terrain.
		if waterMicroZ <= neigh0 && waterMicroZ <= neigh1 {
			continue
		}
		minH := neigh0
		if neigh1 < minH {
			minH = neigh1
		}

		factor := checkerboard + kEdgeFactorOffset[edge]
		for h := minH; h < waterMicroZ; h++ {
			spriteID := int(cliffEdgeImageBase) + int(factor) + int(h&0xF)
			img := w.renderer.GetSprite(spriteID)
			if img == nil {
				continue
			}
			_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
			if !ok {
				continue
			}
			edgeOffsetX, edgeOffsetY := w.getCliffEdgeOffset(edge, h)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(
				drawX+float64(xOff+edgeOffsetX)*scale,
				drawYFlat+float64(yOff+edgeOffsetY)*scale)
			screen.DrawImage(img, op)
		}
	}
}

// paintTrees renders tree sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintTree.cpp  paintTree()
func (w *World) paintTrees(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.trees) == 0 {
		return
	}

	for _, te := range t.trees {
		treeObj := w.renderer.ObjMgr.GetTreeObjectByIndex(int(te.TreeObjectID))
		if treeObj == nil {
			continue
		}

		// Calculate frame number: rotation + growth * numRotations
		// OpenLoco: (viewableRotation % numRotations) + growth * numRotations
		// We use rotation=0 for now (no viewport rotation support yet)
		viewRotation := uint32(te.Rotation) % uint32(treeObj.NumRotations)
		treeFrameNum := viewRotation + uint32(te.Growth)*uint32(treeObj.NumRotations)

		// Select season sprite base
		season := te.Season
		if season >= 6 {
			season = 0
		}
		var spriteBase uint32
		if te.HasSnow {
			spriteBase = treeObj.SnowSprites[season]
		} else {
			spriteBase = treeObj.Sprites[season]
		}

		spriteID := int(spriteBase + treeFrameNum)

		img := w.renderer.GetSprite(spriteID)
		if img == nil {
			continue
		}

		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			continue
		}

		// Position within tile based on quadrant
		quadrant := te.Quadrant % 4
		qx := float64(kTreeQuadrantOffset[quadrant][0]) - 16 // center around tile
		qy := float64(kTreeQuadrantOffset[quadrant][1]) - 16

		// Height offset: drawY already includes the surface baseZ, so we
		// only need the extra height if the tree sits above the surface.
		extraHeight := float64(int(te.BaseZ)-int(t.baseZ)) * 4.0

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(
			drawX+(float64(xOff)+qx)*scale,
			drawY+(float64(yOff)-extraHeight+qy)*scale)
		screen.DrawImage(img, op)
	}
}

// paintBuildings renders building sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintBuilding.cpp  paintBuilding()
// Simplified: renders the first variation's parts stacked vertically.
func (w *World) paintBuildings(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.buildings) == 0 {
		return
	}

	for _, be := range t.buildings {
		bldgObj := w.renderer.ObjMgr.GetBuildingObjectByIndex(int(be.ObjectID))
		if bldgObj == nil || bldgObj.ImageOffset == 0 {
			continue
		}

		// Only render constructed buildings (skip under-construction for now)
		if !be.IsConstructed {
			continue
		}

		rotation := be.Rotation & 0x03
		variation := be.Variation
		if int(variation) >= int(bldgObj.NumVariations) {
			variation = 0
		}

		// For multi-tile (2x2) buildings only the "front" tile renders the sprites.
		// OpenLoco: (sequenceIndex ^ 2) == ((-viewportRotation) & 3).
		// With fixed viewport rotation=0: render seqIdx==2, skip all others.
		if bldgObj.HasFlags(objects.BuildingFlagLargeTile) && be.SequenceIndex != 2 {
			continue
		}

		parts := bldgObj.GetBuildingParts(variation)
		if len(parts) == 0 {
			continue
		}

		// Height offset for building's own baseZ vs surface baseZ
		extraHeight := float64(int(be.BaseZ)-int(t.baseZ)) * 4.0

		// Sub-tile image offset: OpenLoco kImageOffsetBase1x1={16,16,0}, kImageOffsetBase2x2={0,0,0}.
		// Projected to screen: screenX=wy-wx, screenY=(wx+wy)/2.
		// 1x1: {16,16,0} → screenDelta=(0,16); 2x2: {0,0,0} → screenDelta=(0,0).
		offsetX := float64(0)
		offsetY := float64(16)
		if bldgObj.HasFlags(objects.BuildingFlagLargeTile) {
			offsetX = 0
			offsetY = 0
		}

		// Stack parts vertically
		partZ := float64(0)
		for _, part := range parts {
			// Each part has 4 rotation variants: part * 4 + rotation
			spriteID := int(bldgObj.ImageOffset) + int(part)*4 + int(rotation)

			img := w.renderer.GetSprite(spriteID)
			if img == nil {
				continue
			}

			_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
			if !ok {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(
				drawX+(float64(xOff)+offsetX)*scale,
				drawY+(float64(yOff)-extraHeight-partZ+offsetY)*scale)
			screen.DrawImage(img, op)

			// Stack next part on top
			if int(part) < len(bldgObj.PartHeights) {
				partZ += float64(bldgObj.PartHeights[part])
			}
		}
	}
}

// kWallScreenOffsets: screen-pixel (deltaX, deltaY) per rotation.
// OpenLoco source: PaintWall.cpp kOffsets stores world sub-tile Pos3 values:
//
//	rot0={0,0,0}, rot1={1,31,0}, rot2={31,0,0}, rot3={2,1,0}
//
// Projection to screen pixels (matching tileToScreen formula):
//
//	screenX = wy - wx
//	screenY = (wy + wx) / 2
//
// rot0: (0,0)   → (0,   0 )
// rot1: (1,31)  → (30,  16)
// rot2: (31,0)  → (-31, 15)
// rot3: (2,1)   → (-1,  1 )
var kWallScreenOffsets = [4][2]int16{
	{0, 0},    // rotation 0 (SE edge)
	{30, 16},  // rotation 1 (SW edge)
	{-31, 15}, // rotation 2 (NW edge)
	{-1, 1},   // rotation 3 (NE edge)
}

// kWallImageOffsets[rotation][slopeIndex] → WallObj sprite offset.
// OpenLoco source: PaintWall.cpp kImageOffsets[not-twoSided][rotation][slope]
//
//	kFlatSE=0, kFlatNE=1, kSlopedSE=2, kSlopedNE=3, kSlopedNW=4, kSlopedSW=5
//
// slopeIndex: 0=downwards, 1=upwards, 2=flat  (from slopeFlagsToIndex)
var kWallImageOffsets = [4][3]uint32{
	{3, 5, 1}, // rot 0: kSlopedNE, kSlopedSW, kFlatNE
	{2, 4, 0}, // rot 1: kSlopedSE, kSlopedNW, kFlatSE
	{5, 3, 1}, // rot 2: kSlopedSW, kSlopedNE, kFlatNE
	{4, 2, 0}, // rot 3: kSlopedNW, kSlopedSE, kFlatSE
}

// paintTracks renders railway track sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrack.cpp paintTrack()
//
// Sprite layout: kTrackParts[trackID][seqIdx] → image offsets per rotation and layer.
// Mergeable pieces: 3 layers (Ballast=0, Sleeper=1, Rail=2).
// Non-mergeable (slope) pieces: only layer 0.
func (w *World) paintTracks(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.tracks) == 0 {
		return
	}

	for _, te := range t.tracks {
		trackObj := w.renderer.ObjMgr.GetTrackObjectByIndex(int(te.TrackObjectID))
		if trackObj == nil || trackObj.Image == 0 {
			continue
		}

		trackID := int(te.TrackID)
		if trackID >= len(kTrackParts) || kTrackParts[trackID] == nil {
			continue
		}
		pieces := kTrackParts[trackID]

		seqIdx := int(te.SequenceIndex)
		if seqIdx >= len(pieces) {
			continue
		}
		piece := pieces[seqIdx]

		rotation := int(te.Rotation & 0x03)
		extraHeight := float64(int(te.BaseZ)-int(t.baseZ)) * 4.0

		// Determine how many layers to render.
		maxLayer := 3
		if piece.nonMergeable {
			maxLayer = 1
		}

		for layer := 0; layer < maxLayer; layer++ {
			offset := piece.img[rotation][layer]
			if offset == 0 {
				continue
			}
			spriteID := int(trackObj.Image) + int(offset)

			img := w.renderer.GetSprite(spriteID)
			if img == nil {
				continue
			}
			_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
			if !ok {
				continue
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(
				drawX+float64(xOff)*scale,
				drawY+(float64(yOff)-extraHeight)*scale)
			screen.DrawImage(img, op)
		}
	}
}

// paintRoads renders road sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintRoad.cpp paintRoad()
//
// Road IDs 0–2 (straight, left/right very-small curve) use the multi-tile merge system:
// each element on the tile contributes bits to an exits bitmask, and one composite sprite
// is rendered via kExitsToMergeId + kMergeBaseOffset (like OpenLoco's finalisePaintRoad).
// Road IDs 3–9 render directly from the piece tables.
func (w *World) paintRoads(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.roads) == 0 {
		return
	}

	// Accumulate exits for the merge system, keyed by (roadObjectID, baseZ).
	type mergeKey struct{ objID, baseZ uint8 }
	type mergeState struct {
		roadObj     *objects.RoadObject
		exits       uint8 // Style0 / Style2-left exits
		exitsRight  uint8 // Style2-right exits (dual carriageway only)
		extraHeight float64
		hasLeft     bool
		hasRight    bool
	}
	var merges [8]mergeState // fixed-size to avoid allocation on the hot path
	mergeIdx := map[mergeKey]int{}
	nMerges := 0

	for _, re := range t.roads {
		roadObj := w.renderer.ObjMgr.GetRoadObjectByIndex(int(re.RoadObjectID))
		if roadObj == nil || roadObj.Image == 0 {
			continue
		}

		roadID := int(re.RoadID)
		seqIdx := int(re.SequenceIndex)
		rotation := int(re.Rotation & 0x03)
		extraHeight := float64(int(re.BaseZ)-int(t.baseZ)) * 4.0

		// Road IDs 0–2 use the multi-tile merge (junction) system.
		if roadID <= 2 {
			key := mergeKey{re.RoadObjectID, re.BaseZ}
			idx, ok := mergeIdx[key]
			if !ok {
				if nMerges >= len(merges) {
					continue // safety cap
				}
				idx = nMerges
				nMerges++
				merges[idx] = mergeState{roadObj: roadObj, extraHeight: extraHeight}
				mergeIdx[key] = idx
			}
			exits := kRoadMergeExits[roadID][rotation]
			if roadObj.PaintStyle == 2 {
				// Style2 dual carriageway: straight rot 0/1 = left lane, rot 2/3 = right.
				// Very-small-right-curve → right lane; left-curve → left lane.
				isRight := (roadID == 0 && rotation >= 2) || roadID == 2
				if isRight {
					merges[idx].exitsRight |= exits
					merges[idx].hasRight = true
				} else {
					merges[idx].exits |= exits
					merges[idx].hasLeft = true
				}
			} else {
				merges[idx].exits |= exits
				merges[idx].hasLeft = true
			}
			continue
		}

		// Road IDs 3–9: render directly from the piece tables.
		switch roadObj.PaintStyle {
		case 0:
			if roadID >= len(kRoadPartsStyle0) {
				continue
			}
			pieces := kRoadPartsStyle0[roadID]
			if seqIdx >= len(pieces) {
				continue
			}
			offset := pieces[seqIdx].img[rotation]
			if offset == 0 {
				continue
			}
			w.drawTrackRoadSprite(screen, int(roadObj.Image)+int(offset), drawX, drawY, extraHeight, scale)
		case 2:
			// OpenLoco reference: RoadObject.h Style2 namespace
			if roadID >= len(kRoadPartsStyle2) {
				continue
			}
			pieces := kRoadPartsStyle2[roadID]
			if seqIdx >= len(pieces) {
				continue
			}
			offset := pieces[seqIdx].img[rotation]
			if offset == 0 {
				continue
			}
			w.drawTrackRoadSprite(screen, int(roadObj.Image)+int(offset), drawX, drawY, extraHeight, scale)
		case 1:
			// Style1: 3 layers (Ballast, Sleeper, Rail) — no merge system.
			if roadID >= len(kRoadPartsStyle1) {
				continue
			}
			pieces := kRoadPartsStyle1[roadID]
			if seqIdx >= len(pieces) {
				continue
			}
			piece := pieces[seqIdx]
			for layer := 0; layer < 3; layer++ {
				offset := piece.img[rotation][layer]
				if offset == 0 {
					continue
				}
				w.drawTrackRoadSprite(screen, int(roadObj.Image)+int(offset), drawX, drawY, extraHeight, scale)
			}
		}
	}

	// Render the accumulated merge sprites (straight, curve, junction, or crossroads).
	// OpenLoco reference: PaintRoad.cpp finalisePaintRoad()
	for i := 0; i < nMerges; i++ {
		ms := &merges[i]
		if ms.hasLeft {
			mergeID := kExitsToMergeId[ms.exits&0x0F]
			spriteID := int(ms.roadObj.Image) + int(kMergeBaseOffset) + int(mergeID)
			w.drawTrackRoadSprite(screen, spriteID, drawX, drawY, ms.extraHeight, scale)
		}
		if ms.hasRight {
			mergeID := kExitsToMergeId[ms.exitsRight&0x0F]
			spriteID := int(ms.roadObj.Image) + int(kMergeBaseOffsetStyle2Right) + int(mergeID)
			w.drawTrackRoadSprite(screen, spriteID, drawX, drawY, ms.extraHeight, scale)
		}
	}
}

// drawTrackRoadSprite fetches a sprite and draws it at the correct screen position.
func (w *World) drawTrackRoadSprite(screen *ebiten.Image, spriteID int, drawX, drawY, extraHeight, scale float64) {
	img := w.renderer.GetSprite(spriteID)
	if img == nil {
		return
	}
	_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
	if !ok {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(
		drawX+float64(xOff)*scale,
		drawY+(float64(yOff)-extraHeight)*scale)
	screen.DrawImage(img, op)
}

// paintStation renders train station sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrainStation.cpp
//
// Sprite layout (Style0, relative to SequenceImageOffsets[seqIdx]):
//
//	NE orientation (rotation 0/2): +0=BackNE, +1=FrontNE, +2=CanopyNE
//	SE orientation (rotation 1/3): +4=BackSE, +5=FrontSE, +6=CanopySE
func (w *World) paintStation(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.stations) == 0 {
		return
	}

	for _, se := range t.stations {
		stationObj := w.renderer.ObjMgr.GetTrainStationObjectByIndex(int(se.ObjectID))
		if stationObj == nil || stationObj.ImageOffset == 0 {
			continue
		}

		seqIdx := int(se.SequenceIndex)
		if seqIdx >= len(stationObj.SequenceImageOffsets) {
			seqIdx = 0
		}
		seqBase := int(stationObj.SequenceImageOffsets[seqIdx])

		// Select NE or SE sprite set based on rotation.
		var backOff, frontOff, canopyOff int
		if se.Rotation&1 == 0 { // NE-facing (rotations 0, 2)
			backOff = objects.StStraightBackNE
			frontOff = objects.StStraightFrontNE
			canopyOff = objects.StStraightCanopyNE
		} else { // SE-facing (rotations 1, 3)
			backOff = objects.StStraightBackSE
			frontOff = objects.StStraightFrontSE
			canopyOff = objects.StStraightCanopySE
		}

		extraHeight := float64(int(se.BaseZ)-int(t.baseZ)) * 4.0

		for _, sprOffset := range []int{backOff, frontOff, canopyOff} {
			spriteID := seqBase + sprOffset
			w.drawTrackRoadSprite(screen, spriteID, drawX, drawY, extraHeight, scale)
		}
	}
}

// paintWalls renders wall sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintWall.cpp paintWall()
func (w *World) paintWalls(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.walls) == 0 {
		return
	}

	for _, we := range t.walls {
		wallObj := w.renderer.ObjMgr.GetWallObjectByIndex(int(we.WallObjectID))
		if wallObj == nil || wallObj.Sprite == 0 {
			continue
		}

		rot := we.Rotation & 0x03

		// Slope index from EdgeSlope flags (OpenLoco: slopeFlagsToIndex).
		// EdgeSlope::downwards=bit1, EdgeSlope::upwards=bit0.
		slopeIdx := 2 // flat
		if we.EdgeSlope&0x02 != 0 {
			slopeIdx = 0 // downwards
		} else if we.EdgeSlope&0x01 != 0 {
			slopeIdx = 1 // upwards
		}

		spriteOffset := kWallImageOffsets[rot][slopeIdx]
		spriteID := int(wallObj.Sprite + spriteOffset)

		img := w.renderer.GetSprite(spriteID)
		if img == nil {
			continue
		}

		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			continue
		}

		// Height offset for wall's baseZ vs surface baseZ
		extraHeight := float64(int(we.BaseZ)-int(t.baseZ)) * 4.0

		// Screen-space edge offset (world sub-tile Pos3 projected to pixels).
		edgeX := float64(kWallScreenOffsets[rot][0])
		edgeY := float64(kWallScreenOffsets[rot][1])

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(
			drawX+(float64(xOff)+edgeX)*scale,
			drawY+(float64(yOff)-extraHeight+edgeY)*scale)
		screen.DrawImage(img, op)
	}
}

// getWaterImage returns a cached translucent blue diamond overlay for water tiles.
func (w *World) getWaterImage() *ebiten.Image {
	if w.waterImage != nil {
		return w.waterImage
	}
	img := ebiten.NewImage(w.tileW, w.tileH)
	centerX := w.tileW / 2
	centerY := w.tileH / 2
	waterColor := color.RGBA{40, 80, 180, 140}
	for y := 0; y < w.tileH; y++ {
		var halfWidth int
		if y < centerY {
			halfWidth = (y * centerX) / centerY
		} else {
			halfWidth = ((w.tileH - 1 - y) * centerX) / centerY
		}
		for x := centerX - halfWidth; x <= centerX+halfWidth; x++ {
			if x >= 0 && x < w.tileW {
				img.Set(x, y, waterColor)
			}
		}
	}
	w.waterImage = img
	return img
}

// paintWater renders a water surface overlay on a tile.
//
// OpenLoco uses a destination-based palette blend for the +35 sprite:
//   for each non-transparent pixel in the blend shape, remap the destination
//   (terrain) pixel through the water palette map → water-blue output.
// We approximate this by pre-rendering the +35 sprite as solid water blue
// (colour extracted from the water palette map at ImageOffset+41), then
// drawing the +30 sprite (ripple/highlight detail) on top at full opacity.
//
// OpenLoco reference: PaintSurface.cpp:1720-1759 (paintSurfaceWater)
//   sprite+35 .withBlend(ExtColour::water) → dst-based remap (mask + tint)
//   sprite+30 attached → highlight detail drawn on top
func (w *World) paintWater(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	// waterLevel is MicroZ (×16 px/unit); baseZ is SmallZ (×4 px/unit).
	waterHeightDiff := float64(t.waterLevel)*16.0 - float64(t.baseZ)*4.0

	if w.renderer != nil && w.renderer.ObjMgr != nil && w.renderer.ObjMgr.WaterObj != nil {
		waterObj := w.renderer.ObjMgr.WaterObj
		shape := objects.KSlopeToWaterShape[t.slope&0x0F]
		blendID := int(waterObj.GetWaterSpriteIndex(shape, true))
		flatID := int(waterObj.GetWaterSpriteIndex(shape, false))

		drawSprite := func(spriteID int, img *ebiten.Image) {
			_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
			if !ok {
				xOff, yOff = -32, -15
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(
				drawX+float64(xOff)*scale,
				drawY+float64(yOff)*scale-waterHeightDiff*scale)
			screen.DrawImage(img, op)
		}

		// ── layer 1: solid-blue blend sprite (+35) — full tile rhombus ──
		if blendImg := w.buildWaterBlendImage(int(shape), blendID); blendImg != nil {
			drawSprite(blendID, blendImg)
		}
		// ── layer 2: flat ripple/highlight sprite (+30) — drawn on top ──
		if flatImg := w.renderer.GetSprite(flatID); flatImg != nil {
			drawSprite(flatID, flatImg)
		}
		return
	}

	// Fallback: hand-drawn translucent blue diamond
	waterImg := w.getWaterImage()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(
		drawX+float64(-32)*scale,
		drawY+float64(-15)*scale-waterHeightDiff*scale)
	screen.DrawImage(waterImg, op)
}

// buildWaterBlendImage returns a cached solid-blue version of water blend sprite +35.
// Each non-transparent pixel is forced to the water colour derived from the water
// object's palette map at ImageOffset+41, approximating OpenLoco's dst-based blend.
//
// OpenLoco reference: ObjectManager.cpp updateWaterPalette()
//   paletteImageId = waterObj->image + Water::ImageIds::kColourPalette (= 41)
func (w *World) buildWaterBlendImage(shape, blendSpriteID int) *ebiten.Image {
	if w.waterBlendCache == nil {
		w.waterBlendCache = make(map[int]*ebiten.Image)
	}
	if img, ok := w.waterBlendCache[shape]; ok {
		return img
	}
	srcImg := w.renderer.GetSprite(blendSpriteID)
	if srcImg == nil {
		return nil
	}
	wc := w.waterColourFromPalette()
	bw, bh := srcImg.Bounds().Dx(), srcImg.Bounds().Dy()
	pixels := make([]byte, bw*bh*4)
	srcImg.ReadPixels(pixels)
	for i := 0; i < len(pixels); i += 4 {
		if pixels[i+3] > 0 { // non-transparent pixel → force to water blue
			pixels[i+0] = wc.R
			pixels[i+1] = wc.G
			pixels[i+2] = wc.B
			pixels[i+3] = 255
		}
	}
	result := ebiten.NewImage(bw, bh)
	result.WritePixels(pixels)
	w.waterBlendCache[shape] = result
	return result
}

// waterColourFromPalette samples the water palette map (WaterObject sprite +41)
// to derive the actual water blue.  The map converts any terrain index into a
// water-tinted shade; we average several mid-range samples to get a representative
// water colour.
func (w *World) waterColourFromPalette() color.RGBA {
	const defR, defG, defB uint8 = 35, 100, 170
	if w.renderer == nil || w.renderer.G1 == nil || w.renderer.ObjMgr == nil || w.renderer.ObjMgr.WaterObj == nil {
		return color.RGBA{defR, defG, defB, 255}
	}
	palMap, err := w.renderer.G1.GetPaletteMapAt(int(w.renderer.ObjMgr.WaterObj.ImageOffset) + 41)
	if err != nil {
		return color.RGBA{defR, defG, defB, 255}
	}
	samples := []byte{10, 30, 60, 90, 120, 150, 180, 210}
	var rSum, gSum, bSum uint32
	for _, idx := range samples {
		c := w.renderer.G1.Palette[palMap[idx]]
		rSum += uint32(c.R)
		gSum += uint32(c.G)
		bSum += uint32(c.B)
	}
	n := uint32(len(samples))
	return color.RGBA{R: uint8(rSum / n), G: uint8(gSum / n), B: uint8(bSum / n), A: 255}
}

// getCliffEdgeOffset converts kEdgeImageOffset[edge] + Pos3{0,0,h*kMicroZStep}
// to viewport (vpX_off, vpY_off) using gameToScreen:
//
//	vpX_off = posY - posX
//	vpY_off = (posY+posX)/2 - (posZ_base + h*16)
//
// h is in MicroZ units; 1 MicroZ = 16 viewport pixels (kMicroZStep=16).
// See MEASUREMENTS.md "Cliff edge Pos3 → screen offset" table.
func (w *World) getCliffEdgeOffset(edge int, h uint8) (int16, int16) {
	heightPx := int16(h) * 16
	switch edge {
	case EdgeSW: // Pos3{30,0,0}: vpX=0-30=-30, vpY=(0+30)/2-0=15
		return -30, 15 - heightPx
	case EdgeSE: // Pos3{0,30,0}: vpX=30-0=30, vpY=(30+0)/2-0=15
		return 30, 15 - heightPx
	case EdgeNW: // Pos3{0,-2,1}: vpX=-2-0=-2, vpY=(-2+0)/2-1=-2
		return -2, -2 - heightPx
	case EdgeNE: // Pos3{-2,0,1}: vpX=0-(-2)=2, vpY=(0-2)/2-1=-2
		return 2, -2 - heightPx
	default:
		return 0, -heightPx
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	sw, sh := screen.Size()
	scale := 1.0 / float64(int(1)<<w.zoom) // zoom 0→1.0, 1→0.5, 2→0.25, 3→0.125

	// Camera coords are in world-space (unscaled). They are set by
	// PanCamera (gameplay) or SetCamera (title sequence).

	// Draw tiles in depth order (back to front)
	for depth := 0; depth < w.width+w.height; depth++ {
		for y := 0; y < w.height; y++ {
			x := depth - y
			if x < 0 || x >= w.width {
				continue
			}

			t := w.tiles[y][x]
			screenX, screenY := w.tileToScreen(x, y)

			// OpenLoco: loc.z = baseZ * kSmallZStep = baseZ * 4 viewport pixels.
			// Cliff edge sprites fill the gaps at height transitions.
			// OpenLoco ref: TileElementBase.h baseHeight() = _baseZ * kSmallZStep
			heightOffset := float64(t.baseZ) * 4.0
			screenY -= heightOffset

			// Convert world position to screen position: subtract camera,
			// then apply zoom scale. Round to nearest integer so adjacent tile
			// anchors snap consistently — sub-pixel sprite offsets then tend
			// toward overlap rather than leaving gaps.
			drawX := math.Round((screenX - w.camX) * scale)
			drawY := math.Round((screenY - w.camY) * scale)

			// Skip tiles that are entirely off-screen (with margin for sprite offsets)
			if drawX < -128 || drawX > float64(sw)+64 || drawY < -128 || drawY > float64(sh)+64 {
				continue
			}


			// Draw terrain using LandObject embedded sprites
			// OpenLoco reference: Paint/PaintSurface.cpp paintSurface()
			//   imageIndex = landObj->image + variation + displaySlope
			if w.renderer != nil && w.renderer.ObjMgr != nil {
				land := w.renderer.ObjMgr.GetLandObjectByIndex(int(t.terrainIndex))
				if land != nil {
					// Map slope byte (0-31) to display slope (0-18)
					rawSlope := t.slope & 0x1F // Lower 5 bits
					displaySlope := int(0)
					if rawSlope < uint8(len(slopeToDisplaySlope)) {
						displaySlope = int(slopeToDisplaySlope[rawSlope])
					}

					// Zoom-0 flat terrain sprite: skip all zoom3/zoom2/zoom1 images
					// (numAngles * numGrowthStages * 57) then add slope variant.
					// Growth stage variation is NOT applied here: the correct step between
					// growth stages in zoom-0 is (numAngles*19), not NumImagesPerGrowthStage
					// (which includes blend images). Using NumImagesPerGrowthStage pushes
					// spriteIdx out of bounds for any tile with growthStage>0 (blank tile).
					// OpenLoco reference: src/OpenLoco/src/Objects/LandObject.cpp getTerrainImage()
					spriteIdx := land.GetFlatTerrainSpriteIndex() + displaySlope

					if img := w.renderer.GetObjectSprite(land, spriteIdx); img != nil {
						_, _, xOff, yOff, ok := w.renderer.GetObjectSpriteInfo(land, spriteIdx)
						if ok {
							op := &ebiten.DrawImageOptions{}
							op.GeoM.Scale(scale, scale)
							op.GeoM.Translate(
								drawX+float64(xOff)*scale,
								drawY+float64(yOff)*scale)
							screen.DrawImage(img, op)

							// Draw cliff edges for height transitions
							// OpenLoco reference: PaintSurface.cpp:1714-1717
							// Now uses dynamic G1 sprite pool with land.CliffEdgeImage
							if land.CliffEdgeImage > 0 {
								w.paintCliffEdges(screen, x, y, &t, land, drawX, drawY, scale)
							}

							// Draw water when surface is above terrain base (matches OpenLoco attachToPrevious).
							// effectiveDrawY cancels the height offset inside paintWater so the
							// sprite aligns with the terrain; partial-shape sprites handle the rest.
							if t.waterLevel > 0 && uint16(t.waterLevel)*16 > uint16(t.baseZ)*4 {
								effectiveDrawY := drawY + (float64(t.waterLevel)*16.0-float64(t.baseZ)*4.0)*scale
								w.paintWater(screen, &t, drawX, effectiveDrawY, scale)
								// Water cliff edges at shore level boundaries
								w.paintWaterCliffEdges(screen, x, y, &t, land, drawX, drawY, scale)
							}

							// Draw infrastructure and scenery on this tile
							w.paintTracks(screen, &t, drawX, drawY, scale)
							w.paintRoads(screen, &t, drawX, drawY, scale)
							w.paintStation(screen, &t, drawX, drawY, scale)
							w.paintTrees(screen, &t, drawX, drawY, scale)
							w.paintBuildings(screen, &t, drawX, drawY, scale)
							w.paintWalls(screen, &t, drawX, drawY, scale)

							continue
						}
					}
				}
			}

			// Fall back to colored diamond
			img := w.getFallbackImage(t.tileType)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(drawX, drawY)
			screen.DrawImage(img, op)
		}
	}

	// Vehicle rendering — separate post-tile pass (depth approximation).
	// Use real parsed entities when available (loaded .SV5), otherwise demo trains.
	if len(w.entities) > 0 {
		w.paintVehicleEntities(screen, scale)
	} else {
		w.paintDemoTrains(screen, scale)
	}
}

// paintDemoTrains renders all active demo trains after the tile loop.
// OpenLoco reference: src/OpenLoco/src/Paint/PaintVehicle.cpp
func (w *World) paintDemoTrains(screen *ebiten.Image, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return
	}
	vehicles := w.renderer.ObjMgr.Vehicles
	if len(vehicles) == 0 {
		return
	}

	for _, tr := range w.trains {
		if tr.tileX < 0 || tr.tileX >= w.width || tr.tileY < 0 || tr.tileY >= w.height {
			continue
		}

		// Fractional world position (smooth between tiles)
		fracX := float64(tr.tileX) + float64(tr.dx)*tr.progress
		fracY := float64(tr.tileY) + float64(tr.dy)*tr.progress
		vpX := (fracY - fracX) * 32.0
		vpY := (fracY + fracX) * 16.0

		// Height: interpolate between current and next tile
		baseZ := float64(w.tiles[tr.tileY][tr.tileX].baseZ)
		nx := tr.tileX + tr.dx
		ny := tr.tileY + tr.dy
		if nx >= 0 && nx < w.width && ny >= 0 && ny < w.height {
			nextZ := float64(w.tiles[ny][nx].baseZ)
			baseZ = baseZ*(1.0-tr.progress) + nextZ*tr.progress
		}
		vpY -= baseZ * 4.0

		drawX := math.Round((vpX-w.camX)*scale) - 16 // centre horizontally
		drawY := math.Round((vpY-w.camY)*scale) - 16 // centre vertically

		// Direction → sprite rotation (32-direction wheel; 0=SE, 8=SW, 16=NW, 24=NE)
		dir32 := 0
		switch {
		case tr.dy > 0 && tr.dx == 0:
			dir32 = 0
		case tr.dx > 0 && tr.dy == 0:
			dir32 = 8
		case tr.dy < 0 && tr.dx == 0:
			dir32 = 16
		case tr.dx < 0 && tr.dy == 0:
			dir32 = 24
		}

		v := vehicles[tr.vehicleIdx%len(vehicles)]
		if v.ImageOffset == 0 {
			continue // sprites not loaded
		}

		// Use the body sprite for car component 0
		bodyIdx := int(v.CarComponents[0].BodySpriteIdx)
		spriteID, ok := v.GetBodySpriteID(bodyIdx, dir32)
		if !ok {
			continue
		}

		// Apply company-colour palette remap (default ColourBlue=7).
		// OpenLoco reference: PaintVehicle.cpp — ImageId::withPrimary(companyColour)
		const defaultCompanyColour = 7
		img := w.renderer.GetSpriteColoured(int(spriteID), defaultCompanyColour)
		if img == nil {
			img = w.renderer.GetSprite(int(spriteID))
		}
		if img == nil {
			continue
		}
		_, _, xOff, yOff, hasInfo := w.renderer.GetSpriteInfo(int(spriteID))
		if hasInfo {
			drawX += float64(xOff) * scale
			drawY += float64(yOff) * scale
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(drawX, drawY)
		screen.DrawImage(img, op)
	}
}

// paintVehicleEntities renders VehicleBody entities parsed from an .SV5 save file.
// Each body entity carries its own tile position (tileX/tileY) and sprite info
// (objectId, objectSpriteType, spriteYaw) — no interpolation needed.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintVehicle.cpp
func (w *World) paintVehicleEntities(screen *ebiten.Image, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return
	}
	vehicles := w.renderer.ObjMgr.Vehicles
	if len(vehicles) == 0 {
		return
	}

	for i := range w.entities {
		ent := &w.entities[i]
		if ent.EntityType != scenario.VehicleTypeBody {
			continue // only Body entities have renderable car sprites
		}

		// Bounds check: tile coordinates from the save must be on the loaded map.
		tx := int(ent.TileX)
		ty := int(ent.TileY)
		if tx < 0 || tx >= w.width || ty < 0 || ty >= w.height {
			continue
		}

		// Convert tile → viewport (same formula as tileToScreen).
		// For sub-tile precision we shift by the sub-tile offset within the tile:
		//   PosX = tileX*32 + subtileX,  so subtileX = PosX - tileX*32
		subtileX := float64(ent.PosX) - float64(ent.TileX)*32.0
		subtileY := float64(ent.PosY) - float64(ent.TileY)*32.0
		fracX := float64(tx) + subtileX/32.0
		fracY := float64(ty) + subtileY/32.0

		vpX := (fracY - fracX) * 32.0
		vpY := (fracY + fracX) * 16.0

		// Height: use entity's PosZ (world SmallZ) for pixel offset.
		// OpenLoco formula: heightPx = posZ * 4 (kSmallZStep = 4 px/unit).
		// We apply the current global compromise (×2 instead of ×4).
		vpY -= float64(ent.PosZ) * 2.0

		drawX := math.Round((vpX-w.camX)*scale) - 16
		drawY := math.Round((vpY-w.camY)*scale) - 16

		// Vehicle object lookup.  objectId is 0-based index into loaded vehicles.
		if int(ent.ObjectID) >= len(vehicles) {
			continue
		}
		v := vehicles[ent.ObjectID]
		if v == nil || v.ImageOffset == 0 {
			continue
		}

		// Direction: spriteYaw is stored 0-63 (64 steps per full rotation).
		// Halve to get 0-31 for 32-direction sprite tables.
		dir32 := int(ent.SpriteYaw) >> 1

		bodyIdx := int(ent.ObjectSpriteType)
		spriteID, ok := v.GetBodySpriteID(bodyIdx, dir32)
		if !ok {
			continue
		}

		// Apply company-colour palette remap (default ColourBlue=7).
		// OpenLoco reference: PaintVehicle.cpp — ImageId::withPrimary(companyColour)
		const defaultCompanyColour = 7
		img := w.renderer.GetSpriteColoured(int(spriteID), defaultCompanyColour)
		if img == nil {
			img = w.renderer.GetSprite(int(spriteID))
		}
		if img == nil {
			continue
		}
		_, _, xOff, yOff, hasInfo := w.renderer.GetSpriteInfo(int(spriteID))
		if hasInfo {
			drawX += float64(xOff) * scale
			drawY += float64(yOff) * scale
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(drawX, drawY)
		screen.DrawImage(img, op)
	}
}
