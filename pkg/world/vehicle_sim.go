package world

import (
	"math"

	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// vehicle_sim.go — minimal vehicle movement along placed track.
//
// A SimVehicle follows a chain of connected track pieces. Each tick it advances
// a floating-point distance along the current piece (whose geometry comes from
// kTrackCoordinates), interpolates the rendered body entity's world position
// linearly between the piece's begin and end, and hands off to the next piece
// when it runs off the end. With no continuation it reverses (bounces).
//
// This is deliberately simple: curves are treated as straight segments between
// their two endpoints. The piece-to-piece hand-off mirrors the construction
// validation in CanPlaceTrackAtHead — a piece begins where the previous piece
// ends when their rotations and positions agree (kTrackCoordinates RotBegin /
// RotEnd + the DX/DY/DZ world-unit delta).
//
// OpenLoco reference: src/OpenLoco/src/Vehicles/VehicleManager.cpp (movement),
// src/OpenLoco/src/Map/Track/TrackData.cpp (kTrackCoordinates).

// tileCenterOffset centres a vehicle within a tile (tile = 32 world units).
// The value is constant across all pieces, so a piece's end position equals the
// next piece's begin position and vehicles cross junctions seamlessly.
const tileCenterOffset = 16

// SimVehicle is one track-following vehicle. The current track piece is
// identified by its origin (sequence-0) element: tile, TrackID, Rotation, BaseZ
// — the same identity the construction system places. Travel distance is kept
// as a float measured from the piece's begin so sub-tile precision survives
// even though entity positions are int16.
type SimVehicle struct {
	// Current piece identity (origin / SequenceIndex 0 element).
	tileX, tileY int
	trackID      int
	rotation     uint8 // element rotation 0-3
	baseZ        uint8 // origin element BaseZ (SmallZ)

	reversed bool    // false: travelling begin→end; true: end→begin
	dist     float64 // distance from the piece begin, in [0, length]
	speed    float64 // world units per tick

	// consist: rendered body entities driven by this vehicle. Indices into
	// w.entities. For the slice a single body is enough, but the field is a
	// slice so a multi-car consist can be added later.
	entityIdx []int
}

// pieceGeom returns the begin/end world position, world-unit delta and the
// begin/end cardinal rotation of the SimVehicle's current track piece.
func (sv *SimVehicle) pieceGeom() (bx, by, bz, ex, ey, ez float64, c trackCoord) {
	c = kTrackCoordinates[sv.trackID<<3|int(sv.rotation&3)]
	bx = float64(sv.tileX*32 + tileCenterOffset)
	by = float64(sv.tileY*32 + tileCenterOffset)
	bz = float64(sv.baseZ) * 4
	ex = bx + float64(c.DX)
	ey = by + float64(c.DY)
	ez = bz + float64(c.DZ)
	return
}

// pieceLength is the euclidean length of the piece delta in world units.
// A zero-length piece (shouldn't occur) is clamped to 1 to avoid div-by-zero.
func pieceLength(c trackCoord) float64 {
	l := math.Hypot(float64(c.DX), float64(c.DY))
	if l == 0 {
		return 1
	}
	return l
}

// rotToYaw maps a track rotation to a 0-63 sprite yaw: cardinal r (0-3) → r*16,
// diagonal r (12-15) → (r-12)*16+8.
func rotToYaw(r uint8) uint8 {
	if r < 4 {
		return r * 16
	}
	return (r-12)*16 + 8
}

// findTrackOrigin returns the sequence-0 track element on tile (x,y) at the
// given BaseZ whose piece begins with rotation beginRot (kTrackCoordinates
// RotBegin). This is the forward hand-off: the next piece starts where the
// current one ends, exactly as CanPlaceTrackAtHead validates a placement.
func (w *World) findTrackOrigin(x, y int, baseZ, beginRot uint8) (scenario.TrackElement, bool) {
	t := w.getTile(x, y)
	if t == nil {
		return scenario.TrackElement{}, false
	}
	for _, te := range t.tracks {
		if te.SequenceIndex != 0 || te.BaseZ != baseZ {
			continue
		}
		if int(te.TrackID)<<3 >= len(kTrackCoordinates) {
			continue
		}
		if kTrackCoordinates[int(te.TrackID)<<3|int(te.Rotation&3)].RotBegin == beginRot {
			return te, true
		}
	}
	return scenario.TrackElement{}, false
}

// findTrackEnding scans the world for a sequence-0 track element whose piece
// ENDS at tile (x,y) at the given BaseZ with end rotation endRot. This is the
// reverse hand-off (a vehicle bouncing back off a piece begin looks for the
// piece feeding into it). Unlike the forward case the predecessor's origin can
// be on any tile, so this is a full scan; acceptable for the small piece counts
// of the vertical slice.
func (w *World) findTrackEnding(x, y int, baseZ, endRot uint8) (tileX, tileY int, te scenario.TrackElement, ok bool) {
	for ty := 0; ty < w.height; ty++ {
		for tx := 0; tx < w.width; tx++ {
			for _, cand := range w.tiles[ty][tx].tracks {
				if cand.SequenceIndex != 0 {
					continue
				}
				if int(cand.TrackID)<<3 >= len(kTrackCoordinates) {
					continue
				}
				c := kTrackCoordinates[int(cand.TrackID)<<3|int(cand.Rotation&3)]
				endX := tx + int(c.DX)/32
				endY := ty + int(c.DY)/32
				endZ := int(cand.BaseZ) + int(c.DZ)/4
				if endX == x && endY == y && endZ == int(baseZ) && c.RotEnd == endRot {
					return tx, ty, cand, true
				}
			}
		}
	}
	return 0, 0, scenario.TrackElement{}, false
}

// SpawnTestVehicle places a SimVehicle at the start of a track piece in the
// world and appends a single body entity (rendered by paintVehicleEntities).
// It prefers a trackID-0 piece with no predecessor so the vehicle starts at a
// dead-end and travels the whole line; falling back to any track element.
// objID is used as the entity's vehicle ObjectID (index into ObjMgr.Vehicles).
// Returns false if the world contains no track.
func (w *World) SpawnTestVehicle(objID int) bool {
	var (
		best      *scenario.TrackElement
		bestX     int
		bestY     int
		bestScore int // 2: trackID0 + no predecessor, 1: no predecessor, 0: any
		found     bool
	)
	for ty := 0; ty < w.height; ty++ {
		for tx := 0; tx < w.width; tx++ {
			for i := range w.tiles[ty][tx].tracks {
				te := &w.tiles[ty][tx].tracks[i]
				if te.SequenceIndex != 0 || int(te.TrackID)<<3 >= len(kTrackCoordinates) {
					continue
				}
				beginRot := kTrackCoordinates[int(te.TrackID)<<3|int(te.Rotation&3)].RotBegin
				_, _, _, hasPrev := w.findTrackEnding(tx, ty, te.BaseZ, beginRot)
				score := 0
				if !hasPrev {
					score = 1
					if te.TrackID == 0 {
						score = 2
					}
				}
				if !found || score > bestScore {
					best = te
					bestX, bestY = tx, ty
					bestScore = score
					found = true
				}
			}
		}
	}
	if !found {
		return false
	}

	sv := &SimVehicle{
		tileX:    bestX,
		tileY:    bestY,
		trackID:  int(best.TrackID),
		rotation: best.Rotation & 3,
		baseZ:    best.BaseZ,
		speed:    1.0,
	}

	ent := scenario.VehicleEntity{
		EntityType:       scenario.VehicleTypeBody,
		ObjectID:         uint16(objID),
		ObjectSpriteType: 0,
		Owner:            0,
	}
	w.entities = append(w.entities, ent)
	sv.entityIdx = []int{len(w.entities) - 1}

	w.simVehicles = append(w.simVehicles, sv)
	w.writeVehicleEntity(sv)
	return true
}

// TickVehicles advances every SimVehicle one simulation step. Called from the
// game update in gameplay mode.
func (w *World) TickVehicles() { w.UpdateVehicles() }

// UpdateVehicles advances every SimVehicle by its speed and writes the new
// interpolated entity positions.
func (w *World) UpdateVehicles() {
	for _, sv := range w.simVehicles {
		w.advanceVehicle(sv)
		w.writeVehicleEntity(sv)
	}
}

// advanceVehicle moves one vehicle along its piece, handing off to a connected
// piece when it runs off an end, or bouncing when there is none.
func (w *World) advanceVehicle(sv *SimVehicle) {
	// Guard against pathological loops if pieces are extremely short: cap the
	// number of hand-offs processed in a single tick.
	for iter := 0; iter < 16; iter++ {
		_, _, _, _, _, _, c := sv.pieceGeom()
		length := pieceLength(c)

		if !sv.reversed {
			sv.dist += sv.speed
			if sv.dist < length {
				return
			}
			carry := sv.dist - length
			// Reached the end: look for the next piece beginning here.
			endX := sv.tileX + int(c.DX)/32
			endY := sv.tileY + int(c.DY)/32
			endZ := uint8(int(sv.baseZ) + int(c.DZ)/4)
			if te, ok := w.findTrackOrigin(endX, endY, endZ, c.RotEnd); ok {
				sv.tileX, sv.tileY = endX, endY
				sv.trackID = int(te.TrackID)
				sv.rotation = te.Rotation & 3
				sv.baseZ = te.BaseZ
				sv.reversed = false
				sv.dist = carry
				continue
			}
			// No continuation: bounce back onto this piece.
			sv.reversed = true
			sv.dist = length - carry
			if sv.dist < 0 {
				sv.dist = 0
			}
			return
		}

		// Reversed: travelling end→begin.
		sv.dist -= sv.speed
		if sv.dist > 0 {
			return
		}
		carry := -sv.dist
		// Reached the begin: look for the predecessor feeding into this piece.
		beginRot := c.RotBegin
		if px, py, te, ok := w.findTrackEnding(sv.tileX, sv.tileY, sv.baseZ, beginRot); ok {
			pc := kTrackCoordinates[int(te.TrackID)<<3|int(te.Rotation&3)]
			plen := pieceLength(pc)
			sv.tileX, sv.tileY = px, py
			sv.trackID = int(te.TrackID)
			sv.rotation = te.Rotation & 3
			sv.baseZ = te.BaseZ
			sv.reversed = true // enter at its end, travel toward its begin
			sv.dist = plen - carry
			if sv.dist < 0 {
				sv.dist = 0
			}
			continue
		}
		// No predecessor: bounce forward onto this piece.
		sv.reversed = false
		sv.dist = carry
		if sv.dist > length {
			sv.dist = length
		}
		return
	}
}

// writeVehicleEntity writes the vehicle's interpolated world position and yaw
// into its rendered body entities. Sub-tile precision is preserved by rounding
// the float world position to int16 only here.
func (w *World) writeVehicleEntity(sv *SimVehicle) {
	bx, by, bz, ex, ey, ez, c := sv.pieceGeom()
	length := pieceLength(c)
	t := sv.dist / length
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	px := bx + t*(ex-bx)
	py := by + t*(ey-by)
	pz := bz + t*(ez-bz)

	var yaw uint8
	if sv.reversed {
		yaw = (rotToYaw(c.RotBegin) + 32) & 63
	} else {
		yaw = rotToYaw(c.RotEnd)
	}

	for _, idx := range sv.entityIdx {
		if idx < 0 || idx >= len(w.entities) {
			continue
		}
		e := &w.entities[idx]
		e.PosX = coords.WorldUnits(int16(math.Round(px)))
		e.PosY = coords.WorldUnits(int16(math.Round(py)))
		e.PosZ = int16(math.Round(pz))
		e.SpriteYaw = yaw
	}
}
