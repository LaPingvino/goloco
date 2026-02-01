package world

import (
	"image/color"
	"log"
	"math"

	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

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
	// camera offset in pixels
	camX float64
	camY float64
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
		tileW:     64, // standard isometric tile width (matches G1 terrain sprites)
		tileH:     16, // standard isometric tile height (matches G1 terrain sprites)
		tiles:     make([][]tile, 15),
		tileCache: make(map[TileType]*ebiten.Image),
		playerX:   10,
		playerY:   7,
		moveSpeed: 0.15,
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
			}
		}
	}

	w.playerX, w.playerY = w.width/2, w.height/2
	w.playerPX, w.playerPY = w.tileToScreen(w.playerX, w.playerY)
}

// ZoomIn is a stub. Not yet implemented.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Window.cpp
//   Window::viewportZoomIn(bool toCursor)
//   Window::viewportZoomSet(int8_t zoomLevel, bool toCursor)
//
// In OpenLoco the zoom level is clamped per-viewport and the viewport
// position is re-centred (optionally on the cursor).  goloco will need
// a zoom field on World (or a shared Viewport struct) and the
// tileToScreen projection updated accordingly.
func (w *World) ZoomIn() {
	log.Println("[World] ZoomIn: stub — not yet implemented")
}

// ZoomOut is a stub. Not yet implemented.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Window.cpp
//   Window::viewportZoomOut(bool toCursor)
//   Window::viewportZoomSet(int8_t zoomLevel, bool toCursor)
func (w *World) ZoomOut() {
	log.Println("[World] ZoomOut: stub — not yet implemented")
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

func (w *World) Draw(screen *ebiten.Image) {
	sw, sh := screen.Size()

	// Center camera on player
	w.camX = w.playerPX - float64(sw)/2 + float64(w.tileW)/2
	w.camY = w.playerPY - float64(sh)/2 + float64(w.tileH)/2

	// Draw tiles in depth order (back to front)
	for depth := 0; depth < w.width+w.height; depth++ {
		for y := 0; y < w.height; y++ {
			x := depth - y
			if x < 0 || x >= w.width {
				continue
			}

			t := w.tiles[y][x]
			screenX, screenY := w.tileToScreen(x, y)
			drawX := screenX - w.camX
			drawY := screenY - w.camY

			// Try to draw from the LandObject sprite for this terrain index.
			// Sprite 0 in a LandObject is the flat terrain tile (slope 0).
			// OpenLoco reference: Paint/PaintSurface.cpp paintSurface()
			//   landObj->image + variation + displaySlope
			if w.renderer != nil && w.renderer.ObjMgr != nil {
				land := w.renderer.ObjMgr.GetLandObjectByIndex(int(t.terrainIndex))
				if land != nil {
					if img := w.renderer.GetObjectSprite(land, 0); img != nil {
						_, _, xOff, yOff, ok := w.renderer.GetObjectSpriteInfo(land, 0)
						if ok {
							op := &ebiten.DrawImageOptions{}
							op.GeoM.Translate(math.Floor(drawX+float64(xOff)), math.Floor(drawY+float64(yOff)))
							screen.DrawImage(img, op)
							continue
						}
					}
				}
			}

			// Fall back to colored diamond
			img := w.getFallbackImage(t.tileType)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(math.Floor(drawX), math.Floor(drawY))
			screen.DrawImage(img, op)
		}
	}

	// Draw player as a small red marker
	playerDrawX := w.playerPX - w.camX
	playerDrawY := w.playerPY - w.camY - 16 // Raise above ground
	ebitenutil.DrawRect(screen, playerDrawX, playerDrawY, 16, 16, color.RGBA{255, 50, 50, 220})
}
