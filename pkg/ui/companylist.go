package ui

import (
	"image"
	"image/color"
	"sort"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
)

// The company list (Ui/Windows/CompanyList.cpp, the CompanyList tab:
// prepareDraw 0x00435D07, drawScroll 0x00435EA7), opened by the player
// panel's dropdown's last item (CompanyList::open).
//
// PARTIAL: the other tabs (performance, cargo units / distance, values,
// payment rates, speed records: graphs), resizing.

// CompanyListSize is CompanyList::kWindowSize.
var CompanyListSize = [2]int{640, 272}

// CompanyListRow is one company's row.
type CompanyListRow struct {
	ID          int
	Name        string
	Face        int // competitor images[ownerEmotion] (recoloured to FaceColour)
	FaceColour  uint8
	Status      string // getOwnerStatus, formatted
	Performance int16
	Trend       int // performance_index_increase / _decrease (see below)
	Value       int64
	Colour      uint8       // the company's colour (its graph line, shade 6)
	PerfHist    [120]int16  // performanceIndexHistory
	CargoHist   [120]uint32 // cargoUnitsDeliveredHistory (read as int32)
	DistHist    [120]uint32 // cargoUnitsDistanceHistory (read as int32)
	ValueHist   [120]int64  // companyValueHistory (currency48)
	HistorySize int
}

// CompanyListWindow is the window on its CompanyList tab.
type CompanyListWindow struct {
	X, Y, W, H          int
	TitleColour, Colour uint8
	Rows                []CompanyListRow // sorted by SortMode
	SortMode            int              // 0 name, 1 status, 2 performance (open's default), 3 value
	Scroll              int
	Hover               int
	OnClose             func()
	OnRow               func(id int)
	SkinImage           int // the interface skin's image base (the tab icons)
	FrameNo             int
	Tab                 int                // 0 the list, 1 performance, 2 cargo units, 3 cargo distance, 4 values, 5 payment rates, 6 speed records
	Records             [3]SpeedRecord     // land, air, water
	Cargos              []CompanyListCargo // the payment rates tab's lines, by cargo slot
	CurrentMonths       int                // getCurrentYear() × 12 + getCurrentMonth(): the graphs' x range
}

// CompanyPerformanceSize is CompanyPerformance::kWindowSize; the graph
// tabs' kWindowSizes follow it.
var CompanyPerformanceSize = [2]int{635, 322}

var companyTabSizes = [][2]int{CompanyListSize, CompanyPerformanceSize, {640, 272}, {660, 272}, {685, 322}, {495, 342}, {495, 169}}

// SpeedRecord is one of CompanyManager::getRecords(): the speed (mph, 0
// none), whether a company holds it, its face, its name and the date
// (record_date_achieved's {DATE DMY}, formatted).
type SpeedRecord struct {
	Speed      int
	Held       bool
	Face       int
	FaceColour uint8
	Name, Date string
}

// CompanyListCargo is a loaded cargo object for CargoPaymentRates: its name,
// its line colour (_cargoLineColour by slot) and its delivered-cargo payment
// table (Economy::getDeliveryCargoPaymentsTable: 100 units, 10 blocks, 2, 4
// … 120 days).
type CompanyListCargo struct {
	Name     string
	Colour   uint8
	Payments [60]int32
}

// CargoLineColour is CargoPaymentRates' _cargoLineColour.
var CargoLineColour = [32]uint8{27, 4, 16, 7, 18, 13, 25, 8, 21, 0, 2, 3, 5, 6, 9, 10, 11, 12, 14, 15, 17, 19, 1, 20, 22, 23, 24, 26, 28, 29, 30, 1}

// companyListHeaders are the four TableHeaders (x, width) and their labels
// (the sorted one's "▼" variant: the next string id).
var companyListHeaders = [4]struct{ x, w, label int }{
	{4, 175, i18n.TableHeaderCompanyName}, {179, 210, i18n.TableHeaderCompanyStatus},
	{389, 145, i18n.TableHeaderCompanyPerformance}, {534, 100, i18n.TableHeaderCompanyValue},
}

// Sort is sortCompanyList: a stable sort by the mode (getOrder): the name
// and the status by strcmp, the performance and the value highest first.
func (w *CompanyListWindow) Sort() {
	sort.SliceStable(w.Rows, func(i, j int) bool {
		a, b := w.Rows[i], w.Rows[j]
		switch w.SortMode {
		case 0:
			return a.Name < b.Name
		case 1:
			return a.Status < b.Status
		case 2:
			return b.Performance < a.Performance
		case 3:
			return b.Value < a.Value
		}
		return false
	})
}

