package ui

import (
	"image/color"

	"github.com/LaPingvino/goloco/pkg/render"
	"github.com/LaPingvino/goloco/pkg/sprites"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

	// Window colors (Locomotion style)
	frameBg := color.RGBA{209, 205, 181, 255}
	titleBg := color.RGBA{180, 170, 150, 255}
	borderLight := color.RGBA{255, 255, 255, 255}
	borderDark := color.RGBA{100, 95, 85, 255}
	contentBg := color.RGBA{200, 195, 175, 255}

	// Draw window frame background
	for y := w.Y; y < w.Y+w.Height; y++ {
		for x := w.X; x < w.X+w.Width; x++ {
			screen.Set(x, y, frameBg)
		}
	}

	// Draw title bar
	for y := w.Y; y < w.Y+simpleTitleBarHeight; y++ {
		for x := w.X; x < w.X+w.Width; x++ {
			screen.Set(x, y, titleBg)
		}
	}

	// Draw content area
	contentX := w.X + simpleBorderWidth
	contentY := w.Y + simpleTitleBarHeight
	contentW := w.Width - simpleBorderWidth*2
	contentH := w.Height - simpleTitleBarHeight - simpleBorderWidth
	for y := contentY; y < contentY+contentH; y++ {
		for x := contentX; x < contentX+contentW; x++ {
			screen.Set(x, y, contentBg)
		}
	}

	// Draw 3D border effect
	// Top and left (light)
	for x := w.X; x < w.X+w.Width; x++ {
		screen.Set(x, w.Y, borderLight)
		screen.Set(x, w.Y+1, borderLight)
	}
	for y := w.Y; y < w.Y+w.Height; y++ {
		screen.Set(w.X, y, borderLight)
		screen.Set(w.X+1, y, borderLight)
	}
	// Bottom and right (dark)
	for x := w.X; x < w.X+w.Width; x++ {
		screen.Set(x, w.Y+w.Height-1, borderDark)
		screen.Set(x, w.Y+w.Height-2, borderDark)
	}
	for y := w.Y; y < w.Y+w.Height; y++ {
		screen.Set(w.X+w.Width-1, y, borderDark)
		screen.Set(w.X+w.Width-2, y, borderDark)
	}

	// Draw title text
	ebitenutil.DebugPrintAt(screen, w.Title, w.X+6, w.Y+5)

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
		// Fallback close button
		for y := closeY; y < closeY+12; y++ {
			for x := closeX; x < closeX+12; x++ {
				screen.Set(x, y, color.RGBA{200, 80, 80, 255})
			}
		}
	}

	// Draw content if callback is set
	if w.DrawContent != nil {
		w.DrawContent(screen, contentX, contentY, contentW, contentH, renderer)
	}
}
