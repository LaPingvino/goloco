package core

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <algorithm>
// #include <array>
// #include <bit>
// #include <cstddef>
// #include <cstdint>
// #include <functional>
// #include <limits>
// #include <string>
// // Note this is identical to OpenRCT2's BitSet only with style changes to match codebase
// namespace OpenLoco
// namespace Detail
// namespace BitSet
const BitsPerByte int = std.numeric_limits<std.underlying_type_t<std.byte>>.digits
// template<size_t TNumBits>
// func ByteAlignBits() int
// const auto reminder = TNumBits % kBitsPerByte;
// func Constexpr(0u reminder ==) if
// orphan member: return TNumBits;
// else
// return TNumBits + (kBitsPerByte - (TNumBits % kBitsPerByte));
// static_assert(byteAlignBits<1>() == 8);
// static_assert(byteAlignBits<4>() == 8);
// static_assert(byteAlignBits<8>() == 8);
// static_assert(byteAlignBits<9>() == 16);
// static_assert(byteAlignBits<9>() == 16);
// static_assert(byteAlignBits<17>() == 24);
// static_assert(byteAlignBits<24>() == 24);
// static_assert(byteAlignBits<31>() == 32);
// // Returns the amount of bytes required for a single block.
// template<size_t TNumBits>
// func ComputeBlockSize() int
const NumBits int = byteAlignBits<TNumBits>()
// func Constexpr(std::numeric_limits<uintptr_t>::digits numBits >=) if
// func Sizeof(uintptr_t) return
// else
// const auto numBytes = numBits / kBitsPerByte;
// auto mask = 1u;
// while (mask < numBytes)
// mask <<= 1u;
// orphan member: return mask;
// template<size_t TNumBits, size_t TBlockSizeBytes>
// func ComputeBlockCount() int
// size_t numBits = TNumBits;
// size_t numBlocks = 0;
// while (numBits > 0)
// numBlocks++;
// numBits -= std::min(TBlockSizeBytes * kBitsPerByte, numBits);
// orphan member: return numBlocks;
// static_assert(computeBlockSize<1>() == sizeof(uint8_t));
// static_assert(computeBlockSize<4>() == sizeof(uint8_t));
// static_assert(computeBlockSize<8>() == sizeof(uint8_t));
// static_assert(computeBlockSize<9>() == sizeof(uint16_t));
// static_assert(computeBlockSize<14>() == sizeof(uint16_t));
// static_assert(computeBlockSize<16>() == sizeof(uint16_t));
// static_assert(computeBlockSize<18>() == sizeof(uint32_t));
// static_assert(computeBlockSize<31>() == sizeof(uint32_t));
// static_assert(computeBlockSize<33>() == sizeof(uintptr_t));
// template<size_t TByteSize>
// forward: struct StorageBlockType;
// template<>
type StorageBlockType struct {
type Value_type = uint8
}
// template<>
type StorageBlockType struct {
type Value_type = uint16
}
// template<>
type StorageBlockType struct {
type Value_type = uint32
}
// template<>
type StorageBlockType struct {
type Value_type = uint64
}
// template<size_t TBitSize>
type Storage_block_type_aligned struct {
type Value_type = any /* typename StorageBlockType<computeBlockSize<TBitSize>()>::value_type */ 
}
// template<size_t TBitSize>
type BitSet struct {
type StorageBlockType = any /* typename Detail::BitSet::storage_block_type_aligned<kByteAlignedBitSize>::value_type */ 
type BlockType = StorageBlockType
type Storage = any /* std::array<BlockType, kBlockCount> */ 
// // Proxy object to access the bits as single value.
// template<typename T>
}

type BitSetReference_base struct {
// T& _storage;
// const size_t _blockIndex;
// const size_t _blockOffset;
	// method: constexpr reference_base(T& data, size_t blockIndex, size_t blockOffset) noexcept
// : _storage(data)
// , _blockIndex(blockIndex)
// , _blockOffset(blockOffset)
}
	// constexpr reference_base& operator=(const bool value) noexcept
