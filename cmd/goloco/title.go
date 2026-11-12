package main

import (
	"image"
	"image/color"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// title.go — the title screen's windows: TitleLogo, TitleMenu, TitleOptions,
// TitleExit, TitleVersion (drawn and clicked here, not through pkg/ui windows).

// Title screen layout constants — match OpenLoco's window positions.
const (
	// Main menu buttons (TitleMenu.cpp)
	titleBtnSize    = 74
	titleBtnCount   = 4
	titleMenuW      = titleBtnSize * titleBtnCount // 296
	titleMenuH      = titleBtnSize
	titleMenuMargin = 25 // pixels from bottom

	// Exit button (TitleExit.cpp) — bottom-right
	titleExitW = 60
	titleExitH = 28

	// Options button (TitleOptions.cpp) — top-right
	titleOptionsW = 60
	titleOptionsH = 15
)

func (g *Game) handleTitleMenuClick(mx, my int) {
	// --- Exit button: bottom-right corner ---
	if inRect(mx, my, g.sw-titleExitW, g.sh-titleExitH, titleExitW, titleExitH) {
		log.Println("[Game] Exit clicked")
		os.Exit(0)
	}

	// --- Options button: top-right corner ---
	if inRect(mx, my, g.sw-titleOptionsW, 0, titleOptionsW, titleOptionsH) {
		g.openOptionsWindow()
		return
	}

	// --- Main menu buttons: bottom-centre ---
	menuX := (g.sw - titleMenuW) / 2
	menuY := g.sh - titleMenuH - titleMenuMargin
	if !inRect(mx, my, menuX, menuY, titleMenuW, titleMenuH) {
		return
	}
	btnIdx := (mx - menuX) / titleBtnSize

	// clickable(TitleMenu.cpp): scenario_list_btn load_game_btn
	switch btnIdx {
	case 0: // New Game — open scenario browser (SC5 only)
		g.openNewGameWindow()
	case 1:
		g.openLoadGameWindow()
	case 2:
		log.Println("[Game] Title menu: Tutorial (not yet implemented)")
	case 3:
		log.Println("[Game] Title menu: Scenario Editor (not yet implemented)")
	}
}

func (g *Game) drawTitleMenu(screen *ebiten.Image) {
	var logoRect image.Rectangle
	defer func() { // what the title menu draws, for draw8's screen copy (pkg/ui translucent.go)
		ui.MarkForeign(logoRect)                                                     // the logo
		ui.MarkForeign(image.Rect(g.sw-titleOptionsW, 0, g.sw, 30))                  // Options
		ui.MarkForeign(image.Rect(0, g.sh-titleMenuH-titleMenuMargin-4, g.sw, g.sh)) // the globes, the corner texts
	}()
	// Sprite ID constants.
	// OpenLoco reference: src/OpenLoco/src/Graphics/ImageIds.h
	const (
		spriteGlobeSpinBase  = 3552 // title_menu_globe_spin_0..31 (sequential)
		spriteGlobeConstBase = 3584 // title_menu_globe_construct_0..31 (sequential)
		spriteGlobeConstIdle = 3608 // globe_construct_24 — default editor frame
		spriteSparkle        = 3547 // title_menu_sparkle (New Game overlay)
		spriteSave           = 3548 // title_menu_save (Load Game overlay)
		spriteLessonL        = 3549 // title_menu_lesson_l (Tutorial overlay)
	)

	// --- Logo: the original Locomotion logo sprite, like OpenLoco's TitleLogo
	// window (298×170 at the top-left). Requires the disc-aligned G1 layout.
	// OpenLoco reference: src/OpenLoco/src/Ui/Windows/TitleLogo.cpp,
	// ImageIds::locomotion_logo = 3624.
	const spriteLocomotionLogo = 3624
	if logo := g.r.GetSprite(spriteLocomotionLogo); logo != nil {
		_, _, xOff, yOff, _ := g.r.GetSpriteInfo(spriteLocomotionLogo)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(xOff), float64(yOff))
		screen.DrawImage(logo, op)
		logoRect = logo.Bounds().Add(image.Pt(int(xOff), int(yOff)))
	} else {
		// Fallback branding when the logo sprite is unavailable.
		const logoScale = 4.0
		logoText := "GoLoco"
		logoW, _ := ui.MeasureTextBold(logoText)
		logoX := (298 - int(float64(logoW)*logoScale)) / 2
		logoY := 72
		shadow := color.RGBA{0, 0, 0, 220}
		for _, d := range [][2]int{{-4, -4}, {0, -4}, {4, -4}, {-4, 0}, {4, 0}, {-4, 4}, {0, 4}, {4, 4}} {
			ui.DrawTextBoldScaled(screen, logoText, logoX+d[0], logoY+d[1], logoScale, shadow)
		}
		ui.DrawTextBoldScaled(screen, logoText, logoX, logoY, logoScale, color.RGBA{220, 220, 255, 255})
		logoRect = image.Rect(0, 0, 298, logoY+int(9.0*logoScale)+8)
	}
	// upstream draws no subtitle under the logo (TitleLogo.cpp)

	// --- Options button: top-right (sw-60, 0), 60×15 ---
	// OpenLoco: TitleOptions.cpp at (Ui::width()-60, 0)
	optX := g.sw - titleOptionsW
	optHovered := inRect(g.mouseX, g.mouseY, optX, 0, titleOptionsW, titleOptionsH)
	// OpenLoco's TitleOptions window is transparent — plain white text with a
	// subtle darkening only on hover (the solid coloured boxes read as
	// misrendered rectangles).
	if optHovered {
		fillRect(screen, float64(optX), 0, float64(titleOptionsW), float64(titleOptionsH), color.RGBA{0, 0, 0, 90})
	}
	optLabel := i18n.Label(i18n.Options) // TitleOptions.cpp: StringIds::options
	optLabelW, _ := ui.MeasureText(optLabel)
	drawOutlinedText(screen, optLabel, optX+(titleOptionsW-optLabelW)/2, (titleOptionsH-10)/2, color.White)

	// --- Main menu buttons: horizontally centred, 25px from bottom ---
	// OpenLoco: TitleMenu.cpp — kBtnMainSize=74, kWW=296,
	//   window at ((Ui::width()-296)/2, Ui::height()-92-25)
	// Globe animation: (var_846/2) % 32, var_846 incremented each frame.
	animFrame := (g.titleFrame / 2) % 32

	type btnDef struct {
		globeBase  int // base sprite ID for animation frames
		idleGlobe  int // idle (non-hover) sprite ID
		overlay    int // overlay sprite ID (-1 = none)
		label      string
		frameCount int // number of animation frames in the sequence
	}
	btns := [titleBtnCount]btnDef{
		{spriteGlobeSpinBase, spriteGlobeSpinBase, spriteSparkle, i18n.Label(i18n.TitleMenuNewGame), 32},
		{spriteGlobeSpinBase, spriteGlobeSpinBase, spriteSave, i18n.Label(i18n.TitleMenuLoadGame), 32},
		{spriteGlobeSpinBase, spriteGlobeSpinBase, spriteLessonL, i18n.Label(i18n.TitleMenuShowTutorial), 32},
		// globe_construct has 25 frames (0–24 = sprites 3584–3608); clamping to 25
		// prevents overshooting into unrelated sprites (e.g. the Sawyer logo).
		{spriteGlobeConstBase, spriteGlobeConstIdle, -1, i18n.Label(i18n.TitleMenuScenarioEditor), 25},
	}

	menuX := (g.sw - titleMenuW) / 2
	menuY := g.sh - titleMenuH - titleMenuMargin

	tip := ""
	for i, btn := range btns {
		bx := menuX + i*titleBtnSize
		by := menuY

		hovered := inRect(g.mouseX, g.mouseY, bx, by, titleBtnSize, titleBtnSize)
		if hovered {
			tip = btn.label // the button's tooltip (TitleMenu.cpp's title_menu_* StringIds)
		}

		globeID := btn.idleGlobe
		if hovered {
			globeID = btn.globeBase + animFrame%btn.frameCount
		}

		if g.r.GetSprite(globeID) != nil {
			// Draw globe base then overlay (OpenLoco: drawImage x2 per button)
			g.drawUISprite(screen, globeID, bx, by)
			if btn.overlay >= 0 {
				g.drawUISprite(screen, btn.overlay, bx, by)
			}
		} else {
			// Fallback: coloured rectangle + text label
			var bgColor color.RGBA
			if hovered {
				bgColor = color.RGBA{80, 120, 80, 220}
			} else {
				bgColor = color.RGBA{50, 90, 50, 200}
			}
			fillRect(screen, float64(bx), float64(by), float64(titleBtnSize), float64(titleBtnSize), bgColor)
			light := color.RGBA{100, 160, 100, 255}
			dark := color.RGBA{30, 60, 30, 255}
			fillRect(screen, float64(bx), float64(by), float64(titleBtnSize), 2, light)
			fillRect(screen, float64(bx), float64(by), 2, float64(titleBtnSize), light)
			fillRect(screen, float64(bx), float64(by+titleBtnSize-2), float64(titleBtnSize), 2, dark)
			fillRect(screen, float64(bx+titleBtnSize-2), float64(by), 2, float64(titleBtnSize), dark)
			labelW, _ := ui.MeasureText(btn.label)
			labelX := bx + (titleBtnSize-labelW)/2
			labelY := by + titleBtnSize/2 + 4
			ui.DrawText(screen, btn.label, labelX, labelY, color.White)
		}
	}
	g.titleTip = tip

	// --- Exit button: bottom-right (sw-40, sh-28), 40×28 ---
	// OpenLoco: TitleExit.cpp at (Ui::width()-40, Ui::height()-28)
	exitX := g.sw - titleExitW
	exitY := g.sh - titleExitH
	exitHovered := inRect(g.mouseX, g.mouseY, exitX, exitY, titleExitW, titleExitH)
	// Transparent like OpenLoco's TitleExit window; hover darkens.
	if exitHovered {
		fillRect(screen, float64(exitX), float64(exitY), float64(titleExitW), float64(titleExitH), color.RGBA{0, 0, 0, 90})
	}
	// TitleExit.cpp: title_exit_game, centred and wrapped — its {NEWLINE}
	// splits it ("Exit / Game", "Spiel / beenden").
	exitLines := strings.Split(i18n.Label(i18n.TitleExitGame), "{NEWLINE}")
	ey := exitY + (titleExitH-10*len(exitLines))/2
	for _, l := range exitLines {
		lw, _ := ui.MeasureText(l)
		drawOutlinedText(screen, l, exitX+(titleExitW-lw)/2, ey, color.White)
		ey += 10
	}

	// --- Version text: bottom-left ---
	// TitleVersion.cpp: the version and the platform, white with the outline
	// style, at (8, height - 30) and 12 px below.
	{
		wc, oc := ui.TextColours(2) // Colour::white
		ui.DrawTextOutlined(screen, "GoLoco v0.1", 8, g.sh-30, wc, oc)
		ui.DrawTextOutlined(screen, runtime.GOOS+" ("+runtime.GOARCH+")", 8, g.sh-30+12, wc, oc)
	}
}
