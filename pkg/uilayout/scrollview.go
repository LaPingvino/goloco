package uilayout

// ScrollView constants (Ui/ScrollView.h).
const (
	ScrollbarSize   = 11
	ScrollbarMargin = 1
	ThumbSize       = ScrollbarSize - ScrollbarMargin
	MinThumbSize    = ThumbSize * 2
)

// thumb is ScrollView.cpp calculateThumbSizeAndPosition (l. 292) for one
// bar: the thumb's start and end, relative to the widget, for a widget
// widgetSize long (Widget::width() / height()) whose content is contentSize
// long (contentWidth / contentHeight) scrolled to offset; otherBar is the
// other bar's visible flag (the corner it takes off both lengths). The
// uint16 truncations of upstream are kept.
func thumb(otherBar bool, widgetSize, contentSize, offset int) (start, end int) {
	const button = ScrollbarSize // kScrollButtonSize
	ws := uint16(widgetSize)
	bar := ws - button*2
	if otherBar {
		ws -= ScrollbarSize
		bar -= ScrollbarSize
	}
	size := max(uint16(int32(float32(int(bar)*int(ws))/float32(contentSize))), MinThumbSize)
	dist := int(bar) - int(size)
	pos := uint16(int32(float32(dist) * (float32(offset) / float32(contentSize-int(ws)))))
	start = button + int(pos)
	end = min(button+int(pos)+int(size), int(ws)-button-1)
	return start, end
}

// VThumb is calculateThumbSizeAndPosition for the vertical bar of a scroll
// view widgetH tall (no horizontal bar) with content contentH tall scrolled to
// offset: the thumb's top and bottom, relative to the widget.
func VThumb(widgetH, contentH, offset int) (top, bottom int) {
	return thumb(false, widgetH, contentH, offset)
}

// kButtonClickStep: a button click moves the content 3 px.
const scrollButtonStep = 3

// Part is Ui::ScrollPart: what of a scroll view a point is on.
type Part uint8

const (
	PartNone Part = iota
	PartView
	PartHLeft       // hscrollbarButtonLeft
	PartHRight      // hscrollbarButtonRight
	PartHTrackLeft  // hscrollbarTrackLeft
	PartHTrackRight // hscrollbarTrackRight
	PartHThumb      // hscrollbarThumb
	PartVTop        // vscrollbarButtonTop
	PartVBottom     // vscrollbarButtonBottom
	PartVTrackTop   // vscrollbarTrackTop
	PartVTrackBottom
	PartVThumb
)

// Bars is a scroll view as ScrollView.cpp reads it: the widget's
// Widget::width() / height(), its scroll area's contentWidth /
// contentHeight (getScrollSize + 1, as Window::updateScrollWidgets sets
// them) and its Scrollbars (hscrollbarVisible / vscrollbarVisible: the
// widget's flags, set whether the content overflows or not).
type Bars struct {
	W, H               int
	ContentW, ContentH int
	HBar, VBar         bool
}

// HShown / VShown: ScrollView::draw and getPart show a bar only when its
// flag is set and the content is longer than the widget.
func (b Bars) HShown() bool { return b.HBar && b.ContentW > b.W }
func (b Bars) VShown() bool { return b.VBar && b.ContentH > b.H }

// HThumb / VThumb are updateThumbs (0x004CA1ED): the thumbs' hThumbLeft /
// hThumbRight and vThumbTop / vThumbBottom for the offsets.
func (b Bars) HThumb(offX int) (left, right int) { return thumb(b.VBar, b.W, b.ContentW, offX) }
func (b Bars) VThumb(offY int) (top, bottom int) { return thumb(b.HBar, b.H, b.ContentH, offY) }

// axisPart is getPart's walk along one bar (l. 194–233 for the horizontal,
// l. 235–275 for the vertical): pos and size along the bar, relative to the
// widget; otherBar takes the corner's deadzone off its far end.
func axisPart(pos, size int, otherBar bool, thumbStart, thumbEnd int, first, second, trackFirst, trackSecond, th Part) Part {
	if pos < ScrollbarSize {
		return first
	}
	far := size - 1 // right / bottom (widget->right + window.x, relative)
	if otherBar {
		far -= ScrollbarSize
	}
	switch {
	case pos >= far: // within the deadzone
		return PartNone
	case pos >= far-ThumbSize:
		return second
	case pos < thumbStart:
		return trackFirst
	case pos > thumbEnd:
		return trackSecond
	}
	return th
}

