package ui

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/olmath"
	"github.com/hajimehoshi/ebiten/v2"
)

// The news window — a port of Ui/Windows/News/News.cpp and Common.cpp: a
// message shown at the bottom of the screen, 360 wide and centred, sliding
// up 4 px an update (at once for old news and at faster game speeds).
//
//   - News1 (360 × 117), advice: a translucent red (mutedDarkRed) panel, the
//     text in medium bold centred on (180, 17), wrapped at 338.
//   - News2 (360 × 159), general news: a newspaper of the message's year —
//     before 1945 the old paper (the background recoloured translucentBrown1
//     and a brown wash over the page and the first picture), from 1945 the
//     new one, its pictures greyed until 1985; the text in the large font
//     centred on (180, 38), wrapped at 352, the date in the corner.
//
// Under the text one or two subjects: a viewport on an industry, a station,
// a town or a vehicle, a company owner's face, or a turning vehicle (a new
// vehicle), each under a button over its whole column that opens the
// subject's window, with the subject's name in the small font at the bottom.
//
// clickable(News.cpp): close_button viewport1Button viewport2Button
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/News/News.cpp, Common.cpp,
// News.h; Ui/Widgets/NewsPanelWidget.cpp, ImageButtonWidget.cpp,
// ViewportWidget.cpp; Vehicles/VehicleDraw.cpp drawVehicleOverview.

const (
	newsW        = 360
	newsAdviceH  = 117 // News1::kWindowSize
	newsPaperH   = 159 // News2::kWindowSize
	tickerW      = 111 // Ticker::kWindowSize
	tickerH      = 26
	colourGrey   = 1
	colourMutedR = 25 // Colour::mutedDarkRed

	imgNewsOldLeft  = 3539 // ImageIds::news_background_old_left
	imgNewsOldRight = 3540
	imgNewsNewLeft  = 3541
	imgNewsNewRight = 3542
	imgOwnerJailed  = 3630 // ImageIds::owner_jailed

	extTranslucentGrey1  = 53  // ExtColour::translucentGrey1
	extTranslucentBrown1 = 104 // ExtColour::translucentBrown1
)

// NewsSubjectKind is what a subject of a message shows.
type NewsSubjectKind uint8

const (
	NewsSubjectNone    NewsSubjectKind = iota // no subject, or one that is gone
	NewsSubjectView                           // a viewport on it
	NewsSubjectFace                           // SubjectType::companyFace
	NewsSubjectVehicle                        // SubjectType::vehicleImage
)

// NewsSubject is one of a message's two subjects as getView and
// drawViewportString make it.
type NewsSubject struct {
	Kind NewsSubjectKind
	// Present: the type has this subject and the message still names one
	// (hasFirstItem / hasSecondItem, itemSubjects ≠ 0xFFFF).
	Present bool
	// Selectable: the subject's button is there (getView's selectable).
	Selectable bool
	// Caption is drawViewportString's line ("" none).
	Caption string
	// View renders the viewport (NewsSubjectView); Labels its names.
	View   func(w, h int) *gfx8.RenderTarget
	Labels ViewLabels
	// The owner's face (competitor images[ownerEmotion] + 1) in the company's
	// colour, behind bars when jailed.
	Face       int
	FaceColour uint8
	Jailed     bool
	// The vehicle object (drawVehicleOverview), its track's display offset
	// and the player's colour scheme for it.
	Vehicle            *objects.VehicleObject
	VehicleYOffset     int
	Primary, Secondary uint8
}

// NewsWindow is the news window (News1 or News2).
type NewsWindow struct {
	WinBase
	Paper     bool   // News2: general news
	Year      int    // the message's year: the paper's style
	Text      string // the message (Loco codes as runes, newlines as code 5)
	Date      string // news_date's date
	HasSecond bool   // the type's hasSecondItem: the first subject keeps to the left half
	Subjects  [2]NewsSubject
	// PreviewFrame is WindowManager's vehiclePreviewRotationFrame.
	PreviewFrame int
	// OnDismiss is the close button (MessageManager::clearActiveMessage);
	// OnSubject a subject's button.
	OnDismiss func()
	OnSubject func(i int)

	slideInHeight int
	hover         int // the subject button under the mouse (−1: none)
	viewImg       [2]*ebiten.Image
}

