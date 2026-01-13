package main

import (
	graphics "github.com/LaPingvino/goloco/pkg/_conversion_stub/graphics"
	worldstub "github.com/LaPingvino/goloco/pkg/_conversion_stub/world"
	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/world"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"log"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Game struct {
	w *world.World
	r *render.Renderer
}

func NewGame() *Game {
	// Initialize graphics subsystem
	if err := graphics.LoadG1(); err != nil {
		log.Printf("Failed to load graphics: %v", err)
	}
	_ = worldstub.TownManager{}

	// Try to load atlas from assets directory, or generate placeholder
	r := render.NewRenderer()

	// If atlas is nil, try to generate a placeholder
	if r.Atlas == nil {
		log.Printf("No atlas found, generating placeholder...")
		if err := assets.GeneratePlaceholderAtlas("assets"); err != nil {
			log.Printf("Failed to generate placeholder: %v", err)
		} else {
			// Try loading again
			if at, err := render.LoadAtlasFromDir("assets"); err == nil {
				r.Atlas = at
			}
		}
	}

	w := world.NewWorld(r)
	return &Game{w: w, r: r}
}

func (g *Game) Update() error {
	// input handling: arrows and WASD
	var dx, dy int
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		dx = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		dx = 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		dy = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		dy = 1
	}
	if dx != 0 || dy != 0 {
		g.w.HandleInput(dx, dy)
	}
	g.w.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// set renderer screen for future use
	g.r.SetScreen(screen)
	g.w.Draw(screen)
	ebitenutil.DebugPrint(screen, "goloco - minimal shell (arrows/WASD to move)")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := NewGame()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("goloco - experiment")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