// Contains reports whether the point is on the window.
func (w *CompanyListWindow) Contains(x, y int) bool {
	return x >= w.X && x < w.X+w.W && y >= w.Y && y < w.Y+w.H
}

// HandleClick: close, a header (sort by it), a row (its company).
func (w *CompanyListWindow) HandleClick(x, y int) {
	in := func(l, t, ww, h int) bool { return x >= w.X+l && x < w.X+l+ww && y >= w.Y+t && y < w.Y+t+h }
	if in(w.W-15, 2, 13, 13) {
		if w.OnClose != nil {
			w.OnClose()
		}
		return
	}
	for t, size := range companyTabSizes { // Common::switchTab (the ported tabs)
		if in(3+31*t, 15, 31, 27) && t != w.Tab {
			w.Tab, w.W, w.H = t, size[0], size[1]
			return
		}
	}
	if w.Tab != 0 {
		return
	}
	for i, h := range companyListHeaders {
		if in(h.x, 43, h.w, 12) {
			if i != w.SortMode {
				w.SortMode = i
				w.Sort()
			}
			return
		}
	}
	if in(3, 56, w.W-4-3+1, w.H-14-56+1) && w.OnRow != nil { // the scroll view: a row (25 high)
		if r := (y - (w.Y + 56 + 1) + w.Scroll) / 25; r >= 0 && r < len(w.Rows) {
			w.OnRow(w.Rows[r].ID)
		}
	}
}

// Draw is the window's draw and drawScroll: the common widgets, the headers
// (the sorted one's label with "▼"), the rows — the face at (1, 2) and the
// name at (27, 8) (table_item_company), the status at (175, 7), the
// performance at (385 + 0, −1 + 3) and its rating 10 below
// (performance_index), the value at (530, −1 + 8) — and the status bar
// ("{COMMA16} companies").
func (w *CompanyListWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	if w.Tab >= 1 && w.Tab <= 5 {
		w.drawGraphTab(screen)
		return
	}
	if w.Tab == 6 {
		w.drawSpeedRecords(screen)
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	svL, svT, svR, svB := 3, 56, w.W-4, w.H-14
	var cx, cy, cw, ch int
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawTabs(rt, sec)
		for _, h := range companyListHeaders {
			r := min(h.x+h.w-1, w.W-8)
			gfx8.FillRectInset(rt, h.x, 43, r, 43+11, gfx8.InsetColour{Map: sec}, 0)
		}
		cx, cy, cw, ch, _ = drawVScrollView(rt, svL, svT, svR-svL+1, svB-svT+1, gfx8.InsetColour{Map: sec}, len(w.Rows)*25, w.Scroll)
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 3)) // clearSingle(shade 3)
		if w.Hover >= 0 && w.Hover < len(w.Rows) {                      // drawRect(0, y, width, 24, ExtColour::unk30, transparent)
			if m30, err := g1.GetPaletteMap(0x30); err == nil {
				y := cy + w.Hover*25 - w.Scroll
				t, b := max(y, cy), min(y+23, cy+ch-1)
				if t <= b {
					gfx8.FillRectTranslucent(rt, cx, t, cx+cw-1, b, m30)
				}
			}
		}
		for i, r := range w.Rows {
			y := cy + i*25 - w.Scroll
			if y+25 <= cy || y >= cy+ch || r.Face <= 0 || r.Face >= len(g1.Elements) {
				continue
			}
			if sp, err := gfx8.Decode(&g1.Elements[r.Face]); err == nil {
				m, _ := g1.GetPaletteMap(int(r.FaceColour))
				sub := gfx8.NewRenderTarget(cw, ch)
				for yy := range ch {
					copy(sub.Bits[yy*cw:(yy+1)*cw], rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw])
				}
				gfx8.DrawSpriteNoise(sub, sp, 1, y-cy-1+2, gfx8.OpFor(sp, m != nil, false), m, nil)
				for yy := range ch {
					copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
				}
			}
		}
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, i18n.Str(i18n.TitleCompanyList))
	w.drawTabCurrency(screen, pal[paletteBlack0])
	red, _ := TextColours(27)
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	for i, h := range companyListHeaders {
		label := h.label
		if i == w.SortMode {
			label++ // table_header_…_desc
		}
		drawFormattedEx(screen, w.X+h.x+1, w.Y+43, i18n.Str(label), win2, black, red)
	}
	clip := screen.SubImage(image.Rect(w.X+cx, w.Y+cy, w.X+cx+cw, w.Y+cy+ch)).(*ebiten.Image)
	ox, oy := w.X+cx, w.Y+cy
	for i, r := range w.Rows {
		y := i*25 - w.Scroll
		if y+25 <= 0 || y >= ch {
			continue
		}
		tc := black // the hovered row: wcolour2_stringid
		if i == w.Hover {
			tc = win2
		}
		DrawTextBold(clip, r.Name, ox+0+27, oy+y-1+8, tc)
		DrawTextBold(clip, r.Status, ox+175, oy+y+7, tc)
		perf := i18n.PerformanceIndex
		if r.Trend != 0 { // (upstream's condition: increased and decreased both set)
			perf = i18n.PerformanceIndexDecrease
			if r.Trend > 0 {
				perf = i18n.PerformanceIndexIncrease
			}
		}
		rating := i18n.CorporateRatingPlatelayer + min(9, max(0, int(r.Performance)/100))
		drawFormattedEx(clip, ox+385, oy+y-1, ui18nFill(perf, int16OneDecimal(r.Performance), i18n.Str(rating)), tc, tc, red)
		DrawTextBold(clip, L.Currency48(r.Value), ox+530, oy+y-1+8, tc)
	}
	status := ui18nFill(i18n.CompaniesPlural, strconv.Itoa(len(w.Rows)))
	if len(w.Rows) == 1 {
		status = ui18nFill(i18n.CompanySingular, "1")
	}
	drawFormattedEx(screen, w.X+3, w.Y+w.H-13, status, win2, black, red)
}

