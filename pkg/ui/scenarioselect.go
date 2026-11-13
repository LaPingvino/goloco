package ui

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"image/color"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/hajimehoshi/ebiten/v2"
)

// ScenarioSelect — a port of Ui/Windows/ScenarioSelect.cpp (the New Game
// window): 610×412, centred (y at least 28), five category tabs, the scenario
// list (24 px rows) and the selected scenario's details.
// PARTIAL: {COLOUR BLACK} uses PaletteIndex::black0 rather than the g1
// text_palette, wrapping breaks at spaces, the completed tick / "completed by"
// line (no high scores yet), currency swapping.

const (
	ssW, ssH            = 610, 412
	ssRowHeight         = 24
	imgCloseButton      = 2321
	imgWideTab          = 3625
	imgRandomMap        = 2468
	imgCompass          = 2469
	colourMutedSeaGreen = 11
)

// Widget rectangles (x, y, w, h), relative to the window.
var (
	ssCaption = [4]int{1, 1, 608, 34}
	ssClose   = [4]int{595, 2, 13, 13}
	ssPanel   = [4]int{0, 48, 610, 364}
	ssTabs    = [5][4]int{{3, 15, 91, 34}, {94, 15, 91, 34}, {185, 15, 91, 34}, {276, 15, 91, 34}, {367, 15, 91, 34}}
	ssList    = [4]int{3, 52, 431, 356}
)

// ScenarioSelectWindow is the window state.
type ScenarioSelectWindow struct {
	// ObjectiveText, if set, is the scenario index entry's objective string
	// (ScenarioManager loadScenarioProgress: formatChallengeArguments with
	// the delivered cargo's object loaded); else the meta's own description.
	ObjectiveText func(m *scenario.ScenarioMeta) string
	WinBase
	Tab      int
	Scroll   int           // list scroll offset in px
	hold     L.VScrollHold // a press on the scrollbar held (stateScrollLeft)
	Selected *scenario.ScenarioMeta
	byCat    [5][]*scenario.ScenarioMeta
	OnStart  func(*scenario.ScenarioMeta)
	content  [4]int // the list's content rectangle, window-relative (set by Draw)
}

// NewScenarioSelect opens the window for screen size sw × sh.
func NewScenarioSelect(metas []*scenario.ScenarioMeta, sw, sh int) *ScenarioSelectWindow {
	w := &ScenarioSelectWindow{WinBase: WinBase{WinType: WindowTypeScenarioSelect, X: sw/2 - ssW/2, Y: max(sh/2-ssH/2, 28), W: ssW, H: ssH,
		WinFlags: WindowFlagStickToFront}}
	for _, m := range metas {
		c := int(m.Category)
		if c < 0 || c >= 5 {
			c = 0
		}
		w.byCat[c] = append(w.byCat[c], m)
	}
	return w
}