// NewNewsWindow is createNewsWindow: at the screen's bottom centre, 5 px of
// it showing — all of it when the news is old or the game runs faster than
// normal.
func NewNewsWindow(paper, old, speedNormal bool, sw, sh int) *NewsWindow {
	h := newsAdviceH
	if paper {
		h = newsPaperH
	}
	w := &NewsWindow{WinBase: WinBase{WinType: WindowTypeNews, W: newsW, H: h, WinFlags: WindowFlagStickToFront, NoCaption: true},
		Paper: paper, slideInHeight: 5, hover: -1}
	if !speedNormal || old {
		w.slideInHeight = h
	}
	w.X, w.Y = sw/2-newsW/2, sh-w.slideInHeight
	return w
}

// Update is Common::onUpdate, once per interface update: 4 px further up
// until it all shows, kept at the screen's bottom centre.
func (w *NewsWindow) Update(sw, sh int) {
	w.slideInHeight = min(w.slideInHeight+4, w.H)
	w.X, w.Y = sw/2-w.W/2, sh-w.slideInHeight
}

// viewport and button are subject i's widgets as initViewport lays them out
// (inclusive edges): kViewportLayouts — the first at x 4, 351 wide alone and
// 174 beside a second, the second at x 186, 174 wide.
func (w *NewsWindow) viewport(i int) (l, t, r, b int) {
	x, width := w.column(i)
	return x + 2, w.H - 73, x + width - 4, w.H - 73 + 64 - 1
}

func (w *NewsWindow) button(i int) (l, t, r, b int) {
	x, width := w.column(i)
	return x, w.H - 75, x + width, w.H - 1
}

func (w *NewsWindow) column(i int) (x, width int) {
	switch {
	case i == 1:
		return 186, 174
	case w.HasSecond:
		return 4, 174
	}
	return 4, 351
}

// buttonAt is the subject button under a screen point (findWidgetAt: the
// last widget there, hidden ones skipped); −1 none.
func (w *NewsWindow) buttonAt(px, py int) int {
	rx, ry := px-w.X, py-w.Y
	if rx < 0 || ry < 0 || rx >= w.W || ry >= w.H {
		return -1
	}
	for i := 1; i >= 0; i-- {
		if l, t, r, b := w.button(i); w.Subjects[i].Selectable && rx >= l && rx <= r && ry >= t && ry <= b {
			return i
		}
	}
	return -1
}

func (w *NewsWindow) onClose(px, py int) bool {
	rx, ry := px-w.X, py-w.Y
	return rx >= w.W-15 && rx < w.W-2 && ry >= 2 && ry < 15
}

// MouseOver is the hover of the subject buttons (ImageButton::draw's
// hovered state).
func (w *NewsWindow) MouseOver(px, py int) { w.hover = w.buttonAt(px, py) }

// HandleClick is Common::onMouseUp.
func (w *NewsWindow) HandleClick(px, py int) {
	if w.onClose(px, py) {
		if w.OnDismiss != nil {
			w.OnDismiss()
		}
		return
	}
	if i := w.buttonAt(px, py); i >= 0 && w.OnSubject != nil {
		w.OnSubject(i)
	}
}

// WidgetTooltip is the close button's tooltip.
func (w *NewsWindow) WidgetTooltip(px, py int) string {
	if w.onClose(px, py) {
		return i18n.Str(i18n.TooltipCloseWindow)
	}
	return ""
}

