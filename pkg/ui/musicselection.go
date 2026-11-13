package ui

import (
	"image"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/jukebox"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// MusicSelection — a port of Ui/Windows/MusicSelection.cpp, the custom
// playlist's editor: 360×238 (resizable 300×100 … 800×800), the skin's
// windowTitlebarColour and windowOptionsColour; the Title and Years column
// headers (each click: ascending / descending / the original order — years
// descending first), a scroll view of the thirty tracks in 12 px rows with a
// checkbox each, and "N music tracks selected". A click on a row toggles the
// track; the last selected one stays.
//
// clickable(MusicSelection.cpp): close sort_title sort_years
//
// PARTIAL: CJK row height (15).

const (
	msW, msH           = 360, 238
	msMinW, msMinH     = 300, 100
	msMaxW, msMaxH     = 800, 800
	msColumnYearsWidth = 75 // kColumnYearsWidth
	msStatusClearance  = 13 // kStatusBarClearance
	msPadding          = 4
	msRowHeight        = 12
)

// MusicSelectionWindow is the window.
type MusicSelectionWindow struct {
	WinBase
	TitleColour, Colour uint8
	// Cfg is the audio config whose customJukebox the rows toggle; OnChange
	// follows a toggle (Config::write, Audio::revalidateCurrentTrack).
	Cfg      *jukebox.Config
	OnChange func()

	sortMode jukebox.SortMode
	playlist []jukebox.MusicID
	hover    int
	scroll   int
	hold     L.VScrollHold
	content  [4]int
}

// NewMusicSelection is MusicSelection::open's window (createWindow: the
// stack places it), sorted in the original order.
func NewMusicSelection(cfg *jukebox.Config, titleColour, colour uint8) *MusicSelectionWindow {
	w := &MusicSelectionWindow{WinBase: WinBase{WinType: WindowTypeMusicSelection, W: msW, H: msH, MinW: msMinW, MinH: msMinH, MaxW: msMaxW, MaxH: msMaxH,
		WinFlags: WindowFlagResizable}, TitleColour: titleColour, Colour: colour, Cfg: cfg, hover: -1}
	w.setSortMode(jukebox.SortOriginal)
	return w
}

func (w *MusicSelectionWindow) setSortMode(m jukebox.SortMode) {
	w.sortMode, w.playlist = m, jukebox.AllMusicPlaylist(m)
}

// cycleSortMode: (original or other) → order1 → order2 → original.
func (w *MusicSelectionWindow) cycleSortMode(order1, order2 jukebox.SortMode) {
	switch w.sortMode {
	case order2:
		w.setSortMode(jukebox.SortOriginal)
	case order1:
		w.setSortMode(order2)
	default:
		w.setSortMode(order1)
	}
}

// The widgets as onResize lays them out for the window's size (x, y, w, h).
func (w *MusicSelectionWindow) sortYears() [4]int {
	right := w.W - 1 - msPadding
	return [4]int{right - msColumnYearsWidth + 1, 17, msColumnYearsWidth, 12}
}
func (w *MusicSelectionWindow) sortTitle() [4]int {
	return [4]int{msPadding + 1, 17, w.sortYears()[0] - (msPadding + 1), 12}
}
func (w *MusicSelectionWindow) list() [4]int {
	return [4]int{msPadding, 30, w.W - 2*msPadding, w.H - msStatusClearance - 30}
}

func (w *MusicSelectionWindow) contentHeight() int { return msRowHeight * len(w.playlist) }

// PrepareDraw keeps the scroll within the rows the resized list shows.
func (w *MusicSelectionWindow) PrepareDraw() {
	w.scroll = min(w.scroll, max(w.contentHeight()-(w.list()[3]-2), 0))
}

func (w *MusicSelectionWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

func (w *MusicSelectionWindow) onBar(px, py int) bool {
	l := w.list()
	return w.in(l, px, py) && w.contentHeight() > l[3] && px >= w.X+l[0]+l[2]-L.ScrollbarSize-L.ScrollbarMargin
}

// rowAt is the row under a screen point of the list (-1: none).
func (w *MusicSelectionWindow) rowAt(px, py int) int {
	l := w.list()
	if !w.in(l, px, py) || w.onBar(px, py) {
		return -1
	}
	if i := (py - (w.Y + l[1] + 1) + w.scroll) / msRowHeight; py >= w.Y+l[1]+1 && i < len(w.playlist) {
		return i
	}
	return -1
}

// GripBlocked: the scroll view and the status bar reach into the resize
// corner.
func (w *MusicSelectionWindow) GripBlocked(px, py int) bool {
	if w.in(w.list(), px, py) {
		return true
	}
	return w.in([4]int{msPadding, w.H - 12, w.W - ResizeHandleSize - 2*msPadding, 11}, px, py)
}

func (w *MusicSelectionWindow) selected() int {
	n := 0
	for _, on := range w.Cfg.CustomJukebox {
		if on {
			n++
		}
	}
	return n
}

// HandleClick is onMouseUp (close, the headers) and onScrollMouseDown.
func (w *MusicSelectionWindow) HandleClick(px, py int) {
	switch {
	case w.in([4]int{w.W - 15, 2, 13, 13}, px, py):
		w.Close()
	case w.in(w.sortTitle(), px, py):
		w.cycleSortMode(jukebox.SortTitleAscending, jukebox.SortTitleDescending)
	case w.in(w.sortYears(), px, py):
		w.cycleSortMode(jukebox.SortYearsDescending, jukebox.SortYearsAscending)
	case w.onBar(px, py):
		l := w.list()
		w.scroll = w.hold.Begin(py-(w.Y+l[1]), py, l[3], w.contentHeight(), w.scroll)
	default:
		row := w.rowAt(px, py)
		if row < 0 {
			return
		}
		// Toggle the track; "Do not allow there being 0 tracks enabled".
		track := w.playlist[row]
		w.Cfg.CustomJukebox[track] = !w.Cfg.CustomJukebox[track]
		if w.selected() == 0 {
			w.Cfg.CustomJukebox[track] = true
		}
		if w.OnChange != nil {
			w.OnChange()
		}
	}
}

// MouseOver is onScrollMouseOver: the row under the mouse (it stays when the
// mouse leaves the rows, as upstream's rowHover).
func (w *MusicSelectionWindow) MouseOver(px, py int) {
	if i := w.rowAt(px, py); i >= 0 {
		w.hover = i
	}
}

// Drag is the scrollbar's held press each tick.
func (w *MusicSelectionWindow) Drag(px, py int, held bool) {
	l := w.list()
	w.scroll = w.hold.Continue(held, w.onBar(px, py), py-(w.Y+l[1]), py, l[3], w.contentHeight(), w.scroll)
}

// HandleWheel scrolls the list.
func (w *MusicSelectionWindow) HandleWheel(dy float64) {
	w.scroll = min(max(w.scroll-int(dy*3*msRowHeight), 0), max(w.contentHeight()-(w.list()[3]-2), 0))
}

// WidgetTooltip is the tooltip of the widget under a screen point.
func (w *MusicSelectionWindow) WidgetTooltip(px, py int) string {
	switch {
	case w.in([4]int{w.W - 15, 2, 13, 13}, px, py):
		return i18n.Str(i18n.TooltipCloseWindow)
	case w.in(w.sortTitle(), px, py):
		return i18n.Str(i18n.TooltipSortByTrackTitle)
	case w.in(w.sortYears(), px, py):
		return i18n.Str(i18n.TooltipSortByMusicYears)
	case w.in(w.list(), px, py):
		return i18n.Str(i18n.MusicSelectionTooltip)
	}
	return ""
}

// uint16Label puts values into a string's {UINT16} codes, in order.
func uint16Label(id int, vals ...int) string {
	s := i18n.Label(id)
	for _, v := range vals {
		s = strings.Replace(s, "{UINT16}", strconv.Itoa(v), 1)
	}
	return s
}

// musicYears is drawScroll's years text for a track, and whether it is
// right-aligned (a track with an end year only).
func musicYears(mi jukebox.MusicInfo) (text string, right bool) {
	neverPlays := mi.StartYear > mi.EndYear
	hasStart := !(mi.StartYear == jukebox.BeginningOfTime || neverPlays)
	hasEnd := !(mi.EndYear == jukebox.EndOfTime || neverPlays)
	switch {
	case hasStart && hasEnd:
		return uint16Label(i18n.YearRange, mi.StartYear, mi.EndYear), false
	case hasStart:
		return uint16Label(i18n.YearRangeNoEnd, mi.StartYear), false
	case hasEnd:
		return uint16Label(i18n.YearRangeNoStart, mi.EndYear), true
	}
	return i18n.Label(i18n.YearRangeNoStartNoEnd), false
}

// Draw draws the window.
func (w *MusicSelectionWindow) Draw(screen *ebiten.Image) {
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
	st, sy, l := w.sortTitle(), w.sortYears(), w.list()
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, false)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 15, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0) // Panel
		drawResizeHandle(rt, &w.WinBase, sec)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		for _, h := range [][4]int{st, sy} { // TableHeader: Button::draw
			gfx8.FillRectInset(rt, h[0], h[1], h[0]+h[2]-1, h[1]+h[3]-1, gfx8.InsetColour{Map: sec}, 0)
		}
		cx, cy, cw, ch, _ := drawVScrollViewPressed(rt, l[0], l[1], l[2], l[3], gfx8.InsetColour{Map: sec}, w.contentHeight(), w.scroll, w.hold.PressedPart())
		w.content = [4]int{cx, cy, cw, ch}
		if cw <= 0 || ch <= 0 {
			return
		}
		// drawScroll, into the scroll view's own target (its clip).
		sub := gfx8.NewRenderTarget(cw, ch)
		gfx8.FillRect(sub, 0, 0, cw-1, ch-1, gfx8.Shade(sec, 4)) // clear(shade 4)
		m30, _ := g1.GetPaletteMap(0x30)                         // ExtColour::unk30
		for row := range w.playlist {
			y := row*msRowHeight - w.scroll
			if y+msRowHeight <= 0 || y >= ch {
				continue
			}
			if row == w.hover {
				gfx8.FillRectTranslucent(sub, 0, y, cw-1, y+msRowHeight-1, m30)
			}
			gfx8.FillRectInset(sub, 2, y, 11, y+10, gfx8.InsetColour{Map: sec}, gfx8.InsetColourLight|gfx8.InsetFillDarker|gfx8.InsetBorderInset)
		}
		for yy := range ch {
			copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
		}
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, i18n.Label(i18n.MusicSelectionTitle))
	clip := func(x, y, cw, ch int) *ebiten.Image {
		if c, ok := screen.SubImage(image.Rect(w.X+x, w.Y+y, w.X+x+cw, w.Y+y+ch)).(*ebiten.Image); ok {
			return c
		}
		return screen
	}
	// The headers' captions: "{MOVE_X 15}Title", "Years", with ▲ / ▼ for the
	// column the list is sorted by; each clipped to its widget.
	titleID, yearsID := i18n.TableHeaderTitle, i18n.TableHeaderYears
	switch w.sortMode {
	case jukebox.SortTitleAscending:
		titleID = i18n.TableHeaderTitleAsc
	case jukebox.SortTitleDescending:
		titleID = i18n.TableHeaderTitleDesc
	case jukebox.SortYearsAscending:
		yearsID = i18n.TableHeaderYearsAsc
	case jukebox.SortYearsDescending:
		yearsID = i18n.TableHeaderYearsDesc
	}
	DrawTextBold(clip(st[0], st[1], st[2]-2, st[3]), headerText(titleID), w.X+st[0]+15, w.Y+st[1], black)
	DrawTextBold(clip(sy[0], sy[1], sy[2]-2, sy[3]), headerText(yearsID), w.X+sy[0], w.Y+sy[1], black)
	// The status bar: black_stringid, to the resize handle.
	n := w.selected()
	status := uint16Label(i18n.StatusMusicTracksSelectedPlural, n)
	if n == 1 {
		status = uint16Label(i18n.StatusMusicTracksSelectedSingular, n)
	}
	DrawTextBold(clip(msPadding, w.H-12, w.W-ResizeHandleSize-2*msPadding, 11), status, w.X+msPadding, w.Y+w.H-12, black)
	if w.contentHeight() > l[3] {
		drawVScrollArrows(screen, w.X+l[0], w.Y+l[1], l[2], l[3])
	}
	// The rows' text: the tick, the title at 15, the years under their header.
	ct := w.content
	rows := clip(ct[0], ct[1], ct[2], ct[3])
	ox, oy := w.X+ct[0], w.Y+ct[1]
	yearsX := sy[0] - 5 // columnYearsOffset
	for row, track := range w.playlist {
		y := oy + row*msRowHeight - w.scroll
		if y+msRowHeight <= oy || y >= oy+ct[3] {
			continue
		}
		c := black
		if row == w.hover {
			c = win2
		}
		if w.Cfg.CustomJukebox[track] {
			DrawTextBold(rows, string(rune(0xAC)), ox+2, y, win2) // StringIds::checkmark, wcolour2
		}
		mi := jukebox.Info(track)
		DrawTextBold(rows, i18n.Label(mi.TitleID), ox+15, y, c)
		years, right := musicYears(mi)
		if right {
			yw, _ := MeasureTextBold(years)
			DrawTextBold(rows, years, ox+yearsX+msColumnYearsWidth-L.ScrollbarSize-7-yw, y, c)
		} else {
			DrawTextBold(rows, years, ox+yearsX, y, c)
		}
	}
}

// headerText is a table header's caption without its codes ("{MOVE_X 15}
// {COLOUR BLACK}Title ▼": the drawing applies them).
func headerText(id int) string {
	s := i18n.Str(id)
	for strings.HasPrefix(s, "{") {
		end := strings.Index(s, "}")
		if end < 0 {
			break
		}
		s = s[end+1:]
	}
	return s
}
