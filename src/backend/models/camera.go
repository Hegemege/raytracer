package models

import "github.com/go-gl/mathgl/mgl32"

type ProjectionType int

const (
	Perspective ProjectionType = iota
	Ortographic
)

type Camera struct {
	Transform    mgl32.Mat4
	AspectWidth  float32
	AspectHeight float32
	Projection   ProjectionType
}
