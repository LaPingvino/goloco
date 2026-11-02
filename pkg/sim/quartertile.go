package sim

// rotateQuarterTile is QuarterTile::rotate: both nibbles (base quarters, z
// quarters) rotated left by amount within 4 bits (the int arithmetic of
// upstream: bits shifted past bit 7 come back through the >> 4).
func rotateQuarterTile(v, amount uint8) uint8 {
	if amount == 0 || amount > 3 {
		return v
	}
	r1 := uint32(v) << amount
	r2 := r1 >> 4
	lowMask := [4]uint32{0, 0b11101110, 0b11001100, 0b10001000}[amount]
	highMask := [4]uint32{0, 0b00010001, 0b00110011, 0b01110111}[amount]
	return uint8((r1 & lowMask) | (r2 & highMask))
}

// rotl4bit is Numerics::rotl4bit.
func rotl4bit(v, rotation uint8) uint8 {
	return uint8((uint32(v)<<rotation | uint32(v)>>(4-rotation)) & 0xF)
}
