package ui

// DrawFrame draws a frame/border widget
func DrawFrame(window *Window, dc *DrawingContext, widget *Widget) {
	if widget == nil || dc == nil {
		return
	}

	// Calculate absolute position
	x := window.X + widget.X
	y := window.Y + widget.Y

	// Draw frame with inset border
	dc.FillRectInset(x, y, widget.Width, widget.Height, widget.Colour, 1)
}
