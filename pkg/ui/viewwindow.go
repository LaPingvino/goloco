package ui

import (
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/hajimehoshi/ebiten/v2"
)

// ViewWindow is the first tab both the Town window (TownWindow.cpp) and the
// Industry window (IndustryWindow.cpp) open on: frame and caption (the
// place's name), the tabs, the viewport on the place (zoom quarter), the
// centre-view button in its corner, and a status line (the town's size and
// population, the industry's getStatusString). Outside the editor / sandbox
// the expand / demolish buttons are hidden and the viewport takes their 22 px.
// clickable(TownWindow.cpp, IndustryWindow.cpp): close_button centre_on_viewport
// The industry's Transported tab (IndustryWindow.cpp Transported, 300×127):
// last month's received and produced cargo; its Production tabs are disabled
// where it produces no first / second cargo (setDisabledWidgets).
// The Production and Production 2 tabs (299×282) graph the monthly production
// history of its first / second cargo (Common::draw). Each tab captions its own
// way ("{name}", "… - Monthly Production", "… - Statistics").
// PARTIAL: the town's other tabs (Population, Company ratings), the tabs'
// images, renaming by the caption, the editor's buttons, resizing, the viewport's own scrolling, the
// status's clipping (drawStringLeftClipped).
type ViewWindow struct {
	X, Y, W, H          int
	Tabs                int    // 3 (town) or 4 (industry)
	Title               string // the place's name
	Status              string // the status bar's line
	TitleColour, Colour uint8
	// View renders the viewport; OnCentreView moves the main view to the
	// place; OnClose closes the window.
	View         func(w, h int) *gfx8.RenderTarget
	OnCentreView func()
	OnClose      func()
	viewImg      *ebiten.Image

	Tab         int     // currentTab
	TabDisabled [4]bool // disabledWidgets' tabs
	// CargoTab is the Transported tab's index (0: none); Cargo gives its
	// lines (received and produced last month), read each draw.
	CargoTab int
	Cargo    func() (receives, produces bool, lines []CargoMonthLine)
	// Production gives the Production tabs' graph (n 0, 1): the label (the
	// first cargo's unitsAndCargoName on both tabs, as upstream), that
	// cargo's monthly history, and the current month (0–11) and year.
	Production func(n int) (label string, history []uint8, month, year int)
	baseW      int // the first tab's size (switchTab: kWindowSize)
	baseH      int
}

// CargoMonthLine is a Transported tab line: a cargo's unit name, how much,
// and, produced, the percent transported.
type CargoMonthLine struct {
	Produced bool
	Unit     string
	Qty      uint32
	Percent  uint8
}

// Transported tab: kWindowSize (onResize sets it); the Production tabs:
// kMinWindowSize.
const (
	transportedW, transportedH = 300, 127
	productionW, productionH   = 299, 282
)

// caption is Common::prepareDraw's caption per tab: title_town,
// title_industry_monthly_production, title_statistics.
func (w *ViewWindow) caption() string {
	switch {
	case w.Tab == 0:
		return w.Title
	case w.Tab == w.CargoTab:
		return strings.Replace(i18n.Label(i18n.TitleStatistics), "{STRINGID}", w.Title, 1)
	}
	return strings.Replace(i18n.Label(i18n.TitleIndustryMonthlyProduction), "{STRINGID}", w.Title, 1)
}

// NewTownWindow is Town::open's window (223×161, three tabs), its status
// status_town_population "{STRINGID} population {INT32}". Placed as
// WindowManager's default position (here centred on the screen).
func NewTownWindow(sw, sh int, title, size string, population int, titleColour, colour uint8) *ViewWindow {
	status := strings.Replace(i18n.Label(i18n.StatusTownPopulation), "{STRINGID}", size, 1)
	status = strings.Replace(status, "{INT32}", strconv.Itoa(population), 1)
	return newViewWindow(sw, sh, 223, 161, 3, title, status, titleColour, colour)
}

// NewIndustryWindow is IndustryWindow open's window (223×137, four tabs), its
// status the industry's getStatusString.
func NewIndustryWindow(sw, sh int, title, status string, titleColour, colour uint8) *ViewWindow {
	return newViewWindow(sw, sh, 223, 137, 4, title, status, titleColour, colour)
}

