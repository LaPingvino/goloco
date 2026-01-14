package ui

// DrawPanel draws a panel widget (simple filled rectangle)
func DrawPanel(window *Window, dc *DrawingContext, widget *Widget) {
	if widget == nil || dc == nil {
		return
	}

	// Calculate absolute position
	x := window.X + widget.X
	y := window.Y + widget.Y

	// Draw panel background
	dc.FillRect(x, y, widget.Width, widget.Height, uint8(widget.Colour.Color))
}
