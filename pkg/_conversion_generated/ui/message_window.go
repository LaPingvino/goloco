package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Audio/Audio.h"
// #include "Config.h"
// #include "Date.h"
// #include "GameCommands/GameCommands.h"
// #include "Graphics/Colour.h"
// #include "Graphics/Gfx.h"
// #include "Graphics/ImageIds.h"
// #include "Graphics/RenderTarget.h"
// #include "Graphics/TextRenderer.h"
// #include "Input.h"
// #include "Intro.h"
// #include "Localisation/FormatArguments.hpp"
// #include "Localisation/StringIds.h"
// #include "Message.h"
// #include "MessageManager.h"
// #include "Objects/InterfaceSkinObject.h"
// #include "Objects/ObjectManager.h"
// #include "OpenLoco.h"
// #include "Ui.h"
// #include "Ui/Dropdown.h"
// #include "Ui/ScrollView.h"
// #include "Ui/ToolManager.h"
// #include "Ui/Widget.h"
// #include "Ui/Widgets/CaptionWidget.h"
// #include "Ui/Widgets/CheckboxWidget.h"
// #include "Ui/Widgets/DropdownWidget.h"
// #include "Ui/Widgets/FrameWidget.h"
// #include "Ui/Widgets/ImageButtonWidget.h"
// #include "Ui/Widgets/LabelWidget.h"
// #include "Ui/Widgets/PanelWidget.h"
// #include "Ui/Widgets/ScrollViewWidget.h"
// #include "Ui/Widgets/TabWidget.h"
// #include "Ui/WindowManager.h"
// #include "World/CompanyManager.h"
// namespace OpenLoco::Ui::Windows::MessageWindow
// namespace Common
type Widx int

const (
	Frame Widx = 0
	Caption Widx = 1
	Close_button Widx = 2
	Panel Widx = 3
	Tab_messages
	Tab_settings
)
// func MakeCommonWidgets(frameWidth int32, frameHeight int32, windowCaptionId StringId) any
// return makeWidgets(
// Widgets::Frame({ 0, 0 }, { frameWidth, frameHeight }, WindowColour::primary),
// Widgets::Caption({ 1, 1 }, { frameWidth - 2, 13 }, Widgets::Caption::Style::colourText, WindowColour::primary, windowCaptionId),
// Widgets::ImageButton({ frameWidth - 15, 2 }, { 13, 13 }, WindowColour::primary, ImageIds::close_button, StringIds::tooltip_close_window),
// Widgets::Panel({ 0, 41 }, { 366, 175 }, WindowColour::secondary),
// Widgets::Tab({ 3, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tooltip_recent_messages),
// Widgets::Tab({ 34, 15 }, { 31, 27 }, WindowColour::secondary, ImageIds::tab, StringIds::tooltip_message_options));
// func PrepareDraw(self Window) 
// func SwitchTab(self Window, widgetIndex WidgetIndex_t) 
// func OnUpdate(self Window) 
// func DrawTabs(self Window, drawingCtx Gfx::DrawingContext) 
// namespace Messages
// static constexpr Ui::Size kMinWindowSize = { 366, 217 };
// static constexpr Ui::Size kMaxWindowSize = { 366, 1200 };
// static int8_t messageHeight = 39;
type Widx int