// Part is ScrollView::getPart (0x004C8EF0, l. 167) at x, y relative to the
// widget's left and top: the part, and on the view part the point in the
// content (scrollviewLoc: 1 px in, plus the offsets).
func (b Bars) Part(x, y, offX, offY int) (p Part, cx, cy int) {
	switch {
	case b.HShown() && y >= b.H-1-ScrollbarSize:
		l, r := b.HThumb(offX)
		return axisPart(x, b.W, b.VBar, l, r, PartHLeft, PartHRight, PartHTrackLeft, PartHTrackRight, PartHThumb), x, y
	case b.VShown() && x >= b.W-1-ScrollbarSize:
		t, bt := b.VThumb(offY)
		return axisPart(y, b.H, b.HBar, t, bt, PartVTop, PartVBottom, PartVTrackTop, PartVTrackBottom, PartVThumb), x, y
	}
	cx, cy = x-1, y-1
	if cx < 0 || cy < 0 {
		return PartNone, cx, cy
	}
	return PartView, cx + offX, cy + offY
}

// track is the page a track click moves (width() / height() − 2, less the
// other bar), as int16.
func track(size int, otherBar bool) int {
	t := int16(size - 2)
	if otherBar {
		t -= ScrollbarSize
	}
	return int(t)
}

// axisClick is a button or track press on one bar: hButtonLeft / Right
// (0x004C894F / 0x004C89AE), hTrackLeft / Right (0x004C8A36 / 0x004C8AA6)
// and their vertical twins vButtonTop / Bottom (verticalNudgeUp / Down),
// vTrackTop / Bottom: the new offset.
func axisClick(back, fwd, pageBack, pageFwd bool, size, content int, otherBar bool, offset int) int {
	t := track(size, otherBar)
	maxOffset := max(content-t, 0)
	switch {
	case back:
		return max(offset-scrollButtonStep, 0)
	case fwd:
		return int(min(int16(offset+scrollButtonStep), int16(maxOffset)))
	case pageBack:
		return max(offset-t, 0)
	case pageFwd:
		return int(min(int16(offset+t), int16(maxOffset)))
	}
	return offset
}

// Click is scrollLeftBegin's (0x004C8689) button and track handlers for the
// part: the new offsets (the view part and the thumbs move nothing here).
func (b Bars) Click(p Part, offX, offY int) (int, int) {
	switch p {
	case PartHLeft, PartHRight, PartHTrackLeft, PartHTrackRight:
		offX = axisClick(p == PartHLeft, p == PartHRight, p == PartHTrackLeft, p == PartHTrackRight, b.W, b.ContentW, b.VBar, offX)
	case PartVTop, PartVBottom, PartVTrackTop, PartVTrackBottom:
		offY = axisClick(p == PartVTop, p == PartVBottom, p == PartVTrackTop, p == PartVTrackBottom, b.H, b.ContentH, b.HBar, offY)
	}
	return offX, offY
}

// follow is horizontalFollow (0x004C87E1) / verticalFollow (0x004C8898):
// the thumb dragged by delta moves the content by delta × content / track
// (the widget less two thumb sizes and the other bar), as int16, kept within
// 0 … content − (the widget less 2 margins and the other bar).
func follow(size, content int, otherBar bool, offset, delta int) int {
	t := uint16(size - ThumbSize*2)
	if otherBar {
		t -= ScrollbarSize
	}
	if t == 0 {
		return offset
	}
	newOffset := int16(offset + int(int16(delta))*content/int(t))
	frame := size - ScrollbarMargin*2
	if otherBar {
		frame -= ScrollbarSize
	}
	maxOffset := max(int16(content-frame), 0)
	return int(min(max(newOffset, 0), maxOffset))
}

// HFollow / VFollow are horizontalFollow / verticalFollow on the bars.
func (b Bars) HFollow(offX, dx int) int { return follow(b.W, b.ContentW, b.VBar, offX, dx) }
func (b Bars) VFollow(offY, dy int) int { return follow(b.H, b.ContentH, b.HBar, offY, dy) }

// ScrollHold is a left press on a scroll view held (Input::State::scrollLeft):
// the part pressed (_currentScrollArea), whether it shows pressed (the scroll
// area's …Pressed flags; one part at a time), and the pointer's last place
// for a thumb (the tooltip mouse location MouseInput sets at the press).
type ScrollHold struct {
	Part         Part // PartNone: nothing held
	Pressed      bool
	LastX, LastY int
}