// Draw is Common::draw.
func (w *NewsWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	black := pal[paletteBlack0]
	colour := uint8(colourMutedR)
	if w.Paper {
		colour = colourGrey
	}
	prim, err := g1.GetPaletteMap(int(colour))
	if err != nil {
		return
	}
	old := w.Paper && w.Year < 1945
	sprite := func(rt *gfx8.RenderTarget, id, x, y int, m []uint8) {
		if id <= 0 || id >= len(g1.Elements) {
			return
		}
		if sp, err := gfx8.Decode(&g1.Elements[id]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, x, y, gfx8.OpFor(sp, m != nil, false), m, nil)
		}
	}
	// The widgets: the frame (NewsPanel's page, or the window's translucent
	// background and the Panel), the close button, the viewports' insets.
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		if w.Paper {
			var m []uint8
			l, r := imgNewsNewLeft, imgNewsNewRight
			if old {
				l, r = imgNewsOldLeft, imgNewsOldRight
				m, _ = g1.GetPaletteMap(extTranslucentBrown1)
			}
			sprite(rt, l, 0, 0, m)
			sprite(rt, r, w.W/2, 0, m)
		} else {
			m34, _ := g1.GetPaletteMap(extColourUnk34)
			gfx8.FillRectTranslucent(rt, 0, 0, w.W-1, w.H-1, m34)
			gfx8.FillRectInset(rt, 0, 0, w.W-1, w.H-1, translucentInset(colour), 0)
		}
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		if !w.Paper { // Viewport::draw's TextBox; the paper's viewport covers its own
			for i, s := range w.Subjects {
				if s.Kind == NewsSubjectView {
					l, t, r, b := w.viewport(i)
					gfx8.FillRectInset(rt, l, t, r, b, translucentInset(colour), gfx8.InsetBorderInset|gfx8.InsetFillDarker)
				}
			}
		}
	})
	// The viewports: the paper's on the widget, two wider and 64 high; the
	// panel's a pixel in, the widget's width and 62 high — each drawn by its
	// widget, which clips it (the viewport's last columns are not seen).
	for i := range w.Subjects {
		s := &w.Subjects[i]
		if s.Kind != NewsSubjectView || s.View == nil {
			continue
		}
		l, t, r, b := w.viewport(i)
		x, y, vw, vh := l+1, t+1, r-l+1, 62
		if w.Paper {
			x, y, vw, vh = l, t, r-l+1+2, 64
		}
		if rt := s.View(vw, vh); rt != nil {
			rt = cropTarget(rt, min(vw, r+1-x), min(vh, b+1-y))
			drawViewport(screen, &w.viewImg[i], rt, w.X+x, w.Y+y, s.Labels, false, 0, 0, nil)
		}
	}
	// A hovered subject button: ImageButton's draw_3, the translucent inset.
	if w.hover >= 0 && w.hover < 2 && w.Subjects[w.hover].Selectable {
		l, t, r, b := w.button(w.hover)
		draw8(screen, w.X+l, w.Y+t, r-l+1, b-t+1, func(rt *gfx8.RenderTarget) {
			gfx8.FillRectInset(rt, 0, 0, r-l, b-t, translucentInset(colour), 0)
		})
	}
	// The text and the date.
	if w.Paper {
		DrawStringCentredWrapped(screen, w.Text, G1FontLarge, w.X+w.W/2, w.Y+38, 352, black)
		dx, dy := 1, 1
		if old {
			dx, dy = 4, 5
		}
		DrawTextFont(screen, w.Date, G1FontSmall, w.X+dx, w.Y+dy, black)
	} else {
		DrawStringCentredWrapped(screen, w.Text, G1FontMediumBold, w.X+w.W/2, w.Y+17, 338, black)
	}
	// drawNewsSubjectImages: an owner's face, a turning vehicle.
	if w.Paper {
		for i := range w.Subjects {
			s := &w.Subjects[i]
			l, t, r, b := w.viewport(i)
			switch s.Kind {
			case NewsSubjectFace:
				fx, fy := (l+r)/2-31, (t+b)/2-31
				draw8(screen, w.X+fx, w.Y+fy, 64, 64, func(rt *gfx8.RenderTarget) {
					m, _ := g1.GetPaletteMap(int(s.FaceColour))
					sprite(rt, s.Face, 0, 0, m)
					if s.Jailed {
						sprite(rt, imgOwnerJailed, 0, 0, nil)
					}
				})
			case NewsSubjectVehicle:
				if s.Vehicle == nil {
					continue
				}
				// The clip (x + 1, y + 1, width − 2, height − 2); in it the
				// vehicle at (midX, 35) — the widget's midX, as upstream.
				cw, ch := r-l+1-2, b-t+1-2
				yaw, roll := uint8(w.PreviewFrame&0x3F), uint8(((w.PreviewFrame+2)/4)&0x3F)
				draw8(screen, w.X+l+1, w.Y+t+1, cw, ch, func(rt *gfx8.RenderTarget) {
					DrawVehicleOverview(rt, s.Vehicle, (l+r)/2, 35+s.VehicleYOffset, yaw, roll, s.Primary, s.Secondary)
				})
			}
		}
	}
	// The paper's washes: brown over the old page and its first picture
	// (drawEarlyNews), grey over the pictures until 1985 (drawMiddleNews).
	wash := func(x, y, cw, ch, ext int) {
		MarkForeign(image.Rect(w.X+x, w.Y+y, w.X+x+cw, w.Y+y+ch)) // the text is on the screen, not in the 8-bit copy
		draw8(screen, w.X+x, w.Y+y, cw, ch, func(rt *gfx8.RenderTarget) {
			m, _ := g1.GetPaletteMap(ext)
			gfx8.FillRectTranslucent(rt, 0, 0, cw-1, ch-1, m)
		})
	}
	switch {
	case old:
		wash(3, 5, w.W-6, w.H, extTranslucentBrown1)
		l, t, r, b := w.viewport(0)
		wash(l, t, r-l+1, b-t+1, extTranslucentBrown1)
	case w.Paper && w.Year < 1985:
		for i := range w.Subjects {
			if !w.Subjects[i].Present {
				continue
			}
			l, t, r, b := w.viewport(i)
			wash(l, t, r-l+1+1, b-t+1+1, extTranslucentGrey1)
		}
	}
	// drawViewportString: black_tiny_font, centred on the button, clipped to
	// its width − 1, 7 above its bottom.
	for i := range w.Subjects {
		c := w.Subjects[i].Caption
		if c == "" {
			continue
		}
		l, _, r, b := w.button(i)
		c = ClipString(c, G1FontSmall, r-l+1-1)
		DrawTextFont(screen, c, G1FontSmall, w.X+(l+r)/2-MeasureTextFont(c, G1FontSmall)/2, w.Y+b-7, black)
	}
}

