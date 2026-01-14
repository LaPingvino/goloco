package objects

// Cargo object definitions

// StringId represents a string table ID
type StringId = uint16

// CargoObjectFlags are flags for cargo objects
type CargoObjectFlags uint32

const (
	CargoFlagNone      CargoObjectFlags = 0
	CargoFlagUnk0      CargoObjectFlags = 1 << 0
	CargoFlagRefit     CargoObjectFlags = 1 << 1
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

// Validate validates the cargo object
// TODO: Implement validation logic
func (co *CargoObject) Validate() bool {
	return true
}

// Load loads cargo object data from bytes
// TODO: Implement binary loading
func (co *CargoObject) Load(data []byte) error {
	return nil
}

// Unload unloads the cargo object
// TODO: Implement cleanup
func (co *CargoObject) Unload() {
}
