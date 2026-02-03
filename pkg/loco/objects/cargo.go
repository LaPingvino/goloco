package objects

import "log"

// Cargo object definitions

// StringId represents a string table ID
type StringId = uint16

// CargoObjectFlags are flags for cargo objects
type CargoObjectFlags uint32

const (
	CargoFlagNone       CargoObjectFlags = 0
	CargoFlagUnk0       CargoObjectFlags = 1 << 0
	CargoFlagRefit      CargoObjectFlags = 1 << 1
	CargoFlagDelivering CargoObjectFlags = 1 << 2
)

// CargoCategory categorizes cargo types
type CargoCategory uint16

const (
	CargoCategoryGrain      CargoCategory = 1
	CargoCategoryCoal       CargoCategory = 2
	CargoCategoryIronOre    CargoCategory = 3
	CargoCategoryLiquids    CargoCategory = 4
	CargoCategoryPaper      CargoCategory = 5
	CargoCategorySteel      CargoCategory = 6
	CargoCategoryTimber     CargoCategory = 7
	CargoCategoryGoods      CargoCategory = 8
	CargoCategoryFoods      CargoCategory = 9
	CargoCategoryLivestock  CargoCategory = 11
	CargoCategoryPassengers CargoCategory = 12
	CargoCategoryMail       CargoCategory = 13
	CargoCategoryNull       CargoCategory = 0xFFFF
)

// CargoObject defines a cargo type
type CargoObject struct {
	Name                  StringId
	UnitWeight            uint16
	CargoTransferTime     uint16
	UnitsAndCargoName     StringId
	UnitNameSingular      StringId
	UnitNamePlural        StringId
	UnitInlineSprite      uint32
	Category              CargoCategory
	Flags                 CargoObjectFlags
	NumPlatformVariations uint8
	StationCargoDensity   uint8
	PremiumDays           uint8
	MaxNonPremiumDays     uint8
	NonPremiumRate        uint16
	PenaltyRate           uint16
	PaymentFactor         uint16
	PaymentIndex          uint8
	UnitSize              uint8
}

// HasFlags checks if the cargo has specific flags
func (co *CargoObject) HasFlags(flags CargoObjectFlags) bool {
	return (co.Flags & flags) != 0
}

// Validate validates the cargo object. Stub — always returns true.
//
// OpenLoco reference: src/OpenLoco/src/Objects/CargoObject.cpp
//
//	CargoObject::validate()  (at 0x0042F533)
//
// In OpenLoco, validate() checks that referenced string IDs exist,
// that the unit sprite is a valid image ID, and that the category
// is within the known enum range.
func (co *CargoObject) Validate() bool {
	log.Println("[Objects] CargoObject.Validate: stub — always returning true")
	return true
}

// Load loads cargo object data from bytes. Stub — does nothing.
//
// OpenLoco reference: src/OpenLoco/src/Objects/CargoObject.cpp
//
//	CargoObject::load(const LoadedObjectHandle&, std::span<const std::byte>, ...)
//	  (at 0x0042F4D0)
//
// In OpenLoco, load() deserialises the fixed-size CargoObject header from
// the object data span, then resolves dependent string and image IDs via
// the ObjectManager's DependentObjects handle.
func (co *CargoObject) Load(data []byte) error {
	log.Println("[Objects] CargoObject.Load: stub — not yet implemented")
	return nil
}

// Unload releases any resources held by the cargo object. Stub — does nothing.
//
// OpenLoco reference: src/OpenLoco/src/Objects/CargoObject.cpp
//
//	CargoObject::unload()  (at 0x0042F514)
//
// In OpenLoco, unload() clears resolved string and image pointers so the
// object can be safely removed from the ObjectManager.
func (co *CargoObject) Unload() {
	log.Println("[Objects] CargoObject.Unload: stub — not yet implemented")
}
