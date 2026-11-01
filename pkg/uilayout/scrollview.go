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
// pixel height (Widget::height() = h − 1); y is relative to its top.

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
func vMaxOffset(h, contentH int) int { return max(contentH-(h-1-2), 0) }

// VScrollClick is vButtonTop/Bottom and vTrackTop/Bottom: the new offset.
func VScrollClick(part string, h, contentH, offset int) int {
	page := h - 1 - 2 // trackHeight = height() − 2
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

// VScrollDrag is verticalDragFollow: the thumb dragged by dy moves the
// content by dy × contentHeight / trackHeight (trackHeight = height() − 2 −
// 2 thumb sizes), clamped.
func VScrollDrag(h, contentH, offset, dy int) int {
	thumb := ScrollbarSize - ScrollbarMargin // kThumbSize
	track := h - 1 - 2 - 2*thumb
	if track <= 0 {
		return offset
	}
	return min(max(offset+dy*contentH/track, 0), vMaxOffset(h, contentH))
}
