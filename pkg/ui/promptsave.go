package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/hajimehoshi/ebiten/v2"
)

// LoadOrQuitMode is LoadOrQuitMode: what the save prompt asks before.
type LoadOrQuitMode int

const (
	LoadGamePrompt LoadOrQuitMode = iota
	ReturnToTitlePrompt
	QuitGamePrompt
)

// PromptSave is PromptSaveWindow.cpp: 260 × 48, centred, a translucent muted
// dark red; the caption (title_load_game / title_quit_game / _alt), the
// prompt (prompt_save_before_loading / _quitting / _alt) and Save, Don't Save
// and Cancel. OnSave is Game::confirmSaveGame, OnDontSave LoadSaveQuit's
// dontSave; Cancel and the close button close it (closeSavePrompt).
type PromptSave struct {
	X, Y       int
	Mode       LoadOrQuitMode
	OnSave     func()
	OnDontSave func()
	Closed     bool
}

const psW, psH = 260, 48

var (
	psClose  = [4]int{247, 2, 11, 11}
	psSave   = [4]int{8, 33, 78, 12}
	psDont   = [4]int{91, 33, 78, 12}
	psCancel = [4]int{174, 33, 78, 12}
)

// NewPromptSave is PromptSaveWindow::open (createWindowCentred).
func NewPromptSave(mode LoadOrQuitMode, sw, sh int) *PromptSave {
	return &PromptSave{X: (sw - psW) / 2, Y: (sh - psH) / 2, Mode: mode}
}

// W is the window's width (for the open sound's pan).
func (p *PromptSave) W() int { return psW }

func (p *PromptSave) in(r [4]int, px, py int) bool {
	return px >= p.X+r[0] && px < p.X+r[0]+r[2] && py >= p.Y+r[1] && py < p.Y+r[1]+r[3]
}

// HandleClick is onMouseUp.
func (p *PromptSave) HandleClick(px, py int) {
	switch {
	case p.in(psClose, px, py), p.in(psCancel, px, py):
		p.Closed = true
	case p.in(psSave, px, py):
		p.Closed = true
		if p.OnSave != nil {
			p.OnSave()
		}
	case p.in(psDont, px, py):
		p.Closed = true
		if p.OnDontSave != nil {
			p.OnDontSave()
		}
	}
}

func (p *PromptSave) texts() (caption, prompt string) {
	captions := [3]int{i18n.TitleLoadGame, i18n.TitleQuitGame, i18n.TitleQuitGameAlt}
	prompts := [3]int{i18n.PromptSaveBeforeLoading, i18n.PromptSaveBeforeQuitting, i18n.PromptSaveBeforeQuittingAlt}
	return i18n.Label(captions[p.Mode]), i18n.Label(prompts[p.Mode])
}

// Draw is the window's draw: panel, boxed caption, the label centred, the
// three buttons and the close button.
func (p *PromptSave) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	cmap, err := g1.GetPaletteMap(colourMutedDarkRed)
	if err != nil {
		return
	}
	var trans [3][]uint8
	for k := range trans {
		trans[k], _ = g1.GetPaletteMap(int(paint.TranslucentColour[colourMutedDarkRed][k]))
	}
	c := gfx8.InsetColour{Map: cmap, Translucent: true, Trans: trans}
	m2e, _ := g1.GetPaletteMap(extColourUnk2E)
	draw8(screen, p.X, p.Y, psW, psH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 0, psW-1, psH-1, c, 0)                                             // Panel
		gfx8.FillRectInset(rt, 1, 1, 1+258-1, 1+13-1, c, gfx8.InsetBorderInset|gfx8.InsetFillDarker) // Caption boxed
		gfx8.FillRectTranslucent(rt, 2, 2, 2+256-1, 2+11-1, m2e)
		for _, b := range [][4]int{psClose, psSave, psDont, psCancel} {
			gfx8.FillRectInset(rt, b[0], b[1], b[0]+b[2]-1, b[1]+b[3]-1, c, 0)
		}
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, psClose[0], psClose[1], gfx8.OpTransparent|gfx8.OpSrc, cmap, nil)
		}
	})
	pal := g1.GetPalette()
	white := pal[gfx8.Shade(mustMap(colourWhite), 9)]
	black := pal[paletteBlack0]
	text := pal[gfx8.Shade(cmap, 9)]
	caption, prompt := p.texts()
	cw := 258 - 4 - 10
	tw, _ := MeasureTextBold(caption)
	DrawTextOutlined(screen, caption, p.X+1+2+cw/2-tw/2, p.Y+1+1, white, black)
	lw, _ := MeasureTextBold(prompt) // Label promptLabel {2, 17, 256, 12}, centred
	DrawTextBold(screen, prompt, p.X+2+256/2-lw/2, p.Y+17, text)
	for _, b := range []struct {
		r  [4]int
		id int
	}{{psSave, i18n.LabelButtonSave}, {psDont, i18n.LabelButtonDontSave}, {psCancel, i18n.LabelButtonCancel}} {
		s := i18n.Label(b.id)
		bw, _ := MeasureTextBold(s)
		DrawTextBold(screen, s, p.X+b.r[0]+(b.r[2]-1)/2-bw/2, p.Y+b.r[1], text)
	}
}
