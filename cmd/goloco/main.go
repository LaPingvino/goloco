package main

import (
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
}

func NewGame() *Game {
	r := render.NewRenderer()
	w := world.NewWorld(r)
	return &Game{w: w}
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// temporary draw
	ebitenutil.DebugPrint(screen, "goloco - minimal shell")
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
