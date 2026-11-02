package objects

// ClimateObject is ClimateObject.h: the seasons and snow lines.
type ClimateObject struct {
	Header         *ObjectHeader
	Name           string
	FirstSeason    uint8    // 0x02
	SeasonLength   [4]uint8 // 0x03
	WinterSnowLine uint8    // 0x07
	SummerSnowLine uint8    // 0x08
}

// ParseClimateObject is ClimateObject::load (0x0A bytes, then the name).
func ParseClimateObject(header *ObjectHeader, data []byte) *ClimateObject {
	c := &ClimateObject{Header: header}
	if len(data) >= 0x0A {
		c.FirstSeason = data[0x02]
		copy(c.SeasonLength[:], data[0x03:0x07])
		c.WinterSnowLine, c.SummerSnowLine = data[0x07], data[0x08]
		_, c.Name = walkStringTableLang(data, 0x0A)
	}
	return c
}

// ReorderClimate sets the scenario's climate object (its one slot).
func (m *ObjectManager) ReorderClimate(order []string) {
	m.Climate = nil
	if len(order) > 0 && order[0] != "" {
		if obj := m.GetObject(order[0]); obj != nil {
			m.Climate, _ = obj.Object.(*ClimateObject)
		}
	}
}
