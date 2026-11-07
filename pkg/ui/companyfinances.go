package ui

import (
	"fmt"
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
// The scroll view's horizontal bar shows once the years overflow its 499 px
// (getScrollSize: numExpenditureYears × 128, + 1); a held button steps on
// each tick while the pointer stays on it (scrollLeftContinue), pressed
// while it does.

// CompanyFinancesSize is Finances::kWindowSize.
var CompanyFinancesSize = [2]int{636, 319}

// CompanyColourSchemeSize is ColourScheme::kWindowSize.
var CompanyColourSchemeSize = [2]int{265, 252}

// CompanyDetailsSize is Details::kWindowSize.
var CompanyDetailsSize = [2]int{340, 194}

// CompanyStatusSize is Status::kWindowSize.
var CompanyStatusSize = [2]int{270, 182}

// CompanyChallengeSize is Challenge::kWindowSize.
var CompanyChallengeSize = [2]int{320, 182}

// CompanyCargoDeliveredSize is CargoDelivered::kWindowSize.
var CompanyCargoDeliveredSize = [2]int{240, 382}

// Company window tabs (currentTab: widx − tab_status).
const (
	CompanyTabStatus = iota
	CompanyTabDetails
	CompanyTabColourScheme
	CompanyTabFinances
	CompanyTabCargoDelivered
	CompanyTabChallenge
)

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
	lw[cfLoanIncrease] = bvW(87+100-13, 264+1, 12, 10)
	lw[cfLoanAutopay] = bvW(320, 264, 204, 12)
	return lw
}

// CompanyFinances is what the tab shows: Expenditures[i] is the column for
// Years[i] (left to right, oldest first, as drawScroll pairs columnYear and
// columnIndex), the current year marked; the loan, the interest rate and the
// figures below the table, already formatted as the game formats them.
type CompanyFinances struct {
	Years        []int
	NumYears     int // numExpenditureYears (getScrollSize's, uncapped)
	Expenditures [][17]int32
	CurrentYear  int
	Own          bool // the controlling company's: the loan controls show
	Loan         int64
	InterestRate int
	Cash         int64
	Bankrupt     bool
	Value        int64
	Profit       int64
	Autopay      bool // challengeFlags autopayLoan: the checkbox ticked
}

// CompanyFinancesWindow is the window.
type CompanyFinancesWindow struct {
	X, Y, W, H          int
	TitleColour, Colour uint8
	Title               string // the company's name (caption: "{STRINGID} - Finances")
	Data                CompanyFinances
	// Tab is the current tab (CompanyTab…): Finances or Cargo delivered so far.
	Tab int
	// CargoLines is CargoDelivered::draw's list: each delivered cargo's
	// unit name with its count (black_stringid), in cargo order.
	CargoLines []string
	// Challenge is what Challenge::draw shows (the player company's).
	Challenge CompanyChallenge
	// Status is what Status::draw shows: the owner's face (competitor
	// images[ownerEmotion] + 1, in FaceColour), name and status line.
	Status CompanyStatus
	// Details is what Details::draw shows.
	Details CompanyDetails
	// Colours is what ColourScheme::draw shows.
	Colours CompanyColours
	// OnColourMode is a vehicle type\'s checkbox (colourType 1–10): its own
	// scheme on or off (ChangeCompanyColourScheme, setColourMode).
	OnColourMode func(colourType int, on bool)
	// OnRename is onMouseUp caption: renameCompanyPrompt (any tab).
	OnRename func()
	// Positioning: the window follows the pointer (positioningWindow, begun
	// on the caption; released unmoved, the caption's click: rename).
	Positioning        bool
	posLastX, posLastY int
	posMoved           bool
	// OnColour is a swatch\'s onMouseDown (colourType 0 main, 1–10 a vehicle
	// type; primary or secondary) with its screen rect, for showColour.
	OnColour  func(colourType int, primary bool, x, y, w, h int)
	Scroll    int
	hPart     string // _currentScrollArea on the bar ("" none)
	hPressed  bool   // its button / thumb pressed flag (cleared off the part)
	hDragX    int
	SkinImage int // the interface skin's image base (the tab icons)
	FrameNo   int // the window's frameNo: the current tab's icon animates
	OnClose   func()
	// OnLoan is the stepper's − (−1) or + (+1): onMouseDown loan_decrease /
	// loan_increase (a step of £1,000 × the click-repeat step).
	OnLoan func(dir, step int)
	// The company select button (every tab): the company's owner face
	// (images[ownerEmotion], in its colour) and populateCompanySelect's
	// companies; OnCompany is switchCompany.
	Face         int
	FaceColour   uint8
	CompanyID    int
	Companies    []CompanyChoice
	OpenDropdown func(*DropdownMenu)
	OnCompany    func(id int)
	ScreenW      int
	ScreenH      int
	// loanHeld is the loan button pressed (−1 / +1, 0 none) and
	// loanRepeat its _clickRepeatTicks.
	loanHeld, loanRepeat int
	// OnAutopay is onMouseUp loan_autopay: the flag toggled.
	OnAutopay func()
}

// Contains reports whether the point is on the window.
func (w *CompanyFinancesWindow) Contains(x, y int) bool {
	return x >= w.X && x < w.X+w.W && y >= w.Y && y < w.Y+w.H
}

