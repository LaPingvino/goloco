package assets

// The open build's colour remap tables (item O, docs/open-build.md "Colour
// remap tables"): the game's tables exist only in g1.DAT. Made here from
// OpenLoco's own names (Colour.h, MIT): a colour whose name has a swatch in
// PaletteIndex maps its 12 remap slots (primaryRemap 7–9, then 246–254) onto
// that swatch's 12 shades, dark to light. Checked locally against the
// original tables (nothing of them kept): exact for the 16 colours listed;
// black, amber and brown have a same-named swatch but use it otherwise, so
// they are left out, with the colours that have no swatch of their own (grey,
// white, the "dark" variants) — no map rather than a wrong one.

// remapSlots are the primary remap palette indices, in shade order.
var remapSlots = [12]uint8{7, 8, 9, 246, 247, 248, 249, 250, 251, 252, 253, 254}

// colourSwatch is each Colour's same-named PaletteIndex swatch (its shade 0),
// 0 for none — both names from OpenLoco's Colour.h.
var colourSwatch = [32]uint8{
	4:  0x76, // mutedPurple
	5:  0x9A, // purple
	7:  0x82, // blue
	8:  0xBE, // mutedDarkTeal
	11: 0x8E, // mutedSeaGreen
	12: 0x46, // mutedGrassGreen
	13: 0x5E, // green
	14: 0x52, // mutedAvocadoGreen
	15: 0x16, // mutedOliveGreen
	16: 0x2E, // yellow
	18: 0xB2, // orange
	21: 0x22, // mutedDarkYellow
	24: 0x6A, // mutedOrange
	25: 0x3A, // mutedDarkRed
	27: 0xA6, // red
	29: 0xCA, // pink
}

// OpenRemap is colour c's remap table by that rule (ok false: no swatch).
func OpenRemap(c int) ([]byte, bool) {
	if c < 0 || c >= len(colourSwatch) || colourSwatch[c] == 0 {
		return nil, false
	}
	m := make([]byte, 256)
	for i := range m {
		m[i] = byte(i)
	}
	for k, slot := range remapSlots {
		m[slot] = colourSwatch[c] + uint8(k)
	}
	return m, true
}
