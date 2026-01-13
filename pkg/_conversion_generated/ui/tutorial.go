package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Tutorial.h"
// #include "Graphics/Colour.h"
// #include "Graphics/Gfx.h"
// #include "Graphics/TextRenderer.h"
// #include "Localisation/FormatArguments.hpp"
// #include "Localisation/StringIds.h"
// #include "Objects/InterfaceSkinObject.h"
// #include "Objects/ObjectManager.h"
// #include "OpenLoco.h"
// #include "Ui.h"
// #include "Ui/Widget.h"
// #include "Ui/Widgets/Wt3Widget.h"
// #include "Ui/WindowManager.h"
// namespace OpenLoco::Ui::Windows::Tutorial
type Widx int

const (
	Frame Widx = iota
)
// static constexpr Ui::Size kWindowSize = { 140, 29 };
// static constexpr auto widgets = makeWidgets(
// Widgets::Wt3Widget({ 0, 0 }, kWindowSize, WindowColour::primary)
// );
// static const WindowEventList& getEvents();
// // 0x00438CAE
// Window* open()
// auto window = WindowManager::createWindow(
// WindowType::tutorial,
// { kWindowSize.width, Ui::height() - 27 },
// { Ui::width() - 280, 27 },
// WindowFlags::stickToFront | WindowFlags::transparent | WindowFlags::noBackground,
// getEvents());
// window->setWidgets(widgets);
// window->initScrollWidgets();
// auto skin = ObjectManager::get<InterfaceSkinObject>();
// if (skin != nullptr)
// window->setColour(WindowColour::primary, AdvancedColour(skin->mapTooltipObjectColour).translucent());
// window->setColour(WindowColour::secondary, AdvancedColour(skin->mapTooltipCargoColour).translucent());
// orphan member: return window;
// // 0x00439B3D
// func PrepareDraw(self Window) 
// self.widgets[Widx::frame].right = self.width - 1;
// // 0x00439B4A
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// static constexpr StringId titleStringIds[] = {
// StringIds::tutorial_1_title,
// StringIds::tutorial_2_title,
// StringIds::tutorial_3_title,
// auto tr = Gfx::TextRenderer(drawingCtx);
// auto tutorialNumber = OpenLoco::Tutorial::getTutorialNumber();
// orphan member: FormatArguments args{};
// args.push(titleStringIds[tutorialNumber]);
// auto& widget = self.widgets[Widx::frame];
// auto point = Point(self.x + widget.midX(), self.y + widget.top + 4);
// tr.drawStringCentred(point, Colour::black, StringIds::tutorial_text, args);
// point.y += 10;
// tr.drawStringCentred(point, Colour::black, StringIds::tutorial_control);
// static constexpr WindowEventList kEvents = {
// .prepareDraw = prepareDraw,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
