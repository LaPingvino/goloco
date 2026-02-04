package world

import (
	"image/color"
	"math"

	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
	baseZ        uint8 // height in SmallZ units (4 units = 1 pixel vertical)
	slope        uint8 // slope corners and flags from byte 4 of surface element
}

// World holds a tile grid for an isometric game world
type World struct {
	renderer *render.Renderer
	width    int
	height   int
	tiles    [][]tile
	// isometric tile dimensions (standard 2:1 ratio)
	tileW int // tile width in pixels (64)
	tileH int // tile height in pixels (32)
	// cache colored diamond images per tile type (fallback)
	tileCache map[TileType]*ebiten.Image
	// camera offset in pixels (in world-space before zoom is applied)
	camX float64
	camY float64
	// when true, Draw() does not recentre the camera on the player —
	// an external caller (e.g. title sequence) is driving it via SetCamera.
	externalCamera bool
	// zoom level 0–3 matching OpenLoco convention:
	//   0 = full size (1×), 1 = half (2×), 2 = quarter (4×), 3 = eighth (8×)
	// The pixel scale factor is 1 << zoom applied in reverse:
	// world coordinates are divided by (1 << zoom) for screen placement.
	zoom int
	// player tile coords
	playerX int
	playerY int
	// player pixel position (for smooth movement)
	playerPX float64
	playerPY float64
	// movement tweening
	moving       bool
	moveFromX    int
	moveFromY    int
	moveToX      int
	moveToY      int
	moveProgress float64
	moveSpeed    float64
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
		playerX:   10,
		playerY:   7,
		moveSpeed: 0.15,
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

	// Initialize player screen position
	w.playerPX, w.playerPY = w.tileToScreen(w.playerX, w.playerY)
	return w
}

// tileToScreen converts tile coordinates to screen pixel coordinates
func (w *World) tileToScreen(tileX, tileY int) (float64, float64) {
	// Isometric projection:
	// screenX = (tileX - tileY) * tileW/2
	// screenY = (tileX + tileY) * tileH/2
	screenX := float64((tileX-tileY)*w.tileW/2) + float64(w.width*w.tileW/4)
	screenY := float64((tileX+tileY)*w.tileH/2) + 50 // 50px top margin
	return screenX, screenY
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
			}
		}
	}

	w.playerX, w.playerY = w.width/2, w.height/2
	w.playerPX, w.playerPY = w.tileToScreen(w.playerX, w.playerY)
}

