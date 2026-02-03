package ui

import (
	"bytes"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// FontManager manages TrueType fonts for the UI
//
// Uses OpenTTD fonts for proper internationalization support
// Reference: https://github.com/OpenTTD/OpenTTD-TTF
type FontManager struct {
	Regular     *text.GoTextFace
	Bold        *text.GoTextFace
	RegularFace font.Face
	BoldFace    font.Face
}

var globalFontManager *FontManager

// InitFonts loads the OpenTTD fonts from disk
//
// Font files should be in the fonts/ directory:
//   - OpenTTD-Sans.ttf (regular)
//   - OpenTTD-Small.ttf (small/bold variant)
func InitFonts(fontDir string) (*FontManager, error) {
	fm := &FontManager{}

	// Load regular font (using Small variant for compact UI text)
	regularPath := fontDir + "/OpenTTD-Small.ttf"
	regularData, err := os.ReadFile(regularPath)
	if err != nil {
		return nil, err
	}

	regularTTF, err := opentype.Parse(regularData)
	if err != nil {
		return nil, err
	}

	// Create faces at default size (10pt for UI elements)
	const dpi = 72
	const size = 10

	regularFace, err := opentype.NewFace(regularTTF, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	fm.RegularFace = regularFace

	regularSource, err := text.NewGoTextFaceSource(bytes.NewReader(regularData))
	if err != nil {
		return nil, err
	}
	fm.Regular = &text.GoTextFace{
		Source: regularSource,
		Size:   size,
	}

	// Load mono font (used for bold variant)
	boldPath := fontDir + "/OpenTTD-Mono.ttf"
	boldData, err := os.ReadFile(boldPath)
	if err != nil {
		log.Printf("[Font] Bold font not found, using regular: %v", err)
		fm.Bold = fm.Regular
		fm.BoldFace = fm.RegularFace
		globalFontManager = fm
		return fm, nil
	}

	boldTTF, err := opentype.Parse(boldData)
	if err != nil {
		log.Printf("[Font] Bold font parse error, using regular: %v", err)
		fm.Bold = fm.Regular
		fm.BoldFace = fm.RegularFace
		globalFontManager = fm
		return fm, nil
	}

	boldFace, err := opentype.NewFace(boldTTF, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Printf("[Font] Bold face creation error, using regular: %v", err)
		fm.Bold = fm.Regular
		fm.BoldFace = fm.RegularFace
		globalFontManager = fm
		return fm, nil
	}

	fm.BoldFace = boldFace

	boldSource, err := text.NewGoTextFaceSource(bytes.NewReader(boldData))
	if err != nil {
		log.Printf("[Font] Bold source creation error, using regular: %v", err)
		fm.Bold = fm.Regular
		fm.BoldFace = fm.RegularFace
		globalFontManager = fm
		return fm, nil
	}
	fm.Bold = &text.GoTextFace{
		Source: boldSource,
		Size:   size,
	}

	globalFontManager = fm
	log.Println("[Font] Loaded OpenTTD fonts successfully")
	return fm, nil
}

// GetFontManager returns the global font manager
func GetFontManager() *FontManager {
	return globalFontManager
}

// DrawText draws text at the specified position using the regular font
// The Y coordinate is the TOP of the text (not baseline), adjusted for proper alignment
func DrawText(screen *ebiten.Image, str string, x, y int, clr color.Color) {
	if globalFontManager == nil || globalFontManager.Regular == nil {
		return
	}

	op := &text.DrawOptions{}
	// text/v2 uses baseline positioning, but Y should be top of text
	// Subtract to move text up to align properly
	op.GeoM.Translate(float64(x), float64(y-9))
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, globalFontManager.Regular, op)
}

// DrawTextBold draws text at the specified position using the bold font
// The Y coordinate is the TOP of the text (not baseline), adjusted for proper alignment
func DrawTextBold(screen *ebiten.Image, str string, x, y int, clr color.Color) {
	if globalFontManager == nil || globalFontManager.Bold == nil {
		return
	}

	op := &text.DrawOptions{}
	// text/v2 uses baseline positioning, but Y should be top of text
	// Subtract to move text up to align properly
	op.GeoM.Translate(float64(x), float64(y-9))
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, globalFontManager.Bold, op)
}

// MeasureText returns the width and height of text in pixels
func MeasureText(str string) (int, int) {
	if globalFontManager == nil || globalFontManager.RegularFace == nil {
		// Fallback estimate
		return len(str) * 6, 12
	}

	bounds, advance := font.BoundString(globalFontManager.RegularFace, str)
	width := advance.Ceil()
	height := (bounds.Max.Y - bounds.Min.Y).Ceil()

	return width, height
}
