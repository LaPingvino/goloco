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
	trees        []scenario.TreeElement
	buildings    []scenario.BuildingElement
	walls        []scenario.WallElement
	tracks       []scenario.TrackElement
	roads        []scenario.RoadElement
	stations     []scenario.StationElement
	signals      []scenario.SignalElement
	industries   []scenario.IndustryElement
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
	tileCache  map[TileType]*ebiten.Image
	waterImage *ebiten.Image // cached translucent water diamond overlay
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

func (w *World) Update() {
	// Currently a no-op; camera is driven by PanCamera from the game loop.
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
						math.Floor(drawX+float64(xOff+edgeOffsetX)*scale),
						math.Floor(drawY+float64(yOff+edgeOffsetY)*scale))
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
			math.Floor(drawX+(float64(xOff)+qx)*scale),
			math.Floor(drawY+(float64(yOff)-extraHeight+qy)*scale))
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

		// For multi-tile (2x2) buildings, only render from one tile
		// (sequenceIndex check). Simplified: only render sequenceIndex 0.
		if bldgObj.HasFlags(objects.BuildingFlagLargeTile) && be.SequenceIndex != 0 {
			continue
		}

		rotation := be.Rotation & 0x03
		variation := be.Variation
		if int(variation) >= int(bldgObj.NumVariations) {
			variation = 0
		}

		parts := bldgObj.GetBuildingParts(variation)
		if len(parts) == 0 {
			continue
		}

		// Height offset for building's own baseZ vs surface baseZ
		extraHeight := float64(int(be.BaseZ)-int(t.baseZ)) * 4.0

		// Image offset for 1x1 buildings: center of tile (16, 16)
		offsetX := float64(16)
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
				math.Floor(drawX+(float64(xOff)+offsetX)*scale),
				math.Floor(drawY+(float64(yOff)-extraHeight-partZ+offsetY)*scale))
			screen.DrawImage(img, op)

			// Stack next part on top
			if int(part) < len(bldgObj.PartHeights) {
				partZ += float64(bldgObj.PartHeights[part])
			}
		}
	}
}

// Wall sprite offset constants per rotation, from OpenLoco PaintWall.cpp kOffsets.
// These position the wall sprite along the correct tile edge.
var kWallOffsets = [4][2]int16{
	{0, 0},    // rotation 0 (SE edge)
	{1, 31},   // rotation 1 (SW edge)
	{31, 0},   // rotation 2 (NW edge)
	{2, 1},    // rotation 3 (NE edge)
}

// paintTracks renders railway track sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrack.cpp paintTrack()
//
// Sprite layout per TrackObject image table:
//
//	Each track piece has 4 rotation variants.
//	spriteID = trackObj.Image + trackID * 4 + rotation
func (w *World) paintTracks(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.tracks) == 0 {
		return
	}

	for _, te := range t.tracks {
		trackObj := w.renderer.ObjMgr.GetTrackObjectByIndex(int(te.TrackObjectID))
		if trackObj == nil || trackObj.Image == 0 {
			continue
		}

		rotation := te.Rotation & 0x03
		spriteID := int(trackObj.Image) + int(te.TrackID)*4 + int(rotation)

		img := w.renderer.GetSprite(spriteID)
		if img == nil {
			continue
		}

		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			continue
		}

		extraHeight := float64(int(te.BaseZ)-int(t.baseZ)) * 4.0

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(
			math.Floor(drawX+float64(xOff)*scale),
			math.Floor(drawY+(float64(yOff)-extraHeight)*scale))
		screen.DrawImage(img, op)
	}
}

// paintRoads renders road sprites for a tile.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintRoad.cpp paintRoad()
//
// Sprite layout per RoadObject image table:
//
//	Each road piece has 4 rotation variants.
//	spriteID = roadObj.Image + roadID * 4 + rotation
func (w *World) paintRoads(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	if w.renderer == nil || w.renderer.ObjMgr == nil || len(t.roads) == 0 {
		return
	}

	for _, re := range t.roads {
		roadObj := w.renderer.ObjMgr.GetRoadObjectByIndex(int(re.RoadObjectID))
		if roadObj == nil || roadObj.Image == 0 {
			continue
		}

		rotation := re.Rotation & 0x03
		spriteID := int(roadObj.Image) + int(re.RoadID)*4 + int(rotation)

		img := w.renderer.GetSprite(spriteID)
		if img == nil {
			continue
		}

		_, _, xOff, yOff, ok := w.renderer.GetSpriteInfo(spriteID)
		if !ok {
			continue
		}

		extraHeight := float64(int(re.BaseZ)-int(t.baseZ)) * 4.0

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(
			math.Floor(drawX+float64(xOff)*scale),
			math.Floor(drawY+(float64(yOff)-extraHeight)*scale))
		screen.DrawImage(img, op)
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

		// Simplified sprite selection: flat wall for now (ignore slope flags)
		// rotation 0,2 → SE/NW edge → sprite 0 (kFlatSE)
		// rotation 1,3 → NE/SW edge → sprite 1 (kFlatNE)
		var spriteOffset uint32
		switch we.Rotation {
		case 0, 2:
			spriteOffset = 0 // kFlatSE
		case 1, 3:
			spriteOffset = 1 // kFlatNE
		}

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

		// Position along tile edge based on rotation
		rot := we.Rotation & 0x03
		edgeX := float64(kWallOffsets[rot][0])
		edgeY := float64(kWallOffsets[rot][1])

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(
			math.Floor(drawX+(float64(xOff)+edgeX)*scale),
			math.Floor(drawY+(float64(yOff)-extraHeight+edgeY)*scale))
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

// paintWater renders a translucent blue water overlay on a tile.
// The water sits at the waterLevel height, which may be above the terrain baseZ.
//
// OpenLoco reference: PaintSurface.cpp:1720-1759 (paintSurfaceWater)
func (w *World) paintWater(screen *ebiten.Image, t *tile, drawX, drawY, scale float64) {
	waterImg := w.getWaterImage()

	// water field is MicroZ (kMicroZStep=16 px/unit); baseZ is SmallZ (kSmallZStep=4 px/unit).
	// waterHeightDiff = how many px above terrain base the water surface sits.
	// drawY already accounts for baseZ*4, so we only need the additional water offset.
	waterHeightDiff := float64(t.waterLevel)*16.0 - float64(t.baseZ)*4.0

	// Apply the same centering offsets as flat terrain sprites:
	//   xOff=-32  centres the 64px wide diamond on the tile anchor
	//   yOff=-15  places the top of the sprite at the terrain surface level
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(
		math.Floor(drawX+float64(-32)*scale),
		math.Floor(drawY+float64(-15)*scale-waterHeightDiff*scale))
	screen.DrawImage(waterImg, op)
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

							// Draw cliff edges for height transitions
							// OpenLoco reference: PaintSurface.cpp:1714-1717
							// Now uses dynamic G1 sprite pool with land.CliffEdgeImage
							if land.CliffEdgeImage > 0 {
								w.paintCliffEdges(screen, x, y, &t, land, drawX, drawY, scale)
							}

							// Draw water overlay if this tile has water
							if t.waterLevel > 0 {
								w.paintWater(screen, &t, drawX, drawY, scale)
							}

							// Draw infrastructure and scenery on this tile
							w.paintTracks(screen, &t, drawX, drawY, scale)
							w.paintRoads(screen, &t, drawX, drawY, scale)
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
			op.GeoM.Translate(math.Floor(drawX), math.Floor(drawY))
			screen.DrawImage(img, op)
		}
	}

}
