package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// diag.go — diagnostics for headless checks: the --diag / F12 crop and the
// GOLOCO_DIAG_* panels.

const diagCropW = 480

const diagCropH = 360

// diagBackground draws a black+red cross-hatch pattern onto img so that
// transparent/unrendered areas are visually distinct from intentional black pixels.
func diagBackground(img *ebiten.Image) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	pixels := make([]byte, w*h*4)
	const stride = 16 // grid cell size
	red := [4]byte{180, 0, 0, 255}
	dark := [4]byte{20, 20, 20, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			onLine := (x%stride == 0) || (y%stride == 0)
			var c [4]byte
			if onLine {
				c = red
			} else {
				c = dark
			}
			i := (y*w + x) * 4
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = c[0], c[1], c[2], c[3]
		}
	}
	img.WritePixels(pixels)
}

// saveDiagCrop saves a 480×360 PNG cropped from the centre of the screen
// (or centred on tile diagTileX/diagTileY when the world is loaded).
// The file is always named "diag.png" so Claude can read it without guessing the name.
func saveDiagCrop(screen *ebiten.Image, g *Game, tileX, tileY int) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Centre pixel: use tile position when valid, otherwise screen centre.
	cx, cy := sw/2, sh/2
	if g.w != nil && tileX >= 0 && tileY >= 0 {
		cx, cy = g.w.TileToScreenPx(tileX, tileY)
	}

	x0 := cx - diagCropW/2
	y0 := cy - diagCropH/2
	x0 = max(0, min(x0, sw-diagCropW))
	y0 = max(0, min(y0, sh-diagCropH))

	crop := screen.SubImage(image.Rect(x0, y0, x0+diagCropW, y0+diagCropH))

	f, err := os.Create(debugPath("diag.png"))
	if err != nil {
		log.Printf("[Diag] Cannot create diag.png: %v", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, crop); err != nil {
		log.Printf("[Diag] Encode failed: %v", err)
		return
	}
	log.Printf("[Diag] Saved diag.png (%dx%d crop centred on tile %d,%d, screen offset %d,%d)",
		diagCropW, diagCropH, tileX, tileY, x0, y0)
}

// drawInlineDiag (GOLOCO_DIAG_INLINE) is drawVehicleInline for the first ten
// rail and road vehicle types, in a panel at (20, 60).
func (g *Game) drawInlineDiag(screen *ebiten.Image) {
	if os.Getenv("GOLOCO_DIAG_TTF") != "" { // the G1 font and the TTF fallback on one line
		fillRect(screen, 10, 10, 300, 30, color.RGBA{200, 200, 200, 255})
		fillRect(screen, 10, 20, 300, 1, color.RGBA{255, 0, 0, 255}) // y = 20: the text tops
		ui.DrawTextBold(screen, "Bitmap text", 20, 20, color.Black)
		ui.DrawTextBold(screen, "TTF 日本語 text", 120, 20, color.Black)
	}
	if os.Getenv("GOLOCO_DIAG_INLINE") == "" || g.objMgr == nil {
		return
	}
	var vs []*objects.VehicleObject
	var offs []int
	for _, v := range g.objMgr.Vehicles {
		if v != nil && v.ImageOffset != 0 && (v.Mode == objects.TransportModeRail || v.Mode == objects.TransportModeRoad) && len(vs) < 10 {
			vs = append(vs, v)
			offs = append(offs, g.objMgr.VehicleYDisplayOffset(v))
		}
	}
	log.Printf("[diag] inline: %d vehicle types", len(vs))
	ui.DrawVehiclesInlineDiag(screen, 20, 60, vs, offs)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
