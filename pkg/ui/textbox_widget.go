package ui

// DrawTextBox draws a text box widget
func DrawTextBox(window *Window, dc *DrawingContext, widget *Widget) {
	if widget == nil || dc == nil {
		return
	}

	// Calculate absolute position
	x := window.X + widget.X
	y := window.Y + widget.Y

	// Draw textbox background with inset
	dc.FillRectInset(x, y, widget.Width, widget.Height, widget.Colour, 1)

	// Draw text content
	if widget.Text != "" {
		dc.DrawString(x+4, y+4, widget.Text, uint8(widget.Colour.Color))
	}
}
