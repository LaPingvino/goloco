package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Date.h"
// #include "GameCommands/GameCommands.h"
// #include "GameCommands/General/SetGameSpeed.h"
// #include "GameCommands/General/TogglePause.h"
// #include "Graphics/Colour.h"
// #include "Graphics/Gfx.h"
// #include "Graphics/ImageIds.h"
// #include "Graphics/TextRenderer.h"
// #include "Input.h"
// #include "Intro.h"
// #include "Localisation/FormatArguments.hpp"
// #include "Localisation/StringIds.h"
// #include "Network/Network.h"
// #include "Objects/InterfaceSkinObject.h"
// #include "Objects/ObjectManager.h"
// #include "Scenario/Scenario.h"
// #include "Scenario/ScenarioObjective.h"
// #include "SceneManager.h"
// #include "Ui.h"
// #include "Ui/Dropdown.h"
// #include "Ui/ToolTip.h"
// #include "Ui/Widget.h"
// #include "Ui/Widgets/ImageButtonWidget.h"
// #include "Ui/Widgets/Wt3Widget.h"
// #include "Ui/WindowManager.h"
// #include "World/CompanyManager.h"
// namespace OpenLoco::Ui::Windows::TimePanel
// static constexpr Ui::Size kWindowSize = { 140, 27 };
// // When paused, the time panel will alternate between displaying "* paused *" and the in-game date every this many ticks.
var PausedStatusTextDuration = 30 // auto
// namespace Widx
// enum
// outer_frame,
// inner_frame,
// map_chat_menu,
// date_btn,
// pause_btn,
// normal_speed_btn,
// fast_forward_btn,
// extra_fast_forward_btn,
// func FormatChallenge(args FormatArguments) 
// func SendChatMessage(str byte) 
// static constexpr auto _widgets = makeWidgets(
// Widgets::Wt3Widget({ 0, 0 }, { 140, 29 }, WindowColour::primary),
// Widgets::Wt3Widget({ 2, 2 }, { 136, 25 }, WindowColour::primary),
// Widgets::ImageButton({ 113, 1 }, { 26, 26 }, WindowColour::primary),
// Widgets::ImageButton({ 2, 2 }, { 111, 12 }, WindowColour::primary, Widget::kContentNull, StringIds::tooltip_daymonthyear_challenge),
// Widgets::ImageButton({ 18, 15 }, { 20, 12 }, WindowColour::primary, ImageIds::speed_pause, StringIds::tooltip_speed_pause),
// Widgets::ImageButton({ 38, 15 }, { 20, 12 }, WindowColour::primary, ImageIds::speed_normal, StringIds::tooltip_speed_normal),
// Widgets::ImageButton({ 58, 15 }, { 20, 12 }, WindowColour::primary, ImageIds::speed_fast_forward, StringIds::tooltip_speed_fast_forward),
// Widgets::ImageButton({ 78, 15 }, { 20, 12 }, WindowColour::primary, ImageIds::speed_extra_fast_forward, StringIds::tooltip_speed_extra_fast_forward));
// static bool redrawScheduled = false; // 0x0050A004 (2nd bit)
// static const WindowEventList& getEvents();
// Window* open()
// auto window = WindowManager::createWindow(
// WindowType::timeToolbar,
// { Ui::width() - kWindowSize.width, Ui::height() - kWindowSize.height },
// { kWindowSize.width, kWindowSize.height },
// Ui::WindowFlags::stickToFront | Ui::WindowFlags::transparent | Ui::WindowFlags::noBackground,
// getEvents());
// window->setWidgets(_widgets);
// window->var_854 = 0;
// window->numTicksVisible = 0;
// window->initScrollWidgets();
// auto skin = ObjectManager::get<InterfaceSkinObject>();
// if (skin != nullptr)
// window->setColour(WindowColour::primary, AdvancedColour(skin->timeToolbarColour).translucent());
// window->setColour(WindowColour::secondary, AdvancedColour(skin->timeToolbarColour).translucent());
// orphan member: return window;
// // 0x004396A4
// func PrepareDraw(window [[maybe_unused]] Window) 
// window.widgets[Widx::inner_frame].hidden = true;
// window.widgets[Widx::pause_btn].image = Gfx::recolour(ImageIds::speed_pause);
// window.widgets[Widx::normal_speed_btn].image = Gfx::recolour(ImageIds::speed_normal);
// window.widgets[Widx::fast_forward_btn].image = Gfx::recolour(ImageIds::speed_fast_forward);
// window.widgets[Widx::extra_fast_forward_btn].image = Gfx::recolour(ImageIds::speed_extra_fast_forward);
// if (SceneManager::isPaused())
// window.widgets[Widx::pause_btn].image = Gfx::recolour(ImageIds::speed_pause_active);
// func If(SceneManager::getGameSpeed() else
// window.widgets[Widx::normal_speed_btn].image = Gfx::recolour(ImageIds::speed_normal_active);
// func If(SceneManager::getGameSpeed() else
// window.widgets[Widx::fast_forward_btn].image = Gfx::recolour(ImageIds::speed_fast_forward_active);
// func If(SceneManager::getGameSpeed() else
// window.widgets[Widx::extra_fast_forward_btn].image = Gfx::recolour(ImageIds::speed_extra_fast_forward_active);
// if (SceneManager::isNetworked())
// window.widgets[Widx::fast_forward_btn].hidden = true;
// window.widgets[Widx::extra_fast_forward_btn].hidden = true;
// window.widgets[Widx::pause_btn].left = 38;
// window.widgets[Widx::pause_btn].right = 57;
// window.widgets[Widx::normal_speed_btn].left = 58;
// window.widgets[Widx::normal_speed_btn].right = 77;
// else
// window.widgets[Widx::fast_forward_btn].hidden = false;
// window.widgets[Widx::extra_fast_forward_btn].hidden = false;
// window.widgets[Widx::pause_btn].left = 18;
// window.widgets[Widx::pause_btn].right = 37;
// window.widgets[Widx::normal_speed_btn].left = 38;
// window.widgets[Widx::normal_speed_btn].right = 57;
// window.widgets[Widx::fast_forward_btn].left = 58;
// window.widgets[Widx::fast_forward_btn].right = 77;
// window.widgets[Widx::extra_fast_forward_btn].left = 78;
// window.widgets[Widx::extra_fast_forward_btn].right = 97;
// // TODO: use same list as top toolbar
// static constexpr uint32_t map_sprites_by_rotation[] = {
// InterfaceSkin::ImageIds::toolbar_menu_map_north,
// InterfaceSkin::ImageIds::toolbar_menu_map_west,
// InterfaceSkin::ImageIds::toolbar_menu_map_south,
// InterfaceSkin::ImageIds::toolbar_menu_map_east,
// // 0x004397BE
// func Draw(self Ui::Window, drawingCtx Gfx::DrawingContext) 
// auto tr = Gfx::TextRenderer(drawingCtx);
// Widget& frame = self.widgets[Widx::outer_frame];
// drawingCtx.drawRect(self.x + frame.left, self.y + frame.top, frame.width(), frame.height(), enumValue(ExtColour::unk34), Gfx::RectFlags::transparent);
// // Draw widgets.
// self.draw(drawingCtx);
// drawingCtx.drawRectInset(self.x + frame.left + 1, self.y + frame.top + 1, frame.width() - 2, frame.height() - 2, self.getColour(WindowColour::secondary), Gfx::RectInsetFlags::borderInset | Gfx::RectInsetFlags::fillNone);
// orphan member: FormatArguments args{};
// args.push<uint32_t>(getCurrentDay());
// // Show "* Paused *" instead of the date half the time when paused, unless the browse prompt is pausing the game.
// StringId format = StringIds::date_daymonthyear;
// if (SceneManager::isPaused() && (SceneManager::getPauseFlags() & PauseFlags::browsePrompt) == PauseFlags::none)
// if (self.numTicksVisible >= kPausedStatusTextDuration)
// format = StringIds::toolbar_status_paused;
// auto c = self.getColour(WindowColour::primary).opaque();
// if (Input::isHovering(WindowType::timeToolbar, 0, Widx::date_btn))
// c = Colour::white;
// auto& widget = _widgets[Widx::date_btn];
// auto point = Point(self.x + widget.midX(), self.y + widget.top + 1);
// tr.drawStringCentred(point, c, format, args);
// auto skin = ObjectManager::get<InterfaceSkinObject>();
// drawingCtx.drawImage(self.x + _widgets[Widx::map_chat_menu].left - 2, self.y + _widgets[Widx::map_chat_menu].top - 1, skin->img + map_sprites_by_rotation[WindowManager::getCurrentRotation()]);
// // 0x004398FB
// func OnMouseUp(window [[maybe_unused]] Ui::Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case Widx::date_btn:
// MessageWindow::open();
// break;
// case Widx::pause_btn:
// GameCommands::doCommand(GameCommands::PauseGameArgs{}, GameCommands::Flags::apply);
// break;
// case Widx::normal_speed_btn:
// GameCommands::doCommand(GameCommands::SetGameSpeedArgs{ GameSpeed::Normal }, GameCommands::Flags::apply);
// break;
// case Widx::fast_forward_btn:
// GameCommands::doCommand(GameCommands::SetGameSpeedArgs{ GameSpeed::FastForward }, GameCommands::Flags::apply);
// break;
// case Widx::extra_fast_forward_btn:
// GameCommands::doCommand(GameCommands::SetGameSpeedArgs{ GameSpeed::ExtraFastForward }, GameCommands::Flags::apply);
// break;
// // 0x0043A67F
// func MapMouseDown(self Ui::Window, widgetIndex WidgetIndex_t) 
// auto skin = ObjectManager::get<InterfaceSkinObject>();
// if (SceneManager::isNetworked())
// Dropdown::add(0, StringIds::menu_sprite_stringid, { (uint32_t)skin->img + InterfaceSkin::ImageIds::phone, StringIds::chat_send_message });
// Dropdown::add(1, StringIds::menu_sprite_stringid, { (uint32_t)skin->img + map_sprites_by_rotation[WindowManager::getCurrentRotation()], StringIds::menu_map });
// Dropdown::showBelow(self, widgetIndex, 2, 25, (1 << 6));
// Dropdown::setHighlightedItem(1);
// else
// Dropdown::add(0, StringIds::menu_sprite_stringid, { (uint32_t)skin->img + map_sprites_by_rotation[WindowManager::getCurrentRotation()], StringIds::menu_map });
// Dropdown::showBelow(self, widgetIndex, 1, 25, (1 << 6));
// Dropdown::setHighlightedItem(0);
// func BeginSendChatMessage(self Window) 
// const auto* opponent = CompanyManager::getOpponent();
// auto args = FormatArguments::common();
// args.push(opponent->name);
// // TODO: convert this to a builder pattern, with chainable functions to set the different string ids and arguments
// TextInput::openTextInput(&self, StringIds::chat_title, StringIds::chat_instructions, StringIds::empty, Widx::map_chat_menu, args);
// // 0x0043A72F
// func MapDropdown(self Window, widgetIndex [[maybe_unused]] WidgetIndex_t, itemIndex int16) 
// if (itemIndex == -1)
// itemIndex = Dropdown::getHighlightedItem();
// if (SceneManager::isNetworked())
// switch (itemIndex)
// case 0:
// beginSendChatMessage(*self);
// break;
// case 1:
// MapWindow::open();
// break;
// else
// switch (itemIndex)
// case 0:
// MapWindow::open();
// break;
// // 0x043992E
// func OnMouseDown(window Ui::Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case Widx::map_chat_menu:
// mapMouseDown(&window, widgetIndex);
// break;
// // 0x439939
// func OnDropdown(w Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId, item_index int16) 
// switch (widgetIndex)
// case Widx::map_chat_menu:
// mapDropdown(&w, widgetIndex, item_index);
// break;
// // 0x00439944
// static Ui::CursorId onCursor([[maybe_unused]] Ui::Window& self, WidgetIndex_t widgetIdx, [[maybe_unused]] const WidgetId id, [[maybe_unused]] int16_t xPos, [[maybe_unused]] int16_t yPos, Ui::CursorId fallback)
// switch (widgetIdx)
// case Widx::date_btn:
// Ui::ToolTip::setTooltipTimeout(2000);
// break;
// orphan member: return fallback;
// // 0x00439955
// static std::optional<FormatArguments> tooltip([[maybe_unused]] Ui::Window& window, WidgetIndex_t widgetIndex, [[maybe_unused]] const WidgetId id)
// orphan member: FormatArguments args{};
// switch (widgetIndex)
// case Widx::date_btn:
// formatChallenge(args);
// break;
// orphan member: return args;
// // 0x0043995C
// func FormatChallenge(args FormatArguments) 
// args.push(getCurrentDay());
// auto playerCompany = CompanyManager::get(CompanyManager::getControllingId());
// if ((playerCompany->challengeFlags & CompanyFlags::challengeCompleted) != CompanyFlags::none)
// args.push(StringIds::challenge_completed);
// func If(CompanyFlags::challengeFailed (playerCompany->challengeFlags) else
// args.push(StringIds::challenge_failed);
// func If(CompanyFlags::challengeBeatenByOpponent (playerCompany->challengeFlags) else
// args.push(StringIds::empty);
// else
// args.push(StringIds::challenge_progress);
// args.push<uint16_t>(playerCompany->challengeProgress);
// if ((Scenario::getObjective().flags & Scenario::ObjectiveFlags::withinTimeLimit) != Scenario::ObjectiveFlags::none)
// uint16_t monthsLeft = (Scenario::getObjective().timeLimitYears * 12 - Scenario::getObjectiveProgress().monthsInChallenge);
// uint16_t yearsLeft = monthsLeft / 12;
// monthsLeft = monthsLeft % 12;
// args.push(StringIds::challenge_time_left);
// args.push(yearsLeft);
// args.push(monthsLeft);
// else
// args.push(StringIds::empty);
// // 0x00439A15
// func TextInput(w [[maybe_unused]] Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId, str byte) 
// switch (widgetIndex)
// case Widx::map_chat_menu:
// sendChatMessage(str);
// break;
// func SendChatMessage(string byte) 
// Network::sendChatMessage(string);
// func InvalidateFrame() 
// redrawScheduled = true;
// // 0x00439AD9
// func OnUpdate(w Window) 
// w.var_854 += 1;
// if (w.var_854 >= 24)
// w.var_854 = 0;
// w.numTicksVisible += 1;
// if (w.numTicksVisible >= kPausedStatusTextDuration * 2)
// w.numTicksVisible = 0;
// // Determine if the text needs to be updated
// if (SceneManager::isPaused() && (SceneManager::getPauseFlags() & PauseFlags::browsePrompt) == PauseFlags::none)
// if (w.numTicksVisible == 0 || w.numTicksVisible == kPausedStatusTextDuration)
// redrawScheduled = true;
// if (redrawScheduled)
// redrawScheduled = false;
// // Invalidating the inner frame widget effectively causes the entire time panel to be redrawn.
// WindowManager::invalidateWidget(WindowType::timeToolbar, 0, Widx::inner_frame);
// static constexpr WindowEventList kEvents = {
// .onMouseUp = onMouseUp,
// .onMouseHover = onMouseDown,
// .onMouseDown = onMouseDown,
// .onDropdown = onDropdown,
// .onUpdate = onUpdate,
// .textInput = textInput,
// .tooltip = tooltip,
// .cursor = onCursor,
// .prepareDraw = prepareDraw,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
