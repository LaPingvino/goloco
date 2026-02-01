package ui

import (
	"log"

	"github.com/LaPingvino/goloco/pkg/gfx"
	"github.com/LaPingvino/goloco/pkg/graphics"
)

// WidgetType identifies the type of widget
type WidgetType uint8

const (
	WidgetTypeNone WidgetType = iota
	WidgetTypeFrame
	WidgetTypeCaption
	WidgetTypeButton
	WidgetTypePanel
	WidgetTypeTextBox
	WidgetTypeCheckbox
	WidgetTypeImageButton
	WidgetTypeTab
	WidgetTypeScrollView
	WidgetTypeSlider
	WidgetTypeDropdown
	WidgetTypeLabel
	WidgetTypeGroupBox
	WidgetTypeTableHeader
	WidgetTypeViewport
	WidgetTypeColourButton
	WidgetTypeStepper
	WidgetTypeWt3
)

// WidgetIndex is the index of a widget within a window
type WidgetIndex int16

// WidgetFlags control widget behavior and appearance
type WidgetFlags uint16

const (
	WidgetFlagNone        WidgetFlags = 0
	WidgetFlagEnabled     WidgetFlags = 1 << 0
	WidgetFlagDisabled    WidgetFlags = 1 << 1
	WidgetFlagPressed     WidgetFlags = 1 << 2
	WidgetFlagHighlight   WidgetFlags = 1 << 3
	WidgetFlagTransparent WidgetFlags = 1 << 4
	WidgetFlagNoFill      WidgetFlags = 1 << 5
	WidgetFlagResizable   WidgetFlags = 1 << 6
	WidgetFlagClosable    WidgetFlags = 1 << 7
)

// RectInsetFlags for FillRectInset operations (re-exported from graphics package)
const (
	RectInsetNone        = graphics.RectInsetNone
	RectInsetBorder      = graphics.RectInsetBorder
	RectInsetSecondary   = graphics.RectInsetSecondary
	RectInsetTransparent = graphics.RectInsetTransparent
)

// WindowColour represents a colour scheme for a window
type WindowColour uint8

const (
	WindowColourPrimary WindowColour = iota
	WindowColourSecondary
	WindowColourTertiary
	WindowColourQuaternary
)

// WindowType identifies the type of window
type WindowType uint8

const (
	WindowTypeMain WindowType = iota
	WindowTypeViewport
	WindowTypeDialog
	WindowTypeToolbar
	WindowTypeDropdown
	WindowTypeTooltip
	WindowTypeTitleMenu
)

// WindowNumber is a unique identifier for a window instance
type WindowNumber uint16

// WindowFlags control window behavior
type WindowFlags uint32

const (
	WindowFlagNone         WindowFlags = 0
	WindowFlagStickToBack  WindowFlags = 1 << 0
	WindowFlagStickToFront WindowFlags = 1 << 1
	WindowFlagResizable    WindowFlags = 1 << 2
	WindowFlagNoBackground WindowFlags = 1 << 3
	WindowFlagTransparent  WindowFlags = 1 << 4
	WindowFlagNoAutoClose  WindowFlags = 1 << 5
)

// Widget represents a UI element within a window
type Widget struct {
	Type    WidgetType
	Index   WidgetIndex
	X       int16
	Y       int16
	Width   int16
	Height  int16
	Colour  gfx.AdvancedColor
	Content uint32 // for image/string IDs
	Text    string
	Flags   WidgetFlags
	Tooltip string

	// State
	Activated bool
	Visible   bool

	// Callback functions
	Draw     func(*Widget, *graphics.DrawingContext) error
	drawFunc func(*Widget, *graphics.DrawingContext) error
}

// Bounds returns the widget's bounding rectangle
func (w *Widget) Bounds() gfx.Rect {
	return gfx.Rect{
		X:      w.X,
		Y:      w.Y,
		Width:  w.Width,
		Height: w.Height,
	}
}

// Position returns the widget's position as a Point
func (w *Widget) Position() gfx.Point {
	return gfx.Point{X: w.X, Y: w.Y}
}

// Size returns the widget's dimensions
func (w *Widget) Size() gfx.Size {
	return gfx.Size{Width: w.Width, Height: w.Height}
}