// HandleClick handles a press: the close button, the Finances and Cargo
// delivered tabs (Common::switchTab), the loan stepper and autopay (the
// controlling company's own window only). PARTIAL: the other tabs.
func (w *CompanyFinancesWindow) HandleClick(x, y int) {
	lw := CompanyFinancesLayout(w.W, w.H)
	in := func(b BVWidget) bool { return x >= w.X+b.L && x <= w.X+b.R && y >= w.Y+b.T && y <= w.Y+b.B }
	switch {
	case in(lw[cwClose]) && w.OnClose != nil:
		w.OnClose()
	case w.hBarClick(x, y):
	case in(lw[cwCompanySelect]) && len(w.Companies) > 0 && w.OpenDropdown != nil: // onMouseDown company_select
		w.OpenDropdown(CompanySelectDropdown(w.X+lw[cwCompanySelect].L, w.Y+lw[cwCompanySelect].T, w.TitleColour, w.Companies, w.CompanyID, w.ScreenW, w.ScreenH, w.OnCompany))
	case in(lw[cwCaption]): // windowPositionBegin; the rename comes on an unmoved release (PositionEnd)
		w.Positioning, w.posLastX, w.posLastY, w.posMoved = true, x, y, false
	case in(lw[cwTabColour]) && w.Tab != CompanyTabColourScheme:
		w.switchTab(CompanyTabColourScheme)
	case w.Tab == CompanyTabColourScheme && w.Colours.Own && w.OnColour != nil && w.colourSwatchAt(x, y) != nil:
		s := w.colourSwatchAt(x, y)
		w.OnColour(s.colourType, s.primary, w.X+s.x, w.Y+s.y, 16, 16)
	case w.Tab == CompanyTabColourScheme && w.Colours.Own && w.colourCheckAt(x, y) >= 0 && w.OnColourMode != nil:
		i := w.colourCheckAt(x, y) // onMouseUp check_…: the vehicle type's scheme toggled
		w.OnColourMode(i+1, w.Colours.Custom&(1<<(i+1)) == 0)
	case in(lw[cwTabDetails]) && w.Tab != CompanyTabDetails:
		w.switchTab(CompanyTabDetails)
	case in(lw[cwTabStatus]) && w.Tab != CompanyTabStatus:
		w.switchTab(CompanyTabStatus)
	case in(lw[cwTabFinances]) && w.Tab != CompanyTabFinances:
		w.switchTab(CompanyTabFinances)
	case in(lw[cwTabCargo]) && w.Tab != CompanyTabCargoDelivered:
		w.switchTab(CompanyTabCargoDelivered)
	case in(lw[cwTabChallenge]) && w.Tab != CompanyTabChallenge && w.Data.Own: // disableChallengeTab: another company's
		w.switchTab(CompanyTabChallenge)
	case w.Tab != CompanyTabFinances: // the loan controls are the Finances tab's only
	case w.Data.Own && in(lw[cfLoanDecrease]) && w.OnLoan != nil:
		w.loanHeld, w.loanRepeat = -1, 1
		w.OnLoan(-1, 1)
	case w.Data.Own && in(lw[cfLoanIncrease]) && w.OnLoan != nil:
		w.loanHeld, w.loanRepeat = 1, 1
		w.OnLoan(1, 1)
	case w.Data.Own && in(lw[cfLoanAutopay]) && w.OnAutopay != nil:
		w.OnAutopay()
	}
}

var companyExpenditureLabels = [17]int{
	i18n.TrainIncome, i18n.TrainRunningCosts, i18n.BusIncome, i18n.BusRunningCosts,
	i18n.TruckIncome, i18n.TruckRunningCosts, i18n.TramIncome, i18n.TramRunningCosts,
	i18n.AircraftIncome, i18n.AircraftRunningCosts, i18n.ShipIncome, i18n.ShipRunningCosts,
	i18n.Construction, i18n.VehiclePurchases, i18n.VehicleDisposals, i18n.LoanInterest, i18n.Miscellaneous,
}

// switchTab is Common::switchTab: the window takes the tab's size, keeping
// its place.
func (w *CompanyFinancesWindow) switchTab(tab int) {
	w.Tab = tab
	switch tab {
	case CompanyTabFinances:
		w.W, w.H = CompanyFinancesSize[0], CompanyFinancesSize[1]
	case CompanyTabCargoDelivered:
		w.W, w.H = CompanyCargoDeliveredSize[0], CompanyCargoDeliveredSize[1]
	case CompanyTabChallenge:
		w.W, w.H = CompanyChallengeSize[0], CompanyChallengeSize[1]
	case CompanyTabStatus:
		w.W, w.H = CompanyStatusSize[0], CompanyStatusSize[1]
	case CompanyTabDetails:
		w.W, w.H = CompanyDetailsSize[0], CompanyDetailsSize[1]
	case CompanyTabColourScheme:
		w.W, w.H = CompanyColourSchemeSize[0], CompanyColourSchemeSize[1]
	}
	w.Scroll = 0
	if tab == CompanyTabFinances { // switchTab: Finances::scrollToLatestData
		w.ScrollToLatest()
	}
}

// financesContentW is the Finances scroll view's contentWidth.
func (w *CompanyFinancesWindow) financesContentW() int {
	return w.Data.NumYears*companyFinancesExpenditureColumnWidth + 1
}

// ScrollToLatest is Finances::scrollToLatestData: the last full page.
func (w *CompanyFinancesWindow) ScrollToLatest() {
	sc := CompanyFinancesLayout(w.W, w.H)[cfScroll]
	w.Scroll = L.HScrollLatest(sc.R-sc.L+1, w.financesContentW())
}

// hBar reports whether the Finances scroll view shows its horizontal bar.
func (w *CompanyFinancesWindow) hBar() bool {
	sc := CompanyFinancesLayout(w.W, w.H)[cfScroll]
	return w.Tab == CompanyTabFinances && w.financesContentW() > sc.R-sc.L+1
}

