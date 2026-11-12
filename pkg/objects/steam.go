package objects

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// SteamObject is Objects/SteamObject.h (0x28 bytes) as SteamObject::load
// (0x00440CAD) fills it: a vehicle's smoke, steam or sparks. Its two frame
// runs (type 0 and type 1 — Exhaust::isSubObjType1) give each frame's image
// and how far the puff rises; the sprite extents are measured from its
// images (getImagesMaxExtent).
type SteamObject struct {
	Header                                        *ObjectHeader
	Name                                          string
	NumImages                                     uint16
	NumStationaryTicks                            uint8
	SpriteWidth, SpriteHeightNeg, SpriteHeightPos uint8
	Flags                                         uint16
	Image                                         uint32 // baseImageId
	Var0A                                         uint32 // var_0A: the painter adds it to every frame's image
	Frames                                        [2][]SteamFrame
	SoundNames                                    []string // soundEffects' headers (numSoundEffects at 0x1E)
}

// SteamFrame is SteamObject::ImageAndHeight.
type SteamFrame struct{ ImageOffset, Height uint8 }

// SteamObjectFlags.
const (
	SteamFlagApplyWind           = 1 << 0
	SteamFlagDisperseOnCollision = 1 << 1
	SteamFlagHasTunnelSounds     = 1 << 2
	SteamFlagUnk3                = 1 << 3 // painted with a 24×24 box about its centre
)

const steamObjectSize = 0x28

// ParseSteamObject is SteamObject::load: the fixed struct, the name, the two
// 0xFF-terminated frame runs, the sound headers, the image table (g may be
// nil: no images).
func ParseSteamObject(header *ObjectHeader, data []byte, g G1Loader, ext extentMeasurer) (*SteamObject, error) {
	if len(data) < steamObjectSize {
		return nil, fmt.Errorf("steam object too small: %d bytes", len(data))
	}
	o := &SteamObject{Header: header,
		NumImages:          binary.LittleEndian.Uint16(data[0x02:]),
		NumStationaryTicks: data[0x04],
		Flags:              binary.LittleEndian.Uint16(data[0x08:]),
		Var0A:              binary.LittleEndian.Uint32(data[0x0A:]),
	}
	off, name := walkStringTableLang(data, steamObjectSize)
	o.Name = name
	for k := range o.Frames {
		for off+1 < len(data) && data[off] != 0xFF {
			o.Frames[k] = append(o.Frames[k], SteamFrame{data[off], data[off+1]})
			off += 2
		}
		off++ // 0xFF
	}
	for range int(data[0x1E]) {
		o.SoundNames = append(o.SoundNames, headerName(data, off))
		off += HeaderSize
	}
	if g != nil && off <= len(data) {
		res, err := g.LoadImageTable(data[off:])
		if err != nil {
			return nil, fmt.Errorf("steam %q image table: %w", name, err)
		}
		if int(res.TableLength) != len(data)-off { // load's assert: the walk ends at the table
			return nil, fmt.Errorf("steam %q: image table %d bytes, %d left", name, res.TableLength, len(data)-off)
		}
		o.Image = res.ImageOffset
		if ext != nil {
			o.SpriteWidth, o.SpriteHeightNeg, o.SpriteHeightPos = ext.ImagesMaxExtent(o.Image, int(o.NumImages))
		}
	}
	return o, nil
}

// SteamObjectIndex is the slot of the steam object a vehicle's emitter names
// (VehicleObject::load's findObjectHandle); -1 when it isn't loaded.
func (m *ObjectManager) SteamObjectIndex(name string) int {
	for i, s := range m.SteamObjects {
		if s != nil && s.Header != nil && strings.EqualFold(s.Header.GetName(), name) {
			return i
		}
	}
	return -1
}
