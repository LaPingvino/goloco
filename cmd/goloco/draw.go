package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/LaPingvino/goloco/pkg/world"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// draw.go — Game.Draw: the world, then the title screen or the game's windows
// in stacking order, the bottom panels (PlayerInfoPanel, TimePanel), overlays
// and the FPS counter.

func (g *Game) Draw(screen *ebiten.Image) {
	if g.shotsEvery > 0 && g.shotsLeft > 0 && time.Since(g.shotsLast) >= g.shotsEvery && os.Getenv("GOLOCO_SHOT_PAINTDUMP") != "" {
		paint.Dump = &[]string{} // diagnostic: this frame's paint structs, saved with the shot
	}
	defer g.drawCursor(screen)    // a bitmap cursor (SDL's own upstream): after the shots, never in them
	defer g.maybeSaveShot(screen) // captures (--shots) of the finished frame: after the FPS counter
	defer g.drawFPS(screen)       // SoftwareDrawingEngine: the FPS counter over everything, last
	defer g.drawTooltip(screen)   // the tooltip window, over the other windows
	defer g.drawMapTip(screen)    // the map tooltip (stickToFront), under the widget tooltip
	if !g.firstFrameLogged {
		g.firstFrameLogged = true
		log.Printf("[Game] First world frame %v after start", time.Since(processStart).Round(time.Millisecond))
	}
	if fstats.on {
		t0 := time.Now()
		defer func() { fstats.addDraw(t0, time.Since(t0)) }() // runs last: the prewarm is included
	}
	defer g.prewarmTick()

	if g.diagPending {
		diagBackground(screen)
	} else {
		screen.Fill(color.RGBA{0, 0, 0, 255})
	}

	if g.drawMapSave(screen) {
		return
	}
	g.drawWorld(screen)
	if g.inTitleScreen {
		g.drawTitleScreen(screen)
		return
	}
	g.drawGameWindows(screen)
	g.drawOverlays(screen)
}

// drawMapSave is the full-map PNG save: while one runs it takes the frame —
// its progress, and one more chunk rendered — and reports that it did.
func (g *Game) drawMapSave(screen *ebiten.Image) bool {
	done, total := g.w.MapSaveProgress()
	if total <= 0 {
		return false
	}
	pct := float64(done) / float64(total)
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	fillRect(screen, 0, 0, float64(sw), float64(sh), color.RGBA{20, 20, 30, 220})
	barX, barY, barW, barH := float64(sw/2-200), float64(sh/2-12), 400.0, 24.0
	fillRect(screen, barX, barY, barW, barH, color.RGBA{60, 60, 80, 255})
	fillRect(screen, barX, barY, barW*pct, barH, color.RGBA{80, 180, 80, 255})
	label := fmt.Sprintf("Saving full map... %d/%d (%.0f%%)", done, total, pct*100)
	ui.DrawText(screen, label, sw/2-100, sh/2-8, color.RGBA{255, 255, 255, 255})
	if g.w.StepMapSave() {
		if err := g.w.FinishMapSave(); err != nil {
			log.Printf("FinishMapSave: %v", err)
		}
		g.quitAfterSave = true
	}
	return true
}

// drawWorld draws the map view, and saves the diagnostic crop when one is
// pending. The construction ghost is part of it: a ghost element on the map
// (sim's ConstructionToolUpdate / UpdateConstruction), painted with the world.
func (g *Game) drawWorld(screen *ebiten.Image) {
	// If a diagnostic crop is pending for a specific tile, move the camera
	// here (after Update has run) so the title-sequence camera override
	// doesn't undo the positioning we set earlier.
	if g.diagPending && g.diagTileX >= 0 {
		g.w.SetCamera(float64(g.diagTileX), float64(g.diagTileY))
	}

	g.r.SetScreen(screen)
	ui.InvalidateFrame8() // the world (drawScene) starts draw8's screen copy
	g.w.SceneEvery = 1
	if g.uiOpen() {
		g.w.SceneEvery = 3 // the world at a third of the rate, the menus at full (LOOP item 0)
	}
	tw := time.Now()
	g.w.Draw(screen)
	fstats.addWorld(time.Since(tw))
	fstats.parts = g.w.SceneParts

	// Diagnostic crop: save a small PNG centred on the requested tile.
	if g.diagPending {
		g.diagPending = false
		saveDiagCrop(screen, g, g.diagTileX, g.diagTileY)
		world.LogBuildingStats()
		if g.diagQuit {
			g.quitAfterSave = true
		}
	}
}

