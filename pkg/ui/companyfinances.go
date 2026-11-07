package ui

import (
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
)

// The Company window's Finances tab (Ui/Windows/CompanyWindow.cpp Finances:
// prepareDraw 0x004332E4, draw 0x004333D0, drawScroll 0x0043361E), opened
// from the player panel's company value (PlayerInfoPanel companyValueMouseUp
// → CompanyWindow::openFinances).
//
// PARTIAL: the other tabs (drawn, not yet opened), the tab icons and the
// company select button's face, the loan stepper's buttons and autopay
// checkbox (drawn as text only), the horizontal scrollbar (shown upstream
// once the years overflow 499 px: from the fourth year), renaming.

// CompanyFinancesSize is Finances::kWindowSize.
var CompanyFinancesSize = [2]int{636, 319}

// Company window widgets (Common::widx then Finances::widx).
const (
	cwFrame = iota
	cwCaption
	cwClose
	cwPanel
	cwTabStatus
	cwTabDetails
	cwTabColour
	cwTabFinances
	cwTabCargo
	cwTabChallenge
	cwCompanySelect
	cfScroll
	cfCurrentLoan
	cfLoanDecrease
	cfLoanIncrease
	cfLoanAutopay
	cfCount
)

// companyFinancesExpenditureColumnWidth is expenditureColumnWidth.
const companyFinancesExpenditureColumnWidth = 128

// CompanyFinancesLayout is Finances::widgets after prepareDraw for a w × h
// window (makeCommonWidgets, the scroll view, the loan stepper, autopay).
func CompanyFinancesLayout(w, h int) (lw [cfCount]BVWidget) {
	lw[cwFrame] = bvW(0, 0, w, h)
	lw[cwCaption] = bvW(1, 1, w-2, 13)
	lw[cwClose] = bvW(w-15, 2, 13, 13)
	lw[cwPanel] = bvW(0, 41, w, h-41)
	for i := 0; i < 6; i++ {
		lw[cwTabStatus+i] = bvW(3+31*i, 15, 31, 27)
	}
	lw[cwCompanySelect] = bvW(w-28, 14, 26, 26) // prepareDraw: left = width − 28, right = width − 3
	lw[cwCompanySelect].R = w - 3
	lw[cfScroll] = bvW(133, 45, 499, 215)
	lw[cfCurrentLoan] = bvW(87, 264, 100, 12) // stepperWidgets: the value box, then − and +
	lw[cfLoanDecrease] = bvW(87+100-26, 264+1, 13, 10)
	lw[cfLoanIncrease] = bvW(87+100-13, 264+1, 13, 10)
	lw[cfLoanAutopay] = bvW(320, 264, 204, 12)
	return lw
}

// CompanyFinances is what the tab shows: Expenditures[i] is the column for
// Years[i] (left to right, oldest first, as drawScroll pairs columnYear and
// columnIndex), the current year marked; the loan, the interest rate and the
// figures below the table, already formatted as the game formats them.
type CompanyFinances struct {
	Years        []int
	Expenditures [][17]int32
	CurrentYear  int
	Own          bool // the controlling company's: the loan controls show
	Loan         int64
	InterestRate int
	Cash         int64
	Bankrupt     bool
	Value        int64
	Profit       int64
}

// CompanyFinancesWindow is the window.
type CompanyFinancesWindow struct {
	X, Y, W, H          int
	TitleColour, Colour uint8
	Title               string // the company's name (caption: "{STRINGID} - Finances")
	Data                CompanyFinances
	Scroll              int
	OnClose             func()
}

// Contains reports whether the point is on the window.
func (w *CompanyFinancesWindow) Contains(x, y int) bool {
	return x >= w.X && x < w.X+w.W && y >= w.Y && y < w.Y+w.H
}

// HandleClick handles a press: the close button.
func (w *CompanyFinancesWindow) HandleClick(x, y int) {
	b := CompanyFinancesLayout(w.W, w.H)[cwClose]
	if x >= w.X+b.L && x <= w.X+b.R && y >= w.Y+b.T && y <= w.Y+b.B && w.OnClose != nil {
		w.OnClose()
	}
}

