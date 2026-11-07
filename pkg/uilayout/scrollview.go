package uilayout

// ScrollView constants (Ui/ScrollView.h).
const (
	ScrollbarSize   = 11
	ScrollbarMargin = 1
	ThumbSize       = ScrollbarSize - ScrollbarMargin
	MinThumbSize    = ThumbSize * 2
)

// VThumb is calculateThumbSizeAndPosition for the vertical bar of a scroll
// view widgetH tall (no horizontal bar) with content contentH tall scrolled to
// offset: the thumb's top and bottom, relative to the widget. The uint16
// truncations of upstream are kept.
func VThumb(widgetH, contentH, offset int) (top, bottom int) {
	button := ScrollbarSize
	bar := uint16(widgetH - button*2)
	size := uint16(max(int(uint16(float32(bar)*float32(widgetH)/float32(contentH))), MinThumbSize))
	dist := int(bar) - int(size)
	pos := uint16(float32(dist) * (float32(offset) / float32(contentH-widgetH)))
	top = button + int(pos)
	bottom = min(button+int(pos)+int(size), widgetH-button-1)
	return top, bottom
}

// ScrollView's vertical bar, as ScrollView.cpp handles it. h is the widget's
// pixel height (Widget::height(): bottom − top + 1); y is relative to its top.

// kButtonClickStep: a button click moves the content 3 px.
const scrollButtonStep = 3

// VScrollPart is getPart for a point on the vertical bar: "up", "down",
// "trackTop", "thumb" or "trackBottom".
func VScrollPart(y, h, contentH, offset int) string {
	switch {
	case y < ScrollbarMargin+ScrollbarSize:
		return "up"
	case y >= h-ScrollbarMargin-ScrollbarSize:
		return "down"
	}
	top, bot := VThumb(h, contentH, offset)
	switch {
	case y < top:
		return "trackTop"
	case y < bot:
		return "thumb"
	}
	return "trackBottom"
}

// vMaxOffset is the furthest the content scrolls: contentHeight − (height() − 2).
func vMaxOffset(h, contentH int) int { return max(contentH-(h-2), 0) }

// VScrollClick is vButtonTop/Bottom and vTrackTop/Bottom: the new offset.
func VScrollClick(part string, h, contentH, offset int) int {
	page := h - 2 // trackHeight = height() − 2
	switch part {
	case "up":
		return max(offset-scrollButtonStep, 0)
	case "down":
		return min(offset+scrollButtonStep, vMaxOffset(h, contentH))
	case "trackTop":
		return max(offset-page, 0)
	case "trackBottom":
		return min(offset+page, vMaxOffset(h, contentH))
	}
	return offset
}

// VScrollDrag is verticalFollow (scrollLeftContinue on the thumb): the thumb
// dragged by dy moves the content by dy × contentHeight / trackHeight
// (trackHeight = height() − 2 thumb sizes), clamped to contentHeight −
// (height() − 2). (verticalDragFollow, with 2 less, is the right-button
// drag scroll.)
func VScrollDrag(h, contentH, offset, dy int) int {
	thumb := ScrollbarSize - ScrollbarMargin // kThumbSize
	track := h - 2*thumb
	if track <= 0 {
		return offset
	}
	return min(max(offset+dy*contentH/track, 0), vMaxOffset(h, contentH))
}

// The horizontal bar (no vertical bar beside it): w is the widget's pixel
// width (Widget::width()), x relative to its left, content contentW wide.

// HThumb is calculateThumbSizeAndPosition for the horizontal bar.
func HThumb(widgetW, contentW, offset int) (left, right int) {
	return VThumb(widgetW, contentW, offset)
}

// HScrollPart is getPart on the horizontal bar: "left", "right",
// "trackLeft", "thumb" or "trackRight".
func HScrollPart(x, w, contentW, offset int) string {
	switch {
	case x < ScrollbarSize:
		return "left"
	case x >= w-1-ThumbSize:
		return "right"
	}
	l, r := HThumb(w, contentW, offset)
	switch {
	case x < l:
		return "trackLeft"
	case x > r:
		return "trackRight"
	}
	return "thumb"
}

// hMaxOffset is contentWidth − (width() − 2), at least 0.
func hMaxOffset(w, contentW int) int { return max(contentW-(w-2), 0) }

// HScrollClick is hButtonLeft/Right and hTrackLeft/Right: the new offset.
func HScrollClick(part string, w, contentW, offset int) int {
	page := w - 2
	switch part {
	case "left":
		return max(offset-scrollButtonStep, 0)
	case "right":
		return min(offset+scrollButtonStep, hMaxOffset(w, contentW))
	case "trackLeft":
		return max(offset-page, 0)
	case "trackRight":
		return min(offset+page, hMaxOffset(w, contentW))
	}
	return offset
}

// HScrollDrag is horizontalFollow (scrollLeftContinue on the thumb): dx ×
// contentWidth / trackWidth (width() − 2 thumb sizes), clamped to 0 …
// contentWidth − (width() − 2).
func HScrollDrag(w, contentW, offset, dx int) int {
	track := w - 2*ThumbSize
	if track <= 0 {
		return offset
	}
	return min(max(offset+dx*contentW/track, 0), hMaxOffset(w, contentW))
}

// HScrollLatest is Finances::scrollToLatestData: the last full page.
func HScrollLatest(w, contentW int) int { return hMaxOffset(w, contentW) }
