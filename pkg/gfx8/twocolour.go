package gfx8

// SecondaryRemap is PaletteIndex::secondaryRemap0..B (0xCA..0xD5).
var SecondaryRemap = [12]uint8{0xCA, 0xCB, 0xCC, 0xCD, 0xCE, 0xCF, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5}

// TwoColourMap is PaletteMap::getForImage for an image with a secondary
// colour (Gfx::recolour2): the default (identity) map with the primary remap
// entries taken from the primary colour's map and the secondary remap entries
// (0xCA..0xD5) taken from the secondary colour's map at its primary remap
// entries.
func TwoColourMap(primary, secondary []uint8) []uint8 {
	m := make([]uint8, 256)
	for i := range m {
		m[i] = uint8(i)
	}
	for k, idx := range PrimaryRemap {
		m[idx] = primary[idx]
		m[SecondaryRemap[k]] = secondary[idx]
	}
	return m
}
