package objects

import "fmt"

// CurrencyObject is CurrencyObject.h: a 12-byte header (name, prefix and
// suffix StringIds, objectIcon, separator, factor) followed by three string
// tables (name, prefix symbol, suffix symbol) and the image table.
type CurrencyObject struct {
	Header            *ObjectHeader
	Name              string // "British Pounds"
	Prefix, Suffix    string
	Separator, Factor uint8
	// ObjectIcon is the image table's first image: the currency symbol for the
	// small, medium normal, medium bold and large fonts (+0..+3).
	ObjectIcon uint32
}

// CurrencyObjectSize is sizeof(CurrencyObject).
const CurrencyObjectSize = 0x0C

// ParseCurrencyObject reads a currency object's fields and strings, and with a
// G1 loader its image table (objectIcon).
func ParseCurrencyObject(header *ObjectHeader, data []byte, g1 G1Loader) (*CurrencyObject, error) {
	if len(data) < CurrencyObjectSize {
		return nil, fmt.Errorf("currency object too short (%d bytes)", len(data))
	}
	c := &CurrencyObject{Header: header, Separator: data[0x0A], Factor: data[0x0B]}
	off := CurrencyObjectSize
	off, c.Name = walkStringTableLang(data, off)
	off, c.Prefix = walkStringTableLang(data, off)
	off, c.Suffix = walkStringTableLang(data, off)
	if g1 != nil && off < len(data) {
		res, err := g1.LoadImageTable(data[off:])
		if err != nil {
			return nil, fmt.Errorf("currency %q image table: %w", c.Name, err)
		}
		c.ObjectIcon = res.ImageOffset
	}
	return c, nil
}