// drawTitleScreen draws the title menu and the windows that open over it.
func (g *Game) drawTitleScreen(screen *ebiten.Image) {
	g.drawTitleMenu(screen)
	g.drawWindows(screen)                        // the windows opened from the menu
	if g.dropdown != nil && g.dropdown.Visible { // a window's dropdown, on top
		g.dropdown.Draw(screen)
	}
	if g.prompt != nil { // a confirmation, on top of everything
		g.prompt.Draw(screen)
	}
	// A title button's tooltip (ToolTip.cpp), when no window is over it.
	// PARTIAL: shown at once, as goloco's other tooltips (upstream waits).
	if g.titleTip != "" && len(g.wins.Windows) == 0 && g.prompt == nil &&
		(g.dropdown == nil || !g.dropdown.Visible) {
		var c uint8
		if sk := g.skin(); sk != nil {
			c = sk.TooltipColour
		}
		ui.DrawTooltip(screen, g.titleTip, g.mouseX, g.mouseY, c)
	}
	g.drawInlineDiag(screen)
}

// drawGameWindows draws the game's windows — the toolbar and the panels
// among them —, back to front (the stack's order).
func (g *Game) drawGameWindows(screen *ebiten.Image) { g.drawWindows(screen) }

// drawToolbar is ToolbarTop's draw (its window on the stack).
func (g *Game) drawToolbar(screen *ebiten.Image) {
	m := &g.w.Map // the button faces: the last track and road options, in the player's colour
	ctrl, _ := g.w.ControllingCompany()
	g.toolbar.LastRailroadOption, g.toolbar.LastRoadOption = m.LastRailroadOption, m.LastRoadOption
	g.toolbar.LastVehicleType, g.toolbar.LastBuildVehiclesOption = m.LastVehicleType, m.LastBuildVehiclesOption
	g.toolbar.AudioMuted = !g.isAudioEnabled()
	g.toolbar.CompanyColour = g.companyColours(ctrl).Main[0]
	if sk := g.skin(); sk != nil { // ToolbarTop open: the window's colours
		g.toolbar.Colours = [4]uint8{sk.TopToolbarPrimaryColour, sk.TopToolbarSecondaryColour, sk.TopToolbarTertiaryColour, sk.TopToolbarQuaternaryColour}
	}
	g.toolbar.Draw(screen, g.r) // through draw8: the 8-bit screen copy stays in step
}

// drawTopmost draws what opens over the panels and the news (stickToFront
// windows made later than they are): the menus, the text input, a
// confirmation and the error window — before, the panels and the news window
// were drawn over them.
func (g *Game) drawTopmost(screen *ebiten.Image) {
	if g.dropdown != nil && g.dropdown.Visible {
		g.dropdown.Draw(screen)
	}
	if g.colourDropdown != nil && g.colourDropdown.Visible {
		g.colourDropdown.Draw(screen)
	}
	if g.prompt != nil { // a confirmation, on top of everything
		g.prompt.Draw(screen)
	}
	if g.savePrompt != nil { // PromptSaveWindow: stickToFront
		g.savePrompt.Draw(screen)
	}
	if g.errorWin != nil { // Error.cpp: stickToFront
		if g.errorWin.Expired() {
			g.errorWin = nil
		} else {
			g.errorWin.Draw(screen)
		}
	}
}

// drawOverlays draws what lies over the game's windows.
func (g *Game) drawOverlays(screen *ebiten.Image) {
	g.drawTopmost(screen)

	g.drawInlineDiag(screen)
}

