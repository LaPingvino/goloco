package world

import (
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image/color"
	"math"
)

// World holds a small tile grid and a single entity for a minimal playable slice.
type World struct {
	renderer *render.Renderer
	width    int
	height   int
	tiles    [][]color.RGBA
	// tile size in pixels
	ts int
	// cache images per unique tile color
	tileCache map[color.RGBA]*ebiten.Image
	// optional atlas tile name to use for drawing (if non-empty)
	tileImgName string
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
	moveProgress float64 // 0.0..1.0
	moveSpeed    float64 // fraction per update
}

func NewWorld(r *render.Renderer) *World {
	w := &World{
		renderer: r,
		width:    20,
		height:   15,
		// isometric tile size (width, height)
		ts:        64,
		tiles:     make([][]color.RGBA, 15),
		tileCache: make(map[color.RGBA]*ebiten.Image),
		playerX:   10,
		playerY:   7,
		moveSpeed: 0.25, // progress per Update call
	}
	for y := 0; y < w.height; y++ {
		w.tiles[y] = make([]color.RGBA, w.width)
		for x := 0; x < w.width; x++ {
			if (x+y)%2 == 0 {
				w.tiles[y][x] = color.RGBA{R: 0x70, G: 0xB0, B: 0x70, A: 0xFF}
			} else {
				w.tiles[y][x] = color.RGBA{R: 0x60, G: 0xA0, B: 0x60, A: 0xFF}
			}
		}
	}
	// if renderer has an atlas loaded, pick the first PNG as tile image
	if w.renderer != nil && w.renderer.Atlas != nil {
		for name := range w.renderer.Atlas.Images {
			w.tileImgName = name
			break
		}
	}
	// initialize player pixel position at tile center
	w.playerPX = float64(w.playerX * w.ts)
	w.playerPY = float64(w.playerY * w.ts)
	return w
}

func (w *World) Update() {
	// advance movement tween
	if w.moving {
		w.moveProgress += w.moveSpeed
		if w.moveProgress >= 1.0 {
			// finish movement
			w.playerX = w.moveToX
			w.playerY = w.moveToY
			w.playerPX = float64(w.playerX * w.ts)
			w.playerPY = float64(w.playerY * w.ts)
			w.moving = false
			w.moveProgress = 0
		} else {
			// interpolate pixel position between from and to
			fx := float64(w.moveFromX * w.ts)
			fy := float64(w.moveFromY * w.ts)
			tx := float64(w.moveToX * w.ts)
			ty := float64(w.moveToY * w.ts)
			// smoothstep interpolation
			t := w.moveProgress
			t = t * t * (3 - 2*t)
			w.playerPX = fx + (tx-fx)*t
			w.playerPY = fy + (ty-fy)*t
		}
	}
}

func (w *World) HandleInput(dx, dy int) {
	// start movement only if not already moving
	if w.moving {
		return
	}
	nx := w.playerX + dx
	ny := w.playerY + dy
	if nx >= 0 && nx < w.width && ny >= 0 && ny < w.height {
		// initiate tween
		w.moveFromX = w.playerX
		w.moveFromY = w.playerY
		w.moveToX = nx
		w.moveToY = ny
		w.moveProgress = 0
		w.moving = true
	}
}

func (w *World) imageForColor(c color.RGBA) *ebiten.Image {
	if img, ok := w.tileCache[c]; ok {
		return img
	}
	img := ebiten.NewImage(w.ts, w.ts)
	img.Fill(c)
	w.tileCache[c] = img
	return img
}

func (w *World) Draw(screen *ebiten.Image) {
	// determine camera such that player is centered (player screen pos computed via iso)
	sw, sh := screen.Size()
	// compute player screen position via isometric projection
	pxInt, pyInt := render.TileToScreen(w.playerX, w.playerY, w.ts, w.ts/2, w.width*w.ts/2, 40)
	px := float64(pxInt) + float64(w.ts)/2
	py := float64(pyInt) + float64(w.ts)/2
	w.camX = px - float64(sw)/2
	w.camY = py - float64(sh)/2
	// clamp (allow larger ranges since iso extents differ)
	maxCamX := float64((w.width + w.height) * w.ts / 2)
	maxCamY := float64((w.width + w.height) * w.ts / 2)
	if w.camX < 0 {
		w.camX = 0
	}
	if w.camY < 0 {
		w.camY = 0
	}
	if maxCamX > 0 && w.camX > maxCamX {
		w.camX = maxCamX
	}
	if maxCamY > 0 && w.camY > maxCamY {
		w.camY = maxCamY
	}

	// draw tiles in depth order (x + y) for isometric layering
	originX := w.width * w.ts / 2
	originY := 40 // small top offset
	for depth := 0; depth <= w.width+w.height; depth++ {
		for y := 0; y < w.height; y++ {
			x := depth - y
			if x < 0 || x >= w.width {
				continue
			}
			col := w.tiles[y][x]
			// convert tile coords to screen using isometric helper
			sx, sy := render.TileToScreen(x, y, w.ts, w.ts/2, originX, originY)
			// apply camera
			xp := float64(sx) - w.camX
			yp := float64(sy) - w.camY
			if w.tileImgName != "" && w.renderer != nil && w.renderer.Atlas != nil {
				img := w.renderer.Atlas.Get(w.tileImgName)
				if img != nil {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(math.Floor(xp), math.Floor(yp))
					screen.DrawImage(img, op)
					continue
				}
			}
			img := w.imageForColor(col)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(math.Floor(xp), math.Floor(yp))
			screen.DrawImage(img, op)
		}
	}
	// draw player as red square (use ebitenutil for convenience)
	playerDrawX := w.playerPX - w.camX
	playerDrawY := w.playerPY - w.camY
	ebitenutil.DrawRect(screen, playerDrawX, playerDrawY, float64(w.ts), float64(w.ts), color.RGBA{R: 0xFF, G: 0x22, B: 0x22, A: 0xFF})
}