func (w *ScenarioSelectWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

// Contains reports whether the point is on the window.
// clickable(ScenarioSelect.cpp): close tab0 tab1 tab2 tab3 tab4
// HandleClick: tabs switch the category, a list row selects (a second click on
// the selected row starts it — upstream starts on a single click; goloco keeps
// the details visible first), close closes.
func (w *ScenarioSelectWindow) HandleClick(px, py int) {
	if w.in(ssClose, px, py) {
		w.Close()
		return
	}
	for i, r := range ssTabs {
		if w.in(r, px, py) {
			w.Tab, w.Scroll = i, 0
			return
		}
	}
	if content := len(w.byCat[w.Tab]) * ssRowHeight; w.in(ssList, px, py) && content > ssList[3] &&
		px >= w.X+ssList[0]+ssList[2]-L.ScrollbarSize-L.ScrollbarMargin {
		// The scrollbar (ScrollView): buttons and track move the list, the
		// thumb starts a drag (Drag).
		w.Scroll = w.hold.Begin(py-(w.Y+ssList[1]), py, ssList[3], content, w.Scroll)
		return
	}
	if w.in(ssList, px, py) { // onScrollMouseDown: the row clicked starts at once (Scenario::loadAndStart)
		idx := (py - (w.Y + ssList[1] + 1) + w.Scroll) / ssRowHeight
		if idx >= 0 && idx < len(w.byCat[w.Tab]) && w.OnStart != nil {
			w.OnStart(w.byCat[w.Tab][idx])
		}
	}
}

// MouseOver is onScrollMouseOver (0x00443FB2): the row under the pointer is
// the one whose details show.
func (w *ScenarioSelectWindow) MouseOver(px, py int) {
	if !w.in(ssList, px, py) {
		return
	}
	idx := (py - (w.Y + ssList[1] + 1) + w.Scroll) / ssRowHeight
	if idx >= 0 && idx < len(w.byCat[w.Tab]) {
		w.Selected = w.byCat[w.Tab][idx]
	}
}

// HandleWheel scrolls the list.
func (w *ScenarioSelectWindow) HandleWheel(dy float64) {
	maxScroll := max(len(w.byCat[w.Tab])*ssRowHeight-ssList[3], 0)
	w.Scroll = min(max(w.Scroll-int(dy*ssRowHeight), 0), maxScroll)
}

// Draw draws the window.
func (w *ScenarioSelectWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(0) // Colour::black
	sec, _ := g1.GetPaletteMap(colourMutedSeaGreen)
	black := pal[paletteBlack0]
	win2 := pal[gfx8.Shade(sec, 9)]
	// Frame (background style) + "extend background frame with a solid colour".
	DrawWindowFrame(screen, w.X, w.Y, ssW, ssH, 0, false)
	draw8(screen, w.X, w.Y, ssW, ssH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 20, ssW-1, 20+41-1, gfx8.InsetColour{Map: prim}, 0)
		p := ssPanel
		gfx8.FillRectInset(rt, p[0], p[1], p[0]+p[2]-1, p[1]+p[3]-1, gfx8.InsetColour{Map: sec}, 0)
		for i, t := range ssTabs {
			img := imgWideTab
			if i == w.Tab {
				img++
			}
			if sp, err := gfx8.Decode(&g1.Elements[img]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, t[0], t[1], gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
			}
		}
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, ssClose[0], ssClose[1], gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		// The list is a ScrollView: inset frame, scrollbar when needed; its
		// content (drawScroll) cleared to shade 4, the selected row unk30.
		l := ssList
		cx, cy, cw, ch, _ := drawVScrollViewPressed(rt, l[0], l[1], l[2], l[3], gfx8.InsetColour{Map: sec}, len(w.byCat[w.Tab])*ssRowHeight, w.Scroll, w.hold.PressedPart())
		w.content = [4]int{cx, cy, cw, ch}
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 4))
		m30, _ := g1.GetPaletteMap(0x30)
		for i, m := range w.byCat[w.Tab] {
			y := cy + i*ssRowHeight - w.Scroll
			if m == w.Selected && y+ssRowHeight > cy && y < cy+ch {
				// drawRect(0, y, self.width, kRowHeight - 1) in the content target
				gfx8.FillRectTranslucent(rt, cx, max(y, cy), cx+cw-1, min(y+ssRowHeight-2, cy+ch-1), m30)
			}
		}
		// Preview outline (inset, secondary colour) and image.
		if w.Selected != nil {
			bx := ssList[0] + ssList[2] + 4
			by := ssPanel[1] + 5 + 14
			gfx8.FillRectInset(rt, bx+20, by, bx+20+129, by+129, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillNone)
			px, py := bx+21, by+1
			if w.Selected.HasPreview {
				for j := 0; j < 128; j++ {
					for i := 0; i < 128; i++ {
						if v := w.Selected.Preview[j*128+i]; v != 0 {
							rt.Bits[(py+j)*rt.Width+px+i] = v
						}
					}
				}
				if sp, err := gfx8.Decode(&g1.Elements[imgCompass]); err == nil {
					gfx8.DrawSpriteNoise(rt, sp, px, py, gfx8.OpTransparent, nil, nil)
				}
			} else if sp, err := gfx8.Decode(&g1.Elements[imgRandomMap]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, px+1, py+1, gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
			}
		}
	})
	DrawWindowCaption(screen, w.X+ssCaption[0], w.Y+ssCaption[1], ssCaption[2], i18n.Label(i18n.SelectScenarioForNewGame))
	for i, t := range ssTabs {
		off := 0
		if i == w.Tab {
			off = 1
		}
		cx, cy := w.X+t[0]+t[2]/2, w.Y+t[1]+t[3]/2-3-off
		tw, _ := MeasureTextBold(ssGroups()[i])
		DrawTextBold(screen, ssGroups()[i], cx-tw/2, cy, win2) // wcolour2_stringid
	}
	// List rows (content target): the name centred at x = 210, black
	// (selected: window colour 2).
	l, ct := ssList, w.content
	if len(w.byCat[w.Tab])*ssRowHeight > l[3] {
		drawVScrollArrows(screen, w.X+l[0], w.Y+l[1], l[2], l[3])
	}
	for i, m := range w.byCat[w.Tab] {
		y := w.Y + ct[1] + i*ssRowHeight - w.Scroll
		if y+ssRowHeight <= w.Y+ct[1] || y+10 > w.Y+ct[1]+ct[3] {
			continue
		}
		c := black
		if m == w.Selected {
			c = win2
		}
		tw, _ := MeasureTextBold(m.Name)
		DrawTextBold(screen, m.Name, w.X+ct[0]+210-tw/2, y+1, c)
	}
	if w.Selected == nil {
		return
	}
	m := w.Selected
	bx := w.X + ssList[0] + ssList[2] + 4
	by := w.Y + ssPanel[1] + 5
	col := ssPanel[2] - (ssList[0] + ssList[2]) - 6
	tw, _ := MeasureTextBold(m.Name)
	DrawTextBold(screen, m.Name, bx+col/2-tw/2, by, win2)
	if !m.HasPreview {
		y := by + 14 + 1 + 1 + 59
		for k, ln := range wrapTooltip(i18n.Label(i18n.RandomlyGeneratedLandscape), 128, G1FontMediumBold) {
			lw, _ := MeasureTextBold(ln)
			DrawTextBold(screen, ln, bx+20+1+1+64-lw/2, y+k*10, black)
		}
	}
	y := by + 150
	for _, ln := range wrapTooltip(m.Description, 170, G1FontMediumBold) {
		DrawTextBold(screen, ln, bx, y, black)
		y += 10
	}
	y += 5
	DrawTextBold(screen, i18n.Label(i18n.ChallengeLabel), bx, y, win2)
	y += 10
	if m.HasObjective {
		text := m.Objective.DescribeWithCargo(m.ObjectiveCargoName)
		if w.ObjectiveText != nil {
			text = w.ObjectiveText(m)
		}
		for _, ln := range wrapTooltip("  "+text, 170, G1FontMediumBold) { // challenge_value: two spaces, then the objective
			DrawTextBold(screen, ln, bx, y, black)
			y += 10
		}
	}
	y += 5
	// ScenarioSelect.cpp: challenge_start_date, challenge_competing_companies_*.
	y += 10 * drawFormatted(screen, bx, y, strings.Replace(i18n.Str(i18n.ChallengeStartDate), "{UINT16 RAW}", itoa3(int(m.StartYear)), 1), win2, black)
	comp := i18n.Str(i18n.ChallengeCompetingCompaniesNone)
	if m.MaxCompetitors > 0 {
		comp = strings.Replace(i18n.Str(i18n.ChallengeCompetingCompaniesUpTo), "{INT16}", itoa3(int(m.MaxCompetitors)), 1)
	}
	drawFormatted(screen, bx, y, comp, win2, black)
}

