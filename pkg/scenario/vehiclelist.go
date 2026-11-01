package scenario

import "encoding/binary"

// Offsets checked with the compiler against S5::GameState / S5Entity.h.
const (
	gameStateCurrentDayOffset = 0x14 // uint32 general.currentDay (days since 1 Jan 1800... the game epoch)
	vHeadOrdinal              = 0x44 // int16 VehicleHead::ordinalNumber
	vVeh1DayCreated           = 0x4A // uint32 Vehicle1::dayCreated
	vVeh2Profit               = 0x62 // currency32_t Vehicle2::profit[4]
	vVeh2Reliability          = 0x72 // uint8 Vehicle2::reliability
)

// VehicleListEntry is what VehicleList.cpp's drawScroll reads of one train:
// the head's name, ordinal, type and owner, veh1's day created and veh2's
// recent profits and reliability.
type VehicleListEntry struct {
	HeadID       uint16
	Owner        uint8
	VehicleType  uint8 // 0 train, 1 bus, 2 truck, 3 tram, 4 aircraft, 5 ship
	Name         uint16
	Ordinal      int16
	Status       uint8
	DayCreated   uint32
	RecentProfit int32 // veh2->totalRecentProfit()
	Reliability  uint8
}

// ParseCurrentDay returns GameState general.currentDay.
func ParseCurrentDay(gs []byte) uint32 {
	if len(gs) < gameStateCurrentDayOffset+4 {
		return 0
	}
	return binary.LittleEndian.Uint32(gs[gameStateCurrentDayOffset:])
}

// ParseVehicleListEntries collects every vehicle head with its veh1 and veh2,
// in entity-slot order.
func ParseVehicleListEntries(gs []byte) []VehicleListEntry {
	tableOffset := EntityTableOffset
	if len(gs) == GameStateSizeType2 {
		tableOffset = EntityTableOffsetType2
	}
	if len(gs) < tableOffset+MaxEntities*EntitySize {
		return nil
	}
	u16 := func(e []byte, o int) uint16 { return binary.LittleEndian.Uint16(e[o:]) }
	var out []VehicleListEntry
	idx := map[uint16]int{}
	type part struct {
		typ  uint8
		head uint16
		e    []byte
	}
	var parts []part
	for i := 0; i < MaxEntities; i++ {
		e := gs[tableOffset+i*EntitySize : tableOffset+(i+1)*EntitySize]
		if e[ebBaseType] != EntityBaseTypeVehicle {
			continue
		}
		switch e[ebType] {
		case VehicleTypeHead:
			idx[u16(e, ebID)] = len(out)
			out = append(out, VehicleListEntry{
				HeadID: u16(e, ebID), Owner: e[ebOwner], VehicleType: e[vHeadVehicleType],
				Name: u16(e, ebName), Ordinal: int16(u16(e, vHeadOrdinal)), Status: e[vHeadStatus],
			})
		case VehicleTypeVeh1, VehicleTypeVeh2:
			parts = append(parts, part{e[ebType], u16(e, vhCommonHead), e})
		}
	}
	for _, p := range parts {
		i, ok := idx[p.head]
		if !ok {
			continue
		}
		if p.typ == VehicleTypeVeh1 {
			out[i].DayCreated = binary.LittleEndian.Uint32(p.e[vVeh1DayCreated:])
			continue
		}
		var sum int32
		for m := range 4 {
			sum += int32(binary.LittleEndian.Uint32(p.e[vVeh2Profit+m*4:]))
		}
		out[i].RecentProfit = sum
		out[i].Reliability = p.e[vVeh2Reliability]
	}
	return out
}
