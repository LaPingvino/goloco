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
	dropdown      *ui.DropdownMenu
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

	// Initialize fonts
	log.Println("[Game] Loading fonts...")
	_, err := ui.InitFonts("../fonts")
	if err != nil {
		log.Printf("[Game] Failed to load fonts: %v (will use fallback)", err)
	}

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

			// Set base sprite index to after G1 sprites and attach G1 for dynamic loading
			if g1 != nil {
				objMgr.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
				objMgr.G1File = g1
				log.Printf("ObjectManager will use dynamic G1 sprite pool (starting at index %d)", g1.GetSpriteCount())
			}

			// Load InterfaceSkin first (INTERDEF.DAT)
			interfaceSkinPath := filepath.Join(objDataDir, "INTERDEF.DAT")
			if _, err := os.Stat(interfaceSkinPath); err == nil {
				log.Printf("Loading InterfaceSkin from: %s", interfaceSkinPath)
				if _, err := objMgr.LoadObject(interfaceSkinPath); err != nil {
					log.Printf("Failed to load InterfaceSkin: %v", err)
				} else if objMgr.InterfaceSkin != nil {
					log.Printf("Loaded InterfaceSkin: %s with %d sprites",
						objMgr.InterfaceSkin.DisplayName,
						objMgr.InterfaceSkin.ImageCount)
				}
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
			// Reorder land objects to match the terrain-slot order declared
			// in the scenario's RequiredObjects chunk before loading tiles,
			// so that TerrainIndex values resolve to the correct LandObject.
			if objMgr != nil && len(sc.LandObjectOrder) > 0 {
				objMgr.ReorderLandObjects(sc.LandObjectOrder)
				log.Printf("[Game] Reordered land objects to match scenario slot order")
			}
			w.LoadFromScenario(sc)
			log.Printf("[Game] Loaded title scenario: %dx%d map", sc.MapWidth, sc.MapHeight)
		} else {
			log.Printf("[Game] Could not load title scenario: %v", err)
		}
	}

	// Create title sequence
	log.Println("[Game] Creating title sequence...")
	titleSeq := title.NewSequence(audioMgr)

	// Set title screen camera to map center at full zoom (matching OpenLoco)
	// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Main.cpp:45-51
	//   Camera position = (kMapRows * kTileSize) / 2 - 1 = (384 * 32) / 2 - 1 = 6143
	//   Zoom = ZoomLevel::full (0)
	mapW, mapH := w.GetMapSize()
	if mapW > 0 && mapH > 0 {
		titleSeq.Start(mapW, mapH)
		// Set camera to map center (tile coords)
		centerX := float64(mapW) / 2.0
		centerY := float64(mapH) / 2.0
		titleSeq.SetCameraPosition(centerX, centerY, 0) // zoom 0 = full zoom
		w.SetZoom(0)                                    // Set world to full zoom for title screen
	} else {
		titleSeq.Start(384, 384) // Default to standard map size
		titleSeq.SetCameraPosition(192, 192, 0)
		w.SetZoom(0)
	}

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
	g.mouseX, g.mouseY = ebiten.CursorPosition()

	if g.inTitleScreen {
		// Title screen: only handle menu clicks
		// NO camera animation, zoom, scroll, or other gameplay inputs
		// The camera is set once at init and stays fixed
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handleTitleMenuClick(g.mouseX, g.mouseY)
		}
		// Do NOT process any other input on title screen
		return nil
	}

	// --- Gameplay mode ONLY below this point ---

	// Update dropdown hover if visible
	if g.dropdown != nil && g.dropdown.Visible {
		g.dropdown.UpdateHover(g.mouseX, g.mouseY)
	} else {
		g.toolbar.UpdateHover(g.mouseX, g.mouseY)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.windowMgr.HandleDrag(g.mouseX, g.mouseY)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Check dropdown first (it's on top)
		if g.dropdown != nil && g.dropdown.Visible {
			if !g.dropdown.HandleClick(g.mouseX, g.mouseY) {
				// Click outside dropdown - close it
				g.dropdown = nil
			}
		} else if !g.windowMgr.HandleClick(g.mouseX, g.mouseY, true) {
			if btnIdx := g.toolbar.HandleClick(g.mouseX, g.mouseY); btnIdx >= 0 {
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

	// Zoom: keyboard +/- and mouse scroll wheel (only in gameplay mode)
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		g.w.ZoomIn()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		g.w.ZoomOut()
	}
	if _, dy := ebiten.Wheel(); dy != 0 {
		if dy > 0 {
			g.w.ZoomIn()
		} else {
			g.w.ZoomOut()
		}
	}

	// Only update world movement/animation in gameplay mode
	g.w.Update()
	return nil
}

// Title screen layout constants — match OpenLoco's window positions.
const (
	// Main menu buttons (TitleMenu.cpp)
	titleBtnSize    = 74
	titleBtnCount   = 4
	titleMenuW      = titleBtnSize * titleBtnCount // 296
	titleMenuH      = titleBtnSize
	titleMenuMargin = 25 // pixels from bottom

	// Exit button (TitleExit.cpp) — bottom-right
	titleExitW = 40
	titleExitH = 28

	// Options button (TitleOptions.cpp) — top-right
	titleOptionsW = 60
	titleOptionsH = 15
)

var titleButtons = [titleBtnCount]string{"New Game", "Load Game", "Tutorial", "Editor"}

func (g *Game) handleTitleMenuClick(mx, my int) {
	// --- Exit button: bottom-right corner ---
	exitX := screenWidth - titleExitW
	exitY := screenHeight - titleExitH
	if mx >= exitX && mx < screenWidth && my >= exitY && my < screenHeight {
		log.Println("[Game] Exit clicked")
		os.Exit(0)
	}

	// --- Options button: top-right corner ---
	optX := screenWidth - titleOptionsW
	if mx >= optX && mx < screenWidth && my >= 0 && my < titleOptionsH {
		log.Println("[Game] Title menu: Options (not yet implemented)")
		return
	}

	// --- Main menu buttons: bottom-centre ---
	menuX := (screenWidth - titleMenuW) / 2
	menuY := screenHeight - titleMenuH - titleMenuMargin
	relX := mx - menuX
	if relX < 0 || relX >= titleMenuW || my < menuY || my >= menuY+titleMenuH {
		return
	}
	btnIdx := relX / titleBtnSize

	switch btnIdx {
	case 0: // New Game
		if g.titleSeq != nil {
			g.titleSeq.Stop()
		}
		g.inTitleScreen = false
		log.Println("[Game] Title menu: New Game selected")
	case 1:
		log.Println("[Game] Title menu: Load Game (not yet implemented)")
	case 2:
		log.Println("[Game] Title menu: Tutorial (not yet implemented)")
	case 3:
		log.Println("[Game] Title menu: Scenario Editor (not yet implemented)")
	}
}

// handleToolbarButton handles clicks on toolbar buttons, matching the new
// OpenLoco-based button layout.
//
// Button order (indices):
//
//	0: Load/Save, 1: Audio, 2: (reserved/hidden)
//	3: Zoom, 4: Rotate, 5: View
//	6: Terraform, 7: Railroad, 8: Road, 9: Port/Airport, 10: Build Vehicles
//	11: Vehicles, 12: Stations, 13: Towns
func (g *Game) handleToolbarButton(idx int) {
	btn := &g.toolbar.Buttons[idx]
	tooltip := btn.Tooltip

	// Create appropriate dropdown or window based on button tooltip
	var win *ui.SimpleWindow
	switch tooltip {
	case "Load/Save":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTop.cpp:117-154
		items := []ui.DropdownItem{
			{Text: "Load Game", Action: func() { log.Println("[Game] Load Game selected") }},
			{Text: "Save Game", Action: func() { log.Println("[Game] Save Game selected") }},
			{Separator: true},
			{Text: "About", Action: func() { log.Println("[Game] About selected") }},
			{Text: "Options", Action: func() { log.Println("[Game] Options selected") }},
			{Separator: true},
			{Text: "Quit to Menu", Action: func() { g.inTitleScreen = true }},
			{Text: "Quit to Desktop", Action: func() { os.Exit(0) }},
		}
		g.dropdown = ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
		return

	case "Audio":
		items := []ui.DropdownItem{
			{Text: "Stop Music", Action: func() {
				if g.audioMgr != nil {
					g.audioMgr.StopMusic()
					log.Println("[Game] Music stopped")
				}
			}},
		}
		g.dropdown = ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
		return

	case "Zoom":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTopCommon.cpp:95-132
		items := []ui.DropdownItem{
			{Text: "Zoom In", Action: func() { g.w.ZoomIn() }},
			{Text: "Zoom Out", Action: func() { g.w.ZoomOut() }},
		}
		g.dropdown = ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
		return

	case "Rotate":
		// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTopCommon.cpp:135-143
		items := []ui.DropdownItem{
			{Text: "Rotate Clockwise", Action: func() { log.Println("[Game] Rotate CW") }},
			{Text: "Rotate Counter-CW", Action: func() { log.Println("[Game] Rotate CCW") }},
		}
		g.dropdown = ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
		return

	case "View":
		items := []ui.DropdownItem{
			{Text: "Underground View", Action: func() { log.Println("[Game] Underground view") }},
			{Text: "See Through Tracks", Action: func() { log.Println("[Game] See through tracks") }},
		}
		g.dropdown = ui.NewDropdownMenu(btn.X, btn.Y+btn.Height, items)
		return
	case "Terraform":
		log.Println("[Game] Terraform menu (not yet implemented)")
		return
	case "Railroad":
		log.Println("[Game] Railroad construction menu (not yet implemented)")
		return
	case "Road":
		log.Println("[Game] Road construction menu (not yet implemented)")
		return
	case "Port/Airport":
		log.Println("[Game] Port/Airport construction menu (not yet implemented)")
		return
	case "Build Vehicles":
		log.Println("[Game] Build vehicles menu (not yet implemented)")
		return
	case "Vehicles":
		win = ui.NewSimpleWindow("Available Vehicles", 50, 50, 400, 350)
		vehicles := g.objMgr
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			if vehicles == nil {
				ui.DrawText(screen, "No vehicles loaded", x+10, y+20, color.White)
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

			ui.DrawText(screen, fmt.Sprintf("Total Vehicles: %d", len(vehicles.Vehicles)), x+10, y+20, color.White)
			ui.DrawText(screen, fmt.Sprintf("  Trains:   %d", trains), x+10, y+40, color.White)
			ui.DrawText(screen, fmt.Sprintf("  Buses:    %d", buses), x+10, y+55, color.White)
			ui.DrawText(screen, fmt.Sprintf("  Trucks:   %d", trucks), x+10, y+70, color.White)
			ui.DrawText(screen, fmt.Sprintf("  Aircraft: %d", aircraft), x+10, y+85, color.White)
			ui.DrawText(screen, fmt.Sprintf("  Ships:    %d", ships), x+10, y+100, color.White)

			ui.DrawText(screen, "--- Sample Vehicles ---", x+10, y+125, color.White)

			// Show first 10 vehicles
			yOff := y + 140
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
				ui.DrawText(screen, line, x+10, yOff, color.White)
				yOff += 15
				count++
			}
		}
	case "Stations":
		win = ui.NewSimpleWindow("Stations", 150, 120, 350, 250)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ui.DrawText(screen, "Station list will go here", x+10, y+20, color.White)
		}
	case "Towns":
		win = ui.NewSimpleWindow("Towns", 150, 120, 280, 180)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ui.DrawText(screen, "Town list will go here", x+10, y+20, color.White)
		}
	default:
		// Generic window for any unhandled buttons
		win = ui.NewSimpleWindow(tooltip, 100+idx*20, 100+idx*10, 250, 150)
		win.DrawContent = func(screen *ebiten.Image, x, y, w, h int, r *render.Renderer) {
			ui.DrawText(screen, "Coming soon...", x+10, y+20, color.White)
		}
	}

	if win != nil {
		g.windowMgr.OpenWindow(win)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{135, 206, 235, 255})

	g.r.SetScreen(screen)
	g.w.Draw(screen)

	if g.inTitleScreen {
		g.drawTitleMenu(screen)
		return
	}

	// --- Gameplay HUD ---
	g.toolbar.Draw(screen, g.r)
	g.windowMgr.Draw(screen, g.r)

	// Draw dropdown menu on top of everything else
	if g.dropdown != nil && g.dropdown.Visible {
		g.dropdown.Draw(screen)
	}

	statusY := screenHeight - 20
	ebitenutil.DrawRect(screen, 0, float64(statusY), float64(screenWidth), 20, color.RGBA{50, 50, 50, 240})
	statusText := fmt.Sprintf("GoLoco | Mouse: %d,%d | Scroll to zoom", g.mouseX, g.mouseY)
	ui.DrawText(screen, statusText, 4, statusY+14, color.White)
}

