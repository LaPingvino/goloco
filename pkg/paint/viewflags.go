package paint

import "github.com/LaPingvino/goloco/pkg/gfx8"

// ViewFlags is Ui::ViewportFlags (Viewport.hpp): what a viewport shows — the
// View menu's entries (ToolbarTop.cpp viewMenuDropdown), the gridlines, and
// the names a viewport hides. A paint session carries its viewport's
// (SessionOptions::viewFlags). flag_7 and flag_8 (bits 7 and 8) are never set
// upstream and are left out.
type ViewFlags uint16

const (
	ViewUnderground           ViewFlags = 1 << 0 // underground_view
	ViewSeeThroughTracks      ViewFlags = 1 << 1
	ViewHeightMarksOnLand     ViewFlags = 1 << 2
	ViewHeightMarksOnTracks   ViewFlags = 1 << 3 // height_marks_on_tracks_roads
	ViewOneWayDirectionArrows ViewFlags = 1 << 4
	ViewGridlines             ViewFlags = 1 << 5 // gridlines_on_landscape
	ViewSeeThroughScenery     ViewFlags = 1 << 6
	ViewHideTownNames         ViewFlags = 1 << 9
	ViewHideStationNames      ViewFlags = 1 << 10
	ViewSeeThroughRoads       ViewFlags = 1 << 11
	ViewSeeThroughBuildings   ViewFlags = 1 << 12
	ViewSeeThroughTrees       ViewFlags = 1 << 13
	ViewSeeThroughBridges     ViewFlags = 1 << 14

	// ViewSeeThroughAny: any of the six see-through flags (Window.cpp
	// viewportMove repaints such a view instead of shifting its pixels).
	ViewSeeThroughAny = ViewSeeThroughTracks | ViewSeeThroughScenery | ViewSeeThroughRoads | ViewSeeThroughBuildings | ViewSeeThroughTrees | ViewSeeThroughBridges
)

// Has reports whether any of flags is set (Viewport::hasFlags).
func (f ViewFlags) Has(flags ViewFlags) bool { return f&flags != 0 }

// ShouldTryCull is Paint.cpp's shouldTryCullPaintStruct: a struct of type t
// is one the view's see-through flags make see-through (isTypeCullableTrack /
// Road / Tree / Scenery / Building / Bridge).
func ShouldTryCull(t InteractionItem, f ViewFlags) bool {
	switch t {
	case ItemTrack, ItemTrackExtra, ItemSignal, ItemTrainStation:
		return f.Has(ViewSeeThroughTracks)
	case ItemRoadStation, ItemRoad, ItemRoadExtra:
		return f.Has(ViewSeeThroughRoads)
	case ItemIndustryTree, ItemTree:
		return f.Has(ViewSeeThroughTrees)
	case ItemWall:
		return f.Has(ViewSeeThroughScenery)
	case ItemBuilding, ItemIndustry, ItemHeadquarterBuilding, ItemAirport, ItemDock:
		return f.Has(ViewSeeThroughBuildings)
	case ItemBridge:
		return f.Has(ViewSeeThroughBridges)
	}
	return false
}

// culled is drawStructs' see-through step for an image of a struct of type
// t: draw false — cullPaintStructImage: nothing is drawn (the underground
// view, or an image that was blended already: shadows, glass); otherwise the
// image to draw — as it is, or, for a see-through struct, its bare sprite
// translucent in ExtColour::unk30 (ImageId(index).withTranslucency: the
// remap, the mask's image and the noise dropped — the mask itself is the
// struct's, kept).
func (s *Session) culled(t InteractionItem, img Image) (out Image, draw bool) {
	if !ShouldTryCull(t, s.ViewFlags) {
		return img, true
	}
	if s.ViewFlags.Has(ViewUnderground) || img.Op&gfx8.OpDst != 0 {
		return img, false
	}
	var pm []uint8
	if s.ColourMap != nil {
		pm = s.ColourMap(extColourSeeThrough)
	}
	return Image{Sprite: img.Sprite, Op: gfx8.OpTransparent | gfx8.OpDst, PalMap: pm, Mask: img.Mask}, true
}

