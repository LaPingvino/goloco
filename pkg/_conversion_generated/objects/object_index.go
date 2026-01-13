package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "ObjectManager.h" // TODO: Split off entry def to different header
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <array>
// #include <optional>
// #include <span>
// #include <string>
// #include <vector>
// namespace OpenLoco::ObjectManager
// forward: struct ObjectHeader3;
type SelectedObjectsFlags int

const (
	None SelectedObjectsFlags = 0
	Selected SelectedObjectsFlags = 1 << 0
	InUse SelectedObjectsFlags = 1 << 2
	RequiredByAnother SelectedObjectsFlags = 1 << 3
	AlwaysRequired SelectedObjectsFlags = 1 << 4
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(SelectedObjectsFlags);
type ObjectIndexEntry struct {
	Header ObjectHeader
	Header2 ObjectHeader2
	DisplayData ObjectHeader3
// std::string _filepath; // u8string
// std::string _name;
// std::vector<ObjectHeader> _requiredObjects;
// std::vector<ObjectHeader> _alsoLoadObjects;
}
// // Index into the overall ObjectIndex. Note: Not a type specific index!
type ObjectIndexId = int16
const NullObjectIndex ObjectIndexId = -1
type ObjIndexPair struct {
	Index ObjectIndexId
	Object ObjectIndexEntry
}
// func GetCustomObjectsInIndexStatus() bool
// func GetNumInstalledObjects() uint32
// func LoadIndex() 
// std::vector<ObjIndexPair> getAvailableObjects(ObjectType type);
// func GetNumAvailableObjectsByType(type ObjectType) uint16
// func IsObjectInstalled(objectHeader ObjectHeader) bool
// std::optional<ObjectIndexEntry> findObjectInIndex(const ObjectHeader& objectHeader);
// const ObjectIndexEntry& getObjectInIndex(ObjectIndexId index);
// func GetActiveObject(objectType ObjectType, objectIndexFlags any /* std::span<SelectedObjectsFlags> */ ) ObjIndexPair
type ObjectSelectionMeta struct {
// std::array<uint16_t, kMaxObjectTypes> numSelectedObjects;
	NumImages uint32
}
// static_assert(sizeof(ObjectSelectionMeta) == 0x48);
type SelectObjectModes int

const (
	None SelectObjectModes = 0
	Select SelectObjectModes = 1 << 0
	SelectDependents SelectObjectModes = 1 << 1
	SelectAlsoLoads SelectObjectModes = 1 << 2
	MarkAsAlwaysRequired SelectObjectModes = 1 << 3
	DefaultDeselect SelectObjectModes = selectDependents | selectAlsoLoads
	DefaultSelect SelectObjectModes = defaultDeselect | select
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(SelectObjectModes);
type ObjectIndexSelection struct {
	SelectionMetaData ObjectSelectionMeta
// std::vector<SelectedObjectsFlags> objectFlags;
	// method: bool selectObject(SelectObjectModes mode, const ObjectHeader& objHeader);
}
// ObjectIndexSelection& prepareSelectionList(bool markInUse);
// ObjectIndexSelection& getCurrentSelectionList();
// func FreeSelectionList() 
// func MarkOnlyLoadedObjects(objectFlags any /* std::span<SelectedObjectsFlags> */ ) 
// func LoadSelectionListObjects(objectFlags any /* std::span<SelectedObjectsFlags> */ ) 
// func UnloadUnselectedSelectionListObjects(objectFlags any /* std::span<SelectedObjectsFlags> */ ) 
// std::optional<ObjectType> validateObjectSelection(std::span<SelectedObjectsFlags> objectFlags);
