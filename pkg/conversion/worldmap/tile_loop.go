package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Tile.h"
// namespace OpenLoco::World
type TileLoop struct {
	Pos Pos2
// TileLoop() = default;
	// method: explicit TileLoop(const Pos2& startPos)
// : _pos(startPos)
}
func Current() Pos2 {
	// Pos2 next()
	// _pos.x += kTileSize;
	if _pos.x >= kMapWidth - 1 {
	// _pos.x = 0;
	// _pos.y += kTileSize;
	if _pos.y >= kMapHeight - 1 {
	// _pos.y = 0;
	}
	}
	return _pos
	}
	// };
	// // Loops over a range from bottomLeft to topRight inclusive
type TilePosRangeView struct {
	// private:
	var _bottomLeft TilePos2
	var _topRight TilePos2
}

type TilePosRangeViewIterator struct {
	// private:
	var _bottomLeft TilePos2
	var _topRight TilePos2
	var _pos TilePos2
	// public:
	// Iterator(const TilePos2& bottomLeft, const TilePos2& topRight)
	// : _bottomLeft(bottomLeft)
	// , _topRight(topRight)
	// , _pos(bottomLeft)
	}
	// Iterator& operator++()
	if _pos.x >= _topRight.x {
	// _pos.x = _bottomLeft.x;
	// _pos.y++;
	}
	} else {
	// _pos.x++;
	}
	return *this
	}
	// Iterator operator++(int)
	var retval Iterator = *this
	// ++(*this);
	return retval
	}
	// bool operator==(const Iterator& other) const
	return _pos == other._pos
	}
	// const TilePos2& operator*()
	return _pos
	}
	// // iterator traits
type Difference_type = std::ptrdiff_t
type Value_type = TilePos2
type Pointer = TilePos2
type Reference = TilePos2
type Iterator_category = std::forward_iterator_tag
	// };
	// public:
	// TilePosRangeView(const TilePos2& bottomLeft, const TilePos2& topRight)
	// : _bottomLeft(bottomLeft)
	// , _topRight(topRight)
	assert(bottomLeft.x <= topRight.x)
	assert(bottomLeft.y <= topRight.y)
	}
	// Iterator begin() const { return Iterator(_bottomLeft, _topRight); }
	// Iterator end() const
	// // End iterator must be 1 step past the end so that loop is inclusive
	return Iterator(TilePos2(_bottomLeft.x, _topRight.y + 1), _topRight)
	}
	// };
	// TilePosRangeView getClampedRange(const TilePos2& posA, const TilePos2& posB);
	// TilePosRangeView getClampedRange(const Pos2& posA, const Pos2& posB);
	// TilePosRangeView getDrawableTileRange();
	// TilePosRangeView getWorldRange();
	}
