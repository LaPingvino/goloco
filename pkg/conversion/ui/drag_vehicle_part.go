package ui

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/DrawingContext.h"
// #include "Graphics/RenderTarget.h"
// #include "Input.h"
// #include "OpenLoco.h"
// #include "Ui/ToolTip.h"
// #include "Ui/Widget.h"
// #include "Ui/Widgets/Wt3Widget.h"
// #include "Ui/WindowManager.h"
// #include "Vehicles/Vehicle.h"
// #include "Vehicles/VehicleBogie.h"
// #include "Vehicles/VehicleDraw.h"
// namespace OpenLoco::Ui::Windows::DragVehiclePart
type Widx int

const (
	Frame Widx = iota
)
// // 0x00522504
// static constexpr auto widgets = makeWidgets(
// Widgets::Wt3Widget({ 0, 0 }, { 150, 60 }, WindowColour::primary)
// );
// static Vehicles::VehicleBogie* _dragCarComponent = nullptr; // 0x0113614E
// static EntityId _dragVehicleHead = EntityId::null;          // 0x01136156
// static const WindowEventList& getEvents();
// // 0x004B3B7E
// func Open(car Vehicles::Car) 
// WindowManager::close(WindowType::dragVehiclePart);
// _dragCarComponent = car.front;
// _dragVehicleHead = car.front->head;
// WindowManager::invalidate(WindowType::vehicle, enumValue(car.front->head));
// uint16_t width = getWidthVehicleInline(car);
// auto pos = Ui::ToolTip::getTooltipMouseLocation();
// pos.y -= 30;
// pos.x -= width / 2;
// Ui::Size size = { width, 60 };
// auto self = WindowManager::createWindow(WindowType::dragVehiclePart, { pos.x, pos.y }, size, WindowFlags::transparent | WindowFlags::stickToFront, getEvents());
// self->setWidgets(widgets);
// self->widgets[widx::frame].right = width - 1;
// Input::windowPositionBegin(Ui::ToolTip::getTooltipMouseLocation().x, Ui::ToolTip::getTooltipMouseLocation().y, self, widx::frame);
// func OnClose(self [[maybe_unused]] Window) 
// _dragCarComponent = nullptr;
// _dragVehicleHead = EntityId::null;
// func OnUpdate(self Window) 
// if (WindowManager::find(WindowType::vehicle, enumValue(_dragVehicleHead)) == nullptr)
// // Parent window no longer exists; close ourselves
// WindowManager::close(&self);
// // 0x004B62FE
// static Ui::CursorId cursor(Window& self, [[maybe_unused]] const WidgetIndex_t widgetIdx, [[maybe_unused]] const WidgetId id, [[maybe_unused]] const int16_t x, [[maybe_unused]] const int16_t y, [[maybe_unused]] const Ui::CursorId fallback)
// self.height = 0; // Set to zero so that skipped in window find
// Vehicle::Details::scrollDrag(Input::getScrollLastLocation());
// self.height = 60;
// return CursorId::dragHand;
// // 0x004B6271
// func OnMove(self Window, x [[maybe_unused]] int16_t, y [[maybe_unused]] int16_t) 
// const auto height = self.height;
// self.height = 0; // Set to zero so that skipped in window find
// Vehicle::Details::scrollDragEnd(Input::getScrollLastLocation());
// // Reset the height so that invalidation works correctly
// self.height = height;
// WindowManager::invalidate(WindowType::vehicle, enumValue(_dragVehicleHead));
// WindowManager::close(&self);
// // 0x004B6197
// func Draw(self Ui::Window, drawingCtx Gfx::DrawingContext) 
// const auto& rt = drawingCtx.currentRenderTarget();
// auto clipped = Gfx::clipRenderTarget(rt, Ui::Rect(self.x, self.y, self.width, self.height));
// if (clipped)
// drawingCtx.pushRenderTarget(*clipped);
// Vehicles::Vehicle train(_dragVehicleHead);
// for (auto& car : train.cars)
// if (car.front == _dragCarComponent)
// drawVehicleInline(drawingCtx, car, { 0, 19 }, VehicleInlineMode::basic, VehiclePartsToDraw::bogies);
// drawVehicleInline(drawingCtx, car, { 0, 19 }, VehicleInlineMode::basic, VehiclePartsToDraw::bodies);
// break;
// drawingCtx.popRenderTarget();
// static constexpr WindowEventList kEvents = {
// .onClose = onClose,
// .onUpdate = onUpdate,
// .cursor = cursor,
// .onMove = onMove,
// .draw = draw,
// static const WindowEventList& getEvents()
// orphan member: return kEvents;
// Vehicles::VehicleBogie* getDragCarComponent()
// orphan member: return _dragCarComponent;
