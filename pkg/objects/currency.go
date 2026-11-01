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
}

// CurrencyObjectSize is sizeof(CurrencyObject).
const CurrencyObjectSize = 0x0C

// ParseCurrencyObject reads a currency object's fields and strings.
func ParseCurrencyObject(header *ObjectHeader, data []byte) (*CurrencyObject, error) {
	if len(data) < CurrencyObjectSize {
		return nil, fmt.Errorf("currency object too short (%d bytes)", len(data))
	}
	c := &CurrencyObject{Header: header, Separator: data[0x0A], Factor: data[0x0B]}
	off := CurrencyObjectSize
	off, c.Name = walkStringTableLang(data, off)
	off, c.Prefix = walkStringTableLang(data, off)
	_, c.Suffix = walkStringTableLang(data, off)
	return c, nil
}
