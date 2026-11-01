package objects

import "fmt"

// SnowObject — a port of OpenLoco Objects/SnowObject.cpp load(): a string table
// then the image table. ImageOffset is the table's offset; the painter adds
// SnowLine::ImageIds::surfaceFullZoom (pkg/paint SnowSurfaceFullZoom).
type SnowObject struct {
	Header      *ObjectHeader
	Name        string
	ImageOffset uint32
}

// ParseSnowObjectWithG1 parses a snow object; size is sizeof(SnowObject).
func ParseSnowObjectWithG1(header *ObjectHeader, data []byte, size int, g1 G1Loader) (*SnowObject, error) {
	if len(data) < size {
		return nil, fmt.Errorf("snow object too small: %d bytes", len(data))
	}
	o := &SnowObject{Header: header}
	n, name := parseStringTable(data[size:])
	o.Name = name
	if g1 != nil {
		res, err := g1.LoadImageTable(data[size+n:])
		if err != nil {
			return nil, fmt.Errorf("snow %q image table: %w", name, err)
		}
		o.ImageOffset = res.ImageOffset
	}
	return o, nil
}
