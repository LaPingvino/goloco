package ui

import "github.com/LaPingvino/goloco/pkg/gfx8"

// Vehicle window tab icons: Common::drawTabs (Vehicle.cpp 0x004B5F0D) — the
// interface skin's images, the current tab's animated by the window's frameNo:
// main by vehicle type (kTabIconByVehicleType, frames 0–7 at its speed,
// recoloured to the company colour), details the wrench (frameNo >> 1, 16
// frames), cargo (>> 3, 4), finances (>> 1, 16), route (>> 4, 4).
const (
	skinTabWrench         = 97  // InterfaceSkin::ImageIds::tab_wrench_frame0
	skinTabFinances       = 113 // tab_finances_frame0
	skinTabRoutes         = 192 // tab_routes_frame_0
	skinTabCargoDelivered = 198 // tab_cargo_delivered_frame0
)

// vehicleTabIcons is kTabIconByVehicleType: image and frame speed by
// VehicleType (train, bus, truck, tram, aircraft, ship).
var vehicleTabIcons = [6]struct{ image, speed int }{
	{203, 1}, {219, 1}, {235, 1}, {227, 1}, {211, 2}, {243, 3},
}

// drawVehicleTabs draws the five tabs (those not hidden) with their icons;
// active is the current tab's widget, prim the company colour's map.
func (w *VehicleWindow) drawVehicleTabs(rt *gfx8.RenderTarget, lw []BVWidget, active int, sec, prim []uint8) {
	for i := vwTabMain; i <= vwTabRoute; i++ {
		t := lw[i]
		if t.Hidden {
			continue
		}
		icon, iconMap := tabNoIcon, []uint8(nil)
		if w.SkinImage > 0 {
			on := i == active
			frame := func(shift, mask int) int {
				if on {
					return (w.FrameNo >> shift) & mask
				}
				return 0
			}
			switch i {
			case vwTabMain:
				m := vehicleTabIcons[max(0, min(5, w.VehicleType))]
				icon, iconMap = w.SkinImage+m.image+frame(m.speed, 7), prim // Gfx::recolour(…, companyColour)
			case vwTabDetails:
				icon = w.SkinImage + skinTabWrench + frame(1, 0xF)
			case vwTabCargo:
				icon = w.SkinImage + skinTabCargoDelivered + frame(3, 3)
			case vwTabFinances:
				icon = w.SkinImage + skinTabFinances + frame(1, 0xF)
			case vwTabRoute:
				icon = w.SkinImage + skinTabRoutes + frame(4, 3)
			}
		}
		drawTab(rt, t.L, t.T, i == active, sec, icon, iconMap)
	}
}
