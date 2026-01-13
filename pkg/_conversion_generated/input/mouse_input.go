package input

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Audio/Audio.h"
// #include "Config.h"
// #include "Entities/EntityManager.h"
// #include "Input.h"
// #include "Localisation/StringIds.h"
// #include "Logging.h"
// #include "Map/BuildingElement.h"
// #include "Map/MapSelection.h"
// #include "Map/RoadElement.h"
// #include "Map/SignalElement.h"
// #include "Map/StationElement.h"
// #include "Map/TileManager.h"
// #include "Map/TrackElement.h"
// #include "Map/TreeElement.h"
// #include "Map/WallElement.h"
// #include "Objects/ObjectManager.h"
// #include "OpenLoco.h"
// #include "SceneManager.h"
// #include "Tutorial.h"
// #include "Ui/Dropdown.h"
// #include "Ui/ScrollView.h"
// #include "Ui/ToolManager.h"
// #include "Ui/ToolTip.h"
// #include "Ui/ViewportInteraction.h"
// #include "Ui/Widget.h"
// #include "Ui/WindowManager.h"
// #include "Vehicles/Vehicle.h"
// #include "World/CompanyManager.h"
// #include "World/StationManager.h"
// #include "World/TownManager.h"
// #include <map>
// #include <queue>
// using namespace OpenLoco::Ui;
// using namespace OpenLoco::Ui::ScrollView;
// using namespace OpenLoco::Ui::ViewportInteraction;
// using namespace OpenLoco::World;
// namespace OpenLoco::Input
// func StateScrollLeft(cx MouseButton, edx WidgetIndex_t, window Ui::Window, widget Ui::Widget, x int32, y int32) 
// func StateScrollRight(button MouseButton) 
// func StateResizing(button MouseButton, x int32, y int32) 
// func StateWidgetPressed(button MouseButton, x int32, y int32, window Ui::Window, widget Ui::Widget, widgetIndex Ui::WidgetIndex_t) 
// func StateNormal(state MouseButton, x int32, y int32, window Ui::Window, widget Ui::Widget, widgetIndex Ui::WidgetIndex_t) 
// func StateNormalHover(x int32, y int32, window Ui::Window, widget Ui::Widget, widgetIndex Ui::WidgetIndex_t) 
// func StateNormalLeft(x int32, y int32, window Ui::Window, widgetIndex Ui::WidgetIndex_t) 
// func StateNormalRight(x int32, y int32, window Ui::Window, widgetIndex Ui::WidgetIndex_t) 
// func StatePositioningWindow(button MouseButton, x int32, y int32, window Ui::Window, widget Ui::Widget, widgetIndex Ui::WidgetIndex_t) 
// func WindowPositionEnd() 
// func WindowResizeBegin(x int32, y int32, window Ui::Window, widgetIndex Ui::WidgetIndex_t) 
// func ViewportDragBegin(w Window) 
// func ScrollDragBegin(x int32, y int32, pWindow Window, index WidgetIndex_t) 
// func WidgetOverFlatbuttonInvalidate() 
// static bool _pendingMouseInputUpdate; // 0x00525324
// static MouseButton _lastKnownButtonState;
// static Ui::Point _cursorPressed;
// static Ui::CursorId _52336C;
// static Ui::Point _cursor;
// static Ui::Point _cursor2;
// static Ui::WindowType _pressedWindowType;       // 0x0052336F
// static Ui::WindowNumber_t _pressedWindowNumber; // 0x00523370
// static int32_t _pressedWidgetIndex;             // 0x00523372
// static uint16_t _clickRepeatTicks;              // 0x00523376
// static Ui::Point _dragLast;                     // 0x00523378
// static Ui::WindowNumber_t _dragWindowNumber;    // 0x0052337C
// static Ui::WindowType _dragWindowType;          // 0x0052337E
// static uint8_t _dragWidgetIndex;                // 0x0052337F
// static uint8_t _dragScrollIndex;                // 0x00523380
// static uint16_t _ticksSinceDragStart; // 0x0052338E
// static Ui::Point _scrollLast;           // 0x005233A4
// static Ui::WindowType _hoverWindowType; // 0x005233A8
// static uint8_t _5233A9;
// static Ui::WindowNumber_t _hoverWindowNumber; // 0x005233AA
// static Ui::WidgetIndex_t _hoverWidgetIdx;     // 0x005233AC
// static int32_t _mouseDeltaX; // 0x0114084C
// static int32_t _mouseDeltaY; // 0x01140840
// static int32_t _mousePosX;   // 0x005233AE
// static int32_t _mousePosY;   // 0x005233B2
// static Ui::WindowType _focusedWindowType;
// static Ui::WindowNumber_t _focusedWindowNumber;
// static Ui::WidgetIndex_t _focusedWidgetIndex;
// static bool _rightMouseButtonDown;
// static StationId _hoveredStationId = StationId::null; // 0x00F252A4
// static int32_t _cursorWheel;
// static const std::map<Ui::ScrollPart, StringId> kScrollWidgetTooltips = {
// { Ui::ScrollPart::hscrollbarButtonLeft, StringIds::tooltip_scroll_left },
// { Ui::ScrollPart::hscrollbarButtonRight, StringIds::tooltip_scroll_right },
// { Ui::ScrollPart::hscrollbarTrackLeft, StringIds::tooltip_scroll_left_fast },
// { Ui::ScrollPart::hscrollbarTrackRight, StringIds::tooltip_scroll_right_fast },
// { Ui::ScrollPart::hscrollbarThumb, StringIds::tooltip_scroll_left_right },
// { Ui::ScrollPart::vscrollbarButtonTop, StringIds::tooltip_scroll_up },
// { Ui::ScrollPart::vscrollbarButtonBottom, StringIds::tooltip_scroll_down },
// { Ui::ScrollPart::vscrollbarTrackTop, StringIds::tooltip_scroll_up_fast },
// { Ui::ScrollPart::vscrollbarTrackBottom, StringIds::tooltip_scroll_down_fast },
// { Ui::ScrollPart::vscrollbarThumb, StringIds::tooltip_scroll_up_down },
const DropdownItemUndefined int32 = -1
// func InitMouse() 
// _pressedWindowType = Ui::WindowType::undefined;
// Ui::ToolTip::setNotShownTicks(0xFFFFU);
// _hoverWindowType = Ui::WindowType::undefined;
// _focusedWindowType = Ui::WindowType::undefined;
// _mousePosX = 0;
// _mousePosY = 0;
// World::resetMapSelectionFlags();
// func MoveMouse(x int32, y int32, relX int32, relY int32) 
// _cursor = { x, y };
// _mouseDeltaX += relX;
// _mouseDeltaY += relY;
// func ProcessMouseMovement() 
// _mousePosX += _mouseDeltaX;
// _mousePosY += _mouseDeltaY;
// _mouseDeltaX = 0;
// _mouseDeltaY = 0;
// func MouseWheel(wheel int) 
// _cursorWheel += wheel;
// func IsHovering(type Ui::WindowType) bool
// return _hoverWindowType == type;
// func IsHovering(type Ui::WindowType, number Ui::WindowNumber_t) bool
// return (_hoverWindowType == type) && (_hoverWindowNumber == number);
// func IsHovering(type Ui::WindowType, number Ui::WindowNumber_t, widgetIndex Ui::WidgetIndex_t) bool
// return _hoverWindowType == type && _hoverWindowNumber == number && _hoverWidgetIdx == widgetIndex;
// Ui::WidgetIndex_t getHoveredWidgetIndex()
// orphan member: return _hoverWidgetIdx;
// func IsDropdownActive(type Ui::WindowType, number Ui::WindowNumber_t) bool
// if (state() != State::dropdownActive)
// orphan member: return false;
// if (_pressedWindowType != type)
// orphan member: return false;
// if (!hasFlag(Flags::widgetPressed))
// orphan member: return false;
// if (_pressedWindowNumber != number)
// orphan member: return false;
// orphan member: return true;
// func IsDropdownActive(type Ui::WindowType, number Ui::WindowNumber_t, index Ui::WidgetIndex_t) bool
// if (!isDropdownActive(type, number))
// orphan member: return false;
// return _pressedWidgetIndex == index;
// func IsPressed(type Ui::WindowType, number Ui::WindowNumber_t) bool
// if (state() != State::widgetPressed)
// orphan member: return false;
// if (_pressedWindowType != type)
// orphan member: return false;
// if (_pressedWindowNumber != number)
// orphan member: return false;
// if (!hasFlag(Flags::widgetPressed))
// orphan member: return false;
// orphan member: return true;
// func IsPressed(type Ui::WindowType, number Ui::WindowNumber_t, index Ui::WidgetIndex_t) bool
// func IsPressed(type, number) return
// Ui::WidgetIndex_t getPressedWidgetIndex()
// orphan member: return _pressedWidgetIndex;
// func SetPressedWidgetIndex(index Ui::WidgetIndex_t) 
// _pressedWidgetIndex = index;
// func IsFocused(type Ui::WindowType, number Ui::WindowNumber_t) bool
// if (!hasFlag(Flags::widgetFocused))
// orphan member: return false;
// if (_focusedWindowType != type)
// orphan member: return false;
// if (_focusedWindowNumber != number)
// orphan member: return false;
// orphan member: return true;
// func IsFocused(type Ui::WindowType, number Ui::WindowNumber_t, index Ui::WidgetIndex_t) bool
// func IsFocused(type, number) return
// func SetFocus(type Ui::WindowType, number Ui::WindowNumber_t, index Ui::WidgetIndex_t) 
// _focusedWindowType = type;
// _focusedWindowNumber = number;
// _focusedWidgetIndex = index;
// setFlag(Flags::widgetFocused);
// func ResetFocus() 
// _focusedWindowType = WindowType::undefined;
// resetFlag(Flags::widgetFocused);
// // 0x004C6E65
// func UpdateCursorPosition() 
// switch (Tutorial::state())
// case Tutorial::State::none:
// _cursor2 = _cursor;
// break;
// case Tutorial::State::playing:
// _cursor2.x = Tutorial::nextInput();
// _cursor2.y = Tutorial::nextInput();
// Ui::setCursorPosScaled(_cursor2.x, _cursor2.y);
// break;
// case Tutorial::State::recording:
// // Vanilla had tutorial recording here at 0x004C6EC3
// // as tutorials are fixed mouse position there isn't much
// // point implementing this code as per vanilla.
// break;
// func GetHoveredStationId() StationId
// orphan member: return _hoveredStationId;
// func SetHoveredStationId(stationId StationId) 
// _hoveredStationId = stationId;
// func StateViewportLeft(cx MouseButton, x int32, y int32) 
// func StateViewportRight(cx MouseButton, x int32, y int32) 
// // 0x004C7174
// func HandleMouse(x int32, y int32, button MouseButton) 
// _lastKnownButtonState = button;
// Ui::Window* window = WindowManager::findAt(x, y);
// // TODO: I think window can never be null, there is always a main window,
// //       validate this and work with references from here on.
// Ui::WidgetIndex_t widgetIndex = kWidgetIndexNull;
// if (window != nullptr)
// widgetIndex = window->findWidgetAt(x, y);
// auto modalType = WindowManager::getCurrentModalType();
// if (modalType != WindowType::undefined)
// if (window != nullptr)
// if (window->type != modalType)
// if (button == MouseButton::leftPressed)
// WindowManager::bringToFront(modalType);
// Audio::playSound(Audio::SoundId::error, x);
// return;
// if (button == MouseButton::rightPressed)
// return;
// Ui::Widget* widget = nullptr;
// if (widgetIndex != kWidgetIndexNull)
// widget = &window->widgets[widgetIndex];
// switch (state())
// case State::reset:
// Ui::ToolTip::setTooltipMouseLocation({ x, y });
// Ui::ToolTip::setTooltipTimeout(0);
// Ui::ToolTip::setWindowType(Ui::WindowType::undefined);
// state(State::normal);
// resetFlag(Flags::leftMousePressed);
// stateNormal(button, x, y, window, widget, widgetIndex);
// break;
// case State::normal:
// stateNormal(button, x, y, window, widget, widgetIndex);
// break;
// case State::widgetPressed:
// case State::dropdownActive:
// stateWidgetPressed(button, x, y, window, widget, widgetIndex);
// break;
// case State::positioningWindow:
// statePositioningWindow(button, x, y, window, widget, widgetIndex);
// break;
// case State::viewportRight:
// stateViewportRight(button, x, y);
// break;
// case State::viewportLeft:
// stateViewportLeft(button, x, y);
// break;
// case State::scrollLeft:
// stateScrollLeft(button, widgetIndex, window, widget, x, y);
// break;
// case State::resizing:
// stateResizing(button, x, y);
// break;
// case State::scrollRight:
// stateScrollRight(button);
// break;
// // 0x004C7334
// // Left-clicking on a view of the game world (e.g. using terraforming tools, clicking vehicles, buildings, labels)
// func StateViewportLeft(button MouseButton, x int32, y int32) 
// auto window = WindowManager::find(_dragWindowType, _dragWindowNumber);
// if (window == nullptr)
// Input::state(State::reset);
// return;
// switch (button)
// case MouseButton::released:
// // 0x4C735D
// auto viewport = window->viewports[0];
// if (viewport == nullptr)
// viewport = window->viewports[1];
// if (viewport == nullptr)
// Input::state(State::reset);
// return;
// if (window->type == _dragWindowType && window->number == _dragWindowNumber)
// if (Input::hasFlag(Flags::toolActive))
// auto tool = WindowManager::find(ToolManager::getToolWindowType(), ToolManager::getToolWindowNumber());
// if (tool != nullptr)
// // TODO: Handle widget id properly for tools.
// tool->callToolDrag(ToolManager::getToolWidgetIndex(), WidgetId::none, x, y);
// break;
// case MouseButton::leftReleased:
// // 0x4C73C2
// Input::state(State::reset);
// if (window->type != _dragWindowType || window->number != _dragWindowNumber)
// return;
// if (hasFlag(Flags::toolActive))
// auto tool = WindowManager::find(ToolManager::getToolWindowType(), ToolManager::getToolWindowNumber());
// if (tool != nullptr)
// // TODO: Handle widget id properly for tools.
// tool->callToolUp(ToolManager::getToolWidgetIndex(), WidgetId::none, x, y);
// func If(!hasFlag(Flags::leftMousePressed) else
// auto interaction = ViewportInteraction::getItemLeft(x, y);
// switch (interaction.type)
// case InteractionItem::entity:
// auto _entity = reinterpret_cast<EntityBase*>(interaction.object);
// auto veh = _entity->asBase<Vehicles::VehicleBase>();
// if (veh != nullptr)
// Ui::Windows::Vehicle::Main::open(veh);
// break;
// case InteractionItem::townLabel:
// Ui::Windows::Town::open(interaction.value);
// break;
// case InteractionItem::stationLabel:
// Ui::Windows::Station::open(StationId(interaction.value));
// break;
// case InteractionItem::headquarterBuilding:
// const auto* tileElement = reinterpret_cast<World::TileElement*>(interaction.object);
// const auto* building = tileElement->as<World::BuildingElement>();
// if (building != nullptr)
// auto index = building->sequenceIndex();
// const auto firstTile = interaction.pos - World::kOffsets[index];
// const Pos3 pos = { firstTile.x,
// firstTile.y,
// building->baseZ() };
// for (auto& company : CompanyManager::companies())
// if (company.headquartersX == pos.x
// && company.headquartersY == pos.y
// && company.headquartersZ == pos.z)
// Ui::Windows::CompanyWindow::open(company.id());
// break;
// break;
// case InteractionItem::industry:
// Ui::Windows::Industry::open(IndustryId(interaction.value));
// break;
// default:
// break;
// break;
// default:
// break;
// // 0x004C74BB
// // Right mouse dragging in viewports, such as the main display of the game world.
// func StateViewportRight(button MouseButton, x int32, y int32) 
// auto window = WindowManager::find(_dragWindowType, _dragWindowNumber);
// if (window == nullptr)
// Input::state(State::reset);
// return;
// switch (button)
// case MouseButton::released:
// // 4C74E4
// _ticksSinceDragStart += getTimeSinceLastTick();
// auto vp = window->viewports[0];
// if (vp == nullptr)
// vp = window->viewports[1];
// if (vp == nullptr)
// Input::state(State::reset);
// return;
// if (window->hasFlags(WindowFlags::viewportNoScrolling))
// return;
// Ui::Point dragOffset = { x, y };
// if (Tutorial::state() != Tutorial::State::playing)
// // Fix #151: use relative drag from one frame to the next rather than
// //           using the relative position from the message loop
// dragOffset = getNextDragOffset();
// if (dragOffset.x != 0 || dragOffset.y != 0)
// _ticksSinceDragStart = 1000;
// auto* main = WindowManager::getMainWindow();
// if (Windows::Main::viewportIsFocusedOnAnyEntity(*main))
// Windows::Main::viewportUnfocusFromEntity(*main);
// else
// const auto offsetX = dragOffset.x << (vp->zoom + 1);
// const auto offsetY = dragOffset.y << (vp->zoom + 1);
// const auto invert = Config::get().invertRightMouseViewPan ? -1 : 1;
// window->viewportConfigurations[0].savedViewX += offsetX * invert;
// window->viewportConfigurations[0].savedViewY += offsetY * invert;
// break;
// case MouseButton::rightReleased:
// if (_ticksSinceDragStart > 500)
// Input::state(State::reset);
// return;
// Input::state(State::reset);
// ViewportInteraction::handleRightReleased(window, _dragLast.x, _dragLast.y);
// break;
// default:
// break;
// // 0x004C71F6
// func StateScrollLeft(button MouseButton, widgetIndex WidgetIndex_t, window Ui::Window, widget Ui::Widget, x int32, y int32) 
// switch (button)
// case MouseButton::released:
// if (widgetIndex != _pressedWidgetIndex || window->type != _pressedWindowType || window->number != _pressedWindowNumber)
// ScrollView::clearPressedButtons(_pressedWindowType, _pressedWindowNumber, _pressedWidgetIndex);
// return;
// ScrollView::scrollLeftContinue(x, y, *window, widget, widgetIndex);
// break;
// case MouseButton::leftReleased:
// Input::state(State::reset);
// ScrollView::clearPressedButtons(_pressedWindowType, _pressedWindowNumber, _pressedWidgetIndex);
// break;
// default:
// break;
// // 0x004C76A7
// // regs.cx = (uint16_t)button;
// // regs.ax = x;
// // regs.bx = y;
// // Right mouse dragging in scrollview widgets.
// func StateScrollRight(button MouseButton) 
// auto window = WindowManager::find(_dragWindowType, _dragWindowNumber);
// if (window == nullptr)
// Input::state(State::reset);
// return;
// switch (button)
// case MouseButton::released:
// _ticksSinceDragStart += getTimeSinceLastTick();
// const Ui::Point dragOffset = getNextDragOffset();
// if (dragOffset.x != 0 || dragOffset.y != 0)
// _ticksSinceDragStart = 1000;
// Ui::Widget& widget = window->widgets[_dragWidgetIndex];
// const auto invert = Config::get().invertRightMouseViewPan ? -1 : 1;
// Ui::ScrollView::horizontalDragFollow(*window, &widget, _dragWidgetIndex, _dragScrollIndex, dragOffset.x * invert);
// Ui::ScrollView::verticalDragFollow(*window, &widget, _dragWidgetIndex, _dragScrollIndex, dragOffset.y * invert);
// break;
// case MouseButton::rightReleased:
// Input::state(State::reset);
// // in the original assembly code we load into registers values from _dragLast.x, _dragLast.y
// // if _ticksSinceDragStart <= 500, however the result was unused
// break;
// default:
// break;
// func GetLastKnownButtonState() MouseButton
// orphan member: return _lastKnownButtonState;
// // 0x004C7722
// func StateResizing(button MouseButton, x int32, y int32) 
// auto w = WindowManager::find(_dragWindowType, _dragWindowNumber);
// if (w == nullptr)
// state(State::reset);
// return;
// bool doDefault = false;
// int dx = 0, dy = 0;
// switch (button)
// case MouseButton::released:
// doDefault = true;
// break;
// case MouseButton::leftReleased:
// state(State::normal);
// Ui::ToolTip::setTooltipTimeout(0);
// Ui::ToolTip::setWidgetIndex(_pressedWidgetIndex);
// Ui::ToolTip::setWindowType(_dragWindowType);
// Ui::ToolTip::setWindowNumber(_dragWindowNumber);
// if (w->hasFlags(Ui::WindowFlags::finishedResize))
// doDefault = true;
// break;
// if (w->hasFlags(Ui::WindowFlags::beingResized))
// x = w->var_88A - w->width + _dragLast.x;
// y = w->var_88C - w->height + _dragLast.y;
// w->flags &= ~Ui::WindowFlags::beingResized;
// doDefault = true;
// break;
// w->var_88A = w->width;
// w->var_88C = w->height;
// x = _dragLast.x - w->x - w->width + Ui::width();
// y = _dragLast.y - w->y - w->height + Ui::height() - 27;
// w->flags |= Ui::WindowFlags::beingResized;
// if (y >= Ui::height() - 2)
// _dragLast.x = x;
// _dragLast.y = y;
// return;
// dx = x - _dragLast.x;
// dy = y - _dragLast.y;
// if (dx == 0 && dy == 0)
// _dragLast.x = x;
// _dragLast.y = y;
// return;
// break;
// default:
// return;
// if (doDefault)
// if (y >= Ui::height() - 2)
// _dragLast.x = x;
// _dragLast.y = y;
// return;
// dx = x - _dragLast.x;
// dy = y - _dragLast.y;
// if (dx == 0 && dy == 0)
// _dragLast.x = x;
// _dragLast.y = y;
// return;
// w->flags &= ~Ui::WindowFlags::beingResized;
// w->invalidate();
// w->width = std::clamp(w->width + dx, w->minWidth, w->maxWidth);
// w->height = std::clamp(w->height + dy, w->minHeight, w->maxHeight);
// w->flags |= Ui::WindowFlags::finishedResize;
// w->callOnResize();
// w->callPrepareDraw();
// w->scrollAreas[0].contentWidth = -1;
// w->scrollAreas[0].contentHeight = -1;
// w->scrollAreas[1].contentWidth = -1;
// w->scrollAreas[1].contentHeight = -1;
// w->updateScrollWidgets();
// w->invalidate();
// _dragLast.x = x;
// _dragLast.y = y;
// // 0x004C7903
// func StatePositioningWindow(button MouseButton, x int32, y int32, window [[maybe_unused]] Ui::Window, widget [[maybe_unused]] Ui::Widget, widgetIndex [[maybe_unused]] Ui::WidgetIndex_t) 
// auto w = WindowManager::find(_dragWindowType, _dragWindowNumber);
// if (w == nullptr)
// state(State::reset);
// return;
// switch (button)
// case MouseButton::released:
// y = std::clamp(y, 29, Ui::height() - 29);
// auto dx = x - _dragLast.x;
// auto dy = y - _dragLast.y;
// if (w->move(dx, dy))
// _5233A9 = true;
// _dragLast.x = x;
// _dragLast.y = y;
// break;
// case MouseButton::leftReleased:
// windowPositionEnd();
// y = std::clamp(y, 29, Ui::height() - 29);
// int dx = x - _dragLast.x;
// int dy = y - _dragLast.y;
// if (w->move(dx, dy))
// _5233A9 = true;
// _dragLast.x = x;
// _dragLast.y = y;
// if (_5233A9 == false)
// auto dragWindow = WindowManager::find(_dragWindowType, _dragWindowNumber);
// if (dragWindow != nullptr)
// if (dragWindow->isEnabled(_pressedWidgetIndex))
// auto pressedWidget = &dragWindow->widgets[_pressedWidgetIndex];
// Audio::playSound(Audio::SoundId::clickPress, dragWindow->x + pressedWidget->midX());
// dragWindow->callOnMouseUp(_pressedWidgetIndex, pressedWidget->id);
// w->callOnMove(_dragLast.x, _dragLast.y);
// break;
// default:
// break;
// func DropdownRegisterSelection(item int16) 
// auto window = WindowManager::find(_pressedWindowType, _pressedWindowNumber);
// if (window == nullptr)
// return;
// WindowManager::close(Ui::WindowType::dropdown, 0);
// window = WindowManager::find(_pressedWindowType, _pressedWindowNumber);
// bool flagSet = hasFlag(Flags::widgetPressed);
// resetFlag(Flags::widgetPressed);
// if (flagSet)
// WindowManager::invalidateWidget(_pressedWindowType, _pressedWindowNumber, _pressedWidgetIndex);
// Input::state(State::normal);
// Ui::ToolTip::setTooltipTimeout(0);
// Ui::ToolTip::setWidgetIndex(_pressedWidgetIndex);
// Ui::ToolTip::setWindowType(_pressedWindowType);
// Ui::ToolTip::setWindowNumber(_pressedWindowNumber);
// if (WindowManager::getCurrentModalType() == Ui::WindowType::undefined || WindowManager::getCurrentModalType() == window->type)
// window->callOnDropdown(_pressedWidgetIndex, window->widgets[_pressedWidgetIndex].id, item);
// // 0x004C7AE7
// func StateWidgetPressed(button MouseButton, x int32, y int32, window Ui::Window, widget Ui::Widget, widgetIndex Ui::WidgetIndex_t) 
// _cursorPressed = { x, y };
// auto pressedWindow = WindowManager::find(_pressedWindowType, _pressedWindowNumber);
// if (pressedWindow == nullptr)
// Input::state(State::reset);
// return;
// if (Input::state() == State::dropdownActive)
// if (Ui::Dropdown::hasFlags(Ui::Dropdown::Flags::unk1))
// if (widgetIndex == -1 || _pressedWindowType != window->type || _pressedWindowNumber != window->number || _pressedWidgetIndex != widgetIndex)
// if (widgetIndex == -1 || window->type != Ui::WindowType::dropdown)
// WindowManager::close(Ui::WindowType::dropdown, 0);
// if (_pressedWindowType != Ui::WindowType::undefined)
// WindowManager::invalidateWidget(_pressedWindowType, _pressedWindowNumber, _pressedWidgetIndex);
// _pressedWindowType = Ui::WindowType::undefined;
// Input::resetFlag(Flags::widgetPressed);
// Input::state(State::reset);
// return;
// bool doShared = false;
// switch (button)
// case MouseButton::released: // 0
// if (window == nullptr)
// break;
// if (window->type == _pressedWindowType && window->number == _pressedWindowNumber && widgetIndex == _pressedWidgetIndex)
// if (!window->isDisabled(widgetIndex))
// if (_clickRepeatTicks != 0)
// _clickRepeatTicks++;
// // Handle click repeat
// if (window->isHoldable(widgetIndex) && _clickRepeatTicks >= 16 && (_clickRepeatTicks % 4) == 0)
// window->callOnMouseDown(widgetIndex, window->widgets[widgetIndex].id);
// bool flagSet = Input::hasFlag(Flags::widgetPressed);
// Input::setFlag(Flags::widgetPressed);
// if (!flagSet)
// WindowManager::invalidateWidget(_pressedWindowType, _pressedWindowNumber, widgetIndex);
// return;
// break;
// case MouseButton::leftPressed: // 1
// if (Input::state() == State::dropdownActive)
// if (window != nullptr && widgetIndex != kWidgetIndexNull)
// auto buttonWidget = &window->widgets[widgetIndex];
// Audio::playSound(Audio::SoundId::clickUp, window->x + buttonWidget->midX());
// return;
// case MouseButton::leftReleased: // 2
// doShared = true;
// break;
// case MouseButton::rightPressed: // 3
// if (Input::state() == State::dropdownActive)
// doShared = true;
// else
// return;
// break;
// case MouseButton::rightReleased:
// return;
// if (doShared)
// // 0x4C7BC7
// if (Input::state() == State::dropdownActive)
// if (window != nullptr)
// if (window->type == Ui::WindowType::dropdown)
// auto item = Ui::Dropdown::dropdownIndexFromPoint(window, x, y);
// if (item.has_value())
// dropdownRegisterSelection(*item);
// else
// if (window->type == _pressedWindowType && window->number == _pressedWindowNumber && widgetIndex == _pressedWidgetIndex)
// if (hasFlag(Flags::flag1))
// bool flagSet = hasFlag(Flags::flag2);
// setFlag(Flags::flag2);
// if (!flagSet)
// return;
// dropdownRegisterSelection(kDropdownItemUndefined);
// // 0x4C7DA0
// WindowManager::close(Ui::WindowType::dropdown, 0);
// window = WindowManager::find(_pressedWindowType, _pressedWindowNumber);
// Input::state(State::normal);
// Ui::ToolTip::setTooltipTimeout(0);
// Ui::ToolTip::setWidgetIndex(_pressedWidgetIndex);
// Ui::ToolTip::setWindowType(_pressedWindowType);
// Ui::ToolTip::setWindowNumber(_pressedWindowNumber);
// if (window != nullptr)
// Audio::playSound(Audio::SoundId::clickUp, window->x + widget->midX());
// if (window != nullptr && window->type == _pressedWindowType && window->number == _pressedWindowNumber && widgetIndex == _pressedWidgetIndex && !window->isDisabled(widgetIndex))
// WindowManager::invalidateWidget(_pressedWindowType, _pressedWindowNumber, _pressedWidgetIndex);
// window->callOnMouseUp(widgetIndex, window->widgets[widgetIndex].id);
// return;
// // 0x4C7F02
// _clickRepeatTicks = 0;
// if (Input::state() != State::dropdownActive)
// bool flagSet = hasFlag(Flags::widgetPressed);
// resetFlag(Flags::widgetPressed);
// if (flagSet)
// WindowManager::invalidateWidget(_pressedWindowType, _pressedWindowNumber, _pressedWidgetIndex);
// if (Input::state() == State::dropdownActive)
// if (window != nullptr && window->type == Ui::WindowType::dropdown)
// auto item = Ui::Dropdown::dropdownIndexFromPoint(window, x, y);
// if (item.has_value())
// Ui::Dropdown::setHighlightedItem(*item);
// WindowManager::invalidate(Ui::WindowType::dropdown, 0);
// // 0x004C8048
// func StateNormal(state MouseButton, x int32, y int32, window Ui::Window, widget Ui::Widget, widgetIndex Ui::WidgetIndex_t) 
// switch (state)
// case MouseButton::leftPressed:
// stateNormalLeft(x, y, window, widgetIndex);
// break;
// case MouseButton::rightPressed:
// stateNormalRight(x, y, window, widgetIndex);
// break;
// case MouseButton::released:
// stateNormalHover(x, y, window, widget, widgetIndex);
// break;
// default:
// break;
// // 0x004C8098
// func StateNormalHover(x int32, y int32, window Ui::Window, widget Ui::Widget, widgetIndex Ui::WidgetIndex_t) 
// Ui::WindowType windowType = Ui::WindowType::undefined;
// Ui::WindowNumber_t windowNumber = 0;
// if (window != nullptr)
// windowType = window->type;
// windowNumber = window->number;
// if (windowType != _hoverWindowType || windowNumber != _hoverWindowNumber || widgetIndex != _hoverWidgetIdx)
// widgetOverFlatbuttonInvalidate();
// _hoverWindowType = windowType;
// _hoverWindowNumber = windowNumber;
// _hoverWidgetIdx = widgetIndex;
// widgetOverFlatbuttonInvalidate();
// if (window != nullptr && widgetIndex != kWidgetIndexNull)
// if (!window->isDisabled(widgetIndex))
// window->callOnMouseHover(widgetIndex, window->widgets[widgetIndex].id);
// StringId tooltipStringId = StringIds::null;
// if (window != nullptr && widgetIndex != kWidgetIndexNull)
// if (widget->type == Ui::WidgetType::scrollview)
// auto res = Ui::ScrollView::getPart(*window, widget, x, y);
// if (res.area == Ui::ScrollPart::none)
// func If(Ui::ScrollPart::view res.area ==) else
// window->callScrollMouseOver(res.scrollviewLoc.x, res.scrollviewLoc.y, static_cast<uint8_t>(res.index));
// else
// tooltipStringId = kScrollWidgetTooltips.at(res.area);
// if (Ui::ToolTip::getWindowType() != Ui::WindowType::undefined)
// if (tooltipStringId != Ui::ToolTip::getCurrentStringId())
// Ui::Windows::ToolTip::closeAndReset();
// if (Ui::ToolTip::getWindowType() != Ui::WindowType::undefined)
// if (window != nullptr && Ui::ToolTip::getWindowType() == window->type && Ui::ToolTip::getWindowNumber() == window->number && Ui::ToolTip::getWidgetIndex() == widgetIndex)
// auto tooltipTimeout = Ui::ToolTip::getTooltipTimeout();
// tooltipTimeout += getTimeSinceLastTick();
// Ui::ToolTip::setTooltipTimeout(tooltipTimeout);
// if (tooltipTimeout >= 8000)
// WindowManager::close(Ui::WindowType::tooltip);
// else
// Ui::Windows::ToolTip::closeAndReset();
// return;
// const auto tooltipNotShownTicks = Ui::ToolTip::getNotShownTicks();
// const auto tooltipMousePos = Ui::ToolTip::getTooltipMouseLocation();
// if (tooltipNotShownTicks < 500 || (x == tooltipMousePos.x && y == tooltipMousePos.y))
// auto tooltipTimeout = Ui::ToolTip::getTooltipTimeout();
// tooltipTimeout += getTimeSinceLastTick();
// Ui::ToolTip::setTooltipTimeout(tooltipTimeout);
// int bp = 2000;
// if (tooltipNotShownTicks <= 1000)
// bp = 0;
// if (bp > tooltipTimeout)
// return;
// if (tooltipStringId == StringIds::null)
// Ui::Windows::ToolTip::open(window, widgetIndex, x, y);
// else
// Ui::Windows::ToolTip::update(window, widgetIndex, tooltipStringId, x, y);
// Ui::ToolTip::setTooltipTimeout(0);
// Ui::ToolTip::setTooltipMouseLocation({ x, y });
// // 0x004C84BE
// func StateNormalLeft(x int32, y int32, window Ui::Window, widgetIndex Ui::WidgetIndex_t) 
// Ui::WindowType windowType = Ui::WindowType::undefined;
// Ui::WindowNumber_t windowNumber = 0;
// if (window != nullptr)
// windowType = window->type;
// windowNumber = window->number;
// WindowManager::close(Ui::WindowType::error);
// WindowManager::close(Ui::WindowType::tooltip);
// // Window might have changed position in the list, therefore find it again
// window = WindowManager::find(windowType, windowNumber);
// if (window == nullptr)
// return;
// window = WindowManager::bringToFront(*window);
// if (widgetIndex == -1)
// return;
// // Widget may have changed memory position during bringToFront
// auto* widget = &window->widgets[widgetIndex];
// // Handle text input focus
// if (widget->type == WidgetType::textbox)
// setFocus(window->type, window->number, widgetIndex);
// WindowManager::invalidateWidget(_focusedWindowType, _focusedWindowNumber, _focusedWidgetIndex);
// else
// resetFocus();
// // Regular widget handling
// switch (widget->type)
// case Ui::WidgetType::caption:
// windowPositionBegin(x, y, window, widgetIndex);
// break;
// case Ui::WidgetType::panel:
// case Ui::WidgetType::newsPanel:
// case Ui::WidgetType::frame:
// if (window->canResize() && (x >= (window->x + window->width - kResizeHandleSize - 1)) && (y >= (window->y + window->height - kResizeHandleSize - 1)))
// windowResizeBegin(x, y, window, widgetIndex);
// else
// windowPositionBegin(x, y, window, widgetIndex);
// break;
// case Ui::WidgetType::viewport:
// state(State::viewportLeft);
// _dragLast.x = x;
// _dragLast.y = y;
// _dragWindowType = window->type;
// _dragWindowNumber = window->number;
// if (hasFlag(Flags::toolActive))
// auto w = WindowManager::find(ToolManager::getToolWindowType(), ToolManager::getToolWindowNumber());
// if (w != nullptr)
// // TODO: Handle the WidgetId properly for tools.
// w->callToolDown(ToolManager::getToolWidgetIndex(), WidgetId::none, x, y);
// setFlag(Flags::leftMousePressed);
// break;
// case Ui::WidgetType::scrollview:
// state(State::scrollLeft);
// _pressedWidgetIndex = widgetIndex;
// _pressedWindowType = window->type;
// _pressedWindowNumber = window->number;
// Ui::ToolTip::setTooltipMouseLocation({ x, y });
// Ui::ScrollView::scrollLeftBegin(x, y, *window, widget, widgetIndex);
// break;
// default:
// if (window->isEnabled(widgetIndex) && !window->isDisabled(widgetIndex))
// Audio::playSound(Audio::SoundId::clickDown, window->x + widget->midX());
// // Set new cursor down widget
// _pressedWidgetIndex = widgetIndex;
// _pressedWindowType = window->type;
// _pressedWindowNumber = window->number;
// setFlag(Flags::widgetPressed);
// state(State::widgetPressed);
// _clickRepeatTicks = 1;
// WindowManager::invalidateWidget(window->type, window->number, widgetIndex);
// window->callOnMouseDown(widgetIndex, window->widgets[widgetIndex].id);
// break;
// // 0x004C834A
// func StateNormalRight(x int32, y int32, window Ui::Window, widgetIndex Ui::WidgetIndex_t) 
// Ui::WindowType windowType = Ui::WindowType::undefined;
// Ui::WindowNumber_t windowNumber = 0;
// if (window != nullptr)
// windowType = window->type;
// windowNumber = window->number;
// WindowManager::close(Ui::WindowType::tooltip);
// // Window might have changed position in the list, therefore find it again
// window = WindowManager::find(windowType, windowNumber);
// if (window == nullptr)
// return;
// window = WindowManager::bringToFront(*window);
// if (widgetIndex == -1)
// return;
// // Widget may have changed memory position during bringToFront
// auto* widget = &window->widgets[widgetIndex];
// if (WindowManager::getCurrentModalType() != Ui::WindowType::undefined)
// if (WindowManager::getCurrentModalType() == window->type)
// ScrollView::scrollModalRight(x, y, *window, widget, widgetIndex);
// return;
// if (SceneManager::isTitleMode())
// return;
// switch (widget->type)
// default:
// break;
// case Ui::WidgetType::viewport:
// viewportDragBegin(window);
// _dragLast.x = x;
// _dragLast.y = y;
// Ui::hideCursor();
// startCursorDrag();
// _mousePosX = 0;
// _mousePosY = 0;
// setFlag(Flags::rightMousePressed);
// break;
// case Ui::WidgetType::scrollview:
// scrollDragBegin(x, y, window, widgetIndex);
// _mousePosX = 0;
// _mousePosY = 0;
// setFlag(Flags::rightMousePressed);
// break;
// // 0x004C877D
// func WindowPositionBegin(x int32, y int32, window Ui::Window, widgetIndex Ui::WidgetIndex_t) 
// state(State::positioningWindow);
// _pressedWidgetIndex = widgetIndex;
// _dragLast.x = x;
// _dragLast.y = y;
// _dragWindowType = window->type;
// _dragWindowNumber = window->number;
// _5233A9 = false;
// func WindowPositionEnd() 
// state(State::normal);
// Ui::ToolTip::setTooltipTimeout(0);
// Ui::ToolTip::setWidgetIndex(_pressedWidgetIndex);
// Ui::ToolTip::setWindowType(_dragWindowType);
// Ui::ToolTip::setWindowNumber(_dragWindowNumber);
// // 0x004C85D1
// func WindowResizeBegin(x int32, y int32, window Ui::Window, widgetIndex Ui::WidgetIndex_t) 
// state(State::resizing);
// _pressedWidgetIndex = widgetIndex;
// _dragLast.x = x;
// _dragLast.y = y;
// _dragWindowType = window->type;
// _dragWindowNumber = window->number;
// window->flags &= ~Ui::WindowFlags::finishedResize;
// func ViewportDragBegin(w Window) 
// w->flags &= ~Ui::WindowFlags::scrollingToLocation;
// state(State::viewportRight);
// _dragWindowType = w->type;
// _dragWindowNumber = w->number;
// _ticksSinceDragStart = 0;
// func ScrollDragBegin(x int32, y int32, window Ui::Window, widgetIndex Ui::WidgetIndex_t) 
// state(State::scrollRight);
// _dragLast.x = x;
// _dragLast.y = y;
// _dragWindowType = window->type;
// _dragWindowNumber = window->number;
// _dragWidgetIndex = widgetIndex;
// _ticksSinceDragStart = 0;
// _dragScrollIndex = window->getScrollDataIndex(widgetIndex);
// Ui::hideCursor();
// startCursorDrag();
// func WidgetOverFlatbuttonInvalidate() 
// Ui::WindowType windowType = _hoverWindowType;
// Ui::WidgetIndex_t widgetIdx = _hoverWidgetIdx;
// Ui::WindowNumber_t windowNumber = _hoverWindowNumber;
// if (windowType == Ui::WindowType::undefined)
// WindowManager::invalidateWidget(windowType, windowNumber, widgetIdx);
// return;
// Ui::Window* oldWindow = WindowManager::find(windowType, windowNumber);
// if (oldWindow != nullptr)
// oldWindow->callPrepareDraw();
// Ui::Widget* oldWidget = widgetIdx != kWidgetIndexNull ? &oldWindow->widgets[widgetIdx] : nullptr;
// if (oldWidget != nullptr
// && (oldWidget->type == Ui::WidgetType::buttonWithColour || oldWidget->type == Ui::WidgetType::buttonWithImage))
// WindowManager::invalidateWidget(windowType, windowNumber, widgetIdx);
// // 0x004CD47A
// func ProcessMouseOver(x int32, y int32) 
// bool skipItem = false;
// Ui::CursorId cursorId = Ui::CursorId::pointer;
// Windows::MapToolTip::reset();
// if (hasMapSelectionFlag(MapSelectionFlags::hoveringOverStation))
// resetMapSelectionFlag(MapSelectionFlags::hoveringOverStation);
// auto station = StationManager::get(_hoveredStationId);
// if (!station->empty())
// station->invalidate();
// Ui::Window* window = Ui::WindowManager::findAt(x, y);
// if (window != nullptr)
// int16_t widgetIdx = window->findWidgetAt(x, y);
// if (widgetIdx != -1)
// Ui::Widget& widget = window->widgets[widgetIdx];
// switch (widget.type)
// case Ui::WidgetType::panel:
// case Ui::WidgetType::newsPanel:
// case Ui::WidgetType::frame:
// if (window->hasFlags(Ui::WindowFlags::resizable))
// if (window->minWidth != window->maxWidth || window->minHeight != window->maxHeight)
// if (x >= window->x + window->width - 19 && y >= window->y + window->height - 19)
// cursorId = Ui::CursorId::diagonalArrows;
// break;
// [[fallthrough]];
// default:
// _scrollLast.x = x;
// _scrollLast.y = y;
// cursorId = window->callCursor(widgetIdx, window->widgets[widgetIdx].id, x, y, cursorId);
// break;
// case Ui::WidgetType::scrollview:
// _scrollLast.x = x;
// _scrollLast.y = y;
// auto res = Ui::ScrollView::getPart(
// *window,
// &window->widgets[widgetIdx],
// x,
// y);
// if (res.area == Ui::ScrollPart::view)
// cursorId = window->callCursor(widgetIdx, window->widgets[widgetIdx].id, res.scrollviewLoc.x, res.scrollviewLoc.y, cursorId);
// break;
// case Ui::WidgetType::viewport:
// if (Input::hasFlag(Flags::toolActive))
// // 3
// cursorId = ToolManager::getToolCursor();
// auto wnd = Ui::WindowManager::find(ToolManager::getToolWindowType(), ToolManager::getToolWindowNumber());
// if (wnd)
// bool out = false;
// cursorId = wnd->callToolCursor(x, y, cursorId, &out);
// if (out)
// skipItem = true;
// else
// switch (ViewportInteraction::getItemLeft(x, y).type)
// case InteractionItem::entity:
// case InteractionItem::townLabel:
// case InteractionItem::stationLabel:
// case InteractionItem::industry:
// case InteractionItem::headquarterBuilding:
// skipItem = true;
// cursorId = Ui::CursorId::handPointer;
// break;
// default:
// break;
// break;
// if (!skipItem)
// ViewportInteraction::rightOver(x, y);
// if (Input::state() == Input::State::resizing)
// cursorId = Ui::CursorId::diagonalArrows;
// if (cursorId != _52336C)
// _52336C = cursorId;
// Ui::setCursor(cursorId);
// Ui::Point getMouseLocation()
// orphan member: return _cursor;
// Ui::Point getMouseLocation2()
// orphan member: return _cursor2;
// Ui::Point getCursorPressedLocation()
// orphan member: return _cursorPressed;
// Ui::Point getDragLastLocation()
// orphan member: return _dragLast;
// func SetDragLastLocation(pos Ui::Point) 
// _dragLast = pos;
// Ui::Point getScrollLastLocation()
// orphan member: return _scrollLast;
// func GetClickRepeatTicks() uint16
// orphan member: return _clickRepeatTicks;
// func GetClickRepeatStepSize() uint32
// // Each 100 ticks increases step size by a factor of 10
// return static_cast<uint32_t>(std::pow(10, getClickRepeatTicks() / 100));
// func SetClickRepeatTicks(ticks uint16) 
// _clickRepeatTicks = ticks;
// func IsRightMouseButtonDown() bool
// orphan member: return _rightMouseButtonDown;
// func SetRightMouseButtonDown(status bool) 
// _rightMouseButtonDown = status;
// // 0x00113E9E0
// static std::queue<QueuedMouseInput> _mouseQueue;
// // 0x00406FEC
// func EnqueueMouseButton(input QueuedMouseInput) 
const MouseQueueSize uint32 = 64
// if (_mouseQueue.size() >= kMouseQueueSize)
// return;
// _mouseQueue.push(input);
// // 0x00407247
// static std::optional<QueuedMouseInput> dequeueMouseInput()
// if (_mouseQueue.empty())
// return std::nullopt;
// std::optional<QueuedMouseInput> res = _mouseQueue.front();
// _mouseQueue.pop();
// orphan member: return res;
// // 0x004C6FCE
// func Loc_4C6FCE(x int32, y int32) MouseButton
// x = _cursor2.x;
// y = _cursor2.y;
// return MouseButton::released;
// // 0x004C70F1
// func RightMouseButtonReleased(x int32, y int32) MouseButton
// stopCursorDrag();
// resetFlag(Flags::rightMousePressed);
// Ui::setCursor(_52336C);
// if (Tutorial::state() == Tutorial::State::playing)
// x = Tutorial::nextInput();
// y = Tutorial::nextInput();
// else
// x = _mousePosX;
// y = _mousePosY;
// // 0x004C7136, 0x004C7165
// _cursor2.x = 0x80000000;
// return MouseButton::rightReleased;
// // 0x004C6EE6
// func NextMouseInput(x int32, y int32) MouseButton
// if (!hasFlag(Flags::rightMousePressed))
// // Interrupt tutorial on mouse button input.
// auto input = dequeueMouseInput();
// if (Tutorial::state() == Tutorial::State::playing && input.has_value())
// Tutorial::stop();
// // If tutorial is playing, follow the recorded mouse coordinates.
// orphan member: MouseButton button{};
// if (Tutorial::state() == Tutorial::State::playing)
// button = MouseButton(Tutorial::nextInput());
// // 0x004C6F6C
// if (button != MouseButton::released)
// button = MouseButton(Tutorial::nextInput());
// x = Tutorial::nextInput();
// y = Tutorial::nextInput();
// else
// button = loc_4C6FCE(x, y);
// if (x < 0)
// orphan member: return button;
// // 0x004C6F5F
// func If(!input) else
// button = loc_4C6FCE(x, y);
// if (x < 0)
// orphan member: return button;
// func If(input) else
// // 0x004C6F87
// switch (input->button)
// case 1:
// button = MouseButton::leftPressed;
// break;
// case 2:
// button = MouseButton::rightPressed;
// break;
// case 3:
// button = MouseButton::leftReleased;
// break;
// default:
// button = MouseButton::rightReleased;
// x = input->pos.x;
// y = input->pos.y;
// // 0x004C6FE4
// x = std::clamp(x, 0, Ui::width() - 1);
// y = std::clamp(y, 0, Ui::height() - 1);
// orphan member: return button;
// else
// if (Tutorial::state() != Tutorial::State::playing)
// if (!isRightMouseButtonDown())
// func RightMouseButtonReleased(x, y) return
// x = _mousePosX;
// y = _mousePosY;
// _mousePosX = 0;
// _mousePosY = 0;
// return MouseButton::released;
// else
// auto button = MouseButton(Tutorial::nextInput());
// if (button == MouseButton::released)
// func RightMouseButtonReleased(x, y) return
// // Note: seemingly repeats the above, but invoking Tutorial::nextInput() moves the script along!
// auto next = Tutorial::nextInput();
// if (!(next & 0x80))
// func RightMouseButtonReleased(x, y) return
// x = Tutorial::nextInput();
// y = Tutorial::nextInput();
// return MouseButton::released;
// // 0x004C6202
// func ProcessMouseWheel() 
// int wheel = 0;
// for (; _cursorWheel > 0; _cursorWheel--)
// wheel -= 17;
// for (; _cursorWheel < 0; _cursorWheel++)
// wheel += 17;
// if (Tutorial::state() != Tutorial::State::none)
// return;
// if (Input::hasFlag(Input::Flags::rightMousePressed))
// if (SceneManager::isTitleMode())
// return;
// auto main = WindowManager::getMainWindow();
// if (main != nullptr && wheel != 0)
// if (wheel > 0)
// main->viewportRotateRight();
// func If(0 any /* wheel < */ ) else
// main->viewportRotateLeft();
// TownManager::updateLabels();
// StationManager::updateLabels();
// Windows::MapWindow::centerOnViewPoint();
// return;
// if (wheel == 0)
// return;
// WindowManager::wheelInput(wheel);
// Ui::WindowType getPressedWindowType()
// orphan member: return _pressedWindowType;
// func SetPressedWindowType(wndType Ui::WindowType) 
// _pressedWindowType = wndType;
// Ui::WindowNumber_t getPressedWindowNumber()
// orphan member: return _pressedWindowNumber;
// func SetPressedWindowNumber(wndNumber Ui::WindowNumber_t) 
// _pressedWindowNumber = wndNumber;
// func HasPendingMouseInputUpdate() bool
// orphan member: return _pendingMouseInputUpdate;
// func ClearPendingMouseInputUpdate() 
// _pendingMouseInputUpdate = false;
// func SetPendingMouseInputUpdate() 
// _pendingMouseInputUpdate = true;
