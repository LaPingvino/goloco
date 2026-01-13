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
	return _pos
	// };
	// // Loops over a range from bottomLeft to topRight inclusive
type TilePosRangeView struct {
	// private:
	_bottomLeft TilePos2
	_topRight TilePos2

type TilePosRangeViewIterator struct {
	// private:
	_pos TilePos2
	// public:
	// Iterator(const TilePos2& bottomLeft, const TilePos2& topRight)
	// : _bottomLeft(bottomLeft)
	// , _topRight(topRight)
	// , _pos(bottomLeft)
	// Iterator& operator++()
func init() {
	if _pos.x >= _topRight.x {

	// _pos.x = _bottomLeft.x;
	// _pos.y++;
	} else {
	// _pos.x++;
func init() {
	return *this

	// Iterator operator++(int)
	var retval Iterator = *this
	// ++(*this);
func init() {
	return retval

	// bool operator==(const Iterator& other) const
func init() {
	return _pos == other._pos

	// const TilePos2& operator*()
func init() {

	// // iterator traits
// SKIPPED C++ SYNTAX: type Difference_type = std::ptrdiff_t
type Value_type = TilePos2
type Pointer = TilePos2
type Reference = TilePos2
// SKIPPED C++ SYNTAX: type Iterator_category = std::forward_iterator_tag
	// };
	// public:
	// TilePosRangeView(const TilePos2& bottomLeft, const TilePos2& topRight)
	// : _bottomLeft(bottomLeft)
	// , _topRight(topRight)
func init() {
// SKIPPED CONSTRUCTOR: 	assert(bottomLeft.x <= topRight.x)
// SKIPPED CONSTRUCTOR: 	assert(bottomLeft.y <= topRight.y)

	// Iterator begin() const { return Iterator(_bottomLeft, _topRight); }
	// Iterator end() const
	// // End iterator must be 1 step past the end so that loop is inclusive
func init() {
// SKIPPED CONSTRUCTOR: 	return Iterator(TilePos2(_bottomLeft.x, _topRight.y + 1), _topRight)

	// };
	// TilePosRangeView getClampedRange(const TilePos2& posA, const TilePos2& posB);
	// TilePosRangeView getClampedRange(const Pos2& posA, const Pos2& posB);
	// TilePosRangeView getDrawableTileRange();
	// TilePosRangeView getWorldRange();