const (
	Scrollview Widx = 6
)
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(366, 217, StringIds::title_messages),
// Widgets::ScrollView({ 3, 45 }, { 360, 146 }, WindowColour::secondary, Scrollbars::vertical)
// );
// // 0x0042A6F5
// func OnMouseUp(self Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case Common::widx::close_button:
// WindowManager::close(&self);
// break;
// case Common::widx::tab_messages:
// case Common::widx::tab_settings:
// Common::switchTab(self, widgetIndex);
// break;
// // 0x0042A95A
// func OnResize(self Window) 
// auto scrollview = self.widgets[widx::scrollview];
// auto scrollarea = self.scrollAreas[0];
// auto y = scrollarea.contentHeight - scrollview.height() - 1;
// y = std::max(0, y);
// if (y < scrollarea.contentOffsetY)
// scrollarea.contentOffsetY = y;
// Ui::ScrollView::updateThumbs(self, widx::scrollview);
// self.invalidate();
// // 0x0042A847
// func Event_08(self Window) 
// self.flags |= WindowFlags::notScrollView;
// // 0x0042A84F
// func Event_09(self Window) 
// if (!self.hasFlags(WindowFlags::notScrollView))
// return;
// if (self.rowHover == -1)
// return;
// self.rowHover = -1;
// self.invalidate();
// // 0x0042A871
// func GetScrollSize(self [[maybe_unused]] Window, scrollIndex [[maybe_unused]] uint32_t, scrollWidth [[maybe_unused]] int32_t, scrollHeight int32) 
// scrollHeight = MessageManager::getNumMessages() * messageHeight;
// // 0x0042A8B9
// func ScrollMouseDown(self Ui::Window, x [[maybe_unused]] int16_t, y int16, scrollIndex [[maybe_unused]] uint8_t) 
// auto messageIndex = y / messageHeight;
// if (messageIndex >= MessageManager::getNumMessages())
// return;
// if (MessageManager::getActiveIndex() != MessageId::null)
// auto message = MessageManager::get(MessageManager::getActiveIndex());
// if (message->isActive())
// // If the current active message was user selected then remove from queue of active messages
// if (message->isUserSelected())
// message->setActive(false);
// MessageManager::setActiveIndex(MessageId::null);
// WindowManager::close(WindowType::news, 0);
// auto message = MessageManager::get(MessageId(messageIndex));
// message->setUserSelected();
// message->timeActive++;
// NewsWindow::open(MessageId(messageIndex));
// int32_t pan = self.width / 2 + self.x;
// Audio::playSound(Audio::SoundId::clickDown, pan);
// // 0x0042A87C
// func ScrollMouseOver(self Ui::Window, x [[maybe_unused]] int16_t, y int16, scrollIndex [[maybe_unused]] uint8_t) 
// self.flags &= ~(WindowFlags::notScrollView);
// auto messageIndex = y / messageHeight;
// auto messageId = 0xFFFF;
// if (messageIndex < MessageManager::getNumMessages())
// messageId = messageIndex;
// if (self.rowHover != messageId)
// self.rowHover = messageId;
// self.invalidate();
// // 0x0042A70C
// static std::optional<FormatArguments> tooltip([[maybe_unused]] Ui::Window& self, [[maybe_unused]] WidgetIndex_t widgetIndex, [[maybe_unused]] const WidgetId id)
// orphan member: FormatArguments args{};
// args.push(StringIds::tooltip_scroll_message_list);
// orphan member: return args;
// // 0x0042A545
// func PrepareDraw(self Window) 
// Common::prepareDraw(self);
// self.widgets[widx::scrollview].right = self.width - 4;
// self.widgets[widx::scrollview].bottom = self.height - 14;
// // 0x0042A5CC
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// self.draw(drawingCtx);
// Common::drawTabs(self, drawingCtx);
// // 0x0042A5D7
// func DrawScroll(self Ui::Window, drawingCtx Gfx::DrawingContext, scrollIndex [[maybe_unused]] uint32_t) 
// auto colour = Colours::getShade(self.getColour(WindowColour::secondary).c(), 4);
// const auto& rt = drawingCtx.currentRenderTarget();
// auto tr = Gfx::TextRenderer(drawingCtx);
// drawingCtx.clearSingle(colour);
// auto height = 0;
// for (auto i = 0; i < MessageManager::getNumMessages(); i++)
// if (height + messageHeight <= rt.y)
// height += messageHeight;
// continue;
// func If(rt.height height >= rt.y +) else
// break;
// auto message = MessageManager::get(MessageId(i));
// char* buffer = message->messageString;
// auto str = const_cast<char*>(StringManager::getString(StringIds::buffer_2039));
// const size_t bufferLength = 512;
// strncpy(str, buffer, bufferLength);
// auto stringId = StringIds::black_stringid;
// if (self.rowHover == i)
// drawingCtx.drawRect(0, height, self.width, 38, enumValue(ExtColour::unk30), Gfx::RectFlags::transparent);
// stringId = StringIds::wcolour2_stringid;
// auto args = FormatArguments();
// args.push(StringIds::tiny_font_date);
// args.push(message->date);
// auto point = Point(0, height);
// tr.drawStringLeft(point, Colour::black, stringId, args);
// auto args = FormatArguments();
// args.push(StringIds::buffer_2039);
// auto width = self.widgets[widx::scrollview].width() - 14;
// auto point = Point(0, height + 6);
// tr.drawStringLeftWrapped(point, width, Colour::black, stringId, args);
// height += messageHeight;
// // 0x0042A7B9
// func TabReset(self Window) 
// self.minWidth = kMinWindowSize.width;
// self.minHeight = kMinWindowSize.height;
// self.maxWidth = kMaxWindowSize.width;
// self.maxHeight = kMaxWindowSize.height;
// self.width = kMinWindowSize.width;
// self.height = kMinWindowSize.height;
// self.rowHover = -1;
// static constexpr WindowEventList kEvents = {
// .onMouseUp = onMouseUp,
// .onResize = onResize,
// .onUpdate = Common::onUpdate,
// .event_08 = event_08,
// .event_09 = event_09,
// .getScrollSize = getScrollSize,
// .scrollMouseDown = scrollMouseDown,
// .scrollMouseOver = scrollMouseOver,
// .tooltip = tooltip,
// .prepareDraw = prepareDraw,
// .draw = draw,
// .drawScroll = drawScroll,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// // 0x0042A3FF
// func Open() 
// auto window = WindowManager::bringToFront(WindowType::messages);
// if (window != nullptr)
// if (ToolManager::isToolActive(window->type, window->number))
// ToolManager::toolCancel();
// window = WindowManager::bringToFront(WindowType::messages);
// if (window == nullptr)
// int16_t y = 29;
// int16_t x = Ui::width() - 366;
// window = WindowManager::createWindow(
// WindowType::messages,
// { x, y },
// { 366, 217 },
// WindowFlags::lighterFrame,
// Messages::getEvents());
// window->number = 0;
// window->currentTab = 0;
// window->frameNo = 0;
// window->rowHover = -1;
// window->disabledWidgets = 0;
// WindowManager::moveOtherWindowsDown(*window);
// window->minWidth = Messages::kMinWindowSize.width;
// window->minHeight = Messages::kMinWindowSize.height;
// window->maxWidth = Messages::kMaxWindowSize.width;
// window->maxHeight = Messages::kMaxWindowSize.height;
// window->flags |= WindowFlags::resizable;
// window->owner = CompanyManager::getControllingId();
// auto skin = ObjectManager::get<InterfaceSkinObject>();
// window->setColour(WindowColour::secondary, skin->windowPlayerColor);
// window->width = Messages::kMinWindowSize.width;
// window->height = Messages::kMinWindowSize.height;
// window->currentTab = 0;
// window->invalidate();
// window->setWidgets(Messages::widgets);
// window->holdableWidgets = 0;
// window->eventHandlers = &Messages::getEvents();
// window->disabledWidgets = 0;
// window->callOnResize();
// window->callPrepareDraw();
// window->initScrollWidgets();
// int32_t scrollWidth = 0, scrollHeight = 0;
// window->callGetScrollSize(0, scrollWidth, scrollHeight);
// scrollHeight -= window->widgets[Messages::widx::scrollview].height();
// if (static_cast<int16_t>(scrollHeight) < 0)
// scrollHeight = 0;
// window->scrollAreas[0].contentOffsetY = scrollHeight;
// Ui::ScrollView::updateThumbs(*window, Messages::widx::scrollview);
// namespace Settings
// static constexpr Ui::Size kWindowSize = { 366, 155 };
var NumWidgetsPerDropdown = 3 // auto
type Widx int