// drawFormatted draws a pack string's text with its {COLOUR WINDOW_2} /
// {COLOUR BLACK} changes (the label and value colours) and {NEWLINE}s; other
// codes are dropped. It returns the lines drawn.
func drawFormatted(screen *ebiten.Image, x, y int, s string, win2, black color.Color) int {
	lines, cx, col := 1, x, win2
	for len(s) > 0 {
		if s[0] == '{' {
			end := strings.IndexByte(s, '}')
			if end < 0 {
				break
			}
			switch code := s[1:end]; {
			case code == "COLOUR BLACK":
				col = black
			case strings.HasPrefix(code, "COLOUR"):
				col = win2
			case code == "NEWLINE":
				lines++
				cx, y = x, y+10
			}
			s = s[end+1:]
			continue
		}
		next := strings.IndexByte(s, '{')
		if next < 0 {
			next = len(s)
		}
		part := s[:next]
		DrawTextBold(screen, part, cx, y, col)
		w, _ := MeasureTextBold(part)
		cx += w
		s = s[next:]
	}
	return lines
}

func drawLabelValue(screen *ebiten.Image, x, y int, label, value string, lc, vc color.Color) {
	lw, _ := MeasureTextBold(label)
	DrawTextBold(screen, label, x, y, lc)
	DrawTextBold(screen, value, x+lw, y, vc)
}

func itoa3(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// SelectFirst selects the first scenario of the current tab (diagnostics).
func (w *ScenarioSelectWindow) SelectFirst() {
	if len(w.byCat[w.Tab]) > 0 {
		w.Selected = w.byCat[w.Tab][0]
	}
}

// DiagRepeat repeats every category's list n times (diagnostics: to show the
// scrollbar, which no shipped category is long enough to need).
func (w *ScenarioSelectWindow) DiagRepeat(n int) {
	for c := range w.byCat {
		base := w.byCat[c]
		for i := 1; i < n; i++ {
			w.byCat[c] = append(w.byCat[c], base...)
		}
	}
}

// Drag is the scrollbar's held press each tick (VScrollHold.Continue): the
// thumb follows the pointer, a held button repeats while on it.
func (w *ScenarioSelectWindow) Drag(px, py int, held bool) {
	onBar := w.in(ssList, px, py) && px >= w.X+ssList[0]+ssList[2]-L.ScrollbarSize-L.ScrollbarMargin
	w.Scroll = w.hold.Continue(held, onBar, py-(w.Y+ssList[1]), py, ssList[3], len(w.byCat[w.Tab])*ssRowHeight, w.Scroll)
}

// ssGroups are the tabs' group names (scenario_group_*), read when drawn.
func ssGroups() [5]string {
	return [5]string{i18n.Label(i18n.ScenarioGroupBeginner), i18n.Label(i18n.ScenarioGroupEasy), i18n.Label(i18n.ScenarioGroupMedium),
		i18n.Label(i18n.ScenarioGroupChallenging), i18n.Label(i18n.ScenarioGroupExpert)}
}