// Interface skin image ids of the company list's tabs.
const (
	skinTabCompanies              = 174
	skinTabPerformanceIndexFrame0 = 65
	skinTabCargoUnitsFrame0       = 73
	skinTabCargoDistanceFrame0    = 81
	skinTabProductionFrame0       = 89
	skinTabCargoPaymentRates      = 202
	skinTabAwards                 = 45
)

// drawTabs is Common::drawTabs: the companies tab; performance, cargo units,
// cargo distance and values (8 frames, /4, on the current tab) recoloured to
// the secondary colour; payment rates; speed records (tab_awards, recoloured).
func (w *CompanyListWindow) drawTabs(rt *gfx8.RenderTarget, sec []uint8) {
	current := w.Tab
	frame := func(base, tab int) int {
		if tab == current {
			return w.SkinImage + base + (w.FrameNo/4)%8
		}
		return w.SkinImage + base
	}
	icons := []struct {
		img     int
		recolor bool
	}{
		{w.SkinImage + skinTabCompanies, false},
		{frame(skinTabPerformanceIndexFrame0, 1), true},
		{frame(skinTabCargoUnitsFrame0, 2), true},
		{frame(skinTabCargoDistanceFrame0, 3), true},
		{frame(skinTabProductionFrame0, 4), true},
		{w.SkinImage + skinTabCargoPaymentRates, false},
		{w.SkinImage + skinTabAwards, true},
	}
	for i, ic := range icons {
		img := ic.img
		if w.SkinImage == 0 {
			img = tabNoIcon
		}
		var m []uint8
		if ic.recolor {
			m = sec
		}
		drawTab(rt, 3+31*i, 15, i == current, sec, img, m)
	}
}

// MouseOver is onScrollMouseOver (called only over the scroll view): the
// row under the pointer is rowHover, −1 below the last row; leaving the
// list keeps it.
func (w *CompanyListWindow) MouseOver(x, y int) {
	if x < w.X+3 || x > w.X+w.W-4 || y < w.Y+56+1 || y > w.Y+w.H-14 {
		return
	}
	w.Hover = -1
	if r := (y - (w.Y + 56 + 1) + w.Scroll) / 25; r >= 0 && r < len(w.Rows) {
		w.Hover = r
	}
}

const (
	companyListLegendMargin  = 6
	companyListLegendWidth   = 100
	companyListWindowPadding = 4
)

