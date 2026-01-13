package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Window.h"
// namespace OpenLoco::Ui::ScrollView
// // For horizontal scrollbars its N wide, for vertical its N tall
const ScrollbarSize uint8 = 11
var ScrollButtonSize = Ui.Size(kScrollbarSize, kScrollbarSize) // auto
var ScrollbarMargin = 1 // auto
const ThumbSize uint8 = kScrollbarSize - kScrollbarMargin
const MinThumbSize uint8 = kThumbSize * 2
const ButtonClickStep uint8 = 3
type GetPartResult struct {
// Ui::Point scrollviewLoc;
	Area ScrollPart
	Index int
}
// func GetPart(window Ui::Window, widget Ui::Widget, x int16, y int16) GetPartResult
// func UpdateThumbs(window Window, widgetIndex WidgetIndex_t) 
// func ScrollLeftBegin(x int16, y int16, w Ui::Window, widget Ui::Widget, widgetIndex WidgetIndex_t) 
// func ScrollLeftContinue(x int16, y int16, w Ui::Window, widget Ui::Widget, widgetIndex WidgetIndex_t) 
// func ScrollModalRight(x int16, y int16, w Ui::Window, widget Ui::Widget, widgetIndex WidgetIndex_t) 
// func ClearPressedButtons(type WindowType, number WindowNumber_t, widgetIndex WidgetIndex_t) 
// func HorizontalDragFollow(w Ui::Window, widget Ui::Widget, dragWidgetIndex WidgetIndex_t, dragScrollIndex int, deltaX int16) 
// func VerticalDragFollow(w Ui::Window, widget Ui::Widget, dragWidgetIndex WidgetIndex_t, dragScrollIndex int, deltaY int16) 
// func VerticalNudgeUp(w Ui::Window, scrollAreaIndex int, widgetIndex WidgetIndex_t) 
// func VerticalNudgeDown(w Ui::Window, scrollAreaIndex int, widgetIndex WidgetIndex_t) 