const (
	Company_major_news_label Widx = 6
	Company_major_news
	Company_major_news_dropdown
	Competitor_major_news_label
	Competitor_major_news
	Competitor_major_news_dropdown
	Company_minor_news_label
	Company_minor_news
	Company_minor_news_dropdown
	Competitor_minor_news_label
	Competitor_minor_news
	Competitor_minor_news_dropdown
	General_news_label
	General_news
	General_news_dropdown
	Advice_label
	Advice
	Advice_dropdown
	PlaySoundEffects
)
// static constexpr auto widgets = makeWidgets(
// Common::makeCommonWidgets(366, 155, StringIds::title_messages),
// Widgets::Label({ 4, 47 }, { 230, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::company_major_news),
// Widgets::dropdownWidgets({ 236, 47 }, { 124, 12 }, WindowColour::secondary),
// Widgets::Label({ 4, 62 }, { 230, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::competitor_major_news),
// Widgets::dropdownWidgets({ 236, 62 }, { 124, 12 }, WindowColour::secondary),
// Widgets::Label({ 4, 77 }, { 230, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::company_minor_news),
// Widgets::dropdownWidgets({ 236, 77 }, { 124, 12 }, WindowColour::secondary),
// Widgets::Label({ 4, 92 }, { 230, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::competitor_minor_news),
// Widgets::dropdownWidgets({ 236, 92 }, { 124, 12 }, WindowColour::secondary),
// Widgets::Label({ 4, 107 }, { 230, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::general_news),
// Widgets::dropdownWidgets({ 236, 107 }, { 124, 12 }, WindowColour::secondary),
// Widgets::Label({ 4, 122 }, { 230, 12 }, WindowColour::secondary, ContentAlign::left, StringIds::advice),
// Widgets::dropdownWidgets({ 236, 122 }, { 124, 12 }, WindowColour::secondary),
// Widgets::Checkbox({ 4, 137 }, { 346, 12 }, WindowColour::secondary, StringIds::playNewsSoundEffects, StringIds::playNewsSoundEffectsTip)
// );
// // 0x0042AA84
// func OnMouseUp(self Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case Common::widx::close_button:
// WindowManager::close(&self);
// break;
// case Common::widx::tab_messages:
// case Common::widx::tab_settings:
// Common::switchTab(self, widgetIndex);
// break;
// case widx::playSoundEffects:
// Config::get().audio.playNewsSounds ^= 1;
// Config::write();
// WindowManager::invalidateWidget(WindowType::messages, self.number, widgetIndex);
// break;
// constexpr StringId kNewsDropdownStringIds[] = {
// StringIds::message_off,
// StringIds::message_ticker,
// StringIds::message_window,
// // 0x0042AA9F
// func OnMouseDown(self Window, widgetIndex WidgetIndex_t, id [[maybe_unused]] WidgetId) 
// switch (widgetIndex)
// case widx::company_major_news_dropdown:
// case widx::competitor_major_news_dropdown:
// case widx::company_minor_news_dropdown:
// case widx::competitor_minor_news_dropdown:
// case widx::general_news_dropdown:
// case widx::advice_dropdown:
// auto wIndex = widgetIndex - 1;
// auto widget = self.widgets[wIndex];
// auto xPos = widget.left + self.x;
// auto yPos = widget.top + self.y;
// auto width = widget.width() - 2;
// auto height = widget.height() + 2;
// auto flags = 1 << 7;
// Dropdown::show(xPos, yPos, width, height, self.getColour(WindowColour::secondary), 3, flags);
// for (auto i = 0U; i < std::size(kNewsDropdownStringIds); i++)
// Dropdown::add(i, StringIds::dropdown_stringid, kNewsDropdownStringIds[i]);
// auto ddIndex = wIndex - widx::company_major_news;
// auto currentItem = Config::get().newsSettings[ddIndex / kNumWidgetsPerDropdown];
// Config::write();
// Dropdown::setItemSelected(static_cast<size_t>(currentItem));
// break;
// // 0x0042AAAC
// func OnDropdown(self [[maybe_unused]] Window, widgetIndex Ui::WidgetIndex_t, id [[maybe_unused]] WidgetId, itemIndex int16) 
// switch (widgetIndex)
// case widx::company_major_news_dropdown:
// case widx::competitor_major_news_dropdown:
// case widx::company_minor_news_dropdown:
// case widx::competitor_minor_news_dropdown:
// case widx::general_news_dropdown:
// case widx::advice_dropdown:
// if (itemIndex == -1)
// return;
// auto dropdownIndex = (widgetIndex - widx::company_major_news) / kNumWidgetsPerDropdown;
// auto newValue = static_cast<Config::NewsType>(itemIndex);
// if (newValue != Config::get().newsSettings[dropdownIndex])
// Config::get().newsSettings[dropdownIndex] = newValue;
// Config::write();
// Gfx::invalidateScreen();
// break;
// func PrepareDraw(self Window) 
// Common::prepareDraw(self);
// if (Config::get().audio.playNewsSounds)
// self.activatedWidgets |= (1 << widx::playSoundEffects);
// else
// self.activatedWidgets &= ~(1 << widx::playSoundEffects);
// for (auto i = 0; i < 6; i++)
// auto widgetIndex = widx::company_major_news + (kNumWidgetsPerDropdown * i);
// auto setting = static_cast<uint8_t>(Config::get().newsSettings[i]);
// self.widgets[widgetIndex].text = kNewsDropdownStringIds[setting];
// // 0x0042AA02
// func Draw(self Window, drawingCtx Gfx::DrawingContext) 
// self.draw(drawingCtx);
// Common::drawTabs(self, drawingCtx);
// // 0x0042A7E8
// func TabReset(self Window) 
// self.minWidth = kWindowSize.width;
// self.minHeight = kWindowSize.height;
// self.maxWidth = kWindowSize.width;
// self.maxHeight = kWindowSize.height;
// self.width = kWindowSize.width;
// self.height = kWindowSize.height;
// static constexpr WindowEventList kEvents = {
// .onMouseUp = onMouseUp,
// .onMouseDown = onMouseDown,
// .onDropdown = onDropdown,
// .onUpdate = Common::onUpdate,
// .prepareDraw = prepareDraw,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// namespace Common
type TabInformation struct {
// std::span<const Widget> widgets;
// const widx widgetIndex;
// const WindowEventList& events;
}
// static TabInformation tabInformationByTabOffset[] = {
// { Messages::widgets, widx::tab_messages, Messages::getEvents() },
// { Settings::widgets, widx::tab_settings, Settings::getEvents() },
// func PrepareDraw(self Window) 
// // Activate the current tab..
// self.activatedWidgets &= ~((1ULL << tab_messages) | (1ULL << tab_settings));
// self.activatedWidgets |= (1ULL << tabInformationByTabOffset[self.currentTab].widgetIndex);
// self.widgets[Common::widx::frame].right = self.width - 1;
// self.widgets[Common::widx::frame].bottom = self.height - 1;
// self.widgets[Common::widx::panel].right = self.width - 1;
// self.widgets[Common::widx::panel].bottom = self.height - 1;
// self.widgets[Common::widx::caption].right = self.width - 2;
// self.widgets[Common::widx::close_button].left = self.width - 15;
// self.widgets[Common::widx::close_button].right = self.width - 3;
// // 0x0042A716
// func SwitchTab(self Window, widgetIndex WidgetIndex_t) 
// if (ToolManager::isToolActive(self.type, self.number))
// ToolManager::toolCancel();
// self.currentTab = widgetIndex - widx::tab_messages;
// self.frameNo = 0;
// self.flags &= ~(WindowFlags::beingResized);
// self.viewportRemove(0);
// const auto& tabInfo = tabInformationByTabOffset[widgetIndex - widx::tab_messages];
// self.holdableWidgets = 0;
// self.eventHandlers = &tabInfo.events;
// self.activatedWidgets = 0;
// self.setWidgets(tabInfo.widgets);
// self.disabledWidgets = 0;
// self.invalidate();
// if (self.currentTab == widx::tab_messages - widx::tab_messages)
// Messages::tabReset(self);
// if (self.currentTab == widx::tab_settings - widx::tab_messages)
// Settings::tabReset(self);
// self.callOnResize();
// self.callPrepareDraw();
// self.initScrollWidgets();
// self.invalidate();
// self.moveInsideScreenEdges();
// // 0x0042AB92
// func DrawTabs(self Window, drawingCtx Gfx::DrawingContext) 
// auto skin = ObjectManager::get<InterfaceSkinObject>();
// // Message Tab
// uint32_t imageId = skin->img;
// imageId += InterfaceSkin::ImageIds::tab_messages;
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_messages);
// // Setting Tab
// uint32_t imageId = skin->img;
// imageId += InterfaceSkin::ImageIds::tab_message_settings;
// Widget::drawTab(self, drawingCtx, imageId, widx::tab_settings);
// // 0x0042A826 and 0x0042AB6A
// func OnUpdate(self Window) 
// self.frameNo++;
// self.callPrepareDraw();
// WindowManager::invalidateWidget(WindowType::messages, self.number, self.currentTab + Common::widx::tab_messages);
