package ui

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// The Messages window — a port of Ui/Windows/MessageWindow.cpp, opened by
// the time panel's date: at the screen's right under the toolbar (width −
// 366, 29), the player's window (its company's colour, the skin's
// windowPlayerColor), two tabs.
//
//   - Messages (366 × 217, its height to 1200): a scroll view of every
//     message in 39 px rows — the date in the small font, the text under it
//     wrapped to three lines —, scrolled to the newest at the bottom; a click
//     on a row opens it as news.
//   - Settings (366 × 155): for each of the six kinds of news Off, Ticker
//     Message or News Window, and "Play sound effects for new messages".
//
// clickable(MessageWindow.cpp): close_button tab_messages tab_settings
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/MessageWindow.cpp

const (
	messagesW, messagesH       = 366, 217 // Messages::kMinWindowSize
	messagesMaxH               = 1200
	messageSettingsH           = 155 // Settings::kWindowSize
	messageRowHeight           = 39  // messageHeight
	skinImgTabMessages         = 196 // InterfaceSkin::ImageIds::tab_messages
	skinImgTabMessageSettings  = 197
	messageSettingsRowY        = 47 // the first news dropdown's y; 15 a row
	messageSettingsDropdownX   = 236
	messageSettingsDropdownW   = 124
	messageSettingsCheckboxY   = 137
	messageSettingsCheckboxX   = 4
	messageSettingsLabelX      = 4
	messageSettingsRowDistance = 15
)

// MessageRow is a message of the list: tiny_font_date's date and the text
// (Loco codes as runes, newlines as code 5).
type MessageRow struct {
	Date, Text string
}

// MessagesWindow is the window.
type MessagesWindow struct {
	WinBase
	TitleColour, Colour uint8
	SkinImage           int // the interface skin's image base (the tab icons); < 0 none
	Tab                 int // 0 messages, 1 settings

	// Rows gives the messages, oldest first, read each draw; OnRow is a
	// click on one (scrollMouseDown).
	Rows  func() []MessageRow
	OnRow func(i int)

	// Settings are Config's newsSettings by MessageCriticality (0 none, 1
	// ticker, 2 news window) and PlaySounds its audio.playNewsSounds, as the
	// game keeps them; OnSetting and OnPlaySounds change them.
	Settings     [6]int
	PlaySounds   bool
	OnSetting    func(criticality, value int)
	OnPlaySounds func()
	OpenDropdown func(*DropdownMenu)

	cached  []MessageRow // Rows as PrepareDraw last read them
	hover   int          // rowHover (−1 none)
	scroll  int
	hold    L.VScrollHold
	content [4]int
}

// NewMessagesWindow is MessageWindow::open's window, on its Messages tab at
// (screen width − 366, 29); ScrollToEnd, once it has its rows, is open's
// scroll to the newest.
func NewMessagesWindow(sw int, titleColour, colour uint8) *MessagesWindow {
	return &MessagesWindow{WinBase: WinBase{WinType: WindowTypeMessages, X: sw - messagesW, Y: 29, W: messagesW, H: messagesH,
		MinW: messagesW, MinH: messagesH, MaxW: messagesW, MaxH: messagesMaxH, WinFlags: WindowFlagResizable},
		TitleColour: titleColour, Colour: colour, SkinImage: -1, hover: -1}
}

func (w *MessagesWindow) rows() []MessageRow { return w.cached }

// list is the scroll view as prepareDraw lays it out: (3, 45) to (width − 4,
// height − 14), as x, y, width, height.
func (w *MessagesWindow) list() [4]int { return [4]int{3, 45, w.W - 4 - 3 + 1, w.H - 14 - 45 + 1} }

func (w *MessagesWindow) contentHeight() int { return len(w.rows()) * messageRowHeight }

func (w *MessagesWindow) maxScroll() int {
	return max(w.contentHeight()-(w.list()[3]-2), 0)
}

// ScrollToEnd is open's contentOffsetY: the content's height less the scroll
// view's.
func (w *MessagesWindow) ScrollToEnd() {
	w.scroll = max(w.contentHeight()-w.list()[3], 0)
}

// ShowMessages is open()'s end for a window new or already open: the
// Messages tab, scrolled to the newest.
func (w *MessagesWindow) ShowMessages() {
	if w.Tab != 0 {
		w.switchTab(0)
	}
	w.PrepareDraw()
	w.ScrollToEnd()
}

// PrepareDraw is the tabs' size (Messages::tabReset's bounds, Settings::
// tabReset's fixed size) and the scroll kept within the rows.
func (w *MessagesWindow) PrepareDraw() {
	if w.Rows != nil {
		w.cached = w.Rows()
	}
	if w.Tab == 1 {
		w.SetFixedSize(messagesW, messageSettingsH)
		return
	}
	w.SetSize(messagesW, messagesH, messagesW, messagesMaxH)
	w.scroll = min(w.scroll, w.maxScroll())
}