// Begin is stateNormalLeft's scrollview case (MouseInput.cpp l. 1333) and
// scrollLeftBegin: x, y relative to the widget, px, py the pointer. The new
// offsets, and on the view part the content point for callScrollMouseDown.
func (s *ScrollHold) Begin(b Bars, x, y, px, py, offX, offY int) (nx, ny int, view bool, cx, cy int) {
	p, cx, cy := b.Part(x, y, offX, offY)
	s.Part, s.LastX, s.LastY = p, px, py
	s.Pressed = p == PartHLeft || p == PartHRight || p == PartVTop || p == PartVBottom
	nx, ny = b.Click(p, offX, offY)
	return nx, ny, p == PartView, cx, cy
}

// Release is stateScrollLeft's leftReleased: the state resets and
// clearPressedButtons (0x004C72ED).
func (s *ScrollHold) Release() { s.Part, s.Pressed = PartNone, false }

// Continue is stateScrollLeft for a frame the button stays down
// (MouseInput.cpp l. 642): with the pointer off the pressed widget,
// clearPressedButtons alone; on it scrollLeftContinue (0x004C7236): a thumb
// follows the pointer from its last place; another part than the one held
// clears the pressed buttons; on the held part the view drags
// (callScrollMouseDrag: view and its content point), a button steps again and
// a track does nothing. onWidget: the pointer is on this window's widget.
func (s *ScrollHold) Continue(b Bars, onWidget bool, x, y, px, py, offX, offY int) (nx, ny int, view bool, cx, cy int) {
	if s.Part == PartNone {
		return offX, offY, false, 0, 0
	}
	if !onWidget {
		s.Pressed = false
		return offX, offY, false, 0, 0
	}
	switch s.Part {
	case PartHThumb:
		dx := px - s.LastX
		s.LastX = px
		s.Pressed = true
		return b.HFollow(offX, dx), offY, false, 0, 0
	case PartVThumb:
		dy := py - s.LastY
		s.LastY = py
		s.Pressed = true
		return offX, b.VFollow(offY, dy), false, 0, 0
	}
	p, cx, cy := b.Part(x, y, offX, offY)
	if p != s.Part {
		s.Pressed = false
		return offX, offY, false, 0, 0
	}
	switch p {
	case PartView:
		return offX, offY, true, cx, cy
	case PartHLeft, PartHRight, PartVTop, PartVBottom:
		s.Pressed = true
		nx, ny = b.Click(p, offX, offY)
		return nx, ny, false, 0, 0
	}
	return offX, offY, false, 0, 0
}

// PressedPart is the part drawn pressed (PartNone: none).
func (s *ScrollHold) PressedPart() Part {
	if s.Pressed {
		return s.Part
	}
	return PartNone
}

// ScrollView's vertical bar alone (Scrollbars::vertical), as the list
// windows use it: h is the widget's pixel height (Widget::height(): bottom −
// top + 1), y relative to its top, the content contentH tall.

// vBars is a vertical-only scroll view whose bar shows.
func vBars(h, contentH int) Bars { return Bars{H: h, ContentH: contentH, VBar: true} }

var vPartNames = map[Part]string{PartVTop: "up", PartVBottom: "down", PartVTrackTop: "trackTop", PartVThumb: "thumb", PartVTrackBottom: "trackBottom"}

// VScrollPart is getPart for a point on the vertical bar: "up", "down",
// "trackTop", "thumb", "trackBottom" (or "" past the bar's end).
func VScrollPart(y, h, contentH, offset int) string {
	t, bt := VThumb(h, contentH, offset)
	return vPartNames[axisPart(y, h, false, t, bt, PartVTop, PartVBottom, PartVTrackTop, PartVTrackBottom, PartVThumb)]
}

var vParts = map[string]Part{"up": PartVTop, "down": PartVBottom, "trackTop": PartVTrackTop, "trackBottom": PartVTrackBottom}

// VScrollClick is vButtonTop/Bottom and vTrackTop/Bottom: the new offset.
func VScrollClick(part string, h, contentH, offset int) int {
	_, y := vBars(h, contentH).Click(vParts[part], 0, offset)
	return y
}

// VScrollDrag is verticalFollow (scrollLeftContinue on the thumb).
// (verticalDragFollow, with 2 less, is the right-button drag scroll.)
func VScrollDrag(h, contentH, offset, dy int) int {
	return vBars(h, contentH).VFollow(offset, dy)
}

// The horizontal bar alone (Scrollbars::horizontal): w is the widget's pixel
// width (Widget::width()), x relative to its left, content contentW wide.

func hBars(w, contentW int) Bars { return Bars{W: w, ContentW: contentW, HBar: true} }

// HThumb is calculateThumbSizeAndPosition for the horizontal bar.
func HThumb(widgetW, contentW, offset int) (left, right int) {
	return thumb(false, widgetW, contentW, offset)
}

