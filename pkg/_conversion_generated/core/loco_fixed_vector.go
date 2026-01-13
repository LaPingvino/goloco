package core

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <iterator>
// namespace OpenLoco
// template<typename ValueType, size_t Count>
type FixedVector struct {
// ValueType* startAddress = nullptr;
}

type FixedVectorIter struct {
// ValueType* arr;
// size_t i = 0;
	// method: constexpr void findNonEmpty()
// for (; i < Count; ++i)
// if (!arr[i].empty())
// break;
}
}
// public:
// func Iter(_arr ValueType, _index int) constexpr
// : arr(_arr)
// , i(_index)
// // finds first valid entry
// findNonEmpty();
// constexpr Iter& operator++()
// ++i;
// findNonEmpty();
// return *this;
// constexpr Iter operator++(int)
// Iter retval = *this;
// ++(*this);
// orphan member: return retval;
// constexpr bool operator==(Iter other) const
// return i == other.i;
// constexpr ValueType& operator*() const
// return arr[i];
// // iterator traits
type Difference_type = std::ptrdiff_t
type Value_type = ValueType
type Pointer = ValueType
type Reference = ValueType
type Iterator_category = std::forward_iterator_tag
// public:
// FixedVector(ValueType (&_arr)[Count])
// : startAddress(_arr)
// func Begin() Iter
// func Iter(startAddress, 0) return
// func End() Iter
// func Iter(startAddress, Count) return
// [[nodiscard]] bool empty() const
// func Begin() return
// [[nodiscard]] size_t size() const
// return std::distance(begin(), end());
// [[nodiscard]] constexpr size_t capacity() const
// orphan member: return Count;