// hBarClick is a press on the bar (ScrollView's getPart and its handlers):
// true when it was on it.
func (w *CompanyFinancesWindow) hBarClick(x, y int) bool {
	if !w.hBar() {
		return false
	}
	sc := CompanyFinancesLayout(w.W, w.H)[cfScroll]
	if x < w.X+sc.L || x > w.X+sc.R || y < w.Y+sc.B-L.ScrollbarSize || y > w.Y+sc.B {
		return false
	}
	ww := sc.R - sc.L + 1
	part := L.HScrollPart(x-(w.X+sc.L), ww, w.financesContentW(), w.Scroll)
	w.hPart, w.hDragX = part, x // scrollLeftBegin
	w.hPressed = part == "left" || part == "right"
	if part != "thumb" {
		w.Scroll = L.HScrollClick(part, ww, w.financesContentW(), w.Scroll)
	}
	return true
}

// UpdateDrag is stateScrollLeft each tick: the release clears the pressed
// buttons; the thumb follows the pointer (horizontalFollow); a button held
// over its own part steps again, and off it its pressed look clears
// (clearPressedButtons); the track does nothing more.
func (w *CompanyFinancesWindow) UpdateDrag(x, y int, down bool) {
	if w.hPart == "" {
		return
	}
	if !down || w.Tab != CompanyTabFinances {
		w.hPart, w.hPressed = "", false
		return
	}
	sc := CompanyFinancesLayout(w.W, w.H)[cfScroll]
	ww, cw := sc.R-sc.L+1, w.financesContentW()
	switch w.hPart {
	case "thumb":
		w.hPressed = true
		w.Scroll = L.HScrollDrag(ww, cw, w.Scroll, x-w.hDragX)
		w.hDragX = x
	case "left", "right":
		on := y >= w.Y+sc.B-L.ScrollbarSize && y <= w.Y+sc.B && x >= w.X+sc.L && x <= w.X+sc.R &&
			L.HScrollPart(x-(w.X+sc.L), ww, cw, w.Scroll) == w.hPart
		w.hPressed = on
		if on {
			w.Scroll = L.HScrollClick(w.hPart, ww, cw, w.Scroll)
		}
	}
}