// cropTarget is the top-left w × h of a render target.
func cropTarget(rt *gfx8.RenderTarget, w, h int) *gfx8.RenderTarget {
	if w >= rt.Width && h >= rt.Height || w <= 0 || h <= 0 {
		return rt
	}
	w, h = min(w, rt.Width), min(h, rt.Height)
	out := gfx8.NewRenderTarget(w, h)
	for y := range h {
		copy(out.Bits[y*w:(y+1)*w], rt.Bits[y*rt.Width:y*rt.Width+w])
	}
	return out
}

// DrawVehicleOverview is VehicleDraw.cpp drawVehicleOverview (0x004B777B)
// for a vehicle object at (x, y) of rt (y with its display offset): the
// bogies, then the bodies — back to front for a yaw of 8 to 39 —, each
// (the vehicle's half length − its distance) along the yaw, a quarter of it,
// as the screen sees it (gameToScreen, rotation 0).
func DrawVehicleOverview(rt *gfx8.RenderTarget, v *objects.VehicleObject, x, y int, yaw, roll uint8, primary, secondary uint8) {
	g1 := globalG1Font.g1
	items, total := v.DrawItems(yaw, roll)
	mid := total / 2
	pal, _ := g1.GetPaletteMap2(int(primary), int(secondary))
	draw := func(it objects.DrawItem) {
		vx, vy := olmath.ComputeXYVector(mid-it.Dist, yaw)
		vx, vy = vx/4, vy/4
		id := int(v.ImageOffset + it.Image)
		if id <= 0 || id >= len(g1.Elements) {
			return
		}
		if sp, err := gfx8.Decode(&g1.Elements[id]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, x+int(vy-vx), y+int((vx+vy)>>1), gfx8.OpFor(sp, pal != nil, false), pal, nil)
		}
	}
	for _, it := range items {
		if !it.IsBody {
			draw(it)
		}
	}
	if yaw >= 8 && yaw < 40 {
		for i := len(items) - 1; i >= 0; i-- {
			if items[i].IsBody {
				draw(items[i])
			}
		}
		return
	}
	for _, it := range items {
		if it.IsBody {
			draw(it)
		}
	}
}

