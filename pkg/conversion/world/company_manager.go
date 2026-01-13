package world

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Company.h"
// #include "Engine/Limits.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/LocoFixedVector.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <array>
// #include <cstddef>
// namespace OpenLoco::CompanyManager
// func Reset() 
// func GetMaxCompetingCompanies() uint8
// func SetMaxCompetingCompanies(competingCompanies uint8) 
// func GetCompetitorStartDelay() uint8
// func SetCompetitorStartDelay(competetorStartDelay uint8) 
// func GetMaxLoanSize() uint16
// func SetMaxLoanSize(loanSize uint16) 
// func GetStartingLoanSize() uint16
// func GetInflationAdjustedStartingLoan() currency32_t
// func SetStartingLoanSize(loanSize uint16) 
// func Companies() any /* FixedVector<Company, Limits::kMaxCompanies> */ 
// Company* get(CompanyId id);
// func GetControllingId() CompanyId
// func GetSecondaryPlayerId() CompanyId
// func SetControllingId(id CompanyId) 
// func SetSecondaryPlayerId(id CompanyId) 
// Company* getPlayerCompany();
// func GetCompanyColour(id CompanyId) Colour
// func GetPlayerCompanyColour() Colour
// func IsPlayerCompany(id CompanyId) bool
// func Update() 
// func UpdateDaily() 
// func UpdateMonthly1() 
// func UpdateMonthlyHeadquarters() 
// func UpdateQuarterly() 
// func UpdateYearly() 
// func DetermineAvailableVehicles() 
// func CalculateDeliveredCargoPayment(cargoItem uint8, numUnits int32, distance int32, numDays uint16) currency32_t
// Company* getOpponent();
// func GetOwnerStatus(id CompanyId, args FormatArguments) StringId
// func UpdateOwnerStatus() 
// func UpdateColours() 
// func SetPreferredName() 
// func SpendMoneyEffect(loc World::Pos3, company CompanyId, amount currency32_t) 
// func ApplyPaymentToCompany(id CompanyId, payment currency32_t, type ExpenditureType) 
// func EnsureCompanyFunding(id CompanyId, payment currency32_t) bool
// func CompetingColourMask(companyId CompanyId) uint32
// func CompetingColourMask() uint32
// func CreatePlayerCompany() 
// func GetHeadquarterBuildingType() uint8
// // Vector of competitor object index's that are in use that aren't @id's competitor object index.
// std::vector<uint32_t> findAllOtherInUseCompetitors(const CompanyId id);
// func AiDestroy(id CompanyId) 
// func UpdatePlayerInfrastructureOptions() 