// drawGraphTab is the graph tabs' draw and drawGraphAndLegend (0x004365E4):
// each company's history (front to back from its historySize) on the chart —
// 17 below for the x labels, the tab's xOffset left for the y labels every
// yAxisLabelIncrement px, a tick per month and a label every 12, the x range
// the current month — then the legend (a 5 × 5 swatch and the name in the
// small font, 10 apart) at (width − 104, 52). The tabs:
// CompanyPerformance::draw (0x00436490: performanceIndexHistory, xOffset 40,
// labels every 20, percentage_one_decimal_place), CargoUnits (xOffset 45),
// CargoDistance (65; both every 25, cargo_units_delivered "{INT32}") and
// CompanyValues (90, every 25, small_company_value_currency).
// PARTIAL: highlighting a company's line (var_854, hovering the legend).
func (w *CompanyListWindow) drawGraphTab(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	maxHist := 1
	for _, r := range w.Rows {
		maxHist = max(maxHist, r.HistorySize)
	}
	gs := &Chart{Left: 4, Top: 41 + 4, Width: w.W - companyListLegendWidth - companyListLegendMargin - 2*companyListWindowPadding,
		Height: w.H - 41 - 2*companyListWindowPadding, YOffset: 17, XOffset: 40, YAxisLabelIncrement: 20,
		DataEnd: maxHist, XAxisLabelIncrement: 12, YAxisStepSize: 100, Flags: ChartDataFrontToBack,
		XAxisRange: w.CurrentMonths, XAxisStepSize: 1, PointFlags: ChartDrawLines}
	title := i18n.TitleCompanyPerformance
	switch w.Tab {
	case 2:
		gs.XOffset, gs.YAxisLabelIncrement, gs.YAxisStepSize, title = 45, 25, 1000, i18n.TitleCompanyCargoUnits
	case 3:
		gs.XOffset, gs.YAxisLabelIncrement, gs.YAxisStepSize, title = 65, 25, 1000, i18n.TitleCargoDistanceGraphs
	case 4:
		gs.XOffset, gs.YAxisLabelIncrement, gs.YAxisStepSize, title = 90, 25, 10000, i18n.TitleCompanyValues
	case 5: // CargoPaymentRates::draw (0x00437120): back to front, 2 days a tick, a label every 20
		gs.Top, gs.Height = 41+14, w.H-41-20-2*companyListWindowPadding
		gs.XOffset, gs.YAxisLabelIncrement, gs.YAxisStepSize, title = 80, 25, 0, i18n.TitleCargoPaymentRates
		gs.Flags, gs.DataEnd, gs.XAxisLabelIncrement, gs.XAxisRange, gs.XAxisStepSize = 0, 60, 20, 2, 2
	}
	gs.XAxisTickIncrement = (gs.Width - gs.XOffset) / 120
	if w.Tab == 5 {
		gs.XAxisTickIncrement = (gs.Width - gs.XOffset) / 60
	}
	type legendItem struct {
		name   string
		colour uint8
	}
	var legend []legendItem
	for _, c := range w.Cargos {
		if w.Tab != 5 {
			break
		}
		line := make([]int64, len(c.Payments))
		for i, v := range c.Payments {
			line[i] = int64(v)
		}
		cm, _ := g1.GetPaletteMap(int(c.Colour))
		gs.Lines = append(gs.Lines, line)
		gs.DataStart = append(gs.DataStart, 0)
		gs.LineColour = append(gs.LineColour, gfx8.Shade(cm, 6))
		legend = append(legend, legendItem{c.Name, c.Colour})
	}
	for _, r := range w.Rows { // in company order (the legend's)
		if w.Tab == 5 {
			break
		}
		legend = append(legend, legendItem{r.Name, r.Colour})
		line := make([]int64, 120)
		for i := range line {
			switch w.Tab {
			case 1:
				line[i] = int64(r.PerfHist[i])
			case 2:
				line[i] = int64(int32(r.CargoHist[i]))
			case 3:
				line[i] = int64(int32(r.DistHist[i]))
			case 4:
				line[i] = r.ValueHist[i]
			}
		}
		cm, _ := g1.GetPaletteMap(int(r.Colour))
		gs.Lines = append(gs.Lines, line)
		gs.DataStart = append(gs.DataStart, maxHist-r.HistorySize)
		gs.LineColour = append(gs.LineColour, gfx8.Shade(cm, 6))
	}
	var labels []ChartLabel
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawTabs(rt, sec)
		labels = gs.Draw(rt, sec)
		x, y := w.W-companyListLegendWidth-companyListWindowPadding, 52
		for _, r := range legend {
			cm, _ := g1.GetPaletteMap(int(r.colour))
			gfx8.FillRect(rt, x, y+3, x+4, y+7, gfx8.Shade(cm, 6))
			y += 10
		}
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, i18n.Str(title))
	black := pal[paletteBlack0]
	w.drawTabCurrency(screen, black)
	for _, l := range labels { // graph_label_format: the small font, black
		var s string
		if l.XAxis && w.Tab == 5 { // cargo_delivered_days
			s = ui18nFill(i18n.CargoDeliveredDays, strconv.Itoa(int(int16(l.Value))))
			DrawTextFont(screen, s, G1FontSmall, w.X+l.X-MeasureTextFont(s, G1FontSmall)/2, w.Y+l.Y, black)
		} else if l.XAxis { // rawdate_short: the month's short name, a space, the year
			s = i18n.Str(i18n.MonthShortJanuary+int(l.Value%12)) + " " + strconv.Itoa(int(l.Value/12))
			DrawTextFont(screen, s, G1FontSmall, w.X+l.X-MeasureTextFont(s, G1FontSmall)/2, w.Y+l.Y, black)
		} else {
			switch w.Tab {
			case 1: // percentage_one_decimal_place
				s = int16OneDecimal(int16(l.Value)) + "%"
			case 2, 3: // cargo_units_delivered: {INT32} (the currency48 pushed, read as int32)
				s = strconv.Itoa(int(int32(l.Value)))
			case 4: // small_company_value_currency
				s = L.Currency48(l.Value)
			case 5: // cargo_delivered_currency: {CURRENCY32}
				s = L.Currency48(int64(int32(l.Value)))
			}
			DrawTextFont(screen, s, G1FontSmall, w.X+l.X-MeasureTextFont(s, G1FontSmall), w.Y+l.Y, black)
		}
	}
	x, y := w.X+w.W-companyListLegendWidth-companyListWindowPadding, w.Y+52
	for _, r := range legend { // small_black_string, clipped to 94
		DrawTextFont(screen, r.name, G1FontSmall, x+6, y, black)
		y += 10
	}
	if w.Tab == 5 { // the chart's title (100 units, 10 blocks) and "Transit time", centred on the canvas
		mid := w.X + gs.XOffset + (gs.Width-gs.XOffset)/2
		s := stripCodes(ui18nFill(i18n.CargoDeliverGraphTitle, "100", "10"))
		DrawTextFont(screen, s, G1FontSmall, mid-MeasureTextFont(s, G1FontSmall)/2, w.Y+41+1, black)
		s = stripCodes(i18n.Str(i18n.CargoTransitTime))
		DrawTextFont(screen, s, G1FontSmall, mid-MeasureTextFont(s, G1FontSmall)/2, w.Y+w.H-13, black)
	}
}