func newViewWindow(sw, sh, w, h, tabs int, title, status string, titleColour, colour uint8) *ViewWindow {
	return &ViewWindow{X: (sw - w) / 2, Y: (sh - h) / 2, W: w, H: h, Tabs: tabs, Title: title, Status: status,
		TitleColour: titleColour, Colour: colour, baseW: w, baseH: h}
}

// viewport is prepareDraw's viewport outside the editor: 3,44 to
// width-26+22, height-14 (inclusive edges).
func (w *ViewWindow) viewport() (l, t, r, b int) { return 3, 44, w.W - 26 + 22, w.H - 14 }

func (w *ViewWindow) Contains(px, py int) bool {
	return px >= w.X && py >= w.Y && px < w.X+w.W && py < w.Y+w.H
}

// HandleClick is onMouseUp: close, the tabs (switchTab: the first and the
// Transported tab), centre_on_viewport.
func (w *ViewWindow) HandleClick(px, py int) {
	rx, ry := px-w.X, py-w.Y
	if rx >= w.W-15 && rx < w.W-2 && ry >= 2 && ry < 15 {
		if w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	if ry >= 15 && ry < 42 && rx >= 3 && rx < 3+31*w.Tabs {
		i := (rx - 3) / 31
		graph := w.Production != nil && (i == 1 || i == 2) && w.CargoTab != 0
		if !w.TabDisabled[i] && (i == 0 || i == w.CargoTab || graph) {
			w.Tab = i
			switch {
			case i == 0:
				w.W, w.H = w.baseW, w.baseH
			case i == w.CargoTab:
				w.W, w.H = transportedW, transportedH
			default:
				w.W, w.H = productionW, productionH
			}
		}
		return
	}
	if w.Tab != 0 {
		return
	}
	_, _, r, b := w.viewport()
	if rx >= r-24 && rx <= r-1 && ry >= b-24 && ry <= b-1 && w.OnCentreView != nil {
		w.OnCentreView()
	}
}

// Draw is the first tab's draw.
func (w *ViewWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	l, t, r, b := w.viewport()
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for i := 0; i < w.Tabs; i++ {
			if w.TabDisabled[i] {
				drawTabDisabled(rt, 3+31*i, 15, sec)
				continue
			}
			drawTab(rt, 3+31*i, 15, i == w.Tab, sec, tabNoIcon, nil)
		}
		if w.Tab == 0 {
			gfx8.FillRectInset(rt, l, t, r, b, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
		}
	})
	if w.Tab != 0 {
		if w.Tab == w.CargoTab {
			w.drawCargo(screen, g1.GetPalette()[gfx8.Shade(sec, 9)], g1.GetPalette()[paletteBlack0])
		} else {
			w.drawProduction(screen, sec, g1.GetPalette()[paletteBlack0])
		}
		DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, w.caption(), colourWhite)
		return
	}
	if w.View != nil {
		if rt := w.View(r-l-1, b-t-1); rt != nil {
			if sp, err := gfx8.Decode(&g1.Elements[imgCentreView]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, (r-24)-(l+1), (b-24)-(t+1), gfx8.OpTransparent|gfx8.OpSrc, sec, nil)
			}
			pal := g1.GetPalette()
			pix := make([]byte, rt.Width*rt.Height*4)
			for k, ix := range rt.Bits[:rt.Width*rt.Height] {
				c := pal[ix]
				pix[k*4], pix[k*4+1], pix[k*4+2], pix[k*4+3] = c.R, c.G, c.B, 255
			}
			if w.viewImg == nil || w.viewImg.Bounds().Dx() != rt.Width || w.viewImg.Bounds().Dy() != rt.Height {
				w.viewImg = ebiten.NewImage(rt.Width, rt.Height)
			}
			w.viewImg.WritePixels(pix)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(w.X+l+1), float64(w.Y+t+1))
			screen.DrawImage(w.viewImg, op)
			MarkForeign(image.Rect(w.X+l+1, w.Y+t+1, w.X+r, w.Y+b))
		}
	}
	// the status, left-clipped at the status bar (3, height-12) − 1
	black := g1.GetPalette()[paletteBlack0]
	DrawTextFont(screen, w.Status, G1FontMediumNormal, w.X+3-1, w.Y+w.H-12-1, black)
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, w.Title, colourWhite)
}