// drawPlayerPanel draws the ported PlayerInfoPanel (bottom-left).
func (g *Game) drawPlayerPanel(screen *ebiten.Image) {
	if !g.inGame {
		return
	}
	var st ui.PlayerPanelState
	// The controlling company's own cash (the save, kept by the ported sim).
	if cash, ok := g.w.ControllingCompanyCash(); ok {
		st.Cash = cash
	}
	if sk := g.skin(); sk != nil {
		st.SkinColour = sk.PlayerInfoToolbarColour
	}
	g.playerPanelCompany(&st)
	wx, wy := uilayout.PlayerPanelPos(g.sw, g.sh)
	cv, pf := uilayout.PlayerPanelCompanyVal, uilayout.PlayerPanelPerformance
	under := g.barUnderMouse(ui.WindowTypeCompanyInfoPanel) // the hover: the window under the mouse only
	st.HoverCash = under && inRect(g.mouseX, g.mouseY, wx+cv[0], wy+cv[1], cv[2], cv[3])
	st.HoverPerf = under && inRect(g.mouseX, g.mouseY, wx+pf[0], wy+pf[1], pf[2], pf[3])
	ui.DrawPlayerPanel(screen, st)
}

// playerPanelCompany fills the player panel's company part from the
// controlling company: the owner's image (companyFace), performanceIndex and
// the increased/decreasedPerformance flags.
func (g *Game) playerPanelCompany(st *ui.PlayerPanelState) {
	id, _ := g.w.ControllingCompany()
	if id >= len(g.w.CompanyFaces) {
		return
	}
	st.Face, st.FaceColour = g.companyFace(id)
	f := g.w.CompanyFaces[id]
	st.Performance = f.Performance
	switch {
	case f.ChallengeFlags&(1<<4) != 0:
		st.Trend = 1
	case f.ChallengeFlags&(1<<5) != 0:
		st.Trend = -1
	}
}

// drawTimePanel draws the ported TimePanel (bottom-right) and, when its date
// is hovered, the date tooltip with the scenario objective.
func (g *Game) drawTimePanel(screen *ebiten.Image) {
	if !g.inGame {
		return
	}
	mp := &g.w.Map // getCurrentDay's date (pkg/sim dateTick)
	st := ui.TimePanelState{Date: uilayout.DateDMY(int(mp.CurrentDayOfMonth)+1, i18n.Month(int(mp.CurrentMonth)+1), int(mp.CurrentYear)), Rotation: g.w.Rotation()}
	// TimePanel::draw: paused, "* Paused *" instead of the date half the time
	// (numTicksVisible: 30 window updates each). PARTIAL: the browse prompt's
	// own pause (goloco's prompt doesn't pause).
	if g.isPaused && (g.windowFrame/30)%2 == 1 {
		st.Date = i18n.Label(i18n.ToolbarStatusPaused)
	}
	switch {
	case g.isPaused:
		st.Speed = 0
	case g.speedMult <= 1:
		st.Speed = 1
	case g.speedMult == 2:
		st.Speed = 2
	default:
		st.Speed = 3
	}
	if sk := g.skin(); sk != nil {
		st.SkinColour = sk.TimeToolbarColour
		st.SkinImage = int(sk.ImageOffset)
	}
	st.Hover = -1
	if g.barUnderMouse(ui.WindowTypeTimePanel) { // Input::isHovering(timePanel, 0, widget)
		st.Hover = uilayout.TimePanelWidgetAt(g.mouseX, g.mouseY, g.sw, g.sh)
	}
	ui.DrawTimePanel(screen, st)
}

