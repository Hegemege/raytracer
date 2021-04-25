package models

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Ray struct {
	Origin    mgl32.Vec3
	Direction mgl32.Vec3
	Bounce    uint8

	// Image pixel the ray contributes to
	X int
	Y int
}
