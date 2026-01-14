package economy

// Currency type definitions

// Currency32 is a 32-bit currency value
type Currency32 int32

// Currency48 is a 48-bit currency value (6 bytes)
// Used for large monetary amounts in the game
type Currency48 struct {
	Low  uint32
	High int16
}

// NewCurrency48 creates a Currency48 from an int64
func NewCurrency48(value int64) Currency48 {
	return Currency48{
		Low:  uint32(value & 0xFFFFFFFF),
		High: int16((value >> 32) & 0xFFFF),
	}
}

// AsInt64 converts Currency48 to int64
func (c *Currency48) AsInt64() int64 {
	return int64(c.Low) | (int64(c.High) << 32)
}

// Add adds a Currency32 value
func (c *Currency48) Add(value Currency32) Currency48 {
	return NewCurrency48(c.AsInt64() + int64(value))
}

// AddCurrency48 adds another Currency48
func (c *Currency48) AddCurrency48(other Currency48) Currency48 {
	return NewCurrency48(c.AsInt64() + other.AsInt64())
}

// Subtract subtracts a Currency32 value
func (c *Currency48) Subtract(value Currency32) Currency48 {
	return NewCurrency48(c.AsInt64() - int64(value))
}

// SubtractCurrency48 subtracts another Currency48
func (c *Currency48) SubtractCurrency48(other Currency48) Currency48 {
	return NewCurrency48(c.AsInt64() - other.AsInt64())
}

// IsZero returns true if the currency is zero
func (c *Currency48) IsZero() bool {
	return c.Low == 0 && c.High == 0
}

// IsNegative returns true if the currency is negative
func (c *Currency48) IsNegative() bool {
	return c.High < 0
}

// IsPositive returns true if the currency is positive
func (c *Currency48) IsPositive() bool {
	return c.High > 0 || (c.High == 0 && c.Low > 0)
}