// IsEnabled returns true if the widget is enabled
func (w *Widget) IsEnabled() bool {
	return w.Flags&WidgetFlagEnabled != 0 && w.Flags&WidgetFlagDisabled == 0
}

// IsActivated returns true if the widget is activated/pressed
func (w *Widget) IsActivated() bool {
	return w.Activated
}

// Helper methods for AI-generated code compatibility
func (w *Widget) position() gfx.Point {
	return w.Position()
}

func (w *Widget) size() gfx.Size {
	return w.Size()
}

func (w *Widget) colour() gfx.AdvancedColor {
	return w.Colour
}

func (w *Widget) flags() WidgetFlags {
	return w.Flags
}

func (w *Widget) activated() bool {
	return w.Activated
}

func (w *Widget) drawLabel(window *Window, dc *DrawingContext) {
	// Simple label drawing - just text
	if w.Text != "" {
		_ = dc.DrawString(w.X, w.Y, w.Text, uint8(w.Colour.Color))
	}
}

// Window represents a UI window
type Window struct {
	Type      WindowType
	Number    WindowNumber
	X         int16
	Y         int16
	Width     int16
	Height    int16
	MinWidth  int16
	MinHeight int16
	MaxWidth  int16
	MaxHeight int16

	Flags   WindowFlags
	Colours [4]gfx.AdvancedColor

	// Widgets
	Widgets          []Widget
	DisabledWidgets  uint32 // bitfield
	ActivatedWidgets uint32 // bitfield

	// State
	Invalidated bool
	Visible     bool
	Destroyed   bool

	// Optional fields referenced by AI-generated code
	background    *uint8
	Viewports     []*Viewport // capitalized for AI-generated code compatibility
	eventHandlers *WindowEventHandlers
	viewports     []*Viewport
	children      []*Window
	owner         *Window

	// Callback functions
	OnClose  func(*Window)
	OnResize func(*Window, int16, int16)
	OnUpdate func(*Window)
}

// WindowEventHandlers contains callback functions for window events
type WindowEventHandlers struct {
	PrepareDraw func(*Window)
	Draw        func(*Window, *graphics.DrawingContext)
	PostDraw    func(*Window, *graphics.DrawingContext)
}

// Viewport represents a viewport within a window
type Viewport struct {
	X           int16
	Y           int16
	Width       int16
	Height      int16
	Invalidated bool
}

// Bounds returns the window's bounding rectangle
func (w *Window) Bounds() gfx.Rect {
	return gfx.Rect{
		X:      w.X,
		Y:      w.Y,
		Width:  w.Width,
		Height: w.Height,
	}
}

// Position returns the window's position as a Point
func (w *Window) Position() gfx.Point {
	return gfx.Point{X: w.X, Y: w.Y}
}

// position returns the window's position (lowercase for AI-generated code)
func (w *Window) position() gfx.Point {
	return w.Position()
}

// SetSize changes the window's dimensions
func (w *Window) SetSize(size gfx.Size) {
	w.Width = size.Width
	w.Height = size.Height
	if w.OnResize != nil {
		w.OnResize(w, size.Width, size.Height)
	}
}

// FindWidgetAt finds the widget at the given position (relative to window)
func (w *Window) FindWidgetAt(x, y int16) WidgetIndex {
	for i := range w.Widgets {
		widget := &w.Widgets[i]
		if !widget.Visible {
			continue
		}
		if x >= widget.X && x < widget.X+widget.Width &&
			y >= widget.Y && y < widget.Y+widget.Height {
			return WidgetIndex(i)
		}
	}
	return -1
}

// DrawingContext type alias for compatibility with AI-generated code
type DrawingContext = graphics.DrawingContext

// Draw renders the window to the drawing context. Stub — not yet implemented.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Window.cpp
//   Window::draw(Gfx::DrawingContext& drawingCtx)
//   Window::callDraw(Gfx::DrawingContext& ctx)
//
// In OpenLoco, draw() iterates over each widget and dispatches to per-widget
// draw handlers (button, caption, panel, etc.).  It also draws the window
// background via fillRectInset and delegates to custom per-window-type draw
// callbacks registered in the WindowEventTable.
func (w *Window) Draw(dc *DrawingContext) {
	log.Println("[UI] Window.Draw: stub — window", w.Number, "not yet rendered")
}
