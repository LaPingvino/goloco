package ui

// DrawCaption draws a window caption/title bar
func DrawCaption(window *Window, dc *DrawingContext, widget *Widget) {
	if widget == nil || dc == nil {
		return
	}

	// Calculate absolute position
	x := window.X + widget.X
	y := window.Y + widget.Y

	// Draw caption background
	dc.FillRect(x, y, widget.Width, widget.Height, uint8(widget.Colour.Color))

	// Draw caption text
	if widget.Text != "" {
		// Left-aligned text with some padding
		textX := x + 4
		textY := y + widget.Height/2 - 4
		dc.DrawString(textX, textY, widget.Text, 255) // white text
	}
}
