package world

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// TrainCar is one Car of a train as drawVehicleInline reads it: its vehicle
// object, the object's inline y offset, its components' bogies and bodies,
// whether the car is reversed (its first body's isReversed) and whether the
// train's brake lights are on (veh2 brakeLightTimeout).
type TrainCar struct {
	Object     *objects.VehicleObject
	YOffset    int
	Components []objects.CarComponent
	Reversed   bool
	Brake      bool
}

// TrainCars is Vehicle(head).cars: from the head through veh1 and veh2 (the
// NextCarID chain), then cars of (front bogie, back bogie, body) components —
// a body_start begins a car, body_continued ones extend it — up to the tail.
// nil when the chain is broken ("Bad vehicle structure").
func (w *World) TrainCars(headID uint16) []TrainCar {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return nil
	}
	om := w.renderer.ObjMgr
	byID := make(map[uint16]*scenario.VehicleEntity, len(w.Entities))
	for i := range w.Entities {
		byID[w.Entities[i].ID] = &w.Entities[i]
	}
	head := byID[headID]
	if head == nil || head.EntityType != scenario.VehicleTypeHead {
		return nil
	}
	veh1 := byID[head.NextCarID]
	if veh1 == nil {
		return nil
	}
	veh2 := byID[veh1.NextCarID]
	if veh2 == nil || veh2.EntityType != scenario.VehicleTypeVeh2 {
		return nil
	}
	part := func(e *scenario.VehicleEntity) objects.CarPart {
		return objects.CarPart{SpriteType: e.ObjectSpriteType, Reversed: e.Flags38&(1<<1) != 0, Animation: e.Animation,
			CargoFrame: e.CargoFrame, Colour: [2]uint8{e.ColourPrimary, e.ColourSecondary}}
	}
	var cars []TrainCar
	cur := byID[veh2.NextCarID]
	for steps := 0; cur != nil && cur.EntityType != scenario.VehicleTypeTail && steps < 1000; steps++ {
		front := cur
		back := byID[front.NextCarID]
		if front.EntityType != scenario.VehicleTypeBogie || back == nil || back.EntityType != scenario.VehicleTypeBogie {
			return nil
		}
		body := byID[back.NextCarID]
		if body == nil || !scenario.IsVehicleBody(body.EntityType) {
			return nil
		}
		comp := objects.CarComponent{Front: part(front), Back: part(back), Body: part(body)}
		if body.EntityType == scenario.VehicleTypeBodyStart || len(cars) == 0 {
			var obj *objects.VehicleObject
			if int(front.ObjectID) < len(om.Vehicles) {
				obj = om.Vehicles[front.ObjectID]
			}
			if obj == nil {
				return nil
			}
			cars = append(cars, TrainCar{Object: obj, YOffset: om.VehicleYDisplayOffset(obj), Reversed: body.Flags38&(1<<1) != 0,
				Brake: veh2.BrakeLightTimeout != 0})
		}
		c := &cars[len(cars)-1]
		c.Components = append(c.Components, comp)
		cur = byID[body.NextCarID]
	}
	return cars
}
