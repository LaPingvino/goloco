package ui

// DrawButton draws a button widget
func DrawButton(window *Window, dc *DrawingContext, widget *Widget) {
	if widget == nil || dc == nil {
		return
	}

	// Calculate absolute position
	x := window.X + widget.X
	y := window.Y + widget.Y

	// Draw button background with inset if activated
	flags := uint8(0)
	if widget.Activated {
		flags = 1 // border inset
	}

	dc.FillRectInset(x, y, widget.Width, widget.Height, widget.Colour, flags)

	// Draw text if present
	if widget.Text != "" {
		// Center text in button
		textX := x + widget.Width/2
		textY := y + widget.Height/2 - 4 // approximate text height
		dc.DrawStringCentered(textX, textY, widget.Text, uint8(widget.Colour.Color))
	}
}
