package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Input state
var (
	mouseX, mouseY    int
	mousePressed      bool
	mouseJustPressed  bool
	mouseJustReleased bool

	// Dragging state
	draggedWindow *Window
	dragStartX    int16
	dragStartY    int16
	dragOffsetX   int16
	dragOffsetY   int16
)

// UpdateInput updates the input state (call once per frame)
func UpdateInput() {
	// Get current mouse state
	mx, my := ebiten.CursorPosition()
	mouseX = mx
	mouseY = my

	// Detect mouse press/release
	wasPressed := mousePressed
	mousePressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	mouseJustPressed = mousePressed && !wasPressed
	mouseJustReleased = !mousePressed && wasPressed

	// Handle window dragging
	if mouseJustPressed {
		startDrag()
	}

	if mousePressed && draggedWindow != nil {
		updateDrag()
	}

	if mouseJustReleased {
		endDrag()
	}

	// Handle button clicks
	if mouseJustReleased {
		handleClick()
	}
}

// startDrag begins dragging if clicking on a caption bar
func startDrag() {
	x := int16(mouseX)
	y := int16(mouseY)

	// Find window at mouse position
	w := FindWindowAt(x, y)
	if w == nil {
		return
	}

	// Check if clicking on caption bar
	for i := range w.Widgets {
		widget := &w.Widgets[i]
		if widget.Type != WidgetTypeCaption {
			continue
		}

		// Check if mouse is over caption widget
		wx := w.X + widget.X
		wy := w.Y + widget.Y
		if x >= wx && x < wx+widget.Width && y >= wy && y < wy+widget.Height {
			// Start dragging
			draggedWindow = w
			dragStartX = x
			dragStartY = y
			dragOffsetX = x - w.X
			dragOffsetY = y - w.Y

			// Bring to front
			BringWindowToFront(w)
			break
		}
	}
}

// updateDrag updates window position during drag
func updateDrag() {
	if draggedWindow == nil {
		return
	}

	x := int16(mouseX)
	y := int16(mouseY)

	// Update window position
	draggedWindow.X = x - dragOffsetX
	draggedWindow.Y = y - dragOffsetY

	// Clamp to screen bounds
	if draggedWindow.X < 0 {
		draggedWindow.X = 0
	}
	if draggedWindow.Y < 0 {
		draggedWindow.Y = 0
	}

	draggedWindow.Invalidate()
}

// endDrag stops dragging
func endDrag() {
	draggedWindow = nil
}

// handleClick handles mouse clicks on widgets
func handleClick() {
	x := int16(mouseX)
	y := int16(mouseY)

	// Find window at mouse position
	w := FindWindowAt(x, y)
	if w == nil {
		return
	}

	// Check which widget was clicked
	for i := range w.Widgets {
		widget := &w.Widgets[i]
		if !widget.Visible {
			continue
		}

		// Check if mouse is over widget
		wx := w.X + widget.X
		wy := w.Y + widget.Y
		if x >= wx && x < wx+widget.Width && y >= wy && y < wy+widget.Height {
			handleWidgetClick(w, widget)
			break
		}
	}
}

// handleWidgetClick handles a click on a specific widget
func handleWidgetClick(window *Window, widget *Widget) {
	switch widget.Type {
	case WidgetTypeButton:
		// Toggle button activation (visual feedback)
		widget.Activated = !widget.Activated
		window.Invalidate()

		// Check for special button actions
		if widget.Text == "X" || widget.Text == "OK" || widget.Text == "Cancel" {
			// Close the window
			CloseWindow(window)
		}

	case WidgetTypeCheckbox:
		// Toggle checkbox
		widget.Activated = !widget.Activated
		window.Invalidate()

	case WidgetTypeTextBox:
		// Focus textbox (for future text input)
		// TODO: implement text input
	}
}

// GetMousePosition returns the current mouse position
func GetMousePosition() (int, int) {
	return mouseX, mouseY
}

// IsMousePressed returns whether the left mouse button is pressed
func IsMousePressed() bool {
	return mousePressed
}