// extColourSeeThrough is ExtColour::unk30 (translucent grey); colourBlue and
// colourMutedAvocadoGreen the Colours the decorations are drawn in.
const (
	extColourSeeThrough     = 0x30
	colourBlue              = 7
	colourMutedAvocadoGreen = 14
)

// HeightMarkerOffset is Config::heightMarkerOffset: the height markers' units
// (0 height units, 1 feet, 2 metres; default 1) — Options' distance / speed
// and heights dropdowns set it.
var HeightMarkerOffset = 1

// decoration adds a height marker or direction arrow: image in colour, as a
// parent with no interaction.
func (s *Session) decoration(image, colour int, off, bbOff, bbSize Pos3, sprite func(int) *gfx8.Sprite) {
	sp := sprite(image)
	if sp == nil {
		return
	}
	var pm []uint8
	if s.ColourMap != nil {
		pm = s.ColourMap(colour)
	}
	s.AddAsParent(Image{Sprite: sp, Op: gfx8.OpFor(sp, pm != nil, false), PalMap: pm}, off, bbOff, bbSize)
}

// PaintLandHeightMarker is paintSurface's height mark (PaintSurface.cpp,
// height_marks_on_land at full zoom): the marker for landHeight — the
// surface's height at the tile's middle (TileManager::getHeight) — 3 above
// it, in mutedAvocadoGreen, a parent of the surface's type.
func (s *Session) PaintLandHeightMarker(landHeight int, sprite func(int) *gfx8.Sprite) {
	if !s.ViewFlags.Has(ViewHeightMarksOnLand) || s.Zoom != 0 {
		return
	}
	h := landHeight + 3
	if img, ok := HeightMarkerImage(h, s.SeaLevel, HeightMarkerOffset); ok {
		off := Pos3{16, 16, h}
		s.decoration(img, colourMutedAvocadoGreen, off, off, Pos3{1, 1, 0}, sprite)
	}
}

// paintTrackRoadHeightMarker is paintTrack's / paintRoad's height mark
// (height_marks_on_tracks_roads at full zoom): on a piece's first and last
// tile, the marker for the element's height + its decoration offset
// (getTrackDecorationHeightOffset / getRoadDecorationHeightOffset: 16 on the
// high end of a slope) + 8, in blue, sorted in front of everything (bounding
// box at 1000, 1000, 1087).
func (s *Session) paintTrackRoadHeightMarker(height int, offsets [2]int8, first, last bool, sprite func(int) *gfx8.Sprite) {
	if !s.ViewFlags.Has(ViewHeightMarksOnTracks) || s.Zoom != 0 || !(first || last) {
		return
	}
	s.ItemType = ItemNoInteraction
	h := height + int(decorationOffset(offsets, first)) + 8
	if img, ok := HeightMarkerImage(h, s.SeaLevel, HeightMarkerOffset); ok {
		s.decoration(img, colourBlue, Pos3{16, 16, h}, Pos3{1000, 1000, 1087}, Pos3{1, 1, 0}, sprite)
	}
}

// decorationOffset is getTrack / getRoadDecorationHeightOffset's choice.
func decorationOffset(offsets [2]int8, first bool) int8 {
	if first {
		return offsets[0]
	}
	return offsets[1]
}

// HeightMarkerImage is PaintTileDecorations.cpp's getHeightMarkerImage: the
// marker for a world height — height / kMicroZStep − seaLevel + 128 into
// kHeightMarkerImages[Config::heightMarkerOffset] (0 units, 1 imperial, 2
// metric). ok false: outside the table (upstream reads past it).
func HeightMarkerImage(height, seaLevel, markerOffset int) (int, bool) {
	// C++'s division truncates toward zero, as Go's does.
	i := height/microZStep - seaLevel + 128
	if i < 0 || i >= 256 || markerOffset < 0 || markerOffset >= len(HeightMarkerImages) {
		return 0, false
	}
	return int(HeightMarkerImages[markerOffset][i]), true
}
