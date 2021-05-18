package models

import (
	"raytracer/utility"

	"github.com/go-gl/mathgl/mgl32"
)

type Ray struct {
	Origin    mgl32.Vec3
	Direction mgl32.Vec3
	Bounce    uint8

	// Image pixel the ray contributes to
	X int
	Y int

	// Helpers
	InvDirection mgl32.Vec3
	Sign         [3]uint8
}

func NewRay(origin mgl32.Vec3, dir mgl32.Vec3, bounce uint8, x int, y int) *Ray {
	return &Ray{
		Origin:    origin,
		Direction: dir,
		Bounce:    bounce,
		X:         x,
		Y:         y,
		InvDirection: mgl32.Vec3{
			1.0 / dir.X(),
			1.0 / dir.Y(),
			1.0 / dir.Z(),
		},
		Sign: [3]uint8{
			utility.BoolToInt(dir.X() < 0),
			utility.BoolToInt(dir.Y() < 0),
			utility.BoolToInt(dir.Z() < 0),
		},
	}
}