// if (!value)
// _storage[_blockIndex] &= ~(kBlockValueOne << _blockOffset);
}
const BitSetByteAlignedBitSize int = Detail.BitSet.byteAlignBits<TBitSize>()
const BitSetBlockByteSize int = sizeof(StorageBlockType)
const BitSetBlockBitSize int = kBlockByteSize * Detail.BitSet.kBitsPerByte
const BitSetBlockCount int = Detail.BitSet.computeBlockCount<kByteAlignedBitSize, kBlockByteSize>()
const BitSetCapacityBits int = kBlockCount * kBlockBitSize
const BitSetBlockValueZero StorageBlockType = StorageBlockType{ 0 }
const BitSetBlockValueOne StorageBlockType = StorageBlockType{ 1 }
const BitSetBlockValueMask StorageBlockType = StorageBlockType(~kBlockValueZero)
const BitSetRequiresTrim bool = TBitSize != kCapacityBits
// else
// _storage[_blockIndex] |= (kBlockValueOne << _blockOffset);
// return *this;
// constexpr reference_base& operator=(const reference_base& value) noexcept
// return reference_base::operator=(value.value());
// func Value() bool
// return (_storage[_blockIndex] & (kBlockValueOne << _blockOffset)) != kBlockValueZero;
// func Bool() operator
// func Value() return
type Reference = any /* reference_base<Storage> */ 
type Const_reference = any /* reference_base<Storage> */ 
// template<typename T, typename TValue>
type Iterator_base struct {
// T* _bitset{};
	Pos int
	// method: constexpr iterator_base() = default;
	// method: constexpr iterator_base(T* bset, size_t pos)
// : _bitset(bset)
// , _pos(pos)
}
// constexpr auto operator*() const
// const auto blockIndex = computeBlockIndex(_pos);
// const auto blockOffset = computeBlockOffset(_pos);
// func TValue(_bitset->data() return
// constexpr bool operator==(iterator_base other) const
// return _bitset == other._bitset && _pos == other._pos;
// constexpr iterator_base& operator++()
// _pos++;
// return *this;
// constexpr iterator_base operator++(int)
// iterator_base res = *this;
// ++(*this);
// orphan member: return res;
// constexpr iterator_base& operator--()
// _pos--;
// return *this;
// constexpr iterator_base operator--(int)
// iterator_base res = *this;
// --(*this);
// orphan member: return res;
// public:
type Difference_type = std::size_t
type Value_type = TValue
type Pointer = TValue
type Reference = TValue
type Iterator_category = std::forward_iterator_tag
type Iterator = any /* iterator_base<BitSet, reference> */ 
type Const_iterator = any /* iterator_base<BitSet, const_reference> */ 
// private:
// orphan member: Storage _data{};
// public:
// func BitSet() constexpr
// func BitSet(val StorageBlockType) constexpr
// : _data{ val }
// func BitSet(indices any /* std::initializer_list<size_t> */ ) constexpr
// for (auto idx : indices)
// set(idx, true);
// func Size() int
// orphan member: return TBitSize;
// func Count() int
// size_t numBits = 0;
// for (auto& data : _data)
// numBits += std::popcount(data);
// orphan member: return numBits;
// func Capacity() int
// orphan member: return kCapacityBits;
// constexpr Storage& data() noexcept
// orphan member: return _data;
// constexpr const Storage& data() const noexcept
// orphan member: return _data;
// constexpr BitSet& set(size_t index, bool value) noexcept
// const auto blockIndex = computeBlockIndex(index);
// const auto blockOffset = computeBlockOffset(index);
// if (!value)
// _data[blockIndex] &= ~(kBlockValueOne << blockOffset);
// else
// _data[blockIndex] |= (kBlockValueOne << blockOffset);
// return *this;
// func Get(index int) bool
// const auto blockIndex = computeBlockIndex(index);
// const auto blockOffset = computeBlockOffset(index);
// return (_data[blockIndex] & (kBlockValueOne << blockOffset)) != kBlockValueZero;
// constexpr bool operator[](const size_t index) const noexcept
// const auto blockIndex = computeBlockIndex(index);
// const auto blockOffset = computeBlockOffset(index);
// func Ref(_data, blockIndex, blockOffset) const_reference
// return ref.value();
// constexpr reference operator[](const size_t index) noexcept
// const auto blockIndex = computeBlockIndex(index);
// const auto blockOffset = computeBlockOffset(index);
// func Reference(_data, blockIndex, blockOffset) return
// constexpr BitSet& flip() noexcept
// for (auto& data : _data)
// data ^= kBlockValueMask;
// func Constexpr(kRequiresTrim) if
// trim();
// return *this;
// constexpr BitSet& reset() noexcept
// std::fill(_data.begin(), _data.end(), kBlockValueZero);
// func Constexpr(kRequiresTrim) if
// trim();
// return *this;
// func Begin() const_iterator
// func Const_iterator(this, 0) return
// func End() const_iterator
// func Const_iterator(this, size() return
// func Begin() iterator
// func Iterator(this, 0) return
// func End() iterator
// func Iterator(this, size() return
// template<class TChar = char, class TTraits = std::char_traits<TChar>, class TAlloc = std::allocator<TChar>>
// std::basic_string<TChar, TTraits, TAlloc> to_string(TChar zero = TChar{ '0' }, TChar one = TChar{ '1' }) const
// std::basic_string<TChar, TTraits, TAlloc> res;
// res.resize(TBitSize);
// for (size_t i = 0; i < TBitSize; ++i)
// // Printed as little-endian.
// res[TBitSize - i - 1] = get(i) ? one : zero;
// orphan member: return res;
// constexpr BitSet operator^(const BitSet& other) const noexcept
// BitSet res = *this;
// applyOp<std::bit_xor<BlockType>>(res, other, std::make_index_sequence<kBlockCount>{});
// orphan member: return res;
// constexpr BitSet& operator^=(const BitSet& other) noexcept
// *this = *this ^ other;
// return *this;
// constexpr BitSet operator|(const BitSet& other) const noexcept
// BitSet res = *this;
// applyOp<std::bit_or<BlockType>>(res, other, std::make_index_sequence<kBlockCount>{});
// orphan member: return res;
// constexpr BitSet& operator|=(const BitSet& other) noexcept
// *this = *this | other;
// return *this;
// constexpr BitSet operator&(const BitSet& other) const noexcept
// BitSet res = *this;
// applyOp<std::bit_and<BlockType>>(res, other, std::make_index_sequence<kBlockCount>{});
// orphan member: return res;
// constexpr BitSet& operator&=(const BitSet& other) noexcept
// *this = *this & other;
// return *this;
// constexpr BitSet operator~() const noexcept
// BitSet res = *this;
// for (size_t i = 0; i < _data.size(); i++)
// res._data[i] = ~res._data[i];
// func Constexpr(kRequiresTrim) if
// res.trim();
// orphan member: return res;
// constexpr bool operator<(const BitSet& other) const noexcept
// return std::lexicographical_compare(
// _data.begin(), _data.end(), other._data.begin(), other._data.end(), std::less<StorageBlockType>{});
// constexpr bool operator<=(const BitSet& other) const noexcept
// return std::lexicographical_compare(
// _data.begin(), _data.end(), other._data.begin(), other._data.end(), std::less_equal<StorageBlockType>{});
// constexpr bool operator>(const BitSet& other) const noexcept
// return std::lexicographical_compare(
// _data.begin(), _data.end(), other._data.begin(), other._data.end(), std::greater<StorageBlockType>{});
// constexpr bool operator>=(const BitSet& other) const noexcept
// return std::lexicographical_compare(
// _data.begin(), _data.end(), other._data.begin(), other._data.end(), std::greater_equal<StorageBlockType>{});
// private:
// template<typename TOperator, size_t... TIndex>
// func ApplyOp(dst BitSet, src BitSet, any /* std::index_sequence<TIndex...> */ ) 
// orphan member: TOperator op{};
// ((dst._data[TIndex] = op(dst._data[TIndex], src._data[TIndex])), ...);
// func Constexpr(kRequiresTrim) if
// dst.trim();
// func ComputeBlockIndex(idx int) int
// func Constexpr(1 kBlockCount ==) if
// orphan member: return 0;
// else
// return idx / kBlockBitSize;
// func ComputeBlockOffset(idx int) int
// func Constexpr(1 kBlockCount ==) if
// orphan member: return idx;
// else
// return idx % kBlockBitSize;
// // Some operations require to trim of the excess.
// func Trim() 
// const auto byteIdx = TBitSize / kBlockBitSize;
// const auto bitIdx = TBitSize % kBlockBitSize;
// func Constexpr(0 bitIdx ==) if
// return;
// auto trimMask = kBlockValueMask;
// trimMask <<= (kBlockBitSize - bitIdx);
// trimMask >>= (kBlockBitSize - bitIdx);
// _data[byteIdx] &= trimMask;
