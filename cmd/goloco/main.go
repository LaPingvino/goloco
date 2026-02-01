package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/LaPingvino/goloco/pkg/graphics"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/title"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/world"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

type Game struct {
	w             *world.World
	r             *render.Renderer
	toolbar       *ui.Toolbar
	windowMgr     *ui.SimpleWindowManager
	objMgr        *objects.ObjectManager
	audioMgr      *audio.Manager
	titleSeq      *title.Sequence
	mouseX        int
	mouseY        int
	dataDir       string
	inTitleScreen bool
}

func findLocoDataDir() string {
	candidates := []string{
		"../locomotion/Data",
		"locomotion/Data",
		filepath.Join(os.Getenv("HOME"), ".local/share/Steam/steamapps/common/Locomotion/Data"),
		filepath.Join(os.Getenv("HOME"), ".steam/steam/steamapps/common/Locomotion/Data"),
		"/usr/share/games/locomotion/Data",
	}

	for _, dir := range candidates {
		g1Path := filepath.Join(dir, "g1.DAT")
		if _, err := os.Stat(g1Path); err == nil {
			return dir
		}
	}
	return ""
}

func NewGame() *Game {
	log.Println("[Game] Creating new game...")

	r := render.NewRenderer()

	// Initialize audio manager
	log.Println("[Game] Creating audio manager...")
	audioMgr := audio.NewManager()

	// Try to load real Locomotion G1.DAT sprites
	dataDir := findLocoDataDir()
	var objMgr *objects.ObjectManager

	if dataDir != "" {
		log.Printf("[Game] Found Locomotion data directory: %s", dataDir)
		g1Path := filepath.Join(dataDir, "g1.DAT")
		log.Printf("Loading G1 sprites from: %s", g1Path)

		g1, err := assets.LoadG1(g1Path)
		if err != nil {
			log.Printf("Failed to load G1: %v", err)
		} else {
			log.Printf("Loaded %d sprites from G1.DAT", g1.GetSpriteCount())
			r.G1 = g1

			// Populate the UI drawing palette from the loaded G1 so window
			// backgrounds, buttons and other widgets more closely match the
			// original Locomotion look-and-feel.
			//
			// Set the package-level global palette so DrawingContext and UI
			// widgets can use the game's palette everywhere. As a fallback,
			// also attempt to copy the palette into a temporary DrawingContext.
			if graphics.SetGlobalPalette(r.G1) {
				log.Printf("Applied G1 palette to UI drawing context")
			} else {
				// Fallback: still try to copy into a temporary DC for local use
				var tmpDC graphics.DrawingContext
				tmpDC.UseRendererPalette(r.G1)
			}
		}

		// Load objects from ObjData
		objDataDir := filepath.Join(filepath.Dir(dataDir), "ObjData")
		if _, err := os.Stat(objDataDir); err == nil {
			log.Printf("Loading objects from: %s", objDataDir)
			objMgr = objects.NewObjectManager(objDataDir)

			// Set base sprite index to after G1 sprites
			if g1 != nil {
				objMgr.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
			}

			if err := objMgr.LoadAllObjects(); err != nil {
				log.Printf("Failed to load objects: %v", err)
			} else {
				log.Printf("Loaded %d vehicles, %d land objects", len(objMgr.Vehicles), len(objMgr.LandObjects))
			}

			// Connect object manager to renderer for sprite access
			r.ObjMgr = objMgr
		}
	} else {
		log.Printf("Locomotion data directory not found")
	}

	// Fall back to placeholder tiles if G1 not loaded
	if r.G1 == nil {
		log.Printf("Using placeholder tiles...")
		if err := assets.GeneratePlaceholderAtlas("assets"); err != nil {
			log.Printf("Failed to generate placeholder: %v", err)
		} else {
			if at, err := render.LoadAtlasFromDir("assets/extracted"); err == nil {
				r.Atlas = at
				log.Printf("Loaded %d placeholder tiles", len(at.Images))
			}
		}
	}

	w := world.NewWorld(r)
	toolbar := ui.NewToolbar(screenWidth)
	windowMgr := ui.NewSimpleWindowManager()

	// Try to load the title scenario
	if dataDir != "" {
		titlePath := filepath.Join(dataDir, "title.dat")
		log.Printf("[Game] Loading title scenario from: %s", titlePath)
		if sc, err := scenario.LoadScenarioData(titlePath); err == nil {
			w.LoadFromScenario(sc)
			log.Printf("[Game] Loaded title scenario: %dx%d map", sc.MapWidth, sc.MapHeight)
		} else {
			log.Printf("[Game] Could not load title scenario: %v", err)
		}
	}

	// Create title sequence
	log.Println("[Game] Creating title sequence...")
	titleSeq := title.NewSequence(audioMgr)
	titleSeq.Start(64, 64) // Default map size for title animation

	// Start playing title music
	if dataDir != "" {
		musicPath := filepath.Join(dataDir, "css5.dat")
		log.Printf("[Game] Loading music from: %s", musicPath)
		if err := audioMgr.LoadAndPlayMusic(musicPath, true); err != nil {
			log.Printf("[Game] WARNING: Could not load music: %v", err)
		} else {
			log.Println("[Game] Music loaded and playing!")
		}
	}

	log.Println("[Game] Game initialization complete")
	return &Game{
		w:             w,
		r:             r,
		toolbar:       toolbar,
		windowMgr:     windowMgr,
		objMgr:        objMgr,
		audioMgr:      audioMgr,
		titleSeq:      titleSeq,
		dataDir:       dataDir,
		inTitleScreen: true,
	}
}

