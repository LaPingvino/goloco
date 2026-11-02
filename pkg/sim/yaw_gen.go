// Code generated from OpenLoco src/Vehicles/VehicleBody.cpp (calculateYaw*FromVector) by a script in tick 214. DO NOT EDIT.

package sim

// yawFn is one calculateYaw*FromVector: i starts at XNeg for a negative x
// and adds YNeg for a negative y; then one step per threshold xy has reached
// (xy = |x|<<16 / |y|; their nested ifs are a binary search over these);
// the yaw is Table[i].
type yawFn struct {
	XNeg, YNeg int
	Thresholds []uint32
	Table      []uint8
}

// yaw0 is calculateYaw0FromVector.
var yaw0 = yawFn{2, 4, []uint32{65536}, []uint8{16, 32, 16, 0, 48, 32, 48, 0}}

// yaw1Plane is calculateYaw1FromVectorPlane.
var yaw1Plane = yawFn{3, 6, []uint32{3434, 1250501}, []uint8{16, 24, 32, 16, 8, 0, 48, 40, 32, 48, 56, 0}}

// yaw1 is calculateYaw1FromVector.
var yaw1 = yawFn{3, 6, []uint32{27146, 158218}, []uint8{16, 24, 32, 16, 8, 0, 48, 40, 32, 48, 56, 0}}

// yaw2 is calculateYaw2FromVector.
var yaw2 = yawFn{5, 10, []uint32{13036, 43790, 98082, 329472}, []uint8{16, 20, 24, 28, 32, 16, 12, 8, 4, 0, 48, 44, 40, 36, 32, 48, 52, 56, 60, 0}}

// yaw3 is calculateYaw3FromVector.
var yaw3 = yawFn{9, 18, []uint32{6455, 19880, 35030, 53784, 79856, 122609, 216043, 665398}, []uint8{16, 18, 20, 22, 24, 26, 28, 30, 32, 16, 14, 12, 10, 8, 6, 4, 2, 0, 48, 46, 44, 42, 40, 38, 36, 34, 32, 48, 50, 52, 54, 56, 58, 60, 62, 0}}

// yaw4 is calculateYaw4FromVector.
var yaw4 = yawFn{17, 34, []uint32{3220, 9721, 16416, 23449, 30996, 39281, 48605, 59398, 72308, 88365, 109340, 138564, 183161, 261634, 441808, 1334016}, []uint8{16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 48, 47, 46, 45, 44, 43, 42, 41, 40, 39, 38, 37, 36, 35, 34, 33, 32, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 0}}
