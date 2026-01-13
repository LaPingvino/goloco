package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/Gfx.h"
// #include "Ui/Widgets/ImageButtonWidget.h"
// #include "Ui/Widgets/NewsPanelWidget.h"
// #include "Ui/Widgets/ViewportWidget.h"
// #include "Ui/WindowManager.h"
// #include "World/Company.h"
// namespace OpenLoco::Ui::Windows::NewsWindow
type NewsState struct {
// SavedView savedView[2];
	SlideInHeight int32
	NumCharsToDisplay uint16
}
type SubjectType int

const (
	CompanyFace SubjectType = -2
	VehicleImage SubjectType = -3
)
// extern NewsState _nState;
// namespace Common
type Widx int

const (
	Frame Widx = iota
	Close_button
	Viewport1
	Viewport2
	Viewport1Button
	Viewport2Button
)
// template<typename TFrameWidget>
// func MakeCommonWidgets(frameWidth int32, frameHeight int32) any
// return makeWidgets(
// TFrameWidget({ 0, 0 }, { frameWidth, frameHeight }, WindowColour::primary),
// Widgets::ImageButton({ frameWidth - 15, 2 }, { 13, 13 }, WindowColour::primary, ImageIds::close_button, StringIds::tooltip_close_window),
// Widgets::Viewport({ 2, frameHeight - 73 }, { 168, 64 }, WindowColour::primary, Widget::kContentUnk),
// Widgets::Viewport({ 180, frameHeight - 73 }, { 168, 64 }, WindowColour::primary, Widget::kContentUnk),
// Widgets::ImageButton({ 2, frameHeight - 75 }, { 180, 75 }, WindowColour::primary),
// Widgets::ImageButton({ 2, frameHeight - 75 }, { 180, 75 }, WindowColour::primary));
// const WindowEventList& getEvents();
// func InitViewports(self Window) 
// namespace News1
// static constexpr Ui::Size kWindowSize = { 360, 117 };
// std::span<const Widget> getWidgets();
// namespace News2
// static constexpr Ui::Size kWindowSize = { 360, 159 };
// std::span<const Widget> getWidgets();
// namespace Ticker
// static constexpr Ui::Size kWindowSize = { 111, 26 };
type Widx int

const (
	Frame Widx = iota
)
// std::span<const Widget> getWidgets();
// const WindowEventList& getEvents();
