package entities

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Entity.h"
// #include <OpenLoco/Core/Exception.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <cstdio>
// #include <iterator>
// namespace OpenLoco::Vehicles
// forward: struct VehicleHead;
// namespace OpenLoco::EntityManager
type EntityListType int

const (
	Null EntityListType = iota
	NullMoney
	VehicleHead
	Misc EntityListType = 4
	Vehicle EntityListType = 6
)
// func Reset() 
// func FreeUserStrings() 
// template<typename T>
// T* get(EntityId id);
// template<>
// EntityBase* get(EntityId id);
// template<typename T>
// T* get(EntityId id)
// auto* base = get<EntityBase>(id);
// return base != nullptr ? base->asBase<T>() : nullptr;
// func FirstId(list EntityListType) EntityId
// func FirstQuadrantId(loc World::Pos2) EntityId
// func ResetSpatialIndex() 
// func UpdateSpatialIndex() 
// func MoveSpatialEntry(entity EntityBase, loc World::Pos3) 
// EntityBase* createEntityMisc();
// EntityBase* createEntityMoney();
// EntityBase* createEntityVehicle();
// func FreeEntity(entity EntityBase) 
// func GetListCount(list EntityListType) uint16
// func MoveEntityToList(entity EntityBase, list EntityListType) 
// func CheckNumFreeEntities(numNewEntities int) bool
// func ZeroUnused() 
// template<typename TEntityType, EntityId EntityBase::*nextList>
type ListIterator struct {
// TEntityType* entity = nullptr;
// EntityId nextEntityId = EntityId::null;
// ListIterator(const EntityId _headId)
// : nextEntityId(_headId)
// ++(*this);
}
// ListIterator& operator++()
// entity = get<TEntityType>(nextEntityId);
// if (entity)
// nextEntityId = entity->*nextList;
// return *this;
// ListIterator operator++(int)
// ListIterator retval = *this;
// ++(*this);
// orphan member: return retval;
// bool operator==(const ListIterator& other) const
// return entity == other.entity;
// TEntityType* operator*()
// if (entity == nullptr)
// throw Exception::RuntimeError("Bad Entity List!");
// orphan member: return entity;
// // iterator traits
type Difference_type = std::ptrdiff_t
type Value_type = TEntityType
type Pointer = TEntityType
type Reference = TEntityType
type Iterator_category = std::forward_iterator_tag
// template<typename T>
type EntityListIterator = any /* ListIterator<T, &EntityBase::nextEntityId> */ 
// template<typename T, EntityListType list>
type EntityList struct {
// EntityId firstId = EntityId::null;
// EntityList()
// firstId = EntityManager::firstId(list);
}
// func Begin() T
// func T(firstId) return
// func End() T
// func T(EntityId::null) return
type EntityTileList struct {
// EntityId firstId = EntityId::null;
type Iterator = any /* ListIterator<EntityBase, &EntityBase::nextQuadrantId> */ 
// EntityTileList(const World::Pos2& loc)
// firstId = EntityManager::firstQuadrantId(loc);
}
// func Begin() Iterator
// func Iterator(firstId) return
// func End() Iterator
// func Iterator(EntityId::null) return