// drawTabCurrency is drawTabs' text part: the values and payment rates tabs
// carry currency_symbol ("£": the glyph loadCurrency swaps), right-aligned
// at (left + 28, top + 15).
func (w *CompanyListWindow) drawTabCurrency(screen *ebiten.Image, black color.Color) {
	for _, t := range []int{4, 5} {
		sym := stripCodes(i18n.Str(i18n.CurrencySymbol))
		sw, _ := MeasureTextBold(sym)
		DrawTextBold(screen, sym, w.X+3+31*t+28-sw, w.Y+15+14+1, black)
	}
}

// drawSpeedRecords is CompanySpeedRecords::draw: from y + 47, each record
// set — land_speed_record / air / water (11), and if a company holds it its
// face recoloured to its colour at x + 4 (7), "Achieved by … on …" at
// x + 33 (17) — then 5 apart.
func (w *CompanyListWindow) drawSpeedRecords(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawTabs(rt, sec)
		y := 47
		for _, r := range w.Records {
			if r.Speed == 0 {
				continue
			}
			y += 11
			if r.Held {
				if r.Face > 0 && r.Face < len(g1.Elements) {
					if sp, err := gfx8.Decode(&g1.Elements[r.Face]); err == nil {
						m, _ := g1.GetPaletteMap(int(r.FaceColour))
						gfx8.DrawSpriteNoise(rt, sp, 4, y, gfx8.OpFor(sp, m != nil, false), m, nil)
					}
				}
				y += 7 + 17
			}
			y += 5
		}
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, i18n.Str(i18n.TitleSpeedRecords))
	black := pal[paletteBlack0]
	w.drawTabCurrency(screen, black)
	win2, _ := TextColours(int(w.Colour))
	red, _ := TextColours(27)
	y := w.Y + 47
	for i, r := range w.Records {
		if r.Speed == 0 {
			continue
		}
		drawFormattedEx(screen, w.X+4, y, ui18nFill(i18n.LandSpeedRecord+i, L.Velocity(r.Speed)), black, win2, red)
		y += 11
		if r.Held {
			y += 7
			drawFormattedEx(screen, w.X+33, y, ui18nFill(i18n.RecordDateAchieved, r.Name, r.Date), black, win2, red)
			y += 17
		}
		y += 5
	}
}
