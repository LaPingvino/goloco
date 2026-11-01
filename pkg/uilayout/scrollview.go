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
