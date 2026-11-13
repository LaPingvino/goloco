package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// TextRenderer's word wrap (OpenLoco reference: src/OpenLoco/src/Graphics/
// TextRenderer.cpp wrapString 0x00495301, wrapStringTicker 0x0049544E,
// drawStringCentredWrapped, drawStringLeftWrapped, drawStringTicker) for
// text in one font whose only control code is the newline (5): what the
// messages are (MessageManager::post formats them, their arguments filled
// in).

const controlNewline = 5 // ControlCodes::newline

// FromLocoBytes is a Loco-encoded string (a message's text, an object's
// string: one byte a character) as goloco's strings carry it — Loco codes as
// runes.
func FromLocoBytes(s string) string {
	r := make([]rune, len(s))
	for i := 0; i < len(s); i++ {
		r[i] = rune(s[i])
	}
	return string(r)
}

func charWidth(font int, r rune) int {
	if globalG1Font == nil || r < 32 || r > 255 {
		return 0
	}
	return int(globalG1Font.widths[font][r-32])
}

// wrapLines is wrapString and wrapStringTicker in one: the lines the text
// breaks into for a width, and — the ticker's — how many lines it takes to
// show the first numChars characters.
//
// A line runs until its width reaches the limit (the character that reaches
// it included), then goes back to its last space, which the break replaces;
// a line without a space breaks where it stands. A newline ends a line. As
// upstream: a text ending in a newline has an empty last line, and the
// ticker's count doesn't grow for a line broken inside a word.
func wrapLines(text string, font, width, numChars int) (lines []string, tickerLines int) {
	buf := []rune(text)
	tickerLines = 1
	charNum := numChars
	for ptr := 0; ptr < len(buf); {
		start, wordStart := ptr, ptr
		lineWidth, lastWordCharNum := 0, charNum
		forced := false
		for ; ptr < len(buf) && lineWidth < width; ptr++ {
			c := buf[ptr]
			if c == controlNewline {
				lines = append(lines, string(buf[start:ptr]))
				ptr++
				if charNum > 0 {
					tickerLines++
				}
				forced = true
				break
			}
			if c < 32 {
				continue
			}
			if c == ' ' {
				wordStart, lastWordCharNum = ptr, charNum
			}
			charNum--
			lineWidth += charWidth(font, c)
		}
		if forced {
			if ptr == len(buf) {
				lines = append(lines, "") // wrapCount++ once more at the string's end
			}
			continue
		}
		if start == wordStart || (ptr == len(buf) && lineWidth < width) {
			lines = append(lines, string(buf[start:ptr]))
			continue
		}
		charNum = lastWordCharNum
		if charNum > 0 {
			tickerLines++
		}
		lines = append(lines, string(buf[start:wordStart]))
		ptr = wordStart + 1
	}
	return lines, tickerLines
}

// WrapString is wrapString: the text's lines for a width in a font.
func WrapString(text string, font, width int) []string {
	lines, _ := wrapLines(text, font, width, 0)
	return lines
}

// DrawStringCentredWrapped is drawStringCentredWrapped: the lines centred on
// x, the block centred on y by half line heights ((lineHeight / 2) × (lines −
// 1) up).
func DrawStringCentredWrapped(screen *ebiten.Image, text string, font, x, y, width int, clr color.Color) {
	lines := WrapString(text, font, width)
	lh := g1FontHeight(font)
	y -= (lh / 2) * (len(lines) - 1)
	for _, l := range lines {
		DrawTextFont(screen, l, font, x-MeasureTextFont(l, font)/2, y, clr)
		y += lh
	}
}

// DrawStringLeftWrapped is drawStringLeftWrapped; the y below the last line.
func DrawStringLeftWrapped(screen *ebiten.Image, text string, font, x, y, width int, clr color.Color) int {
	lh := g1FontHeight(font)
	for _, l := range WrapString(text, font, width) {
		DrawTextFont(screen, l, font, x, y, clr)
		y += lh
	}
	return y
}

// DrawStringTicker is drawStringTicker: the first numChars characters of the
// wrapped text, each line centred on x, scrolled up so that the last of the
// numLines lines it may show is the one being written (drawStringMaxChars:
// every character of a line counts, drawn or scrolled out of sight).
func DrawStringTicker(screen *ebiten.Image, text string, font, x, y, numLines, numChars, width int, clr color.Color) {
	lines, need := wrapLines(text, font, width, numChars)
	lh := g1FontHeight(font)
	if from := need - numLines; from > 0 {
		y -= lh * from
	}
	left := numChars
	for i := 0; i < len(lines) && i < need; i++ {
		l := []rune(lines[i])
		px := x - MeasureTextFont(lines[i], font)/2
		n := min(left, len(l))
		DrawTextFont(screen, string(l[:n]), font, px, y, clr)
		left -= n
		y += lh
	}
}