var companyExpenditureLabels = [17]int{
	i18n.TrainIncome, i18n.TrainRunningCosts, i18n.BusIncome, i18n.BusRunningCosts,
	i18n.TruckIncome, i18n.TruckRunningCosts, i18n.TramIncome, i18n.TramRunningCosts,
	i18n.AircraftIncome, i18n.AircraftRunningCosts, i18n.ShipIncome, i18n.ShipRunningCosts,
	i18n.Construction, i18n.VehiclePurchases, i18n.VehicleDisposals, i18n.LoanInterest, i18n.Miscellaneous,
}

// Draw is Finances::draw and drawScroll.
func (w *CompanyFinancesWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw := CompanyFinancesLayout(w.W, w.H)
	d := &w.Data
	years := len(d.Years)
	sc := lw[cfScroll]
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[cwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[cwClose].L, lw[cwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for i := 0; i < 6; i++ {
			t := lw[cwTabStatus+i]
			drawTab(rt, t.L, t.T, cwTabStatus+i == cwTabFinances, sec, tabNoIcon, nil)
		}
		// The labels' zebra stripes (getShade(secondary, 6), crossHatching).
		stripe := gfx8.Shade(sec, 6)
		for i, y := 0, 62; i < 17; i, y = i+1, y+10 {
			if i%2 == 0 {
				fillRectCrossHatch(rt, 4, y, 129, y+9, stripe)
			}
		}
		// The scroll view: its frame, then the stripes across the content.
		gfx8.FillRectInset(rt, sc.L, sc.T, sc.R, sc.B, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
		cx, cy := sc.L+L.ScrollbarMargin, sc.T+L.ScrollbarMargin
		cw, ch := sc.R-sc.L+1-2*L.ScrollbarMargin, sc.B-sc.T+1-2*L.ScrollbarMargin
		sub := gfx8.NewRenderTarget(cw, ch) // the scroll's content, then copied into the view
		for yy := range ch {
			copy(sub.Bits[yy*cw:(yy+1)*cw], rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw])
		}
		for i, y := 0, 47-sc.T+14; i < 17; i, y = i+1, y+10 {
			if i%2 == 0 {
				fillRectCrossHatch(sub, 0, y, min(cw-1, companyFinancesExpenditureColumnWidth*17), y+9, stripe)
			}
		}
		// drawFinanceSum's rule above each column's sum (black0).
		x := 132 - sc.L + companyFinancesExpenditureColumnWidth - w.Scroll
		for i := 0; i < years; i++ {
			y := 46 - sc.T + 14 + 17*10 + 4
			gfx8.FillRect(sub, max(0, x-companyFinancesExpenditureColumnWidth+10), y-2, min(cw-1, x), y-2, paletteBlack0)
			x += companyFinancesExpenditureColumnWidth
		}
		for yy := range ch {
			copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
		}
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, ui18nFill(i18n.TitleCompanyFinances, w.Title), int(w.TitleColour))
	red, _ := TextColours(27) // Colour::red ({COLOUR RED})
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	// 'Expenditure/Income', underlined (drawStringLeftUnderline: y + 11).
	drawFormattedEx(screen, w.X+5, w.Y+47, i18n.Str(i18n.ExpenditureIncome), win2, black, red)
	drawUnderline(screen, w.X+5, w.Y+47+11, i18n.Str(i18n.ExpenditureIncome), win2)
	for i, y := 0, w.Y+62; i < 17; i, y = i+1, y+10 {
		drawFormattedEx(screen, w.X+5, y-1, ui18nFill(i18n.Wcolour2Stringid, i18n.Str(companyExpenditureLabels[i])), win2, black, red)
	}
	// The years' columns, clipped to the scroll view.
	clip := screen.SubImage(image.Rect(w.X+sc.L+1, w.Y+sc.T+1, w.X+sc.R, w.Y+sc.B)).(*ebiten.Image)
	x := w.X + sc.L + 1 + 132 - sc.L + companyFinancesExpenditureColumnWidth - w.Scroll
	for i := 0; i < years; i++ {
		y := w.Y + 46 // drawFinanceYear: right-aligned, underlined; the current year in WINDOW_2
		yc := black
		if d.Years[i] == d.CurrentYear {
			yc = win2
		}
		ys := strconv.Itoa(d.Years[i])
		tw, _ := MeasureTextBold(ys)
		DrawTextBold(clip, ys, x-tw, y, yc)
		drawUnderline(clip, x-tw, y+11, ys, yc)
		y += 14
		var sum int64
		for j := 0; j < 17; j++ {
			v := int64(d.Expenditures[i][j])
			sum += v
			if v != 0 {
				s := L.Currency48(v)
				vw, _ := MeasureTextBold(s)
				DrawTextBold(clip, s, x-vw, y, black)
			}
			y += 10
		}
		y += 4 // drawFinanceSum: plus_currency48 in black, or currency48 in red
		var c color.Color = black
		s := ui18nFill(i18n.PlusCurrency48, L.Currency48(sum))
		if sum < 0 {
			s, c = L.Currency48(sum), red
		}
		sw, _ := MeasureTextBold(s)
		DrawTextBold(clip, s, x-sw, y, c)
		x += companyFinancesExpenditureColumnWidth
	}
	// Below the table.
	loanTop := w.Y + lw[cfCurrentLoan].T
	if d.Own {
		drawFormattedEx(screen, w.X+7, loanTop, i18n.Str(i18n.CompanyCurrentLoan), win2, black, red)
		b := lw[cfCurrentLoan] // the stepper's value (company_current_loan_value), its box
		DrawTextBold(screen, L.Currency48(d.Loan), w.X+b.L+2, w.Y+b.T, black)
		drawFormattedEx(screen, w.X+b.R+3, w.Y+b.T+1, ui18nFill(i18n.InterestPerYear, strconv.Itoa(d.InterestRate)), win2, black, red)
	}
	cash := i18n.CashPositive
	switch {
	case d.Bankrupt:
		cash = i18n.CashBankrupt
	case d.Cash < 0:
		cash = i18n.CashNegative
	}
	drawFormattedEx(screen, w.X+7, loanTop+13, ui18nFill(cash, L.Currency48(d.Cash)), win2, black, red)
	drawFormattedEx(screen, w.X+7, loanTop+26, ui18nFill(i18n.CompanyValue, L.Currency48(d.Value)), win2, black, red)
	drawFormattedEx(screen, w.X+7, loanTop+39, ui18nFill(i18n.ProfitFromVehicles, L.Currency48(d.Profit)), win2, black, red)
}

// ui18nFill is FillArgs on a game string.
func ui18nFill(id int, args ...string) string { return FillArgs(i18n.Str(id), args...) }

// drawUnderline is the underline drawStringLeftUnderline adds (the text's
// colour, one pixel, as wide as the text). PARTIAL: the second line in
// textRemap1 (an outlined colour's).
func drawUnderline(screen *ebiten.Image, x, y int, text string, c color.Color) {
	tw, _ := MeasureTextBold(stripCodes(text))
	for i := 0; i < tw; i++ {
		screen.Set(x+i, y, c)
	}
}

// stripCodes drops a game string's {…} codes.
func stripCodes(s string) string {
	var b strings.Builder
	for len(s) > 0 {
		i := strings.IndexByte(s, '{')
		if i < 0 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:i])
		j := strings.IndexByte(s[i:], '}')
		if j < 0 {
			break
		}
		s = s[i+j+1:]
	}
	return b.String()
}

// drawFormattedEx is drawFormatted with {COLOUR RED} and {MOVE_X n} (the pen
// to x + n).
func drawFormattedEx(screen *ebiten.Image, x, y int, s string, win2, black, red color.Color) {
	cx, col := x, win2
	for len(s) > 0 {
		if s[0] == '{' {
			end := strings.IndexByte(s, '}')
			if end < 0 {
				break
			}
			switch code := s[1:end]; {
			case code == "COLOUR BLACK":
				col = black
			case code == "COLOUR RED":
				col = red
			case strings.HasPrefix(code, "COLOUR"):
				col = win2
			case strings.HasPrefix(code, "MOVE_X "):
				if n, err := strconv.Atoi(strings.TrimPrefix(code, "MOVE_X ")); err == nil {
					cx = x + n
				}
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
		pw, _ := MeasureTextBold(part)
		cx += pw
		s = s[next:]
	}
}
