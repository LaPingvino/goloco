package ui

// DrawCheckbox draws a checkbox widget
func DrawCheckbox(window *Window, dc *DrawingContext, widget *Widget) {
	if widget == nil || dc == nil {
		return
	}

	// Calculate absolute position
	x := window.X + widget.X
	y := window.Y + widget.Y

	// Draw checkbox box
	dc.FillRectInset(x, y, 12, 12, widget.Colour, 1)

	// Draw check mark if activated
	if widget.Activated {
		// Simple X mark
		dc.FillRect(x+3, y+5, 6, 2, 0) // horizontal
		dc.FillRect(x+5, y+3, 2, 6, 0) // vertical
	}

	// Draw label text
	if widget.Text != "" {
		dc.DrawString(x+16, y+2, widget.Text, uint8(widget.Colour.Color))
	}
}
