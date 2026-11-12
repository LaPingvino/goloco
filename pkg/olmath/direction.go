package olmath

// YawToDirectionVector is Trigonometry.hpp kYawToDirectionVector: the
// direction of each of the 64 yaws, scaled to 256 (dumped from upstream).
var YawToDirectionVector = [64][2]int32{
	{-256, 0},
	{-255, 25},
	{-251, 50},
	{-245, 74},
	{-237, 98},
	{-226, 121},
	{-213, 142},
	{-198, 162},
	{-181, 181},
	{-162, 198},
	{-142, 213},
	{-121, 226},
	{-98, 237},
	{-74, 245},
	{-50, 251},
	{-25, 255},
	{0, 256},
	{25, 255},
	{50, 251},
	{74, 245},
	{98, 237},
	{121, 226},
	{142, 213},
	{162, 198},
	{181, 181},
	{198, 162},
	{213, 142},
	{226, 121},
	{237, 98},
	{245, 74},
	{251, 50},
	{255, 25},
	{256, 0},
	{255, -25},
	{251, -50},
	{245, -74},
	{237, -98},
	{226, -121},
	{213, -142},
	{198, -162},
	{181, -181},
	{162, -198},
	{142, -213},
	{121, -226},
	{98, -237},
	{74, -245},
	{50, -251},
	{25, -255},
	{0, -256},
	{-25, -255},
	{-50, -251},
	{-74, -245},
	{-98, -237},
	{-121, -226},
	{-142, -213},
	{-162, -198},
	{-181, -181},
	{-198, -162},
	{-213, -142},
	{-226, -121},
	{-237, -98},
	{-245, -74},
	{-251, -50},
	{-255, -25},
}

// ComputeXYVector is computeXYVector(magnitude, yaw): the direction vector
// times magnitude / 256 (integer division per component, as TVector's).
func ComputeXYVector(magnitude int32, yaw uint8) (x, y int32) {
	v := YawToDirectionVector[yaw&63]
	return v[0] * magnitude / 256, v[1] * magnitude / 256
}

// PitchHorizontalFactor is kPitchHorizontalFactor (0x00503B50): −sin(pitch)·256
// per Pitch; computeXYMagnitude(height, pitch) is factor · height / 256.
var PitchHorizontalFactor = [13]int16{0, -26, -52, -75, -98, 26, 52, 75, 98, -44, 44, -84, 84}