var hPartNames = map[Part]string{PartHLeft: "left", PartHRight: "right", PartHTrackLeft: "trackLeft", PartHThumb: "thumb", PartHTrackRight: "trackRight"}

// HScrollPart is getPart on the horizontal bar: "left", "right",
// "trackLeft", "thumb" or "trackRight" ("" past the bar's end).
func HScrollPart(x, w, contentW, offset int) string {
	l, r := HThumb(w, contentW, offset)
	return hPartNames[axisPart(x, w, false, l, r, PartHLeft, PartHRight, PartHTrackLeft, PartHTrackRight, PartHThumb)]
}

var hParts = map[string]Part{"left": PartHLeft, "right": PartHRight, "trackLeft": PartHTrackLeft, "trackRight": PartHTrackRight}

// HScrollClick is hButtonLeft/Right and hTrackLeft/Right: the new offset.
func HScrollClick(part string, w, contentW, offset int) int {
	x, _ := hBars(w, contentW).Click(hParts[part], offset, 0)
	return x
}

// HScrollDrag is horizontalFollow (scrollLeftContinue on the thumb).
func HScrollDrag(w, contentW, offset, dx int) int {
	return hBars(w, contentW).HFollow(offset, dx)
}

// HScrollLatest is Finances::scrollToLatestData: the last full page
// (contentWidth − (width() − 2), at least 0).
func HScrollLatest(w, contentW int) int { return max(contentW-(w-2), 0) }

// VScrollHold is a press on a vertical bar held down (the stateScrollLeft
// input state): its part (_currentScrollArea), whether that part shows
// pressed, and the pointer's last y for the thumb.
type VScrollHold struct {
	Part    string // "" when nothing is held
	Pressed bool
	LastY   int
}

// Begin is scrollLeftBegin on the bar: the part pressed (y relative to the
// widget, py the pointer's), and the offset after vButton… / vTrack….
func (s *VScrollHold) Begin(y, py, h, contentH, offset int) int {
	part := VScrollPart(y, h, contentH, offset)
	s.Part, s.LastY = part, py
	s.Pressed = part == "up" || part == "down"
	if part == "thumb" {
		return offset
	}
	return VScrollClick(part, h, contentH, offset)
}

// Continue is stateScrollLeft for a tick: released, the hold ends
// (clearPressedButtons); the thumb follows the pointer (verticalFollow); a
// button over its own part (onBar, y relative to the widget) steps again,
// off it it shows unpressed; the track does nothing more.
func (s *VScrollHold) Continue(held, onBar bool, y, py, h, contentH, offset int) int {
	if s.Part == "" {
		return offset
	}
	if !held {
		s.Part, s.Pressed = "", false
		return offset
	}
	switch s.Part {
	case "thumb":
		s.Pressed = true
		dy := py - s.LastY
		s.LastY = py
		if dy != 0 {
			return VScrollDrag(h, contentH, offset, dy)
		}
	case "up", "down":
		s.Pressed = onBar && VScrollPart(y, h, contentH, offset) == s.Part
		if s.Pressed {
			return VScrollClick(s.Part, h, contentH, offset)
		}
	}
	return offset
}

// PressedPart is the part drawn pressed ("" none).
func (s *VScrollHold) PressedPart() string {
	if s.Pressed {
		return s.Part
	}
	return ""
}

// DragFollow is ScrollView::verticalDragFollow (0x004C8E2E) and
// horizontalDragFollow (0x004C8CFD) for one axis: the right-button drag in a
// scroll view (MouseInput's stateScrollRight). bar is the axis's own
// …scrollbarVisible flag (nothing moves without it), otherBar the other
// axis's; size is Widget::width() / height(), content the area's
// contentWidth / contentHeight (getScrollSize + 1). The track is size − 2 −
// 2 thumb sizes (minus the other bar), the content moves by delta × content
// / track, as int16, kept within 0 … content − (size − 2, minus the other
// bar). OpenLoco reference: src/OpenLoco/src/Ui/ScrollView.cpp.
func DragFollow(bar, otherBar bool, size, content, offset, delta int) int {
	if !bar {
		return offset
	}
	track := uint16(size - 2 - ThumbSize - ThumbSize)
	if otherBar {
		track -= ScrollbarSize
	}
	if track == 0 {
		return offset
	}
	newOffset := int16(offset + int(int16(delta))*content/int(track))
	frame := size - 2
	if otherBar {
		frame -= ScrollbarSize
	}
	maxOffset := max(int16(content-frame), 0)
	return int(min(max(newOffset, 0), maxOffset))
}
