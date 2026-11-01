package objects

import (
	"fmt"
	"strings"
)

// TunnelObject — a port of OpenLoco Objects/TunnelObject.cpp load(): a string
// table then the image table (4 portal images: two per drawn edge).
type TunnelObject struct {
	Header *ObjectHeader
	Name   string
	Image  uint32
}

// ParseTunnelObjectWithG1 parses a tunnel object (g1 may be nil: no images).
func ParseTunnelObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*TunnelObject, error) {
	if len(data) < TunnelObjectSize {
		return nil, fmt.Errorf("tunnel object too small: %d bytes", len(data))
	}
	o := &TunnelObject{Header: header}
	n, name := parseStringTable(data[TunnelObjectSize:])
	o.Name = name
	if g1 != nil {
		res, err := g1.LoadImageTable(data[TunnelObjectSize+n:])
		if err != nil {
			return nil, fmt.Errorf("tunnel %q image table: %w", name, err)
		}
		o.Image = res.ImageOffset
	}
	return o, nil
}

// headerName is the name field of an embedded ObjectHeader at data[off:]
// (bytes 4..11, NUL-padded), as GetName returns it.
func headerName(data []byte, off int) string {
	if off+HeaderSize > len(data) {
		return ""
	}
	return strings.TrimRight(string(data[off+4:off+12]), "\x00 ")
}

// TunnelImage is trackObj->tunnel / roadObj->tunnel resolved to the tunnel
// object's image base: upstream resolves the embedded tunnel header to the
// loaded object's id at load time; looking the same header name up gives that
// object. -1 when it is not loaded (upstream: tunnel 0xFF).
func (m *ObjectManager) TunnelImage(name string) int {
	if name == "" {
		return -1
	}
	if obj := m.GetObject(strings.ToUpper(name)); obj != nil {
		if t, ok := obj.Object.(*TunnelObject); ok {
			return int(t.Image)
		}
	}
	return -1
}