func (g *Game) drawTitleMenu(screen *ebiten.Image) {
	// --- Logo: top-left (0,0) 298×170 ---
	// Display "GoLoco" branding using OpenTTD font
	ebitenutil.DrawRect(screen, 0, 0, 298, 170, color.RGBA{30, 30, 60, 200})
	// Draw large "GoLoco" text centered
	logoText := "GoLoco"
	logoW, _ := ui.MeasureTextBold(logoText)
	logoX := (298 - logoW) / 2
	ui.DrawTextBold(screen, logoText, logoX, 60, color.RGBA{220, 220, 255, 255})
	// Subtitle
	subtitle := "A Locomotion Reimplementation"
	subW, _ := ui.MeasureText(subtitle)
	subX := (298 - subW) / 2
	ui.DrawText(screen, subtitle, subX, 100, color.RGBA{180, 180, 200, 255})

	// --- Options button: top-right (screenWidth-60, 0) 60×15 ---
	optX := screenWidth - titleOptionsW
	optHovered := g.mouseX >= optX && g.mouseX < screenWidth &&
		g.mouseY >= 0 && g.mouseY < titleOptionsH
	optBg := color.RGBA{60, 60, 100, 200}
	if optHovered {
		optBg = color.RGBA{80, 80, 130, 220}
	}
	ebitenutil.DrawRect(screen, float64(optX), 0, float64(titleOptionsW), float64(titleOptionsH), optBg)
	ebitenutil.DrawRect(screen, float64(optX), 0, float64(titleOptionsW), 1, color.RGBA{120, 120, 180, 255})
	ebitenutil.DrawRect(screen, float64(optX), 0, 1, float64(titleOptionsH), color.RGBA{120, 120, 180, 255})
	ebitenutil.DrawRect(screen, float64(optX), float64(titleOptionsH-1), float64(titleOptionsW), 1, color.RGBA{30, 30, 50, 255})
	ebitenutil.DrawRect(screen, float64(screenWidth-1), 0, 1, float64(titleOptionsH), color.RGBA{30, 30, 50, 255})
	optLabel := "Options"
	optLabelW, _ := ui.MeasureText(optLabel)
	optLabelX := optX + (titleOptionsW-optLabelW)/2
	ui.DrawText(screen, optLabel, optLabelX, 12, color.White)

	// --- Main menu buttons: bottom-centre ---
	// OpenLoco reference: src/OpenLoco/src/Graphics/ImageIds.h
	//   title_menu_globe_spin_0 = 3552 (idle sprite for New/Load/Tutorial)
	//   title_menu_globe_construct_24 = 3608 (idle sprite for Editor)
	const titleMenuGlobeSpin0 = 3552
	const titleMenuGlobeConstruct24 = 3608

	menuX := (screenWidth - titleMenuW) / 2
	menuY := screenHeight - titleMenuH - titleMenuMargin

	for i := 0; i < titleBtnCount; i++ {
		bx := menuX + i*titleBtnSize
		by := menuY

		hovered := g.mouseX >= bx && g.mouseX < bx+titleBtnSize &&
			g.mouseY >= by && g.mouseY < by+titleBtnSize

		// Choose sprite based on button type
		spriteID := titleMenuGlobeSpin0
		if i == 3 { // Editor button uses construct globe
			spriteID = titleMenuGlobeConstruct24
		}

		// Try to draw globe sprite
		globeSprite := g.r.GetSprite(spriteID)
		if globeSprite != nil {
			// Center the sprite in the button
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(bx), float64(by))
			screen.DrawImage(globeSprite, opts)
		} else {
			// Fallback: draw colored rectangle with border
			var bgColor color.RGBA
			if hovered {
				bgColor = color.RGBA{80, 120, 80, 220}
			} else {
				bgColor = color.RGBA{50, 90, 50, 200}
			}
			ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(titleBtnSize), float64(titleBtnSize), bgColor)

			light := color.RGBA{100, 160, 100, 255}
			dark := color.RGBA{30, 60, 30, 255}
			ebitenutil.DrawRect(screen, float64(bx), float64(by), float64(titleBtnSize), 2, light)
			ebitenutil.DrawRect(screen, float64(bx), float64(by), 2, float64(titleBtnSize), light)
			ebitenutil.DrawRect(screen, float64(bx), float64(by+titleBtnSize-2), float64(titleBtnSize), 2, dark)
			ebitenutil.DrawRect(screen, float64(bx+titleBtnSize-2), float64(by), 2, float64(titleBtnSize), dark)

			label := titleButtons[i]
			labelW, _ := ui.MeasureText(label)
			labelX := bx + (titleBtnSize-labelW)/2
			labelY := by + titleBtnSize/2 + 4
			ui.DrawText(screen, label, labelX, labelY, color.White)
		}
	}

	// --- Exit button: bottom-right (screenWidth-40, screenHeight-28) 40×28 ---
	exitX := screenWidth - titleExitW
	exitY := screenHeight - titleExitH
	exitHovered := g.mouseX >= exitX && g.mouseX < screenWidth &&
		g.mouseY >= exitY && g.mouseY < screenHeight
	exitBg := color.RGBA{100, 40, 40, 200}
	if exitHovered {
		exitBg = color.RGBA{140, 60, 60, 220}
	}
	ebitenutil.DrawRect(screen, float64(exitX), float64(exitY), float64(titleExitW), float64(titleExitH), exitBg)
	ebitenutil.DrawRect(screen, float64(exitX), float64(exitY), float64(titleExitW), 2, color.RGBA{180, 100, 100, 255})
	ebitenutil.DrawRect(screen, float64(exitX), float64(exitY), 2, float64(titleExitH), color.RGBA{180, 100, 100, 255})
	ebitenutil.DrawRect(screen, float64(exitX), float64(screenHeight-2), float64(titleExitW), 2, color.RGBA{60, 20, 20, 255})
	ebitenutil.DrawRect(screen, float64(screenWidth-2), float64(exitY), 2, float64(titleExitH), color.RGBA{60, 20, 20, 255})
	exitLabel := "Exit"
	exitLabelW, _ := ui.MeasureText(exitLabel)
	exitLabelX := exitX + (titleExitW-exitLabelW)/2
	exitLabelY := exitY + (titleExitH)/2 + 4
	ui.DrawText(screen, exitLabel, exitLabelX, exitLabelY, color.White)

	// --- Version text: bottom-left (8, screenHeight-30) ---
	ui.DrawText(screen, "GoLoco v0.1", 8, screenHeight-20, color.RGBA{200, 200, 200, 255})
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
