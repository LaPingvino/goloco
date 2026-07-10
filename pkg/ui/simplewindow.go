package ui

import (
	"image/color"

	"github.com/LaPingvino/goloco/pkg/gfx"
	"github.com/LaPingvino/goloco/pkg/graphics"
	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/sprites"
	"github.com/hajimehoshi/ebiten/v2"
)

// SimpleWindow represents a basic draggable game window for simple UI
// This is separate from the more complex Window type in types.go
type SimpleWindow struct {
	Title   string
	X       int
	Y       int
	Width   int
	Height  int
	Visible bool

	// Internal state
	dragging bool
	dragOffX int
	dragOffY int

	// Content callback
	DrawContent func(screen *ebiten.Image, x, y, w, h int, renderer *render.Renderer)

	// OnContentClick is called when the user clicks inside the content area.
	// relX, relY are coordinates relative to the top-left of the content area.
	OnContentClick func(relX, relY int)

	// OnScrollWheel is called when the mouse wheel is used over this window.
	// dy is the vertical scroll delta (positive = scroll up/away from user).
	OnScrollWheel func(dy float64)
}

// SimpleWindowManager manages all open simple windows
type SimpleWindowManager struct {
	Windows []*SimpleWindow
}

// NewSimpleWindowManager creates a new window manager
func NewSimpleWindowManager() *SimpleWindowManager {
	return &SimpleWindowManager{
		Windows: make([]*SimpleWindow, 0),
	}
}

// NewSimpleWindow creates a new window
func NewSimpleWindow(title string, x, y, width, height int) *SimpleWindow {
	return &SimpleWindow{
		Title:   title,
		X:       x,
		Y:       y,
		Width:   width,
		Height:  height,
		Visible: true,
	}
}

// OpenWindow adds a window to the manager
func (wm *SimpleWindowManager) OpenWindow(w *SimpleWindow) {
	// Check if window with same title already exists
	for _, existing := range wm.Windows {
		if existing.Title == w.Title {
			existing.Visible = true
			return
		}
	}
	wm.Windows = append(wm.Windows, w)
}

// CloseWindow removes a window from the manager
func (wm *SimpleWindowManager) CloseWindow(title string) {
	for i, w := range wm.Windows {
		if w.Title == title {
			wm.Windows = append(wm.Windows[:i], wm.Windows[i+1:]...)
			return
		}
	}
}

// Draw draws all visible windows
func (wm *SimpleWindowManager) Draw(screen *ebiten.Image, renderer *render.Renderer) {
	for _, w := range wm.Windows {
		if w.Visible {
			w.Draw(screen, renderer)
		}
	}
}

// HandleClick handles mouse clicks, returns true if a window consumed the click
func (wm *SimpleWindowManager) HandleClick(x, y int, pressed bool) bool {
	// Check windows in reverse order (topmost first)
	for i := len(wm.Windows) - 1; i >= 0; i-- {
		w := wm.Windows[i]
		if !w.Visible {
			continue
		}

		// Check if click is in window bounds
		if x >= w.X && x < w.X+w.Width && y >= w.Y && y < w.Y+w.Height {
			if pressed {
				// Check close button (top right)
				closeX := w.X + w.Width - 18
				closeY := w.Y + 4
				if x >= closeX && x < closeX+14 && y >= closeY && y < closeY+14 {
					w.Visible = false
					return true
				}

				// Start dragging from title bar
				if y < w.Y+22 {
					w.dragging = true
					w.dragOffX = x - w.X
					w.dragOffY = y - w.Y
				} else if w.OnContentClick != nil {
					// Click in content area — pass relative coordinates
					contentX := w.X + simpleBorderWidth
					contentY := w.Y + simpleTitleBarHeight
					w.OnContentClick(x-contentX, y-contentY)
				}
			} else {
				w.dragging = false
			}
			return true
		}
	}
	return false
}

// HandleDrag handles mouse drag
func (wm *SimpleWindowManager) HandleDrag(x, y int) {
	for _, w := range wm.Windows {
		if w.dragging {
			w.X = x - w.dragOffX
			w.Y = y - w.dragOffY
			// Keep on screen
			if w.X < 0 {
				w.X = 0
			}
			if w.Y < 0 {
				w.Y = 0
			}
		}
	}
}

// StopDrag stops all dragging
func (wm *SimpleWindowManager) StopDrag() {
	for _, w := range wm.Windows {
		w.dragging = false
	}
}

const (
	simpleTitleBarHeight = 22
	simpleBorderWidth    = 3
)

