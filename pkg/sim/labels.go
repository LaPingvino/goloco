package sim

// Name labels' frames: Town::updateLabel (Town.cpp) and Station::updateLabel
// (Station.cpp 0x0048DCA5) — the box a town's or station's name takes at
// each zoom, which drawTownNames / drawStationNames then test and draw at.
// Rotation 0 (goloco's views don't rotate).

// G1 fonts (Gfx::Font, as glyph blocks after g1FontBase).
const (
	fontMediumNormal = 0
	fontMediumBold   = 1
	fontSmall        = 2
)

var (
	kZoomToTownFonts    = [4]int{fontMediumBold, fontMediumBold, fontMediumNormal, fontMediumNormal}
	kZoomToStationFonts = [4]int{fontMediumBold, fontMediumBold, fontSmall, fontSmall}
	// kZoomToStationBorder's end width and height by zoom.
	kZoomToStationBorder = [4]struct{ width, height int }{{3, 11}, {3, 11}, {1, 7}, {1, 7}}
)

// stringWidth is Gfx::TextRenderer::getStringWidth on plain text in a font:
// each character's glyph width − 1 (characters 123–150 none). Runes are
// Loco codes (names are ASCII; the transport icons 177–184).
func stringWidth(g1 G1Raw, font int, s string) int {
	w := 0
	for _, r := range s {
		chr := int(r)
		if chr < 32 || chr > 255 || chr >= 123 && chr <= 150 {
			continue
		}
		if gw, _, _, ok := g1.ElementRaw(g1FontBase + font*224 + chr - 32); ok {
			w += int(uint8(int8(gw) - 1))
		}
	}
	return w
}

// gameToScreen0 is World::gameToScreen at rotation 0.
func gameToScreen0(x, y, z int) (int, int) { return y - x, ((x + y) >> 1) - z }

// UpdateTownLabel is Town::updateLabel: the name centred 48 px above the
// town's tile centre on the ground, 11 high, 2 wider than the text in that
// zoom's font — computed in zoom-0 pixels and shifted down to each zoom.
func (m *Map) UpdateTownLabel(i int, g1 G1Raw) {
	t := &m.Towns[i]
	name := m.TownName(t.Name, i)
	land, _ := m.GetHeight(int32(t.X), int32(t.Y))
	sx, sy := gameToScreen0(int(t.X)+TileSize/2, int(t.Y)+TileSize/2, int(land))
	sy -= 48
	for z := 0; z < 4; z++ {
		w := (stringWidth(g1, kZoomToTownFonts[z], name) + 2) << z
		h := 11 << z
		x0, y0 := sx-w/2, sy-h/2
		f := &t.LabelFrame
		f.Left[z], f.Right[z] = int16(x0>>z), int16((x0+w)>>z)
		f.Top[z], f.Bottom[z] = int16(y0>>z), int16((y0+h)>>z)
	}
}

// stationIcons is getTransportIconsFromStationFlags's text: the transport
// glyphs of the modes the station serves, rail, road, air, water in order
// (the label_icons_* strings).
func stationIcons(flags uint16) string {
	s := ""
	for i, c := range [4]rune{177, 182, 183, 184} {
		if flags&(1<<i) != 0 {
			s += string(c)
		}
	}
	return s
}

// UpdateStationLabel is Station::updateLabel: "{name} {icons}" in that
// zoom's font plus both ends, the end's height, centred on the station's
// position and 32 px up — sized in viewport pixels (<< zoom) and placed back
// in that zoom's (>> zoom).
func (m *Map) UpdateStationLabel(id int, g1 G1Raw) {
	st := &m.StationNames[id]
	text := m.StationName(id) + " " + stationIcons(st.Flags)
	vx, vy := gameToScreen0(int(st.X), int(st.Y), int(st.Z))
	for z := 0; z < 4; z++ {
		b := kZoomToStationBorder[z]
		w := (stringWidth(g1, kZoomToStationFonts[z], text) + b.width*2) << z
		h := b.height << z
		left := vx - w/2
		top := vy - h/2 - 32
		f := &st.LabelFrame
		f.Left[z], f.Right[z] = int16(left>>z), int16((left+w)>>z)
		f.Top[z], f.Bottom[z] = int16(top>>z), int16((top+h)>>z)
	}
}
