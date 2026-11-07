package ui

import (
	"image"
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
// payment rates, speed records: graphs), resizing, the hover row.

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
}

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
	for _, t := range []int{4, 5} { // the values and payment rates tabs: the currency symbol, right-aligned at (left + 28, top + 15)
		sym := L.Currency.Prefix + L.Currency.Suffix
		sw, _ := MeasureTextBold(sym)
		DrawTextBold(screen, sym, w.X+3+31*t+28-sw, w.Y+15+14+1, pal[paletteBlack0])
	}
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
		DrawTextBold(clip, r.Name, ox+0+27, oy+y-1+8, black)
		DrawTextBold(clip, r.Status, ox+175, oy+y+7, black)
		perf := i18n.PerformanceIndex
		if r.Trend != 0 { // (upstream's condition: increased and decreased both set)
			perf = i18n.PerformanceIndexDecrease
			if r.Trend > 0 {
				perf = i18n.PerformanceIndexIncrease
			}
		}
		rating := i18n.CorporateRatingPlatelayer + min(9, max(0, int(r.Performance)/100))
		drawFormattedEx(clip, ox+385, oy+y-1, ui18nFill(perf, int16OneDecimal(r.Performance), i18n.Str(rating)), black, black, red)
		DrawTextBold(clip, L.Currency48(r.Value), ox+530, oy+y-1+8, black)
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
	current := 0
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
