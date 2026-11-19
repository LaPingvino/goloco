package ui

import "image"

// ViewportWidgets is a window with a viewport widget (WidgetType::viewport):
// the Town and Industry windows (ViewWindow), the Station and the Vehicle
// window, each on its first tab. MouseInput.cpp reads it twice: a press or
// hover on the widget (findWidgetAt: the last widget holding the point, so
// the centre-view button over the viewport is not it) begins viewportLeft /
// viewportRight and getItemLeft there; ViewportInteraction's
// getMapCoordinatesFromPos and rightOver read the viewport itself
// (containsWindowPos), wherever on the window the point is.
// OpenLoco reference: src/OpenLoco/src/Input/MouseInput.cpp,
// src/OpenLoco/src/Ui/Window.cpp (findWidgetAt),
// src/OpenLoco/src/Ui/Windows/TownWindow.cpp, IndustryWindow.cpp,
// StationWindow.cpp, Vehicle.cpp (initViewport: the viewport one pixel in
// from the widget's edges, widget.width() − 2 × widget.height() − 2)
type ViewportWidgets interface {
	// ViewportRect is the viewport's rectangle on the screen; ok false when
	// the window shows none (another tab, nothing to view).
	ViewportRect() (r image.Rectangle, ok bool)
	// ViewportWidgetAt reports that findWidgetAt at screen (x, y) is the
	// viewport widget.
	ViewportWidgetAt(x, y int) bool
}

// viewportInside is the viewport of a widget l, t, r, b (inclusive,
// window-relative) on the screen: one pixel in on each side.
func viewportInside(b *WinBase, l, t, r, bt int) image.Rectangle {
	return image.Rect(b.X+l+1, b.Y+t+1, b.X+r, b.Y+bt)
}

// inWidget is findWidgetAt's test for one widget: l ≤ x − window x ≤ r and
// the same for y.
func inWidget(b *WinBase, x, y, l, t, r, bt int) bool {
	x, y = x-b.X, y-b.Y
	return x >= l && x <= r && y >= t && y <= bt
}

// ViewportRect is the Town / Industry window's viewport on its first tab.
func (w *ViewWindow) ViewportRect() (image.Rectangle, bool) {
	if w.Tab != 0 || w.View == nil {
		return image.Rectangle{}, false
	}
	l, t, r, b := w.viewport()
	return viewportInside(&w.WinBase, l, t, r, b), true
}

// ViewportWidgetAt: the viewport widget, less the centre-view button
// (centre_on_viewport: right and bottom one in from the viewport's, 24 × 24)
// that comes after it.
func (w *ViewWindow) ViewportWidgetAt(x, y int) bool {
	if w.Tab != 0 || w.View == nil {
		return false
	}
	l, t, r, b := w.viewport()
	return inWidget(&w.WinBase, x, y, l, t, r, b) && !inWidget(&w.WinBase, x, y, r-24, b-24, r-1, b-1)
}

// ViewportRect is the Station window's viewport on its Station tab.
func (w *StationWindow) ViewportRect() (image.Rectangle, bool) {
	if w.Tab != 0 || w.View == nil {
		return image.Rectangle{}, false
	}
	l, t, r, b := w.viewport()
	return viewportInside(&w.WinBase, l, t, r, b), true
}

// ViewportWidgetAt: as the Town window's, the centre-view button excepted.
func (w *StationWindow) ViewportWidgetAt(x, y int) bool {
	if w.Tab != 0 || w.View == nil {
		return false
	}
	l, t, r, b := w.viewport()
	return inWidget(&w.WinBase, x, y, l, t, r, b) && !inWidget(&w.WinBase, x, y, r-24, b-24, r-1, b-1)
}

// ViewportRect is the Vehicle window's viewport on its Main tab.
func (w *VehicleWindow) ViewportRect() (image.Rectangle, bool) {
	if w.Tab != 0 || w.View == nil {
		return image.Rectangle{}, false
	}
	v := VehicleMainLayout(w.W, w.H, w.Main)[vmViewport]
	return viewportInside(&w.WinBase, v.L, v.T, v.R, v.B), true
}

// ViewportWidgetAt is findWidgetAt over the Main tab's widgets: the last
// one shown that holds the point must be the viewport.
func (w *VehicleWindow) ViewportWidgetAt(x, y int) bool {
	if w.Tab != 0 || w.View == nil {
		return false
	}
	found := -1
	for i, b := range VehicleMainLayout(w.W, w.H, w.Main) {
		if !b.Hidden && inWidget(&w.WinBase, x, y, b.L, b.T, b.R, b.B) {
			found = i
		}
	}
	return found == vmViewport
}