func (g *Game) Update() error {
	// Track mouse position
	g.mouseX, g.mouseY = ebiten.CursorPosition()
	g.toolbar.UpdateHover(g.mouseX, g.mouseY)

	// Handle window dragging
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.windowMgr.HandleDrag(g.mouseX, g.mouseY)
	}

	// Handle mouse clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Check windows first
		if !g.windowMgr.HandleClick(g.mouseX, g.mouseY, true) {
			// Then check toolbar
			if btnIdx := g.toolbar.HandleClick(g.mouseX, g.mouseY); btnIdx >= 0 {
				log.Printf("Toolbar button %d clicked: %s", btnIdx, g.toolbar.Buttons[btnIdx].Tooltip)
				g.toolbar.Buttons[btnIdx].Pressed = true
				g.handleToolbarButton(btnIdx)
			}
		}
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		for i := range g.toolbar.Buttons {
			g.toolbar.Buttons[i].Pressed = false
		}
		g.windowMgr.StopDrag()
	}

	// Handle zoom with +/- keys (scroll wheel handled in world.Update)
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		g.w.ZoomIn()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		g.w.ZoomOut()
	}

	// Advance title sequence camera animation
	if g.titleSeq != nil && g.titleSeq.IsRunning() {
		g.titleSeq.Update()
	}

	// World update handles arrow keys, mouse edge pan, and scroll zoom
	g.w.Update()
	return nil
}

