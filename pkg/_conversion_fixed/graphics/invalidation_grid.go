package graphics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <algorithm>
// #include <cstdint>
// #include <vector>
// namespace OpenLoco::Gfx
type InvalidationGrid struct {
	BlockWidth  uint16
	BlockHeight uint16
	ColumnCount uint32
	RowCount    uint32
	// std::vector<uint8> _blocks;
	ScreenWidth  uint32
	ScreenHeight uint32
	// method: uint32 getRowCount() const noexcept;
	// method: uint32 getColumnCount() const noexcept;
	// method: uint32 getBlockWidth() const noexcept;
	// method: uint32 getBlockHeight() const noexcept;
	// method: void reset(int32 width, int32 height, uint32 blockWidth, uint32 blockHeight) noexcept;
	// method: void invalidate(int32 left, int32 top, int32 right, int32 bottom) noexcept;
	// template<typename F>
	// method: void traverseDirtyCells(F&& func)
	// const auto columnCount = _columnCount;
	// const auto rowCount = _rowCount;
	// const auto blockWidth = _blockWidth;
	// const auto blockHeight = _blockHeight;
	// auto& blocks = _blocks;
	// for (uint32 column = 0; column < columnCount; column++)
	// for (uint32 row = 0; row < rowCount; row++)
	// const auto rowStartOffset = row * columnCount;
	// if (blocks[rowStartOffset + column] != 0)
	// uint32 rowEndOffset = rowStartOffset;
	// uint32 numRowsDirty = 0;
	// // Count amount of dirty rows at current column.
	// while (true)
	// if (row + numRowsDirty + 1 >= rowCount || blocks[rowEndOffset + column + columnCount] == 0)
	// break;
}

// numRowsDirty++;
// rowEndOffset += columnCount;
// // Clear rows at the current column.
// for (auto rowOffset = rowStartOffset; rowOffset <= rowEndOffset; rowOffset += columnCount)
// blocks[rowOffset + column] = 0;
// // Convert to pixel coordinates.
// const auto left = column * blockWidth;
// const auto top = row * blockHeight;
// const auto right = (column + 1) * blockWidth;
// const auto bottom = (row + numRowsDirty + 1) * blockHeight;
// if (left < _screenWidth && top < _screenHeight)
// func(left, top, std::min(right, _screenWidth), std::min(bottom, _screenHeight));