// switchTab is Common::switchTab: the Messages tab back at its smallest and
// at its first message.
func (w *MessagesWindow) switchTab(tab int) {
	w.Tab, w.hover = tab, -1
	w.WinFlags &^= WindowFlagMaximised
	if tab == 0 {
		w.W, w.H = messagesW, messagesH
	}
	w.scroll = 0 // initScrollWidgets
	w.PrepareDraw()
	w.MoveInsideScreenEdges()
}

func (w *MessagesWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

func (w *MessagesWindow) onBar(px, py int) bool {
	l := w.list()
	return w.Tab == 0 && w.in(l, px, py) && w.contentHeight() > l[3] && px >= w.X+l[0]+l[2]-L.ScrollbarSize-L.ScrollbarMargin
}

// rowAt is the message under a screen point of the list (−1: none).
func (w *MessagesWindow) rowAt(px, py int) int {
	l := w.list()
	if w.Tab != 0 || !w.in(l, px, py) || w.onBar(px, py) || py < w.Y+l[1]+1 {
		return -1
	}
	if i := (py - (w.Y + l[1] + 1) + w.scroll) / messageRowHeight; i < len(w.rows()) {
		return i
	}
	return -1
}

// GripBlocked: the scroll view reaches into the resize corner.
func (w *MessagesWindow) GripBlocked(px, py int) bool { return w.Tab == 0 && w.in(w.list(), px, py) }

// settingsDropdown is news setting i's dropdown (the combo box and its
// button) as x, y, width, height.
func settingsDropdown(i int) [4]int {
	return [4]int{messageSettingsDropdownX, messageSettingsRowY + messageSettingsRowDistance*i, messageSettingsDropdownW, 12}
}

func newsSettingNames() [3]string {
	return [3]string{i18n.Label(i18n.MessageOff), i18n.Label(i18n.MessageTicker), i18n.Label(i18n.MessageWindow)}
}

// HandleClick is onMouseUp (close, the tabs, the checkbox), the Settings
// tab's onMouseDown (a dropdown: Dropdown::show at its combo box, the
// widget's width − 2, the current setting marked) and scrollMouseDown.
func (w *MessagesWindow) HandleClick(px, py int) {
	switch {
	case w.in([4]int{w.W - 15, 2, 13, 13}, px, py):
		w.Close()
		return
	case w.in([4]int{3, 15, 31, 27}, px, py):
		w.switchTab(0)
		return
	case w.in([4]int{34, 15, 31, 27}, px, py):
		w.switchTab(1)
		return
	}
	if w.Tab == 1 {
		if w.in([4]int{messageSettingsCheckboxX, messageSettingsCheckboxY, 346, 12}, px, py) && w.OnPlaySounds != nil {
			w.OnPlaySounds()
			return
		}
		for i := range w.Settings {
			r := settingsDropdown(i)
			if !w.in(r, px, py) || w.OpenDropdown == nil {
				continue
			}
			var items []DropdownItem
			for v, name := range newsSettingNames() {
				items = append(items, DropdownItem{Text: name, Selected: v == w.Settings[i], MoveX10: true})
			}
			d := NewDropdownShow(w.X+r[0], w.Y+r[1], r[2]-2, r[3]+2, w.Colour, items, 1<<20, 1<<20)
			d.OnSelect = func(v int) {
				if v >= 0 && v != w.Settings[i] && w.OnSetting != nil {
					w.OnSetting(i, v)
				}
			}
			w.OpenDropdown(d)
			return
		}
		return
	}
	if w.onBar(px, py) {
		l := w.list()
		w.scroll = w.hold.Begin(py-(w.Y+l[1]), py, l[3], w.contentHeight(), w.scroll)
		return
	}
	if i := w.rowAt(px, py); i >= 0 && w.OnRow != nil {
		w.OnRow(i)
	}
}

// MouseOver is scrollMouseOver with listWindowOnHandleInputEnd: the row
// under the mouse, none once the mouse is off the rows.
func (w *MessagesWindow) MouseOver(px, py int) { w.hover = w.rowAt(px, py) }

// Drag is the scrollbar's held press each tick.
func (w *MessagesWindow) Drag(px, py int, held bool) {
	if w.Tab != 0 {
		return
	}
	l := w.list()
	w.scroll = w.hold.Continue(held, w.onBar(px, py), py-(w.Y+l[1]), py, l[3], w.contentHeight(), w.scroll)
}

// HandleWheel scrolls the list.
func (w *MessagesWindow) HandleWheel(dy float64) {
	if w.Tab == 0 {
		w.scroll = min(max(w.scroll-int(dy*messageRowHeight), 0), w.maxScroll())
	}
}

// WidgetTooltip is the tooltip of the widget under a screen point.
func (w *MessagesWindow) WidgetTooltip(px, py int) string {
	switch {
	case w.in([4]int{w.W - 15, 2, 13, 13}, px, py):
		return i18n.Str(i18n.TooltipCloseWindow)
	case w.in([4]int{3, 15, 31, 27}, px, py):
		return i18n.Str(i18n.TooltipRecentMessages)
	case w.in([4]int{34, 15, 31, 27}, px, py):
		return i18n.Str(i18n.TooltipMessageOptions)
	case w.Tab == 1 && w.in([4]int{messageSettingsCheckboxX, messageSettingsCheckboxY, 346, 12}, px, py):
		return i18n.Str(i18n.PlayNewsSoundEffectsTip)
	}
	return ""
}

// Draw is Messages::draw + drawScroll, or Settings::draw.
func (w *MessagesWindow) Draw(screen *ebiten.Image) {
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
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	rows := w.rows()
	l := w.list()
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true) // WindowFlags::lighterFrame
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 41, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0) // Panel
		drawResizeHandle(rt, &w.WinBase, sec)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for i, img := range [2]int{skinImgTabMessages, skinImgTabMessageSettings} { // drawTabs
			icon := tabNoIcon
			if w.SkinImage >= 0 {
				icon = w.SkinImage + img
			}
			drawTab(rt, 3+31*i, 15, i == w.Tab, sec, icon, nil)
		}
		if w.Tab == 1 {
			for i := range w.Settings {
				r := settingsDropdown(i)
				optDropdown(rt, r[0], r[1], r[2], sec)
			}
			gfx8.FillRectInset(rt, messageSettingsCheckboxX, messageSettingsCheckboxY, messageSettingsCheckboxX+10-1, messageSettingsCheckboxY+10-1,
				gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
			return
		}
		cx, cy, cw, ch, _ := drawVScrollViewPressed(rt, l[0], l[1], l[2], l[3], gfx8.InsetColour{Map: sec}, w.contentHeight(), w.scroll, w.hold.PressedPart())
		// ScrollView::draw crops the content a scrollbar margin short of its
		// size, right and bottom.
		cw, ch = cw-L.ScrollbarMargin, ch-L.ScrollbarMargin
		w.content = [4]int{cx, cy, cw, ch}
		if cw <= 0 || ch <= 0 {
			return
		}
		// drawScroll's clear and the hovered row's wash (0, y, the window's
		// width, 38 high), in the scroll view's own clip.
		gfx8.FillRect(rt, cx, cy, cx+cw-1, cy+ch-1, gfx8.Shade(sec, 4))
		if w.hover >= 0 && w.hover < len(rows) {
			m30, _ := g1.GetPaletteMap(0x30) // ExtColour::unk30
			y := w.hover*messageRowHeight - w.scroll
			if t, b := max(y, 0), min(y+38-1, ch-1); t <= b {
				gfx8.FillRectTranslucent(rt, cx, cy+t, cx+cw-1, cy+b, m30)
			}
		}
	})
	// Caption::Style::colourText: the title in the window's primary colour.
	DrawWindowCaptionColourText(screen, w.X+1, w.Y+1, w.W-2, i18n.Label(i18n.TitleMessages), w.TitleColour)
	if w.Tab == 1 {
		w.drawSettings(screen)
		return
	}
	if w.contentHeight() > l[3] {
		drawVScrollArrows(screen, w.X+l[0], w.Y+l[1], l[2], l[3])
	}
	ct := w.content
	clip, ok := screen.SubImage(image.Rect(w.X+ct[0], w.Y+ct[1], w.X+ct[0]+ct[2], w.Y+ct[1]+ct[3])).(*ebiten.Image)
	if !ok || ct[2] <= 0 || ct[3] <= 0 {
		return
	}
	// The rows: tiny_font_date at (0, y), the text from (0, y + 6) wrapped at
	// the scroll view's width − 14; black, the hovered row in the window's
	// colour (wcolour2_stringid).
	ox, oy := w.X+ct[0], w.Y+ct[1]
	for i, r := range rows {
		y := oy + i*messageRowHeight - w.scroll
		if y+messageRowHeight <= oy || y >= oy+ct[3] {
			continue
		}
		c := black
		if i == w.hover {
			c = win2
		}
		DrawTextFont(clip, r.Date, G1FontSmall, ox, y, c)
		DrawStringLeftWrapped(clip, r.Text, G1FontMediumBold, ox, y+6, l[2]-14, c)
	}
}

// drawSettings is the Settings tab's labels, dropdown values and checkbox.
func (w *MessagesWindow) drawSettings(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	sec, _ := g1.GetPaletteMap(int(w.Colour))
	black, text := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	names := newsSettingNames()
	arrow := string(rune(170)) // StringIds::dropdown
	aw, _ := MeasureTextBold(arrow)
	for i, v := range w.Settings {
		r := settingsDropdown(i)
		DrawTextBold(screen, i18n.Label(i18n.CompanyMajorNews+i), w.X+messageSettingsLabelX, w.Y+r[1], text)
		if v >= 0 && v < len(names) {
			DrawTextBold(screen, names[v], w.X+r[0]+1, w.Y+r[1]+1, black)
		}
		DrawTextBold(screen, arrow, w.X+r[0]+r[2]-12+(11-1)/2-aw/2, w.Y+r[1]+1, black)
	}
	if w.PlaySounds {
		DrawTextBold(screen, string(rune(0xAC)), w.X+messageSettingsCheckboxX, w.Y+messageSettingsCheckboxY, text)
	}
	DrawTextBold(screen, i18n.Label(i18n.PlayNewsSoundEffects), w.X+messageSettingsCheckboxX+14, w.Y+messageSettingsCheckboxY, text)
}