func (g *Game) handleToolbarButton(idx int) {
	btn := &g.toolbar.Buttons[idx]
	tooltip := btn.Tooltip

	// Create appropriate window based on button
	var win *ui.SimpleWindow
	switch tooltip {
	case "Vehicles":
		win = ui.NewSimpleWindow("Available Vehicles", 50, 50, 400, 350)
		vehicles := g.objMgr
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			if vehicles == nil {
				ebitenutil.DebugPrintAt(screen, "No vehicles loaded", x+10, y+10)
				return
			}

			// Group by type
			trains := 0
			buses := 0
			trucks := 0
			aircraft := 0
			ships := 0
			for _, v := range vehicles.Vehicles {
				switch v.Type {
				case objects.VehicleTypeTrain:
					trains++
				case objects.VehicleTypeBus:
					buses++
				case objects.VehicleTypeTruck:
					trucks++
				case objects.VehicleTypeAircraft:
					aircraft++
				case objects.VehicleTypeShip:
					ships++
				}
			}

			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Total Vehicles: %d", len(vehicles.Vehicles)), x+10, y+10)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Trains:   %d", trains), x+10, y+30)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Buses:    %d", buses), x+10, y+45)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Trucks:   %d", trucks), x+10, y+60)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Aircraft: %d", aircraft), x+10, y+75)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Ships:    %d", ships), x+10, y+90)

			ebitenutil.DebugPrintAt(screen, "--- Sample Vehicles ---", x+10, y+115)

			// Show first 10 vehicles
			yOff := y + 130
			count := 0
			for _, v := range vehicles.Vehicles {
				if count >= 10 {
					break
				}
				name := v.DisplayName
				if name == "" {
					name = v.Header.GetName()
				}
				if len(name) > 28 {
					name = name[:28] + "..."
				}
				line := fmt.Sprintf("%-30s %3d km/h", name, int(v.GetSpeedKmh()))
				ebitenutil.DebugPrintAt(screen, line, x+10, yOff)
				yOff += 15
				count++
			}
		}
	case "Towns":
		win = ui.NewSimpleWindow("Towns", 150, 120, 280, 180)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ebitenutil.DebugPrintAt(screen, "Town list will go here", x+10, y+10)
		}
	case "Industries":
		win = ui.NewSimpleWindow("Industries", 180, 140, 320, 220)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ebitenutil.DebugPrintAt(screen, "Industry list will go here", x+10, y+10)
		}
	case "Companies":
		win = ui.NewSimpleWindow("Companies", 120, 100, 350, 250)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ebitenutil.DebugPrintAt(screen, "Company: GoLoco Transport", x+10, y+10)
			ebitenutil.DebugPrintAt(screen, "Balance: $1,000,000", x+10, y+30)
			ebitenutil.DebugPrintAt(screen, "Vehicles: 0", x+10, y+50)
		}
	case "Finances":
		win = ui.NewSimpleWindow("Finances", 140, 110, 400, 300)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ebitenutil.DebugPrintAt(screen, "Income:      $0", x+10, y+10)
			ebitenutil.DebugPrintAt(screen, "Expenses:    $0", x+10, y+30)
			ebitenutil.DebugPrintAt(screen, "Profit:      $0", x+10, y+50)
		}
	case "Map":
		win = ui.NewSimpleWindow("Map", 200, 80, 256, 256)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			drawH := h
			if drawH > 200 {
				drawH = 200
			}
			drawW := w
			if drawW > 200 {
				drawW = 200
			}
			ebitenutil.DrawRect(screen, float64(x), float64(y), float64(drawW), float64(drawH), color.RGBA{34, 139, 34, 255})
		}
	default:
		// Generic window for other buttons
		win = ui.NewSimpleWindow(tooltip, 100+idx*20, 100+idx*10, 250, 150)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ebitenutil.DebugPrintAt(screen, "Coming soon...", x+10, y+10)
		}
	}

	if win != nil {
		g.windowMgr.OpenWindow(win)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen with sky color
	screen.Fill(color.RGBA{135, 206, 235, 255})

	// Draw world
	g.r.SetScreen(screen)
	g.w.Draw(screen)

	// Draw toolbar
	g.toolbar.Draw(screen, g.r)

	// Draw windows
	g.windowMgr.Draw(screen, g.r)

	// Draw status bar at bottom
	statusY := screenHeight - 20
	ebitenutil.DrawRect(screen, 0, float64(statusY), float64(screenWidth), 20, color.RGBA{50, 50, 50, 240})

	// Status text
	statusText := fmt.Sprintf("GoLoco | Mouse: %d,%d | WASD to move | Using Locomotion sprites", g.mouseX, g.mouseY)
	ebitenutil.DebugPrintAt(screen, statusText, 4, statusY+4)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := NewGame()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("GoLoco - Locomotion Reimplementation")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
