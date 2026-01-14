package gfx

// Basic geometry types used throughout the UI and rendering systems

// Point represents a 2D position
type Point struct {
	X int16
	Y int16
}

// Add returns a new point with the coordinates added
func (p Point) Add(other Point) Point {
	return Point{X: p.X + other.X, Y: p.Y + other.Y}
}

// Sub returns a new point with the coordinates subtracted
func (p Point) Sub(other Point) Point {
	return Point{X: p.X - other.X, Y: p.Y - other.Y}
}

// Size represents dimensions
type Size struct {
	Width  int16
	Height int16
}

// Rect represents a rectangle
type Rect struct {
	X      int16
	Y      int16
	Width  int16
	Height int16
}

// Right returns the right edge X coordinate
func (r Rect) Right() int16 {
	return r.X + r.Width
}

// Bottom returns the bottom edge Y coordinate
func (r Rect) Bottom() int16 {
	return r.Y + r.Height
}

// Contains checks if a point is inside the rectangle
func (r Rect) Contains(x, y int16) bool {
	return x >= r.X && x < r.Right() && y >= r.Y && y < r.Bottom()
}

// Intersects checks if two rectangles overlap
func (r Rect) Intersects(other Rect) bool {
	return r.X < other.Right() && r.Right() > other.X &&
		r.Y < other.Bottom() && r.Bottom() > other.Y
}

// Color represents a palette color index (0-255)
type Color uint8

// Standard OpenLoco colors
const (
	ColorBlack Color = iota
	ColorGrey
	ColorWhite
	ColorDarkPurple
	ColorLightPurple
	ColorBrightPurple
	ColorDarkBlue
	ColorLightBlue
	ColorIcyBlue
	ColorTeal
	ColorAquamarine
	ColorSaturatedGreen
	ColorDarkGreen
	ColorMossGreen
	ColorBrightGreen
	ColorOliveGreen
	ColorDarkOliveGreen
	ColorBrightYellow
	ColorYellow
	ColorMossYellow
	ColorDarkYellow
	ColorLightOrange
	ColorDarkOrange
	ColorLightBrown
	ColorSaturatedBrown
	ColorDarkBrown
	ColorSalmon
	ColorButterflyRed
	ColorDarkRed
	ColorBrightRed
	ColorDarkPink
	ColorBrightPink
)

// AdvancedColor represents a color with special effects
type AdvancedColor struct {
	Color       Color
	Translucent bool
	Outline     bool
	Inset       bool
}

// NewColor creates a basic AdvancedColor from a palette color
func NewColor(c Color) AdvancedColor {
	return AdvancedColor{Color: c}
}

// WithTranslucent returns a copy with translucency enabled
func (ac AdvancedColor) WithTranslucent() AdvancedColor {
	ac.Translucent = true
	return ac
}

// WithOutline returns a copy with outline enabled
func (ac AdvancedColor) WithOutline() AdvancedColor {
	ac.Outline = true
	return ac
}

// WithInset returns a copy with inset enabled
func (ac AdvancedColor) WithInset() AdvancedColor {
	ac.Inset = true
	return ac
}
