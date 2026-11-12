package main

import (
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// diag.go — diagnostics for headless checks: the --diag / F12 crop and the
// GOLOCO_DIAG_* panels.

const diagCropW = 480

const diagCropH = 360

// diagBackground draws a black+red cross-hatch pattern onto img so that
// transparent/unrendered areas are visually distinct from intentional black pixels.
func diagBackground(img *ebiten.Image) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	pixels := make([]byte, w*h*4)
	const stride = 16 // grid cell size
	red := [4]byte{180, 0, 0, 255}
	dark := [4]byte{20, 20, 20, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			onLine := (x%stride == 0) || (y%stride == 0)
			var c [4]byte
			if onLine {
				c = red
			} else {
				c = dark
			}
			i := (y*w + x) * 4
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = c[0], c[1], c[2], c[3]
		}
	}
	img.WritePixels(pixels)
}

// saveDiagCrop saves a 480×360 PNG cropped from the centre of the screen
// (or centred on tile diagTileX/diagTileY when the world is loaded).
// The file is always named "diag.png" so Claude can read it without guessing the name.
func saveDiagCrop(screen *ebiten.Image, g *Game, tileX, tileY int) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Centre pixel: use tile position when valid, otherwise screen centre.
	cx, cy := sw/2, sh/2
	if tileX >= 0 && tileY >= 0 {
		cx, cy = g.w.TileToScreenPx(tileX, tileY)
	}

	x0 := cx - diagCropW/2
	y0 := cy - diagCropH/2
	x0 = max(0, min(x0, sw-diagCropW))
	y0 = max(0, min(y0, sh-diagCropH))

	crop := screen.SubImage(image.Rect(x0, y0, x0+diagCropW, y0+diagCropH))

	if err := savePNG(debugPath("diag.png"), crop); err != nil {
		log.Printf("[Diag] %v", err)
		return
	}
	log.Printf("[Diag] Saved diag.png (%dx%d crop centred on tile %d,%d, screen offset %d,%d)",
		diagCropW, diagCropH, tileX, tileY, x0, y0)
}