// timePanelTooltip is TimePanel's tooltip text for a widget: the speed
// buttons' tooltip_speed_*, and the date's tooltip_daymonthyear_challenge
// with formatChallenge's arguments (TimePanel.cpp 0x0043995C): getCurrentDay,
// then challenge_completed / challenge_failed / empty (beaten by an
// opponent) / challenge_progress with challenge_time_left (or empty).
func (g *Game) timePanelTooltip(widget int) string {
	if sp := widget - uilayout.TimePanelWidxPause; sp >= 0 && sp < 4 {
		return i18n.Str(i18n.TooltipSpeedPause + sp)
	}
	if widget != uilayout.TimePanelWidxDate {
		return ""
	}
	mp := &g.w.Map
	date := uilayout.DateDMY(int(mp.CurrentDayOfMonth)+1, i18n.Month(int(mp.CurrentMonth)+1), int(mp.CurrentYear))
	var status string
	completed, failed, beaten, progress := mp.ChallengeStatus()
	switch {
	case completed:
		status = i18n.Str(i18n.ChallengeCompleted)
	case failed:
		status = i18n.Str(i18n.ChallengeFailed)
	case beaten:
	default:
		var left string
		if mp.Objective.Flags&(1<<2) != 0 { // withinTimeLimit
			months := uint16(mp.Objective.TimeLimitYears)*12 - mp.ObjectiveMonthsInChallenge
			left = i18n.Str(i18n.ChallengeTimeLeft)
			left = strings.Replace(left, "{INT16}", strconv.Itoa(int(months/12)), 1)
			left = strings.Replace(left, "{INT16}", strconv.Itoa(int(months%12)), 1)
		}
		status = strings.Replace(i18n.Str(i18n.ChallengeProgress), "{INT16}", strconv.Itoa(int(progress)), 1)
		status = strings.Replace(status, "{STRINGID}", left, 1)
	}
	tip := strings.Replace(i18n.Str(i18n.TooltipDaymonthyearChallenge), "{DATE DMY}", date, 1)
	return strings.Replace(tip, "{STRINGID}", status, 1)
}

// barUnderMouse reports the bar of that type as the window under the mouse
// (findAt): its hover.
func (g *Game) barUnderMouse(typ ui.WindowType) bool {
	w := g.wins.FindAt(g.mouseX, g.mouseY)
	return w != nil && w.Base().WinType == typ
}

// drawFPS is Gfx::drawFPS (Graphics/FPSCounter.cpp), drawn when
// Config::showFPS is on (Options → Display, or the test page's "Show FPS"):
// the frames per second measured over a second, "%.0f" (below 10: "%.1f"),
// medium bold, white outlined, centred at the top.
func (g *Game) drawFPS(screen *ebiten.Image) {
	if !g.showFPS {
		return
	}
	g.fpsFrames++
	now := time.Now()
	if g.fpsRef.IsZero() {
		g.fpsRef = now
	}
	if el := now.Sub(g.fpsRef).Seconds(); el > 1.0 {
		g.fpsCurrent = float64(g.fpsFrames) / el
		g.fpsFrames, g.fpsRef = 0, now
	}
	format := "%.0f"
	if g.fpsCurrent < 10 {
		format = "%.1f"
	}
	str := fmt.Sprintf(format, g.fpsCurrent)
	w, _ := ui.MeasureTextBold(str)
	wc, oc := ui.TextColours(2)
	x, y := g.sw/2-w/2, 2 // FPSCounter.cpp (2da9f6f): beside a centred toolbar, top left
	if g.toolbar.Centred {
		x, y = 4, 4
	}
	ui.DrawTextOutlined(screen, str, x, y, wc, oc)
}

// fillRect is a thin wrapper around vector.FillRect using float64 args (no anti-aliasing).
// Replaces the deprecated ebitenutil.DrawRect.
func fillRect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

// drawUISprite draws a G1 sprite at (x,y) in screen coordinates, with no
// palette remap. The sprite's G1 xOffset/yOffset are applied, matching
// OpenLoco's drawImage().
//
// OpenLoco reference: src/OpenLoco/src/Graphics/SoftwareDrawingContext.cpp drawImage()
func (g *Game) drawUISprite(screen *ebiten.Image, spriteID, x, y int) {
	img := g.r.GetSprite(spriteID)
	if img == nil {
		return
	}
	_, _, xOff, yOff, ok := g.r.GetSpriteInfo(spriteID)
	if !ok {
		xOff, yOff = 0, 0
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x+int(xOff)), float64(y+int(yOff)))
	screen.DrawImage(img, opts)
}

// drawOutlinedText draws white-ish text with a 1px dark outline so it stays
// readable over the moving title-screen world (like OpenLoco's title labels).
func drawOutlinedText(screen *ebiten.Image, str string, x, y int, clr color.Color) {
	shadow := color.RGBA{0, 0, 0, 200}
	for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		ui.DrawText(screen, str, x+d[0], y+d[1], shadow)
	}
	ui.DrawText(screen, str, x, y, clr)
}