// Draw is the current tab's draw.
func (w *CompanyFinancesWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	if w.Tab == CompanyTabCargoDelivered {
		w.drawCargoDelivered(screen)
		return
	}
	if w.Tab == CompanyTabChallenge {
		w.drawChallenge(screen)
		return
	}
	if w.Tab == CompanyTabStatus {
		w.drawStatus(screen)
		return
	}
	if w.Tab == CompanyTabDetails {
		w.drawDetails(screen)
		return
	}
	if w.Tab == CompanyTabColourScheme {
		w.drawColourScheme(screen)
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
			w.drawCompanySelect(rt, lw[cwCompanySelect])
		}
		w.drawCompanyTabs(rt, lw[:], cwTabFinances, sec, prim)
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
		if w.hBar() { // drawHScroll, below the content
			ch -= L.ScrollbarSize
			w.drawHBar(rt, sc, sec)
		}
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
		if d.Own { // the stepper: TextBox (inset, fillDarker), then its two Buttons
			b := lw[cfCurrentLoan]
			gfx8.FillRectInset(rt, b.L, b.T, b.R, b.B, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			for _, i := range []int{cfLoanDecrease, cfLoanIncrease} {
				b := lw[i]
				gfx8.FillRectInset(rt, b.L, b.T, b.R, b.B, gfx8.InsetColour{Map: sec}, 0)
			}
			// Checkbox::drawCheckBox: a 10 × 10 box (inset, fillDarker).
			a := lw[cfLoanAutopay]
			gfx8.FillRectInset(rt, a.L, a.T, a.L+9, a.T+9, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
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
	// The years' columns, clipped to the scroll view (above the bar).
	clipB := w.Y + sc.B
	if w.hBar() {
		clipB -= L.ScrollbarSize
		pb := pal[paletteBlack0] // the arrows: "\x90\xBE" / "\x90\xAF" at kArrowOffset
		bx, by, bw := w.X+sc.L+L.ScrollbarMargin, w.Y+sc.B+1-L.ScrollbarSize-L.ScrollbarMargin, sc.R-sc.L+1-2*L.ScrollbarMargin
		DrawTextBold(screen, string(rune(0xBE)), bx+2, by-1, pb)
		DrawTextBold(screen, string(rune(0xAF)), bx+bw-L.ScrollbarSize+L.ScrollbarMargin+2, by-1, pb)
	}
	clip := screen.SubImage(image.Rect(w.X+sc.L+1, w.Y+sc.T+1, w.X+sc.R, clipB)).(*ebiten.Image)
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
		b := lw[cfCurrentLoan] // the stepper's value (company_current_loan_value: TextBox::drawText at +1, +1; colour FD: the string's own, black)
		DrawTextBold(screen, L.Currency48(d.Loan), w.X+b.L+1, w.Y+b.T+1, black)
		for _, s := range []struct {
			i  int
			id int
		}{{cfLoanDecrease, i18n.StepperMinus}, {cfLoanIncrease, i18n.StepperPlus}} { // Button: its text centred
			bb := lw[s.i]
			t := stripCodes(i18n.Str(s.id))
			tw, _ := MeasureTextBold(t)
			DrawTextBold(screen, t, w.X+(bb.L+bb.R+1)/2-tw/2, w.Y+bb.T, black)
		}
		// The autopay checkbox: the check mark (0xAC) when set, the label 4
		// past the box — both in the window's colour (opaque).
		a := lw[cfLoanAutopay]
		wc, _ := TextColours(int(w.Colour))
		if d.Autopay {
			DrawTextBold(screen, string(rune(0xAC)), w.X+a.L, w.Y+a.T, wc)
		}
		DrawTextBold(screen, stripCodes(i18n.Str(i18n.LoanAutopay)), w.X+a.L+10+4, w.Y+a.T, wc)
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

// drawFormattedEx is drawFormatted with {COLOUR RED}, {MOVE_X n} (the pen to
// x + n), {NEWLINE} (the next line, 10 below) and {SPRITE n} (g1 image n
// inline, the pen past its width). It returns the y after the last line.
func drawFormattedEx(screen *ebiten.Image, x, y int, s string, win2, black, red color.Color) int {
	cx, col, y0 := x, win2, y
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
			case code == "NEWLINE":
				cx, y = x, y+10
			case strings.HasPrefix(code, "NEWLINE "): // {NEWLINE x y}: the pen to (x, y) from the string's origin
				var nx, ny int
				if _, err := fmt.Sscanf(strings.TrimPrefix(code, "NEWLINE "), "%d %d", &nx, &ny); err == nil {
					cx, y = x+nx, y0+ny
				}
			case strings.HasPrefix(code, "SPRITE "):
				if n, err := strconv.Atoi(strings.TrimPrefix(code, "SPRITE ")); err == nil && globalG1Font != nil && n >= 0 && n < len(globalG1Font.g1.Elements) {
					if sp, err := gfx8.Decode(&globalG1Font.g1.Elements[n]); err == nil {
						sx, sy := cx, y
						draw8(screen, sx+sp.XOffset, sy+sp.YOffset, sp.Width, sp.Height, func(rt *gfx8.RenderTarget) {
							gfx8.DrawSpriteNoise(rt, sp, -sp.XOffset, -sp.YOffset, gfx8.OpFor(sp, false, false), nil, nil)
						})
						cx += sp.Width
					}
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
	return y
}

// Interface skin image ids of the Company window's tabs (InterfaceSkin::ImageIds).
const (
	skinTabCompany              = 173
	skinTabCompanyDetails       = 181
	skinTabColourSchemeFrame0   = 49
	skinTabFinancesFrame0       = 113
	skinTabCargoDeliveredFrame0 = 198
	skinTabCupFrame0            = 129
)

// drawCompanyTabs is Common::drawTabs: each tab's skin icon — status;
// details recoloured to the primary colour; colour scheme (8 frames, /4),
// finances (16 frames, /2), cargo delivered (4, /4) and challenge (16, /4)
// animating on the current tab.
func (w *CompanyFinancesWindow) drawCompanyTabs(rt *gfx8.RenderTarget, lw []BVWidget, current int, sec, prim []uint8) {
	frame := func(base, n, div, tab int) int {
		if tab == current {
			return w.SkinImage + base + (w.FrameNo/div)%n
		}
		return w.SkinImage + base
	}
	icons := []struct {
		img int
		m   []uint8
	}{
		{w.SkinImage + skinTabCompany, nil},
		{w.SkinImage + skinTabCompanyDetails, prim},
		{frame(skinTabColourSchemeFrame0, 8, 4, cwTabColour), nil},
		{frame(skinTabFinancesFrame0, 16, 2, cwTabFinances), nil},
		{frame(skinTabCargoDeliveredFrame0, 4, 4, cwTabCargo), nil},
		{frame(skinTabCupFrame0, 16, 4, cwTabChallenge), nil},
	}
	for i, ic := range icons {
		t := lw[cwTabStatus+i]
		img := ic.img
		if w.SkinImage == 0 {
			img = tabNoIcon
		}
		drawTab(rt, t.L, t.T, cwTabStatus+i == current, sec, img, ic.m)
	}
}

// drawCargoDelivered is CargoDelivered::draw (0x00433ACD): the common
// widgets, "Cargo delivered:" at (5, 47), then each cargo the company has
// delivered at (10, y) ten apart, or cargo_delivered_none.
func (w *CompanyFinancesWindow) drawCargoDelivered(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw := CompanyFinancesLayout(w.W, w.H)
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[cwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[cwClose].L, lw[cwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			w.drawCompanySelect(rt, lw[cwCompanySelect])
		}
		w.drawCompanyTabs(rt, lw[:], cwTabCargo, sec, prim)
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, ui18nFill(i18n.TitleCompanyCargoDelivered, w.Title), int(w.TitleColour))
	red, _ := TextColours(27)
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	y := w.Y + 47
	drawFormattedEx(screen, w.X+5, y, i18n.Str(i18n.CargoDelivered), win2, black, red)
	y += 10
	for _, l := range w.CargoLines {
		drawFormattedEx(screen, w.X+10, y, ui18nFill(i18n.BlackStringid, l), win2, black, red)
		y += 10
	}
	if len(w.CargoLines) == 0 {
		drawFormattedEx(screen, w.X+10, y, i18n.Str(i18n.CargoDeliveredNone), win2, black, red)
	}
}

// CompanyChallenge is the Challenge tab's content.
type CompanyChallenge struct {
	Details         string // scenarioDetails
	Objective       string // formatChallengeArguments, filled in (challenge_value's {STRINGID})
	Progress        int    // challengeProgress (%)
	Completed       bool
	Failed          bool
	CompletedMonths int // completedChallengeInMonths
	TimeLimited     bool
	MonthsLeft      int // timeLimitYears × 12 − monthsInChallenge
}

// ChallengeObjective is what formatChallengeArguments reads of the
// objective and its progress.
type ChallengeObjective struct {
	Type             uint8 // companyValue, vehicleProfit, performanceIndex, cargoDelivery
	Flags            uint8 // beTopCompany, beWithinTopThreeCompanies, withinTimeLimit
	CompanyValue     uint32
	MonthlyProfit    uint32
	PerformanceIndex uint8
	CargoUnitPlural  string // the cargo's unitNamePlural ("" for none)
	CargoAmount      uint32
	TimeLimitYears   uint8
	UntilYear        int  // progress.timeLimitUntilYear
	TitleOrEditor    bool // "within N years" instead of "by the end of"
}

// ChallengeObjectiveText is Scenario::formatChallengeArguments: the
// objective's sentence, then its first two qualifiers (top company / top
// three / the time limit), as the {STRINGID}s it fills.
func ChallengeObjectiveText(o ChallengeObjective) string {
	var extra []string
	if o.Flags&1 != 0 {
		extra = append(extra, i18n.Str(i18n.AndBeTheTopPerformingCompany))
	}
	if o.Flags&2 != 0 {
		extra = append(extra, i18n.Str(i18n.AndBeOneOfTheTop3PerformingCompanies))
	}
	if o.Flags&4 != 0 {
		if o.TitleOrEditor {
			extra = append(extra, ui18nFill(i18n.WithinYears, strconv.Itoa(int(o.TimeLimitYears))))
		} else {
			extra = append(extra, ui18nFill(i18n.ByTheEndOf, strconv.Itoa(o.UntilYear)))
		}
	}
	extra = append(extra, "", "") // the two empty string ids pushed last
	a, b := extra[0], extra[1]
	switch o.Type {
	case 0:
		return ui18nFill(i18n.AchieveACompanyValueOf, L.Currency48(int64(o.CompanyValue)), a, b)
	case 1:
		return ui18nFill(i18n.AchieveAMonthlyProfitFromVehiclesOf, L.Currency48(int64(o.MonthlyProfit)), a, b)
	case 2:
		pi := int16(o.PerformanceIndex) * 10 // formatPerformanceIndex: the index and its rating's name
		rating := i18n.CorporateRatingPlatelayer + min(9, int(pi)/100)
		return ui18nFill(i18n.AchieveAPerformanceIndexOf, int16OneDecimal(pi), i18n.Str(rating), a, b)
	case 3:
		cargo := ""
		if o.CargoUnitPlural != "" {
			cargo = FillArgs(o.CargoUnitPlural, Comma32(int64(o.CargoAmount)))
		}
		return ui18nFill(i18n.Deliver, cargo, a, b)
	}
	return ""
}

// drawChallenge is Challenge::draw (0x00433DEB): the scenario's details, the
// challenge and its objective, then the outcome — success in years and
// months, failure — or the progress and, time-limited, the time left.
// PARTIAL: beaten by the other player (no multiplayer).
func (w *CompanyFinancesWindow) drawChallenge(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw := CompanyFinancesLayout(w.W, w.H)
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[cwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[cwClose].L, lw[cwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			w.drawCompanySelect(rt, lw[cwCompanySelect])
		}
		w.drawCompanyTabs(rt, lw[:], cwTabChallenge, sec, prim)
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, ui18nFill(i18n.TitleCompanyChallenge, w.Title), int(w.TitleColour))
	red, _ := TextColours(27)
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	c := &w.Challenge
	x, y := w.X+5, w.Y+47
	// drawStringLeftWrapped: each line 10 below the last; it returns the
	// position after the last line.
	wrapped := func(s string, width int, col color.Color) {
		for _, part := range strings.Split(s, "{NEWLINE}") {
			lead := ""
			if strings.HasPrefix(part, "{COLOUR") { // a line's colour stays with its first word
				if e := strings.IndexByte(part, '}'); e > 0 {
					lead, part = part[:e+1], part[e+1:]
				}
			}
			for i, ln := range WrapTextBold(stripCodes(part), width) {
				if i == 0 && lead != "" {
					drawFormattedEx(screen, x, y, lead+ln, win2, black, red)
				} else {
					DrawTextBold(screen, ln, x, y, col)
				}
				y += 10
			}
		}
	}
	wrapped(c.Details, w.W-10, black)
	y += 5
	drawFormattedEx(screen, x, y, i18n.Str(i18n.ChallengeLabel), win2, black, red)
	y += 10
	wrapped(ui18nFill(i18n.ChallengeValue, c.Objective), w.W-10, black)
	y += 5
	switch {
	case c.Completed:
		wrapped(ui18nFill(i18n.SuccessYouCompletedTheChallengeInYearsMonths, strconv.Itoa(c.CompletedMonths/12), strconv.Itoa(c.CompletedMonths%12)), w.W-10, black)
		return
	case c.Failed:
		wrapped(i18n.Str(i18n.FailedYouFailedToCompleteTheChallenge), w.W-10, black)
		return
	}
	drawFormattedEx(screen, x, y, ui18nFill(i18n.ProgressTowardsCompletingChallengePercent, strconv.Itoa(c.Progress)), win2, black, red)
	y += 10
	if c.TimeLimited {
		drawFormattedEx(screen, x, y, ui18nFill(i18n.TimeRemainingYearsMonths, strconv.Itoa(c.MonthsLeft/12), strconv.Itoa(c.MonthsLeft%12)), win2, black, red)
	}
}

// CompanyStatus is the Status tab's content.
type CompanyStatus struct {
	Face       int // the owner's image (images[ownerEmotion] + 1); 0 none
	FaceColour uint8
	OwnerName  string
	Status     string // getOwnerStatus, filled in
	Own        bool   // the controlling company's: change_owner_name shows
}

// drawStatus is Status::draw (0x00432055) at 270 × 182 (prepareDraw's
// widgets: the face 177–242 × 57–122, change_owner_name 154–266 × 124–145,
// the status label 3–256 × 170–179): "Owner:" centred 12 above the face, the
// face at its widget + 1, the owner's name centred in change_owner_name + 5,
// the status line at the label − 1. PARTIAL: the headquarters viewport and
// its centre button (no headquarters yet), the jail bars, the company
// select's face, renaming.
func (w *CompanyFinancesWindow) drawStatus(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw := CompanyFinancesLayout(w.W, w.H)
	st := &w.Status
	faceL, faceT, faceR := w.W-93, 57, w.W-28
	nameL, nameT, nameR := w.W-116, 124, w.W-4
	labL, labT := 3, w.H-12 // its right (w.W − 14) clips the line: PARTIAL, not clipped
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[cwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[cwClose].L, lw[cwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			w.drawCompanySelect(rt, lw[cwCompanySelect])
		}
		w.drawCompanyTabs(rt, lw[:], cwTabStatus, sec, prim)
		if st.Face > 0 && st.Face < len(g1.Elements) {
			if sp, err := gfx8.Decode(&g1.Elements[st.Face]); err == nil {
				m, _ := g1.GetPaletteMap(int(st.FaceColour))
				gfx8.DrawSpriteNoise(rt, sp, faceL+1, faceT+1, gfx8.OpFor(sp, m != nil, false), m, nil)
			}
		}
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, ui18nFill(i18n.TitleCompany, w.Title), int(w.TitleColour))
	red, _ := TextColours(27)
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	owner := stripCodes(i18n.Str(i18n.WindowOwner))
	ow, _ := MeasureTextBold(owner)
	drawFormattedEx(screen, w.X+(faceL+faceR)/2-ow/2, w.Y+faceT-12, i18n.Str(i18n.WindowOwner), win2, black, red)
	for i, ln := range WrapTextBold(stripCodes(st.OwnerName), nameR-nameL) { // drawStringCentredWrapped (new_owner's {POP16} prints nothing)
		lw2, _ := MeasureTextBold(ln)
		DrawTextBold(screen, ln, w.X+(nameL+nameR)/2-lw2/2, w.Y+nameT+5+i*10, black)
	}
	drawFormattedEx(screen, w.X+labL-1, w.Y+labT-1, ui18nFill(i18n.BlackStringid, st.Status), win2, black, red)
}

// CompanyDetails is the Details tab's content.
type CompanyDetails struct {
	Started      string // startedDate, formatted (company_details_started)
	Performance  int16  // performanceIndex
	Trend        int    // +1 increased, −1 decreased (challengeFlags)
	OwnerName    string
	Counts       [6]uint16 // transportTypeCount
	Headquarters bool      // headquartersX != −1
}

var companyDetailsCounts = [6]int{i18n.CompanyDetailsTrainsCount, i18n.CompanyDetailsBusesCount, i18n.CompanyDetailsTrucksCount,
	i18n.CompanyDetailsTramsCount, i18n.CompanyDetailsAircraftCount, i18n.CompanyDetailsShipsCount}

// drawDetails is Details::draw (0x00432919): the started date and the
// performance (its trend's string) from (3, 48), the owner 25 below, each
// non-zero vehicle count 5 below that; "Headquarters" over the viewport
// (219–314 × 54–173: a TextBox's inset) and, with none built,
// not_yet_constructed in it. PARTIAL: the AI details, the headquarters
// viewport and the build / rotate buttons (building one), the company
// select's face.
func (w *CompanyFinancesWindow) drawDetails(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw := CompanyFinancesLayout(w.W, w.H)
	d := &w.Details
	vpL, vpT, vpR, vpB := 219, 54, 219+96-1, 54+120-1
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[cwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[cwClose].L, lw[cwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			w.drawCompanySelect(rt, lw[cwCompanySelect])
		}
		w.drawCompanyTabs(rt, lw[:], cwTabDetails, sec, prim)
		gfx8.FillRectInset(rt, vpL, vpT, vpR, vpB, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, ui18nFill(i18n.TitleCompanyDetails, w.Title), int(w.TitleColour))
	red, _ := TextColours(27)
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	x, y := w.X+3, w.Y+48
	drawFormattedEx(screen, x, y, ui18nFill(i18n.CompanyDetailsStarted, d.Started), win2, black, red)
	y += 10
	perf := i18n.CompanyDetailsPerformance
	switch {
	case d.Trend < 0:
		perf = i18n.CompanyDetailsPerformanceDecreasing
	case d.Trend > 0:
		perf = i18n.CompanyDetailsPerformanceIncreasing
	}
	rating := i18n.CorporateRatingPlatelayer + min(9, max(0, int(d.Performance)/100))                                  // formatPerformanceIndex
	drawFormattedEx(screen, x, y, ui18nFill(perf, int16OneDecimal(d.Performance), i18n.Str(rating)), win2, black, red) // two lines ({NEWLINE})
	y += 25
	drawFormattedEx(screen, x, y, ui18nFill(i18n.OwnerLabel, stripCodes(d.OwnerName)), win2, black, red)
	y += 10 + 5
	for i, n := range d.Counts {
		if n != 0 {
			drawFormattedEx(screen, x, y, ui18nFill(companyDetailsCounts[i], strconv.Itoa(int(n))), win2, black, red)
			y += 10
		}
	}
	hq := stripCodes(i18n.Str(i18n.Wcolour2Headquarters))
	hw, _ := MeasureTextBold(hq)
	drawFormattedEx(screen, w.X+(vpL+vpR)/2-hw/2, w.Y+vpT-12, i18n.Str(i18n.Wcolour2Headquarters), win2, black, red)
	if !d.Headquarters {
		for i, ln := range WrapTextBold(stripCodes(i18n.Str(i18n.NotYetConstructed)), vpR-vpL-1) {
			lw2, _ := MeasureTextBold(ln)
			DrawTextBold(screen, ln, w.X+(vpL+vpR)/2-lw2/2, w.Y+(vpT+vpB)/2-5+i*10, black)
		}
	}
}

// CompanyColours is the Colour scheme tab's content: the main scheme, each
// vehicle type's, which types have their own (customVehicleColoursSet, bit
// 1–10), and whether it is the controlling company's (else all disabled).
type CompanyColours struct {
	Main    [2]uint8
	Vehicle [10][2]uint8
	Custom  uint32
	Own     bool
}

const imgColourSwatchRecolourable = 2306 // ImageIds::colour_swatch_recolourable

var companyColourLabels = [10]int{i18n.ColourSteamLocomotives, i18n.ColourDieselLocomotives, i18n.ColourElectricLocomotives,
	i18n.ColourMultipleUnits, i18n.ColourPassengerVehicles, i18n.ColourFreightVehicles, i18n.ColourBuses, i18n.ColourTrucks,
	i18n.ColourAircraft, i18n.ColourShips}

// colourCheckAt is the vehicle-type checkbox (0–9: Checkbox (15, 81 + 17i),
// 204 × 12) under a point, −1 none.
func (w *CompanyFinancesWindow) colourCheckAt(x, y int) int {
	for i := 0; i < 10; i++ {
		l, t := w.X+15, w.Y+81+17*i
		if x >= l && x < l+204 && y >= t && y < t+12 {
			return i
		}
	}
	return -1
}

// drawColourScheme is ColourScheme::draw (0x00432F9A) after prepareDraw: the
// main scheme's two swatches (221 / 239, 48), "Main colour scheme" at
// (6, 48 + 3) and "Special colour schemes used for" 17 below; each vehicle
// type's checkbox (15, 81 + 17i: Checkbox::draw — the box, the tick when its
// bit is set, the label 14 right, in the window's colour) and, when set, its
// two swatches (221 / 239, 78 + 17i). A swatch is ColourButton::draw: the
// recolourable swatch image in its colour. PARTIAL: the hover / pressed
// swatch images, another company's disabled look, the company select's face.
func (w *CompanyFinancesWindow) drawColourScheme(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw := CompanyFinancesLayout(w.W, w.H)
	c := &w.Colours
	swatch := func(rt *gfx8.RenderTarget, x, y int, colour uint8) {
		if imgColourSwatchRecolourable >= len(g1.Elements) {
			return
		}
		if sp, err := gfx8.Decode(&g1.Elements[imgColourSwatchRecolourable]); err == nil {
			m, _ := g1.GetPaletteMap(int(colour))
			gfx8.DrawSpriteNoise(rt, sp, x, y, gfx8.OpFor(sp, m != nil, false), m, nil)
		}
	}
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[cwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[cwClose].L, lw[cwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
			w.drawCompanySelect(rt, lw[cwCompanySelect])
		}
		w.drawCompanyTabs(rt, lw[:], cwTabColour, sec, prim)
		swatch(rt, 221, 48, c.Main[0])
		swatch(rt, 239, 48, c.Main[1])
		for i := 0; i < 10; i++ {
			y := 81 + 17*i
			gfx8.FillRectInset(rt, 15, y, 15+9, y+9, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			if c.Custom&(1<<(i+1)) != 0 {
				swatch(rt, 221, 78+17*i, c.Vehicle[i][0])
				swatch(rt, 239, 78+17*i, c.Vehicle[i][1])
			}
		}
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, ui18nFill(i18n.TitleCompanyColourScheme, w.Title), int(w.TitleColour))
	red, _ := TextColours(27)
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	drawFormattedEx(screen, w.X+6, w.Y+48+3, i18n.Str(i18n.MainColourScheme), win2, black, red)
	drawFormattedEx(screen, w.X+6, w.Y+48+3+17, i18n.Str(i18n.SpecialColourSchemesUsedFor), win2, black, red)
	wc, _ := TextColours(int(w.Colour))
	for i := 0; i < 10; i++ {
		y := w.Y + 81 + 17*i
		if c.Custom&(1<<(i+1)) != 0 {
			DrawTextBold(screen, string(rune(0xAC)), w.X+15, y, wc)
		}
		DrawTextBold(screen, stripCodes(i18n.Str(companyColourLabels[i])), w.X+15+10+4, y, wc)
	}
}

type colourSwatch struct {
	colourType int
	primary    bool
	x, y       int
}

// colourSwatchAt is the ColourButton (16 × 16) under a point: the main
// scheme's (221 / 239, 48), or a vehicle type's when its scheme is set
// (221 / 239, 78 + 17i); nil none.
func (w *CompanyFinancesWindow) colourSwatchAt(x, y int) *colourSwatch {
	try := func(t int, primary bool, sx, sy int) *colourSwatch {
		if x >= w.X+sx && x < w.X+sx+16 && y >= w.Y+sy && y < w.Y+sy+16 {
			return &colourSwatch{t, primary, sx, sy}
		}
		return nil
	}
	for _, p := range []bool{true, false} {
		sx := 221
		if !p {
			sx = 239
		}
		if s := try(0, p, sx, 48); s != nil {
			return s
		}
		for i := 0; i < 10; i++ {
			if w.Colours.Custom&(1<<(i+1)) != 0 {
				if s := try(i+1, p, sx, 78+17*i); s != nil {
					return s
				}
			}
		}
	}
	return nil
}

// PositionMove is statePositioningWindow while held: the pointer's y clamped
// to 29 … screen height − 29, the window moved by its travel.
func (w *CompanyFinancesWindow) PositionMove(x, y, screenH int) {
	y = min(max(y, 29), screenH-29)
	if dx, dy := x-w.posLastX, y-w.posLastY; dx != 0 || dy != 0 {
		w.X += dx
		w.Y += dy
		w.posMoved = true
	}
	w.posLastX, w.posLastY = x, y
}

// PositionEnd is its leftReleased: the last move, then — never moved — the
// caption's mouse-up (renameCompanyPrompt).
func (w *CompanyFinancesWindow) PositionEnd(x, y, screenH int) {
	w.PositionMove(x, y, screenH)
	w.Positioning = false
	if !w.posMoved && w.OnRename != nil {
		w.OnRename()
	}
}

// SwitchTab is Common::switchTab from outside (CompanyWindow::open's tab).
func (w *CompanyFinancesWindow) SwitchTab(tab int) { w.switchTab(tab) }

// UpdateHeld is MouseInput's click repeat for the holdable loan buttons:
// while the button stays pressed over its widget the ticks count, and from
// 16 every 4th repeats onMouseDown with getClickRepeatStepSize (10 to the
// power ticks / 100); the release resets them.
func (w *CompanyFinancesWindow) UpdateHeld(x, y int, down bool) {
	if !down || w.loanHeld == 0 {
		w.loanHeld, w.loanRepeat = 0, 0
		return
	}
	lw := CompanyFinancesLayout(w.W, w.H)
	b := lw[cfLoanIncrease]
	if w.loanHeld < 0 {
		b = lw[cfLoanDecrease]
	}
	if x < w.X+b.L || x > w.X+b.R || y < w.Y+b.T || y > w.Y+b.B || w.Tab != CompanyTabFinances {
		return
	}
	w.loanRepeat++
	if w.loanRepeat >= 16 && w.loanRepeat%4 == 0 && w.OnLoan != nil {
		step := 1
		for i := 0; i < w.loanRepeat/100; i++ {
			step *= 10
		}
		w.OnLoan(w.loanHeld, step)
	}
}

// drawCompanySelect is Common::drawCompanySelect: the owner's face at the
// button's (left + 1, top + 1), recoloured to the company's colour.
func (w *CompanyFinancesWindow) drawCompanySelect(rt *gfx8.RenderTarget, b BVWidget) {
	g1 := globalG1Font.g1
	if w.Face <= 0 || w.Face >= len(g1.Elements) {
		return
	}
	if sp, err := gfx8.Decode(&g1.Elements[w.Face]); err == nil {
		m, _ := g1.GetPaletteMap(int(w.FaceColour))
		gfx8.DrawSpriteNoise(rt, sp, b.L+1, b.T+1, gfx8.OpFor(sp, m != nil, false), m, nil)
	}
}

// drawHBar is ScrollViewWidget's drawHScroll: the hatched track (shade 7,
// hatched shade 3) with its lines at 2, 3, 7 and 8, the left and right
// buttons and the thumb (hThumbLeft … hThumbRight − 1), at the view's
// bottom inside its 1 px margin.
func (w *CompanyFinancesWindow) drawHBar(rt *gfx8.RenderTarget, sc BVWidget, sec []uint8) {
	c := gfx8.InsetColour{Map: sec}
	ww, wh := sc.R-sc.L+1, sc.B-sc.T+1
	x, y := sc.L+L.ScrollbarMargin, sc.T+wh-L.ScrollbarSize-L.ScrollbarMargin
	bw := ww - 2*L.ScrollbarMargin
	s3, s7 := gfx8.Shade(sec, 3), gfx8.Shade(sec, 7)
	gfx8.FillRect(rt, x, y, x+bw-1, y+L.ScrollbarSize-1, s7)
	fillRectCrossHatch(rt, x, y, x+bw-1, y+L.ScrollbarSize-1, s3)
	for _, ln := range [][2]int{{2, int(s3)}, {3, int(s7)}, {7, int(s3)}, {8, int(s7)}} {
		gfx8.FillRect(rt, x, y+ln[0], x+bw-1, y+ln[0], uint8(ln[1]))
	}
	pressed := func(part string) int { // hscrollbar…Pressed: borderInset
		if w.hPressed && w.hPart == part {
			return gfx8.InsetBorderInset
		}
		return 0
	}
	gfx8.FillRectInset(rt, x, y, x+L.ScrollbarSize-1, y+L.ScrollbarSize-1, c, pressed("left"))
	rx := x + bw - L.ScrollbarSize + L.ScrollbarMargin
	gfx8.FillRectInset(rt, rx, y, rx+L.ScrollbarSize-1, y+L.ScrollbarSize-1, c, pressed("right"))
	l, r := L.HThumb(ww, w.financesContentW(), w.Scroll)
	gfx8.FillRectInset(rt, x+l, y, x+l+(r-l-L.ScrollbarMargin)-1, y+L.ScrollbarSize-1, c, pressed("thumb"))
}

// StripCodes is stripCodes for callers outside the package: the text with
// its {…} formatting codes taken out.
func StripCodes(s string) string { return stripCodes(s) }