// drawInlineDiag (GOLOCO_DIAG_INLINE) is drawVehicleInline for the first ten
// rail and road vehicle types, in a panel at (20, 60).
func (g *Game) drawInlineDiag(screen *ebiten.Image) {
	if os.Getenv("GOLOCO_DIAG_TTF") != "" { // the G1 font and the TTF fallback on one line
		fillRect(screen, 10, 10, 300, 30, color.RGBA{200, 200, 200, 255})
		fillRect(screen, 10, 20, 300, 1, color.RGBA{255, 0, 0, 255}) // y = 20: the text tops
		ui.DrawTextBold(screen, "Bitmap text", 20, 20, color.Black)
		ui.DrawTextBold(screen, "TTF 日本語 text", 120, 20, color.Black)
	}
	if os.Getenv("GOLOCO_DIAG_INLINE") == "" || g.r.ObjMgr == nil {
		return
	}
	var vs []*objects.VehicleObject
	var offs []int
	for _, v := range g.r.ObjMgr.Vehicles {
		if v != nil && v.ImageOffset != 0 && (v.Mode == objects.TransportModeRail || v.Mode == objects.TransportModeRoad) && len(vs) < 10 {
			vs = append(vs, v)
			offs = append(offs, g.r.ObjMgr.VehicleYDisplayOffset(v))
		}
	}
	log.Printf("[diag] inline: %d vehicle types", len(vs))
	ui.DrawVehiclesInlineDiag(screen, 20, 60, vs, offs)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// diagOpenWindows runs the GOLOCO_DIAG_* switches that open a window or set a
// situation up, each once, after a frame is drawn (the next frame shows it).
func (g *Game) diagOpenWindows() {
	if os.Getenv("GOLOCO_DIAG_TOWNS") != "" && g.townList == nil && !g.diagTownsDone {
		g.diagTownsDone = true
		if d := os.Getenv("GOLOCO_DIAG_TOWNS"); d == "stations" {
			g.openStationsWindow()
		} else if d == "vehicles" {
			g.openVehiclesWindow()
		} else {
			g.openTownsWindow()
		}
		g.townList.Hover = 2
		if d := os.Getenv("GOLOCO_DIAG_TOWNS"); d == "scrolled" { // a click on the track below the thumb
			l := g.townList.List
			g.townList.HandleClick(g.townList.X+l[0]+l[2]-6, g.townList.Y+l[1]+l[3]-20)
		}
	}
	if d := os.Getenv("GOLOCO_DIAG_CONSTRUCTION"); (d == "window" || d == "road" || d == "extend" || d == "signal" || d == "overhead" || d == "station" || d == "sigdrop" || d == "bridgedrop") && !g.diagConsDone {
		g.diagConsDone = true
		if d == "signal" {
			g.consTab = 2
		}
		if d == "station" {
			g.consTab = 1
		}
		if d == "bridgedrop" { // the bridge list open
			defer func() {
				x, y := g.consWindowOrigin()
				f, _ := g.constructionFrame(x, y)
				g.openBridgeDropdown(x, y, f.SkinColour, f.CompanyColour)
			}()
		}
		if d == "sigdrop" { // the signal tab with its dropdown open
			g.consTab = 2
			defer func() {
				x, y := g.consWindowOrigin()
				g.openConsDropdown(x, y, g.r.ObjMgr.InterfaceSkin.WindowConstructionColour, false)
			}()
		}
		if d == "overhead" { // the tram road (its overhead wires), one ticked
			g.buildMode = buildModeRoad
			for range 8 {
				if st := g.overheadTabState(); st.Names[0] != "" {
					break
				}
				g.consCycleType(1)
			}
			g.consTab, g.consMods = 3, 1
		}
		if d == "extend" {
			g.buildMode = buildModeTrack
			mw, mh := g.w.GetMapSize()
			g.w.SetConstructionHead(mw/2, mh/2)
		}
		if d == "road" {
			g.buildMode = buildModeRoad
			if names := g.consObjectNames(); len(names) > 0 && names[0] == "" {
				g.consCycleType(1)
			}
		}
		g.windowMgr.OpenWindow(g.newPortedConstructionWindow())
	}
	if v := os.Getenv("GOLOCO_DIAG_BUILDVEHICLE"); v != "" && !g.buildVehDiagDone && (!diagNeedsScene[v] || g.w.SceneReady()) {
		g.buildVehDiagDone = true
		g.diagScenario(v)
	}
	if os.Getenv("GOLOCO_DIAG_VEHLIST") != "" && !g.vehDiagDone { // the vehicle list (trains of their first owner)
		g.vehDiagDone = true
		g.openVehicleList(0)
		if os.Getenv("GOLOCO_DIAG_VEHLIST") == "cargo" && g.townList != nil { // transporting cargo, the cargo dropdown open
			w := g.townList
			w.HandleClick(w.X+270, w.Y+205)
			if g.dropdown != nil && g.dropdown.OnSelect != nil {
				g.dropdown.OnSelect(1)
				g.dropdown = nil
			}
			w.HandleClick(w.X+400, w.Y+205)
		}
		if os.Getenv("GOLOCO_DIAG_VEHLIST") == "company" && g.townList != nil { // the company select dropdown open
			w := g.townList
			w.HandleClick(w.X+w.W-20, w.Y+20)
		}
	}
	if os.Getenv("GOLOCO_DIAG_OPTIONS") != "" && g.options == nil && !g.diagOptionsDone {
		g.diagOptionsDone = true
		g.openOptionsWindow()
		o := g.options
		switch os.Getenv("GOLOCO_DIAG_OPTIONS") {
		case "controls":
			o.Tab = ui.OptTabControls
			o.Controls = [3]bool{true, false, true}
		case "regional":
			o.Tab = ui.OptTabRegional
		case "framesolid":
			ui.FrameStyle = 1 // the Rendering tab with solid frames
			o.Tab = ui.OptTabRendering
		case "prompt": // the replace-file confirmation
			g.options = nil
			g.prompt = ui.NewPromptOkCancel("Save Game", "Replace existing file: “Bottleneck Blues”?", "Replace", g.sw, g.sh, nil)
		case "renderscale":
			o.Tab = ui.OptTabRendering // and a click on Vehicles min. display scale
			o.HandleClick(o.X+240, o.Y+66)
		case "framelimit":
			o.Tab = ui.OptTabDisplay // and a click on Frame rate limit
			o.HandleClick(o.X+240, o.Y+114)
		case "displayres":
			o.Tab = ui.OptTabDisplay // in fullscreen mode, a click on Display Resolution
			o.ScreenMode = 1
			o.HandleClick(o.X+240, o.Y+82)
		case "displaymode":
			o.Tab = ui.OptTabDisplay // and a click on Screen mode
			o.HandleClick(o.X+240, o.Y+66)
		case "regionalcurrency":
			o.Tab = ui.OptTabRegional // and a click on Preferred currency
			o.HandleClick(o.X+190, o.Y+122)
		case "regionalcurrent":
			o.Tab = ui.OptTabRegional // and a click on Current currency (as in a game)
			o.TitleMode = false
			o.HandleClick(o.X+190, o.Y+107)
		case "regionalalways":
			o.Tab = ui.OptTabRegional // in a game, a click on 'always' (the current currency greys)
			o.TitleMode = false
			o.HandleClick(o.X+100, o.Y+153)
		case "regionallang":
			o.Tab = ui.OptTabRegional // and a click on Language
			o.HandleClick(o.X+190, o.Y+52)
		case "regionalheights":
			o.Tab = ui.OptTabRegional // and a click on Heights
			o.HandleClick(o.X+190, o.Y+87)
		case "regionaldrop":
			o.Tab = ui.OptTabRegional // and a click on Distance and Speed
			o.HandleClick(o.X+190, o.Y+72)
		case "display":
			o.Tab = ui.OptTabDisplay
		case "company":
			o.Tab = ui.OptTabCompany
		case "rendering":
			o.Tab = ui.OptTabRendering
		case "misc":
			o.Tab = ui.OptTabMisc
		}
	}
	if os.Getenv("GOLOCO_DIAG_BROWSE") != "" && g.browse == nil && !g.diagBrowseDone {
		g.diagBrowseDone = true
		if os.Getenv("GOLOCO_DIAG_BROWSE") == "preview" { // a synthetic saved-game preview on the first file
			g.openLoadGameWindow()
			if g.browse != nil {
				img := make([]byte, 250*200)
				for i := range img {
					img[i] = byte(10 + (i%250)/25*10 + (i/250)/50) // bands of palette colours
				}
				g.browse.Preview = &scenario.SaveDetails{Company: "Joop's Railways", Owner: "Joop", Date: 55000, ChallengeProgress: 42, Image: img}
				g.browse.ChangeDir(filepath.Join(g.browse.Dir, "Scenarios"))
				g.browse.SelectFirstFile()
			}
		} else if os.Getenv("GOLOCO_DIAG_BROWSE") == "save" {
			g.openSaveWindow()
		} else {
			g.openFileWindow("Load Game", false)
		}
	}
	if os.Getenv("GOLOCO_DIAG_NEWGAME") != "" && g.scenarioSelect == nil && !g.diagNewGameDone {
		// Diagnostic: open the ScenarioSelect window with its first scenario selected.
		g.diagNewGameDone = true
		g.openNewGameWindow()
		g.scenarioSelect.SelectFirst()
		if os.Getenv("GOLOCO_DIAG_NEWGAME") == "many" {
			g.scenarioSelect.DiagRepeat(3)
			g.scenarioSelect.Tab = 3
			g.scenarioSelect.HandleWheel(-4)
		}
	}
}

// diagDrawPanels draws the GOLOCO_DIAG_* panels over the finished frame, every
// frame, for headless shots of single widgets.
func (g *Game) diagDrawPanels(screen *ebiten.Image) {
	if os.Getenv("GOLOCO_DIAG_TIMEPANEL") != "" {
		st := ui.TimePanelState{Date: uilayout.DateDMY(1, "January", 1950), Speed: 1}
		if sk := g.skin(); sk != nil {
			st.SkinColour = sk.TimeToolbarColour
			st.SkinImage = int(sk.ImageOffset)
		}
		ui.DrawTimePanel(screen, st)
		ps := ui.PlayerPanelState{Cash: 2500000}
		if sk := g.skin(); sk != nil {
			ps.SkinColour = sk.PlayerInfoToolbarColour
		}
		g.playerPanelCompany(&ps)
		ui.DrawPlayerPanel(screen, ps)
	}
	if os.Getenv("GOLOCO_DIAG_WINDOW") != "" {
		// Diagnostic: an upstream-style window frame + caption (FrameWidget,
		// CaptionWidget ports) for headless shots; the colour is the skin's window colour.
		var c uint8 = 1
		if sk := g.skin(); sk != nil {
			c = sk.WindowColour
		}
		ui.DrawWindowFrame(screen, 200, 60, 360, 220, c, false)
		ui.DrawWindowCaption(screen, 201, 61, 358, "Select Scenario")
	}
	if os.Getenv("GOLOCO_DIAG_CONSTRUCTION") == "frame" && g.r.ObjMgr != nil {
		// Diagnostic: the ported construction frame for the first track object.
		f := &ui.ConstructionFrame{X: 10, Y: 40, Title: "Railway", FrameNo: g.titleFrame}
		if sk := g.r.ObjMgr.InterfaceSkin; sk != nil {
			f.SkinColour = sk.WindowConstructionColour
		}
		if len(g.w.CompanyColours) > 0 {
			f.CompanyColour = g.w.CompanyColours[0]
		}
		if len(g.r.ObjMgr.TrackObjects) > 0 && g.r.ObjMgr.TrackObjects[0] != nil {
			f.TrackImage = int(g.r.ObjMgr.TrackObjects[0].Image)
			f.Title = g.r.ObjMgr.TrackObjects[0].Name
		}
		f.Draw(screen)
		if len(g.r.ObjMgr.TrackObjects) > 0 && g.r.ObjMgr.TrackObjects[0] != nil {
			ui.DrawConstructionTab(screen, f, ui.ConstructionTabState{Traits: g.r.ObjMgr.TrackObjects[0].TrackPieces,
				Piece: "straight", Gradient: "level"})
		}
		f.DrawCaption(screen)
	}
	if os.Getenv("GOLOCO_DIAG_PROGRESS") != "" {
		// Diagnostic: the ProgressBar window at 3/4 in both styles, for headless shots.
		ui.DrawProgressBar(screen, "Loading...", 190, 0)
	}
	if os.Getenv("GOLOCO_DIAG_TOOLTIP") != "" {
		// Diagnostic: a tooltip over the frame (ToolTip.cpp port), for headless shots.
		var c uint8
		if sk := g.skin(); sk != nil {
			c = sk.TooltipColour
		}
		ui.DrawTooltip(screen, "Change viewport settings", 380, 120, c)
		ui.DrawTooltip(screen, "Build tracks, stations, signals and train depots, and place them on the map", 560, 240, c)
	}
	if os.Getenv("GOLOCO_DIAG_DROPDOWN") != "" {
		// Diagnostic: a toolbar-style dropdown on every frame (drawn last, just
		// before the capture), for headless screenshots of the Dropdown.cpp port.
		d := ui.NewDropdownMenu(40, 28, []ui.DropdownItem{{Text: "Load Game"}, {Text: "Save Game"}, {Separator: true},
			{Text: "Screenshot"}, {Separator: true}, {Text: "About", Disabled: true}, {Text: "Options"}, {Separator: true},
			{Text: "Quit to Menu"}, {Text: "Quit to Desktop"}})
		d.Highlighted = 1
		if sk := g.skin(); sk != nil {
			d.SetToolbarColour(sk.TopToolbarPrimaryColour)
		}
		d.Draw(screen)
	}
}
