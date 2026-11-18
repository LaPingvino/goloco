package ui

import "github.com/hajimehoshi/ebiten/v2"

// BarWindow is one of the game scene's bars as a window of the stack:
// ToolbarTop (0, 0, width × 28), CompanyInfoPanel (0, height − 27, 140 × 27)
// and TimePanel (width − 140, height − 27, 140 × 27) — stickToFront,
// transparent and noBackground windows that Gui::init opens in that order
// (OpenLoco reference: src/OpenLoco/src/Gui.cpp init, Ui/Windows/
// ToolbarTop.cpp, CompanyInfoPanel.cpp, TimePanel.cpp open). Their drawing
// and their widgets are the game's (cmd/goloco): DrawBar, ClickBar and
// WidgetAt.
type BarWindow struct {
	WinBase
	DrawBar  func(screen *ebiten.Image)
	ClickBar func(x, y int)
	// WidgetAt reports a widget at a screen point (findWidgetAt): a
	// noBackground window is under the mouse only there.
	WidgetAt func(x, y int) bool
}

// NewBarWindow is a bar of the given type: stickToFront and noBackground,
// without a caption.
func NewBarWindow(typ WindowType, x, y, w, h int) *BarWindow {
	return &BarWindow{WinBase: WinBase{WinType: typ, X: x, Y: y, W: w, H: h,
		WinFlags: WindowFlagStickToFront | WindowFlagNoBackground, NoCaption: true}}
}

func (w *BarWindow) Draw(screen *ebiten.Image) {
	if w.DrawBar != nil {
		w.DrawBar(screen)
	}
}

func (w *BarWindow) HandleClick(x, y int) {
	if w.ClickBar != nil {
		w.ClickBar(x, y)
	}
}

func (w *BarWindow) HasWidgetAt(x, y int) bool { return w.WidgetAt != nil && w.WidgetAt(x, y) }

// IsBar reports one of the bars' window types.
func IsBar(w Win) bool {
	switch w.Base().WinType {
	case WindowTypeTopToolbar, WindowTypeCompanyInfoPanel, WindowTypeTimePanel:
		return true
	}
	return false
}
