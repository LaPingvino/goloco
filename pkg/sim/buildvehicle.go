package sim

import (
	"sort"
	"strings"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// The Build Vehicle window's list (BuildVehicle.cpp generateBuildableVehiclesArray
// 0x4B9165): which vehicles it offers for a vehicle type and a track-type tab,
// filtered and sorted as the window's options say (item 5P.f).

// Build Vehicle filter flags (VehicleFilterFlags) and sort keys (VehicleSortBy).
const (
	BuildFilterPowered   = 1 << 0
	BuildFilterUnpowered = 1 << 1
	BuildFilterLocked    = 1 << 2
	BuildFilterUnlocked  = 1 << 3

	BuildSortDesignYear = 0
	BuildSortName       = 1
	BuildSortCost       = 2
	BuildSortCapacity   = 3
	BuildSortPower      = 4
	BuildSortMaxSpeed   = 5
	BuildSortObsolete   = 6
	BuildSortWeight     = 7
	BuildSortLength     = 8
)

// BuildVehicleOptions are the window's statics: _vehicleFilterFlags (default
// all four), _vehicleSortBy, _vehicleSortAscending (default true),
// _cargoSupportedFilter (0xFF any, 0xFE cargoless, else a cargo type), the
// search box's text, Config::displayLockedVehicles.
type BuildVehicleOptions struct {
	Filter         uint8
	SortBy         int
	Ascending      bool
	CargoFilter    uint8
	Search         string
	DisplayLocked  bool
	TargetHead     uint16 // a vehicle being added to (the Vehicle window's), or entityIDNull
	ControllingCmp uint8  // CompanyManager::getControllingId
}

// DefaultBuildVehicleOptions are the window's starting statics.
func DefaultBuildVehicleOptions(company uint8) BuildVehicleOptions {
	return BuildVehicleOptions{Filter: BuildFilterPowered | BuildFilterUnpowered | BuildFilterLocked | BuildFilterUnlocked,
		Ascending: true, CargoFilter: 0xFF, TargetHead: entityIDNull, ControllingCmp: company}
}

const roadFlagAnyRoadTypeCompatibleObj = 1 << 3 // RoadObjectFlags::anyRoadTypeCompatible

// BuildableVehicles is generateBuildableVehiclesArray: the vehicle object ids
// offered for vehicleType on the track-type tab trackType (0xFF all; a road
// type with bit 7; a rail track type).
func (m *Map) BuildableVehicles(om *objects.ObjectManager, vehicleType, trackType uint8, o BuildVehicleOptions) []uint16 {
	if trackType != 0xFF && trackType&(1<<7) != 0 {
		idx := int(trackType &^ (1 << 7))
		if idx < len(om.RoadObjects) && om.RoadObjects[idx] != nil && om.RoadObjects[idx].Flags&roadFlagAnyRoadTypeCompatibleObj != 0 {
			trackType = 0xFE
		}
	}
	company := o.ControllingCmp
	var target *Train
	if o.TargetHead != entityIDNull {
		if tr, err := m.TrainOf(o.TargetHead); err == nil {
			target = &tr
			company = m.Entities[tr.Head].Owner
		}
	}
	type bv struct {
		id                   uint16
		name                 string
		powered              bool
		designed, obsolete   uint16
		cost                 int32
		power, speed, weight uint16
		length               int
		capacity             uint8
	}
	var list []bv
	showUnpowered, showPowered := o.Filter&BuildFilterUnpowered != 0, o.Filter&BuildFilterPowered != 0
	showUnlocked, showLocked := o.Filter&BuildFilterUnlocked != 0, o.Filter&BuildFilterLocked != 0 && o.DisplayLocked
	for id, v := range om.Vehicles {
		if v == nil {
			continue
		}
		oid := uint16(id)
		if target != nil && !m.isVehicleTypeCompatible(*target, oid, om) {
			continue
		}
		if uint8(v.Type) != vehicleType {
			continue
		}
		unlocked := m.isVehicleIndexUnlocked(company, oid)
		if !(showUnlocked && unlocked || showLocked && !unlocked) {
			continue
		}
		if o.Search != "" && !strings.Contains(v.DisplayName, o.Search) {
			continue
		}
		if trackType != 0xFF {
			want := trackType
			if trackType&(1<<7) != 0 {
				if v.Mode != objects.TransportModeRoad {
					continue
				}
				if trackType == 0xFE {
					want = 0xFF
				} else {
					want = trackType &^ (1 << 7)
				}
			} else if v.Mode != objects.TransportModeRail {
				continue
			}
			if want != om.TrackTypeFor(v) {
				continue
			}
		}
		cargo := om.CargoFor(v)
		if o.CargoFilter != 0xFF && o.CargoFilter != 0xFE && (cargo.CompatibleCargoCategories[0]|cargo.CompatibleCargoCategories[1])&(1<<o.CargoFilter) == 0 {
			continue
		}
		powered := v.Power > 0
		if !(powered && showPowered || !powered && showUnpowered) {
			continue
		}
		if o.CargoFilter == 0xFE && (cargo.CompatibleCargoCategories[0] != 0 || cargo.CompatibleCargoCategories[1] != 0) {
			continue
		}
		list = append(list, bv{oid, v.DisplayName, powered, v.Designed, v.Obsolete, m.InflationAdjustedCost(v.CostFactor, v.CostIndex, 10),
			v.Power, v.Speed, v.Weight, vehicleLength(v), cargo.MaxCargo[0]})
	}
	less := func(a, b bv) bool {
		switch o.SortBy {
		case BuildSortName:
			return a.name < b.name
		case BuildSortCost:
			return a.cost < b.cost
		case BuildSortPower:
			return a.power < b.power
		case BuildSortMaxSpeed:
			return a.speed < b.speed
		case BuildSortWeight:
			return a.weight < b.weight
		case BuildSortLength:
			return a.length < b.length
		case BuildSortCapacity:
			return a.capacity < b.capacity
		case BuildSortObsolete:
			return a.obsolete < b.obsolete
		default:
			return a.designed < b.designed
		}
	}
	// std::sort (not stable): ties keep goloco's slice order — PARTIAL: an
	// introsort may order equal keys differently from upstream's.
	sort.SliceStable(list, func(i, j int) bool {
		if o.Ascending {
			return less(list[i], list[j])
		}
		return less(list[j], list[i])
	})
	if o.Filter&(BuildFilterPowered|BuildFilterUnpowered) == BuildFilterPowered|BuildFilterUnpowered {
		sort.SliceStable(list, func(i, j int) bool { return list[i].powered && !list[j].powered })
	}
	out := make([]uint16, len(list))
	for i, b := range list {
		out[i] = b.id
	}
	return out
}