// Draw draws the window
func (w *SimpleWindow) Draw(screen *ebiten.Image, renderer *render.Renderer) {
	if !w.Visible {
		return
	}

	// Window colors (Locomotion style) - prefer using loaded G1/global palette when available
	frameBg := color.RGBA{209, 205, 181, 255}
	titleBg := color.RGBA{180, 170, 150, 255}
	contentBg := color.RGBA{200, 195, 175, 255}

	// If a renderer/global palette is available, map these target RGBA values to
	// the closest palette entries and use palette colors so the UI matches the
	// game's look-and-feel more closely. (Bevel light/dark shades come from
	// FillRectInset's palette-index offsets, not from separate colours here.)
	if graphics.IsGlobalPaletteLoaded() {
		frameBg = graphics.GetGlobalColor(graphics.MatchPaletteIndex(frameBg))
		titleBg = graphics.GetGlobalColor(graphics.MatchPaletteIndex(titleBg))
		contentBg = graphics.GetGlobalColor(graphics.MatchPaletteIndex(contentBg))
	}

	// Draw window frame, title bar and content area using the DrawingContext so
	// palette-aware colors are used when a G1 palette is loaded.
	dc := graphics.NewDrawingContext(screen)
	frameIdx := graphics.MatchPaletteIndex(frameBg)
	titleIdx := graphics.MatchPaletteIndex(titleBg)
	contentIdx := graphics.MatchPaletteIndex(contentBg)

	// Draw full frame background with palette index
	_ = dc.FillRect(int16(w.X), int16(w.Y), int16(w.Width), int16(w.Height), frameIdx)

	// Draw title bar
	_ = dc.FillRect(int16(w.X), int16(w.Y), int16(w.Width), int16(simpleTitleBarHeight), titleIdx)

	// Draw content area
	contentX := w.X + simpleBorderWidth
	contentY := w.Y + simpleTitleBarHeight
	contentW := w.Width - simpleBorderWidth*2
	contentH := w.Height - simpleBorderWidth - simpleTitleBarHeight
	_ = dc.FillRect(int16(contentX), int16(contentY), int16(contentW), int16(contentH), contentIdx)

	// Draw a beveled border for the window using FillRectInset which will use
	// nearby palette indices to simulate highlights and shadows. Pass the frame
	// palette index as the base advanced color.
	_ = dc.FillRectInset(int16(w.X), int16(w.Y), int16(w.Width), int16(w.Height), gfx.AdvancedColor{Color: gfx.Color(frameIdx)}, RectInsetBorder|RectInsetSecondary)

	// Draw title text using the drawing context so we can use the active palette.
	// Prefer a readable text color mapped to the global palette when available.
	// Desired default for title text is black; map it to the closest palette index.
	textTarget := color.RGBA{0, 0, 0, 255}
	textIdx := graphics.MatchPaletteIndex(textTarget)
	_ = dc.DrawString(int16(w.X+6), int16(w.Y+5), w.Title, textIdx)

	// Draw close button using G1 sprite if available
	closeX := w.X + w.Width - 18
	closeY := w.Y + 4
	if renderer != nil && renderer.G1 != nil {
		if img := renderer.GetSprite(sprites.CloseButton); img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(closeX), float64(closeY))
			screen.DrawImage(img, op)
		}
	} else {
		// Fallback: draw a beveled close button with an × label.
		const cw, ch = 14, 14
		bgC := color.RGBA{192, 56, 56, 255}
		bgIdx := graphics.MatchPaletteIndex(bgC)
		var lightIdx, darkIdx uint8
		if bgIdx <= 252 {
			lightIdx = bgIdx + 4
		} else {
			lightIdx = bgIdx
		}
		if bgIdx >= 4 {
			darkIdx = bgIdx - 4
		} else {
			darkIdx = bgIdx
		}
		_ = dc.FillRect(int16(closeX), int16(closeY), cw, ch, bgIdx)
		_ = dc.FillRect(int16(closeX), int16(closeY), cw, 1, lightIdx)
		_ = dc.FillRect(int16(closeX), int16(closeY), 1, ch, lightIdx)
		_ = dc.FillRect(int16(closeX), int16(closeY+ch-1), cw, 1, darkIdx)
		_ = dc.FillRect(int16(closeX+cw-1), int16(closeY), 1, ch, darkIdx)
		// × label (white)
		textIdx := graphics.MatchPaletteIndex(color.RGBA{255, 255, 255, 255})
		_ = dc.DrawString(int16(closeX+4), int16(closeY-2), "x", textIdx)
	}

	// Draw content if callback is set
	if w.DrawContent != nil {
		w.DrawContent(screen, contentX, contentY, contentW, contentH, renderer)
	}
}
