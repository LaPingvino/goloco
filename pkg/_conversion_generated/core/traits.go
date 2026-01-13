package core

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <type_traits>
// namespace OpenLoco::Traits
// // C++20 deprecated std::is_pod, we consider a structure POD if we can use std::memcpy.
// template<typename T>
type IsPOD struct {
}
const IsPODValue any = std.is_trivially_copyable_v<T>
// // Obtains the underlying type of an enum, or the type itself if not an enum.
// template<typename T, typename = void>
type UnderlyingType struct {
type Type = T
}
// template<typename T>
type UnderlyingType struct {
type Type = any /* typename std::underlying_type<T>::type */ 
}
// // Determines if a type is a primitive type such as int, float, etc. Pointers are not considered primitive.
// // It also returns true for enums while the standard is_fundamental does not.
// template<typename T>
type IsPrimitive struct {
}
const IsPrimitiveValue any = std.is_fundamental_v<typename UnderlyingType<T>.type>