// NewsTicker is the news ticker — Ui/Windows/News/Ticker.cpp: a message
// written out two characters at a time over the time panel's date and speed
// buttons (111 × 26 at the screen's right − 138, bottom − 26), in the small
// font on white (general news) or red (advice), four lines scrolling. The
// findAt of the mouse looks through it (ignoreInFindAt), so a click there is
// the time panel's; while the mouse is over it, and 12 updates after, it
// shows nothing — the date can be read.
//
// PARTIAL: the ticker sound (no sound effects in goloco). Its onMouseUp
// (the message opened as a window) is not ported: no press reaches a window
// findAt ignores.
type NewsTicker struct {
	WinBase
	General bool   // isGeneralNews: white, else red
	Text    string // the message (Loco codes as runes)

	numCharsToDisplay uint16 // _nState.numCharsToDisplay
	var852            int
	screenW, screenH  int // the screen it was last placed on
}

// NewNewsTicker is NewsWindow::open's ticker: at (width − 138, height − 25),
// its last row below the screen — where the real game shows it; onResize's
// height − 26 comes only with a new screen size.
func NewNewsTicker(sw, sh int) *NewsTicker {
	return &NewsTicker{WinBase: WinBase{WinType: WindowTypeNews, X: sw - 138, Y: sh - 25, W: tickerW, H: tickerH,
		WinFlags: WindowFlagStickToFront | WindowFlagIgnoreInFindAt, NoCaption: true}, screenW: sw, screenH: sh}
}

// HandleClick: nothing reaches it (see the type).
func (w *NewsTicker) HandleClick(px, py int) {}

// Update is Ticker's onUpdate, once per interface update — under the mouse
// the pause of 12 updates starts again; unpaused it counts down, and at 0
// two more quarter characters show — and its onResize when the screen's size
// has changed.
func (w *NewsTicker) Update(sw, sh int, underMouse, paused bool) {
	if sw != w.screenW || sh != w.screenH {
		w.screenW, w.screenH = sw, sh
		w.X, w.Y = sw-tickerW-27, sh-tickerH
	}
	if underMouse {
		w.var852 = 12
	}
	if w.var852 != 0 && !paused {
		w.var852--
	}
	if w.var852 == 0 && !paused {
		w.numCharsToDisplay += 2
	}
}

// tickerText is draw's buffer: the message, a newline as two spaces.
func tickerText(text string) string {
	var out []rune
	for _, r := range text {
		if r == controlNewline {
			out = append(out, ' ', ' ')
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

// Draw is Ticker's draw: nothing during the mouse's pause; else the window
// cleared to shade 5 of white or mutedDarkRed and the text so far — once all
// of it shows, bit 15 of the count stops the writing.
func (w *NewsTicker) Draw(screen *ebiten.Image) {
	if globalG1Font == nil || w.var852 != 0 {
		return
	}
	g1 := globalG1Font.g1
	colour := colourWhite
	if !w.General {
		colour = colourMutedR
	}
	cm, err := g1.GetPaletteMap(colour)
	if err != nil {
		return
	}
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRect(rt, 0, 0, w.W-1, w.H-1, gfx8.Shade(cm, 5))
	})
	text := tickerText(w.Text)
	if int(w.numCharsToDisplay>>2) > len([]rune(w.Text))+1 { // i: the message's characters and its terminator
		w.numCharsToDisplay |= 1 << 15
	}
	r := image.Rect(w.X, w.Y, w.X+w.W, w.Y+w.H)
	clip, ok := screen.SubImage(r).(*ebiten.Image)
	if !ok {
		return
	}
	MarkForeign(r)
	DrawStringTicker(clip, text, G1FontSmall, w.X+55, w.Y, 4, int(w.numCharsToDisplay&^(1<<15))>>2, 109, g1.GetPalette()[paletteBlack0])
}