// drawCargo is Transported::draw: from (3, 45), "Received last month:" in
// WINDOW_2 and each required cargo's amount 4 in, wrapped at 290; 4 below,
// "Produced last month:" and each produced cargo's amount with its percent
// transported.
func (w *ViewWindow) drawCargo(screen *ebiten.Image, win2, black color.Color) {
	if w.Cargo == nil {
		return
	}
	receives, produces, lines := w.Cargo()
	x, y := w.X+3, w.Y+45
	block := func(heading int, produced bool) {
		DrawTextBold(screen, i18n.Label(heading), x, y, win2)
		x, y = x+4, y+10
		for _, l := range lines {
			if l.Produced != produced {
				continue
			}
			text := FillArgs(ObjectString(l.Unit), Comma32(int64(l.Qty)))
			if produced { // transported_cargo "{STRINGID} ({INT16}% transported)"
				text = strings.Replace(strings.Replace(i18n.Label(i18n.TransportedCargo), "{STRINGID}", text, 1), "{INT16}", strconv.Itoa(int(l.Percent)), 1)
			}
			for _, part := range WrapTextBold(text, 290) {
				DrawTextBold(screen, part, x, y, black)
				y += 10
			}
		}
	}
	if receives {
		block(i18n.ReceivedCargo, false)
		x, y = x-4, y+4
	}
	if produces {
		block(i18n.ProducedCargo, true)
	}
}

// drawProduction is Common::draw (0x00456079) for a Production tab: the
// "(units of cargo)" label at (2, 44); from the graph's bottom (height − 7)
// up to y 68 every 20 px a grid line (x 41, 239 wide, shade 4) labelled
// 0, 1000, 2000 … right-aligned at x 39 (population_graph_people, as
// upstream); then back from the newest month a pixel per month: each
// January a year label (every third, centred at y 56) and a vertical grid
// line, and the history joined month to month in shade 7 (the newest
// segment not drawn: it may be zeroed).
func (w *ViewWindow) drawProduction(screen *ebiten.Image, sec []uint8, black color.Color) {
	if w.Production == nil {
		return
	}
	label, hist, month, year := w.Production(w.Tab - 1)
	small := func(s string, x, y int) { DrawTextFont(screen, s, G1FontSmall, x, y, black) }
	bottom := w.H - 7
	grid := gfx8.Shade(sec, 4)
	type tick struct {
		x, y int
		s    string
	}
	var ticks []tick
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		yTick := 0
		for y := bottom; y >= 68; y -= 20 {
			gfx8.FillRect(rt, 41, y, 41+239-1, y, grid)
			s := strconv.Itoa(yTick)
			ticks = append(ticks, tick{39 - MeasureTextFont(s, G1FontSmall), y - 6, s})
			yTick += 1000
		}
		yearSkip := 0
		for i := len(hist) - 1; i > 0; i-- {
			x := 41 + i
			if month == 0 {
				if yearSkip == 0 {
					s := strconv.Itoa(year)
					ticks = append(ticks, tick{x - MeasureTextFont(s, G1FontSmall)/2, 56, s})
				}
				gfx8.FillRect(rt, x, 56+11, x, 56+11+w.H-74-1, grid)
			}
			if i < len(hist)-1 {
				y1, y2 := bottom-int(hist[i]), bottom-int(hist[i+1])
				chartDrawLine(rt, x, y1, x+1, y2, gfx8.Shade(sec, 7))
			}
			if month == 0 {
				month, year = 11, year-1
				if yearSkip++; yearSkip >= 3 {
					yearSkip = 0
				}
			} else {
				month--
			}
		}
	})
	small(strings.Replace(i18n.Label(i18n.ProductionGraphLabel), "{STRINGID}", ObjectString(label), 1), w.X+2, w.Y-24+68)
	for _, t := range ticks {
		small(t.s, w.X+t.x, w.Y+t.y)
	}
}
