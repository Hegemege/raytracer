package models

import "github.com/go-gl/mathgl/mgl32"

type Object struct {
	Position   mgl32.Vec3
	MaterialID int
	Material   *Material
}
