package gfx8

// Zoomed drawing — a port of Graphics/SoftwareDrawingContext.cpp
// getDrawImagePosArgs and DrawSpriteBMP.hpp / DrawSpriteRLE.hpp for zoom 1-3.
// Coordinates are viewport pixels at zoom 0 relative to the target's origin
// (upstream's pos - rt.x/y with rt.x = rt.y = 0); the target holds the zoomed
// pixels, so its logical size is Width<<zoom × Height<<zoom.

// posArgs is getDrawImagePosArgs<zoom, rle>: source start, destination start
// (zoomed pixels) and size (source pixels) after alignment and clipping.
func posArgs(rt *RenderTarget, s *Sprite, x, y, zoom int, rle bool) (srcX, srcY, dstX, dstY, w, h int, ok bool) {
	if zoom > 0 && s.NoZoomDraw {
		return
	}
	zoomMask := ^0 << zoom
	low := ^zoomMask
	rtW, rtH := rt.Width<<zoom, rt.Height<<zoom
	if zoom > 0 && rle {
		x -= low
		y -= low
	}
	h = s.Height
	dstTop := y + s.YOffset
	if !rle {
		dstTop &= zoomMask // "the RLE version does not use the zoom mask on y"
	}
	if dstTop < 0 {
		h += dstTop
		if h <= 0 {
			return
		}
		srcY -= dstTop
		dstTop = 0
	} else if zoom > 0 && rle {
		srcY -= dstTop & low
		h += dstTop & low
	}
	if b := dstTop + h; b > rtH {
		h -= b - rtH
	}
	if h <= 0 {
		return
	}
	dstY = dstTop >> zoom
	w = s.Width
	dstLeft := (x + s.XOffset + low) & zoomMask
	if dstLeft < 0 {
		w += dstLeft
		if w <= 0 {
			return
		}
		srcX -= dstLeft
		dstLeft = 0
	} else if zoom > 0 && rle {
		srcX -= dstLeft & low
	}
	if r := dstLeft + w; r > rtW {
		w -= r - rtW
		if w <= 0 {
			return
		}
	}
	dstX = dstLeft >> zoom
	return srcX, srcY, dstX, dstY, w, h, true
}

// DrawSpriteZoom draws s at zoom 0-3 (drawImagePaletteSet without the
// hasZoomSprites recursion, which needs the sprite table: see the caller).
func DrawSpriteZoom(rt *RenderTarget, s *Sprite, x, y, zoom int, op BlendOp, palMap []uint8, noise *Sprite) {
	if zoom == 0 {
		DrawSpriteNoise(rt, s, x, y, op, palMap, noise)
		return
	}
	if noise != nil && !s.RLE {
		op |= OpNoiseMask
	} else {
		op &^= OpNoiseMask
	}
	srcX, srcY, dstX, dstY, w, h, ok := posArgs(rt, s, x, y, zoom, s.RLE)
	if !ok {
		return
	}
	step := 1 << zoom
	if !s.RLE {
		// drawBMPSprite
		for row := 0; h > 0; h, row = h-step, row+1 {
			sy := srcY + row*step
			drow := (dstY + row) * rt.Width
			for col, rem := 0, w; rem > 0; col, rem = col+1, rem-step {
				i := s.Width*sy + srcX + col*step
				nm := uint8(0xFF)
				if op&OpNoiseMask != 0 {
					if i < len(noise.Pix) {
						nm = noise.Pix[i]
					} else {
						nm = 0
					}
				}
				blitPixel(op, s.Pix[i], &rt.Bits[drow+dstX+col], palMap, nm)
			}
		}
		return
	}
	// drawRLESprite
	dst0 := dstY
	if srcY < 0 {
		srcY += step
		h -= step
		dst0++
	}
	for i := 0; i < h; i += step {
		sy := srcY + i
		line := (dst0 + i>>zoom) * rt.Width
		for _, run := range s.Runs[sy] {
			srcPos, num := run[0], run[1]
			xx := run[0] - srcX
			if xx > 0 {
				if mod := xx & (step - 1); mod != 0 {
					off := step - mod
					xx += off
					srcPos += off
					num -= off
				}
			} else if xx < 0 {
				srcPos += -xx
				num += xx
				xx = 0
			}
			num = min(num, w-xx)
			d := line + dstX + xx>>zoom
			for ; num > 0; num, srcPos, d = num-step, srcPos+step, d+1 {
				src := uint8(0)
				if srcPos < s.Width {
					src = s.Pix[sy*s.Width+srcPos]
				}
				blitPixel(op, src, &rt.Bits[d], palMap, 0xFF)
			}
		}
	}
}
