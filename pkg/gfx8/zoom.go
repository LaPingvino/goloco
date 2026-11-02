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
	if rle {
		dstTop -= rt.OriginY
	} else {
		dstTop = (dstTop & zoomMask) - rt.OriginY // "the RLE version does not use the zoom mask on y"
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
	dstLeft := ((x + s.XOffset + low) & zoomMask) - rt.OriginX
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
		DrawSpriteNoise(rt, s, x-rt.OriginX, y-rt.OriginY, op, palMap, noise)
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

// DrawImageZoom is drawImagePaletteSet (0x00448D90) for image id: at zoom > 0
// an image with hasZoomSprites is drawn instead as image − zoomOffset into the
// same pixels seen one zoom level closer (rt.x/y and the position halved,
// zoom − 1), recursively; otherwise DrawSpriteZoom. sprite looks an image up
// (nil = no element). The blend op is the caller's for the whole recursion.
func DrawImageZoom(rt *RenderTarget, sprite func(id int) *Sprite, id, x, y, zoom int, op BlendOp, palMap []uint8, noise *Sprite) {
	s := sprite(id)
	if s == nil {
		return
	}
	if zoom > 0 && s.HasZoomSprites {
		half := *rt // the same bits; width/height in zoomed pixels are unchanged
		half.OriginX, half.OriginY = rt.OriginX>>1, rt.OriginY>>1
		DrawImageZoom(&half, sprite, id-s.ZoomOffset, x>>1, y>>1, zoom-1, op, palMap, noise)
		return
	}
	DrawSpriteZoom(rt, s, x, y, zoom, op, palMap, noise)
}

// DrawSpriteAtZoom is DrawImageZoom starting from an already looked-up
// sprite (its Index names the image for the hasZoomSprites recursion).
func DrawSpriteAtZoom(rt *RenderTarget, s *Sprite, sprite func(id int) *Sprite, x, y, zoom int, op BlendOp, palMap []uint8, noise *Sprite) {
	if s == nil {
		return
	}
	if zoom > 0 && s.HasZoomSprites && s.Index > 0 {
		half := *rt
		half.OriginX, half.OriginY = rt.OriginX>>1, rt.OriginY>>1
		DrawImageZoom(&half, sprite, s.Index-s.ZoomOffset, x>>1, y>>1, zoom-1, op, palMap, noise)
		return
	}
	DrawSpriteZoom(rt, s, x, y, zoom, op, palMap, noise)
}

// DrawSpriteMaskedZoom is drawImageMasked (0x00450705, drawImageMaskedZoom<z>
// and drawMaskedZoom) at zoom 0-3 with the target's origin: BMP image and
// mask only (RLE is unsupported upstream). At zoom > 0 an image with zoom
// sprites whose mask also has them recurses one level closer (origin and
// position halved, both images − their zoomOffset); a mask with noZoomDraw
// draws nothing. The mask is read at the image's own offsets.
func DrawSpriteMaskedZoom(rt *RenderTarget, img, mask *Sprite, sprite func(id int) *Sprite, x, y, zoom int) {
	if img == nil || mask == nil || img.RLE || mask.RLE {
		return
	}
	if zoom > 0 && img.HasZoomSprites {
		if mask.NoZoomDraw {
			return
		}
		if mask.HasZoomSprites && img.Index > 0 && mask.Index > 0 && sprite != nil {
			half := *rt
			half.OriginX, half.OriginY = rt.OriginX>>1, rt.OriginY>>1
			DrawSpriteMaskedZoom(&half, sprite(img.Index-img.ZoomOffset), sprite(mask.Index-mask.ZoomOffset), sprite, x>>1, y>>1, zoom-1)
			return
		}
	}
	if zoom == 0 {
		DrawSpriteMasked(rt, img, mask, x-rt.OriginX, y-rt.OriginY)
		return
	}
	h, w := img.Height, img.Width
	data := 0 // imageDataPos − offset
	zm := int(uint16(0xFFFF << zoom))
	offsetX := 1<<zoom - 1
	rtW, rtH := rt.Width<<zoom, rt.Height<<zoom // rt.width / rt.height
	dst := 0
	dstTop := int(int16(int(int16((img.YOffset+y)&zm)) - rt.OriginY))
	if dstTop >= 0 {
		dst += (dstTop >> zoom) * rt.Width // pitch 0 + width >> zoom
	} else {
		if dstTop+h == 0 {
			return
		}
		h += dstTop
		if h < 0 {
			return
		}
		data = int(uint16(-dstTop)) * img.Width
		dstTop = 0
	}
	if b := h + dstTop; b > rtH {
		h -= b - rtH
		if h <= 0 {
			return
		}
	}
	rowSize := 0
	dstWrap := rt.Width
	dstLeft := int(int16(int(int16((img.XOffset+x+offsetX)&zm)) - rt.OriginX))
	if dstLeft < 0 {
		w += dstLeft
		if w <= 0 {
			return
		}
		rowSize -= dstLeft
		data -= dstLeft
		dstLeft = 0
	}
	if w+dstLeft > rtW {
		r := w + dstLeft - rtW
		w -= r
		if w <= 0 {
			return
		}
		rowSize += r
	}
	dst += dstLeft >> zoom
	// drawMaskedZoom<zoom>
	sh, sw := h>>zoom, w>>zoom
	if sh == 0 || sw == 0 {
		return
	}
	skip := 1 << zoom
	low := skip - 1
	scaledRowSize := int(uint16(w))&low + low*int(uint16(rowSize+w)) + rowSize
	dstWrap -= sw
	for row := 0; row < sh; row++ {
		for col := 0; col < sw; col++ {
			if data >= 0 && data < len(img.Pix) && data < len(mask.Pix) && dst >= 0 && dst < len(rt.Bits) {
				if v := mask.Pix[data] & img.Pix[data]; v != 0 {
					rt.Bits[dst] = v
				}
			}
			data += skip
			dst++
		}
		data += scaledRowSize
		dst += dstWrap
	}
}
