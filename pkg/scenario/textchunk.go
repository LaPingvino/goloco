package scenario

import (
	"bytes"
	"encoding/binary"

	"github.com/LaPingvino/goloco/pkg/i18n"
)

// GolocoTextMagic opens goloco's text chunk (a deviation, LOOP.md N56): the
// user strings and messages holding characters the Loco code page lacks.
// The 8-bit tables get "?" for them as upstream's convertUnicodeToLoco would;
// this chunk, after the tile elements where OpenLoco's loadSave never reads,
// keeps the text itself. Layout: the magic, a version (u16), then records of
// kind (u8: 0 user string, 1 message), index (u16), length (u16), bytes — a
// user string's UTF-8, a message's in-memory Loco bytes (i18n.LocoWide).
const GolocoTextMagic = "GOLOCOTX"

const textChunkVersion = 1

// EncodeTextChunk is the chunk for sc's wide texts; nil when there are none
// (the file is then exactly OpenLoco's).
func (sc *Scenario) EncodeTextChunk() []byte {
	var b bytes.Buffer
	rec := func(kind uint8, i int, s []byte) {
		if b.Len() == 0 {
			b.WriteString(GolocoTextMagic)
			binary.Write(&b, binary.LittleEndian, uint16(textChunkVersion))
		}
		binary.Write(&b, binary.LittleEndian, kind)
		binary.Write(&b, binary.LittleEndian, uint16(i))
		binary.Write(&b, binary.LittleEndian, uint16(len(s)))
		b.Write(s)
	}
	for i, s := range sc.UserStrings {
		if i < userStringsCount && i18n.HasWide(s) {
			rec(0, i, []byte(s))
		}
	}
	for i, m := range sc.Messages {
		if i < 200 && i18n.HasWideLoco(m.Text) {
			rec(1, i, m.Text)
		}
	}
	if b.Len() == 0 {
		return nil
	}
	return b.Bytes()
}

// decodeTextChunk lays the chunk's texts over the 8-bit ones read from the
// game state — each only where the 8-bit text is still what this one saves
// as (an edit by another program wins).
func (sc *Scenario) decodeTextChunk(c []byte) {
	c = c[len(GolocoTextMagic):]
	if len(c) < 2 || binary.LittleEndian.Uint16(c) != textChunkVersion {
		return
	}
	c = c[2:]
	for len(c) >= 5 {
		kind, i, n := c[0], int(binary.LittleEndian.Uint16(c[1:])), int(binary.LittleEndian.Uint16(c[3:]))
		if len(c) < 5+n {
			return
		}
		s := c[5 : 5+n]
		c = c[5+n:]
		switch kind {
		case 0:
			if i < len(sc.UserStrings) {
				b := make([]byte, userStringSize)
				putCString(b, string(s), userStringSize)
				if string(bytes.TrimRight(b, "\x00")) == locoRunesBytes(sc.UserStrings[i]) {
					sc.UserStrings[i] = string(s)
				}
			}
		case 1:
			if i < len(sc.Messages) {
				t := i18n.StripWide(s)
				if len(t) > 197 {
					t = t[:197]
				}
				if bytes.Equal(t, sc.Messages[i].Text) {
					sc.Messages[i].Text = append([]byte(nil), s...)
				}
			}
		}
	}
}

// locoRunesBytes is a user string as parseUserStrings gives it (one rune a
// byte) back as its bytes.
func locoRunesBytes(s string) string {
	b := make([]byte, 0, len(s))
	for _, r := range s {
		b = append(b, byte(r))
	}
	return string(b)
}
