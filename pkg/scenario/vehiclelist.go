package scenario

import "encoding/binary"

// Offsets checked with the compiler against S5::GameState / S5Entity.h.
const (
	gameStateCurrentDayOffset = 0x14 // uint32 general.currentDay (days since 1 Jan 1800... the game epoch)
	vHeadOrdinal              = 0x44 // int16 VehicleHead::ordinalNumber
	vVeh1DayCreated           = 0x4A // uint32 Vehicle1::dayCreated
	vVeh2Profit               = 0x62 // currency32_t Vehicle2::profit[4]
	vVeh2Reliability          = 0x72 // uint8 Vehicle2::reliability
	// getStatus's fields (S5Entity.h / S5GameState.h, from the compiler).
	vHeadTileX            = 0x30 // int16 tileX (-1: not placed)
	vHeadVehicleFlags     = 0x0C // uint16 EntityBase vehicleFlags
	vHeadOrderTableOffset = 0x46 // uint32
	vHeadCurrentOrder     = 0x4A // uint16
	vHeadSizeOfOrderTable = 0x4C // uint16
	vHeadStationID        = 0x54 // uint16
	vHeadManualPower      = 0x6E // int8
	vVeh2CurrentSpeed     = 0x56 // int32 Speed32 (16.16 mph)
	vVeh2Flags73          = 0x73 // uint8 bit 0 isBrokenDown
	gameStateOrders       = 0x461E44
	gameStateOrdersType2  = 0x460224
	maxOrders             = 256000
)

// ParseOrders returns GameState orders[] (the shared order table).
// GameState routings[kMaxVehicles][kMaxRoutingsPerVehicle] (uint16): directly
// after userStrings, so 0x442A44 (GameStateType2: 0x440E24, the same 0x1C20
// shift as userStrings and orders).
const (
	gameStateRoutings      = 0x442A44
	gameStateRoutingsType2 = 0x440E24
	MaxRoutingVehicles     = 1000
	RoutingsPerVehicle     = 64
)

// ParseRoutings reads the routing table (RoutingManager), flat:
// [vehicleRef*64 + index] — i.e. indexed by a RoutingHandle's _data.
func ParseRoutings(gs []byte) []uint16 {
	off := gameStateRoutings
	if len(gs) == GameStateSizeType2 {
		off = gameStateRoutingsType2
	}
	n := MaxRoutingVehicles * RoutingsPerVehicle
	if len(gs) < off+2*n {
		return nil
	}
	r := make([]uint16, n)
	for i := range r {
		r[i] = binary.LittleEndian.Uint16(gs[off+2*i:])
	}
	return r
}

func ParseOrders(gs []byte) []byte {
	off := gameStateOrders
	if len(gs) == GameStateSizeType2 {
		off = gameStateOrdersType2
	}
	if len(gs) < off+maxOrders {
		return nil
	}
	return append([]byte(nil), gs[off:off+maxOrders]...)
}

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
	// getStatus: the head's tileX, vehicleFlags, orders, stationId and
	// manualPower; veh2's broken-down flag and currentSpeed.
	TileX            int16
	VehicleFlags     uint16
	OrderTableOffset uint32
	CurrentOrder     uint16
	SizeOfOrderTable uint16
	StationID        uint16
	ManualPower      int8
	BrokenDown       bool
	CurrentSpeed     int32
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
				TileX: int16(u16(e, vHeadTileX)), VehicleFlags: u16(e, vHeadVehicleFlags),
				OrderTableOffset: binary.LittleEndian.Uint32(e[vHeadOrderTableOffset:]), CurrentOrder: u16(e, vHeadCurrentOrder),
				SizeOfOrderTable: u16(e, vHeadSizeOfOrderTable), StationID: u16(e, vHeadStationID), ManualPower: int8(e[vHeadManualPower]),
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
		out[i].BrokenDown = p.e[vVeh2Flags73]&1 != 0
		out[i].CurrentSpeed = int32(binary.LittleEndian.Uint32(p.e[vVeh2CurrentSpeed:]))
	}
	return out
}