// SetCamera overrides the camera position directly (used by the title
// sequence to pan across the map).  The coordinates are in tile space;
// they are converted to screen pixels using the same projection as
// tileToScreen so the camera tracks smoothly.
func (w *World) SetCamera(tileX, tileY float64) {
	w.externalCamera = true
	// Convert tile coords to world-space pixel position using the same
	// projection as tileToScreen, then subtract half the viewport extent
	// in world-space so the target tile lands at the centre.
	// Half-viewport in world-space = (screenSize / 2) / scale = (screenSize / 2) << zoom
	worldX := (tileX-tileY)*float64(w.tileW)/2 + float64(w.width*w.tileW/4)
	worldY := (tileX+tileY)*float64(w.tileH)/2 + 50
	halfW := 400.0 * float64(int(1)<<w.zoom) // 400 = screenWidth/2
	halfH := 300.0 * float64(int(1)<<w.zoom) // 300 = screenHeight/2
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

func (w *World) Update() {
	if w.moving {
		w.moveProgress += w.moveSpeed
		if w.moveProgress >= 1.0 {
			w.playerX = w.moveToX
			w.playerY = w.moveToY
			w.playerPX, w.playerPY = w.tileToScreen(w.playerX, w.playerY)
			w.moving = false
			w.moveProgress = 0
		} else {
			// Smoothstep interpolation
			t := w.moveProgress
			t = t * t * (3 - 2*t)

			fromX, fromY := w.tileToScreen(w.moveFromX, w.moveFromY)
			toX, toY := w.tileToScreen(w.moveToX, w.moveToY)
			w.playerPX = fromX + (toX-fromX)*t
			w.playerPY = fromY + (toY-fromY)*t
		}
	}
}

func (w *World) HandleInput(dx, dy int) {
	if w.moving {
		return
	}
	nx := w.playerX + dx
	ny := w.playerY + dy
	if nx >= 0 && nx < w.width && ny >= 0 && ny < w.height {
		// Don't walk on water
		if w.tiles[ny][nx].tileType == TileWater {
			return
		}
		w.moveFromX = w.playerX
		w.moveFromY = w.playerY
		w.moveToX = nx
		w.moveToY = ny
		w.moveProgress = 0
		w.moving = true
	}
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

// getNeighborTile returns the neighbor tile in the given edge direction
// Edge directions match OpenLoco's neighbor offset order
func (w *World) getNeighborTile(x, y int, edge int) *tile {
	// Neighbor offsets for each edge (rotation 0)
	// From OpenLoco kNeighbourOffsets
	var dx, dy int
	switch edge {
	case EdgeSW: // 0
		dx, dy = -1, 0
	case EdgeSE: // 1
		dx, dy = 0, 1
	case EdgeNW: // 2
		dx, dy = 0, -1
	case EdgeNE: // 3
		dx, dy = 1, 0
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

// getCornerHeights returns the absolute corner heights for a tile
func (w *World) getCornerHeights(t *tile) CornerHeight {
	if t == nil {
		return CornerHeight{0, 0, 0, 0}
	}

	// Base height in MicroZ units (SmallZ * 4)
	microZ := uint8(t.baseZ) * 4

	// Get relative corner heights from slope
	slope := t.slope & 0x1F // Lower 5 bits
	rel := cornerHeights[slope]

	// Add base height to get absolute corner heights
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

// paintCliffEdges renders cliff edge sprites for height transitions
// OpenLoco reference: PaintSurface.cpp paintSurfaceCliffEdge()
func (w *World) paintCliffEdges(screen *ebiten.Image, x, y int, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.G1 == nil {
		return
	}

	// Get corner heights for this tile
	selfCorners := w.getCornerHeights(t)

	// Check all 4 edges in order: NW, NE, SW, SE (matching OpenLoco order)
	// OpenLoco paints in order: 2, 3, 0, 1
	edges := []int{EdgeNW, EdgeNE, EdgeSW, EdgeSE}

	for _, edge := range edges {
		neighbor := w.getNeighborTile(x, y, edge)
		if neighbor == nil {
			continue
		}

		// Get edge heights
		edgeHeights := w.getEdgeHeights(x, y, edge, selfCorners)

		// Check if there's a height difference requiring cliff edges
		// Cliff edges are needed when self corners are higher than neighbor corners
		if edgeHeights.Self0 <= edgeHeights.Neighbor0 && edgeHeights.Self1 <= edgeHeights.Neighbor1 {
			continue // No cliff needed
		}

		// Paint cliff edge sections
		// For now, use hardcoded G1 cliff edge sprites (3726+ from ImageIds.h)
		// TODO: Use land.CliffEdgeImage when we have dynamic sprite loading
		const cliffEdgeBase = 3726 // cliffEdge0MaskSlope0

		// Determine which corner needs cliff sections
		minHeight := edgeHeights.Neighbor0
		if edgeHeights.Neighbor1 < minHeight {
			minHeight = edgeHeights.Neighbor1
		}
		maxHeight := edgeHeights.Self0
		if edgeHeights.Self1 > maxHeight {
			maxHeight = edgeHeights.Self1
		}

		// Paint cliff sections for each height level
		for h := minHeight; h < maxHeight; h++ {
			// Simple cliff edge sprite (height & 0xF gives us variant)
			spriteID := int(cliffEdgeBase) + int(h&0xF)

			if img := w.renderer.GetSprite(spriteID); img != nil {
				_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
				if ok {
					// Offset cliff edge sprite based on edge direction
					edgeOffsetX, edgeOffsetY := w.getCliffEdgeOffset(edge, h)

					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(scale, scale)
					op.GeoM.Translate(
						math.Floor(drawX+float64(xOff+edgeOffsetX)*scale),
						math.Floor(drawY+float64(yOff+edgeOffsetY)*scale))
					screen.DrawImage(img, op)
				}
			}
		}
	}
}

// getCliffEdgeOffset returns the pixel offset for cliff edge sprites
// OpenLoco reference: kEdgeImageOffset in PaintSurface.cpp
func (w *World) getCliffEdgeOffset(edge int, height uint8) (int16, int16) {
	// Height offset (each MicroZ level is 1 pixel vertical)
	heightOffset := int16(height)

	switch edge {
	case EdgeSW: // 0
		return 30, -heightOffset
	case EdgeSE: // 1
		return 0, 30 - heightOffset
	case EdgeNW: // 2
		return 0, -2 - heightOffset + 1
	case EdgeNE: // 3
		return -2, -heightOffset + 1
	default:
		return 0, -heightOffset
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	sw, sh := screen.Size()
	scale := 1.0 / float64(int(1)<<w.zoom) // zoom 0→1.0, 1→0.5, 2→0.25, 3→0.125

	// Centre camera on player unless an external source (title sequence)
	// is driving it.  Camera coords are in world-space (unscaled).
	if !w.externalCamera {
		w.camX = w.playerPX - float64(sw)/scale/2 + float64(w.tileW)/2
		w.camY = w.playerPY - float64(sh)/scale/2 + float64(w.tileH)/2
	}

	// Draw tiles in depth order (back to front)
	for depth := 0; depth < w.width+w.height; depth++ {
		for y := 0; y < w.height; y++ {
			x := depth - y
			if x < 0 || x >= w.width {
				continue
			}

			t := w.tiles[y][x]
			screenX, screenY := w.tileToScreen(x, y)

			// Apply height offset (4 SmallZ units = 1 pixel vertical)
			// OpenLoco reference: Map/Tile.h - SmallZ units
			heightOffset := float64(t.baseZ) / 4.0
			screenY -= heightOffset

			// Convert world position to screen position: subtract camera,
			// then apply zoom scale.
			drawX := (screenX - w.camX) * scale
			drawY := (screenY - w.camY) * scale

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

					// Calculate sprite index: base flat terrain index (57) + displaySlope
					// This gives us the correct slope variant from the LandObject's sprite table
					spriteIdx := land.GetFlatTerrainSpriteIndex() + displaySlope

					if img := w.renderer.GetObjectSprite(land, spriteIdx); img != nil {
						_, _, xOff, yOff, ok := w.renderer.GetObjectSpriteInfo(land, spriteIdx)
						if ok {
							op := &ebiten.DrawImageOptions{}
							op.GeoM.Scale(scale, scale)
							op.GeoM.Translate(
								math.Floor(drawX+float64(xOff)*scale),
								math.Floor(drawY+float64(yOff)*scale))
							screen.DrawImage(img, op)

							// TODO: Draw cliff edges for height transitions
							// OpenLoco reference: PaintSurface.cpp:1714-1717
							// DISABLED: The G1 sprites at 3726+ are MASK sprites (palette index 255)
							// They show as green strokes covering everything.
							// Need to use land.CliffEdgeImage with proper texture sprites instead.
							// This requires implementing dynamic G1 sprite loading for DAT objects.
							// w.paintCliffEdges(screen, x, y, &t, drawX, drawY, scale)

							continue
						}
					}
				}
			}

			// Fall back to colored diamond
			img := w.getFallbackImage(t.tileType)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(math.Floor(drawX), math.Floor(drawY))
			screen.DrawImage(img, op)
		}
	}

	// Draw player as a small red marker (only when not in title-sequence mode)
	if !w.externalCamera {
		playerDrawX := (w.playerPX - w.camX) * scale
		playerDrawY := (w.playerPY - w.camY) * scale
		markerSize := 16.0 * scale
		ebitenutil.DrawRect(screen, playerDrawX, playerDrawY-markerSize, markerSize, markerSize, color.RGBA{255, 50, 50, 220})
	}
}
